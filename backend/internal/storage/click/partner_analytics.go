package click

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/go-clickhouse/ch"

	"dd-prediction-api/internal/model"
)

type PartnerAnalyticsRepository struct {
	db *ch.DB
}

func NewPartnerAnalyticsRepository(db *ch.DB) *PartnerAnalyticsRepository {
	return &PartnerAnalyticsRepository{db: db}
}

// GetDailyIncome returns partner commission income aggregated by calendar day (USD).
// Uses tx_type = TransactionTypeDuelCommission (3) entries for the project.
func (r *PartnerAnalyticsRepository) GetDailyIncome(
	ctx context.Context,
	projectID uuid.UUID,
	from, to time.Time,
) ([]model.PartnerDailyIncome, error) {
	var rows []struct {
		Day    time.Time `ch:"day"`
		Amount float64   `ch:"amount_usd"`
	}

	err := r.db.NewSelect().
		TableExpr("duel_transactions").
		ColumnExpr("toDate(created_at) AS day").
		ColumnExpr("sum(amount) AS amount_usd").
		Where("project_id = ?", projectID).
		Where("tx_type = ?", model.TransactionTypeDuelCommission).
		Where("created_at >= ?", from).
		Where("created_at < ?", to).
		GroupExpr("day").
		OrderExpr("day ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("get daily income: %w", err)
	}

	result := make([]model.PartnerDailyIncome, 0, len(rows))
	for _, r := range rows {
		result = append(result, model.PartnerDailyIncome{
			Day:    r.Day.Format("2006-01-02"),
			Amount: r.Amount,
		})
	}
	return result, nil
}

// GetMonthlyActiveUsers returns the count of unique users who placed at least one
// prediction (tx_type = TransactionTypeDuelPrediction) per calendar month.
func (r *PartnerAnalyticsRepository) GetMonthlyActiveUsers(
	ctx context.Context,
	projectID uuid.UUID,
) ([]model.PartnerMonthlyMAU, error) {
	var rows []struct {
		Month       string `ch:"month"`
		ActiveUsers uint64 `ch:"active_users"`
	}

	err := r.db.NewSelect().
		TableExpr("duel_transactions").
		ColumnExpr("formatDateTime(toStartOfMonth(created_at), '%Y-%m') AS month").
		ColumnExpr("uniqExact(user_id) AS active_users").
		Where("project_id = ?", projectID).
		Where("tx_type = ?", model.TransactionTypeDuelPrediction).
		GroupExpr("month").
		OrderExpr("month ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("get monthly active users: %w", err)
	}

	result := make([]model.PartnerMonthlyMAU, 0, len(rows))
	for _, r := range rows {
		result = append(result, model.PartnerMonthlyMAU{
			Month:       r.Month,
			ActiveUsers: r.ActiveUsers,
		})
	}
	return result, nil
}

// GetMonthlyVolume returns the total duel volume (sum of prediction amounts) per calendar month.
func (r *PartnerAnalyticsRepository) GetMonthlyVolume(
	ctx context.Context,
	projectID uuid.UUID,
) ([]struct {
	Month  string
	Volume float64
}, error) {
	var rows []struct {
		Month  string  `ch:"month"`
		Volume float64 `ch:"volume"`
	}

	err := r.db.NewSelect().
		TableExpr("duel_transactions").
		ColumnExpr("formatDateTime(toStartOfMonth(created_at), '%Y-%m') AS month").
		ColumnExpr("sum(amount) AS volume").
		Where("project_id = ?", projectID).
		Where("tx_type = ?", model.TransactionTypeDuelPrediction).
		GroupExpr("month").
		OrderExpr("month ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("get monthly volume: %w", err)
	}

	result := make([]struct {
		Month  string
		Volume float64
	}, 0, len(rows))
	for _, r := range rows {
		result = append(result, struct {
			Month  string
			Volume float64
		}{Month: r.Month, Volume: r.Volume})
	}
	return result, nil
}

// GetCurrentMonthVolume returns total prediction volume for the current calendar month.
func (r *PartnerAnalyticsRepository) GetCurrentMonthVolume(
	ctx context.Context,
	projectID uuid.UUID,
) (float64, error) {
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	var result struct {
		Volume float64 `ch:"volume"`
	}
	err := r.db.NewSelect().
		TableExpr("duel_transactions").
		ColumnExpr("sum(amount) AS volume").
		Where("project_id = ?", projectID).
		Where("tx_type = ?", model.TransactionTypeDuelPrediction).
		Where("created_at >= ?", monthStart).
		Scan(ctx, &result)
	if err != nil {
		return 0, fmt.Errorf("get current month volume: %w", err)
	}
	return result.Volume, nil
}
