package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/pkg/apperrors"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	solanaPacketDataSizeBytes = 1232

	combinedExternalTxMaxComputeUnits = 1_400_000
)

func (s *WalletService) InitAndJoinSolanaRoom(
	ctx context.Context,
	duel *model.Duel,
	user *model.User,
	answer uint8,
	tokenInfo *model.DuelTokenInfo,
) (string, string, error) {
	var (
		duelPrice = uint64(duel.DuelPrice * tokenInfo.CryptoDuelPriceMultiplier())

		initEndpoint string
		initReqBody  = map[string]interface{}{
			"theme":       model.DuelTopicPlaceholder,
			"description": duel.Question,
			"percent":     uint32(duel.CommissionRate),
			"bet":         duelPrice,
			"end":         uint32(duel.Deadline.Unix()),
			"pda_nr":      uint32(duel.RoomNumber),
		}
	)

	if duel.Symbol == model.SOLSymbol {
		initEndpoint = "init-sol"
	} else {
		if duel.CreatedAt.After(s.newContractStartDate) {
			initEndpoint = "init-new"
		} else {
			initEndpoint = "init"
		}

		initReqBody["token_mint"] = tokenInfo.Mint
	}

	txInit, roomTokenPDA, err := s.WalletAuthorityClient.GetInitTxFromContract(initReqBody, initEndpoint)
	if err != nil {
		return "", "", err
	}

	userPublicKey, err := solana.PublicKeyFromBase58(user.WalletAddress)
	if err != nil {
		return "", "", apperrors.Internal("failed to parse active wallet address", err)
	}

	var (
		userTokenAccount solana.PublicKey
		hasEnoughBalance bool
	)

	if tokenInfo.Symbol == model.SOLSymbol {
		userTokenAccount = userPublicKey

		hasEnoughBalance, err = s.HasEnoughSolBalance(ctx, userTokenAccount, float64(duelPrice))
		if err != nil {
			return "", "", err
		}
	} else {
		userTokenAccount, err = s.findATA(userPublicKey, tokenInfo.Mint, tokenInfo.ProgramID)
		if err != nil {
			return "", "", apperrors.Internal("failed to get user associated token address", err)
		}

		hasEnoughBalance, err = s.HasEnoughTokenBalance(ctx, userTokenAccount, float64(duelPrice))
		if err != nil {
			return "", "", err
		}
	}

	if !hasEnoughBalance {
		return "", "", apperrors.BadRequest("not enough balance to proceed a transaction")
	}

	var (
		joinEndpoint string

		joinReqBody = map[string]interface{}{
			"multiplier":   1,
			"answer":       answer,
			"pda_nr":       duel.RoomNumber,
			"user_address": userPublicKey,
		}
	)

	if tokenInfo.Symbol == model.SOLSymbol {
		joinEndpoint = "join-sol"
	} else {
		if duel.CreatedAt.After(s.newContractStartDate) {
			joinEndpoint = "join-new"

			joinReqBody["token_program"] = tokenInfo.ProgramID
			joinReqBody["token_mint"] = tokenInfo.Mint
		} else {
			joinEndpoint = "join"
		}

		joinReqBody["from_token_account"] = userTokenAccount
		joinReqBody["to_token_account"] = roomTokenPDA
	}

	txJoin, err := s.WalletAuthorityClient.GetTxFromContract(joinReqBody, joinEndpoint)
	if err != nil {
		return "", "", err
	}

	instructions, err := GetTxInstructions(txInit, txJoin)
	if err != nil {
		return "", "", err
	}

	recentBlockhashResp, err := s.SolanaRPC.GetLatestBlockhash(ctx, Finalized)
	if err != nil {
		return "", "", apperrors.Internal("failed to get recent blockhash", err)
	}

	tx, err := solana.NewTransaction(
		instructions,
		recentBlockhashResp.Value.Blockhash,
		solana.TransactionPayer(userPublicKey))
	if err != nil {
		return "", "", apperrors.ServiceUnavailable("failed to generate a transaction", err)
	}

	_, err = tx.PartialSign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(s.solanaAdminPrivateKey.PublicKey()) {
			return &s.solanaAdminPrivateKey
		}
		return nil
	})
	if err != nil {
		return "", "", apperrors.ServiceUnavailable("failed to sign transaction", err)
	}

	txBytes, err := tx.MarshalBinary()
	if err != nil {
		return "", "", apperrors.ServiceUnavailable("failed to marshal transaction", err)
	}

	encodedTx := base64.StdEncoding.EncodeToString(txBytes)

	return encodedTx, roomTokenPDA, nil
}

func (s *WalletService) JoinSolanaRoom(
	ctx context.Context,
	duel *model.Duel,
	user *model.User,
	answer uint8,
	tokenInfo *model.DuelTokenInfo,
) (string, error) {
	var (
		duelPrice = duel.DuelPrice * tokenInfo.CryptoDuelPriceMultiplier()
	)

	pdaInfo, err := s.WalletAuthorityClient.GetPdaInfo(map[string]any{"pda_nr": duel.RoomNumber})
	if err != nil {
		return "", err
	}

	var pdaInitBytes uint64
	if duel.CreatedAt.After(s.newPdaInitBytesDate) && duel.Symbol != model.SOLSymbol {
		pdaInitBytes = model.NewPdaInitBytes
	} else {
		pdaInitBytes = model.OldPdaInitBytes
	}

	multiplier := ((pdaInfo.BytesSize - pdaInitBytes) / model.PdaMultiplierStepBytes) + 1

	requiredMultiplier := uint64(1)
	if duel.PlayersCount > 0 {
		requiredMultiplier = (duel.PlayersCount-1)/10 + 1
	}
	if multiplier >= requiredMultiplier {
		return s.joinSolanaRoomWithExternalWallet(ctx, duel, user, answer, tokenInfo, multiplier)
	}

	// extend room size
	multiplier++

	userPublicKey, err := solana.PublicKeyFromBase58(user.WalletAddress)
	if err != nil {
		return "", apperrors.Internal("failed to parse active wallet address", err)
	}

	var (
		userTokenAccount solana.PublicKey
		hasEnoughBalance bool
	)

	if tokenInfo.Symbol == model.SOLSymbol {
		userTokenAccount = userPublicKey

		hasEnoughBalance, err = s.HasEnoughSolBalance(ctx, userTokenAccount, float64(duelPrice))
		if err != nil {
			return "", err
		}
	} else {
		userTokenAccount, err = s.findATA(userPublicKey, tokenInfo.Mint, tokenInfo.ProgramID)
		if err != nil {
			return "", apperrors.Internal("failed to get user associated token address", err)
		}

		hasEnoughBalance, err = s.HasEnoughTokenBalance(ctx, userTokenAccount, duelPrice)
		if err != nil {
			return "", err
		}
	}

	if !hasEnoughBalance {
		return "", apperrors.BadRequest("not enough balance to proceed a transaction")
	}

	var (
		reallocateEndpoint string = "memory"
		joinEndpoint       string

		reallocateReqBody = map[string]any{"pda_nr": duel.RoomNumber}
		joinReqBody       = map[string]interface{}{
			"multiplier":   multiplier,
			"answer":       answer,
			"pda_nr":       duel.RoomNumber,
			"user_address": userPublicKey,
		}
	)

	if tokenInfo.Symbol == model.SOLSymbol {
		joinEndpoint = "join-sol"
	} else {
		if duel.CreatedAt.After(s.newContractStartDate) {
			reallocateEndpoint = "memory-new"
			joinEndpoint = "join-new"

			joinReqBody["token_program"] = tokenInfo.ProgramID
			joinReqBody["token_mint"] = tokenInfo.Mint
		} else {
			joinEndpoint = "join"
		}

		joinReqBody["from_token_account"] = userTokenAccount
		joinReqBody["to_token_account"] = duel.RoomTokenPDA
	}

	txReallocate, err := s.WalletAuthorityClient.GetTxFromContract(reallocateReqBody, reallocateEndpoint)
	if err != nil {
		return "", err
	}

	txJoin, err := s.WalletAuthorityClient.GetTxFromContract(joinReqBody, joinEndpoint)
	if err != nil {
		return "", err
	}

	instructions, err := GetTxInstructions(txReallocate, txJoin)
	if err != nil {
		return "", err
	}

	recentBlockhashResp, err := s.SolanaRPC.GetLatestBlockhash(ctx, Finalized)
	if err != nil {
		return "", apperrors.Internal("failed to get recent blockhash", err)
	}

	tx, err := solana.NewTransaction(
		instructions,
		recentBlockhashResp.Value.Blockhash,
		solana.TransactionPayer(userPublicKey))
	if err != nil {
		return "", apperrors.ServiceUnavailable("failed to generate a transaction", err)
	}

	_, err = tx.PartialSign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(s.solanaAdminPrivateKey.PublicKey()) {
			return &s.solanaAdminPrivateKey
		}
		return nil
	})
	if err != nil {
		return "", apperrors.ServiceUnavailable("failed to sign transaction", err)
	}

	txBytes, err := tx.MarshalBinary()
	if err != nil {
		return "", apperrors.ServiceUnavailable("failed to marshal transaction", err)
	}

	encodedTx := base64.StdEncoding.EncodeToString(txBytes)

	return encodedTx, nil
}

func (s *WalletService) joinSolanaRoomWithExternalWallet(
	ctx context.Context,
	duel *model.Duel,
	user *model.User,
	answer uint8,
	tokenInfo *model.DuelTokenInfo,
	multiplier uint64,
) (string, error) {
	duelPrice := duel.DuelPrice * tokenInfo.CryptoDuelPriceMultiplier()

	userPublicKey, err := solana.PublicKeyFromBase58(user.WalletAddress)
	if err != nil {
		return "", apperrors.Internal("failed to parse active wallet address", err)
	}

	var (
		userTokenAccount solana.PublicKey
		hasEnoughBalance bool
	)

	if tokenInfo.Symbol == model.SOLSymbol {
		userTokenAccount = userPublicKey

		hasEnoughBalance, err = s.HasEnoughSolBalance(ctx, userTokenAccount, float64(duelPrice))
		if err != nil {
			return "", err
		}
	} else {
		userTokenAccount, err = s.findATA(userPublicKey, tokenInfo.Mint, tokenInfo.ProgramID)
		if err != nil {
			return "", apperrors.Internal("failed to get user associated token address", err)
		}

		hasEnoughBalance, err = s.HasEnoughTokenBalance(ctx, userTokenAccount, duelPrice)
		if err != nil {
			return "", err
		}
	}

	if !hasEnoughBalance {
		return "", apperrors.BadRequest("not enough balance to proceed a transaction")
	}

	var (
		joinEndpoint string

		joinReqBody = map[string]interface{}{
			"multiplier":   multiplier,
			"answer":       answer,
			"pda_nr":       duel.RoomNumber,
			"user_address": userPublicKey,
		}
	)

	if tokenInfo.Symbol == model.SOLSymbol {
		joinEndpoint = "join-sol"
	} else {
		if duel.CreatedAt.After(s.newContractStartDate) {
			joinEndpoint = "join-new"

			joinReqBody["token_program"] = tokenInfo.ProgramID
			joinReqBody["token_mint"] = tokenInfo.Mint
		} else {
			joinEndpoint = "join"
		}

		joinReqBody["from_token_account"] = userTokenAccount
		joinReqBody["to_token_account"] = duel.RoomTokenPDA
	}

	txJoin, err := s.WalletAuthorityClient.GetTxFromContract(joinReqBody, joinEndpoint)
	if err != nil {
		return "", err
	}

	instructions, err := GetTxInstructions(txJoin)
	if err != nil {
		return "", err
	}

	recentBlockhashResp, err := s.SolanaRPC.GetLatestBlockhash(ctx, Finalized)
	if err != nil {
		return "", apperrors.Internal("failed to get recent blockhash", err)
	}

	tx, err := solana.NewTransaction(
		instructions,
		recentBlockhashResp.Value.Blockhash,
		solana.TransactionPayer(userPublicKey))
	if err != nil {
		return "", apperrors.ServiceUnavailable("failed to generate a transaction", err)
	}

	_, err = tx.PartialSign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(s.solanaAdminPrivateKey.PublicKey()) {
			return &s.solanaAdminPrivateKey
		}
		return nil
	})
	if err != nil {
		return "", apperrors.ServiceUnavailable("failed to sign transaction", err)
	}

	txBytes, err := tx.MarshalBinary()
	if err != nil {
		return "", apperrors.ServiceUnavailable("failed to marshal transaction", err)
	}

	encodedTx := base64.StdEncoding.EncodeToString(txBytes)

	return encodedTx, nil
}

func buildCombinedExternalWalletTransaction(
	encodedTxs []string,
	payer solana.PublicKey,
	recentBlockhash solana.Hash,
	adminPrivateKey solana.PrivateKey,
	priorityMicroLamports uint64,
	maxComputeUnits uint32,
) (*solana.Transaction, error) {
	if len(encodedTxs) == 0 {
		return nil, apperrors.BadRequest("no transactions to combine")
	}

	instructions := make([]solana.Instruction, 0, len(encodedTxs)*4)
	for _, encodedTx := range encodedTxs {
		encodedTx = strings.TrimSpace(encodedTx)
		if encodedTx == "" {
			return nil, apperrors.BadRequest("empty transaction payload")
		}

		tx, err := solana.TransactionFromBase64(encodedTx)
		if err != nil {
			return nil, apperrors.BadRequest("failed to decode transaction payload", err)
		}

		txInstructions, err := RemoveComputeBudgetInstructionsFromTx(tx)
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, txInstructions...)
	}

	if len(instructions) == 0 {
		return nil, apperrors.BadRequest("no instructions to combine")
	}

	cuPriceInstruction, err := computebudget.NewSetComputeUnitPriceInstructionBuilder().
		SetMicroLamports(priorityMicroLamports).
		ValidateAndBuild()
	if err != nil {
		return nil, apperrors.Internal("failed to set transaction compute unit price", err)
	}

	cuLimitInstruction, err := computebudget.NewSetComputeUnitLimitInstructionBuilder().
		SetUnits(maxComputeUnits).
		ValidateAndBuild()
	if err != nil {
		return nil, apperrors.Internal("failed to set transaction compute unit limit", err)
	}

	instructions = append([]solana.Instruction{cuPriceInstruction, cuLimitInstruction}, instructions...)

	tx, err := solana.NewTransaction(
		instructions,
		recentBlockhash,
		solana.TransactionPayer(payer),
	)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to generate a transaction", err)
	}

	_, err = tx.PartialSign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(adminPrivateKey.PublicKey()) {
			return &adminPrivateKey
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to sign transaction", err)
	}

	return tx, nil
}

func (s *WalletService) PrecheckExternalEncodedTxLimits(
	ctx context.Context,
	encodedTx string,
) error {
	encodedTx = strings.TrimSpace(encodedTx)
	if encodedTx == "" {
		return apperrors.BadRequest("empty transaction payload")
	}

	tx, err := solana.TransactionFromBase64(encodedTx)
	if err != nil {
		return apperrors.BadRequest("failed to decode transaction payload", err)
	}

	return s.precheckPreparedExternalTxLimits(ctx, tx, combinedExternalTxMaxComputeUnits)
}

func (s *WalletService) precheckPreparedExternalTxLimits(
	ctx context.Context,
	tx *solana.Transaction,
	maxComputeUnits uint32,
) error {
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		return apperrors.ServiceUnavailable("failed to marshal transaction", err)
	}

	if len(txBytes) > solanaPacketDataSizeBytes {
		return apperrors.BadRequest(
			fmt.Sprintf("transaction too large: %d bytes (max %d)", len(txBytes), solanaPacketDataSizeBytes),
		)
	}

	computeUnits, err := s.GetSimulationComputeUnits(ctx, tx)
	if err != nil {
		return apperrors.BadRequest("transaction compute precheck failed", err)
	}
	if computeUnits > maxComputeUnits {
		return apperrors.BadRequest(
			fmt.Sprintf("transaction exceeds compute limit: %d > %d", computeUnits, maxComputeUnits),
		)
	}

	return nil
}

func (s *WalletService) RewardDuelWinners(
	ctx context.Context,
	winAmount uint64,
	winners []model.CryptoDuelPlayer,
	tokenInfo *model.DuelTokenInfo,
	projectID uuid.UUID,
) ([]string, error) {
	if len(winners) == 0 {
		return []string{}, nil
	}

	adminKey, err := s.getProjectAdminKey(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if tokenInfo.Symbol != model.SOLSymbol {
		adminTokenAccount, err := s.findATA(adminKey.PublicKey(), tokenInfo.Mint, tokenInfo.ProgramID)
		if err != nil || adminTokenAccount == ZeroValuePublicKey {
			return nil, apperrors.Internal("failed to find associated token account for user rewarding", err)
		}
	}

	allTransferInstructions, err := s.getTransferInstruction(adminKey, winAmount, winners, tokenInfo)
	if err != nil {
		return nil, err
	}

	batches := separateInstructions(allTransferInstructions)
	return s.sendBatchTransferTransactions(ctx, batches, adminKey)
}

// RewardDuelWinnersWithAmounts used for duel rewarding with ranged price duel type
func (s *WalletService) RewardDuelWinnersWithAmounts(
	ctx context.Context,
	winners []model.CryptoDuelPlayer,
	tokenInfo *model.DuelTokenInfo,
	projectID uuid.UUID,
) ([]string, error) {
	return s.RewardDuelWinners(ctx, 0, winners, tokenInfo, projectID)
}

func (s *WalletService) sendBatchTransferTransactions(
	ctx context.Context,
	batches [][]solana.Instruction,
	adminKey solana.PrivateKey,
) ([]string, error) {
	txHashes := make([]string, 0, len(batches))

	for _, instructions := range batches {
		tx, err := s.NewTransactionForSimulation(
			instructions,
			txSignerPrivateKeyGetter(adminKey),
			solana.TransactionPayer(adminKey.PublicKey()))
		if err != nil {
			return nil, err
		}

		computeUnits, err := s.GetSimulationComputeUnits(ctx, tx)
		if err != nil {
			if len(instructions) > 1 {
				computeUnits = FallBackCUTransfer * uint32(len(instructions))
			}
		}

		computeUnits = uint32(float64(computeUnits)*CUExtraCapacityCoefficient + 300)
		cuPriceInstruction, err := computebudget.NewSetComputeUnitPriceInstructionBuilder().
			SetMicroLamports(s.PriorityTracker.GetHighPriorityMicroLamports()).
			ValidateAndBuild()
		if err != nil {
			return nil, apperrors.Internal("failed to set transaction compute unit price", err)
		}

		cuLimitInstruction, err := computebudget.NewSetComputeUnitLimitInstructionBuilder().
			SetUnits(computeUnits).
			ValidateAndBuild()
		if err != nil {
			return nil, apperrors.Internal("failed to set transaction compute unit limit", err)
		}

		// Compute Unit Price and Compute Unit Limit instructions must be first
		instructions = append([]solana.Instruction{cuPriceInstruction, cuLimitInstruction}, instructions...)

		tx, err = solana.NewTransaction(
			instructions,
			solana.Hash{},
			solana.TransactionPayer(adminKey.PublicKey()))
		if err != nil {
			return nil, apperrors.Internal("failed to create transaction", err)
		}

		txHash, err := s.sendTransaction(
			ctx,
			tx,
			txSignerPrivateKeyGetter(adminKey))
		if err != nil {
			return nil, err
		}

		txHashes = append(txHashes, txHash.String())
	}

	return txHashes, nil
}

func (s *WalletService) TransferSymbol(
	ctx context.Context,
	recipientAddress solana.PublicKey,
	amount uint64,
	tokenInfo *model.DuelTokenInfo,
) (string, error) {
	if amount == 0 {
		return "", nil
	}

	if tokenInfo.Symbol == model.SOLSymbol {
		transferInstruction, err := system.NewTransferInstruction(
			amount,
			s.solanaAdminPrivateKey.PublicKey(),
			recipientAddress,
		).ValidateAndBuild()
		if err != nil {
			return "", apperrors.Internal("failed to build transfer transaction", err)
		}

		txHash, err := s.SendTransaction(ctx, []solana.Instruction{transferInstruction})
		if err != nil {
			return "", err
		}

		return txHash, nil
	}

	adminTokenAccount, err := s.findATA(s.solanaAdminPrivateKey.PublicKey(), tokenInfo.Mint, tokenInfo.ProgramID)
	if err != nil || adminTokenAccount == ZeroValuePublicKey {
		return "", apperrors.Internal("failed to find associated token account for user rewarding", err)
	}

	duelOwnerATA, err := s.findATA(recipientAddress, tokenInfo.Mint, tokenInfo.ProgramID)
	if err != nil || duelOwnerATA == ZeroValuePublicKey {
		return "", apperrors.BadRequest("failed to find recipient associated token account", err)
	}

	info, err := s.SolanaRPC.GetAccountInfo(ctx, duelOwnerATA)
	if err != nil && !errors.Is(err, rpc.ErrNotFound) {
		return "", apperrors.ServiceUnavailable("failed to get recipient's account info", err)
	}

	mint, err := solana.PublicKeyFromBase58(tokenInfo.Mint)
	if err != nil {
		return "", apperrors.Internal("failed to get token mint", err)
	}

	inst := make([]solana.Instruction, 0, 2)
	if info == nil || info.Value == nil || info.Value.Owner == ZeroValuePublicKey {
		initTokenAccountInstruction, err := associatedtokenaccount.NewCreateInstruction(
			s.solanaAdminPrivateKey.PublicKey(),
			recipientAddress,
			mint).ValidateAndBuild()
		if err != nil {
			return "", apperrors.Internal("failed to build token account initialization instruction", err)
		}

		inst = append(inst, initTokenAccountInstruction)
	}

	transferInstruction, err := token.NewTransferInstruction(
		amount,
		adminTokenAccount,
		duelOwnerATA,
		s.solanaAdminPrivateKey.PublicKey(),
		[]solana.PublicKey{s.solanaAdminPrivateKey.PublicKey()}).ValidateAndBuild()
	if err != nil {
		return "", apperrors.Internal("failed to build transfer transaction", err)
	}

	inst = append(inst, transferInstruction)

	txHash, err := s.SendTransaction(ctx, inst)
	if err != nil {
		return "", err
	}

	return txHash, nil
}

func (s *WalletService) SendTransaction(
	ctx context.Context,
	instructions []solana.Instruction,
) (string, error) {
	tx, err := s.NewTransactionForSimulation(
		instructions,
		txSignerPrivateKeyGetter(s.solanaAdminPrivateKey),
		solana.TransactionPayer(s.solanaAdminPrivateKey.PublicKey()))
	if err != nil {
		return "", err
	}

	computeUnits, err := s.GetSimulationComputeUnits(ctx, tx)
	if err != nil {
		if len(instructions) > 1 {
			computeUnits = FallBackCUTransfer * uint32(len(instructions))
		}
	}

	computeUnits = uint32(float64(computeUnits)*CUExtraCapacityCoefficient + 300)
	cuPriceInstruction, err := computebudget.NewSetComputeUnitPriceInstructionBuilder().
		SetMicroLamports(s.PriorityTracker.GetHighPriorityMicroLamports()).
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
		solana.TransactionPayer(s.solanaAdminPrivateKey.PublicKey()))
	if err != nil {
		return "", apperrors.Internal("failed to create transaction", err)
	}

	txHash, err := s.sendTransaction(
		ctx,
		tx,
		txSignerPrivateKeyGetter(s.solanaAdminPrivateKey))
	if err != nil {
		return "", err
	}

	return txHash.String(), nil
}

func (s *WalletService) TransferBulkSolanaChain(
	ctx context.Context,
	amount uint64,
	players []model.CryptoDuelPlayer,
	tokenInfo *model.DuelTokenInfo,
) ([]string, error) {

	if tokenInfo.Symbol != model.SOLSymbol {
		adminTokenAccount, err := s.findATA(s.solanaAdminPrivateKey.PublicKey(), tokenInfo.Mint, tokenInfo.ProgramID)
		if err != nil || adminTokenAccount == ZeroValuePublicKey {
			return nil, apperrors.Internal("failed to find associated token account for user rewarding", err)
		}
	}
	allTransferInstructions, err := s.getTransferInstruction(s.solanaAdminPrivateKey, amount, players, tokenInfo)
	if err != nil {
		return nil, err
	}

	batches := separateInstructions(allTransferInstructions)
	return s.sendBatchTransferTransactions(ctx, batches, s.solanaAdminPrivateKey)
}

func (s *WalletService) getTransferInstruction(
	sender solana.PrivateKey,
	amount uint64,
	players []model.CryptoDuelPlayer,
	tokenInfo *model.DuelTokenInfo,
) ([]solana.Instruction, error) {
	if tokenInfo.Symbol == model.SOLSymbol {
		instructions, err := getSolTransferInstructions(sender, amount, players)
		return instructions, err
	}

	senderTokenAccount, err := s.findATA(sender.PublicKey(), tokenInfo.Mint, tokenInfo.ProgramID)
	if err != nil || senderTokenAccount == ZeroValuePublicKey {
		return nil, apperrors.Internal("failed to find associated token account for user rewarding", err)
	}

	var (
		instructions = make([]solana.Instruction, 0, len(players)+2)
	)

	ctx := context.Background()
	for _, player := range players {
		recipient, err := solana.PublicKeyFromBase58(player.PublicAddress)
		if err != nil || recipient == ZeroValuePublicKey {
			return nil, apperrors.BadRequest("recipient is not valid solana address", err)
		}

		// check for ranged/fixed duel type
		// if amount == 0 -> ranged price duel type
		rewardAmount := amount
		if player.Amount > 0 {
			rewardAmount = player.Amount
		}

		playerInstructions, err := s.buildRewardInstructionsForRecipient(
			ctx,
			player.UserID,
			sender,
			senderTokenAccount,
			recipient,
			rewardAmount,
			tokenInfo,
		)
		if err != nil {
			return nil, err
		}

		instructions = append(instructions, playerInstructions...)
	}

	return instructions, nil
}

func getSolTransferInstructions(
	sender solana.PrivateKey,
	amount uint64,
	players []model.CryptoDuelPlayer,
) ([]solana.Instruction, error) {
	from := sender.PublicKey()
	instructions := make([]solana.Instruction, 0, len(players))

	for _, player := range players {
		recipient, err := solana.PublicKeyFromBase58(player.PublicAddress)
		if err != nil || recipient == ZeroValuePublicKey {
			return nil, apperrors.BadRequest("recipient is not valid solana address", err)
		}

		transferAmount := amount
		if player.Amount > 0 {
			transferAmount = player.Amount
		}

		transferInstruction, err := system.NewTransferInstruction(
			transferAmount,
			from,
			recipient,
		).ValidateAndBuild()
		if err != nil {
			return nil, apperrors.Internal("failed to build transfer transaction", err)
		}

		instructions = append(instructions, transferInstruction)
	}

	return instructions, nil
}

const BatchTransferUserLimit = 10

func separateInstructions(instructions []solana.Instruction) [][]solana.Instruction {
	separatedInstructions := make([][]solana.Instruction, 0, len(instructions)/BatchTransferUserLimit+1)

	for i := 0; i < len(instructions); i += BatchTransferUserLimit {
		sliceEnd := min(i+BatchTransferUserLimit, len(instructions))
		separatedInstructions = append(separatedInstructions, instructions[i:sliceEnd])
	}

	return separatedInstructions
}

func (s *WalletService) WithdrawSolanaRoom(
	ctx context.Context,
	duel *model.Duel,
	tokenInfo *model.DuelTokenInfo,
) (string, error) {
	reqBody := map[string]any{
		"pda_nr":     duel.RoomNumber,
		"token_mint": tokenInfo.Mint,
	}

	endpoint := "withdraw"
	if duel.CreatedAt.After(s.newContractStartDate) {
		endpoint = "withdraw-new"
	}

	signature, err := s.WalletAuthorityClient.GetSignatureFromContract(
		reqBody,
		endpoint,
	)
	if err != nil {
		return "", err
	}

	return signature, nil
}

// Close for the duels with spl tokens
func (s *WalletService) CloseSolanaSPLTokensRoom(
	ctx context.Context,
	duel *model.Duel,
	tokenInfo *model.DuelTokenInfo,
) (string, error) {
	adminTokenAccount, err := s.findATA(s.solanaAdminPrivateKey.PublicKey(), tokenInfo.Mint, tokenInfo.ProgramID)
	if err != nil || adminTokenAccount == ZeroValuePublicKey {
		return "", apperrors.Internal("failed to find associated token account for user rewarding", err)
	}

	reqBody := map[string]any{
		"pda_nr":     duel.RoomNumber,
		"token_mint": tokenInfo.Mint,
	}

	endpoint := "close-room"
	if duel.CreatedAt.After(s.newContractStartDate) {
		endpoint = "close-room-new"
	}

	signature, err := s.WalletAuthorityClient.GetSignatureFromContract(
		reqBody,
		endpoint,
	)
	if err != nil {
		return "", err
	}

	return signature, nil
}

// Close for the duels with SOL
func (s *WalletService) CloseSolanaRoom(
	ctx context.Context,
	duel *model.Duel,
) (string, error) {
	reqBody := map[string]any{
		"pda_nr": duel.RoomNumber,
	}

	signature, err := s.WalletAuthorityClient.GetSignatureFromContract(
		reqBody,
		"close-room-sol",
	)
	if err != nil {
		return "", err
	}

	return signature, nil
}

func findProgramAddress(seed1, seed2 []byte, programID solana.PublicKey) (solana.PublicKey, uint8) {
	seed := [][]byte{seed1, seed2}
	pda, bump, err := solana.FindProgramAddress(seed, programID)
	if err != nil {
		return solana.PublicKey{}, 0
	}

	return pda, bump
}

func uint32ToBytesLE(value uint32) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, value)

	return buf.Bytes()
}

func (s *WalletService) GetPDAByRoomNumber(num uint32) (solana.PublicKey, error) {
	programID, err := solana.PublicKeyFromBase58(s.contractAddress)
	if err != nil {
		return solana.PublicKey{}, apperrors.Internal("failed to get contract public address", err)
	}

	seed1 := []byte("data")
	seed2 := uint32ToBytesLE(num)

	pda, _ := findProgramAddress(seed1, seed2, programID)

	return pda, nil
}

type CompiledInstruction struct {
	programID solana.PublicKey
	accounts  []*solana.AccountMeta
	data      []byte
}

func decompileInstruction(instruction solana.CompiledInstruction, tx *solana.Transaction) (*CompiledInstruction, error) {
	programID, err := tx.ResolveProgramIDIndex(instruction.ProgramIDIndex)
	if err != nil {
		return nil, fmt.Errorf("resolve program ID: %w", err)
	}

	accounts, err := instruction.ResolveInstructionAccounts(&tx.Message)
	if err != nil {
		return nil, fmt.Errorf("resolve instruction account: %w", err)
	}

	return &CompiledInstruction{
		programID: programID,
		accounts:  accounts,
		data:      instruction.Data,
	}, nil
}

func (c *CompiledInstruction) ProgramID() solana.PublicKey {
	return c.programID
}

func (c *CompiledInstruction) Accounts() []*solana.AccountMeta {
	return c.accounts
}

func (c *CompiledInstruction) Data() ([]byte, error) {
	return c.data, nil
}

var ComputeBudgetProgramID = solana.MustPublicKeyFromBase58("ComputeBudget111111111111111111111111111111")

func RemoveComputeBudgetInstructionsFromTx(tx *solana.Transaction) ([]solana.Instruction, error) {
	instructions := make([]solana.Instruction, 0, len(tx.Message.Instructions))

	for _, instruction := range tx.Message.Instructions {
		instr, err := decompileInstruction(instruction, tx)
		if err != nil {
			return nil, apperrors.Internal("failed to decompile instruction", err)
		}

		if instr.ProgramID() == ComputeBudgetProgramID {
			continue
		}

		instructions = append(instructions, instr)
	}

	return instructions, nil
}

// outdated prob
func (s *WalletService) GetInstructionsFromContractService(
	ctx context.Context,
	reqBody map[string]any,
	endpoint string,
	signer solana.PrivateKey,
) ([]solana.Instruction, error) {
	tx, err := s.WalletAuthorityClient.GetTxFromContract(reqBody, endpoint)
	if err != nil {
		return nil, err
	}

	recentBlockHashResp, err := s.SolanaRPC.GetLatestBlockhash(ctx, Finalized)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to get latest block hash", err)
	}

	tx.Message.RecentBlockhash = recentBlockHashResp.Value.Blockhash

	_, err = tx.Sign(txSignerPrivateKeyGetter(signer))
	if err != nil {
		return nil, apperrors.Internal("failed to sign a transaction", err)
	}

	computeUnits, err := s.GetSimulationComputeUnits(ctx, tx)
	if err != nil {
		return nil, err
	}

	computeUnits = uint32(float64(computeUnits) * CUExtraCapacityCoefficient)
	cuPriceInstruction, err := computebudget.NewSetComputeUnitPriceInstructionBuilder().
		SetMicroLamports(s.PriorityTracker.GetMediumPriorityMicroLamports()).
		ValidateAndBuild()
	if err != nil {
		return nil, apperrors.Internal("failed to set transaction compute unit price", err)
	}

	cuLimitInstruction, err := computebudget.NewSetComputeUnitLimitInstructionBuilder().
		SetUnits(computeUnits).
		ValidateAndBuild()
	if err != nil {
		return nil, apperrors.Internal("failed to set transaction compute unit limit", err)
	}

	instructions := make([]solana.Instruction, 0, 3)
	instructions = append(instructions, cuPriceInstruction, cuLimitInstruction)

	for _, instruction := range tx.Message.Instructions {
		inst, err := decompileInstruction(instruction, tx)
		if err != nil {
			return nil, apperrors.ServiceUnavailable("failed to decompile instruction", err)
		}

		instructions = append(instructions, inst)
	}

	return instructions, nil
}

func GetTxInstructions(txs ...*solana.Transaction) ([]solana.Instruction, error) {
	instructionsCount := 0
	for _, tx := range txs {
		if tx == nil {
			continue
		}

		instructionsCount += len(tx.Message.Instructions)
	}

	instructions := make([]solana.Instruction, 0, instructionsCount)

	for _, tx := range txs {
		if tx == nil {
			continue
		}

		for _, instruction := range tx.Message.Instructions {
			inst, err := decompileInstruction(instruction, tx)
			if err != nil {
				return nil, apperrors.ServiceUnavailable("failed to decompile instruction", err)
			}

			instructions = append(instructions, inst)
		}
	}

	return instructions, nil
}

func (s *WalletService) buildRewardInstructionsForRecipient(
	ctx context.Context,
	userID uuid.UUID,
	sender solana.PrivateKey,
	senderTokenAccount solana.PublicKey,
	recipient solana.PublicKey,
	grossRewardAmount uint64,
	tokenInfo *model.DuelTokenInfo,
) ([]solana.Instruction, error) {
	var instructions = make([]solana.Instruction, 0, 4)

	recipientTokenAccount, err := s.findATA(recipient, tokenInfo.Mint, tokenInfo.ProgramID)
	if err != nil {
		return nil, apperrors.Internal("failed to find recipient ATA", err)
	}

	info, err := s.GetAccountInfo(ctx, recipientTokenAccount)
	if err != nil && !errors.Is(err, rpc.ErrNotFound) {
		return nil, apperrors.Internal("failed to get ATA info", err)
	}

	ataExists := info != nil && info.Value != nil && info.Value.Owner != (solana.PublicKey{})

	solBalance, err := s.getSolBalance(ctx, recipient, Finalized)
	if err != nil {
		return nil, apperrors.Internal("failed to get recipient SOL balance", err)
	}

	// SOL to send & ATA creation cost
	recipientTopUpLamports, ataInitLamports, err := s.calculateRecipientCostsLamports(ctx, ataExists, solBalance)
	if err != nil {
		return nil, apperrors.Internal("failed to calculate recipient costs", err)
	}

	// total cost to subtract from reward
	totalChargeLamports := recipientTopUpLamports + ataInitLamports
	netRewardAmount := grossRewardAmount

	if totalChargeLamports > 0 {
		feeInRewardToken, err := s.convertLamportsToRewardTokenFee(ctx, totalChargeLamports, tokenInfo)
		if err != nil {
			return nil, apperrors.Internal("failed to convert lamports to reward token fee", err)
		}

		// reward too small to cover costs
		if feeInRewardToken >= grossRewardAmount {
			return nil, apperrors.BadRequest("reward amount is too small to cover activation cost", nil)
		}

		netRewardAmount -= feeInRewardToken
	}

	if recipientTopUpLamports > 0 {
		// send SOL to recipient
		solTransferInstr, err := system.NewTransferInstruction(
			recipientTopUpLamports,
			sender.PublicKey(),
			recipient,
		).ValidateAndBuild()
		if err != nil {
			return nil, apperrors.Internal("failed to build SOL transfer instruction", err)
		}

		instructions = append(instructions, solTransferInstr)
	}

	mint, err := solana.PublicKeyFromBase58(tokenInfo.Mint)
	if err != nil {
		return nil, apperrors.Internal("failed to get token mint", err)
	}

	if !ataExists {
		// create ATA (paid by sender)
		createATAInstr, err := associatedtokenaccount.NewCreateInstruction(
			sender.PublicKey(),
			recipient,
			mint,
		).ValidateAndBuild()
		if err != nil {
			return nil, apperrors.Internal("failed to build ATA creation instruction", err)
		}

		instructions = append(instructions, createATAInstr)
	}

	// transfer SPL tokens
	transferInstruction, err := token.NewTransferInstruction(
		netRewardAmount,
		senderTokenAccount,
		recipientTokenAccount,
		sender.PublicKey(),
		[]solana.PublicKey{sender.PublicKey()},
	).ValidateAndBuild()
	if err != nil {
		return nil, apperrors.Internal("failed to build SPL transfer instruction", err)
	}

	instructions = append(instructions, transferInstruction)

	return instructions, nil
}

func (s *WalletService) convertLamportsToRewardTokenFee(
	ctx context.Context,
	lamports uint64,
	tokenInfo *model.DuelTokenInfo,
) (uint64, error) {
	if tokenInfo.USDPrice <= 0 {
		return 0, apperrors.BadRequest("token USD price must be greater than zero", nil)
	}

	solPrice, err := s.CoinService.TokenPriceRepository.GetLastPrice(ctx, solana.SolMint.String())
	if err != nil {
		return 0, apperrors.Internal("failed to get last solana price", err)
	}
	if solPrice == nil || solPrice.UsdPrice <= 0 {
		return 0, apperrors.NotFound("solana price not found", nil)
	}

	solPriceDecimal := decimal.NewFromFloat(solPrice.UsdPrice)
	tokenPriceDecimal := decimal.NewFromFloat(tokenInfo.USDPrice)

	// lamports -> SOL
	solAmount := decimal.NewFromInt(int64(lamports)).
		Div(decimal.NewFromInt(1_000_000_000))

	// SOL -> USD
	usdAmount := solAmount.Mul(solPriceDecimal)

	// round up
	tokenDivider := decimal.NewFromInt(1).Shift(int32(tokenInfo.Decimals))
	feeInToken := usdAmount.Div(tokenPriceDecimal)
	feeInTokenSmallestUnit := feeInToken.Mul(tokenDivider).Ceil()

	return feeInTokenSmallestUnit.BigInt().Uint64(), nil
}

func (s *WalletService) calculateRecipientCostsLamports(
	ctx context.Context,
	ataExists bool,
	solBalance uint64,
) (recipientTopUpLamports uint64, ataInitLamports uint64, err error) {
	const minOperationalLamports uint64 = 1_000_000 // 0.001 SOL
	const txFeeBufferLamports uint64 = 10_000

	if !ataExists {
		rentLamports, err := s.getTokenAccountRentExemption(ctx)
		if err != nil {
			return 0, 0, err
		}

		// ATA rent + small fee buffer
		ataInitLamports = rentLamports + txFeeBufferLamports
	}

	if solBalance < minOperationalLamports {
		// missing SOL for recipient
		recipientTopUpLamports = minOperationalLamports - solBalance
	}

	return recipientTopUpLamports, ataInitLamports, nil
}

func (s *WalletService) getTokenAccountRentExemption(ctx context.Context) (uint64, error) {
	const tokenAccountSize = 165

	lamports, err := s.SolanaRPC.GetMinimumBalanceForRentExemption(
		ctx,
		tokenAccountSize,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return 0, err
	}

	return lamports, nil
}

// SC - Smart Contranct
func (s *WalletService) validateCreateCryptoDuelSCTransaction(
	ctx context.Context,
	duel *model.CreateDuelReq,
	txHash string,
	tokenInfo *model.DuelTokenInfo,
) (uint64, string, error) {

	sig, err := solana.SignatureFromBase58(txHash)
	if err != nil {
		return 0, "", apperrors.BadRequest("failed to parse tx hash")
	}

	sent, err := s.SigTracker.SubscribeForSignatureStatus(sig, TxConfirmationTimeout)
	if err != nil && !errors.Is(err, rpc.ErrNotConfirmed) {
		return 0, "", apperrors.Internal("subscribe to signature status", err)
	}

	if !sent {
		return 0, "", apperrors.Internal("tx was not confirmed: " + sig.String())
	}

	tx, err := s.SolanaRPC.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return 0, "", apperrors.Internal("failed to get transaction by tx hash", err)
	}

	logs := strings.Join(tx.Meta.LogMessages, ", ")

	if !strings.Contains(logs, "Instruction: Init") {
		return 0, "", apperrors.BadRequest("invalid instruction")
	}
	if !strings.Contains(logs, fmt.Sprintf("Theme: %s", model.DuelTopicPlaceholder)) {
		return 0, "", apperrors.BadRequest("invalid theme")
	}
	if !strings.Contains(logs, fmt.Sprintf("Description: %s", duel.Question)) {
		return 0, "", apperrors.BadRequest("invalid description")
	}

	if !strings.Contains(logs, fmt.Sprintf("Bet: %d", int64(duel.DuelPrice*tokenInfo.CryptoDuelPriceMultiplier()))) {
		return 0, "", apperrors.BadRequest("invalid bet")
	}
	if !strings.Contains(logs, fmt.Sprintf("Current executing program address: %s", s.contractAddress)) {
		return 0, "", apperrors.BadRequest("invalid program address")
	}

	var (
		joinedRoomRegex *regexp.Regexp
		initedRoomRegex *regexp.Regexp
	)

	if tokenInfo.Symbol == model.SOLSymbol {
		joinedRoomRegex = model.JoinedSOLRoomRegex
		initedRoomRegex = model.RoomSOLAccountRegex
	} else {
		joinedRoomRegex = model.JoinedSPLRoomRegex
		initedRoomRegex = model.InitedRoomRegex
	}

	roomNumberStr, found := parseTxLogs(joinedRoomRegex, logs)
	if !found {
		return 0, "", apperrors.BadRequest("room number not found")
	}
	roomNumber, err := strconv.ParseUint(roomNumberStr, 10, 64)
	if err != nil {
		return 0, "", err
	}

	roomTokenPda, found := parseTxLogs(initedRoomRegex, logs)
	if !found {
		return 0, "", apperrors.BadRequest("room token pda not found")
	}
	return roomNumber, roomTokenPda, nil
}

func parseTxLogs(regExp *regexp.Regexp, log string) (string, bool) {
	m := regExp.FindStringSubmatch(log)
	if len(m) < 2 {
		return "", false
	}
	return m[1], true
}

func parseTxLogsAll(regExp *regexp.Regexp, log string) []string {
	matches := regExp.FindAllStringSubmatch(log, -1)
	if len(matches) == 0 {
		return nil
	}

	out := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		out = append(out, match[1])
	}

	return out
}

func parseRoomTokenPDAByJoinedRoom(logLines []string) map[uint64]string {
	roomToPDA := make(map[uint64]string)
	pendingInitPDAs := make([]string, 0)

	for _, line := range logLines {
		if pda, ok := parseTxLogs(model.InitedRoomRegex, line); ok {
			pendingInitPDAs = append(pendingInitPDAs, pda)
			continue
		}

		roomStr, ok := parseTxLogs(model.JoinedSPLRoomRegex, line)
		if !ok {
			continue
		}

		room, err := strconv.ParseUint(roomStr, 10, 64)
		if err != nil {
			continue
		}

		if len(pendingInitPDAs) == 0 {
			continue
		}

		roomToPDA[room] = pendingInitPDAs[0]
		pendingInitPDAs = pendingInitPDAs[1:]
	}

	return roomToPDA
}

// SC - Smart Contranct
func (s *WalletService) validateJoinCryptoDuelSCTransaction(
	ctx context.Context,
	txHash string,
	duel *model.Duel,
) (string, error) {

	sig, err := solana.SignatureFromBase58(txHash)
	if err != nil {
		return "", apperrors.BadRequest("failed to parse tx hash")
	}

	sent, err := s.SigTracker.SubscribeForSignatureStatus(sig, TxConfirmationTimeout)
	if err != nil && !errors.Is(err, rpc.ErrNotConfirmed) {
		return "", apperrors.Internal("subscribe to signature status", err)
	}

	if !sent {
		return "", apperrors.Internal("tx was not confirmed: " + sig.String())
	}

	tx, err := s.SolanaRPC.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return "", apperrors.Internal("failed to get transaction by tx hash", err)
	}

	logs := strings.Join(tx.Meta.LogMessages, ", ")

	if !strings.Contains(logs, "Instruction: Join") {
		return "", apperrors.BadRequest("invalid instruction")
	}
	if !strings.Contains(logs, fmt.Sprintf("Current executing program address: %s", s.contractAddress)) {
		return "", apperrors.BadRequest("invalid program address")
	}

	roomNumberStr, found := parseTxLogs(model.JoinedSPLRoomRegex, logs)
	if !found {
		return "", apperrors.BadRequest("room number not found")
	}
	roomNumber, err := strconv.ParseUint(roomNumberStr, 10, 64)
	if err != nil {
		return "", apperrors.BadRequest("invalid room number format")
	}
	if roomNumber != duel.RoomNumber {
		return "", apperrors.BadRequest("room number mismatch")
	}

	var roomTokenPda string
	if strings.Contains(logs, "Instruction: Init") {
		roomTokenPda, found = parseTxLogs(model.InitedRoomRegex, logs)
		if !found {
			return "", apperrors.BadRequest("room token pda not found")
		}
	}

	return roomTokenPda, nil
}

func (s *WalletService) validateJoinCryptoDuelSOLTransaction(
	ctx context.Context,
	txHash string,
	sender string,
	duel *model.Duel,
) (string, error) {
	sig, err := solana.SignatureFromBase58(txHash)
	if err != nil {
		return "", apperrors.BadRequest("failed to parse tx hash")
	}

	sent, err := s.SigTracker.SubscribeForSignatureStatus(sig, TxConfirmationTimeout)
	if err != nil && !errors.Is(err, rpc.ErrNotConfirmed) {
		return "", apperrors.Internal("subscribe to signature status", err)
	}

	if !sent {
		return "", apperrors.Internal("tx was not confirmed: " + sig.String())
	}

	tx, err := s.SolanaRPC.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return "", apperrors.Internal("failed to get transaction by tx hash", err)
	}

	exists, err := s.TxRepository.Exists(ctx, &model.TransactionType{Signature: sig.String()})
	if err != nil {
		return "", apperrors.Internal("failed to check if transaction already exists", err)
	}
	if exists {
		return "", apperrors.BadRequest("transaction already exists")
	}

	logs := strings.Join(tx.Meta.LogMessages, ", ")

	if !strings.Contains(logs, "Instruction: JoinSol") {
		return "", apperrors.BadRequest("invalid instruction")
	}
	if !strings.Contains(logs, fmt.Sprintf("Current executing program address: %s", s.contractAddress)) {
		return "", apperrors.BadRequest("invalid program address")
	}

	senderAddress, err := solana.PublicKeyFromBase58(sender)
	if err != nil {
		return "", apperrors.BadRequest("invalid sender address")
	}

	if !strings.Contains(logs, fmt.Sprintf("Your address: %s", senderAddress.String())) {
		return "", apperrors.BadRequest("sender address mismatch")
	}

	roomNumberStr, found := parseTxLogs(model.JoinedSOLRoomRegex, logs)
	if !found {
		return "", apperrors.BadRequest("room number not found")
	}
	roomNumber, err := strconv.ParseUint(roomNumberStr, 10, 64)
	if err != nil {
		return "", apperrors.BadRequest("invalid room number format")
	}
	if roomNumber != duel.RoomNumber {
		return "", apperrors.BadRequest("room number mismatch")
	}

	_, found = parseTxLogs(model.RoomSOLAccountRegex, logs)
	if !found {
		return "", apperrors.BadRequest("room SOL account not found")
	}

	return "", nil
}
