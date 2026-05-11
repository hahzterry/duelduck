package click

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/uptrace/go-clickhouse/ch"

	"dd-prediction-api/internal/model"
)

type IDuelTransactionRepository interface {
	BulkInsert(ctx context.Context, txs []model.DuelTransaction) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.DuelTransaction, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]model.DuelTransaction, error)
}

type DuelTransactionRepository struct {
	db *ch.DB
}

func NewDuelTransactionRepository(db *ch.DB) *DuelTransactionRepository {
	return &DuelTransactionRepository{db: db}
}

func (r *DuelTransactionRepository) BulkInsert(ctx context.Context, txs []model.DuelTransaction) error {
	if len(txs) == 0 {
		return nil
	}
	if _, err := r.db.NewInsert().Model(&txs).Exec(ctx); err != nil {
		return fmt.Errorf("insert duel_transactions: %w", err)
	}
	return nil
}

func (r *DuelTransactionRepository) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]model.DuelTransaction, error) {
	var txs []model.DuelTransaction
	err := r.db.NewSelect().
		Model(&txs).
		Where("user_id = ?", userID).
		OrderExpr("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("get duel_transactions by user_id: %w", err)
	}
	return txs, nil
}

func (r *DuelTransactionRepository) GetByProjectID(
	ctx context.Context,
	projectID uuid.UUID,
) ([]model.DuelTransaction, error) {
	var txs []model.DuelTransaction
	err := r.db.NewSelect().
		Model(&txs).
		Where("project_id = ?", projectID).
		OrderExpr("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("get duel_transactions by project_id: %w", err)
	}
	return txs, nil
}
