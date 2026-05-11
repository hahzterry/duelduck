//go:build integration

package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/pkg/mtype"
)

func TestUserRepository_Create_FindByEmail(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	repo := newUserRepo(db, tx)
	ctx := context.Background()

	user := model.NewUser(mtype.RoleUser)
	user.Email = mtype.Email("integ_findbyemail@example.com")
	require.NoError(t, repo.Create(ctx, user))

	found, err := repo.FindByEmail(ctx, user.Email)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Email, found.Email)
	assert.True(t, found.IsActive)
}

func TestUserRepository_FindByEmail_notFound(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	repo := newUserRepo(db, tx)
	ctx := context.Background()

	found, err := repo.FindByEmail(ctx, mtype.Email("nobody@example.com"))
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestUserRepository_GetByID(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	repo := newUserRepo(db, tx)
	ctx := context.Background()

	user := model.NewUser(mtype.RoleAdmin)
	require.NoError(t, repo.Create(ctx, user))

	found, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, mtype.Role(mtype.RoleAdmin), found.Role)
}

func TestUserRepository_GetByID_notFound(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	repo := newUserRepo(db, tx)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New())
	require.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestUserRepository_FindByProjectIDAndWallet(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	repo := newUserRepo(db, tx)
	ctx := context.Background()

	projectID := uuid.New()
	const wallet = "SolanaWalletTestAddress1111"

	user := model.NewUser(mtype.RoleUser)
	user.ProjectID = projectID
	user.WalletAddress = wallet
	require.NoError(t, repo.Create(ctx, user))

	found, err := repo.FindByProjectIDAndWallet(ctx, projectID, wallet)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)

	// Wrong wallet returns nil without error.
	notFound, err := repo.FindByProjectIDAndWallet(ctx, projectID, "WrongWallet")
	require.NoError(t, err)
	assert.Nil(t, notFound)

	// Right wallet, wrong project returns nil.
	notFound2, err := repo.FindByProjectIDAndWallet(ctx, uuid.New(), wallet)
	require.NoError(t, err)
	assert.Nil(t, notFound2)
}

func TestUserRepository_BlockProfile_UnblockProfile(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	repo := newUserRepo(db, tx)
	ctx := context.Background()

	user := model.NewUser(mtype.RoleUser)
	require.NoError(t, repo.Create(ctx, user))

	// Initially active.
	got, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.True(t, got.IsActive)

	// Block.
	require.NoError(t, repo.BlockProfile(ctx, user.ID))
	got, err = repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)

	// Unblock.
	require.NoError(t, repo.UnblockProfile(ctx, user.ID))
	got, err = repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.True(t, got.IsActive)
}

func TestUserRepository_CountNewUsersLast24Hours(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	repo := newUserRepo(db, tx)
	ctx := context.Background()

	before, err := repo.CountNewUsersLast24Hours(ctx)
	require.NoError(t, err)

	u1 := model.NewUser(mtype.RoleUser)
	u2 := model.NewUser(mtype.RolePartner)
	require.NoError(t, repo.Create(ctx, u1))
	require.NoError(t, repo.Create(ctx, u2))

	after, err := repo.CountNewUsersLast24Hours(ctx)
	require.NoError(t, err)
	assert.Equal(t, before+2, after)
}

func TestUserRepository_EmailUniqueness(t *testing.T) {
	db := openTestDB(t)
	tx := beginTx(t, db)
	repo := newUserRepo(db, tx)
	ctx := context.Background()

	email := mtype.Email("unique_check@example.com")

	u1 := model.NewUser(mtype.RoleUser)
	u1.Email = email
	require.NoError(t, repo.Create(ctx, u1))

	u2 := model.NewUser(mtype.RoleUser)
	u2.Email = email
	err := repo.Create(ctx, u2)
	require.Error(t, err, "duplicate email must be rejected by the DB")
}
