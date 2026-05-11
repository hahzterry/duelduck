package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"dd-prediction-api/internal/model"
)

type ProjectCommissionAccrualRepository struct {
	db bun.IDB
}

func NewProjectCommissionAccrualRepository(db *bun.DB) *ProjectCommissionAccrualRepository {
	return &ProjectCommissionAccrualRepository{db: db}
}

func (r *ProjectCommissionAccrualRepository) WithTx(tx bun.Tx) *ProjectCommissionAccrualRepository {
	return &ProjectCommissionAccrualRepository{db: tx}
}

func (r *ProjectCommissionAccrualRepository) Create(ctx context.Context, accrual *model.ProjectCommissionAccrual) error {
	_, err := r.db.NewInsert().Model(accrual).Exec(ctx)
	return err
}

// GetUnclaimed returns partner-unclaimed accruals for billing months up to claimableThrough.
// A billing month is unclaimed if no partner claim covers it for this project.
func (r *ProjectCommissionAccrualRepository) GetUnclaimed(ctx context.Context, projectID uuid.UUID, claimableThrough time.Time) ([]model.ProjectCommissionAccrual, error) {
	var accruals []model.ProjectCommissionAccrual
	err := r.db.NewSelect().
		Model(&accruals).
		Where("pca.project_id = ?", projectID).
		Where("pca.billing_month <= ?", claimableThrough.Format("2006-01-02")).
		Where(`NOT EXISTS (
			SELECT 1 FROM project_commission_claims pcc
			WHERE pcc.type = ?
			  AND pcc.project_id = pca.project_id
			  AND pcc.period_start <= pca.billing_month
			  AND pcc.period_end >= pca.billing_month
		)`, model.ClaimTypePartner).
		OrderExpr("pca.billing_month ASC").
		Scan(ctx)
	return accruals, err
}

// GetUnclaimedDD returns DD-unclaimed accruals for billing months up to claimableThrough.
// A billing month is unclaimed for DD if no dd_profit claim covers it across all projects.
func (r *ProjectCommissionAccrualRepository) GetUnclaimedDD(ctx context.Context, claimableThrough time.Time) ([]model.ProjectCommissionAccrual, error) {
	var accruals []model.ProjectCommissionAccrual
	err := r.db.NewSelect().
		Model(&accruals).
		Where("pca.billing_month <= ?", claimableThrough.Format("2006-01-02")).
		Where(`NOT EXISTS (
			SELECT 1 FROM project_commission_claims pcc
			WHERE pcc.type = ?
			  AND pcc.period_start <= pca.billing_month
			  AND pcc.period_end >= pca.billing_month
		)`, model.ClaimTypeDDProfit).
		OrderExpr("pca.billing_month ASC").
		Scan(ctx)
	return accruals, err
}

type MonthlyAccrualAggregate struct {
	Month    string // "2026-01"
	GrossUSD float64
}

func (r *ProjectCommissionAccrualRepository) GetMonthlyAggregates(ctx context.Context, projectID uuid.UUID) ([]MonthlyAccrualAggregate, error) {
	var rows []struct {
		Month    string  `bun:"month"`
		GrossUSD float64 `bun:"gross_usd"`
	}
	err := r.db.NewSelect().
		Model((*model.ProjectCommissionAccrual)(nil)).
		ColumnExpr("TO_CHAR(billing_month, 'YYYY-MM') AS month").
		ColumnExpr("SUM(commission_usd) AS gross_usd").
		Where("project_id = ?", projectID).
		GroupExpr("billing_month").
		OrderExpr("billing_month ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	result := make([]MonthlyAccrualAggregate, 0, len(rows))
	for _, r := range rows {
		result = append(result, MonthlyAccrualAggregate{Month: r.Month, GrossUSD: r.GrossUSD})
	}
	return result, nil
}

func (r *ProjectCommissionAccrualRepository) GetSummariesByProject(ctx context.Context, projectID uuid.UUID) ([]model.ProjectCommissionSummary, error) {
	var rows []struct {
		Symbol        string  `bun:"symbol"`
		CommissionRaw uint64  `bun:"commission_raw"`
		CommissionUSD float64 `bun:"commission_usd"`
	}

	err := r.db.NewSelect().
		Model((*model.ProjectCommissionAccrual)(nil)).
		ColumnExpr("symbol").
		ColumnExpr("SUM(commission_raw) AS commission_raw").
		ColumnExpr("SUM(commission_usd) AS commission_usd").
		Where("project_id = ?", projectID).
		GroupExpr("symbol").
		OrderExpr("symbol ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	var totalUSD float64
	for _, r := range rows {
		totalUSD += r.CommissionUSD
	}
	rate := model.PlatformCommissionRate(totalUSD)

	summaries := make([]model.ProjectCommissionSummary, 0, len(rows))
	for _, row := range rows {
		feeRaw := uint64(float64(row.CommissionRaw) * rate)
		summaries = append(summaries, model.ProjectCommissionSummary{
			ProjectID:      projectID,
			Symbol:         row.Symbol,
			CommissionRaw:  row.CommissionRaw,
			CommissionUSD:  row.CommissionUSD,
			PlatformRate:   rate,
			PlatformFeeRaw: feeRaw,
			PlatformFeeUSD: row.CommissionUSD * rate,
		})
	}

	return summaries, nil
}
