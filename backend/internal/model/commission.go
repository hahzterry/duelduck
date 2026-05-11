package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const (
	ClaimTypePartner  = "partner"
	ClaimTypeDDProfit = "dd_profit"
)

type ProjectCommissionAccrual struct {
	bun.BaseModel `bun:"table:project_commission_accruals,alias:pca" json:"-"`

	ID            uuid.UUID `bun:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	ProjectID     uuid.UUID `bun:",type:uuid,notnull" json:"project_id"`
	DuelID        uuid.UUID `bun:",type:uuid,notnull" json:"duel_id"`
	Symbol        string    `bun:"type:varchar(20),notnull" json:"symbol"`
	CommissionRaw uint64    `bun:",notnull" json:"commission_raw"`
	CommissionUSD float64   `bun:"type:decimal(15,6),notnull" json:"commission_usd"`
	BillingMonth  time.Time `bun:"type:date,notnull" json:"billing_month"`
	AccruedAt     time.Time `bun:",notnull,default:current_timestamp" json:"accrued_at"`
}

type ProjectCommissionClaim struct {
	bun.BaseModel `bun:"table:project_commission_claims,alias:pcc" json:"-"`

	ID            uuid.UUID         `bun:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Type          string            `bun:"type:varchar(20),notnull" json:"type"`
	ProjectID     *uuid.UUID        `bun:",type:uuid" json:"project_id,omitempty"`
	PeriodStart   time.Time         `bun:"type:date,notnull" json:"period_start"`
	PeriodEnd     time.Time         `bun:"type:date,notnull" json:"period_end"`
	GrossUSD      float64           `bun:"type:decimal(15,6),notnull" json:"gross_usd"`
	PlatformRate  float64           `bun:"type:decimal(5,4),notnull" json:"platform_rate"`
	AmountUSD     float64           `bun:"type:decimal(15,6),notnull" json:"amount_usd"`
	WalletAddress string            `bun:",notnull" json:"wallet_address"`
	TxHashes      map[string]string `bun:"type:jsonb,notnull" json:"tx_hashes"`
	ClaimedAt     time.Time         `bun:",notnull,default:current_timestamp" json:"claimed_at"`
}

type ProjectCommissionClaimReq struct {
	WalletAddress string `json:"wallet_address"`
}

type ProjectCommissionSummary struct {
	ProjectID      uuid.UUID `json:"project_id"`
	Symbol         string    `json:"symbol"`
	CommissionRaw  uint64    `json:"commission_raw"`
	CommissionUSD  float64   `json:"commission_usd"`
	PlatformRate   float64   `json:"platform_rate"`
	PlatformFeeRaw uint64    `json:"platform_fee_raw"`
	PlatformFeeUSD float64   `json:"platform_fee_usd"`
}

type SymbolAccrual struct {
	Symbol    string  `json:"symbol"`
	AmountRaw uint64  `json:"amount_raw"`
	AmountUSD float64 `json:"amount_usd"`
}

type PartnerDashboard struct {
	APIKey           string        `json:"api_key"`
	Status           ProjectStatus `json:"status"`
	DuelCount        int           `json:"duel_count"`
	TotalVolumeUSD   float64       `json:"total_volume_usd"`
	MonthlyVolumeUSD float64       `json:"monthly_volume_usd"`
	PlatformRate     float64       `json:"platform_rate"`
	NetBySymbol      []SymbolNet   `json:"net_by_symbol"`
}

// SymbolNet is the partner's net income for a specific token after DD commission.
type SymbolNet struct {
	Symbol   string  `json:"symbol"`
	NetRaw   uint64  `json:"net_raw"`
	NetUSD   float64 `json:"net_usd"`
	GrossUSD float64 `json:"gross_usd"`
	PlatRate float64 `json:"platform_rate"`
}

// PartnerDailyIncome is income earned per calendar day (USDC equivalent).
type PartnerDailyIncome struct {
	Day    string  `json:"day"`
	Amount float64 `json:"amount_usd"`
}

// PartnerMonthlyMAU is unique active users per calendar month.
type PartnerMonthlyMAU struct {
	Month       string `json:"month"`
	ActiveUsers uint64 `json:"active_users"`
}

// PartnerMonthlyDDCommission is DD commission per billing month.
type PartnerMonthlyDDCommission struct {
	Month    string  `json:"month"`
	GrossUSD float64 `json:"gross_usd"`
	DDAmount float64 `json:"dd_amount_usd"`
	DDRate   float64 `json:"dd_rate"`
}

// PlatformCommissionRate returns the platform's cut based on gross commission income in USD.
func PlatformCommissionRate(totalUSD float64) float64 {
	switch {
	case totalUSD <= 1000:
		return 0
	case totalUSD <= 10_000:
		return 0.10
	case totalUSD <= 50_000:
		return 0.075
	case totalUSD <= 100_000:
		return 0.05
	default:
		return 0.025
	}
}

func BillingMonthStart(t time.Time) time.Time {
	return time.Date(t.UTC().Year(), t.UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
}

// ClaimableThrough returns the latest billing month that can be claimed right now.
// A billing month becomes claimable as soon as it ends (from the 1st of the next month).
func ClaimableThrough(now time.Time) time.Time {
	return BillingMonthStart(now).AddDate(0, -1, 0)
}
