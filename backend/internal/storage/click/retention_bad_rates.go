package click

import (
	"context"
	"dd-prediction-api/internal/model"
	"fmt"
	"math"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
)

type RetentionBadRatesRepository struct {
	chDB *ch.DB
}

func NewRetentionBadRatesRepository(chDB *ch.DB) *RetentionBadRatesRepository {
	return &RetentionBadRatesRepository{
		chDB: chDB,
	}
}

func (r *RetentionBadRatesRepository) GetHealthMetrics(
	ctx context.Context,
	period model.MetricsPeriod,
) (*model.HealthDashboardMetrics, error) {

	currStart, currEnd := period.Current()
	prevStart, prevEnd := period.Previous()

	var result model.HealthDashboardMetrics

	// --- Retention 7 days ---
	newUsersCurr, err := r.countNewUsers(ctx, currStart, currEnd)
	if err != nil {
		return nil, fmt.Errorf("countNewUsers current failed: %w", err)
	}

	ret7CurrUsers, err := r.countRetention7Days(ctx, currStart, currEnd)
	if err != nil {
		return nil, fmt.Errorf("countRetention7Days failed: %w", err)
	}

	result.Retention7 = retentionMetric(ret7CurrUsers, newUsersCurr)

	// --- Retention 30 days ---
	ret30CurrUsers, err := r.countRetention30Days(ctx, currStart, currEnd)
	if err != nil {
		return nil, fmt.Errorf("countRetention30Days failed: %w", err)
	}
	result.Retention30 = retentionMetric(ret30CurrUsers, newUsersCurr)

	// --- Bounce rate (1 activity in period) ---
	bounceCurr, err := r.countBounceUsers(ctx, currStart, currEnd)
	if err != nil {
		return nil, fmt.Errorf("countBounceUsers failed: %w", err)
	}
	visitedCurr, err := r.countVisitedUsers(ctx, currStart, currEnd)
	if err != nil {
		return nil, fmt.Errorf("countVisitedUsers current failed: %w", err)
	}

	bouncePrev, _ := r.countBounceUsers(ctx, prevStart, prevEnd)
	visitedPrev, _ := r.countVisitedUsers(ctx, prevStart, prevEnd)
	result.BounceRate = badRateMetric(bounceCurr, visitedCurr, bouncePrev, visitedPrev)

	// --- Churn rate (active in previous, inactive in current) ---
	churnCurrUsers, prevActiveUsers, err := r.countChurnUsers(ctx, currStart, currEnd, prevStart, prevEnd)
	if err != nil {
		return nil, fmt.Errorf("countChurnUsers failed: %w", err)
	}

	var churnPrevUsers, prevPrevActiveUsers int64
	if !prevStart.IsZero() {
		prevPeriod := model.MetricsPeriod{
			Type:      "custom",
			StartDate: prevStart,
			EndDate:   prevEnd,
		}
		prevPrevStart, prevPrevEnd := prevPeriod.Previous()
		churnPrevUsers, prevPrevActiveUsers, _ = r.countChurnUsers(ctx, prevStart, prevEnd, prevPrevStart, prevPrevEnd)
	}

	result.ChurnRate = badRateMetric(churnCurrUsers, prevActiveUsers, churnPrevUsers, prevPrevActiveUsers)

	return &result, nil
}

func retentionPercent(retainedUsers, newUsers int64) int64 {
	if newUsers <= 0 || retainedUsers <= 0 {
		return 0
	}

	return int64(math.Round(float64(retainedUsers) * 100 / float64(newUsers)))
}

func retentionMetric(retainedUsers, newUsers int64) model.IntMetric {
	return model.IntMetric{
		Value:      retainedUsers,
		Growth:     fmt.Sprintf("%d", retentionPercent(retainedUsers, newUsers)),
		IsPositive: true,
	}
}

func ratePercent(numerator, denominator int64) int64 {
	if numerator <= 0 || denominator <= 0 {
		return 0
	}

	return int64(math.Round(float64(numerator) * 100 / float64(denominator)))
}

func badRateMetric(currNumerator, currDenominator, prevNumerator, prevDenominator int64) model.IntMetric {
	currRate := ratePercent(currNumerator, currDenominator)
	prevRate := ratePercent(prevNumerator, prevDenominator)
	diff := currRate - prevRate

	if diff < 0 {
		diff = -diff
	}

	return model.IntMetric{
		Value:      currRate,
		Growth:     fmt.Sprintf("%d", diff),
		IsPositive: currRate <= prevRate,
	}
}

func (r *RetentionBadRatesRepository) countNewUsers(ctx context.Context, start, end time.Time) (int64, error) {
	query := `
		SELECT count(*)
		FROM user_registrations
		WHERE registered_at >= ? AND registered_at < ?
	`
	var cnt int64
	err := r.chDB.QueryRowContext(ctx, query, start, end).Scan(&cnt)
	return cnt, err
}

func (r *RetentionBadRatesRepository) countVisitedUsers(ctx context.Context, start, end time.Time) (int64, error) {
	query := `
		SELECT count(DISTINCT user_id)
		FROM user_activities
		WHERE created_at >= ? AND created_at < ?
	`
	var cnt int64
	err := r.chDB.QueryRowContext(ctx, query, start, end).Scan(&cnt)
	return cnt, err
}

// Retention 7 days: >=10 activities in first 7 days after registration excluding day of registration
func (r *RetentionBadRatesRepository) countRetention7Days(ctx context.Context, start, end time.Time) (int64, error) {
	query := `
		WITH cohorts AS (
			SELECT user_id, registered_at AS reg_time
			FROM user_registrations
			WHERE registered_at >= ? AND registered_at < ?
		),
		active AS (
			SELECT c.user_id
			FROM cohorts c
			INNER JOIN user_activities a ON a.user_id = c.user_id
			WHERE a.created_at >= c.reg_time + INTERVAL 1 DAY
			  AND a.created_at < c.reg_time + INTERVAL 8 DAY
			GROUP BY c.user_id
			HAVING sum(a.activity_count) >= 10
		)
		SELECT count(*) FROM active
	`
	var cnt int64
	err := r.chDB.QueryRowContext(ctx, query, start, end).Scan(&cnt)
	return cnt, err
}

// Retention 30 days: active users in days 1-30 minus active users in days 1-7.
func (r *RetentionBadRatesRepository) countRetention30Days(ctx context.Context, start, end time.Time) (int64, error) {
	query := `
		WITH cohorts AS (
			SELECT user_id, registered_at AS reg_time
			FROM user_registrations
			WHERE registered_at >= ? AND registered_at < ?
		),
		active_1_7 AS (
			SELECT c.user_id
			FROM cohorts c
			INNER JOIN user_activities a ON a.user_id = c.user_id
			WHERE a.created_at >= c.reg_time + INTERVAL 1 DAY
			  AND a.created_at < c.reg_time + INTERVAL 8 DAY
			GROUP BY c.user_id
			HAVING sum(a.activity_count) >= 10
		),
		active_1_30 AS (
			SELECT c.user_id
			FROM cohorts c
			INNER JOIN user_activities a ON a.user_id = c.user_id
			WHERE a.created_at >= c.reg_time + INTERVAL 1 DAY
			  AND a.created_at < c.reg_time + INTERVAL 31 DAY
			GROUP BY c.user_id
			HAVING sum(a.activity_count) >= 10
		)
		SELECT count(DISTINCT a.user_id)
		FROM active_1_30 a
		LEFT JOIN active_1_7 b ON a.user_id = b.user_id
		WHERE b.user_id IS NULL
	`
	var cnt int64
	err := r.chDB.QueryRowContext(ctx, query, start, end).Scan(&cnt)
	return cnt, err
}

// Bounce rate: user with only 1 activity in the period
func (r *RetentionBadRatesRepository) countBounceUsers(ctx context.Context, start, end time.Time) (int64, error) {
	query := `
		SELECT count(DISTINCT user_id)
		FROM (
			SELECT user_id
			FROM user_activities
			WHERE created_at >= ? AND created_at < ?
			GROUP BY user_id
			HAVING sum(activity_count) = 1
		) AS bounce_users
	`
	var cnt int64
	err := r.chDB.QueryRowContext(ctx, query, start, end).Scan(&cnt)
	return cnt, err
}

func (r *RetentionBadRatesRepository) countActiveUsers(ctx context.Context, start, end time.Time) (int64, error) {
	query := `
		SELECT count(*)
		FROM (
			SELECT user_id
			FROM user_activities
			WHERE created_at >= ? AND created_at < ?
			GROUP BY user_id
			HAVING sum(activity_count) >= 10
		) AS active_users
	`
	var cnt int64
	err := r.chDB.QueryRowContext(ctx, query, start, end).Scan(&cnt)
	return cnt, err
}

// Churn rate: users active in previous period but inactive in current period.
func (r *RetentionBadRatesRepository) countChurnUsers(
	ctx context.Context,
	currStart, currEnd, prevStart, prevEnd time.Time,
) (int64, int64, error) {
	if prevStart.IsZero() {
		return 0, 0, nil
	}

	query := `
		SELECT count(DISTINCT prev.user_id)
		FROM (
			SELECT user_id
			FROM user_activities
			WHERE created_at >= ? AND created_at < ?
			GROUP BY user_id
			HAVING sum(activity_count) >= 10
		) AS prev
		LEFT JOIN (
			SELECT user_id
			FROM user_activities
			WHERE created_at >= ? AND created_at < ?
			GROUP BY user_id
			HAVING sum(activity_count) >= 10
		) AS curr ON prev.user_id = curr.user_id
		WHERE curr.user_id IS NULL
	`
	var churned int64
	err := r.chDB.QueryRowContext(ctx, query, prevStart, prevEnd, currStart, currEnd).Scan(&churned)
	if err != nil {
		return 0, 0, err
	}

	prevActiveUsers, err := r.countActiveUsers(ctx, prevStart, prevEnd)
	if err != nil {
		return 0, 0, err
	}

	return churned, prevActiveUsers, nil
}
