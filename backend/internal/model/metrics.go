package model

import (
	"fmt"
	"math"
	"time"
)

type MetricsPeriod struct {
	Type      string
	Month     string
	Year      int
	StartDate time.Time
	EndDate   time.Time
}

type Period struct {
	Type  string    `json:"type"`
	Label string    `json:"label"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func (p MetricsPeriod) Current() (start, end time.Time) {
	now := time.Now().UTC()

	switch p.Type {
	case "all_time":
		return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), now

	case "month":
		year := p.Year
		if year == 0 {
			year = now.Year()
		}
		month := MonthFromName(p.Month)
		if month == 0 {
			month = int(now.Month())
		}
		firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Second)
		return firstDay, lastDay

	case "custom":
		if p.StartDate.IsZero() || p.EndDate.IsZero() || p.EndDate.Before(p.StartDate) {
			return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
				now
		}
		return p.StartDate, p.EndDate

	default:
		return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), now
	}
}

func (p MetricsPeriod) Previous() (start, end time.Time) {
	currStart, currEnd := p.Current()

	switch p.Type {
	case "all_time":
		return time.Time{}, time.Time{}

	case "month":
		return currStart.AddDate(0, -1, 0), currStart.Add(-time.Second)

	case "custom":
		duration := currEnd.Sub(currStart)
		prevEnd := currStart
		prevStart := prevEnd.Add(-duration)
		return prevStart, prevEnd

	default:
		return time.Time{}, time.Time{}
	}
}

func MonthFromName(name string) int {
	switch name {
	case "January":
		return 1
	case "February":
		return 2
	case "March":
		return 3
	case "April":
		return 4
	case "May":
		return 5
	case "June":
		return 6
	case "July":
		return 7
	case "August":
		return 8
	case "September":
		return 9
	case "October":
		return 10
	case "November":
		return 11
	case "December":
		return 12
	default:
		return 0
	}
}

type IntMetric struct {
	Value      int64  `json:"value"`
	Growth     string `json:"growth"`
	IsPositive bool   `json:"is_positive"`
}

type FloatMetric struct {
	Value      float64 `json:"value"`
	Growth     string  `json:"growth"`
	IsPositive bool    `json:"is_positive"`
}

type TotalMetrics struct {
	Registrations IntMetric `json:"registrations"`
	ActiveUsers   IntMetric `json:"active_users"`
	Creators      IntMetric `json:"creators"`
}

type GrowthMetrics struct {
	NewUsers         IntMetric           `json:"new_users"`
	ActiveUsers      IntMetric           `json:"active_users"`
	Creators         IntMetric           `json:"creators"`
	ActiveUserGrowth PercentChangeMetric `json:"active_user_growth"`
}

type PredictionsMetrics struct {
	TotalVotes    IntMetric   `json:"total_votes"`
	PerActiveUser FloatMetric `json:"per_active_user"`
	PerCreator    FloatMetric `json:"per_creator"`
	TotalTVL      FloatMetric `json:"total_tvl"`
}

type MainDashboardMetrics struct {
	Period      Period             `json:"period"`
	Total       TotalMetrics       `json:"total"`
	Growth      GrowthMetrics      `json:"growth"`
	Predictions PredictionsMetrics `json:"predictions"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type PercentChangeMetric struct {
	Value      string `json:"value"`
	IsPositive bool   `json:"is_positive"`
}

type PlatformPerformanceMetrics struct {
	Period      Period             `json:"period"`
	Users       GrowthMetrics      `json:"users"`
	Predictions PredictionsMetrics `json:"predictions"`
	Revenue     RevenueDuels       `json:"revenue"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func CalculateGrowthPercent(current, previous float64) (string, bool) {
	if previous == 0 {
		if current > 0 {
			return "∞", true
		}
		return "0.00", true
	}

	pct := ((current - previous) / previous) * 100
	isPositive := pct >= 0

	value := math.Abs(pct)
	formatted := fmt.Sprintf("%.2f", value)

	return formatted, isPositive
}

func CalculateTotalGrowthPercent(current, previous float64) (string, bool) {
	if previous == 0 {
		if current > 0 {
			return "∞", true
		}
		return "0.00", true
	}

	pct := (current / previous) * 100
	value := math.Abs(pct)

	formatted := fmt.Sprintf("%.2f", value)

	return formatted, true
}

func (p MetricsPeriod) ToJSONStruct() Period {
	start, end := p.Current()

	var label string

	switch p.Type {
	case "all_time":
		label = "All Time"

	case "month":
		monthName := p.Month
		if monthName == "" {
			monthName = time.Month(start.Month()).String()
		}
		year := p.Year
		if year == 0 {
			year = start.Year()
		}
		label = fmt.Sprintf("%s %d", monthName, year)
	case "custom":
		label = fmt.Sprintf(
			"%d %s %d – %d %s %d",
			start.Day(), start.Month().String()[:3], start.Year(),
			end.Day(), end.Month().String()[:3], end.Year(),
		)

	default:
		label = "Period"
	}

	return Period{
		Type:  p.Type,
		Label: label,
		Start: start.UTC(),
		End:   end.UTC(),
	}
}

type HealthDashboardMetrics struct {
	Retention7  IntMetric `json:"retention_7"`
	Retention30 IntMetric `json:"retention_30"`
	BounceRate  IntMetric `json:"bounce_rate"`
	ChurnRate   IntMetric `json:"churn_rate"`
	Period      Period    `json:"period"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Stats struct {
	TotalUsers  int `json:"total_users"`
	ActiveUsers int `json:"active_users"`
	ActiveDuels int `json:"active_duels"`
}

type RevenueDashboardMetrics struct {
	Duels RevenueDuels `json:"duels"`
	//Swap  SwapRevenueMetrics  `json:"swap"`
	//API   APIRevenueMetrics   `json:"api"`

	Period    Period    `json:"period"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RevenueDuels struct {
	AOV                 FloatMetric `json:"aov"`
	AvgPredictionPrice  FloatMetric `json:"avg_prediction_price"`
	AvgFeePerPrediction FloatMetric `json:"avg_fee_per_prediction"`
	Revenue             FloatMetric `json:"revenue"`
}

type DuelsDashboardMetrics struct {
	DuelsStatusMetrics  DuelsStatusMetrics  `json:"status"`
	DuelsTVLMetrics     DuelsTVLMetrics     `json:"tvl"`
	DuelsRevenueMetrics DuelsRevenueMetrics `json:"revenue"`

	Period    Period    `json:"period"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DuelsStatusMetrics struct {
	ActiveDuels  IntMetric `json:"active_duels"`
	PendingDuels IntMetric `json:"pending_duels"`
	ClosedDuels  IntMetric `json:"closed_duels"`
	RefundDuels  IntMetric `json:"refund_duels"`
}

type DuelsTVLMetrics struct {
	ActiveDuels  FloatMetric `json:"active_duels"`
	PendingDuels FloatMetric `json:"pending_duels"`
	ClosedDuels  FloatMetric `json:"closed_duels"`
	RefundDuels  FloatMetric `json:"refund_duels"`
}

type DuelsRevenueMetrics struct {
	ActiveDuels  FloatMetric `json:"active_duels"`
	PendingDuels FloatMetric `json:"pending_duels"`
	ClosedDuels  FloatMetric `json:"closed_duels"`
	RefundDuels  FloatMetric `json:"refund_duels"`
}
