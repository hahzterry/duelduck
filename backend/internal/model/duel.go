package model

import (
	"math"
	"math/rand/v2"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const (
	_                           = uint8(iota)
	DuelStatusDenied            // rejected before user participation
	DuelStatusActive            // open for voting
	DuelStatusWaitingForResolve // deadline passed, awaiting result
	DuelStatusResolved          // winner determined
	DuelStatusRefunded          // money returned to participants
)

const (
	USDCSymbol = "USDC"
	SOLSymbol  = "SOL"
)

var (
	JoinedSPLRoomRegex  = regexp.MustCompile(`joined room (\d+)!`)
	JoinedSOLRoomRegex  = regexp.MustCompile(`joined SOL room (\d+)!`)
	InitedRoomRegex     = regexp.MustCompile(`Room token account:\s*([1-9A-HJ-NP-Za-km-z]{32,44})`)
	RoomSOLAccountRegex = regexp.MustCompile(`Room SOL account:\s*([1-9A-HJ-NP-Za-km-z]{32,44})`)
)

const (
	SamePredictionCancellationReason     = "All users made the same prediction"
	LackOfParticipantsCancellationReason = "The duel was canceled due to a lack of participants"
	AIModerationCancellationReason       = "The duel was automatically canceled after AI moderation"
)

const (
	DuelTopicPlaceholder = "topic"
)

const (
	USDDuelMinJoinPrice   = 0.999 // since 1 USDC ~ 0.999 USD
	USDDuelMaxJoinPrice   = 5000.0
	MaxDuelCommissionRate = 20
)

const (
	DuelPriceTypeFixed = "fixed"
	DuelPriceTypeRange = "range"
)

type Duel struct {
	bun.BaseModel `bun:"table:duels,alias:duels" json:"-"`

	ID                   uuid.UUID      `bun:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	OwnerID              uuid.UUID      `bun:"owner_id,type:uuid,notnull" json:"owner_id"`
	ProjectID            uuid.UUID      `bun:"project_id,type:uuid,notnull" json:"project_id"`
	ResolvedBy           uuid.UUID      `bun:"resolved_by,type:uuid,default:null" json:"resolved_by"`
	ApprovedBy           uuid.UUID      `bun:"approved_by,type:uuid,default:null" json:"approved_by"`
	ResolvedAt           *time.Time     `bun:"resolved_at,type:timestamp" json:"resolved_at"`
	IsOwnerResolving     bool           `bun:",notnull,default:false" json:"is_owner_resolving"`
	RoomNumber           uint64         `bun:"room_number,type:integer,nullzero" json:"room_number"`
	RoomTokenPDA         string         `bun:"room_token_pda,type:string,nullzero" json:"room_token_pda"`
	Symbol               string         `bun:"symbol,type:string,notnull" json:"symbol"`
	PlayersCount         uint64         `bun:"players_count,type:integer,notnull,default:0" json:"players_count"`
	RefundedPlayersCount uint64         `bun:"refunded_players_count,type:integer,notnull" json:"refunded_players_count"`
	WinnersCount         uint64         `bun:"winners_count,type:integer,notnull" json:"winners_count"`
	Username             string         `bun:"username,type:varchar(17),notnull" json:"username"`
	Status               uint8          `bun:"status,type:integer,notnull,default:0" json:"status"`
	LogoURL              string         `bun:"logo_url,type:text" json:"logo_url"`
	Question             string         `bun:"question,type:text" json:"question"`
	Slug                 string         `bun:"slug,type:text,nullzero" json:"slug"`
	SourceOfTruth        string         `bun:"source_of_truth,type:text" json:"source_of_truth"`
	DuelPrice            float64        `bun:"duel_price,type:int,notnull" json:"duel_price"`
	PriceType            string         `bun:"price_type,type:varchar(10),notnull,default:'fixed'" json:"price_type"`
	MinPrice             *float64       `bun:"min_price,type:decimal(15,9)" json:"min_price,omitempty"`
	MaxPrice             *float64       `bun:"max_price,type:decimal(15,9)" json:"max_price,omitempty"`
	USDPrice             *float64       `bun:"usd_price,type:decimal(15,9)" json:"usd_price"`
	Commission           uint64         `bun:"commission,type:integer,notnull" json:"commission"`
	CommissionRate       uint64         `bun:"commission_rate,type:integer,notnull,default:0" json:"commission_rate"`
	DuelInfo             map[string]any `bun:"duel_info,type:json" json:"duel_info"`
	FinalResult          *uint8         `bun:"final_result,type:integer" json:"final_result"`
	CancellationReason   string         `bun:"cancellation_reason,type:text" json:"cancellation_reason"`
	Deadline             time.Time      `bun:"deadline,notnull,default:current_timestamp" json:"deadline"`
	CreatedAt            time.Time      `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt            time.Time      `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

func DuelByCreateReqAdmin(req *CreateDuelReq, user *User) *Duel {
	duel := DuelByCreateReq(req, user)

	duel.Status = DuelStatusActive // duel created by admin so no need to review
	duel.LogoURL = req.LogoURL

	return duel
}

func DuelByCreateReq(req *CreateDuelReq, user *User) *Duel {
	roomNum := uint64(rand.Int64N(math.MaxUint32-10_000) + 10_000 + 1)

	now := time.Now()
	return &Duel{
		ID:               uuid.New(),
		IsOwnerResolving: req.IsOwnerResolving,
		RoomNumber:       roomNum,
		Symbol:           req.Symbol,
		Status:           DuelStatusActive,
		OwnerID:          user.ID,
		LogoURL:          req.LogoURL,
		Question:         req.Question,
		SourceOfTruth:    req.SourceOfTruth,
		Deadline:         req.Deadline,
		DuelPrice:        req.DuelPrice,
		PriceType:        req.PriceType,
		MinPrice:         req.MinPrice,
		MaxPrice:         req.MaxPrice,
		CommissionRate:   req.CommissionRate,
		DuelInfo:         req.DuelInfo,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

type CreateDuelReq struct {
	Symbol           string         `json:"symbol"`
	LogoURL          string         `json:"logo_url"`
	IsOwnerResolving bool           `json:"is_owner_resolving"`
	Question         string         `json:"question"`
	SourceOfTruth    string         `json:"source_of_truth"`
	Deadline         time.Time      `json:"deadline"`
	DuelPrice        float64        `json:"duel_price"`
	PriceType        string         `json:"price_type"`
	MinPrice         *float64       `json:"min_price,omitempty"`
	MaxPrice         *float64       `json:"max_price,omitempty"`
	CommissionRate   uint64         `json:"commission_rate"`
	DuelInfo         map[string]any `json:"duel_info"`
	Answer           uint8          `json:"answer"`
}

type CreateCryptoDuelReq struct {
	CreateDuelReq CreateDuelReq `json:"duel"`
	Hash          string        `json:"tx_hash"`
}

type JoinCryptoDuelReq struct {
	JoinDuelReq JoinDuelReq `json:"duel"`
	Hash        string      `json:"tx_hash"`
}

type JoinDuelReq struct {
	DuelID                uuid.UUID  `json:"duel_id"`
	Answer                uint8      `json:"answer"`
	InvitedBy             string     `json:"invited_by"`
	ExternalSource        string     `json:"external_source"`
	MultiDuelID           *uuid.UUID `json:"multi_duel_id,omitempty"`
	MultiDuelReferralCode string     `json:"multi_duel_referral_code,omitempty"`
	WalletID              uuid.UUID  `json:"-"`
	PaidPrice             *float64   `json:"paid_price,omitempty"`
}

type DuelResolveReq struct {
	DuelID uuid.UUID `json:"duel_id"`
	Answer uint8     `json:"answer"`
}

type DuelResolveParams struct {
	DuelID        uuid.UUID `json:"duel_id"`
	Answer        uint8     `json:"answer"`
	JoinNotBefore time.Time `json:"not_before"`
}

type DuelApproveReq struct {
	DuelID     uuid.UUID  `json:"duel_id"`
	CategoryID *uuid.UUID `json:"category_id,omitempty"`
}

type DuelCancelReq struct {
	DuelID             uuid.UUID `json:"duel_id"`
	Status             uint8     `json:"status"`
	CancellationReason string    `json:"cancellation_reason"`
}

type DuelAdminEditReq struct {
	ID            uuid.UUID      `json:"id"`
	Question      string         `json:"question"`
	DuelType      string         `json:"duel_type"`
	SourceOfTruth string         `json:"source_of_truth"`
	LogoURL       string         `json:"logo_url"`
	DuelInfo      map[string]any `json:"duel_info"`
	Deadline      time.Time      `json:"deadline"`
}

type DuelTokenInfo struct {
	Mint          string  `json:"mint"`
	Symbol        string  `bun:"" json:"symbol"`
	Decimals      uint8   `bun:"" json:"decimals"`
	USDPrice      float64 `bun:"usd_price,notnull" json:"usd_price"`
	ProgramID     string  `json:"program_id"`
	TokenImageUrl string  `json:"token_image_url"`
}

func (i *DuelTokenInfo) CryptoDuelPriceMultiplier() float64 {
	return math.Pow10(int(i.Decimals))
}

type DuelParams struct {
	Pool         float64
	Commission   float64
	PlayersCount float64
	WinnersCount float64
}

func NewDuelParams(
	duelPrice float64,
	commission uint64,
	playersCount uint64,
	winnersCount uint64,
) DuelParams {
	return DuelParams{
		Pool:         float64(playersCount) * duelPrice,
		Commission:   float64(commission),
		PlayersCount: float64(playersCount),
		WinnersCount: float64(winnersCount),
	}
}

type TxHashResp struct {
	TxHash string `json:"tx_hash"`
}

type CreateCryptoDuelResp struct {
	Duel   *Duel       `json:"duel"`
	Result *TxHashResp `json:"result"`
}

type JoinCryptoDuelResp struct {
	Player *Player     `json:"player"`
	Result *TxHashResp `json:"result"`
}

type ResolveCryptoDuelResp struct {
	TxHashes []string `json:"tx_hashes"`
	Duel     *Duel    `json:"duel,omitempty"`
}

type CancelCryptoDuelResp struct {
	TxHashes []string `json:"tx_hashes"`
}

func AutoCancelReq(duel *Duel) *DuelCancelReq {
	cancellationReason := SamePredictionCancellationReason
	if duel.PlayersCount <= 1 {
		cancellationReason = LackOfParticipantsCancellationReason
	}

	return &DuelCancelReq{
		DuelID:             duel.ID,
		Status:             DuelStatusRefunded,
		CancellationReason: cancellationReason,
	}
}

func (p DuelParams) CalculateFinalCryptoReward(priceMultiplier float64) uint64 {
	percentValue := p.Pool * p.Commission * priceMultiplier / 100
	finalPool := p.Pool*priceMultiplier - percentValue

	return uint64(finalPool / p.WinnersCount)
}

func (p DuelParams) CalculateCryptoCommissionReward(priceMultiplier float64) uint64 {
	return uint64(p.Pool * p.Commission * priceMultiplier / 100)
}

type DuelShow struct {
	bun.BaseModel `bun:"table:duels,alias:duels" json:"-"`

	Duel
	MultiDuelID   *uuid.UUID `bun:",column:multi_duel_id" json:"multi_duel_id"`
	MultiDuelSlug *string    `bun:",column:multi_duel_slug" json:"multi_duel_slug"`
	OwnerImageURL string     `bun:"owner_image_url" json:"owner_image_url"`
	YesCount      uint64     `bun:",column:yes_count" json:"yes_count"`
	NoCount       uint64     `bun:",column:no_count" json:"no_count"`
	Joined        bool       `bun:",column:joined" json:"joined"`
	YourAnswer    *int       `bun:",column:your_answer" json:"your_answer"`
	PlayerStatus  uint8      `bun:",column:player_status" json:"player_status"`
	WinAmount     *float64   `bun:",column:win_amount" json:"win_amount"`
	PotentialWin  *float64   `bun:"-" json:"potential_win"`
}

type DuelListWithTotalResp struct {
	Duels []DuelShow `json:"duels"`
	Total int        `json:"total"`
}

type DuelSymbolVolume24h struct {
	Symbol       string  `json:"symbol"`
	NativeVolume float64 `json:"native_volume"`
	USDCVolume   float64 `json:"usdc_volume"`
}

type DuelActiveWagerVolume24h struct {
	TotalUSDC float64               `json:"total_usdc"`
	BySymbol  []DuelSymbolVolume24h `json:"by_symbol"`
}

// DuelFeeIncome24h — platform and creators fee income for last 24h.
type DuelFeeIncome24h struct {
	PlatformFeeUSDC float64 `json:"platform_fee_usdc"`
	UsersFeeUSDC    float64 `json:"users_fee_usdc"`
	UsersWhoEarned  int64   `json:"users_who_earned"`
}

// UsersActivity24h — new users, players, creators counts for last 24h.
type UsersActivity24h struct {
	NewUsers int64 `json:"new_users"`
	Players  int64 `json:"players"`
	Creators int64 `json:"creators"`
}

type PublicProfileDuelCounts struct {
	Created int64 `bun:"created" json:"created"`
	Playing int64 `bun:"playing" json:"playing"`
	Won     int64 `bun:"won" json:"won"`
	Lost    int64 `bun:"lost" json:"lost"`
}

type FinancialTxEntry struct {
	TxType uint8     `json:"tx_type"`
	TxHash string    `json:"tx_hash,omitempty"`
	Amount float64   `json:"amount"`
	Date   time.Time `json:"date"`
}
