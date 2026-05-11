package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"dd-prediction-api/internal/model"
)

type IProjectStatusHistoryRepository interface {
	Insert(ctx context.Context, projectID uuid.UUID, status model.ProjectStatus) error
	WithTx(tx bun.Tx) IProjectStatusHistoryRepository
}

type ProjectStatusHistoryRepository struct {
	db bun.IDB
}

func NewProjectStatusHistoryRepository(db *bun.DB) *ProjectStatusHistoryRepository {
	return &ProjectStatusHistoryRepository{db: db}
}

func (r *ProjectStatusHistoryRepository) WithTx(tx bun.Tx) IProjectStatusHistoryRepository {
	return &ProjectStatusHistoryRepository{db: tx}
}

func (r *ProjectStatusHistoryRepository) Insert(ctx context.Context, projectID uuid.UUID, status model.ProjectStatus) error {
	entry := &model.ProjectStatusHistory{
		ProjectID: projectID,
		Status:    status,
		ChangedAt: time.Now().UTC(),
	}
	_, err := r.db.NewInsert().Model(entry).Exec(ctx)
	return err
}

func (r *ProjectStatusHistoryRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]model.ProjectStatusHistory, error) {
	var history []model.ProjectStatusHistory
	err := r.db.NewSelect().
		Model(&history).
		Where("project_id = ?", projectID).
		OrderExpr("changed_at DESC").
		Scan(ctx)
	return history, err
}
