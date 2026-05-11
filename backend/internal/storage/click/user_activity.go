package click

import (
	"context"
	"fmt"

	"github.com/uptrace/go-clickhouse/ch"

	"dd-prediction-api/internal/model"
)

type UserActivityRepository struct {
	db *ch.DB
}

func NewUserActivityRepository(db *ch.DB) *UserActivityRepository {
	return &UserActivityRepository{db: db}
}

func (r *UserActivityRepository) Insert(ctx context.Context, activity model.UserActivity) error {
	meta := activity.Metadata
	if meta == "" {
		meta = "{}"
	}
	if _, err := r.db.NewInsert().Model(&activity).Exec(ctx); err != nil {
		return fmt.Errorf("failed to insert user activity: %w", err)
	}
	return nil
}

func (r *UserActivityRepository) BulkInsert(ctx context.Context, activities []model.UserActivity) error {
	if len(activities) == 0 {
		return nil
	}

	if _, err := r.db.NewInsert().Model(&activities).Exec(ctx); err != nil {
		return fmt.Errorf("failed to bulk insert user activities: %w", err)
	}
	return nil
}
