//go:build integration

package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/pkg/mtype"
	pkgrepo "dd-prediction-api/pkg/repository"
)

// openTestDB opens a *bun.DB against TEST_DATABASE_URL.
// The test is skipped when the variable is not set.
func openTestDB(t *testing.T) *bun.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run integration tests")
	}
	conf, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	pool, err := pgxpool.NewWithConfig(context.Background(), conf)
	require.NoError(t, err)
	sqlDB := stdlib.OpenDBFromPool(pool)
	db := bun.NewDB(sqlDB, pgdialect.New())
	require.NoError(t, db.Ping())
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// beginTx starts a transaction and registers a Rollback cleanup so every test
// leaves the database in a clean state.
func beginTx(t *testing.T, db *bun.DB) bun.Tx {
	t.Helper()
	tx, err := db.Begin()
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback() })
	return tx
}

// newUserRepo creates a UserRepository that executes within tx.
func newUserRepo(db *bun.DB, tx bun.Tx) *UserRepository {
	return &UserRepository{
		Generic: pkgrepo.NewGenericRepository[model.User, uuid.UUID](db).WithTx(tx),
	}
}

// newProjectRepo creates a ProjectRepository that executes within tx.
func newProjectRepo(db *bun.DB, tx bun.Tx) *ProjectRepository {
	return &ProjectRepository{
		Generic: pkgrepo.NewGenericRepository[model.Project, uuid.UUID](db).WithTx(tx),
	}
}

// insertTestUser inserts a minimal user and returns it.
func insertTestUser(t *testing.T, repo *UserRepository, ctx context.Context) *model.User {
	t.Helper()
	u := model.NewUser(mtype.RoleUser)
	require.NoError(t, repo.Create(ctx, u))
	return u
}

// insertTestProject inserts a project whose partner is partnerID and returns it.
func insertTestProject(t *testing.T, repo *ProjectRepository, ctx context.Context, partnerID uuid.UUID) *model.Project {
	t.Helper()
	id := uuid.New()
	p := model.NewProject(id, "test_key_"+id.String(), partnerID)
	require.NoError(t, repo.Create(ctx, p))
	return p
}
