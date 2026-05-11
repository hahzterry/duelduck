package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	sol "dd-prediction-api/internal/client/solana"
	"dd-prediction-api/internal/model"
	"dd-prediction-api/pkg/apperrors"

	"github.com/gagliardetto/solana-go"
	lookup "github.com/gagliardetto/solana-go/programs/address-lookup-table"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	USDCMintAddress = solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")

	OldTokenProgramID = solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	NewTokenProgramID = solana.MustPublicKeyFromBase58("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb")
)

const (
	USDCMintDecimals uint8 = 6
	SolMintDecimals  uint8 = 9
)

type proceedTransferData struct {
	SenderID              uuid.UUID
	SenderAccount         solana.PrivateKey
	SenderTokenAddress    solana.PublicKey
	RecipientAddress      solana.PublicKey
	RecipientTokenAddress solana.PublicKey
	Mint                  solana.PublicKey
	Amount                uint64
	Decimals              uint8
}

func (s *WalletService) newProceedTransferData(
	senderID uuid.UUID,
	sender solana.PrivateKey,
	recipient solana.PublicKey,
	mint solana.PublicKey,
	amount uint64,
	decimals uint8,
	solanaToken *model.SolanaToken,
) (*proceedTransferData, error) {

	senderTokenAddress, err := s.findATA(sender.PublicKey(), solanaToken.Mint, solanaToken.ProgramID)
	if err != nil {
		return nil, apperrors.BadRequest("failed to get sender's associated token account", err)
	}

	recipientTokenAddress, err := s.findATA(recipient, solanaToken.Mint, solanaToken.ProgramID)
	if err != nil {
		return nil, apperrors.BadRequest("failed to get recipient's associated token account", err)
	}

	return &proceedTransferData{
		SenderID:              senderID,
		SenderAccount:         sender,
		SenderTokenAddress:    senderTokenAddress,
		RecipientAddress:      recipient,
		RecipientTokenAddress: recipientTokenAddress,
		Mint:                  mint,
		Amount:                amount,
		Decimals:              decimals,
	}, nil
}

var ZeroValuePublicKey solana.PublicKey

func (s *WalletService) GetMintInfo(ctx context.Context, tokenKey solana.PublicKey) (*token.Mint, error) {
	var mint token.Mint
	err := s.SolanaRPC.GetAccountDataBorshInto(ctx, tokenKey, &mint)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to get mint details", err)
	}

	return &mint, nil
}

func (s *WalletService) senderExists(ctx context.Context, senderID uuid.UUID) error {
	ok, err := s.UserRepository.Exists(ctx, &model.User{ID: senderID})
	if err != nil {
		return apperrors.Internal("failed to get sender by id", err)
	}

	if !ok {
		return apperrors.BadRequest("user with id does not exist", nil)
	}

	return nil
}

const (
	Finalized = rpc.CommitmentFinalized
	Confirmed = rpc.CommitmentConfirmed
)

func (s *WalletService) getTokenBalance(
	ctx context.Context,
	ata solana.PublicKey,
	commitment rpc.CommitmentType,
) (uint64, error) {
	balance, err := s.SolanaRPC.GetTokenAccountBalance(ctx, ata, commitment)
	if err != nil {
		if s.isAccountUninitialized(err) {
			return 0, model.ErrTokenAccountUninitialized
		}

		return 0, apperrors.ServiceUnavailable("failed to get ata balance", err)
	}

	if balance == nil || balance.Value == nil {
		return 0, apperrors.ServiceUnavailable("failed to get ata balance: balance is nil", nil)
	}

	balanceAmount, err := strconv.ParseUint(balance.Value.Amount, 10, 64)
	if err != nil {
		return 0, apperrors.ServiceUnavailable("failed to parse token balance amount", err)
	}

	return balanceAmount, nil
}

func (s *WalletService) GetTokenBalance(
	ctx context.Context,
	ata solana.PublicKey,
) (uint64, error) {
	return s.getTokenBalance(ctx, ata, Confirmed)
}

func (s *WalletService) HasEnoughTokenBalance(
	ctx context.Context,
	ata solana.PublicKey,
	requiredAmount float64,
) (bool, error) {
	tokenBalance, err := s.getTokenBalance(ctx, ata, Confirmed)
	if err != nil {
		return false, err
	}

	return float64(tokenBalance) >= requiredAmount, nil
}

func (s *WalletService) HasEnoughTokenBalanceFinalized(
	ctx context.Context,
	ata solana.PublicKey,
	requiredAmount uint64,
) (bool, error) {
	balanceAmount, err := s.getTokenBalance(ctx, ata, Finalized)
	if err != nil {
		return false, err
	}

	return balanceAmount >= requiredAmount, nil
}

func (s *WalletService) getSolBalance(
	ctx context.Context,
	pk solana.PublicKey,
	commitment rpc.CommitmentType,
) (uint64, error) {
	balance, err := s.SolanaRPC.GetBalance(ctx, pk, commitment)
	if err != nil || balance == nil {
		return 0, apperrors.ServiceUnavailable("failed to get ata balance", err)
	}

	return balance.Value, nil
}

func (s *WalletService) GetSolBalance(
	ctx context.Context,
	pk solana.PublicKey,
) (uint64, error) {
	return s.getSolBalance(ctx, pk, Confirmed)
}

func (s *WalletService) HasEnoughSolBalance(
	ctx context.Context,
	pk solana.PublicKey,
	requiredAmount float64,
) (bool, error) {
	balance, err := s.GetSolBalance(ctx, pk)
	if err != nil {
		return false, err
	}

	if float64(balance) < requiredAmount {
		return false, nil
	}

	return true, nil
}

const (
	// TransferTransactionInstructionsCount
	// Fist two are Compute Unit Price and Compute Unit Limit Instructions
	// Recipient token account initialization instruction if needed
	// And token transfer instruction
	TransferTransactionInstructionsCount = 4

	FallBackCUTransfer = 5026

	// FallBackCUTransferChecked and FallBackCUTransferWithTokenAccountInit are average values
	// we got by simulating transactions
	FallBackCUTransferChecked              = 6254
	FallBackCUTransferWithTokenAccountInit = 30195

	FallbackComputeUnitPrice = 343400

	// CUExtraCapacityCoefficient sometimes transaction
	// may consume a bit more Compute Units then usual
	CUExtraCapacityCoefficient = 1.05
)

func (s *WalletService) proceedTransfer(ctx context.Context, data *proceedTransferData) (string, error) {
	instructions := make([]solana.Instruction, 0, TransferTransactionInstructionsCount)

	info, err := s.SolanaRPC.GetAccountInfo(ctx, data.RecipientTokenAddress)
	if err != nil && !errors.Is(err, rpc.ErrNotFound) {
		return "", apperrors.ServiceUnavailable("failed to get recipient's account info", err)
	}

	if info == nil || info.Value == nil || info.Value.Owner == ZeroValuePublicKey {
		initTokenAccountInstruction, err := associatedtokenaccount.NewCreateInstruction(
			data.SenderAccount.PublicKey(),
			data.RecipientAddress,
			data.Mint).ValidateAndBuild()
		if err != nil {
			return "", apperrors.Internal("failed to build token account initialization instruction", err)
		}

		instructions = append(instructions, initTokenAccountInstruction)
	}

	transferInstruction, err := token.NewTransferCheckedInstruction(
		data.Amount,
		data.Decimals,
		data.SenderTokenAddress,
		data.Mint,
		data.RecipientTokenAddress,
		data.SenderAccount.PublicKey(),
		[]solana.PublicKey{data.SenderAccount.PublicKey()}).ValidateAndBuild()
	if err != nil {
		return "", apperrors.Internal("failed to build token transfer instruction", err)
	}

	instructions = append(instructions, transferInstruction)

	tx, err := s.NewTransactionForSimulation(
		instructions,
		txSignerPrivateKeyGetter(data.SenderAccount),
		solana.TransactionPayer(data.SenderAccount.PublicKey()))
	if err != nil {
		return "", err
	}

	computeUnits, err := s.GetSimulationComputeUnits(ctx, tx)
	if err != nil {
		if errors.Is(err, sol.ErrInsufficientFunds) {
			return "", err
		}

		if len(instructions) > 1 {
			computeUnits = FallBackCUTransferWithTokenAccountInit
		} else {
			computeUnits = FallBackCUTransferChecked
		}
	}

	computeUnits = uint32(float64(computeUnits) * CUExtraCapacityCoefficient)
	cuPriceInstruction, err := computebudget.NewSetComputeUnitPriceInstructionBuilder().
		SetMicroLamports(s.PriorityTracker.GetMediumPriorityMicroLamports()).
		ValidateAndBuild()
	if err != nil {
		return "", apperrors.Internal("failed to set transaction compute unit price", err)
	}

	cuLimitInstruction, err := computebudget.NewSetComputeUnitLimitInstructionBuilder().
		SetUnits(computeUnits).
		ValidateAndBuild()
	if err != nil {
		return "", apperrors.Internal("failed to set transaction compute unit limit", err)
	}

	// Compute Unit Price and Compute Unit Limit instructions must be first
	instructions = append([]solana.Instruction{cuPriceInstruction, cuLimitInstruction}, instructions...)

	tx, err = solana.NewTransaction(
		instructions,
		solana.Hash{},
		solana.TransactionPayer(data.SenderAccount.PublicKey()))
	if err != nil {
		return "", apperrors.Internal("failed to create transaction", err)
	}

	sig, err := s.sendTxWithTracker(
		ctx,
		tx,
		txSignerPrivateKeyGetter(data.SenderAccount))
	if err != nil {
		return "", err
	}

	return sig.String(), nil
}

func (s *WalletService) transferSol(
	ctx context.Context,
	senderID uuid.UUID,
	sender solana.PrivateKey,
	recipient solana.PublicKey,
	rawAmount uint64,
) (string, error) {
	instructions := make([]solana.Instruction, 0, TransferTransactionInstructionsCount)

	transferInstruction := system.NewTransferInstruction(
		rawAmount,
		sender.PublicKey(),
		recipient,
	).Build()

	instructions = append(instructions, transferInstruction)

	smTxInstructions := []solana.Instruction{
		computebudget.NewSetComputeUnitLimitInstructionBuilder().
			SetUnits(1_400_000).
			Build(),
		computebudget.NewSetComputeUnitPriceInstruction(FallbackComputeUnitPrice).
			Build(),
	}

	smTxInstructions = append(smTxInstructions, instructions...)

	tx, err := s.NewTransactionForSimulation(
		smTxInstructions,
		txSignerPrivateKeyGetter(sender),
		solana.TransactionPayer(sender.PublicKey()),
	)
	if err != nil {
		return "", err
	}

	computeUnits, err := s.GetSimulationComputeUnits(ctx, tx)
	if err != nil {
		if errors.Is(err, sol.ErrInsufficientFunds) {
			return "", err
		}

		computeUnits = FallBackCUTransferChecked
	}

	computeUnits = uint32(float64(computeUnits) * CUExtraCapacityCoefficient)
	cuPriceInstruction, err := computebudget.NewSetComputeUnitPriceInstructionBuilder().
		SetMicroLamports(s.PriorityTracker.GetMediumPriorityMicroLamports()).
		ValidateAndBuild()
	if err != nil {
		return "", apperrors.Internal("failed to set transaction compute unit price", err)
	}

	cuLimitInstruction, err := computebudget.NewSetComputeUnitLimitInstructionBuilder().
		SetUnits(computeUnits).
		ValidateAndBuild()
	if err != nil {
		return "", apperrors.Internal("failed to set transaction compute unit limit", err)
	}

	// Compute Unit Price and Compute Unit Limit instructions must be first
	instructions = append([]solana.Instruction{cuPriceInstruction, cuLimitInstruction}, instructions...)

	tx, err = solana.NewTransaction(
		instructions,
		solana.Hash{},
		solana.TransactionPayer(sender.PublicKey()))
	if err != nil {
		return "", apperrors.Internal("failed to create transaction", err)
	}

	// Send transaction
	sig, err := s.sendTxWithTracker(
		ctx,
		tx,
		txSignerPrivateKeyGetter(sender),
	)
	if err != nil {
		return "", err
	}

	return sig.String(), nil
}

var (
	TransactionMaxRetryCount uint = 10
)

func (s *WalletService) sendTransaction(
	ctx context.Context,
	tx *solana.Transaction,
	privateKeyGetter func(key solana.PublicKey) *solana.PrivateKey,
) (solana.Signature, error) {
	opts := rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: Finalized,
		MaxRetries:          &TransactionMaxRetryCount,
	}

	if err := s.RefreshBlockHash(ctx, &tx.Message); err != nil {
		return solana.Signature{}, err
	}

	_, err := tx.Sign(privateKeyGetter)
	if err != nil {
		return solana.Signature{}, apperrors.Internal("failed to sign transaction", err)
	}

	sig, err := s.SolanaRPC.SendTransactionWithOpts(ctx, tx, opts)
	if err != nil {
		if s.isAccountUninitialized(err) {
			zap.L().Error("account is uninitialized", zap.Error(err))
			return solana.Signature{}, model.ErrTokenAccountUninitialized
		}
		return solana.Signature{}, apperrors.Internal("failed to send transaction", err)
	}

	return sig, nil
}

func (s *WalletService) RefreshBlockHash(ctx context.Context, txMessage *solana.Message) error {
	recentBlockHashResp, err := s.SolanaRPC.GetLatestBlockhash(ctx, Finalized)
	if err != nil {
		return apperrors.ServiceUnavailable("failed to get latest block hash", err)
	}

	txMessage.RecentBlockhash = recentBlockHashResp.Value.Blockhash
	return nil
}

func (s *WalletService) GetSimulationComputeUnits(
	ctx context.Context,
	tx *solana.Transaction,
) (uint32, error) {
	opts := &rpc.SimulateTransactionOpts{
		ReplaceRecentBlockhash: true,
		SigVerify:              false, // conflicts with ReplaceRecentBlockhash
	}

	result, err := s.simulateTransaction(ctx, tx, opts)
	if err != nil {
		return 0, err
	}

	if result == nil {
		return 0, apperrors.Internal("transaction simulation: tx value is nil")
	}

	if result.Err != nil {
		zap.L().Error("transaction simulation", zap.Any("err", result.Err))

		if s.isAccountUninitialized(fmt.Errorf("%v", result.Err)) {
			return 0, apperrors.Internal(model.ErrSolanaAccountUninitialized)
		}

		return 0, sol.ParseLogsForError(result.Logs)
	}

	if result.UnitsConsumed == nil {
		return 0, apperrors.Internal("transaction simulation: 0 units consumed")
	}

	return uint32(*result.UnitsConsumed), nil
}

func (s *WalletService) simulateTransaction(
	ctx context.Context,
	tx *solana.Transaction,
	opts *rpc.SimulateTransactionOpts,
) (*rpc.SimulateTransactionResult, error) {
	sTx, err := s.SolanaRPC.SimulateTransactionWithOpts(ctx, tx, opts)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to send simulation transaction", err)
	}

	if sTx == nil {
		return nil, apperrors.ServiceUnavailable("failed to get simulation transaction compute units, tx is nil", err)
	}

	return sTx.Value, nil
}

func txSignerPrivateKeyGetter(sender solana.PrivateKey) func(solana.PublicKey) *solana.PrivateKey {
	return func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(sender.PublicKey()) {
			return &sender
		}

		return nil
	}
}

func txTwoSignersPrivateKeyGetter(admin, user solana.PrivateKey) func(solana.PublicKey) *solana.PrivateKey {
	return func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(admin.PublicKey()) {
			return &admin
		}

		if key.Equals(user.PublicKey()) {
			return &user
		}

		return nil
	}
}

func (s *WalletService) NewTransactionForSimulation(
	instructions []solana.Instruction,
	privateKeyGetter func(key solana.PublicKey) *solana.PrivateKey,
	opts ...solana.TransactionOption,
) (*solana.Transaction, error) {
	tx, err := solana.NewTransaction(
		instructions,
		solana.Hash{}, // latest block hash will be set just before transaction sending or transaction simulation
		opts...)
	if err != nil {
		return nil, apperrors.Internal("failed to create transaction for simulation", err)
	}

	_, err = tx.Sign(privateKeyGetter)
	if err != nil {
		return nil, apperrors.Internal("failed to sign a transaction for simulation", err)
	}

	return tx, nil
}

// EstimateTxFeeLamports estimates total fee (base + priority) for an already built transaction.
func (s *WalletService) EstimateTxFeeLamports(
	ctx context.Context,
	tx *solana.Transaction,
) (uint64, error) {
	// Estimate compute units via simulation.
	cu, err := s.GetSimulationComputeUnits(ctx, tx)
	if err != nil || cu == 0 {
		// fallback if simulation fails or returns 0
		cu = FallBackCUTransferChecked
	}

	// Priority fee from CU and price per CU (in micro-lamports).
	microLamports := s.PriorityTracker.GetMediumPriorityMicroLamports()
	priorityFee := uint64(cu) * uint64(microLamports) / 1_000_000

	// Base fee via getFeeForMessage.
	msgBytes, err := tx.Message.MarshalBinary()
	if err != nil {
		return 0, apperrors.Internal("failed to marshal tx message", err)
	}

	msgBase64 := base64.StdEncoding.EncodeToString(msgBytes)

	feeResp, err := s.SolanaRPC.GetFeeForMessage(
		ctx,
		msgBase64,
		rpc.CommitmentProcessed,
	)

	// small fallback
	baseFee := uint64(5_000)
	if err == nil && feeResp != nil && feeResp.Value != nil {
		baseFee = *feeResp.Value
	}

	return baseFee + priorityFee, nil
}

func (s *WalletService) processTransactionWithAddressLookups(
	ctx context.Context,
	txx *solana.Transaction,
) error {
	if !txx.Message.IsVersioned() {
		return apperrors.ServiceUnavailable("invalid tx: only versioned transactions can contain lookups")
	}

	tblKeys := txx.Message.GetAddressTableLookups().GetTableIDs()
	if len(tblKeys) == 0 {
		return apperrors.ServiceUnavailable("no lookup tables in versioned tx")
	}

	numLookups := txx.Message.GetAddressTableLookups().NumLookups()
	if numLookups == 0 {
		return apperrors.ServiceUnavailable("no lookups in versioned transaction")
	}

	resolutions := make(map[solana.PublicKey]solana.PublicKeySlice, len(tblKeys))

	for _, key := range tblKeys {
		info, err := s.SolanaRPC.GetAccountInfo(ctx, key)
		if err != nil {
			return apperrors.ServiceUnavailable("failed to get account info", err)
		}

		tableContent, err := lookup.DecodeAddressLookupTableState(info.GetBinary())
		if err != nil {
			return apperrors.Internal("failed to decode address lookup table state", err)
		}

		resolutions[key] = tableContent.Addresses
	}

	err := txx.Message.SetAddressTables(resolutions)
	if err != nil {
		return apperrors.Internal("failed to set address tables", err)
	}

	err = txx.Message.ResolveLookups()
	if err != nil {
		return apperrors.Internal("failed to resolve lookups", err)
	}

	return nil
}

const TxConfirmationTimeout = time.Second * 40

func (s *WalletService) sendTxWithTracker(
	ctx context.Context,
	tx *solana.Transaction,
	privateKeyGetter func(key solana.PublicKey) *solana.PrivateKey,
) (solana.Signature, error) {
	sig, err := s.sendTransaction(ctx, tx, privateKeyGetter)
	if err != nil {
		return solana.Signature{}, err
	}

	sent, err := s.SigTracker.SubscribeForSignatureStatus(sig, TxConfirmationTimeout)
	if err != nil && !errors.Is(err, rpc.ErrNotConfirmed) {
		return solana.Signature{}, apperrors.Internal("subscribe to signature status", err)
	}

	if !sent {
		return solana.Signature{}, apperrors.Internal("tx was not confirmed: " + sig.String())
	}

	return sig, nil
}

func (s *WalletService) findATA(
	pk solana.PublicKey,
	mintStr,
	programID string,
) (solana.PublicKey, error) {
	mint, err := solana.PublicKeyFromBase58(mintStr)
	if err != nil {
		return solana.PublicKey{}, apperrors.Internal("failed to get token mint", err)
	}

	var ata solana.PublicKey

	if programID == OldTokenProgramID.String() {
		ata, _, err = solana.FindAssociatedTokenAddress(pk, mint)
	} else if programID == NewTokenProgramID.String() {
		ata, _, err = solana.FindProgramAddress(
			[][]byte{
				pk.Bytes(),
				NewTokenProgramID.Bytes(),
				mint.Bytes(),
			},
			associatedtokenaccount.ProgramID,
		)
	}
	if err != nil {
		return solana.PublicKey{}, apperrors.Internal("failed to get user associated token address", err)
	}

	return ata, nil
}
