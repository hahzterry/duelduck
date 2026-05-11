package repository

import (
	"context"
	"math"
	"time"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/pkg/mtype"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type MetricsRepository struct {
	db *bun.DB
}

func NewMetricsRepository(db *bun.DB) *MetricsRepository {
	return &MetricsRepository{db: db}
}

func (r *MetricsRepository) GetMainDashboardMetrics(
	ctx context.Context,
	period model.MetricsPeriod,
) (*model.MainDashboardMetrics, error) {
	currStart, currEnd := period.Current()
	prevStart, prevEnd := period.Previous()

	var m model.MainDashboardMetrics

	if err := r.fillTotalMetrics(ctx, currStart, currEnd, &m.Total); err != nil {
		return nil, err
	}
	if err := r.fillGrowthMetrics(ctx, currStart, currEnd, prevStart, prevEnd, &m.Growth); err != nil {
		return nil, err
	}

	if err := r.fillPredictionsMetrics(ctx, currStart, currEnd, prevStart, prevEnd, &m.Predictions); err != nil {
		return nil, err
	}

	m.UpdatedAt = time.Now().UTC()

	return &m, nil
}

func (r *MetricsRepository) fillTotalMetrics(
	ctx context.Context,
	periodStart, periodEnd time.Time,
	dst *model.TotalMetrics,
) error {
	// Registrations
	totalUsers, err := r.countTotalUsers(ctx)
	if err != nil {
		return err
	}
	dst.Registrations.Value = totalUsers

	newThisPeriod, err := r.countNewUsers(ctx, periodStart, periodEnd)
	if err != nil {
		return err
	}
	beforeUsers, _ := r.countNewUsers(ctx, time.Time{}, periodStart)
	dst.Registrations.Growth, dst.Registrations.IsPositive = model.CalculateTotalGrowthPercent(
		float64(newThisPeriod), float64(beforeUsers),
	)

	// Active Users
	everActive, err := r.countEverActiveUsers(ctx)
	if err != nil {
		return err
	}
	dst.ActiveUsers.Value = everActive

	activeThisPeriod, err := r.countDistinctActiveUsers(ctx, periodStart, periodEnd)
	if err != nil {
		return err
	}
	beforeActive, _ := r.countDistinctActiveUsers(ctx, time.Time{}, periodStart)
	dst.ActiveUsers.Growth, dst.ActiveUsers.IsPositive = model.CalculateTotalGrowthPercent(
		float64(activeThisPeriod), float64(beforeActive),
	)

	// Creators
	everCreators, err := r.countEverCreators(ctx)
	if err != nil {
		return err
	}
	dst.Creators.Value = everCreators

	creatorsThisPeriod, err := r.countDistinctCreators(ctx, periodStart, periodEnd)
	if err != nil {
		return err
	}
	beforeCreators, _ := r.countDistinctCreators(ctx, time.Time{}, periodStart)
	dst.Creators.Growth, dst.Creators.IsPositive = model.CalculateTotalGrowthPercent(
		float64(creatorsThisPeriod), float64(beforeCreators),
	)

	return nil
}

func (r *MetricsRepository) fillGrowthMetrics(
	ctx context.Context,
	currStart, currEnd, prevStart, prevEnd time.Time,

	dst *model.GrowthMetrics,
) error {
	currNew, _ := r.countNewUsers(ctx, currStart, currEnd)
	dst.NewUsers = model.IntMetric{
		Value: currNew,
	}

	currActive, _ := r.countDistinctActiveUsers(ctx, currStart, currEnd)
	prevActive, _ := r.countDistinctActiveUsers(ctx, prevStart, prevEnd)
	growthActive, isPosActive := model.CalculateGrowthPercent(float64(currActive), float64(prevActive))
	dst.ActiveUsers = model.IntMetric{
		Value: currActive,
	}
	dst.ActiveUserGrowth = model.PercentChangeMetric{
		Value:      growthActive,
		IsPositive: isPosActive,
	}

	currCreators, _ := r.countDistinctCreators(ctx, currStart, currEnd)
	dst.Creators = model.IntMetric{
		Value: currCreators,
	}

	return nil
}

func (r *MetricsRepository) fillPredictionsMetrics(
	ctx context.Context,
	currStart, currEnd, prevStart, prevEnd time.Time,

	dst *model.PredictionsMetrics,
) error {
	currVotes, _ := r.countVotes(ctx, currStart, currEnd)
	prevVotes, _ := r.countVotes(ctx, prevStart, prevEnd)
	growthVotes, isPosVotes := model.CalculateGrowthPercent(float64(currVotes), float64(prevVotes))
	dst.TotalVotes = model.IntMetric{
		Value:      currVotes,
		Growth:     growthVotes,
		IsPositive: isPosVotes,
	}

	currPerActive, _ := r.avgPredictionsPerActiveUser(ctx, currStart, currEnd)
	prevPerActive, _ := r.avgPredictionsPerActiveUser(ctx, prevStart, prevEnd)
	growthPerActive, isPosActive := model.CalculateGrowthPercent(currPerActive, prevPerActive)
	dst.PerActiveUser = model.FloatMetric{
		Value:      math.Round(currPerActive*100) / 100,
		Growth:     growthPerActive,
		IsPositive: isPosActive,
	}

	currPerCreator, _ := r.avgDuelsPerCreator(ctx, currStart, currEnd)
	prevPerCreator, _ := r.avgDuelsPerCreator(ctx, prevStart, prevEnd)
	growthPerCreator, isPosCreator := model.CalculateGrowthPercent(currPerCreator, prevPerCreator)
	dst.PerCreator = model.FloatMetric{
		Value:      math.Round(currPerCreator*100) / 100,
		Growth:     growthPerCreator,
		IsPositive: isPosCreator,
	}

	currTVL, _ := r.totalVotesVolume(ctx, currStart, currEnd)
	prevTVL, _ := r.totalVotesVolume(ctx, prevStart, prevEnd)
	growthTVL, isPosTVL := model.CalculateGrowthPercent(currTVL, prevTVL)
	dst.TotalTVL = model.FloatMetric{
		Value:      math.Round(currTVL*100) / 100,
		Growth:     growthTVL,
		IsPositive: isPosTVL,
	}

	return nil
}

func (r *MetricsRepository) countTotalUsers(ctx context.Context) (int64, error) {
	count, err := r.db.NewSelect().Model((*model.User)(nil)).Count(ctx)
	return int64(count), err
}

func (r *MetricsRepository) countNewUsers(ctx context.Context, start, end time.Time) (int64, error) {
	q := r.db.NewSelect().Model((*model.User)(nil)).ColumnExpr("COUNT(*)")

	if !start.IsZero() {
		q = q.Where("created_at >= ?", start)
	}
	if !end.IsZero() {
		q = q.Where("created_at < ?", end)
	}

	var cnt int64
	err := q.Scan(ctx, &cnt)
	return cnt, err
}

func (r *MetricsRepository) countEverActiveUsers(ctx context.Context) (int64, error) {
	q := r.db.NewSelect().
		Model((*model.Player)(nil)).
		ColumnExpr("COUNT(DISTINCT user_id)").
		Join("INNER JOIN duels d ON d.id = players.duel_id")

	var cnt int64
	err := q.Scan(ctx, &cnt)
	return cnt, err
}

func (r *MetricsRepository) countDistinctActiveUsers(
	ctx context.Context,
	start, end time.Time,
) (int64, error) {
	q := r.db.NewSelect().
		Model((*model.Player)(nil)).
		ColumnExpr("COUNT(DISTINCT user_id)").
		Where("players.created_at >= ? AND players.created_at < ?", start, end).
		Join("INNER JOIN duels d ON d.id = players.duel_id")

	var cnt int64
	err := q.Scan(ctx, &cnt)
	return cnt, err
}

func (r *MetricsRepository) countVotes(
	ctx context.Context,
	start, end time.Time,
) (int64, error) {
	q := r.db.NewSelect().
		Model((*model.Player)(nil)).
		ColumnExpr("COUNT(*)").
		Where("players.created_at >= ? AND players.created_at < ?", start, end).
		Join("INNER JOIN duels d ON d.id = players.duel_id")

	var cnt int64
	err := q.Scan(ctx, &cnt)
	return cnt, err
}

func (r *MetricsRepository) countEverCreators(ctx context.Context) (int64, error) {
	q := r.db.NewSelect().
		Model((*model.Duel)(nil)).
		ColumnExpr("COUNT(DISTINCT owner_id)")

	var cnt int64
	err := q.Scan(ctx, &cnt)
	return cnt, err
}

func (r *MetricsRepository) countDistinctCreators(
	ctx context.Context,
	start, end time.Time,
) (int64, error) {
	q := r.db.NewSelect().
		Model((*model.Duel)(nil)).
		ColumnExpr("COUNT(DISTINCT owner_id)").
		Where("created_at >= ? AND created_at < ?", start, end)

	var cnt int64
	err := q.Scan(ctx, &cnt)
	return cnt, err
}

func (r *MetricsRepository) avgPredictionsPerNewUser(ctx context.Context, start, end time.Time) (float64, error) {
	var result struct {
		NewUsers    int64 `bun:"new_users"`
		Predictions int64 `bun:"predictions"`
	}

	q := r.db.NewSelect().
		Model((*model.User)(nil)).
		ColumnExpr("COUNT(DISTINCT u.id) AS new_users").
		ColumnExpr("COUNT(p.id) AS predictions").
		Join("LEFT JOIN players p ON p.user_id = u.id AND p.created_at >= ? AND p.created_at < ?", start, end).
		Where("u.created_at >= ? AND u.created_at < ?", start, end)

	err := q.Scan(ctx, &result)
	if err != nil {
		return 0, err
	}

	if result.NewUsers == 0 {
		return 0, nil
	}
	return float64(result.Predictions) / float64(result.NewUsers), nil
}

func (r *MetricsRepository) avgPredictionsPerActiveUser(
	ctx context.Context,
	start, end time.Time,
) (float64, error) {
	var res struct {
		ActiveUsers int64 `bun:"active_users"`
		Predictions int64 `bun:"predictions"`
	}

	q := r.db.NewSelect().
		Model((*model.Player)(nil)).
		ColumnExpr("COUNT(DISTINCT user_id) AS active_users").
		ColumnExpr("COUNT(*) AS predictions").
		Where("players.created_at >= ? AND players.created_at < ?", start, end).
		Join("INNER JOIN duels d ON d.id = players.duel_id")

	err := q.Scan(ctx, &res)
	if err != nil {
		return 0, err
	}
	if res.ActiveUsers == 0 {
		return 0, nil
	}
	return float64(res.Predictions) / float64(res.ActiveUsers), nil
}

func (r *MetricsRepository) avgDuelsPerCreator(
	ctx context.Context,
	start, end time.Time,
) (float64, error) {
	var res struct {
		Creators int64 `bun:"creators"`
		Duels    int64 `bun:"duels"`
	}

	q := r.db.NewSelect().
		Model((*model.Duel)(nil)).
		ColumnExpr("COUNT(DISTINCT owner_id) AS creators").
		ColumnExpr("COUNT(*) AS duels").
		Where("created_at >= ? AND created_at < ?", start, end)

	err := q.Scan(ctx, &res)
	if err != nil {
		return 0, err
	}
	if res.Creators == 0 {
		return 0, nil
	}
	return float64(res.Duels) / float64(res.Creators), nil
}

func (r *MetricsRepository) totalVotesVolume(
	ctx context.Context,
	start, end time.Time,
) (float64, error) {
	priceExpr := duelPriceExpr()

	q := r.db.NewSelect().
		TableExpr("players AS p").
		Join("INNER JOIN duels d ON d.id = p.duel_id").
		Where("p.created_at >= ? AND p.created_at < ?", start, end).
		ColumnExpr("COALESCE(SUM(" + priceExpr + "), 0)")

	var volume float64
	if err := q.Scan(ctx, &volume); err != nil {
		return 0, err
	}

	return volume, nil
}

type UserRegistrationRow struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

func (r *MetricsRepository) GetUsersRegisteredAfter(
	ctx context.Context,
	from time.Time,
	limit int,
	offset int,
) ([]UserRegistrationRow, error) {

	rows := make([]UserRegistrationRow, 0)

	err := r.db.NewSelect().
		Table("users").
		Column("id", "created_at").
		Where("created_at >= ?", from).
		OrderExpr("created_at ASC").
		Limit(limit).
		Offset(offset).
		Scan(ctx, &rows)

	return rows, err
}

type RevenueDuelsVoteMetrics struct {
	Votes       int64   `bun:"predictions"`
	Volume      float64 `bun:"volume"`
	PlatformFee float64 `bun:"platform_fee"`
}

type RevenueDuelsDuelMetrics struct {
	Duels       int64   `bun:"duels"`
	TotalPrice  float64 `bun:"total_price"`
	TotalFeePct float64 `bun:"total_fee_pct"`
}

func duelPriceExpr() string {
	return "COALESCE(d.usd_price, 0)"
}

func (r *MetricsRepository) GetRevenueDuelsVoteMetrics(
	ctx context.Context,
	start, end time.Time,
) (*RevenueDuelsVoteMetrics, error) {
	result := new(RevenueDuelsVoteMetrics)

	priceExpr := duelPriceExpr()

	q := r.db.NewSelect().
		TableExpr("players AS p").
		Join("INNER JOIN duels d ON d.id = p.duel_id").
		Join("INNER JOIN users u ON u.id = d.owner_id").
		Where("p.created_at >= ? AND p.created_at < ?", start, end).
		ColumnExpr("COUNT(p.id) AS predictions").
		ColumnExpr("COALESCE(SUM("+priceExpr+"), 0) AS volume").
		ColumnExpr(
			"COALESCE(SUM(CASE WHEN u.role IN (?) THEN "+priceExpr+" * d.commission / 100.0 ELSE "+priceExpr+" * d.commission / 200.0 END), 0) AS platform_fee",
			bun.In([]uint8{mtype.RolePartnerAdmin, mtype.RoleAdmin}),
		)

	if err := q.Scan(ctx, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *MetricsRepository) GetRevenueDuelsDuelMetrics(
	ctx context.Context,
	start, end time.Time,

) (*RevenueDuelsDuelMetrics, error) {
	result := new(RevenueDuelsDuelMetrics)

	priceExpr := duelPriceExpr()

	q := r.db.NewSelect().
		TableExpr("duels AS d").
		Where("d.created_at >= ? AND d.created_at < ?", start, end).
		ColumnExpr("COUNT(d.id) AS duels").
		ColumnExpr("COALESCE(SUM(" + priceExpr + "), 0)::DOUBLE PRECISION AS total_price").
		ColumnExpr("COALESCE(SUM(d.commission), 0)::DOUBLE PRECISION AS total_fee_pct")

	if err := q.Scan(ctx, result); err != nil {
		return nil, err
	}

	return result, nil
}

type DuelsStatusAggregates struct {
	ActiveCount  int64   `bun:"active_count"`
	PendingCount int64   `bun:"pending_count"`
	ClosedCount  int64   `bun:"closed_count"`
	RefundCount  int64   `bun:"refund_count"`
	ActiveTVL    float64 `bun:"active_tvl"`
	PendingTVL   float64 `bun:"pending_tvl"`
	ClosedTVL    float64 `bun:"closed_tvl"`
	RefundTVL    float64 `bun:"refund_tvl"`
	ActiveFee    float64 `bun:"active_fee"`
	PendingFee   float64 `bun:"pending_fee"`
	ClosedFee    float64 `bun:"closed_fee"`
	RefundFee    float64 `bun:"refund_fee"`
}

type duelStateMetrics struct {
	Count int64   `bun:"cnt"`
	TVL   float64 `bun:"tvl"`
	Fee   float64 `bun:"fee"`
}

func (r *MetricsRepository) getStableStatusMetrics(
	ctx context.Context,
	start, end time.Time,

	statuses ...uint8,
) (*duelStateMetrics, error) {
	result := new(duelStateMetrics)
	priceExpr := duelPriceExpr()

	lastStatusBeforeStartSubQuery := r.db.NewSelect().
		TableExpr("duel_status_history").
		ColumnExpr("DISTINCT ON (duel_id) duel_id, status").
		Where("changed_at < ?", start).
		OrderExpr("duel_id, changed_at DESC, id DESC")

	changedInPeriodSubQuery := r.db.NewSelect().
		TableExpr("duel_status_history").
		ColumnExpr("DISTINCT duel_id").
		Where("changed_at >= ? AND changed_at < ?", start, end).
		OrderExpr("duel_id")

	q := r.db.NewSelect().
		TableExpr("duels AS d").
		Join("INNER JOIN (?) ls ON ls.duel_id = d.id", lastStatusBeforeStartSubQuery).
		Join("LEFT JOIN (?) cp ON cp.duel_id = d.id", changedInPeriodSubQuery).
		Where("ls.status IN (?)", bun.In(statuses)).
		Where("cp.duel_id IS NULL").
		Where("d.created_at < ?", start).
		ColumnExpr("COUNT(*) AS cnt").
		ColumnExpr("COALESCE(SUM(" + priceExpr + " * d.players_count), 0) AS tvl").
		ColumnExpr("COALESCE(SUM(" + priceExpr + " * d.players_count * d.commission / 100.0), 0) AS fee")

	if err := q.Scan(ctx, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *MetricsRepository) getPendingMetrics(
	ctx context.Context,
	start, end time.Time,

) (*duelStateMetrics, error) {
	result := new(duelStateMetrics)
	priceExpr := duelPriceExpr()

	lastStatusBeforeStartSubQuery := r.db.NewSelect().
		TableExpr("duel_status_history").
		ColumnExpr("DISTINCT ON (duel_id) duel_id, status").
		Where("changed_at < ?", start).
		OrderExpr("duel_id, changed_at DESC, id DESC")

	changedInPeriodSubQuery := r.db.NewSelect().
		TableExpr("duel_status_history").
		ColumnExpr("DISTINCT duel_id").
		Where("changed_at >= ? AND changed_at < ?", start, end).
		OrderExpr("duel_id")

	q := r.db.NewSelect().
		TableExpr("duels AS d").
		Join("INNER JOIN (?) ls ON ls.duel_id = d.id", lastStatusBeforeStartSubQuery).
		Join("LEFT JOIN (?) cp ON cp.duel_id = d.id", changedInPeriodSubQuery).
		Where("d.created_at < ?", start).
		Where("ls.status = ?", model.DuelStatusActive).
		Where("cp.duel_id IS NULL").
		Where("d.deadline < ?", start).
		ColumnExpr("COUNT(*) AS cnt").
		ColumnExpr("COALESCE(SUM(" + priceExpr + " * d.players_count), 0) AS tvl").
		ColumnExpr("COALESCE(SUM(" + priceExpr + " * d.players_count * d.commission / 100.0), 0) AS fee")

	if err := q.Scan(ctx, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *MetricsRepository) GetDuelsStatusAggregates(
	ctx context.Context,
	start, end time.Time,

) (*DuelsStatusAggregates, error) {
	active, err := r.getStableStatusMetrics(
		ctx, start, end, model.DuelStatusActive,
	)
	if err != nil {
		return nil, err
	}

	closed, err := r.getStableStatusMetrics(
		ctx, start, end, model.DuelStatusResolved,
	)
	if err != nil {
		return nil, err
	}

	refund, err := r.getStableStatusMetrics(
		ctx,
		start,
		end,
		model.DuelStatusRefunded,
		model.DuelStatusDenied,
	)
	if err != nil {
		return nil, err
	}

	pending, err := r.getPendingMetrics(ctx, start, end)
	if err != nil {
		return nil, err
	}

	return &DuelsStatusAggregates{
		ActiveCount:  active.Count,
		PendingCount: pending.Count,
		ClosedCount:  closed.Count,
		RefundCount:  refund.Count,
		ActiveTVL:    active.TVL,
		PendingTVL:   pending.TVL,
		ClosedTVL:    closed.TVL,
		RefundTVL:    refund.TVL,
		ActiveFee:    active.Fee,
		PendingFee:   pending.Fee,
		ClosedFee:    closed.Fee,
		RefundFee:    refund.Fee,
	}, nil
}
