package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"dd-prediction-api/internal/model"
)

type ProjectCommissionClaimRepository struct {
	db bun.IDB
}

func NewProjectCommissionClaimRepository(db *bun.DB) *ProjectCommissionClaimRepository {
	return &ProjectCommissionClaimRepository{db: db}
}

func (r *ProjectCommissionClaimRepository) WithTx(tx bun.Tx) *ProjectCommissionClaimRepository {
	return &ProjectCommissionClaimRepository{db: tx}
}

func (r *ProjectCommissionClaimRepository) Create(ctx context.Context, claim *model.ProjectCommissionClaim) error {
	_, err := r.db.NewInsert().Model(claim).Exec(ctx)
	return err
}

func (r *ProjectCommissionClaimRepository) GetByProjectAndType(ctx context.Context, projectID uuid.UUID, claimType string) ([]model.ProjectCommissionClaim, error) {
	var claims []model.ProjectCommissionClaim
	err := r.db.NewSelect().
		Model(&claims).
		Where("project_id = ?", projectID).
		Where("type = ?", claimType).
		OrderExpr("claimed_at DESC").
		Scan(ctx)
	return claims, err
}

func (r *ProjectCommissionClaimRepository) GetByType(ctx context.Context, claimType string) ([]model.ProjectCommissionClaim, error) {
	var claims []model.ProjectCommissionClaim
	err := r.db.NewSelect().
		Model(&claims).
		Where("type = ?", claimType).
		OrderExpr("claimed_at DESC").
		Scan(ctx)
	return claims, err
}
