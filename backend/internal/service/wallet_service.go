package service

import (
	"context"
	"strings"

	"dd-prediction-api/config"
	"dd-prediction-api/internal/client/jupiter"
	"dd-prediction-api/internal/client/solscan"
	"dd-prediction-api/internal/client/walletauthority"
	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/storage/cypher"
	"dd-prediction-api/internal/storage/repository"
	"dd-prediction-api/pkg/apperrors"
	repo "dd-prediction-api/pkg/repository"
	"dd-prediction-api/pkg/sigtracker"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type IWalletService interface {
	InitAndJoinSolanaRoom(ctx context.Context, duel *model.Duel, user *model.User, answer uint8, tokenInfo *model.DuelTokenInfo, paidPrice float64) (string, string, error)
	JoinSolanaRoom(ctx context.Context, duel *model.Duel, user *model.User, answer uint8, tokenInfo *model.DuelTokenInfo, paidPrice float64) (string, error)
	RewardDuelWinners(ctx context.Context, winAmount uint64, winners []model.CryptoDuelPlayer, tokenInfo *model.DuelTokenInfo, projectID uuid.UUID) ([]string, error)
	RewardDuelOwnerWithCommission(ctx context.Context, publicAddress string, commissionReward uint64, tokenInfo *model.DuelTokenInfo) (string, error)
	SendProjectCommission(ctx context.Context, partnerWalletAddress string, amount uint64, tokenInfo *model.DuelTokenInfo) (string, error)
	TransferSymbol(ctx context.Context, recipientAddress solana.PublicKey, amount uint64, tokenInfo *model.DuelTokenInfo, projectID uuid.UUID) (string, error)
	SendTransaction(ctx context.Context, instructions []solana.Instruction, adminKey solana.PrivateKey) (string, error)
	TransferBulkSolanaChain(ctx context.Context, amount uint64, players []model.CryptoDuelPlayer, tokenInfo *model.DuelTokenInfo, projectID uuid.UUID) ([]string, error)
	WithdrawSolanaRoom(ctx context.Context, duel *model.Duel, tokenInfo *model.DuelTokenInfo) (string, error)
	CloseSolanaSPLTokensRoom(ctx context.Context, duel *model.Duel, tokenInfo *model.DuelTokenInfo) (string, error)
	CloseSolanaRoom(ctx context.Context, duel *model.Duel) (string, error)
	GetPDAByRoomNumber(num uint32) (solana.PublicKey, error)
	GetInstructionsFromContractService(ctx context.Context, reqBody map[string]any, endpoint string, signer solana.PrivateKey) ([]solana.Instruction, error)
	GetTransactions(ctx context.Context, publicAddress string, req *model.GetTransactionHistoryReq) ([]model.Transaction, error)
	GetTokenAccounts(ctx context.Context, publicAddress string, req *model.GetTokenAccountsReq) (*model.SolscanTokenAccountsResp, error)
	GetMintInfo(ctx context.Context, tokenKey solana.PublicKey) (*token.Mint, error)
	GetTokenBalance(ctx context.Context, ata solana.PublicKey) (uint64, error)
	HasEnoughTokenBalance(ctx context.Context, ata solana.PublicKey, requiredAmount float64) (bool, error)
	HasEnoughTokenBalanceFinalized(ctx context.Context, ata solana.PublicKey, requiredAmount uint64) (bool, error)
	GetSolBalance(ctx context.Context, pk solana.PublicKey) (uint64, error)
	HasEnoughSolBalance(ctx context.Context, pk solana.PublicKey, requiredAmount float64) (bool, error)
	RefreshBlockHash(ctx context.Context, txMessage *solana.Message) error
	GetSimulationComputeUnits(ctx context.Context, tx *solana.Transaction) (uint32, error)
	NewTransactionForSimulation(instructions []solana.Instruction, privateKeyGetter func(key solana.PublicKey) *solana.PrivateKey, opts ...solana.TransactionOption) (*solana.Transaction, error)
	GetTokenInfo(address string) (model.TokenInfo, error)
	EstimateTxFeeLamports(ctx context.Context, tx *solana.Transaction) (uint64, error)
	EnsureUserHasLamportsForTx(ctx context.Context, userAddress solana.PublicKey, rewardTx *solana.Transaction) (solana.Signature, error)
	GetAccountInfo(ctx context.Context, userTokenAccount solana.PublicKey) (out *rpc.GetAccountInfoResult, err error)
	InitializeATA(ctx context.Context, userID uuid.UUID, userAddress solana.PublicKey, mint solana.PublicKey) (string, error)
	SendTxWithTracker(ctx context.Context, tx *solana.Transaction, privateKeyGetter func(key solana.PublicKey) *solana.PrivateKey) (solana.Signature, error)
}

type WalletService struct {
	CoinService           *CoinService
	PriorityTracker       *PriorityTracker
	SigTracker            *sigtracker.TxTracker
	UserRepository        *repository.UserRepository
	TxRepository          *repository.TransactionRepository
	privateKeyRepository  *cypher.PrivateKeyRepository
	SolanaRPC             *rpc.Client
	Jupiter               *jupiter.Client
	Solscan               *solscan.Client
	HTTPClient            *resty.Client
	TransactionManager    *repo.TransactionManager
	WalletAuthorityClient *walletauthority.Client
	contractAddress       string
}

func NewWalletService(
	c *config.Config,
	coinService *CoinService,
	tracker *PriorityTracker,
	sigTracker *sigtracker.TxTracker,
	solanaRPC *rpc.Client,
	userRepository *repository.UserRepository,
	txRepository *repository.TransactionRepository,
	privateKeyRepository *cypher.PrivateKeyRepository,
	jupiter *jupiter.Client,
	solscan *solscan.Client,
	transactionManager *repo.TransactionManager,
	walletAuthority *walletauthority.Client,
) (*WalletService, error) {

	sigTracker.Start()

	w := &WalletService{
		CoinService:           coinService,
		PriorityTracker:       tracker,
		SigTracker:            sigTracker,
		UserRepository:        userRepository,
		TxRepository:          txRepository,
		privateKeyRepository:  privateKeyRepository,
		SolanaRPC:             solanaRPC,
		HTTPClient:            resty.New(),
		Jupiter:               jupiter,
		Solscan:               solscan,
		contractAddress:       c.App.ContractAddress,
		TransactionManager:    transactionManager,
		WalletAuthorityClient: walletAuthority,
	}

	return w, nil
}

func (s *WalletService) getProjectAdminKey(ctx context.Context, projectID uuid.UUID) (solana.PrivateKey, error) {
	encoded, err := s.privateKeyRepository.GetPrivateKeyBase58(ctx, projectID)
	if err != nil {
		return nil, apperrors.Internal("failed to get project admin key", err)
	}
	return solana.PrivateKeyFromBase58(encoded)
}

func (w *WalletService) SendTxWithTracker(
	ctx context.Context,
	tx *solana.Transaction,
	privateKeyGetter func(key solana.PublicKey) *solana.PrivateKey,
) (solana.Signature, error) {
	return w.sendTxWithTracker(ctx, tx, privateKeyGetter)
}

func (w *WalletService) GetAccountInfo(ctx context.Context, userTokenAccount solana.PublicKey) (out *rpc.GetAccountInfoResult, err error) {
	return w.SolanaRPC.GetAccountInfo(ctx, userTokenAccount)
}

func (s *WalletService) GetTokenAccounts(
	ctx context.Context,
	publicAddress string,
	req *model.GetTokenAccountsReq,
) (*model.SolscanTokenAccountsResp, error) {
	if publicAddress == "" {
		return nil, apperrors.BadRequest("public address is required")
	}
	if _, err := solana.PublicKeyFromBase58(publicAddress); err != nil {
		return nil, apperrors.BadRequest("invalid solana public address", err)
	}
	return s.Solscan.GetTokenAccounts(ctx, publicAddress, req)
}

func (s *WalletService) isAccountUninitialized(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "could not find account") ||
		strings.Contains(err.Error(), "AccountNotFound") ||
		strings.Contains(err.Error(), "AccountNotInitialized")
}
