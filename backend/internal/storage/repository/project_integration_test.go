//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dd-prediction-api/internal/model"
)

func TestProjectRepository_Create_GetByID(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	userRepo := newUserRepo(db, tx)
	projRepo := newProjectRepo(db, tx)
	ctx := context.Background()

	user := insertTestUser(t, userRepo, ctx)
	proj := insertTestProject(t, projRepo, ctx, user.ID)

	found, err := projRepo.GetByID(ctx, proj.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, proj.ID, found.ID)
	assert.Equal(t, user.ID, found.PartnerID)
	assert.Equal(t, proj.APIKey, found.APIKey)
	assert.False(t, found.IsBlocked)
}

func TestProjectRepository_GetByAPIKey(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	userRepo := newUserRepo(db, tx)
	projRepo := newProjectRepo(db, tx)
	ctx := context.Background()

	user := insertTestUser(t, userRepo, ctx)
	proj := insertTestProject(t, projRepo, ctx, user.ID)

	found, err := projRepo.GetByAPIKey(ctx, proj.APIKey)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, proj.ID, found.ID)
}

func TestProjectRepository_GetByPartnerID(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	userRepo := newUserRepo(db, tx)
	projRepo := newProjectRepo(db, tx)
	ctx := context.Background()

	user := insertTestUser(t, userRepo, ctx)
	proj := insertTestProject(t, projRepo, ctx, user.ID)

	found, err := projRepo.GetByPartnerID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, proj.ID, found.ID)
}

func TestProjectRepository_GetByPartnerID_notFound(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	projRepo := newProjectRepo(db, tx)
	ctx := context.Background()

	found, err := projRepo.GetByPartnerID(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestProjectRepository_HasPartnerProject(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	userRepo := newUserRepo(db, tx)
	projRepo := newProjectRepo(db, tx)
	ctx := context.Background()

	user := insertTestUser(t, userRepo, ctx)

	has, err := projRepo.HasPartnerProject(ctx, user.ID)
	require.NoError(t, err)
	assert.False(t, has)

	insertTestProject(t, projRepo, ctx, user.ID)

	has, err = projRepo.HasPartnerProject(ctx, user.ID)
	require.NoError(t, err)
	assert.True(t, has)
}

func TestProjectRepository_SetBlocked_IsBlocked(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	userRepo := newUserRepo(db, tx)
	projRepo := newProjectRepo(db, tx)
	ctx := context.Background()

	user := insertTestUser(t, userRepo, ctx)
	proj := insertTestProject(t, projRepo, ctx, user.ID)

	// Initially not blocked.
	blocked, err := projRepo.IsBlocked(ctx, proj.ID)
	require.NoError(t, err)
	assert.False(t, blocked)

	// Block.
	require.NoError(t, projRepo.SetBlocked(ctx, proj.ID, true))
	blocked, err = projRepo.IsBlocked(ctx, proj.ID)
	require.NoError(t, err)
	assert.True(t, blocked)

	// Unblock.
	require.NoError(t, projRepo.SetBlocked(ctx, proj.ID, false))
	blocked, err = projRepo.IsBlocked(ctx, proj.ID)
	require.NoError(t, err)
	assert.False(t, blocked)
}

func TestProjectRepository_APIKeyUniqueness(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	userRepo := newUserRepo(db, tx)
	projRepo := newProjectRepo(db, tx)
	ctx := context.Background()

	u1 := insertTestUser(t, userRepo, ctx)
	u2 := insertTestUser(t, userRepo, ctx)

	sharedKey := "shared_api_key_unique_check"

	id1 := uuid.New()
	p1 := model.NewProject(id1, sharedKey, u1.ID)
	require.NoError(t, projRepo.Create(ctx, p1))

	id2 := uuid.New()
	p2 := model.NewProject(id2, sharedKey, u2.ID)
	err := projRepo.Create(ctx, p2)
	require.Error(t, err, "duplicate api_key must be rejected by the DB")
}
