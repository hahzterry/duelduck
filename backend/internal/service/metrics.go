package service

import (
	"context"
	"errors"
	"math"
	"strconv"
	"time"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/storage/click"
	"dd-prediction-api/internal/storage/repository"
)

var (
	ErrInvalidFilterType = errors.New("invalid filter type")
	ErrInvalidDateFormat = errors.New("invalid date format")
	ErrInvalidMonth      = errors.New("invalid month")
	ErrStartAfterEnd     = errors.New("start date must be before end date")
	ErrFutureDate        = errors.New("dates cannot be in the future")
)

type MetricsFilter struct {
	Type     string `form:"type"`  // "all_time", "month", "custom"
	Month    string `form:"month"` // "January", "February", ...
	Year     string `form:"year"`  // "2026"
	StartStr string `form:"start"` // "01/01/2026" for custom
	EndStr   string `form:"end"`   // "31/01/2026" for custom
}

type MetricsService struct {
	metricsRepo                 *repository.MetricsRepository
	userRegistrationRepository  *click.UserRegistrationRepository
	retentionBadRatesRepository *click.RetentionBadRatesRepository
	tokenPriceRepo              *click.TokenPriceRepository
	coinService                 *CoinService
}

func NewMetricsService(
	metricsRepo *repository.MetricsRepository,
	clickRepo *click.UserRegistrationRepository,
	retentionBadRatesRepo *click.RetentionBadRatesRepository,
	tokenPriceRepo *click.TokenPriceRepository,
	coinService *CoinService,
) *MetricsService {
	return &MetricsService{
		metricsRepo:                 metricsRepo,
		userRegistrationRepository:  clickRepo,
		retentionBadRatesRepository: retentionBadRatesRepo,
		tokenPriceRepo:              tokenPriceRepo,
		coinService:                 coinService,
	}
}

// ------------------------------ MAIN ------------------------------

func (s *MetricsService) GetMainDashboardMetrics(
	ctx context.Context,
	filter *MetricsFilter,
) (*model.MainDashboardMetrics, error) {

	period, err := s.parsePeriod(filter)
	if err != nil {
		return nil, err
	}

	metrics, err := s.metricsRepo.GetMainDashboardMetrics(ctx, period)
	if err != nil {
		return nil, err
	}

	metrics.Period = period.ToJSONStruct()
	metrics.UpdatedAt = time.Now().UTC()

	if period.Type == "all_time" {
		zero := "0"

		metrics.Total.Registrations.Growth, metrics.Total.Registrations.IsPositive = zero, true
		metrics.Total.ActiveUsers.Growth, metrics.Total.ActiveUsers.IsPositive = zero, true
		metrics.Total.Creators.Growth, metrics.Total.Creators.IsPositive = zero, true

		metrics.Growth.NewUsers.Growth, metrics.Growth.NewUsers.IsPositive = zero, true
		metrics.Growth.ActiveUsers.Growth, metrics.Growth.ActiveUsers.IsPositive = zero, true
		metrics.Growth.Creators.Growth, metrics.Growth.Creators.IsPositive = zero, true
		metrics.Growth.ActiveUserGrowth.Value, metrics.Growth.ActiveUserGrowth.IsPositive = zero, true

		metrics.Predictions.TotalVotes.Growth, metrics.Predictions.TotalVotes.IsPositive = zero, true
		metrics.Predictions.PerActiveUser.Growth, metrics.Predictions.PerActiveUser.IsPositive = zero, true
		metrics.Predictions.PerCreator.Growth, metrics.Predictions.PerCreator.IsPositive = zero, true
		metrics.Predictions.TotalTVL.Growth, metrics.Predictions.TotalTVL.IsPositive = zero, true
	}

	return metrics, nil
}

func (s *MetricsService) GetPlatformPerformanceMetrics(
	ctx context.Context,
	filter *MetricsFilter,
) (*model.PlatformPerformanceMetrics, error) {
	mainMetrics, err := s.GetMainDashboardMetrics(ctx, filter)
	if err != nil {
		return nil, err
	}

	revenueMetrics, err := s.GetRevenueDashboardMetrics(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &model.PlatformPerformanceMetrics{
		Period:      mainMetrics.Period,
		Users:       mainMetrics.Growth,
		Predictions: mainMetrics.Predictions,
		Revenue:     revenueMetrics.Duels,
		UpdatedAt:   time.Now().UTC(),
	}, nil
}

// ------------------------------ HEALTH ------------------------------

func (s *MetricsService) GetHealthDashboardMetrics(
	ctx context.Context,
	filter *MetricsFilter,
) (*model.HealthDashboardMetrics, error) {

	period, err := s.parsePeriod(filter)
	if err != nil {
		return nil, err
	}

	metrics, err := s.retentionBadRatesRepository.GetHealthMetrics(ctx, period)
	if err != nil {
		return nil, err
	}

	metrics.Period = period.ToJSONStruct()
	metrics.UpdatedAt = time.Now().UTC()

	if period.Type == "all_time" {
		zero := "0"
		metrics.BounceRate.Growth, metrics.BounceRate.IsPositive = zero, true
		metrics.ChurnRate.Growth, metrics.ChurnRate.IsPositive = zero, true
	}

	return metrics, nil
}

// ------------------------------ REVENUE ------------------------------

func (s *MetricsService) GetRevenueDashboardMetrics(
	ctx context.Context,
	filter *MetricsFilter,
) (*model.RevenueDashboardMetrics, error) {

	period, err := s.parsePeriod(filter)
	if err != nil {
		return nil, err
	}

	currStart, currEnd := period.Current()
	prevStart, prevEnd := period.Previous()

	curr, err := s.calcDuels(ctx, currStart, currEnd)
	if err != nil {
		return nil, err
	}
	prev, err := s.calcDuels(ctx, prevStart, prevEnd)
	if err != nil {
		return nil, err
	}

	var res model.RevenueDashboardMetrics
	res.Period = period.ToJSONStruct()
	res.UpdatedAt = time.Now().UTC()

	if period.Type == "all_time" {
		res.Duels.Revenue = zeroGrowthOrNA(curr.RevenueUSD, curr.RevenueOK)
		res.Duels.AOV = zeroGrowthOrNA(curr.AOVUSD, curr.AOVOK)
		res.Duels.AvgPredictionPrice = zeroGrowthOrNA(curr.AvgPredictionPriceUSD, curr.AvgPredictionPriceOK)
		res.Duels.AvgFeePerPrediction = zeroGrowthOrNA(curr.AvgFeePerPredictionPct, curr.AvgFeePctOK)
		return &res, nil
	}

	res.Duels.Revenue = floatMetricOrNA(curr.RevenueUSD, curr.RevenueOK, prev.RevenueUSD, prev.RevenueOK)
	res.Duels.AOV = floatMetricOrNA(curr.AOVUSD, curr.AOVOK, prev.AOVUSD, prev.AOVOK)
	res.Duels.AvgPredictionPrice = floatMetricOrNA(curr.AvgPredictionPriceUSD, curr.AvgPredictionPriceOK, prev.AvgPredictionPriceUSD, prev.AvgPredictionPriceOK)
	res.Duels.AvgFeePerPrediction = floatMetricOrNA(curr.AvgFeePerPredictionPct, curr.AvgFeePctOK, prev.AvgFeePerPredictionPct, prev.AvgFeePctOK)

	return &res, nil
}

type duelsCalc struct {
	RevenueUSD             float64
	AvgPredictionPriceUSD  float64
	AvgFeePerPredictionPct float64
	AOVUSD                 float64

	RevenueOK            bool
	AvgPredictionPriceOK bool
	AvgFeePctOK          bool
	AOVOK                bool
}

func naFloatMetric() model.FloatMetric {
	return model.FloatMetric{
		Value:      0,
		Growth:     "n/a",
		IsPositive: false,
	}
}

func intMetric(curr, prev int64) model.IntMetric {
	g, pos := model.CalculateGrowthPercent(float64(curr), float64(prev))
	return model.IntMetric{Value: curr, Growth: g, IsPositive: pos}
}

func zeroGrowthIntMetric(value int64) model.IntMetric {
	return model.IntMetric{Value: value, Growth: "0", IsPositive: true}
}

func floatMetricOrNA(curr float64, currOK bool, prev float64, prevOK bool) model.FloatMetric {
	if !currOK || !prevOK {
		return naFloatMetric()
	}
	g, pos := model.CalculateGrowthPercent(curr, prev)
	return model.FloatMetric{Value: math.Round(curr*100) / 100, Growth: g, IsPositive: pos}
}

func zeroGrowthOrNA(value float64, ok bool) model.FloatMetric {
	if !ok {
		return naFloatMetric()
	}
	return model.FloatMetric{
		Value:      math.Round(value*100) / 100,
		Growth:     "0",
		IsPositive: true,
	}
}

func averagePerDuel(total float64, duels int64) float64 {
	if duels == 0 {
		return 0
	}

	return total / float64(duels)
}

func averagePerVote(volume float64, votes int64) float64 {
	if votes == 0 {
		return 0
	}

	return volume / float64(votes)
}

func calculateDuelsRevenueMetrics(
	voteMetrics *repository.RevenueDuelsVoteMetrics,
	duelMetrics *repository.RevenueDuelsDuelMetrics,
) duelsCalc {
	return duelsCalc{
		RevenueUSD:             voteMetrics.PlatformFee,
		AvgPredictionPriceUSD:  averagePerDuel(duelMetrics.TotalPrice, duelMetrics.Duels),
		AvgFeePerPredictionPct: averagePerDuel(duelMetrics.TotalFeePct, duelMetrics.Duels),
		AOVUSD:                 averagePerVote(voteMetrics.Volume, voteMetrics.Votes),

		RevenueOK:            true,
		AvgPredictionPriceOK: true,
		AvgFeePctOK:          true,
		AOVOK:                true,
	}
}

func (s *MetricsService) calcDuels(
	ctx context.Context,
	start, end time.Time,
) (duelsCalc, error) {
	voteMetrics, err := s.metricsRepo.GetRevenueDuelsVoteMetrics(ctx, start, end)
	if err != nil {
		return duelsCalc{}, err
	}

	duelMetrics, err := s.metricsRepo.GetRevenueDuelsDuelMetrics(ctx, start, end)
	if err != nil {
		return duelsCalc{}, err
	}

	return calculateDuelsRevenueMetrics(voteMetrics, duelMetrics), nil
}

// ------------------------------ DUELS ------------------------------

func (s *MetricsService) GetDuelsDashboardMetrics(
	ctx context.Context,
	filter *MetricsFilter,
) (*model.DuelsDashboardMetrics, error) {

	period, err := s.parsePeriod(filter)
	if err != nil {
		return nil, err
	}

	currStart, currEnd := period.Current()
	prevStart, prevEnd := period.Previous()

	curr, err := s.metricsRepo.GetDuelsStatusAggregates(ctx, currStart, currEnd)
	if err != nil {
		return nil, err
	}
	prev, err := s.metricsRepo.GetDuelsStatusAggregates(ctx, prevStart, prevEnd)
	if err != nil {
		return nil, err
	}

	var res model.DuelsDashboardMetrics

	if period.Type == "all_time" {
		res.DuelsStatusMetrics.ActiveDuels = zeroGrowthIntMetric(curr.ActiveCount)
		res.DuelsStatusMetrics.PendingDuels = zeroGrowthIntMetric(curr.PendingCount)
		res.DuelsStatusMetrics.ClosedDuels = zeroGrowthIntMetric(curr.ClosedCount)
		res.DuelsStatusMetrics.RefundDuels = zeroGrowthIntMetric(curr.RefundCount)

		res.DuelsTVLMetrics.ActiveDuels = zeroGrowthOrNA(curr.ActiveTVL, true)
		res.DuelsTVLMetrics.PendingDuels = zeroGrowthOrNA(curr.PendingTVL, true)
		res.DuelsTVLMetrics.ClosedDuels = zeroGrowthOrNA(curr.ClosedTVL, true)
		res.DuelsTVLMetrics.RefundDuels = zeroGrowthOrNA(curr.RefundTVL, true)

		res.DuelsRevenueMetrics.ActiveDuels = zeroGrowthOrNA(curr.ActiveFee, true)
		res.DuelsRevenueMetrics.PendingDuels = zeroGrowthOrNA(curr.PendingFee, true)
		res.DuelsRevenueMetrics.ClosedDuels = zeroGrowthOrNA(curr.ClosedFee, true)
		res.DuelsRevenueMetrics.RefundDuels = zeroGrowthOrNA(-1*curr.RefundFee, true)
		return &res, nil
	}

	res.DuelsStatusMetrics.ActiveDuels = intMetric(curr.ActiveCount, prev.ActiveCount)
	res.DuelsStatusMetrics.PendingDuels = intMetric(curr.PendingCount, prev.PendingCount)
	res.DuelsStatusMetrics.ClosedDuels = intMetric(curr.ClosedCount, prev.ClosedCount)
	res.DuelsStatusMetrics.RefundDuels = intMetric(curr.RefundCount, prev.RefundCount)

	res.DuelsTVLMetrics.ActiveDuels = floatMetricOrNA(curr.ActiveTVL, true, prev.ActiveTVL, true)
	res.DuelsTVLMetrics.PendingDuels = floatMetricOrNA(curr.PendingTVL, true, prev.PendingTVL, true)
	res.DuelsTVLMetrics.ClosedDuels = floatMetricOrNA(curr.ClosedTVL, true, prev.ClosedTVL, true)
	res.DuelsTVLMetrics.RefundDuels = floatMetricOrNA(curr.RefundTVL, true, prev.RefundTVL, true)

	res.DuelsRevenueMetrics.ActiveDuels = floatMetricOrNA(curr.ActiveFee, true, prev.ActiveFee, true)
	res.DuelsRevenueMetrics.PendingDuels = floatMetricOrNA(curr.PendingFee, true, prev.PendingFee, true)
	res.DuelsRevenueMetrics.ClosedDuels = floatMetricOrNA(curr.ClosedFee, true, prev.ClosedFee, true)
	res.DuelsRevenueMetrics.RefundDuels = floatMetricOrNA(-1*curr.RefundFee, true, -1*prev.RefundFee, true)

	res.Period = period.ToJSONStruct()
	res.UpdatedAt = time.Now().UTC()

	return &res, nil
}

// ------------------------------ SYNC USERS ------------------------------

const (
	userRegBatchSize = 20_000
	userRegDelay     = 24 * time.Hour
	userRegRetention = 30 * 24 * time.Hour
)

func (s *MetricsService) SyncUserRegistrationsToClick(ctx context.Context) error {
	maxTime, ok, err := s.userRegistrationRepository.GetMaxRegisteredAt(ctx)
	if err != nil {
		return err
	}

	var from time.Time
	if !ok {
		from = time.Now().UTC().Add(-userRegRetention)
	} else {
		from = maxTime.Add(-userRegDelay)
	}

	offset := 0

	for {
		users, err := s.metricsRepo.GetUsersRegisteredAfter(
			ctx,
			from,
			userRegBatchSize,
			offset,
		)
		if err != nil {
			return err
		}
		if len(users) == 0 {
			return nil
		}

		batch := make([]model.UserRegistration, 0, len(users))
		for _, u := range users {
			batch = append(batch, model.UserRegistration{
				UserID:       u.ID,
				RegisteredAt: u.CreatedAt,
			})
		}

		if err := s.userRegistrationRepository.BulkInsert(ctx, batch); err != nil {
			return err
		}

		if len(users) < userRegBatchSize {
			return nil
		}

		offset += userRegBatchSize
	}
}

// ------------------------------ PERIOD PARSING ------------------------------

func (s *MetricsService) parsePeriod(filter *MetricsFilter) (model.MetricsPeriod, error) {
	typ := filter.Type
	if typ == "" {
		typ = "all_time"
	}

	period := model.MetricsPeriod{Type: typ}

	switch period.Type {
	case "month":
		period.Month = filter.Month

		if filter.Year != "" {
			y, err := strconv.Atoi(filter.Year)
			if err != nil || y < 2000 || y > 2100 {
				return model.MetricsPeriod{}, ErrInvalidDateFormat
			}
			period.Year = y
		} else {
			period.Year = time.Now().UTC().Year()
		}

		now := time.Now().UTC()
		requestedMonth := model.MonthFromName(filter.Month)
		if requestedMonth == 0 {
			return model.MetricsPeriod{}, ErrInvalidMonth
		}
		requestedDate := time.Date(period.Year, time.Month(requestedMonth), 1, 0, 0, 0, 0, time.UTC)
		if requestedDate.After(now) {
			return model.MetricsPeriod{}, ErrFutureDate
		}

	case "custom":
		loc := time.UTC

		if filter.StartStr != "" && filter.EndStr != "" {
			start, err := time.ParseInLocation("02-01-2006", filter.StartStr, loc)
			if err != nil {
				return model.MetricsPeriod{}, ErrInvalidDateFormat
			}

			end, err := time.ParseInLocation("02-01-2006", filter.EndStr, loc)
			if err != nil {
				return model.MetricsPeriod{}, ErrInvalidDateFormat
			}
			end = end.Add(24 * time.Hour)

			if start.After(end) {
				return model.MetricsPeriod{}, ErrStartAfterEnd
			}
			if end.After(time.Now().UTC()) {
				return model.MetricsPeriod{}, ErrFutureDate
			}

			period.StartDate = start
			period.EndDate = end
		}

	case "all_time":
	default:
		return model.MetricsPeriod{}, ErrInvalidFilterType
	}

	return period, nil
}
