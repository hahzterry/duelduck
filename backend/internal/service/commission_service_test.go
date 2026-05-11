package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/storage/click"
	"dd-prediction-api/internal/storage/repository"
)

// ── mocks ────────────────────────────────────────────────────────────────────

type mockProjectClaimRepo struct{ mock.Mock }

func (m *mockProjectClaimRepo) Create(ctx context.Context, claim *model.ProjectCommissionClaim) error {
	return m.Called(ctx, claim).Error(0)
}

func (m *mockProjectClaimRepo) GetByProjectAndType(ctx context.Context, projectID uuid.UUID, claimType string) ([]model.ProjectCommissionClaim, error) {
	args := m.Called(ctx, projectID, claimType)
	res, _ := args.Get(0).([]model.ProjectCommissionClaim)
	return res, args.Error(1)
}

func (m *mockProjectClaimRepo) GetByType(ctx context.Context, claimType string) ([]model.ProjectCommissionClaim, error) {
	args := m.Called(ctx, claimType)
	res, _ := args.Get(0).([]model.ProjectCommissionClaim)
	return res, args.Error(1)
}

type mockProjectAccrualRepo struct{ mock.Mock }

func (m *mockProjectAccrualRepo) GetUnclaimed(ctx context.Context, projectID uuid.UUID, through time.Time) ([]model.ProjectCommissionAccrual, error) {
	args := m.Called(ctx, projectID, through)
	res, _ := args.Get(0).([]model.ProjectCommissionAccrual)
	return res, args.Error(1)
}

func (m *mockProjectAccrualRepo) GetUnclaimedDD(ctx context.Context, through time.Time) ([]model.ProjectCommissionAccrual, error) {
	args := m.Called(ctx, through)
	res, _ := args.Get(0).([]model.ProjectCommissionAccrual)
	return res, args.Error(1)
}

func (m *mockProjectAccrualRepo) GetSummariesByProject(ctx context.Context, projectID uuid.UUID) ([]model.ProjectCommissionSummary, error) {
	args := m.Called(ctx, projectID)
	res, _ := args.Get(0).([]model.ProjectCommissionSummary)
	return res, args.Error(1)
}

func (m *mockProjectAccrualRepo) GetMonthlyAggregates(ctx context.Context, projectID uuid.UUID) ([]repository.MonthlyAccrualAggregate, error) {
	args := m.Called(ctx, projectID)
	res, _ := args.Get(0).([]repository.MonthlyAccrualAggregate)
	return res, args.Error(1)
}

type mockCommissionProjectRepo struct{ mock.Mock }

func (m *mockCommissionProjectRepo) GetByPartnerID(ctx context.Context, partnerID uuid.UUID) (*model.Project, error) {
	args := m.Called(ctx, partnerID)
	res, _ := args.Get(0).(*model.Project)
	return res, args.Error(1)
}

func (m *mockCommissionProjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	args := m.Called(ctx, id)
	res, _ := args.Get(0).(*model.Project)
	return res, args.Error(1)
}

func (m *mockCommissionProjectRepo) GetByAPIKey(ctx context.Context, k string) (*model.Project, error) {
	args := m.Called(ctx, k)
	res, _ := args.Get(0).(*model.Project)
	return res, args.Error(1)
}

func (m *mockCommissionProjectRepo) IsBlocked(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *mockCommissionProjectRepo) SetBlocked(ctx context.Context, id uuid.UUID, blocked bool) error {
	return m.Called(ctx, id, blocked).Error(0)
}

func (m *mockCommissionProjectRepo) Create(ctx context.Context, p *model.Project) error {
	return m.Called(ctx, p).Error(0)
}

func (m *mockCommissionProjectRepo) HasPartnerProject(ctx context.Context, partnerID uuid.UUID) (bool, error) {
	args := m.Called(ctx, partnerID)
	return args.Bool(0), args.Error(1)
}

func (m *mockCommissionProjectRepo) GetAll(ctx context.Context) ([]model.Project, error) {
	args := m.Called(ctx)
	res, _ := args.Get(0).([]model.Project)
	return res, args.Error(1)
}

func (m *mockCommissionProjectRepo) SetStatus(ctx context.Context, id uuid.UUID, status model.ProjectStatus) error {
	return m.Called(ctx, id, status).Error(0)
}

func (m *mockCommissionProjectRepo) Update(ctx context.Context, id uuid.UUID, name, siteURL string) error {
	return m.Called(ctx, id, name, siteURL).Error(0)
}

func (m *mockCommissionProjectRepo) UpdatePermissions(ctx context.Context, id uuid.UUID, duelsEnabled, selfResolveEnabled bool) error {
	return m.Called(ctx, id, duelsEnabled, selfResolveEnabled).Error(0)
}

func (m *mockCommissionProjectRepo) WithTx(_ bun.Tx) repository.IProjectRepository {
	return m
}

type mockCommissionDuelRepo struct{ mock.Mock }

func (m *mockCommissionDuelRepo) GetProjectDuelStats(ctx context.Context, projectID uuid.UUID) (*repository.ProjectDuelStats, error) {
	args := m.Called(ctx, projectID)
	res, _ := args.Get(0).(*repository.ProjectDuelStats)
	return res, args.Error(1)
}

type mockCommissionWalletService struct{ mock.Mock }

func (m *mockCommissionWalletService) TransferSymbol(ctx context.Context, recipientAddress solana.PublicKey, amount uint64, info *model.DuelTokenInfo) (string, error) {
	args := m.Called(ctx, recipientAddress, amount, info)
	return args.String(0), args.Error(1)
}

type mockCommissionCoinService struct{ mock.Mock }

func (m *mockCommissionCoinService) findSolanaTokenBySymbol(ctx context.Context, symbol string) (*model.SolanaToken, error) {
	args := m.Called(ctx, symbol)
	res, _ := args.Get(0).(*model.SolanaToken)
	return res, args.Error(1)
}

type mockPartnerAnalytics struct{ mock.Mock }

func (m *mockPartnerAnalytics) GetDailyIncome(ctx context.Context, projectID uuid.UUID, from, to time.Time) ([]model.PartnerDailyIncome, error) {
	args := m.Called(ctx, projectID, from, to)
	res, _ := args.Get(0).([]model.PartnerDailyIncome)
	return res, args.Error(1)
}

func (m *mockPartnerAnalytics) GetMonthlyActiveUsers(ctx context.Context, projectID uuid.UUID) ([]model.PartnerMonthlyMAU, error) {
	args := m.Called(ctx, projectID)
	res, _ := args.Get(0).([]model.PartnerMonthlyMAU)
	return res, args.Error(1)
}

func (m *mockPartnerAnalytics) GetCurrentMonthVolume(ctx context.Context, projectID uuid.UUID) (float64, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).(float64), args.Error(1)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newTestCommissionService(
	claimRepo IProjectCommissionClaimRepository,
	accrualRepo IProjectCommissionAccrualRepository,
	projectRepo repository.IProjectRepository,
	duelRepo ICommissionDuelRepository,
	walletSvc ICommissionWalletService,
	coinSvc ICoinService,
	duelTxRepos ...click.IDuelTransactionRepository,
) *CommissionService {
	analytics := new(mockPartnerAnalytics)
	analytics.On("GetCurrentMonthVolume", mock.Anything, mock.Anything).Return(0.0, nil)
	var duelTxRepo click.IDuelTransactionRepository
	if len(duelTxRepos) > 0 {
		duelTxRepo = duelTxRepos[0]
	}
	return &CommissionService{
		ProjectClaimRepo:             claimRepo,
		ProjectCommissionAccrualRepo: accrualRepo,
		ProjectRepository:            projectRepo,
		DuelRepository:               duelRepo,
		WalletService:                walletSvc,
		CoinService:                  coinSvc,
		DuelTxRepository:             duelTxRepo,
		PartnerAnalytics:             analytics,
		ddProfitWallet:               "11111111111111111111111111111113",
	}
}

// ── GetPartnerDashboard ───────────────────────────────────────────────────────

func TestGetPartnerDashboard(t *testing.T) {
	partnerID := uuid.New()
	projectID := uuid.New()

	project := &model.Project{
		ID:        projectID,
		APIKey:    "dd_test_key",
		PartnerID: partnerID,
		Status:    model.ProjectStatusActive,
	}

	stats := &repository.ProjectDuelStats{
		DuelCount:      10,
		TotalVolumeUSD: 500.0,
	}

	summaries := []model.ProjectCommissionSummary{
		{CommissionUSD: 50.0, PlatformFeeUSD: 5.0},
		{CommissionUSD: 30.0, PlatformFeeUSD: 3.0},
	}

	tests := []struct {
		name    string
		setup   func(pRepo *mockCommissionProjectRepo, dRepo *mockCommissionDuelRepo, aRepo *mockProjectAccrualRepo)
		wantErr bool
		check   func(t *testing.T, dashboard *model.PartnerDashboard)
	}{
		{
			name: "success returns dashboard with aggregated data",
			setup: func(pRepo *mockCommissionProjectRepo, dRepo *mockCommissionDuelRepo, aRepo *mockProjectAccrualRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				dRepo.On("GetProjectDuelStats", mock.Anything, projectID).Return(stats, nil)
				aRepo.On("GetSummariesByProject", mock.Anything, projectID).Return(summaries, nil)
			},
			check: func(t *testing.T, dashboard *model.PartnerDashboard) {
				require.NotNil(t, dashboard)
				assert.Equal(t, "dd_test_key", dashboard.APIKey)
				assert.Equal(t, 10, dashboard.DuelCount)
				assert.Equal(t, 500.0, dashboard.TotalVolumeUSD)
				assert.Equal(t, model.ProjectStatusActive, dashboard.Status)
				assert.Len(t, dashboard.NetBySymbol, 2)
				var totalGross, totalFee float64
				for _, s := range dashboard.NetBySymbol {
					totalGross += s.GrossUSD
					totalFee += s.GrossUSD - s.NetUSD
				}
				assert.Equal(t, 80.0, totalGross)
				assert.Equal(t, 8.0, totalFee)
			},
		},
		{
			name: "project not found returns error",
			setup: func(pRepo *mockCommissionProjectRepo, dRepo *mockCommissionDuelRepo, aRepo *mockProjectAccrualRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "project repo error propagates",
			setup: func(pRepo *mockCommissionProjectRepo, dRepo *mockCommissionDuelRepo, aRepo *mockProjectAccrualRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "duel stats repo error propagates",
			setup: func(pRepo *mockCommissionProjectRepo, dRepo *mockCommissionDuelRepo, aRepo *mockProjectAccrualRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				dRepo.On("GetProjectDuelStats", mock.Anything, projectID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "accrual summaries repo error propagates",
			setup: func(pRepo *mockCommissionProjectRepo, dRepo *mockCommissionDuelRepo, aRepo *mockProjectAccrualRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				dRepo.On("GetProjectDuelStats", mock.Anything, projectID).Return(stats, nil)
				aRepo.On("GetSummariesByProject", mock.Anything, projectID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pRepo := new(mockCommissionProjectRepo)
			dRepo := new(mockCommissionDuelRepo)
			aRepo := new(mockProjectAccrualRepo)
			wSvc := new(mockCommissionWalletService)
			cSvc := new(mockCommissionCoinService)
			cRepo := new(mockProjectClaimRepo)

			tt.setup(pRepo, dRepo, aRepo)

			svc := newTestCommissionService(cRepo, aRepo, pRepo, dRepo, wSvc, cSvc)
			dashboard, err := svc.GetPartnerDashboard(context.Background(), partnerID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, dashboard)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, dashboard)
				}
			}
		})
	}
}

// ── GetProjectClaimHistory ────────────────────────────────────────────────────

func TestGetProjectClaimHistory(t *testing.T) {
	partnerID := uuid.New()
	projectID := uuid.New()

	project := &model.Project{ID: projectID, PartnerID: partnerID}

	claims := []model.ProjectCommissionClaim{
		{ID: uuid.New(), Type: model.ClaimTypePartner, ProjectID: &projectID, AmountUSD: 100.0},
		{ID: uuid.New(), Type: model.ClaimTypePartner, ProjectID: &projectID, AmountUSD: 50.0},
	}

	tests := []struct {
		name    string
		setup   func(pRepo *mockCommissionProjectRepo, cRepo *mockProjectClaimRepo)
		wantErr bool
		check   func(t *testing.T, claims []model.ProjectCommissionClaim)
	}{
		{
			name: "success returns claims",
			setup: func(pRepo *mockCommissionProjectRepo, cRepo *mockProjectClaimRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				cRepo.On("GetByProjectAndType", mock.Anything, projectID, model.ClaimTypePartner).Return(claims, nil)
			},
			check: func(t *testing.T, returned []model.ProjectCommissionClaim) {
				assert.Len(t, returned, 2)
				assert.Equal(t, 100.0, returned[0].AmountUSD)
				assert.Equal(t, 50.0, returned[1].AmountUSD)
			},
		},
		{
			name: "project not found returns error",
			setup: func(pRepo *mockCommissionProjectRepo, cRepo *mockProjectClaimRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "no claims returns empty slice",
			setup: func(pRepo *mockCommissionProjectRepo, cRepo *mockProjectClaimRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				cRepo.On("GetByProjectAndType", mock.Anything, projectID, model.ClaimTypePartner).Return([]model.ProjectCommissionClaim{}, nil)
			},
			check: func(t *testing.T, returned []model.ProjectCommissionClaim) {
				assert.Empty(t, returned)
			},
		},
		{
			name: "claim repo error propagates",
			setup: func(pRepo *mockCommissionProjectRepo, cRepo *mockProjectClaimRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				cRepo.On("GetByProjectAndType", mock.Anything, projectID, model.ClaimTypePartner).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pRepo := new(mockCommissionProjectRepo)
			cRepo := new(mockProjectClaimRepo)
			aRepo := new(mockProjectAccrualRepo)
			dRepo := new(mockCommissionDuelRepo)
			wSvc := new(mockCommissionWalletService)
			cSvc := new(mockCommissionCoinService)

			tt.setup(pRepo, cRepo)

			svc := newTestCommissionService(cRepo, aRepo, pRepo, dRepo, wSvc, cSvc)
			returned, err := svc.GetProjectClaimHistory(context.Background(), partnerID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, returned)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, returned)
				}
			}
		})
	}
}

// ── GetDDClaimHistory ─────────────────────────────────────────────────────────

func TestGetDDClaimHistory(t *testing.T) {
	claims := []model.ProjectCommissionClaim{
		{ID: uuid.New(), Type: model.ClaimTypeDDProfit},
		{ID: uuid.New(), Type: model.ClaimTypeDDProfit},
	}

	tests := []struct {
		name    string
		setup   func(cRepo *mockProjectClaimRepo)
		wantErr bool
		check   func(t *testing.T, returned []model.ProjectCommissionClaim)
	}{
		{
			name: "success returns DD profit claims",
			setup: func(cRepo *mockProjectClaimRepo) {
				cRepo.On("GetByType", mock.Anything, model.ClaimTypeDDProfit).Return(claims, nil)
			},
			check: func(t *testing.T, returned []model.ProjectCommissionClaim) {
				assert.Len(t, returned, 2)
				for _, c := range returned {
					assert.Equal(t, model.ClaimTypeDDProfit, c.Type)
				}
			},
		},
		{
			name: "repo error propagates",
			setup: func(cRepo *mockProjectClaimRepo) {
				cRepo.On("GetByType", mock.Anything, model.ClaimTypeDDProfit).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cRepo := new(mockProjectClaimRepo)
			aRepo := new(mockProjectAccrualRepo)
			pRepo := new(mockCommissionProjectRepo)
			dRepo := new(mockCommissionDuelRepo)
			wSvc := new(mockCommissionWalletService)
			cSvc := new(mockCommissionCoinService)

			tt.setup(cRepo)

			svc := newTestCommissionService(cRepo, aRepo, pRepo, dRepo, wSvc, cSvc)
			returned, err := svc.GetDDClaimHistory(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, returned)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, returned)
				}
			}
		})
	}
}

// ── ClaimProjectCommission ────────────────────────────────────────────────────

func TestClaimProjectCommission(t *testing.T) {
	partnerID := uuid.New()
	projectID := uuid.New()

	project := &model.Project{ID: projectID, PartnerID: partnerID}

	usdcToken := &model.SolanaToken{
		Mint:      "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		Symbol:    "USDC",
		Decimals:  6,
		USDPrice:  1.0,
		ProgramID: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
	}

	// Use $2000 so PlatformCommissionRate returns 10% (> 0), enabling net transfer.
	accruals := []model.ProjectCommissionAccrual{
		{
			ID:            uuid.New(),
			ProjectID:     projectID,
			DuelID:        uuid.New(),
			Symbol:        "USDC",
			CommissionRaw: 2_000_000_000,
			CommissionUSD: 2000.0,
			BillingMonth:  time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	req := &model.ProjectCommissionClaimReq{WalletAddress: "11111111111111111111111111111112"}

	tests := []struct {
		name    string
		setup   func(pRepo *mockCommissionProjectRepo, aRepo *mockProjectAccrualRepo, cRepo *mockProjectClaimRepo, wSvc *mockCommissionWalletService, cSvc *mockCommissionCoinService)
		wantErr bool
		check   func(t *testing.T, claim *model.ProjectCommissionClaim)
	}{
		{
			name: "success — single USDC accrual — correct net amount",
			setup: func(pRepo *mockCommissionProjectRepo, aRepo *mockProjectAccrualRepo, cRepo *mockProjectClaimRepo, wSvc *mockCommissionWalletService, cSvc *mockCommissionCoinService) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				aRepo.On("GetUnclaimed", mock.Anything, projectID, mock.AnythingOfType("time.Time")).Return(accruals, nil)
				cSvc.On("findSolanaTokenBySymbol", mock.Anything, "USDC").Return(usdcToken, nil)
				wSvc.On("TransferSymbol", mock.Anything, solana.MustPublicKeyFromBase58(req.WalletAddress), mock.AnythingOfType("uint64"), mock.Anything).Return("tx-hash-abc", nil)
				cRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.ProjectCommissionClaim")).Return(nil)
			},
			check: func(t *testing.T, claim *model.ProjectCommissionClaim) {
				require.NotNil(t, claim)
				assert.Equal(t, model.ClaimTypePartner, claim.Type)
				assert.Equal(t, &projectID, claim.ProjectID)
				assert.Equal(t, req.WalletAddress, claim.WalletAddress)
				assert.Greater(t, claim.GrossUSD, 0.0)
				assert.NotEmpty(t, claim.TxHashes)
			},
		},
		{
			name: "project not found returns error",
			setup: func(pRepo *mockCommissionProjectRepo, aRepo *mockProjectAccrualRepo, cRepo *mockProjectClaimRepo, wSvc *mockCommissionWalletService, cSvc *mockCommissionCoinService) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "project repo error propagates",
			setup: func(pRepo *mockCommissionProjectRepo, aRepo *mockProjectAccrualRepo, cRepo *mockProjectClaimRepo, wSvc *mockCommissionWalletService, cSvc *mockCommissionCoinService) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "no unclaimed accruals returns bad request",
			setup: func(pRepo *mockCommissionProjectRepo, aRepo *mockProjectAccrualRepo, cRepo *mockProjectClaimRepo, wSvc *mockCommissionWalletService, cSvc *mockCommissionCoinService) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				aRepo.On("GetUnclaimed", mock.Anything, projectID, mock.AnythingOfType("time.Time")).Return([]model.ProjectCommissionAccrual{}, nil)
			},
			wantErr: true,
		},
		{
			name: "wallet transfer error propagates",
			setup: func(pRepo *mockCommissionProjectRepo, aRepo *mockProjectAccrualRepo, cRepo *mockProjectClaimRepo, wSvc *mockCommissionWalletService, cSvc *mockCommissionCoinService) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				aRepo.On("GetUnclaimed", mock.Anything, projectID, mock.AnythingOfType("time.Time")).Return(accruals, nil)
				cSvc.On("findSolanaTokenBySymbol", mock.Anything, "USDC").Return(usdcToken, nil)
				wSvc.On("TransferSymbol", mock.Anything, solana.MustPublicKeyFromBase58(req.WalletAddress), mock.AnythingOfType("uint64"), mock.Anything).Return("", errors.New("transfer failed"))
			},
			wantErr: true,
		},
		{
			name: "coin service error propagates",
			setup: func(pRepo *mockCommissionProjectRepo, aRepo *mockProjectAccrualRepo, cRepo *mockProjectClaimRepo, wSvc *mockCommissionWalletService, cSvc *mockCommissionCoinService) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				aRepo.On("GetUnclaimed", mock.Anything, projectID, mock.AnythingOfType("time.Time")).Return(accruals, nil)
				cSvc.On("findSolanaTokenBySymbol", mock.Anything, "USDC").Return(nil, errors.New("coin error"))
			},
			wantErr: true,
		},
		{
			name: "claim repo Create error propagates",
			setup: func(pRepo *mockCommissionProjectRepo, aRepo *mockProjectAccrualRepo, cRepo *mockProjectClaimRepo, wSvc *mockCommissionWalletService, cSvc *mockCommissionCoinService) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				aRepo.On("GetUnclaimed", mock.Anything, projectID, mock.AnythingOfType("time.Time")).Return(accruals, nil)
				cSvc.On("findSolanaTokenBySymbol", mock.Anything, "USDC").Return(usdcToken, nil)
				wSvc.On("TransferSymbol", mock.Anything, solana.MustPublicKeyFromBase58(req.WalletAddress), mock.AnythingOfType("uint64"), mock.Anything).Return("tx-hash-abc", nil)
				cRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.ProjectCommissionClaim")).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pRepo := new(mockCommissionProjectRepo)
			aRepo := new(mockProjectAccrualRepo)
			cRepo := new(mockProjectClaimRepo)
			dRepo := new(mockCommissionDuelRepo)
			wSvc := new(mockCommissionWalletService)
			cSvc := new(mockCommissionCoinService)

			tt.setup(pRepo, aRepo, cRepo, wSvc, cSvc)

			svc := newTestCommissionService(cRepo, aRepo, pRepo, dRepo, wSvc, cSvc)
			claim, err := svc.ClaimProjectCommission(context.Background(), partnerID, req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claim)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, claim)
				}
			}
		})
	}
}

// ── ClaimDDProfit ─────────────────────────────────────────────────────────────

func TestClaimDDProfit(t *testing.T) {
	const ddWallet = "11111111111111111111111111111113"

	usdcToken := &model.SolanaToken{
		Mint:      "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		Symbol:    "USDC",
		Decimals:  6,
		USDPrice:  1.0,
		ProgramID: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
	}

	// PlatformCommissionRate returns 0 for <= $1000, so use $2000 to get 10% rate.
	accruals := []model.ProjectCommissionAccrual{
		{
			ID:            uuid.New(),
			ProjectID:     uuid.New(),
			DuelID:        uuid.New(),
			Symbol:        "USDC",
			CommissionRaw: 2_000_000_000,
			CommissionUSD: 2000.0,
			BillingMonth:  time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	tests := []struct {
		name    string
		setup   func(aRepo *mockProjectAccrualRepo, cRepo *mockProjectClaimRepo, wSvc *mockCommissionWalletService, cSvc *mockCommissionCoinService)
		wantErr bool
		check   func(t *testing.T, claim *model.ProjectCommissionClaim)
	}{
		{
			name: "success — correct DD fee calculated",
			setup: func(aRepo *mockProjectAccrualRepo, cRepo *mockProjectClaimRepo, wSvc *mockCommissionWalletService, cSvc *mockCommissionCoinService) {
				aRepo.On("GetUnclaimedDD", mock.Anything, mock.AnythingOfType("time.Time")).Return(accruals, nil)
				cSvc.On("findSolanaTokenBySymbol", mock.Anything, "USDC").Return(usdcToken, nil)
				wSvc.On("TransferSymbol", mock.Anything, solana.MustPublicKeyFromBase58(ddWallet), mock.AnythingOfType("uint64"), mock.Anything).Return("dd-tx-hash", nil)
				cRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.ProjectCommissionClaim")).Return(nil)
			},
			check: func(t *testing.T, claim *model.ProjectCommissionClaim) {
				require.NotNil(t, claim)
				assert.Equal(t, model.ClaimTypeDDProfit, claim.Type)
				assert.Nil(t, claim.ProjectID)
				assert.Equal(t, ddWallet, claim.WalletAddress)
				assert.Equal(t, 2000.0, claim.GrossUSD)
				assert.Greater(t, claim.PlatformRate, 0.0)
				assert.NotEmpty(t, claim.TxHashes)
			},
		},
		{
			name: "no unclaimed accruals returns bad request",
			setup: func(aRepo *mockProjectAccrualRepo, cRepo *mockProjectClaimRepo, wSvc *mockCommissionWalletService, cSvc *mockCommissionCoinService) {
				aRepo.On("GetUnclaimedDD", mock.Anything, mock.AnythingOfType("time.Time")).Return([]model.ProjectCommissionAccrual{}, nil)
			},
			wantErr: true,
		},
		{
			name: "accrual repo error propagates",
			setup: func(aRepo *mockProjectAccrualRepo, cRepo *mockProjectClaimRepo, wSvc *mockCommissionWalletService, cSvc *mockCommissionCoinService) {
				aRepo.On("GetUnclaimedDD", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "wallet transfer error propagates",
			setup: func(aRepo *mockProjectAccrualRepo, cRepo *mockProjectClaimRepo, wSvc *mockCommissionWalletService, cSvc *mockCommissionCoinService) {
				aRepo.On("GetUnclaimedDD", mock.Anything, mock.AnythingOfType("time.Time")).Return(accruals, nil)
				cSvc.On("findSolanaTokenBySymbol", mock.Anything, "USDC").Return(usdcToken, nil)
				wSvc.On("TransferSymbol", mock.Anything, solana.MustPublicKeyFromBase58(ddWallet), mock.AnythingOfType("uint64"), mock.Anything).Return("", errors.New("transfer error"))
			},
			wantErr: true,
		},
		{
			name: "claim Create error propagates",
			setup: func(aRepo *mockProjectAccrualRepo, cRepo *mockProjectClaimRepo, wSvc *mockCommissionWalletService, cSvc *mockCommissionCoinService) {
				aRepo.On("GetUnclaimedDD", mock.Anything, mock.AnythingOfType("time.Time")).Return(accruals, nil)
				cSvc.On("findSolanaTokenBySymbol", mock.Anything, "USDC").Return(usdcToken, nil)
				wSvc.On("TransferSymbol", mock.Anything, solana.MustPublicKeyFromBase58(ddWallet), mock.AnythingOfType("uint64"), mock.Anything).Return("dd-tx-hash", nil)
				cRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.ProjectCommissionClaim")).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cRepo := new(mockProjectClaimRepo)
			aRepo := new(mockProjectAccrualRepo)
			pRepo := new(mockCommissionProjectRepo)
			dRepo := new(mockCommissionDuelRepo)
			wSvc := new(mockCommissionWalletService)
			cSvc := new(mockCommissionCoinService)

			tt.setup(aRepo, cRepo, wSvc, cSvc)

			svc := newTestCommissionService(cRepo, aRepo, pRepo, dRepo, wSvc, cSvc)
			claim, err := svc.ClaimDDProfit(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claim)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, claim)
				}
			}
		})
	}
}

// ── pure helpers ──────────────────────────────────────────────────────────────

func TestEarliestBillingMonth(t *testing.T) {
	jan := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	mar := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	accruals := []model.ProjectCommissionAccrual{
		{BillingMonth: feb},
		{BillingMonth: jan},
		{BillingMonth: mar},
	}

	result := earliestBillingMonth(accruals)
	assert.Equal(t, jan, result)
}

func TestGroupBySymbol(t *testing.T) {
	accruals := []model.ProjectCommissionAccrual{
		{Symbol: "USDC", CommissionRaw: 1_000_000, CommissionUSD: 1.0},
		{Symbol: "USDC", CommissionRaw: 500_000, CommissionUSD: 0.5},
		{Symbol: "SOL", CommissionRaw: 100_000, CommissionUSD: 2.0},
	}

	groups := groupBySymbol(accruals)
	require.Len(t, groups, 2)

	bySymbol := make(map[string]model.SymbolAccrual, len(groups))
	for _, g := range groups {
		bySymbol[g.Symbol] = g
	}

	usdc := bySymbol["USDC"]
	assert.Equal(t, uint64(1_500_000), usdc.AmountRaw)
	assert.Equal(t, 1.5, usdc.AmountUSD)

	sol := bySymbol["SOL"]
	assert.Equal(t, uint64(100_000), sol.AmountRaw)
	assert.Equal(t, 2.0, sol.AmountUSD)
}

// ── GetProjectTransactions ────────────────────────────────────────────────────

func TestGetProjectTransactions(t *testing.T) {
	partnerID := uuid.New()
	projectID := uuid.New()
	project := &model.Project{ID: projectID, PartnerID: partnerID}

	txs := []model.DuelTransaction{
		{TxType: model.TransactionTypeDuelPrediction},
		{TxType: model.TransactionTypeDuelCommission},
	}

	tests := []struct {
		name    string
		setup   func(pRepo *mockCommissionProjectRepo, txRepo *mockDuelTxRepo)
		wantErr bool
		check   func(t *testing.T, result []model.DuelTransaction)
	}{
		{
			name: "success returns transactions",
			setup: func(pRepo *mockCommissionProjectRepo, txRepo *mockDuelTxRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				txRepo.On("GetByProjectID", mock.Anything, projectID).Return(txs, nil)
			},
			check: func(t *testing.T, result []model.DuelTransaction) {
				assert.Len(t, result, 2)
			},
		},
		{
			name: "project not found returns error",
			setup: func(pRepo *mockCommissionProjectRepo, txRepo *mockDuelTxRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "tx repo error propagates",
			setup: func(pRepo *mockCommissionProjectRepo, txRepo *mockDuelTxRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				txRepo.On("GetByProjectID", mock.Anything, projectID).Return(nil, errors.New("ch error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pRepo := new(mockCommissionProjectRepo)
			txRepo := new(mockDuelTxRepo)
			tt.setup(pRepo, txRepo)

			svc := newTestCommissionService(
				new(mockProjectClaimRepo), new(mockProjectAccrualRepo),
				pRepo, new(mockCommissionDuelRepo),
				new(mockCommissionWalletService), new(mockCommissionCoinService),
				txRepo,
			)
			result, err := svc.GetProjectTransactions(context.Background(), partnerID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, result)
				}
			}
		})
	}
}

// ── GetDailyIncome ────────────────────────────────────────────────────────────

func TestGetDailyIncome(t *testing.T) {
	partnerID := uuid.New()
	projectID := uuid.New()
	project := &model.Project{ID: projectID, PartnerID: partnerID}
	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	income := []model.PartnerDailyIncome{
		{Day: "2026-05-01", Amount: 10.5},
		{Day: "2026-05-02", Amount: 20.0},
	}

	tests := []struct {
		name    string
		setup   func(pRepo *mockCommissionProjectRepo, analytics *mockPartnerAnalytics)
		wantErr bool
		check   func(t *testing.T, result []model.PartnerDailyIncome)
	}{
		{
			name: "success returns daily income",
			setup: func(pRepo *mockCommissionProjectRepo, analytics *mockPartnerAnalytics) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				analytics.On("GetDailyIncome", mock.Anything, projectID, from, to).Return(income, nil)
			},
			check: func(t *testing.T, result []model.PartnerDailyIncome) {
				require.Len(t, result, 2)
				assert.Equal(t, "2026-05-01", result[0].Day)
				assert.Equal(t, 10.5, result[0].Amount)
			},
		},
		{
			name: "project not found returns error",
			setup: func(pRepo *mockCommissionProjectRepo, analytics *mockPartnerAnalytics) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "analytics error propagates",
			setup: func(pRepo *mockCommissionProjectRepo, analytics *mockPartnerAnalytics) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				analytics.On("GetDailyIncome", mock.Anything, projectID, from, to).Return(nil, errors.New("ch error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pRepo := new(mockCommissionProjectRepo)
			analytics := new(mockPartnerAnalytics)
			analytics.On("GetCurrentMonthVolume", mock.Anything, mock.Anything).Return(0.0, nil)
			tt.setup(pRepo, analytics)

			svc := newTestCommissionService(
				new(mockProjectClaimRepo), new(mockProjectAccrualRepo),
				pRepo, new(mockCommissionDuelRepo),
				new(mockCommissionWalletService), new(mockCommissionCoinService),
			)
			svc.PartnerAnalytics = analytics

			result, err := svc.GetDailyIncome(context.Background(), partnerID, from, to)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, result)
				}
			}
		})
	}
}

// ── GetMonthlyActiveUsers ─────────────────────────────────────────────────────

func TestGetMonthlyActiveUsers(t *testing.T) {
	partnerID := uuid.New()
	projectID := uuid.New()
	project := &model.Project{ID: projectID, PartnerID: partnerID}

	mau := []model.PartnerMonthlyMAU{
		{Month: "2026-03", ActiveUsers: 120},
		{Month: "2026-04", ActiveUsers: 95},
	}

	tests := []struct {
		name    string
		setup   func(pRepo *mockCommissionProjectRepo, analytics *mockPartnerAnalytics)
		wantErr bool
		check   func(t *testing.T, result []model.PartnerMonthlyMAU)
	}{
		{
			name: "success returns monthly active users",
			setup: func(pRepo *mockCommissionProjectRepo, analytics *mockPartnerAnalytics) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				analytics.On("GetMonthlyActiveUsers", mock.Anything, projectID).Return(mau, nil)
			},
			check: func(t *testing.T, result []model.PartnerMonthlyMAU) {
				require.Len(t, result, 2)
				assert.Equal(t, uint64(120), result[0].ActiveUsers)
			},
		},
		{
			name: "project not found returns error",
			setup: func(pRepo *mockCommissionProjectRepo, analytics *mockPartnerAnalytics) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "analytics error propagates",
			setup: func(pRepo *mockCommissionProjectRepo, analytics *mockPartnerAnalytics) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				analytics.On("GetMonthlyActiveUsers", mock.Anything, projectID).Return(nil, errors.New("ch error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pRepo := new(mockCommissionProjectRepo)
			analytics := new(mockPartnerAnalytics)
			analytics.On("GetCurrentMonthVolume", mock.Anything, mock.Anything).Return(0.0, nil)
			tt.setup(pRepo, analytics)

			svc := newTestCommissionService(
				new(mockProjectClaimRepo), new(mockProjectAccrualRepo),
				pRepo, new(mockCommissionDuelRepo),
				new(mockCommissionWalletService), new(mockCommissionCoinService),
			)
			svc.PartnerAnalytics = analytics

			result, err := svc.GetMonthlyActiveUsers(context.Background(), partnerID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, result)
				}
			}
		})
	}
}

// ── GetMonthlyDDCommission ────────────────────────────────────────────────────

func TestGetMonthlyDDCommission(t *testing.T) {
	partnerID := uuid.New()
	projectID := uuid.New()
	project := &model.Project{ID: projectID, PartnerID: partnerID}

	aggregates := []repository.MonthlyAccrualAggregate{
		{Month: "2026-03", GrossUSD: 2000.0},
		{Month: "2026-04", GrossUSD: 500.0},
	}

	tests := []struct {
		name    string
		setup   func(pRepo *mockCommissionProjectRepo, aRepo *mockProjectAccrualRepo)
		wantErr bool
		check   func(t *testing.T, result []model.PartnerMonthlyDDCommission)
	}{
		{
			name: "success computes dd commission per month",
			setup: func(pRepo *mockCommissionProjectRepo, aRepo *mockProjectAccrualRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				aRepo.On("GetMonthlyAggregates", mock.Anything, projectID).Return(aggregates, nil)
			},
			check: func(t *testing.T, result []model.PartnerMonthlyDDCommission) {
				require.Len(t, result, 2)
				// 2026-03: $2000 gross → 10% rate
				assert.Equal(t, "2026-03", result[0].Month)
				assert.Equal(t, 2000.0, result[0].GrossUSD)
				assert.InDelta(t, 0.10, result[0].DDRate, 0.001)
				assert.InDelta(t, 200.0, result[0].DDAmount, 0.001)
				// 2026-04: $500 gross → 0% rate (≤ $1000)
				assert.Equal(t, "2026-04", result[1].Month)
				assert.Equal(t, 0.0, result[1].DDRate)
				assert.Equal(t, 0.0, result[1].DDAmount)
			},
		},
		{
			name: "project not found returns error",
			setup: func(pRepo *mockCommissionProjectRepo, aRepo *mockProjectAccrualRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "accrual repo error propagates",
			setup: func(pRepo *mockCommissionProjectRepo, aRepo *mockProjectAccrualRepo) {
				pRepo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				aRepo.On("GetMonthlyAggregates", mock.Anything, projectID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pRepo := new(mockCommissionProjectRepo)
			aRepo := new(mockProjectAccrualRepo)
			tt.setup(pRepo, aRepo)

			svc := newTestCommissionService(
				new(mockProjectClaimRepo), aRepo,
				pRepo, new(mockCommissionDuelRepo),
				new(mockCommissionWalletService), new(mockCommissionCoinService),
			)
			result, err := svc.GetMonthlyDDCommission(context.Background(), partnerID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, result)
				}
			}
		})
	}
}
