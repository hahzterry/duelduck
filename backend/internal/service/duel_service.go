package service

import (
	"context"
	"fmt"
	"time"

	"dd-prediction-api/internal/storage/click"

	"github.com/gagliardetto/solana-go"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"go.uber.org/zap"

	"dd-prediction-api/config"
	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/storage/repository"
	"dd-prediction-api/pkg/apperrors"
	"dd-prediction-api/pkg/mtype"
	repo "dd-prediction-api/pkg/repository"
)

type ITokenPriceRepository interface {
	GetLastPrice(ctx context.Context, token string) (*model.TokenPrice, error)
	GetLastPricesBySymbol(ctx context.Context, symbols []string) (map[string]float64, error)
}

type DuelService struct {
	WalletService                *WalletService
	CoinService                  ICoinService
	UserRepository               *repository.UserRepository
	TxRepository                 *repository.TransactionRepository
	DuelRepository               *repository.DuelRepository
	PlayerRepository             *repository.PlayerRepository
	ProjectRepository            *repository.ProjectRepository
	TransactionManager           *repo.TransactionManager
	TokenPriceRepository         ITokenPriceRepository
	DuelTxRepository             click.IDuelTransactionRepository
	ProjectCommissionAccrualRepo *repository.ProjectCommissionAccrualRepository
	HTTPShareClient              *resty.Client
	ShareImageAPI                string
	ContractAddress              string
	PublicDomain                 string
}

func NewDuelService(
	c *config.Config,
	walletService *WalletService,
	coinService *CoinService,
	userRepository *repository.UserRepository,
	txRepository *repository.TransactionRepository,
	duelRepository *repository.DuelRepository,
	playerRepository *repository.PlayerRepository,
	projectRepository *repository.ProjectRepository,
	transactionManager *repo.TransactionManager,
	tokenPriceRepository *click.TokenPriceRepository,
	duelTxRepository click.IDuelTransactionRepository,
	projectCommissionAccrualRepo *repository.ProjectCommissionAccrualRepository,
) (*DuelService, error) {
	return &DuelService{
		WalletService:                walletService,
		CoinService:                  coinService,
		UserRepository:               userRepository,
		TxRepository:                 txRepository,
		DuelRepository:               duelRepository,
		PlayerRepository:             playerRepository,
		ProjectRepository:            projectRepository,
		TransactionManager:           transactionManager,
		TokenPriceRepository:         tokenPriceRepository,
		DuelTxRepository:             duelTxRepository,
		ProjectCommissionAccrualRepo: projectCommissionAccrualRepo,
		ContractAddress:              c.App.ContractAddress,
		HTTPShareClient:              resty.New(),
		ShareImageAPI:                c.App.ShareImageAPI,
		PublicDomain:                 c.HTTP.PublicDomain,
	}, nil
}

// logTxAsync persists duel transactions to ClickHouse asynchronously
func (s *DuelService) logTxAsync(txs []model.DuelTransaction) {
	if len(txs) == 0 {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.DuelTxRepository.BulkInsert(ctx, txs); err != nil {
			zap.L().Error("failed to persist duel transactions to clickhouse", zap.Error(err))
		}
	}()
}

func (s *DuelService) CreateNewCryptoDuelAdmin(
	ctx context.Context,
	userID uuid.UUID,
	req *model.CreateDuelReq,
) (*model.Duel, error) {
	if _, err := s.validatePaymentType(req); err != nil {
		return nil, err
	}

	user, err := s.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to find user", err)
	}

	duel := model.DuelByCreateReqAdmin(req, user)
	s.tryUpdateDuelUSDPrice(ctx, duel, false)

	if err = s.DuelRepository.Create(ctx, duel); err != nil {
		return nil, apperrors.Internal("failed to create crypto duel", err)
	}

	if err := s.sendDuelShareImageReq(duel); err != nil {
		zap.L().Error("failed on duel share image request", zap.Error(err))
	}

	return duel, nil
}

func (s *DuelService) SignCreateCryptoDuelTransaction(
	ctx context.Context,
	userID uuid.UUID,
	req *model.CreateDuelReq,
) (string, error) {
	tokenInfo, err := s.validatePaymentType(req)
	if err != nil {
		return "", err
	}

	user, err := s.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return "", apperrors.Internal("failed to get user", err)
	}

	if err := s.validateProjectDuelPermissions(ctx, user, req); err != nil {
		return "", err
	}

	if req.SourceOfTruth == "" {
		return "", apperrors.BadRequest("no source of truth is specified")
	}

	duel := model.DuelByCreateReq(req, user)

	tx, _, err := s.WalletService.InitAndJoinSolanaRoom(ctx, duel, user, req.Answer, tokenInfo, duel.DuelPrice)
	if err != nil {
		return "", err
	}

	return tx, nil
}

func (s *DuelService) CreateCryptoDuel(
	ctx context.Context,
	userID uuid.UUID,
	req *model.CreateCryptoDuelReq,
) (*model.CreateCryptoDuelResp, error) {
	tokenInfo, err := s.getDuelTokenInfoBySymbol(ctx, req.CreateDuelReq.Symbol)
	if err != nil {
		return nil, err
	}

	roomNumber, roomTokenPda, err := s.WalletService.
		validateCreateCryptoDuelSCTransaction(
			ctx,
			&req.CreateDuelReq,
			req.Hash,
			tokenInfo,
		)
	if err != nil {
		zap.L().Warn("transaction validation error", zap.Error(err))
		return nil, apperrors.BadRequest("transaction validation failed")
	}

	user, err := s.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to get user", err)
	}

	if err := s.validateProjectDuelPermissions(ctx, user, &req.CreateDuelReq); err != nil {
		return nil, err
	}

	if req.CreateDuelReq.SourceOfTruth == "" {
		return nil, apperrors.BadRequest("no source of truth is specified")
	}

	duel := model.DuelByCreateReq(&req.CreateDuelReq, user)

	duel.RoomNumber = roomNumber
	duel.RoomTokenPDA = roomTokenPda

	if err = s.createAndJoinCryptoDuel(ctx, duel, user, req.CreateDuelReq.Answer, req.Hash); err != nil {
		return nil, err
	}

	if err := s.sendDuelShareImageReq(duel); err != nil {
		zap.L().Error("failed on duel share image request", zap.Error(err))
	}

	return &model.CreateCryptoDuelResp{Duel: duel}, nil
}

func (s *DuelService) createAndJoinCryptoDuel(
	ctx context.Context,
	duel *model.Duel,
	user *model.User,
	duelOwnerAnswer uint8,
	txHash string,
) error {
	join := &model.JoinDuelReq{
		DuelID:    duel.ID,
		Answer:    duelOwnerAnswer,
		PaidPrice: &duel.DuelPrice,
	}

	s.tryUpdateDuelUSDPrice(ctx, duel, false)

	err := s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {
			err := s.DuelRepository.WithTx(tx).Create(ctx, duel)
			if err != nil {
				return apperrors.Internal("failed to create duel", err)
			}

			_, err = s.DuelRepository.WithTx(tx).JoinDuel(ctx, user.ID, join, duel)
			if err != nil {
				return apperrors.Internal("failed to join duel owner to the duel", err)
			}

			if txHash != "" {
				txRecord := &model.TransactionType{Signature: txHash, TxType: model.TransactionTypeDuelPrediction}

				if err = s.TxRepository.WithTx(tx).Create(ctx, txRecord); err != nil {
					return apperrors.Internal("failed to join duel owner to the duel", err)
				}
			}

			return nil
		})
	if err != nil {
		return err
	}

	if txHash != "" {
		s.logTxAsync([]model.DuelTransaction{
			model.NewDuelTransaction(txHash, model.TransactionTypeDuelPrediction, user.ID, duel.ID, duel.DuelPrice),
		})
	}

	return nil
}

func (s *DuelService) SignJoinCryptoDuelTransaction(
	ctx context.Context,
	userID uuid.UUID,
	req *model.JoinDuelReq,
) (*model.TxHashResp, error) {
	user, err := s.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to get user", err)
	}

	duel, err := s.DuelRepository.GetByID(ctx, req.DuelID)
	if err != nil {
		return nil, apperrors.Internal("failed to get duel", err)
	}

	tokenInfo, err := s.getDuelTokenInfoBySymbol(context.Background(), duel.Symbol)
	if err != nil {
		return nil, apperrors.Internal("failed to find solana token by symbol", err)
	}

	if err = s.isAbleToJoinDuel(ctx, duel, userID); err != nil {
		return nil, err
	}

	var paidPrice float64
	if duel.PriceType == model.DuelPriceTypeRange {
		if req.PaidPrice == nil {
			return nil, apperrors.BadRequest("paid_price is required for range type duels")
		}
		if duel.MinPrice != nil && *req.PaidPrice < *duel.MinPrice {
			return nil, apperrors.BadRequest("paid_price is below min_price")
		}
		if duel.MaxPrice != nil && *req.PaidPrice > *duel.MaxPrice {
			return nil, apperrors.BadRequest("paid_price exceeds max_price")
		}
		paidPrice = *req.PaidPrice
	} else {
		paidPrice = duel.DuelPrice
	}

	var tx string
	if duel.PlayersCount == 0 {
		tx, _, err = s.WalletService.InitAndJoinSolanaRoom(
			ctx,
			duel,
			user,
			req.Answer,
			tokenInfo,
			paidPrice,
		)
	} else {
		tx, err = s.WalletService.JoinSolanaRoom(
			ctx,
			duel,
			user,
			req.Answer,
			tokenInfo,
			paidPrice,
		)
	}
	if err != nil {
		return nil, err
	}

	return &model.TxHashResp{TxHash: tx}, nil
}

func (s *DuelService) JoinCryptoDuel(
	ctx context.Context,
	userID uuid.UUID,
	req *model.JoinCryptoDuelReq,
) (*model.JoinCryptoDuelResp, error) {

	user, err := s.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to get user", err)
	}

	duel, err := s.DuelRepository.GetByID(ctx, req.JoinDuelReq.DuelID)
	if err != nil {
		return nil, apperrors.Internal("failed to get duel", err)
	}

	if err = s.isAbleToJoinDuel(ctx, duel, userID); err != nil {
		return nil, err
	}

	tokenInfo, err := s.getDuelTokenInfoBySymbol(ctx, duel.Symbol)
	if err != nil {
		return nil, apperrors.Internal("failed to find solana token by symbol", err)
	}

	if duel.PriceType == model.DuelPriceTypeRange {
		if req.JoinDuelReq.PaidPrice == nil {
			return nil, apperrors.BadRequest("paid_price is required for range type duels")
		}
		if duel.MinPrice != nil && *req.JoinDuelReq.PaidPrice < *duel.MinPrice {
			return nil, apperrors.BadRequest("paid_price is below min_price")
		}
		if duel.MaxPrice != nil && *req.JoinDuelReq.PaidPrice > *duel.MaxPrice {
			return nil, apperrors.BadRequest("paid_price exceeds max_price")
		}
	} else {
		req.JoinDuelReq.PaidPrice = &duel.DuelPrice
	}

	var roomTokenPda string
	if duel.Symbol != model.SOLSymbol {
		roomTokenPda, err = s.WalletService.
			validateJoinCryptoDuelSCTransaction(
				ctx,
				req.Hash,
				duel,
				*req.JoinDuelReq.PaidPrice,
				tokenInfo,
			)
	} else {
		// SOL transfer has to be validated separately
		// cause its have another type of transfer instructions
		roomTokenPda, err = s.WalletService.
			validateJoinCryptoDuelSOLTransaction(
				ctx,
				req.Hash,
				user.WalletAddress,
				duel,
				*req.JoinDuelReq.PaidPrice,
				tokenInfo,
			)

	}
	if err != nil {
		zap.L().Warn("transaction validation error", zap.Error(err))
		return nil, apperrors.BadRequest("transaction validation failed")
	}

	var player *model.Player
	err = s.TransactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		if roomTokenPda != "" {
			err = s.DuelRepository.WithTx(tx).SetRoomTokenPDA(ctx, duel.ID, roomTokenPda)
			if err != nil {
				return apperrors.Internal("failed to set room token pda", err)
			}
		}

		player, err = s.DuelRepository.WithTx(tx).JoinDuel(ctx, userID, &req.JoinDuelReq, duel)
		if err != nil {
			return apperrors.Internal("failed to join duel", err)
		}

		txRecord := &model.TransactionType{Signature: req.Hash, TxType: model.TransactionTypeDuelPrediction}
		if err = s.TxRepository.WithTx(tx).Create(ctx, txRecord); err != nil {
			return apperrors.Internal("failed to create transaction record", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logTxAsync([]model.DuelTransaction{
		model.NewDuelTransaction(req.Hash, model.TransactionTypeDuelPrediction, userID, req.JoinDuelReq.DuelID, duel.DuelPrice),
	})

	return &model.JoinCryptoDuelResp{Player: player}, nil
}

func (s *DuelService) ResolveCryptoDuelByOwner(
	ctx context.Context,
	ownerID uuid.UUID,
	req *model.DuelResolveReq,
) (*model.ResolveCryptoDuelResp, error) {
	duel, err := s.DuelRepository.GetByID(ctx, req.DuelID)
	if err != nil {
		return nil, apperrors.Internal("failed to get duel by id", err)
	}

	err = s.validateResolveDuelByOwner(ownerID, duel)
	if err != nil {
		return nil, err
	}

	duel.ResolvedBy = ownerID
	params := &model.DuelResolveParams{
		DuelID:        duel.ID,
		Answer:        req.Answer,
		JoinNotBefore: duel.Deadline,
	}
	txHashes, err := s.resolveCryptoDuel(ctx, duel, params)
	if err != nil {
		return nil, err
	}

	// Fetch one more time with all updated duel states
	duel, err = s.DuelRepository.GetByID(ctx, req.DuelID)
	if err != nil {
		return nil, apperrors.Internal("failed to get duel by id", err)
	}

	return &model.ResolveCryptoDuelResp{
		TxHashes: txHashes,
		Duel:     duel,
	}, nil
}

func (s *DuelService) ResolveCryptoDuelByAdmin(
	ctx context.Context,
	moderatorID uuid.UUID,
	req *model.DuelResolveReq,
) (*model.ResolveCryptoDuelResp, error) {
	duel, err := s.DuelRepository.GetByID(ctx, req.DuelID)
	if err != nil {
		return nil, apperrors.Internal("failed to get duel by id", err)
	}

	if duel.Status != model.DuelStatusActive && duel.Status != model.DuelStatusWaitingForResolve {
		return nil, apperrors.BadRequest("resolve is not possible from current status")
	}

	if time.Now().UTC().Before(duel.Deadline) {
		return nil, apperrors.BadRequest("cannot resolve duel before deadline")
	}

	if duel.IsOwnerResolving &&
		time.Now().UTC().Before(duel.CreatedAt.Add(30*24*time.Hour)) {
		return nil, apperrors.BadRequest("cannot resolve owner resolving duel now")
	}

	params := &model.DuelResolveParams{
		DuelID:        duel.ID,
		Answer:        req.Answer,
		JoinNotBefore: duel.Deadline,
	}
	txHashes, err := s.resolveCryptoDuelByAdmin(ctx, moderatorID, duel, params)
	if err != nil {
		return nil, err
	}

	return &model.ResolveCryptoDuelResp{
		TxHashes: txHashes,
	}, nil
}

func (s *DuelService) resolveCryptoDuelByAdmin(
	ctx context.Context,
	moderatorID uuid.UUID,
	duel *model.Duel,
	params *model.DuelResolveParams,
) ([]string, error) {
	duel.ResolvedBy = moderatorID

	txHashes, err := s.resolveCryptoDuel(ctx, duel, params)
	if err != nil {
		return nil, err
	}

	return txHashes, nil
}

func (s *DuelService) resolveCryptoDuel(
	ctx context.Context,
	duel *model.Duel,
	req *model.DuelResolveParams,
) ([]string, error) {
	duelWinners, err := s.PlayerRepository.GetDuelWinners(ctx, duel.ID, req.Answer, req.JoinNotBefore)
	if err != nil {
		return nil, apperrors.Internal("failed to count players with specific answer", err)
	}
	allDuelWinnersCount := uint64(len(duelWinners))

	userIDsToRefund, err := s.PlayerRepository.FindDuelPlayersUserIDsToRefund(ctx, duel.ID, req.JoinNotBefore)
	if err != nil {
		return nil, apperrors.Internal("failed to count players that must be refunded", err)
	}

	now := time.Now().UTC()
	duel.ResolvedAt = &now

	duelPlayersCount, err := s.PlayerRepository.CountDuelPlayers(ctx, duel.ID)
	if err != nil {
		return nil, apperrors.Internal("failed to count duel players", err)
	}

	playersPool := duelPlayersCount - uint64(len(userIDsToRefund))
	if playersPool == 0 || playersPool == allDuelWinnersCount || allDuelWinnersCount == 0 {
		txHashes, err := s.cancelCryptoDuel(ctx, duel, model.AutoCancelReq(duel))
		if err != nil {
			return nil, err
		}

		return txHashes, nil
	}

	var (
		roomClosingTxHash     string
		roomWithdrawingTxHash string
		commissionAmountRaw   uint64
		commissionAmountUSD   float64
	)

	tokenInfo, err := s.getDuelTokenInfoBySymbol(context.Background(), duel.Symbol)
	if err != nil {
		return nil, err
	}

	playersCount := duelPlayersCount - duel.RefundedPlayersCount
	duelParams := model.NewDuelParams(
		duel.DuelPrice, duel.CommissionRate,
		playersCount, allDuelWinnersCount,
	)

	if duel.Symbol == model.SOLSymbol {
		roomClosingTxHash, err = s.WalletService.CloseSolanaRoom(ctx, duel)
		if err != nil {
			zap.L().Error("failed to close solana room after refund",
				zap.Error(err),
				zap.String("duel_id", duel.ID.String()),
				zap.Uint64("room_number", duel.RoomNumber))
			return nil, apperrors.Internal("failed to close solana room after resolve", err)
		}
	} else {
		if duel.RoomTokenPDA != "" {
			roomWithdrawingTxHash, err = s.WalletService.WithdrawSolanaRoom(ctx, duel, tokenInfo)
			if err != nil {
				zap.L().Error("failed to withdraw solana room",
					zap.Error(err),
					zap.String("duel_id", duel.ID.String()),
					zap.Uint64("room_number", duel.RoomNumber))
				return nil, apperrors.Internal("failed to withdraw solana room", err)
			}

			// delay betweetn withdrawing and closing rooms
			time.Sleep(30 * time.Second)
		}

		roomClosingTxHash, err = s.WalletService.CloseSolanaSPLTokensRoom(ctx, duel, tokenInfo)
		if err != nil {
			zap.L().Error("failed to close solana room after resolve",
				zap.Error(err),
				zap.String("duel_id", duel.ID.String()),
				zap.Uint64("room_number", duel.RoomNumber))
			return nil, apperrors.Internal("failed to close solana room after resolve", err)
		}
	}

	txRecords := make([]model.TransactionType, 0)

	var refundedPlayersTxHashes []string
	if len(userIDsToRefund) > 0 {
		refundedPlayersTxHashes, err = s.partialCryptoRefund(ctx, duel, req.JoinNotBefore)
		if err != nil {
			return nil, err
		}
	}

	unpaidWinners, err := s.PlayerRepository.GetCryptoDuelWinners(ctx, duel.ID, req.Answer)
	if err != nil {
		return nil, apperrors.Internal("failed to count players with specific answer", err)
	}

	if len(unpaidWinners) == 0 {
		return []string{}, nil
	}

	var (
		duelRewardTxHashes []string
		priceMultiplier    = tokenInfo.CryptoDuelPriceMultiplier()
		winAmount          uint64
		playersWinAmount   float64
	)

	if duel.PriceType == model.DuelPriceTypeRange {
		totalPool, err := s.PlayerRepository.SumActivePaidPrice(ctx, duel.ID, duel.DuelPrice)
		if err != nil {
			return nil, apperrors.Internal("failed to sum active paid prices", err)
		}

		commissionFraction := float64(duel.CommissionRate) / 100.0
		netPool := totalPool * (1.0 - commissionFraction)
		commissionAmountRaw = uint64(totalPool * commissionFraction * priceMultiplier)

		var winnersPool float64
		for _, w := range unpaidWinners {
			if w.PaidPrice != nil {
				winnersPool += *w.PaidPrice
			} else {
				winnersPool += duel.DuelPrice
			}
		}

		for i, w := range unpaidWinners {
			paid := duel.DuelPrice
			if w.PaidPrice != nil {
				paid = *w.PaidPrice
			}
			unpaidWinners[i].Amount = uint64(netPool * (paid / winnersPool) * priceMultiplier)
			duelWinners[i].WinAmount = float64(unpaidWinners[i].Amount) / priceMultiplier
			duelWinners[i].IsWinner = true
			duelWinners[i].FinalStatus = model.PlayerStatusResolved
		}

		duelRewardTxHashes, err = s.WalletService.RewardDuelWinnersWithAmounts(ctx, unpaidWinners, tokenInfo, duel.ProjectID)
		if err != nil {
			return nil, err
		}
	} else {
		winAmount = duelParams.CalculateFinalCryptoReward(priceMultiplier)
		playersWinAmount = float64(winAmount) / priceMultiplier

		duelRewardTxHashes, err = s.WalletService.RewardDuelWinners(ctx, winAmount, unpaidWinners, tokenInfo, duel.ProjectID)
		if err != nil {
			return nil, err
		}

		commissionAmountRaw = duelParams.CalculateCryptoCommissionReward(priceMultiplier)
	}

	txRecords = append(
		txRecords,
		model.NewTransactionsWithSameType(
			model.TransactionTypeDuelReward,
			duelRewardTxHashes...,
		)...,
	)
	if len(refundedPlayersTxHashes) > 0 {
		txRecords = append(
			txRecords,
			model.NewTransactionsWithSameType(
				model.TransactionTypeDuelRefund,
				refundedPlayersTxHashes...,
			)...,
		)
	}

	s.tryUpdateDuelUSDPrice(ctx, duel, true)

	if commissionAmountRaw > 0 {
		project, err := s.ProjectRepository.GetByID(ctx, duel.ProjectID)
		if err != nil {
			return nil, apperrors.Internal("failed to get project for commission", err)
		}

		if project.IsUsersDuelsEnabled {
			commissionAmountRaw = uint64(commissionAmountRaw / 2)
			commissionAmountUSD = float64(commissionAmountRaw) / priceMultiplier * tokenInfo.USDPrice

			owner, err := s.UserRepository.GetByID(ctx, duel.OwnerID)
			if err != nil {
				return nil, apperrors.Internal("failed to get duel owner for commission", err)
			}

			recipientAddress, err := solana.PublicKeyFromBase58(owner.WalletAddress)
			if err != nil || recipientAddress == ZeroValuePublicKey {
				return nil, apperrors.BadRequest("recipient is not valid solana address", err)
			}

			if _, err = s.WalletService.TransferSymbol(ctx, recipientAddress, commissionAmountRaw, tokenInfo, duel.ProjectID); err != nil {
				return nil, err
			}
		} else {
			commissionAmountUSD = float64(commissionAmountRaw) / priceMultiplier * tokenInfo.USDPrice
		}
	}

	duel.Status = model.DuelStatusResolved
	duel.FinalResult = &req.Answer
	duel.WinnersCount = allDuelWinnersCount
	duel.Commission = commissionAmountRaw

	winnerIDs := make([]uuid.UUID, 0, len(duelWinners))
	for i := range duelWinners {
		winnerIDs = append(winnerIDs, duelWinners[i].UserID)
	}

	err = s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {
			if duel.PriceType == model.DuelPriceTypeRange {
				err = s.PlayerRepository.WithTx(tx).UpdateDuelWinnersVariableAmounts(ctx, duelWinners)
			} else {
				err = s.PlayerRepository.WithTx(tx).UpdateDuelWinners(ctx, duelWinners, playersWinAmount)
			}
			if err != nil {
				return err
			}

			err = s.PlayerRepository.WithTx(tx).SetStatusToActiveByDuelID(ctx, duel.ID, model.PlayerStatusResolved)
			if err != nil {
				return err
			}

			if err = s.DuelRepository.WithTx(tx).Update(ctx, duel); err != nil {
				return err
			}

			if err = s.TxRepository.WithTx(tx).BulkInsert(ctx, txRecords); err != nil {
				return err
			}

			if commissionAmountRaw > 0 {
				now := time.Now().UTC()
				projectAccrual := &model.ProjectCommissionAccrual{
					ProjectID:     duel.ProjectID,
					DuelID:        duel.ID,
					Symbol:        duel.Symbol,
					CommissionRaw: commissionAmountRaw,
					CommissionUSD: commissionAmountUSD,
					BillingMonth:  model.BillingMonthStart(now),
					AccruedAt:     now,
				}
				if err = s.ProjectCommissionAccrualRepo.WithTx(tx).Create(ctx, projectAccrual); err != nil {
					return err
				}
			}

			return nil
		})
	if err != nil {
		return nil, apperrors.Internal("failed to resolve a duel", err)
	}

	rewardTxs := make([]model.DuelTransaction, 0, len(unpaidWinners))
	for _, w := range unpaidWinners {
		amount := playersWinAmount
		if duel.PriceType == model.DuelPriceTypeRange {
			amount = float64(w.Amount) / priceMultiplier
		}
		rewardTxs = append(rewardTxs, model.NewDuelTransaction("", model.TransactionTypeDuelReward, w.UserID, duel.ID, amount))
	}
	s.logTxAsync(rewardTxs)

	allTxHashes := append(duelRewardTxHashes, roomWithdrawingTxHash, roomClosingTxHash)
	allTxHashes = append(allTxHashes, refundedPlayersTxHashes...)

	return allTxHashes, nil
}

func (s *DuelService) ApproveCryptoDuel(ctx context.Context,
	approvedBy uuid.UUID,
	req *model.DuelApproveReq,
) (*model.JoinCryptoDuelResp, error) {
	duel, err := s.DuelRepository.GetByID(ctx, req.DuelID)
	if err != nil {
		return nil, apperrors.Internal("failed to get duel by id", err)
	}

	if duel.Status != model.DuelStatusActive {
		return nil, apperrors.BadRequest("approve is not possible from current status")
	}

	user, err := s.UserRepository.GetByID(ctx, duel.OwnerID)
	if err != nil {
		return nil, apperrors.Internal("failed to get user by id", err)
	}

	player, err := s.PlayerRepository.GetByUserID(ctx, user.ID, duel.ID)
	if err != nil {
		return nil, apperrors.Internal("failed to get player by id", err)
	}

	duel.ApprovedBy = approvedBy
	err = s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {
			return s.DuelRepository.WithTx(tx).Update(ctx, duel)
		})
	if err != nil {
		return nil, apperrors.Internal("failed to update a duel", err)
	}

	return &model.JoinCryptoDuelResp{
		Player: player,
	}, nil
}

func (s *DuelService) CancelCryptoDuelByAdmin(ctx context.Context,
	resolvedBy uuid.UUID,
	req *model.DuelCancelReq,
) (*model.CancelCryptoDuelResp, error) {
	duel, err := s.DuelRepository.GetByID(ctx, req.DuelID)
	if err != nil {
		return nil, apperrors.Internal("failed to get duel by id", err)
	}

	if err = validateDuelCancel(duel, req); err != nil {
		return nil, err
	}

	txHashes, err := s.cancelCryptoDuelByAdmin(ctx, resolvedBy, duel, req)
	if err != nil {
		return nil, err
	}

	err = s.TxRepository.BulkInsertWithSameTxType(ctx, model.TransactionTypeDuelRefund, txHashes)
	if err != nil {
		return nil, apperrors.Internal("failed to create transaction records", err)
	}

	return &model.CancelCryptoDuelResp{TxHashes: txHashes}, nil
}

func (s *DuelService) cancelCryptoDuelByAdmin(ctx context.Context,
	resolvedBy uuid.UUID,
	duel *model.Duel,
	req *model.DuelCancelReq,
) ([]string, error) {
	txSignatures, err := s.cancelCryptoDuel(ctx, duel, req)
	if err != nil {
		return nil, err
	}

	return txSignatures, nil
}

func (s *DuelService) cancelCryptoDuel(ctx context.Context,
	duel *model.Duel,
	req *model.DuelCancelReq,
) ([]string, error) {
	var (
		txHashes              = make([]string, 0)
		roomClosingTxHash     = ""
		roomWithdrawingTxHash = ""
		cancelPlayers         []model.CryptoDuelPlayer
	)

	duelPlayersCount, err := s.PlayerRepository.CountDuelPlayers(ctx, duel.ID)
	if err != nil {
		return nil, apperrors.Internal("failed to count duel players", err)
	}

	// If duel is in review, and it should be cancelled,
	// we don't have to return any money, since we didn't charge anything from user
	// charge users before approve | users can join duels in-review
	if s.hasChargedDuelPriceFromUser(duelPlayersCount, duel.Status, req.Status) {
		cancelPlayers, err = s.PlayerRepository.GetCryptoDuelPlayers(ctx, duel.ID)
		players := cancelPlayers
		if err != nil {
			return nil, apperrors.Internal("failed to get duel players", err)
		}

		tokenInfo, err := s.getDuelTokenInfoBySymbol(context.Background(), duel.Symbol)
		if err != nil {
			return nil, apperrors.Internal("failed to find solana token by symbol", err)
		}

		priceMultiplier := tokenInfo.CryptoDuelPriceMultiplier()
		var duelPrice = uint64(duel.DuelPrice * priceMultiplier)

		if duel.Symbol == model.SOLSymbol {
			roomClosingTxHash, err = s.WalletService.CloseSolanaRoom(ctx, duel)
			if err != nil {
				zap.L().Error("failed to close solana room after refund",
					zap.Error(err),
					zap.String("duel_id", duel.ID.String()),
					zap.Uint64("room_number", duel.RoomNumber))
			}
		} else {
			if duel.RoomTokenPDA != "" {
				roomWithdrawingTxHash, err = s.WalletService.WithdrawSolanaRoom(ctx, duel, tokenInfo)
				if err != nil {
					zap.L().Error("failed to withdraw solana room",
						zap.Error(err),
						zap.String("duel_id", duel.ID.String()),
						zap.Uint64("room_number", duel.RoomNumber))
					return nil, apperrors.Internal("failed to withdraw solana room", err)
				}

				// delay betweetn withdrawing and closing rooms
				time.Sleep(30 * time.Second)
			}

			roomClosingTxHash, err = s.WalletService.CloseSolanaSPLTokensRoom(ctx, duel, tokenInfo)
			if err != nil {
				zap.L().Error("failed to close solana room after refund",
					zap.Error(err),
					zap.String("duel_id", duel.ID.String()),
					zap.Uint64("room_number", duel.RoomNumber))
				return nil, apperrors.Internal("failed to close solana room after resolve", err)
			}
		}

		if duel.PriceType == model.DuelPriceTypeRange {
			for i, p := range cancelPlayers {
				if p.PaidPrice != nil {
					cancelPlayers[i].Amount = uint64(*p.PaidPrice * priceMultiplier)
				} else {
					cancelPlayers[i].Amount = duelPrice
				}
			}
			txHashes, err = s.WalletService.TransferBulkSolanaChain(ctx, 0, players, tokenInfo, duel.ProjectID)
		} else {
			txHashes, err = s.WalletService.TransferBulkSolanaChain(ctx, duelPrice, players, tokenInfo, duel.ProjectID)
		}
		if err != nil {
			return nil, apperrors.ServiceUnavailable("failed to refund duel: "+duel.ID.String(), err)
		}
	}

	duel.Status = req.Status
	duel.CancellationReason = req.CancellationReason

	err = s.TransactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		err := s.DuelRepository.WithTx(tx).Update(ctx, duel)
		if err != nil {
			return apperrors.Internal("failed to update duel status", err)
		}

		err = s.PlayerRepository.WithTx(tx).SetStatusToAll(ctx, duel.ID, model.PlayerStatusRefunded)
		if err != nil {
			return apperrors.Internal("failed to update crypto players status", err)
		}

		if len(txHashes) > 0 {
			err = s.TxRepository.WithTx(tx).BulkInsertWithSameTxType(
				ctx,
				model.TransactionTypeDuelRefund,
				txHashes)
			if err != nil {
				return apperrors.Internal("failed to create transaction records", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(cancelPlayers) > 0 {
		refundTxs := make([]model.DuelTransaction, 0, len(cancelPlayers))
		for _, p := range cancelPlayers {
			refundAmount := duel.DuelPrice
			if duel.PriceType == model.DuelPriceTypeRange && p.PaidPrice != nil {
				refundAmount = *p.PaidPrice
			}
			refundTxs = append(refundTxs, model.NewDuelTransaction("", model.TransactionTypeDuelRefund, p.UserID, duel.ID, refundAmount))
		}
		s.logTxAsync(refundTxs)
	}

	return append(txHashes, roomWithdrawingTxHash, roomClosingTxHash), nil
}

func (s *DuelService) partialCryptoRefund(
	ctx context.Context,
	duel *model.Duel,
	votedAfter time.Time,
) ([]string, error) {
	if duel.Status != model.DuelStatusActive {
		return []string{}, nil
	}

	players, err := s.PlayerRepository.GetDuelPlayersToRefund(ctx, duel.ID, votedAfter)
	if err != nil {
		return nil, apperrors.Internal("failed to get crypto duel players", err)
	}

	tokenInfo, err := s.getDuelTokenInfoBySymbol(context.Background(), duel.Symbol)
	if err != nil {
		return nil, apperrors.Internal("failed to find solana token by symbol", err)
	}

	var (
		txHashes        []string
		priceMultiplier = tokenInfo.CryptoDuelPriceMultiplier()
		duelPrice       = uint64(duel.DuelPrice * priceMultiplier)
	)

	if duel.PriceType == model.DuelPriceTypeRange {
		for i, p := range players {
			if p.PaidPrice != nil {
				players[i].Amount = uint64(*p.PaidPrice * priceMultiplier)
			} else {
				players[i].Amount = duelPrice
			}
		}
		txHashes, err = s.WalletService.TransferBulkSolanaChain(ctx, 0, players, tokenInfo, duel.ProjectID)
	} else {
		txHashes, err = s.WalletService.TransferBulkSolanaChain(ctx, duelPrice, players, tokenInfo, duel.ProjectID)
	}
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to refund duel: "+duel.ID.String(), err)
	}

	err = s.TransactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		err = s.PlayerRepository.WithTx(tx).SetStatus(ctx, players, model.PlayerStatusRefunded)
		if err != nil {
			return apperrors.Internal("failed to update crypto players status", err)
		}

		duel.RefundedPlayersCount += uint64(len(players))
		err = s.DuelRepository.WithTx(tx).SetRefundedPlayersCount(ctx, duel.ID, duel.RefundedPlayersCount)
		if err != nil {
			return apperrors.Internal("failed to update duel", err)
		}

		err = s.TxRepository.WithTx(tx).BulkInsertWithSameTxType(
			ctx,
			model.TransactionTypeDuelRefund,
			txHashes)
		if err != nil {
			return apperrors.Internal("failed to create transaction records", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(players) > 0 {
		refundTxs := make([]model.DuelTransaction, 0, len(players))
		for _, p := range players {
			refundAmount := duel.DuelPrice
			if duel.PriceType == model.DuelPriceTypeRange && p.PaidPrice != nil {
				refundAmount = *p.PaidPrice
			}
			refundTxs = append(refundTxs, model.NewDuelTransaction("", model.TransactionTypeDuelRefund, p.UserID, duel.ID, refundAmount))
		}
		s.logTxAsync(refundTxs)
	}

	return txHashes, nil
}

func (s *DuelService) hasChargedDuelPriceFromUser(duelPlayersCount uint64, duelOldStatus, duelNewStatus uint8) bool {
	if duelOldStatus != model.DuelStatusActive {
		return false
	}
	if duelNewStatus != model.DuelStatusRefunded && duelNewStatus != model.DuelStatusDenied {
		return false
	}
	return duelPlayersCount > 0
}

func validateDuelCancel(duel *model.Duel, req *model.DuelCancelReq) error {
	if req.Status != model.DuelStatusDenied && req.Status != model.DuelStatusRefunded {
		return apperrors.BadRequest("cancel is not possible with provided status")
	}

	if req.Status == model.DuelStatusDenied && duel.Status != model.DuelStatusActive {
		return apperrors.BadRequest("cancel is not possible from current duel status")
	}

	if req.Status == model.DuelStatusRefunded &&
		duel.Status != model.DuelStatusActive &&
		duel.Status != model.DuelStatusWaitingForResolve {
		return apperrors.BadRequest("refund is not possible from current duel status")
	}

	return nil
}

func (s *DuelService) getDuelTokenInfoBySymbol(
	ctx context.Context,
	symbol string,
) (*model.DuelTokenInfo, error) {
	solanaToken, err := s.CoinService.findSolanaTokenBySymbol(ctx, symbol)
	if err != nil {
		return nil, err
	}

	if solanaToken == nil {
		return nil, apperrors.NotFound("solana token not found")
	}

	return &model.DuelTokenInfo{
		Mint:          solanaToken.Mint,
		Symbol:        solanaToken.Symbol,
		Decimals:      solanaToken.Decimals,
		USDPrice:      solanaToken.USDPrice,
		ProgramID:     solanaToken.ProgramID,
		TokenImageUrl: solanaToken.ImageURL,
	}, nil
}

func (s *DuelService) isAbleToJoinDuel(ctx context.Context, duel *model.Duel, userID uuid.UUID) error {
	if duel.Deadline.Before(time.Now()) {
		return apperrors.BadRequest("join is not possible since duel's deadline has passed")
	}

	if duel.Status != model.DuelStatusActive {
		return apperrors.BadRequest("failed to join duel")
	}

	isPlayer, err := s.PlayerRepository.UserAlreadyParticipant(ctx, userID, duel.ID)
	if err != nil {
		return apperrors.Internal("failed to check if user is already participating in duel", err)
	}

	if isPlayer {
		return apperrors.BadRequest("user is already participating in this duel")
	}

	return nil
}

func (s *DuelService) GetDuelByIDUnauthorized(ctx context.Context, duelID uuid.UUID) (*model.DuelShow, error) {
	duel, err := s.DuelRepository.GetDuelShowByID(ctx, uuid.Nil, duelID)
	if err != nil {
		if repo.IsErrNoRows(err) {
			return nil, apperrors.NotFound("duel not found")
		}
		return nil, apperrors.Internal("failed to get duel", err)
	}

	if duel.Status == model.DuelStatusDenied {
		return nil, apperrors.Gone("duel has been removed")
	}

	return duel, nil
}

func (s *DuelService) GetDuelBySlugUnauthorized(ctx context.Context, slug string) (*model.DuelShow, error) {
	duel, err := s.DuelRepository.GetDuelShowBySlug(ctx, uuid.Nil, slug)
	if err != nil {
		if repo.IsErrNoRows(err) {
			return nil, apperrors.NotFound("duel not found")
		}
		return nil, apperrors.Internal("failed to get duel", err)
	}

	if duel.Status == model.DuelStatusDenied {
		return nil, apperrors.Gone("duel has been removed")
	}

	return duel, nil
}

func (s *DuelService) GetDuelByID(ctx context.Context, duelID uuid.UUID, userID uuid.UUID, role mtype.Role) (*model.DuelShow, []model.PlayerShow, error) {
	duel, err := s.DuelRepository.GetDuelShowByID(ctx, userID, duelID)
	if err != nil {
		return nil, nil, apperrors.Internal("failed to get duel", err)
	}

	fillPotentialWin(duel)

	players, err := s.PlayerRepository.GetAllPlayersByDuelID(ctx, duelID, nil)
	if err != nil {
		return nil, nil, apperrors.Internal("failed to get players", err)
	}

	return duel, players, nil
}

func (s *DuelService) GetDuelBySlug(ctx context.Context, slug string, userID uuid.UUID, role mtype.Role) (*model.DuelShow, []model.PlayerShow, error) {
	duel, err := s.DuelRepository.GetDuelShowBySlug(ctx, userID, slug)
	if err != nil {
		return nil, nil, apperrors.Internal("failed to get duel by slug", err)
	}

	fillPotentialWin(duel)

	players, err := s.PlayerRepository.GetAllPlayersByDuelID(ctx, duel.ID, nil)
	if err != nil {
		return nil, nil, apperrors.Internal("failed to get players", err)
	}

	return duel, players, nil
}

func (s *DuelService) GetAllDuelsUnauthorized(ctx context.Context, options *repo.Options) ([]model.DuelShow, error) {
	duels, err := s.DuelRepository.GetAllDuels(ctx, uuid.Nil, options)
	if err != nil {
		return nil, apperrors.Internal("failed to get duels", err)
	}

	return duels, nil
}

func (s *DuelService) GetAllDuels(ctx context.Context, userID uuid.UUID, options *repo.Options) ([]model.DuelShow, error) {
	duels, err := s.DuelRepository.GetAllDuels(ctx, userID, options)
	if err != nil {
		return nil, apperrors.Internal("failed to get duels", err)
	}

	for i := range duels {
		fillPotentialWin(&duels[i])
	}

	return duels, nil
}

func (s *DuelService) GetAllDuelsWhereParticipant(ctx context.Context, userID uuid.UUID, options *repo.Options) ([]model.DuelShow, error) {
	duels, err := s.DuelRepository.GetAllDuelsWhereParticipate(ctx, userID, options)
	if err != nil {
		return nil, apperrors.Internal("failed to get duels", err)
	}

	for i := range duels {
		fillPotentialWin(&duels[i])
	}

	return duels, nil
}

func (s *DuelService) GetFinancialHistory(
	ctx context.Context,
	userID uuid.UUID,
) ([]model.FinancialTxEntry, error) {
	txs, err := s.DuelTxRepository.GetByUserID(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to get financial history", err)
	}

	entries := make([]model.FinancialTxEntry, 0, len(txs))
	for _, tx := range txs {
		entries = append(entries, model.FinancialTxEntry{
			TxType: tx.TxType,
			TxHash: tx.Signature,
			Amount: tx.Amount,
			Date:   tx.CreatedAt,
		})
	}

	return entries, nil
}

func (s *DuelService) GetMyHistoryDuels(ctx context.Context, userID uuid.UUID, options *repo.Options) ([]model.DuelShow, error) {
	duels, err := s.DuelRepository.GetHistoryByUserID(ctx, userID, options)
	if err != nil {
		return nil, apperrors.Internal("failed to get duels history", err)
	}

	for i := range duels {
		fillPotentialWin(&duels[i])
	}

	return duels, nil
}

func (s *DuelService) GetMyDuels(ctx context.Context, userID uuid.UUID, options *repo.Options) ([]model.DuelShow, error) {
	duels, err := s.DuelRepository.GetUserDuels(ctx, userID, options)
	if err != nil {
		return nil, apperrors.Internal("failed to get my duels", err)
	}

	for i := range duels {
		fillPotentialWin(&duels[i])
	}

	return duels, nil
}

func (s *DuelService) EditDuel(ctx context.Context,
	req *model.DuelAdminEditReq,
) error {
	existing, err := s.DuelRepository.GetByID(ctx, req.ID)
	if err != nil {
		return apperrors.Internal("failed to get duel", err)
	}

	if existing.Status == model.DuelStatusResolved || existing.Status == model.DuelStatusRefunded {
		return apperrors.BadRequest("cannot edit a resolved or refunded duel")
	}

	duel := &model.Duel{
		ID:            req.ID,
		Question:      req.Question,
		SourceOfTruth: req.SourceOfTruth,
		Deadline:      req.Deadline,
		LogoURL:       req.LogoURL,
		DuelInfo:      req.DuelInfo,
	}

	err = s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {
			if err := s.DuelRepository.WithTx(tx).Update(ctx, duel); err != nil {
				return apperrors.Internal("failed to edit duel", err)
			}
			return nil
		})
	if err != nil {
		return err
	}

	return nil
}

func (s *DuelService) RefreshMaterializedView(ctx context.Context, viewName string) error {
	if err := s.DuelRepository.RefreshMaterializedView(ctx, viewName); err != nil {
		return apperrors.Internal("failed to refresh leaderboard", err)
	}

	return nil
}

func (s *DuelService) GetSelfResolveUnresolvedDuelsCount(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	count, err := s.DuelRepository.GetUnresolvedDuelsByUserID(ctx, userID)
	if err != nil {
		return 0, apperrors.Internal("failed to get self resolve unresolved duels by id", err)
	}

	return count, nil
}

func (s *DuelService) GetTransactionsBySignatures(
	ctx context.Context,
	signatures []string,
) ([]model.TransactionType, error) {
	transactions, err := s.TxRepository.GetTransactionsBySignatures(ctx, signatures)
	if err != nil {
		return nil, apperrors.Internal("failed to get transaction records", err)
	}

	return transactions, nil
}

func (s *DuelService) GetTokenAccounts(
	ctx context.Context,
	req *model.GetTokenAccountsReq,
) (*model.SolscanTokenAccountsResp, error) {
	txs, err := s.WalletService.GetTokenAccounts(ctx, req.Address, req)
	if err != nil {
		return nil, err
	}

	return txs, nil
}

func (s *DuelService) validateResolveDuelByOwner(ownerID uuid.UUID, duel *model.Duel) error {
	if !duel.IsOwnerResolving {
		return apperrors.BadRequest("this duel is owner resolving only")
	}

	if duel.OwnerID != ownerID {
		return apperrors.BadRequest("only the owner of the duel can resolve it")
	}

	if duel.Status != model.DuelStatusActive && duel.Status != model.DuelStatusWaitingForResolve {
		return apperrors.BadRequest("resolve is not possible from current status")
	}

	if time.Now().UTC().Before(duel.Deadline) {
		return apperrors.BadRequest("cannot resolve duel before deadline")
	}

	return nil
}

func (s *DuelService) GetActiveDuelsWagerVolumeLast24Hours(ctx context.Context) (*model.DuelActiveWagerVolume24h, error) {
	symbols, err := s.DuelRepository.GetActiveDuelSymbols(ctx)
	if err != nil {
		return nil, apperrors.Internal("failed to get active duel symbols", err)
	}
	if len(symbols) == 0 {
		return &model.DuelActiveWagerVolume24h{
			TotalUSDC: 0,
			BySymbol:  []model.DuelSymbolVolume24h{},
		}, nil
	}

	rows, err := s.DuelRepository.GetActiveDuelsNativeWagerVolumeLast24Hours(ctx, symbols)
	if err != nil {
		return nil, apperrors.Internal("failed to get native wager volume for active duels", err)
	}

	priceMap, err := s.TokenPriceRepository.GetLastPricesBySymbol(ctx, symbols)
	if err != nil {
		return nil, apperrors.Internal("failed to get last prices by symbol", err)
	}

	out := &model.DuelActiveWagerVolume24h{
		BySymbol: make([]model.DuelSymbolVolume24h, 0, len(rows)),
	}

	for _, r := range rows {
		price := priceMap[r.Symbol]
		if r.Symbol == model.USDCSymbol {
			price = 1
		}
		usdcVol := r.NativeVolume * price

		out.TotalUSDC += usdcVol
		out.BySymbol = append(out.BySymbol, model.DuelSymbolVolume24h{
			Symbol:       r.Symbol,
			NativeVolume: r.NativeVolume,
			USDCVolume:   usdcVol,
		})
	}

	return out, nil
}

func (s *DuelService) GetFeeIncomeLast24Hours(ctx context.Context) (*model.DuelFeeIncome24h, error) {
	out := &model.DuelFeeIncome24h{}

	symbols, err := s.DuelRepository.GetDuelSymbolsLast24Hours(ctx)
	if err != nil {
		return nil, apperrors.Internal("failed to get duel symbols for last 24 hours", err)
	}

	platformRows, err := s.DuelRepository.GetActiveDuelsNativePlatformFeeLast24Hours(ctx)
	if err != nil {
		return nil, apperrors.Internal("failed to get platform native fee for active duels", err)
	}

	platformPriceMap, err := s.TokenPriceRepository.GetLastPricesBySymbol(ctx, symbols)
	if err != nil {
		return nil, apperrors.Internal("failed to get last prices by symbol", err)
	}

	for _, r := range platformRows {
		price := platformPriceMap[r.Symbol]
		if r.Symbol == model.USDCSymbol {
			price = 1
		}
		out.PlatformFeeUSDC += r.NativeFee * price
	}

	creatorRows, err := s.DuelRepository.GetCreatorsCommissionNativeByResolvedDuelsLast24Hours(ctx)
	if err != nil {
		return nil, apperrors.Internal("failed to get creators commission earnings", err)
	}

	creatorSymbolsSet := make(map[string]struct{}, len(creatorRows))
	for _, r := range creatorRows {
		creatorSymbolsSet[r.Symbol] = struct{}{}
	}
	creatorSymbols := make([]string, 0, len(creatorSymbolsSet))
	for sym := range creatorSymbolsSet {
		creatorSymbols = append(creatorSymbols, sym)
	}

	creatorPriceMap, err := s.TokenPriceRepository.GetLastPricesBySymbol(ctx, creatorSymbols)
	if err != nil {
		return nil, apperrors.Internal("failed to get last prices by symbol", err)
	}

	earnedUsers := make(map[string]struct{}, len(creatorRows))
	for _, r := range creatorRows {
		price := creatorPriceMap[r.Symbol]
		if r.Symbol == model.USDCSymbol {
			price = 1
		}
		if r.NativeEarning > 0 {
			earnedUsers[r.OwnerID.String()] = struct{}{}
		}
		out.UsersFeeUSDC += r.NativeEarning * price
	}
	out.UsersWhoEarned = int64(len(earnedUsers))

	return out, nil
}

func (s *DuelService) GetUsersActivityLast24Hours(ctx context.Context) (*model.UsersActivity24h, error) {
	newUsers, err := s.UserRepository.CountNewUsersLast24Hours(ctx)
	if err != nil {
		return nil, apperrors.Internal("failed to count new users for last 24 hours", err)
	}

	players, err := s.PlayerRepository.CountDistinctPlayersLast24Hours(ctx)
	if err != nil {
		return nil, apperrors.Internal("failed to count distinct players for last 24 hours", err)
	}

	creators, err := s.DuelRepository.CountDistinctCreatorsLast24Hours(ctx)
	if err != nil {
		return nil, apperrors.Internal("failed to count distinct creators for last 24 hours", err)
	}

	return &model.UsersActivity24h{
		NewUsers: newUsers,
		Players:  players,
		Creators: creators,
	}, nil
}

func (s *DuelService) GetDuelParticipants(ctx context.Context, duelID uuid.UUID) ([]model.PlayerShow, error) {
	return s.PlayerRepository.GetAllPlayersByDuelID(ctx, duelID, &repo.Options{})
}

func fillPotentialWin(duel *model.DuelShow) {
	if !duel.Joined || duel.YourAnswer == nil {
		return
	}
	if duel.Status == model.DuelStatusResolved ||
		duel.Status == model.DuelStatusDenied ||
		duel.Status == model.DuelStatusRefunded {
		return
	}

	var votersOnMySide uint64
	if *duel.YourAnswer == 1 {
		votersOnMySide = duel.YesCount
	} else {
		votersOnMySide = duel.NoCount
	}
	if votersOnMySide == 0 {
		return
	}

	pool := float64(duel.PlayersCount) * duel.DuelPrice
	net := pool * (1.0 - float64(duel.CommissionRate)/100.0)
	v := net / float64(votersOnMySide)
	duel.PotentialWin = &v
}

func (s *DuelService) validateProjectDuelPermissions(ctx context.Context, user *model.User, req *model.CreateDuelReq) error {
	if user.Role.HasProjectAdminRights() {
		return nil
	}

	project, err := s.ProjectRepository.GetByID(ctx, user.ProjectID)
	if err != nil {
		return apperrors.Internal("failed to get project", err)
	}

	if !project.IsUsersDuelsEnabled {
		return apperrors.Forbidden("user duels are not enabled for this project")
	}

	if req.IsOwnerResolving && !project.IsSelfResolvedEnabled {
		return apperrors.Forbidden("self-resolve duels are not enabled for this project")
	}

	return nil
}

func (s *DuelService) validatePaymentType(req *model.CreateDuelReq) (*model.DuelTokenInfo, error) {
	if req.PriceType == "" {
		req.PriceType = model.DuelPriceTypeFixed
	}

	solanaToken, err := s.getTokenUSDPriceBySymbol(context.Background(), req.Symbol, false)
	if err != nil {
		return nil, err
	}

	minPrice := model.USDDuelMinJoinPrice / solanaToken.USDPrice
	maxPrice := model.USDDuelMaxJoinPrice / solanaToken.USDPrice

	switch req.PriceType {
	case model.DuelPriceTypeFixed:
		if req.DuelPrice < minPrice || req.DuelPrice > maxPrice {
			return nil, apperrors.BadRequest("invalid duel price")
		}
	case model.DuelPriceTypeRange:
		if req.MinPrice == nil || req.MaxPrice == nil {
			return nil, apperrors.BadRequest("min_price and max_price are required for range type")
		}
		if *req.MinPrice < minPrice || *req.MinPrice > maxPrice {
			return nil, apperrors.BadRequest("min_price is out of allowed range")
		}
		if *req.MaxPrice < minPrice || *req.MaxPrice > maxPrice {
			return nil, apperrors.BadRequest("max_price is out of allowed range")
		}
		if *req.MinPrice >= *req.MaxPrice {
			return nil, apperrors.BadRequest("min_price must be less than max_price")
		}
	default:
		return nil, apperrors.BadRequest("invalid price type")
	}

	if req.DuelInfo == nil {
		req.DuelInfo = make(map[string]any, 3)
	}

	req.DuelInfo["token_image_url"] = solanaToken.ImageURL
	req.DuelInfo["token_mint"] = solanaToken.Mint
	req.DuelInfo["token_is_verified"] = solanaToken.IsVerified

	return &model.DuelTokenInfo{
		Mint:      solanaToken.Mint,
		Symbol:    solanaToken.Symbol,
		Decimals:  solanaToken.Decimals,
		USDPrice:  solanaToken.USDPrice,
		ProgramID: solanaToken.ProgramID,
	}, nil
}

func (s *DuelService) getTokenUSDPriceBySymbol(
	ctx context.Context,
	symbol string,
	forceRefresh bool,
) (*model.SolanaToken, error) {
	solanaToken, err := s.CoinService.findSolanaTokenBySymbol(ctx, symbol)
	if err != nil {
		return nil, apperrors.Internal("failed to find solana token by symbol", err)
	}

	if solanaToken == nil {
		return nil, apperrors.NotFound("solana token not found")
	}

	if forceRefresh {
		resp, err := s.WalletService.Jupiter.Price(ctx, solanaToken.Mint)
		if err != nil {
			return nil, err
		}

		tokenInfo, ok := resp[solanaToken.Mint]
		if !ok {
			return nil, apperrors.ServiceUnavailable("failed to get token info from jupiter")
		}

		if tokenInfo.UsdPrice <= 0 {
			return nil, apperrors.ServiceUnavailable("invalid token usd price from jupiter")
		}

		solanaToken.USDPrice = tokenInfo.UsdPrice
	} else {
		tokenPrice, err := s.TokenPriceRepository.GetLastPrice(ctx, solanaToken.Mint)
		if err != nil {
			return nil, err
		}

		if tokenPrice != nil {
			solanaToken.USDPrice = tokenPrice.UsdPrice
		}
	}

	if solanaToken.USDPrice <= 0 {
		return nil, apperrors.ServiceUnavailable("invalid token usd price")
	}

	return solanaToken, nil
}

func (s *DuelService) tryUpdateDuelUSDPrice(
	ctx context.Context,
	duel *model.Duel,
	forceRefresh bool,
) {
	if err := s.updateDuelUSDPrice(ctx, duel, forceRefresh); err != nil {
		// Price conversion is analytics-only and should not block core duel flows.
		duel.USDPrice = nil
		zap.L().Warn(
			"failed to update duel usd price; continuing without usd_price",
			zap.String("duel_id", duel.ID.String()),
			zap.String("symbol", duel.Symbol),
			zap.Float64("duel_price", duel.DuelPrice),
			zap.Bool("force_refresh", forceRefresh),
			zap.Error(err),
		)
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}

func (s *DuelService) updateDuelUSDPrice(
	ctx context.Context,
	duel *model.Duel,
	forceRefresh bool,
) error {
	if duel.Symbol == model.USDCSymbol {
		duel.USDPrice = float64Ptr(duel.DuelPrice)
		return nil
	}

	solanaToken, err := s.getTokenUSDPriceBySymbol(ctx, duel.Symbol, forceRefresh)
	if err != nil {
		return err
	}

	usdPrice := duel.DuelPrice * solanaToken.USDPrice
	duel.USDPrice = float64Ptr(usdPrice)

	return nil
}

func (s *DuelService) sendDuelShareImageReq(duel *model.Duel) error {
	resp, err := s.HTTPShareClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(duel).
		Post(s.ShareImageAPI + "duel")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("status: %d, body: %s", resp.StatusCode(), resp.String())
	}

	return nil
}
