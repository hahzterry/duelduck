package service

import (
	"context"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"

	"dd-prediction-api/config"
	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/storage/click"
	"dd-prediction-api/internal/storage/repository"
	"dd-prediction-api/pkg/apperrors"
)

type IPartnerAnalyticsRepository interface {
	GetDailyIncome(ctx context.Context, projectID uuid.UUID, from, to time.Time) ([]model.PartnerDailyIncome, error)
	GetMonthlyActiveUsers(ctx context.Context, projectID uuid.UUID) ([]model.PartnerMonthlyMAU, error)
	GetCurrentMonthVolume(ctx context.Context, projectID uuid.UUID) (float64, error)
}

type IProjectCommissionClaimRepository interface {
	Create(ctx context.Context, claim *model.ProjectCommissionClaim) error
	GetByProjectAndType(ctx context.Context, projectID uuid.UUID, claimType string) ([]model.ProjectCommissionClaim, error)
	GetByType(ctx context.Context, claimType string) ([]model.ProjectCommissionClaim, error)
}

type IProjectCommissionAccrualRepository interface {
	GetUnclaimed(ctx context.Context, projectID uuid.UUID, through time.Time) ([]model.ProjectCommissionAccrual, error)
	GetUnclaimedDD(ctx context.Context, through time.Time) ([]model.ProjectCommissionAccrual, error)
	GetSummariesByProject(ctx context.Context, projectID uuid.UUID) ([]model.ProjectCommissionSummary, error)
	GetMonthlyAggregates(ctx context.Context, projectID uuid.UUID) ([]repository.MonthlyAccrualAggregate, error)
}

type ICommissionDuelRepository interface {
	GetProjectDuelStats(ctx context.Context, projectID uuid.UUID) (*repository.ProjectDuelStats, error)
}

type ICommissionWalletService interface {
	TransferSymbol(ctx context.Context, recipientAddress solana.PublicKey, amount uint64, tokenInfo *model.DuelTokenInfo) (string, error)
}

type CommissionService struct {
	ProjectClaimRepo             IProjectCommissionClaimRepository
	ProjectCommissionAccrualRepo IProjectCommissionAccrualRepository
	ProjectRepository            repository.IProjectRepository
	DuelRepository               ICommissionDuelRepository
	WalletService                ICommissionWalletService
	CoinService                  ICoinService
	DuelTxRepository             click.IDuelTransactionRepository
	PartnerAnalytics             IPartnerAnalyticsRepository
	ddProfitWallet               string
}

func NewCommissionService(
	cfg *config.Config,
	projectClaimRepo *repository.ProjectCommissionClaimRepository,
	projectCommissionAccrualRepo *repository.ProjectCommissionAccrualRepository,
	projectRepository repository.IProjectRepository,
	duelRepository *repository.DuelRepository,
	walletService ICommissionWalletService,
	coinService ICoinService,
	duelTxRepository click.IDuelTransactionRepository,
	partnerAnalytics *click.PartnerAnalyticsRepository,
) *CommissionService {
	return &CommissionService{
		ProjectClaimRepo:             projectClaimRepo,
		ProjectCommissionAccrualRepo: projectCommissionAccrualRepo,
		ProjectRepository:            projectRepository,
		DuelRepository:               duelRepository,
		WalletService:                walletService,
		CoinService:                  coinService,
		DuelTxRepository:             duelTxRepository,
		PartnerAnalytics:             partnerAnalytics,
		ddProfitWallet:               cfg.App.DDProfitWalletAddress,
	}
}

// ClaimProjectCommission claims all unclaimed partner commissions up to the current claimable period.
// Only the net amount (after DD rate) is sent to the partner wallet.
func (s *CommissionService) ClaimProjectCommission(ctx context.Context, partnerID uuid.UUID, req *model.ProjectCommissionClaimReq) (*model.ProjectCommissionClaim, error) {
	now := time.Now().UTC()
	claimableThrough := model.ClaimableThrough(now)

	project, err := s.ProjectRepository.GetByPartnerID(ctx, partnerID)
	if err != nil {
		return nil, apperrors.Internal("failed to get project", err)
	}
	if project == nil {
		return nil, apperrors.NotFound("partner project not found")
	}

	recipientAddress, err := solana.PublicKeyFromBase58(req.WalletAddress)
	if err != nil || recipientAddress == ZeroValuePublicKey {
		return nil, apperrors.BadRequest("recipient is not valid solana address", err)
	}

	accruals, err := s.ProjectCommissionAccrualRepo.GetUnclaimed(ctx, project.ID, claimableThrough)
	if err != nil {
		return nil, apperrors.Internal("failed to get unclaimed project accruals", err)
	}
	if len(accruals) == 0 {
		return nil, apperrors.BadRequest("no pending project commissions to claim")
	}

	periodStart := earliestBillingMonth(accruals)

	var totalUSD float64
	for _, a := range accruals {
		totalUSD += a.CommissionUSD
	}

	rate := model.PlatformCommissionRate(totalUSD)
	netUSD := totalUSD * (1 - rate)

	grouped := groupBySymbol(accruals)
	txHashes := make(map[string]string, len(grouped))

	for _, g := range grouped {
		info, err := s.resolveTokenInfo(ctx, g.Symbol)
		if err != nil {
			return nil, err
		}

		netRaw := uint64(float64(g.AmountRaw) * (1 - rate))
		if netRaw == 0 {
			continue
		}

		txHash, err := s.WalletService.TransferSymbol(ctx, recipientAddress, netRaw, info)
		if err != nil {
			return nil, err
		}
		if txHash != "" {
			txHashes[g.Symbol] = txHash
		}
	}

	claim := &model.ProjectCommissionClaim{
		ID:            uuid.New(),
		Type:          model.ClaimTypePartner,
		ProjectID:     &project.ID,
		PeriodStart:   periodStart,
		PeriodEnd:     claimableThrough,
		GrossUSD:      totalUSD,
		PlatformRate:  rate,
		AmountUSD:     netUSD,
		WalletAddress: req.WalletAddress,
		TxHashes:      txHashes,
		ClaimedAt:     now,
	}

	if err = s.ProjectClaimRepo.Create(ctx, claim); err != nil {
		return nil, apperrors.Internal("failed to save project commission claim", err)
	}

	return claim, nil
}

// ClaimDDProfit claims the platform fee (DD profit) from all accruals not yet claimed.
// Sends to the wallet configured via DD_PROFIT_WALLET_ADDRESS env var.
func (s *CommissionService) ClaimDDProfit(ctx context.Context) (*model.ProjectCommissionClaim, error) {
	now := time.Now().UTC()
	claimableThrough := model.ClaimableThrough(now)

	accruals, err := s.ProjectCommissionAccrualRepo.GetUnclaimedDD(ctx, claimableThrough)
	if err != nil {
		return nil, apperrors.Internal("failed to get unclaimed DD profit accruals", err)
	}
	if len(accruals) == 0 {
		return nil, apperrors.BadRequest("no pending DD profit to claim")
	}

	periodStart := earliestBillingMonth(accruals)

	var totalGrossUSD float64
	for _, a := range accruals {
		totalGrossUSD += a.CommissionUSD
	}

	rate := model.PlatformCommissionRate(totalGrossUSD)
	feeUSD := totalGrossUSD * rate

	grouped := groupBySymbol(accruals)
	txHashes := make(map[string]string, len(grouped))

	recipientAddress, err := solana.PublicKeyFromBase58(s.ddProfitWallet)
	if err != nil || recipientAddress == ZeroValuePublicKey {
		return nil, apperrors.BadRequest("recipient is not valid solana address", err)
	}

	for _, g := range grouped {
		info, err := s.resolveTokenInfo(ctx, g.Symbol)
		if err != nil {
			return nil, err
		}

		feeRaw := uint64(float64(g.AmountRaw) * rate)
		if feeRaw == 0 {
			continue
		}

		txHash, err := s.WalletService.TransferSymbol(ctx, recipientAddress, feeRaw, info)
		if err != nil {
			return nil, err
		}
		if txHash != "" {
			txHashes[g.Symbol] = txHash
		}
	}

	claim := &model.ProjectCommissionClaim{
		ID:            uuid.New(),
		Type:          model.ClaimTypeDDProfit,
		ProjectID:     nil,
		PeriodStart:   periodStart,
		PeriodEnd:     claimableThrough,
		GrossUSD:      totalGrossUSD,
		PlatformRate:  rate,
		AmountUSD:     feeUSD,
		WalletAddress: s.ddProfitWallet,
		TxHashes:      txHashes,
		ClaimedAt:     now,
	}

	if err = s.ProjectClaimRepo.Create(ctx, claim); err != nil {
		return nil, apperrors.Internal("failed to save DD profit claim", err)
	}

	return claim, nil
}

func (s *CommissionService) GetProjectClaimHistory(ctx context.Context, partnerID uuid.UUID) ([]model.ProjectCommissionClaim, error) {
	project, err := s.ProjectRepository.GetByPartnerID(ctx, partnerID)
	if err != nil {
		return nil, apperrors.Internal("failed to get project", err)
	}
	if project == nil {
		return nil, apperrors.NotFound("partner project not found")
	}

	claims, err := s.ProjectClaimRepo.GetByProjectAndType(ctx, project.ID, model.ClaimTypePartner)
	if err != nil {
		return nil, apperrors.Internal("failed to get project claim history", err)
	}
	return claims, nil
}

func (s *CommissionService) GetDDClaimHistory(ctx context.Context) ([]model.ProjectCommissionClaim, error) {
	claims, err := s.ProjectClaimRepo.GetByType(ctx, model.ClaimTypeDDProfit)
	if err != nil {
		return nil, apperrors.Internal("failed to get DD claim history", err)
	}
	return claims, nil
}

func (s *CommissionService) GetPartnerDashboard(ctx context.Context, partnerID uuid.UUID) (*model.PartnerDashboard, error) {
	project, err := s.ProjectRepository.GetByPartnerID(ctx, partnerID)
	if err != nil {
		return nil, apperrors.Internal("failed to get project", err)
	}
	if project == nil {
		return nil, apperrors.NotFound("partner project not found")
	}

	stats, err := s.DuelRepository.GetProjectDuelStats(ctx, project.ID)
	if err != nil {
		return nil, apperrors.Internal("failed to get project duel stats", err)
	}

	summaries, err := s.ProjectCommissionAccrualRepo.GetSummariesByProject(ctx, project.ID)
	if err != nil {
		return nil, apperrors.Internal("failed to get project commission summaries", err)
	}

	monthlyVolume, err := s.PartnerAnalytics.GetCurrentMonthVolume(ctx, project.ID)
	if err != nil {
		monthlyVolume = 0
	}

	var totalGrossUSD float64
	netBySymbol := make([]model.SymbolNet, 0, len(summaries))
	for _, sum := range summaries {
		totalGrossUSD += sum.CommissionUSD
		netBySymbol = append(netBySymbol, model.SymbolNet{
			Symbol:   sum.Symbol,
			NetRaw:   sum.CommissionRaw - sum.PlatformFeeRaw,
			NetUSD:   sum.CommissionUSD - sum.PlatformFeeUSD,
			GrossUSD: sum.CommissionUSD,
			PlatRate: sum.PlatformRate,
		})
	}

	return &model.PartnerDashboard{
		APIKey:           project.APIKey,
		Status:           project.Status,
		DuelCount:        stats.DuelCount,
		TotalVolumeUSD:   stats.TotalVolumeUSD,
		MonthlyVolumeUSD: monthlyVolume,
		PlatformRate:     model.PlatformCommissionRate(totalGrossUSD),
		NetBySymbol:      netBySymbol,
	}, nil
}

func (s *CommissionService) GetProjectTransactions(ctx context.Context, partnerID uuid.UUID) ([]model.DuelTransaction, error) {
	project, err := s.ProjectRepository.GetByPartnerID(ctx, partnerID)
	if err != nil {
		return nil, apperrors.Internal("failed to get project", err)
	}
	if project == nil {
		return nil, apperrors.NotFound("partner project not found")
	}

	txs, err := s.DuelTxRepository.GetByProjectID(ctx, project.ID)
	if err != nil {
		return nil, apperrors.Internal("failed to get project transactions", err)
	}
	return txs, nil
}

func (s *CommissionService) GetDailyIncome(ctx context.Context, partnerID uuid.UUID, from, to time.Time) ([]model.PartnerDailyIncome, error) {
	project, err := s.ProjectRepository.GetByPartnerID(ctx, partnerID)
	if err != nil {
		return nil, apperrors.Internal("failed to get project", err)
	}
	if project == nil {
		return nil, apperrors.NotFound("partner project not found")
	}

	result, err := s.PartnerAnalytics.GetDailyIncome(ctx, project.ID, from, to)
	if err != nil {
		return nil, apperrors.Internal("failed to get daily income", err)
	}
	return result, nil
}

func (s *CommissionService) GetMonthlyActiveUsers(ctx context.Context, partnerID uuid.UUID) ([]model.PartnerMonthlyMAU, error) {
	project, err := s.ProjectRepository.GetByPartnerID(ctx, partnerID)
	if err != nil {
		return nil, apperrors.Internal("failed to get project", err)
	}
	if project == nil {
		return nil, apperrors.NotFound("partner project not found")
	}

	result, err := s.PartnerAnalytics.GetMonthlyActiveUsers(ctx, project.ID)
	if err != nil {
		return nil, apperrors.Internal("failed to get monthly active users", err)
	}
	return result, nil
}

func (s *CommissionService) GetMonthlyDDCommission(ctx context.Context, partnerID uuid.UUID) ([]model.PartnerMonthlyDDCommission, error) {
	project, err := s.ProjectRepository.GetByPartnerID(ctx, partnerID)
	if err != nil {
		return nil, apperrors.Internal("failed to get project", err)
	}
	if project == nil {
		return nil, apperrors.NotFound("partner project not found")
	}

	monthly, err := s.ProjectCommissionAccrualRepo.GetMonthlyAggregates(ctx, project.ID)
	if err != nil {
		return nil, apperrors.Internal("failed to get monthly commission aggregates", err)
	}

	result := make([]model.PartnerMonthlyDDCommission, 0, len(monthly))
	for _, m := range monthly {
		rate := model.PlatformCommissionRate(m.GrossUSD)
		result = append(result, model.PartnerMonthlyDDCommission{
			Month:    m.Month,
			GrossUSD: m.GrossUSD,
			DDAmount: m.GrossUSD * rate,
			DDRate:   rate,
		})
	}
	return result, nil
}

func (s *CommissionService) resolveTokenInfo(ctx context.Context, symbol string) (*model.DuelTokenInfo, error) {
	tokenInfo, err := s.CoinService.findSolanaTokenBySymbol(ctx, symbol)
	if err != nil || tokenInfo == nil {
		return nil, apperrors.Internal("failed to get token info for "+symbol, err)
	}
	return &model.DuelTokenInfo{
		Mint:      tokenInfo.Mint,
		Symbol:    tokenInfo.Symbol,
		Decimals:  tokenInfo.Decimals,
		USDPrice:  tokenInfo.USDPrice,
		ProgramID: tokenInfo.ProgramID,
	}, nil
}

func earliestBillingMonth(accruals []model.ProjectCommissionAccrual) time.Time {
	earliest := accruals[0].BillingMonth
	for _, a := range accruals[1:] {
		if a.BillingMonth.Before(earliest) {
			earliest = a.BillingMonth
		}
	}
	return earliest
}

func groupBySymbol(accruals []model.ProjectCommissionAccrual) []model.SymbolAccrual {
	m := make(map[string]*model.SymbolAccrual)
	for _, a := range accruals {
		if s, ok := m[a.Symbol]; ok {
			s.AmountRaw += a.CommissionRaw
			s.AmountUSD += a.CommissionUSD
		} else {
			m[a.Symbol] = &model.SymbolAccrual{
				Symbol:    a.Symbol,
				AmountRaw: a.CommissionRaw,
				AmountUSD: a.CommissionUSD,
			}
		}
	}
	result := make([]model.SymbolAccrual, 0, len(m))
	for _, v := range m {
		result = append(result, *v)
	}
	return result
}
