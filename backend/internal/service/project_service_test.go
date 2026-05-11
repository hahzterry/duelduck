package service

import (
	"context"
	"crypto/ed25519"
	"errors"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/storage/cypher"
	"dd-prediction-api/internal/storage/repository"
	"dd-prediction-api/pkg/apikey"
	"dd-prediction-api/pkg/mtype"
)

// ── mocks ────────────────────────────────────────────────────────────────────

type mockProjectRepo struct{ mock.Mock }

func (m *mockProjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	args := m.Called(ctx, id)
	res, _ := args.Get(0).(*model.Project)
	return res, args.Error(1)
}
func (m *mockProjectRepo) GetByAPIKey(ctx context.Context, k string) (*model.Project, error) {
	args := m.Called(ctx, k)
	res, _ := args.Get(0).(*model.Project)
	return res, args.Error(1)
}
func (m *mockProjectRepo) IsBlocked(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}
func (m *mockProjectRepo) SetBlocked(ctx context.Context, id uuid.UUID, blocked bool) error {
	return m.Called(ctx, id, blocked).Error(0)
}
func (m *mockProjectRepo) Create(ctx context.Context, p *model.Project) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockProjectRepo) HasPartnerProject(ctx context.Context, partnerID uuid.UUID) (bool, error) {
	args := m.Called(ctx, partnerID)
	return args.Bool(0), args.Error(1)
}
func (m *mockProjectRepo) GetByPartnerID(ctx context.Context, partnerID uuid.UUID) (*model.Project, error) {
	args := m.Called(ctx, partnerID)
	res, _ := args.Get(0).(*model.Project)
	return res, args.Error(1)
}
func (m *mockProjectRepo) GetAll(ctx context.Context) ([]model.Project, error) {
	args := m.Called(ctx)
	res, _ := args.Get(0).([]model.Project)
	return res, args.Error(1)
}
func (m *mockProjectRepo) SetStatus(ctx context.Context, id uuid.UUID, status model.ProjectStatus) error {
	return m.Called(ctx, id, status).Error(0)
}
func (m *mockProjectRepo) Update(ctx context.Context, id uuid.UUID, name, siteURL string) error {
	return m.Called(ctx, id, name, siteURL).Error(0)
}
func (m *mockProjectRepo) UpdatePermissions(ctx context.Context, id uuid.UUID, duelsEnabled, selfResolveEnabled bool) error {
	return m.Called(ctx, id, duelsEnabled, selfResolveEnabled).Error(0)
}
func (m *mockProjectRepo) WithTx(_ bun.Tx) repository.IProjectRepository { return m }

type mockStatusHistoryRepo struct{ mock.Mock }

func (m *mockStatusHistoryRepo) Insert(ctx context.Context, projectID uuid.UUID, status model.ProjectStatus) error {
	return m.Called(ctx, projectID, status).Error(0)
}
func (m *mockStatusHistoryRepo) WithTx(_ bun.Tx) repository.IProjectStatusHistoryRepository { return m }

type mockBlockedCache struct{ mock.Mock }

func (m *mockBlockedCache) Block(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockBlockedCache) Unblock(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockBlockedCache) IsBlocked(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

type mockPrivateKeyRepo struct{ mock.Mock }

func (m *mockPrivateKeyRepo) WritePrivateKey(ctx context.Context, id uuid.UUID, mnemonic string, pk ed25519.PrivateKey) error {
	return m.Called(ctx, id, mnemonic, pk).Error(0)
}
func (m *mockPrivateKeyRepo) GetPrivateKeyBase58(ctx context.Context, id uuid.UUID) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}
func (m *mockPrivateKeyRepo) GetWalletData(ctx context.Context, id uuid.UUID) (*cypher.VaultCustomData, error) {
	args := m.Called(ctx, id)
	res, _ := args.Get(0).(*cypher.VaultCustomData)
	return res, args.Error(1)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newTestProjectService(repo *mockProjectRepo, cache *mockBlockedCache, secret string) *ProjectService {
	histRepo := &mockStatusHistoryRepo{}
	histRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	return &ProjectService{
		ProjectRepository:       repo,
		BlockedProjectCache:     cache,
		StatusHistoryRepository: histRepo,
		apiKeySecret:            secret,
	}
}

// ── BlockAPIKey ───────────────────────────────────────────────────────────────

func TestBlockAPIKey(t *testing.T) {
	projectID := uuid.New()

	tests := []struct {
		name    string
		setup   func(repo *mockProjectRepo, cache *mockBlockedCache)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(repo *mockProjectRepo, cache *mockBlockedCache) {
				repo.On("SetBlocked", mock.Anything, projectID, true).Return(nil)
				cache.On("Block", mock.Anything, projectID).Return(nil)
			},
		},
		{
			name: "repo error stops before cache",
			setup: func(repo *mockProjectRepo, cache *mockBlockedCache) {
				repo.On("SetBlocked", mock.Anything, projectID, true).Return(errors.New("db error"))
				// cache.Block must NOT be called
			},
			wantErr: true,
		},
		{
			name: "cache error propagates",
			setup: func(repo *mockProjectRepo, cache *mockBlockedCache) {
				repo.On("SetBlocked", mock.Anything, projectID, true).Return(nil)
				cache.On("Block", mock.Anything, projectID).Return(errors.New("redis error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProjectRepo{}
			cache := &mockBlockedCache{}
			tt.setup(repo, cache)

			err := newTestProjectService(repo, cache, "secret").BlockAPIKey(context.Background(), projectID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			repo.AssertExpectations(t)
			cache.AssertExpectations(t)
		})
	}
}

// ── UnblockAPIKey ─────────────────────────────────────────────────────────────

func TestUnblockAPIKey(t *testing.T) {
	projectID := uuid.New()

	tests := []struct {
		name    string
		setup   func(repo *mockProjectRepo, cache *mockBlockedCache)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(repo *mockProjectRepo, cache *mockBlockedCache) {
				repo.On("SetBlocked", mock.Anything, projectID, false).Return(nil)
				cache.On("Unblock", mock.Anything, projectID).Return(nil)
			},
		},
		{
			name: "repo error stops before cache",
			setup: func(repo *mockProjectRepo, cache *mockBlockedCache) {
				repo.On("SetBlocked", mock.Anything, projectID, false).Return(errors.New("db down"))
			},
			wantErr: true,
		},
		{
			name: "cache error propagates",
			setup: func(repo *mockProjectRepo, cache *mockBlockedCache) {
				repo.On("SetBlocked", mock.Anything, projectID, false).Return(nil)
				cache.On("Unblock", mock.Anything, projectID).Return(errors.New("redis down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProjectRepo{}
			cache := &mockBlockedCache{}
			tt.setup(repo, cache)

			err := newTestProjectService(repo, cache, "secret").UnblockAPIKey(context.Background(), projectID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			repo.AssertExpectations(t)
			cache.AssertExpectations(t)
		})
	}
}

// ── IsAPIKeyBlocked ───────────────────────────────────────────────────────────

func TestIsAPIKeyBlocked(t *testing.T) {
	projectID := uuid.New()

	tests := []struct {
		name        string
		setup       func(cache *mockBlockedCache)
		wantBlocked bool
		wantErr     bool
	}{
		{
			name: "blocked",
			setup: func(cache *mockBlockedCache) {
				cache.On("IsBlocked", mock.Anything, projectID).Return(true, nil)
			},
			wantBlocked: true,
		},
		{
			name: "not blocked",
			setup: func(cache *mockBlockedCache) {
				cache.On("IsBlocked", mock.Anything, projectID).Return(false, nil)
			},
			wantBlocked: false,
		},
		{
			name: "cache error propagates",
			setup: func(cache *mockBlockedCache) {
				cache.On("IsBlocked", mock.Anything, projectID).Return(false, errors.New("redis timeout"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &mockBlockedCache{}
			tt.setup(cache)

			blocked, err := newTestProjectService(&mockProjectRepo{}, cache, "secret").
				IsAPIKeyBlocked(context.Background(), projectID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantBlocked, blocked)
			cache.AssertExpectations(t)
		})
	}
}

// ── ParseAPIKey ───────────────────────────────────────────────────────────────

func TestParseAPIKey(t *testing.T) {
	const secret = "test-signing-key"
	projectID := uuid.New()
	validKey := apikey.Generate(projectID, secret)

	tests := []struct {
		name   string
		key    string
		wantID uuid.UUID
		wantOK bool
	}{
		{
			name:   "valid key returns correct project id",
			key:    validKey,
			wantID: projectID,
			wantOK: true,
		},
		{
			name:   "wrong secret",
			key:    apikey.Generate(projectID, "other-secret"),
			wantOK: false,
		},
		{
			name:   "empty key",
			key:    "",
			wantOK: false,
		},
		{
			name:   "garbage string",
			key:    "not-a-valid-key-at-all",
			wantOK: false,
		},
		{
			name:   "valid key with different project id still parses",
			key:    apikey.Generate(uuid.New(), secret),
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestProjectService(&mockProjectRepo{}, &mockBlockedCache{}, secret)
			gotID, ok := svc.ParseAPIKey(tt.key)

			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK && tt.wantID != uuid.Nil {
				assert.Equal(t, tt.wantID, gotID)
			}
		})
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) SetProjectID(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) error {
	return m.Called(ctx, userID, projectID).Error(0)
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	res, _ := args.Get(0).(*model.User)
	return res, args.Error(1)
}
func (m *mockUserRepo) FindByProjectIDAndWallet(ctx context.Context, projectID uuid.UUID, walletAddress string) (*model.User, error) {
	args := m.Called(ctx, projectID, walletAddress)
	res, _ := args.Get(0).(*model.User)
	return res, args.Error(1)
}
func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email mtype.Email) (*model.User, error) {
	args := m.Called(ctx, email)
	res, _ := args.Get(0).(*model.User)
	return res, args.Error(1)
}
func (m *mockUserRepo) BlockProfile(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockUserRepo) UnblockProfile(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockUserRepo) WithTx(_ bun.Tx) repository.IUserRepository { return m }

type mockTransactionManager struct{ mock.Mock }

func (m *mockTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context, bun.Tx) error) error {
	return fn(ctx, bun.Tx{})
}

func TestCreate(t *testing.T) {
	partnerID := uuid.New()
	const secret = "test-secret"

	tests := []struct {
		name    string
		setup   func(repo *mockProjectRepo, userRepo *mockUserRepo, pkRepo *mockPrivateKeyRepo)
		wantErr bool
		check   func(t *testing.T, project *model.Project)
	}{
		{
			name: "success — project has valid wallet address and api key",
			setup: func(repo *mockProjectRepo, userRepo *mockUserRepo, pkRepo *mockPrivateKeyRepo) {
				repo.On("HasPartnerProject", mock.Anything, partnerID).Return(false, nil)
				pkRepo.On("WritePrivateKey", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(p *model.Project) bool {
					return p.PartnerID == partnerID && p.APIKey != "" && p.WalletAddress != ""
				})).Return(nil)
				userRepo.On("SetProjectID", mock.Anything, partnerID, mock.Anything).Return(nil)
			},
			check: func(t *testing.T, project *model.Project) {
				require.NotNil(t, project)
				assert.Equal(t, partnerID, project.PartnerID)
				assert.NotEmpty(t, project.APIKey)
				assert.False(t, project.IsBlocked)
				assert.NotEmpty(t, project.WalletAddress)
				_, err := solana.PublicKeyFromBase58(project.WalletAddress)
				assert.NoError(t, err, "WalletAddress must be a valid Solana public key")
			},
		},
		{
			name: "partner already has project returns error",
			setup: func(repo *mockProjectRepo, userRepo *mockUserRepo, pkRepo *mockPrivateKeyRepo) {
				repo.On("HasPartnerProject", mock.Anything, partnerID).Return(true, nil)
			},
			wantErr: true,
		},
		{
			name: "repo check error propagates",
			setup: func(repo *mockProjectRepo, userRepo *mockUserRepo, pkRepo *mockPrivateKeyRepo) {
				repo.On("HasPartnerProject", mock.Anything, partnerID).Return(false, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "vault write error stops creation before db",
			setup: func(repo *mockProjectRepo, userRepo *mockUserRepo, pkRepo *mockPrivateKeyRepo) {
				repo.On("HasPartnerProject", mock.Anything, partnerID).Return(false, nil)
				pkRepo.On("WritePrivateKey", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("vault unavailable"))
				// DB must NOT be called
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProjectRepo{}
			userRepo := &mockUserRepo{}
			pkRepo := &mockPrivateKeyRepo{}
			tt.setup(repo, userRepo, pkRepo)

			histRepo := &mockStatusHistoryRepo{}
			histRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			svc := &ProjectService{
				ProjectRepository:       repo,
				UserRepository:          userRepo,
				TransactionManager:      &mockTransactionManager{},
				BlockedProjectCache:     &mockBlockedCache{},
				StatusHistoryRepository: histRepo,
				privateKeyRepository:    pkRepo,
				apiKeySecret:            secret,
			}

			project, err := svc.Create(context.Background(), partnerID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, project)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, project)
				}
			}
			repo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			pkRepo.AssertExpectations(t)
		})
	}
}

// ── GetOwnProject ─────────────────────────────────────────────────────────────

func TestGetOwnProject(t *testing.T) {
	partnerID := uuid.New()
	projectID := uuid.New()
	project := &model.Project{ID: projectID, PartnerID: partnerID}

	tests := []struct {
		name    string
		setup   func(repo *mockProjectRepo)
		wantErr bool
	}{
		{
			name: "success returns project",
			setup: func(repo *mockProjectRepo) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
			},
		},
		{
			name: "project not found returns error",
			setup: func(repo *mockProjectRepo) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "repo error propagates",
			setup: func(repo *mockProjectRepo) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProjectRepo{}
			cache := &mockBlockedCache{}
			tt.setup(repo)

			result, err := newTestProjectService(repo, cache, "secret").GetOwnProject(context.Background(), partnerID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, projectID, result.ID)
			}
			repo.AssertExpectations(t)
		})
	}
}

// ── Edit ──────────────────────────────────────────────────────────────────────

func TestEdit(t *testing.T) {
	partnerID := uuid.New()
	projectID := uuid.New()
	project := &model.Project{ID: projectID, PartnerID: partnerID, Name: "old", SiteURL: "https://old.com"}
	req := &model.EditProjectReq{Name: "new", SiteURL: "https://new.com"}

	tests := []struct {
		name    string
		setup   func(repo *mockProjectRepo)
		wantErr bool
	}{
		{
			name: "success updates and returns project",
			setup: func(repo *mockProjectRepo) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				repo.On("Update", mock.Anything, projectID, "new", "https://new.com").Return(nil)
			},
		},
		{
			name: "project not found returns error",
			setup: func(repo *mockProjectRepo) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "update repo error propagates",
			setup: func(repo *mockProjectRepo) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				repo.On("Update", mock.Anything, projectID, "new", "https://new.com").Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProjectRepo{}
			cache := &mockBlockedCache{}
			tt.setup(repo)

			result, err := newTestProjectService(repo, cache, "secret").Edit(context.Background(), partnerID, req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "new", result.Name)
				assert.Equal(t, "https://new.com", result.SiteURL)
			}
			repo.AssertExpectations(t)
		})
	}
}

// ── Disable ───────────────────────────────────────────────────────────────────

func TestDisable(t *testing.T) {
	partnerID := uuid.New()
	projectID := uuid.New()

	tests := []struct {
		name    string
		project *model.Project
		setup   func(repo *mockProjectRepo, cache *mockBlockedCache)
		wantErr bool
	}{
		{
			name:    "success disables active project",
			project: &model.Project{ID: projectID, PartnerID: partnerID, Status: model.ProjectStatusActive},
			setup: func(repo *mockProjectRepo, cache *mockBlockedCache) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(&model.Project{ID: projectID, PartnerID: partnerID, Status: model.ProjectStatusActive}, nil)
				repo.On("SetStatus", mock.Anything, projectID, model.ProjectStatusDisabled).Return(nil)
				cache.On("Block", mock.Anything, projectID).Return(nil)
			},
		},
		{
			name:    "already disabled is idempotent",
			project: &model.Project{ID: projectID, PartnerID: partnerID, Status: model.ProjectStatusDisabled},
			setup: func(repo *mockProjectRepo, cache *mockBlockedCache) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(&model.Project{ID: projectID, Status: model.ProjectStatusDisabled}, nil)
			},
		},
		{
			name: "blocked project returns forbidden error",
			setup: func(repo *mockProjectRepo, cache *mockBlockedCache) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(&model.Project{ID: projectID, Status: model.ProjectStatusBlocked}, nil)
			},
			wantErr: true,
		},
		{
			name: "project not found returns error",
			setup: func(repo *mockProjectRepo, cache *mockBlockedCache) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProjectRepo{}
			cache := &mockBlockedCache{}
			tt.setup(repo, cache)

			err := newTestProjectService(repo, cache, "secret").Disable(context.Background(), partnerID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			repo.AssertExpectations(t)
			cache.AssertExpectations(t)
		})
	}
}

// ── Enable ────────────────────────────────────────────────────────────────────

func TestEnable(t *testing.T) {
	partnerID := uuid.New()
	projectID := uuid.New()

	tests := []struct {
		name    string
		setup   func(repo *mockProjectRepo, cache *mockBlockedCache)
		wantErr bool
	}{
		{
			name: "success enables disabled project",
			setup: func(repo *mockProjectRepo, cache *mockBlockedCache) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(&model.Project{ID: projectID, Status: model.ProjectStatusDisabled}, nil)
				repo.On("SetStatus", mock.Anything, projectID, model.ProjectStatusActive).Return(nil)
				cache.On("Unblock", mock.Anything, projectID).Return(nil)
			},
		},
		{
			name: "already active is idempotent",
			setup: func(repo *mockProjectRepo, cache *mockBlockedCache) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(&model.Project{ID: projectID, Status: model.ProjectStatusActive}, nil)
			},
		},
		{
			name: "blocked project returns forbidden error",
			setup: func(repo *mockProjectRepo, cache *mockBlockedCache) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(&model.Project{ID: projectID, Status: model.ProjectStatusBlocked}, nil)
			},
			wantErr: true,
		},
		{
			name: "project not found returns error",
			setup: func(repo *mockProjectRepo, cache *mockBlockedCache) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProjectRepo{}
			cache := &mockBlockedCache{}
			tt.setup(repo, cache)

			err := newTestProjectService(repo, cache, "secret").Enable(context.Background(), partnerID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			repo.AssertExpectations(t)
			cache.AssertExpectations(t)
		})
	}
}

// ── UpdatePermissions ─────────────────────────────────────────────────────────

func TestUpdatePermissions(t *testing.T) {
	partnerID := uuid.New()
	projectID := uuid.New()
	project := &model.Project{ID: projectID, PartnerID: partnerID, Status: model.ProjectStatusActive}

	tests := []struct {
		name    string
		req     *model.UpdateProjectPermissionsReq
		setup   func(repo *mockProjectRepo)
		wantErr bool
	}{
		{
			name: "success enables both flags",
			req:  &model.UpdateProjectPermissionsReq{IsUsersDuelsEnabled: true, IsSelfResolvedEnabled: true},
			setup: func(repo *mockProjectRepo) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				repo.On("UpdatePermissions", mock.Anything, projectID, true, true).Return(nil)
			},
		},
		{
			name: "disabling duels forces self-resolve off",
			req:  &model.UpdateProjectPermissionsReq{IsUsersDuelsEnabled: false, IsSelfResolvedEnabled: true},
			setup: func(repo *mockProjectRepo) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				repo.On("UpdatePermissions", mock.Anything, projectID, false, false).Return(nil)
			},
		},
		{
			name: "project not found returns error",
			req:  &model.UpdateProjectPermissionsReq{IsUsersDuelsEnabled: true},
			setup: func(repo *mockProjectRepo) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "repo error propagates",
			req:  &model.UpdateProjectPermissionsReq{IsUsersDuelsEnabled: true, IsSelfResolvedEnabled: false},
			setup: func(repo *mockProjectRepo) {
				repo.On("GetByPartnerID", mock.Anything, partnerID).Return(project, nil)
				repo.On("UpdatePermissions", mock.Anything, projectID, true, false).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProjectRepo{}
			cache := &mockBlockedCache{}
			tt.setup(repo)

			err := newTestProjectService(repo, cache, "secret").UpdatePermissions(context.Background(), partnerID, tt.req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}
