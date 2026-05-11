package click

import (
	"context"
	"fmt"

	"github.com/uptrace/go-clickhouse/ch"
	"go.uber.org/zap"

	"dd-prediction-api/internal/model"
)

type AdminAuditLogRepository struct {
	db *ch.DB
}

func NewAdminAuditLogRepository(db *ch.DB) *AdminAuditLogRepository {
	return &AdminAuditLogRepository{db: db}
}

func (r *AdminAuditLogRepository) Insert(ctx context.Context, log model.AdminAuditLog) {
	if _, err := r.db.NewInsert().Model(&log).Exec(ctx); err != nil {
		zap.L().Error("failed to insert admin audit log",
			zap.String("action", log.Action),
			zap.String("moderator_id", log.ModeratorID.String()),
			zap.Error(fmt.Errorf("%w", err)),
		)
	}
}
