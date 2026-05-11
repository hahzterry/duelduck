package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/go-clickhouse/ch"
)

type UserActivity struct {
	ch.CHModel `ch:"table:user_activities"`

	UserID        uuid.UUID `ch:"user_id"`
	ActivityType  string    `ch:"activity_type"`
	ActivityCount uint32    `ch:"activity_count"`
	CreatedAt     time.Time `ch:"created_at"`
	Metadata      string    `ch:"metadata"`
}

const (
	ActivityLogin          = "login"
	ActivityDuelJoinFree   = "duel_join_free"
	ActivityDuelJoinPaid   = "duel_join_paid"
	ActivityDuelCreateFree = "duel_create_free"
	ActivityDuelCreatePaid = "duel_create_paid"
	ActivityPageView       = "page_view"
	ActivityTokenRefresh   = "token_refresh"
)

type ctxKey string

const (
	CtxTrackActivityKey ctxKey = "track_activity"
)

func ShouldTrackActivity(ctx context.Context) bool {
	if v, ok := ctx.Value(CtxTrackActivityKey).(bool); ok {
		return v
	}
	return false
}

type UserRegistration struct {
	ch.CHModel `ch:"table:user_registrations"`

	UserID       uuid.UUID `ch:"user_id"`
	RegisteredAt time.Time `ch:"registered_at"`
}
