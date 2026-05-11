package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/pkg/mtype"
	"dd-prediction-api/pkg/repository"
)

type IUserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByProjectIDAndWallet(ctx context.Context, projectID uuid.UUID, walletAddress string) (*model.User, error)
	Create(ctx context.Context, model *model.User) error
	FindByEmail(ctx context.Context, email mtype.Email) (*model.User, error)
	SetProjectID(ctx context.Context, userID, projectID uuid.UUID) error
	BlockProfile(ctx context.Context, userID uuid.UUID) error
	UnblockProfile(ctx context.Context, userID uuid.UUID) error
	WithTx(tx bun.Tx) IUserRepository
}

type UserRepository struct {
	repository.Generic[model.User, uuid.UUID]
}

func (r *UserRepository) WithTx(tx bun.Tx) IUserRepository {
	return &UserRepository{Generic: r.Generic.WithTx(tx)}
}

func NewUserRepository(
	genericRepository repository.Generic[model.User, uuid.UUID],
) *UserRepository {
	return &UserRepository{
		Generic: genericRepository,
	}
}

func (r *UserRepository) FindByEmail(
	ctx context.Context,
	email mtype.Email,
) (*model.User, error) {
	var user = new(model.User)

	err := r.DB.NewSelect().
		Model(user).
		Where("email = ?", email.String()).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) FindByProjectIDAndWallet(
	ctx context.Context,
	projectID uuid.UUID,
	walletAddress string,
) (*model.User, error) {
	var user = new(model.User)

	err := r.DB.NewSelect().
		Model(user).
		Where("project_id = ?", projectID).
		Where("wallet_address = ?", walletAddress).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) SetProjectID(
	ctx context.Context,
	userID,
	projectID uuid.UUID,
) error {

	_, err := r.DB.NewUpdate().
		Model((*model.User)(nil)).
		Set("project_id = ?", projectID).
		Where("id = ?", userID).
		Exec(ctx)

	return err
}

func (r *UserRepository) BlockProfile(ctx context.Context, userID uuid.UUID) error {
	_, err := r.DB.NewUpdate().
		Model((*model.User)(nil)).
		Set("is_active = ?", false).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", userID).
		Exec(ctx)
	return err
}

func (r *UserRepository) UnblockProfile(ctx context.Context, userID uuid.UUID) error {
	_, err := r.DB.NewUpdate().
		Model((*model.User)(nil)).
		Set("is_active = ?", true).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", userID).
		Exec(ctx)
	return err
}

func (r *UserRepository) CountNewUsersLast24Hours(ctx context.Context) (int64, error) {
	var cnt int64
	err := r.DB.NewSelect().
		Model((*model.User)(nil)).
		ColumnExpr("COUNT(*)").
		Where("created_at >= (NOW() - INTERVAL '24 hours')").
		Scan(ctx, &cnt)
	return cnt, err
}
