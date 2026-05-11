//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dd-prediction-api/internal/model"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// makeAccrual returns a minimal accrual for projectID with the given billing month.
func makeAccrual(projectID uuid.UUID, billingMonth time.Time) *model.ProjectCommissionAccrual {
	return &model.ProjectCommissionAccrual{
		ID:            uuid.New(),
		ProjectID:     projectID,
		DuelID:        uuid.New(), // no FK constraint on this column
		Symbol:        "USDC",
		CommissionRaw: 1_000_000,
		CommissionUSD: 1.0,
		BillingMonth:  billingMonth,
	}
}

// makeClaim builds a partner claim that covers billingMonth for the given project.
func makeClaim(projectID *uuid.UUID, claimType string, billingMonth time.Time) *model.ProjectCommissionClaim {
	return &model.ProjectCommissionClaim{
		ID:            uuid.New(),
		Type:          claimType,
		ProjectID:     projectID,
		PeriodStart:   billingMonth,
		PeriodEnd:     billingMonth.AddDate(0, 1, -1), // last day of the month
		GrossUSD:      1.0,
		PlatformRate:  0,
		AmountUSD:     1.0,
		WalletAddress: "TestWallet111",
		TxHashes:      map[string]string{"USDC": "tx_hash_abc"},
	}
}

// ── ProjectCommissionAccrualRepository ───────────────────────────────────────

func TestAccrualRepository_GetUnclaimed_noClaimsReturnsAll(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	accrualRepo := &ProjectCommissionAccrualRepository{db: tx}
	ctx := context.Background()

	projectID := uuid.New()
	jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	require.NoError(t, accrualRepo.Create(ctx, makeAccrual(projectID, jan)))
	require.NoError(t, accrualRepo.Create(ctx, makeAccrual(projectID, feb)))

	claimableThrough := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	results, err := accrualRepo.GetUnclaimed(ctx, projectID, claimableThrough)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestAccrualRepository_GetUnclaimed_claimedMonthExcluded(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	accrualRepo := &ProjectCommissionAccrualRepository{db: tx}
	claimRepo := &ProjectCommissionClaimRepository{db: tx}

	// Need a real project_id because project_commission_claims has FK.
	userRepo := newUserRepo(db, tx)
	projRepo := newProjectRepo(db, tx)
	ctx := context.Background()

	user := insertTestUser(t, userRepo, ctx)
	proj := insertTestProject(t, projRepo, ctx, user.ID)
	projectID := proj.ID

	jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	// Insert two accruals: Jan and Feb.
	require.NoError(t, accrualRepo.Create(ctx, makeAccrual(projectID, jan)))
	require.NoError(t, accrualRepo.Create(ctx, makeAccrual(projectID, feb)))

	// A partner claim that covers January only.
	require.NoError(t, claimRepo.Create(ctx, makeClaim(&projectID, model.ClaimTypePartner, jan)))

	claimableThrough := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	results, err := accrualRepo.GetUnclaimed(ctx, projectID, claimableThrough)
	require.NoError(t, err)

	// Only February should be returned.
	require.Len(t, results, 1)
	assert.Equal(t, feb, results[0].BillingMonth)
}

func TestAccrualRepository_GetUnclaimed_futureBillingMonthExcluded(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	accrualRepo := &ProjectCommissionAccrualRepository{db: tx}
	ctx := context.Background()

	projectID := uuid.New()
	past := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	future := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)

	require.NoError(t, accrualRepo.Create(ctx, makeAccrual(projectID, past)))
	require.NoError(t, accrualRepo.Create(ctx, makeAccrual(projectID, future)))

	// claimableThrough before the future accrual.
	claimableThrough := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	results, err := accrualRepo.GetUnclaimed(ctx, projectID, claimableThrough)
	require.NoError(t, err)

	require.Len(t, results, 1)
	assert.Equal(t, past, results[0].BillingMonth)
}

func TestAccrualRepository_GetUnclaimed_differentProjectIsolated(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	accrualRepo := &ProjectCommissionAccrualRepository{db: tx}
	ctx := context.Background()

	p1 := uuid.New()
	p2 := uuid.New()
	jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	require.NoError(t, accrualRepo.Create(ctx, makeAccrual(p1, jan)))
	require.NoError(t, accrualRepo.Create(ctx, makeAccrual(p2, jan)))

	claimableThrough := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	p1Results, err := accrualRepo.GetUnclaimed(ctx, p1, claimableThrough)
	require.NoError(t, err)
	assert.Len(t, p1Results, 1)

	p2Results, err := accrualRepo.GetUnclaimed(ctx, p2, claimableThrough)
	require.NoError(t, err)
	assert.Len(t, p2Results, 1)
}

func TestAccrualRepository_GetUnclaimedDD_noClaimsReturnsAll(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	accrualRepo := &ProjectCommissionAccrualRepository{db: tx}
	ctx := context.Background()

	jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	a1 := makeAccrual(uuid.New(), jan)
	a2 := makeAccrual(uuid.New(), feb)
	require.NoError(t, accrualRepo.Create(ctx, a1))
	require.NoError(t, accrualRepo.Create(ctx, a2))

	claimableThrough := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	results, err := accrualRepo.GetUnclaimedDD(ctx, claimableThrough)
	require.NoError(t, err)

	// Verify by UUID so the check is immune to existing data in the DB.
	ids := make(map[uuid.UUID]bool, len(results))
	for _, r := range results {
		ids[r.ID] = true
	}
	assert.True(t, ids[a1.ID], "a1 must appear in unclaimed DD list")
	assert.True(t, ids[a2.ID], "a2 must appear in unclaimed DD list")
}

func TestAccrualRepository_GetUnclaimedDD_ddClaimedMonthExcluded(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	accrualRepo := &ProjectCommissionAccrualRepository{db: tx}
	claimRepo := &ProjectCommissionClaimRepository{db: tx}
	ctx := context.Background()

	jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	projectID := uuid.New()
	accrualJan := makeAccrual(projectID, jan)
	accrualFeb := makeAccrual(projectID, feb)
	require.NoError(t, accrualRepo.Create(ctx, accrualJan))
	require.NoError(t, accrualRepo.Create(ctx, accrualFeb))

	// DD profit claim covering January (project_id = nil).
	require.NoError(t, claimRepo.Create(ctx, makeClaim(nil, model.ClaimTypeDDProfit, jan)))

	claimableThrough := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	results, err := accrualRepo.GetUnclaimedDD(ctx, claimableThrough)
	require.NoError(t, err)

	// January accrual must not appear; February must appear.
	for _, r := range results {
		assert.NotEqual(t, accrualJan.ID, r.ID, "claimed accrual must not appear in unclaimed list")
	}
	var foundFeb bool
	for _, r := range results {
		if r.ID == accrualFeb.ID {
			foundFeb = true
		}
	}
	assert.True(t, foundFeb, "unclaimed February accrual must appear")
}

func TestAccrualRepository_GetSummariesByProject(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	accrualRepo := &ProjectCommissionAccrualRepository{db: tx}
	ctx := context.Background()

	projectID := uuid.New()
	jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// Two USDC accruals (1.0 + 0.5 = 1.5 USD, 1_000_000 + 500_000 raw) and one SOL accrual (2.0 USD).
	a1 := makeAccrual(projectID, jan)
	a1.Symbol = "USDC"
	a1.CommissionRaw = 1_000_000
	a1.CommissionUSD = 1.0

	a2 := makeAccrual(projectID, jan)
	a2.Symbol = "USDC"
	a2.CommissionRaw = 500_000
	a2.CommissionUSD = 0.5

	a3 := makeAccrual(projectID, jan)
	a3.Symbol = "SOL"
	a3.CommissionRaw = 200_000
	a3.CommissionUSD = 2.0

	require.NoError(t, accrualRepo.Create(ctx, a1))
	require.NoError(t, accrualRepo.Create(ctx, a2))
	require.NoError(t, accrualRepo.Create(ctx, a3))

	summaries, err := accrualRepo.GetSummariesByProject(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, summaries, 2)

	bySymbol := make(map[string]model.ProjectCommissionSummary, 2)
	for _, s := range summaries {
		bySymbol[s.Symbol] = s
	}

	usdc := bySymbol["USDC"]
	assert.Equal(t, uint64(1_500_000), usdc.CommissionRaw)
	assert.InDelta(t, 1.5, usdc.CommissionUSD, 0.001)
	assert.Equal(t, projectID, usdc.ProjectID)

	sol := bySymbol["SOL"]
	assert.Equal(t, uint64(200_000), sol.CommissionRaw)
	assert.InDelta(t, 2.0, sol.CommissionUSD, 0.001)

	// totalUSD = 3.5 which is <= 1000, so platform rate = 0.
	// PlatformCommissionRate(3.5) returns 0 (first tier).
	assert.Equal(t, 0.0, usdc.PlatformRate)
	assert.Equal(t, uint64(0), usdc.PlatformFeeRaw)
}

func TestAccrualRepository_GetSummariesByProject_higherTierRate(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	accrualRepo := &ProjectCommissionAccrualRepository{db: tx}
	ctx := context.Background()

	projectID := uuid.New()
	jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// 2000 USD total triggers the 10% platform tier.
	a := makeAccrual(projectID, jan)
	a.Symbol = "USDC"
	a.CommissionRaw = 2_000_000_000
	a.CommissionUSD = 2000.0
	require.NoError(t, accrualRepo.Create(ctx, a))

	summaries, err := accrualRepo.GetSummariesByProject(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, summaries, 1)

	s := summaries[0]
	assert.InDelta(t, 0.10, s.PlatformRate, 0.001)
	assert.InDelta(t, 200.0, s.PlatformFeeUSD, 0.01)
}

func TestAccrualRepository_GetSummariesByProject_empty(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	accrualRepo := &ProjectCommissionAccrualRepository{db: tx}
	ctx := context.Background()

	summaries, err := accrualRepo.GetSummariesByProject(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, summaries)
}

// ── ProjectCommissionClaimRepository ─────────────────────────────────────────

func TestClaimRepository_Create_GetByProjectAndType(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	userRepo := newUserRepo(db, tx)
	projRepo := newProjectRepo(db, tx)
	claimRepo := &ProjectCommissionClaimRepository{db: tx}
	ctx := context.Background()

	user := insertTestUser(t, userRepo, ctx)
	proj := insertTestProject(t, projRepo, ctx, user.ID)
	projectID := proj.ID

	jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	claim := makeClaim(&projectID, model.ClaimTypePartner, jan)
	require.NoError(t, claimRepo.Create(ctx, claim))

	results, err := claimRepo.GetByProjectAndType(ctx, projectID, model.ClaimTypePartner)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, claim.ID, results[0].ID)
	assert.Equal(t, model.ClaimTypePartner, results[0].Type)
	assert.Equal(t, &projectID, results[0].ProjectID)
}

func TestClaimRepository_GetByProjectAndType_wrongTypeReturnsEmpty(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	userRepo := newUserRepo(db, tx)
	projRepo := newProjectRepo(db, tx)
	claimRepo := &ProjectCommissionClaimRepository{db: tx}
	ctx := context.Background()

	user := insertTestUser(t, userRepo, ctx)
	proj := insertTestProject(t, projRepo, ctx, user.ID)
	projectID := proj.ID

	jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, claimRepo.Create(ctx, makeClaim(&projectID, model.ClaimTypePartner, jan)))

	results, err := claimRepo.GetByProjectAndType(ctx, projectID, model.ClaimTypeDDProfit)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestClaimRepository_GetByType_DDProfit(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	claimRepo := &ProjectCommissionClaimRepository{db: tx}
	ctx := context.Background()

	jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	// DD profit claims have project_id = nil.
	dd1 := makeClaim(nil, model.ClaimTypeDDProfit, jan)
	dd2 := makeClaim(nil, model.ClaimTypeDDProfit, feb)
	require.NoError(t, claimRepo.Create(ctx, dd1))
	require.NoError(t, claimRepo.Create(ctx, dd2))

	results, err := claimRepo.GetByType(ctx, model.ClaimTypeDDProfit)
	require.NoError(t, err)

	ids := make(map[uuid.UUID]bool, len(results))
	for _, r := range results {
		ids[r.ID] = true
	}
	assert.True(t, ids[dd1.ID], "dd1 must appear")
	assert.True(t, ids[dd2.ID], "dd2 must appear")
}

func TestClaimRepository_GetByType_partnerClaimsExcludedFromDD(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	userRepo := newUserRepo(db, tx)
	projRepo := newProjectRepo(db, tx)
	claimRepo := &ProjectCommissionClaimRepository{db: tx}
	ctx := context.Background()

	user := insertTestUser(t, userRepo, ctx)
	proj := insertTestProject(t, projRepo, ctx, user.ID)
	projectID := proj.ID

	jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	partnerClaim := makeClaim(&projectID, model.ClaimTypePartner, jan)
	ddClaim := makeClaim(nil, model.ClaimTypeDDProfit, jan)
	require.NoError(t, claimRepo.Create(ctx, partnerClaim))
	require.NoError(t, claimRepo.Create(ctx, ddClaim))

	ddResults, err := claimRepo.GetByType(ctx, model.ClaimTypeDDProfit)
	require.NoError(t, err)

	for _, r := range ddResults {
		assert.NotEqual(t, partnerClaim.ID, r.ID, "partner claim must not appear in DD results")
	}

	var foundDD bool
	for _, r := range ddResults {
		if r.ID == ddClaim.ID {
			foundDD = true
		}
	}
	assert.True(t, foundDD, "DD claim must appear in DD results")
}

func TestClaimRepository_TxHashesRoundtrip(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	claimRepo := &ProjectCommissionClaimRepository{db: tx}
	userRepo := newUserRepo(db, tx)
	projRepo := newProjectRepo(db, tx)
	ctx := context.Background()

	user := insertTestUser(t, userRepo, ctx)
	proj := insertTestProject(t, projRepo, ctx, user.ID)
	projectID := proj.ID

	jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	claim := makeClaim(&projectID, model.ClaimTypePartner, jan)
	claim.TxHashes = map[string]string{
		"USDC": "usdc_tx_hash_111",
		"SOL":  "sol_tx_hash_222",
	}
	require.NoError(t, claimRepo.Create(ctx, claim))

	results, err := claimRepo.GetByProjectAndType(ctx, projectID, model.ClaimTypePartner)
	require.NoError(t, err)
	require.Len(t, results, 1)

	assert.Equal(t, "usdc_tx_hash_111", results[0].TxHashes["USDC"])
	assert.Equal(t, "sol_tx_hash_222", results[0].TxHashes["SOL"])
}
