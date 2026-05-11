package repository

import (
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type APIErrorRepository struct {
	DB *bun.DB
}

func NewAPIErrorRepository(db *bun.DB) *APIErrorRepository {
	return &APIErrorRepository{DB: db}
}

func (r *APIErrorRepository) LogAsync(service, endpoint string, statusCode int, body []byte) {
	go func() {
		zap.L().Warn("external API error",
			zap.String("service", service),
			zap.String("endpoint", endpoint),
			zap.Int("status_code", statusCode),
			zap.ByteString("body", body),
		)
	}()
}
