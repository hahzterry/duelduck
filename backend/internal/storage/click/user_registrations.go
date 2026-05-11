package click

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/go-clickhouse/ch"

	"dd-prediction-api/internal/model"
)

type UserRegistrationRepository struct {
	db *ch.DB
}

func NewUserRegistrationRepository(db *ch.DB) *UserRegistrationRepository {
	return &UserRegistrationRepository{db: db}
}

func (r *UserRegistrationRepository) GetMaxRegisteredAt(ctx context.Context) (time.Time, bool, error) {
	var cnt int64
	if err := r.db.NewSelect().
		Table("user_registrations").
		ColumnExpr("count()").
		Scan(ctx, &cnt); err != nil {
		return time.Time{}, false, err
	}

	if cnt == 0 {
		return time.Time{}, false, nil
	}

	var t time.Time
	if err := r.db.NewSelect().
		Table("user_registrations").
		ColumnExpr("max(registered_at)").
		Scan(ctx, &t); err != nil {
		return time.Time{}, false, err
	}

	return t, true, nil
}

func (r *UserRegistrationRepository) BulkInsert(ctx context.Context, items []model.UserRegistration) error {
	if len(items) == 0 {
		return nil
	}
	if _, err := r.db.NewInsert().Model(&items).Exec(ctx); err != nil {
		return fmt.Errorf("insert user_registrations: %w", err)
	}
	return nil
}
