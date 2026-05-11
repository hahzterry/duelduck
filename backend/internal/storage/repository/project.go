package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/pkg/repository"
)

type IProjectRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*model.Project, error)
	GetByPartnerID(ctx context.Context, partnerID uuid.UUID) (*model.Project, error)
	GetAll(ctx context.Context) ([]model.Project, error)
	IsBlocked(ctx context.Context, id uuid.UUID) (bool, error)
	SetBlocked(ctx context.Context, id uuid.UUID, blocked bool) error
	SetStatus(ctx context.Context, id uuid.UUID, status model.ProjectStatus) error
	Update(ctx context.Context, id uuid.UUID, name, siteURL string) error
	UpdatePermissions(ctx context.Context, id uuid.UUID, duelsEnabled, selfResolveEnabled bool) error
	Create(ctx context.Context, model *model.Project) error
	HasPartnerProject(ctx context.Context, partnerID uuid.UUID) (bool, error)
	WithTx(tx bun.Tx) IProjectRepository
}

type ProjectRepository struct {
	repository.Generic[model.Project, uuid.UUID]
}

func (r *ProjectRepository) WithTx(tx bun.Tx) IProjectRepository {
	return &ProjectRepository{Generic: r.Generic.WithTx(tx)}
}

func NewProjectRepository(
	genericRepository repository.Generic[model.Project, uuid.UUID],
) *ProjectRepository {
	return &ProjectRepository{
		Generic: genericRepository,
	}
}

func (r *ProjectRepository) GetByAPIKey(
	ctx context.Context,
	apiKey string,
) (*model.Project, error) {
	var project model.Project
	err := r.DB.NewSelect().
		Model(&project).
		Where("api_key = ?", apiKey).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) IsBlocked(ctx context.Context, id uuid.UUID) (bool, error) {
	var result struct {
		Status model.ProjectStatus `bun:"status"`
	}
	err := r.DB.NewSelect().
		Model((*model.Project)(nil)).
		Column("status").
		Where("id = ?", id).
		Scan(ctx, &result)
	if err != nil {
		return false, err
	}
	return result.Status != model.ProjectStatusActive, nil
}

func (r *ProjectRepository) SetBlocked(ctx context.Context, id uuid.UUID, blocked bool) error {
	status := model.ProjectStatusActive
	if blocked {
		status = model.ProjectStatusBlocked
	}
	_, err := r.DB.NewUpdate().
		Model((*model.Project)(nil)).
		Set("status = ?", status).
		Set("is_blocked = ?", blocked).
		Set("updated_at = NOW()").
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *ProjectRepository) SetStatus(ctx context.Context, id uuid.UUID, status model.ProjectStatus) error {
	isBlocked := status != model.ProjectStatusActive
	_, err := r.DB.NewUpdate().
		Model((*model.Project)(nil)).
		Set("status = ?", status).
		Set("is_blocked = ?", isBlocked).
		Set("updated_at = NOW()").
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *ProjectRepository) Update(ctx context.Context, id uuid.UUID, name, siteURL string) error {
	_, err := r.DB.NewUpdate().
		Model((*model.Project)(nil)).
		Set("name = ?", name).
		Set("site_url = ?", siteURL).
		Set("updated_at = NOW()").
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *ProjectRepository) UpdatePermissions(ctx context.Context, id uuid.UUID, duelsEnabled, selfResolveEnabled bool) error {
	_, err := r.DB.NewUpdate().
		Model((*model.Project)(nil)).
		Set("is_users_duels_enabled = ?", duelsEnabled).
		Set("is_self_resolved_enabled = ?", selfResolveEnabled).
		Set("updated_at = NOW()").
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *ProjectRepository) GetAll(ctx context.Context) ([]model.Project, error) {
	var projects []model.Project
	err := r.DB.NewSelect().
		Model(&projects).
		OrderExpr("created_at DESC").
		Scan(ctx)
	return projects, err
}

func (r *ProjectRepository) GetByPartnerID(ctx context.Context, partnerID uuid.UUID) (*model.Project, error) {
	var project model.Project
	err := r.DB.NewSelect().
		Model(&project).
		Relation("StatusHistory", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.OrderExpr("changed_at DESC")
		}).
		Where("p.partner_id = ?", partnerID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) HasPartnerProject(
	ctx context.Context,
	partnerID uuid.UUID,
) (bool, error) {
	return r.DB.NewSelect().
		Model((*model.Project)(nil)).
		Where("partner_id = ?", partnerID).
		Exists(ctx)
}
