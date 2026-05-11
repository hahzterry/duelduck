package service

import (
	"context"

	"crypto/ed25519"
	"dd-prediction-api/config"
	"dd-prediction-api/internal/model"

	"dd-prediction-api/internal/storage/cache"
	"dd-prediction-api/internal/storage/cypher"
	"dd-prediction-api/internal/storage/repository"
	"dd-prediction-api/pkg/apikey"
	"dd-prediction-api/pkg/apperrors"
	repo "dd-prediction-api/pkg/repository"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ProjectService struct {
	ProjectRepository       repository.IProjectRepository
	UserRepository          repository.IUserRepository
	StatusHistoryRepository repository.IProjectStatusHistoryRepository
	TransactionManager      repo.ITransactionManager
	BlockedProjectCache     cache.IBlockedProjectCache
	privateKeyRepository    cypher.IPrivateKeyRepository
	apiKeySecret            string
}

func NewProjectService(
	c *config.Config,
	projectRepository *repository.ProjectRepository,
	userRepository *repository.UserRepository,
	statusHistoryRepository *repository.ProjectStatusHistoryRepository,
	transactionManager *repo.TransactionManager,
	blockedProjectCache *cache.BlockedProjectCache,
	privateKeyRepository *cypher.PrivateKeyRepository,
) *ProjectService {
	return &ProjectService{
		ProjectRepository:       projectRepository,
		UserRepository:          userRepository,
		StatusHistoryRepository: statusHistoryRepository,
		TransactionManager:      transactionManager,
		BlockedProjectCache:     blockedProjectCache,
		privateKeyRepository:    privateKeyRepository,
		apiKeySecret:            c.Auth.SecretSignKey,
	}
}

func (s *ProjectService) Create(
	ctx context.Context,
	partnerID uuid.UUID,
) (*model.Project, error) {
	exists, err := s.ProjectRepository.HasPartnerProject(ctx, partnerID)
	if exists {
		return nil, apperrors.AlreadyExist("partner has already created 1 project")
	}
	if err != nil {
		return nil, apperrors.Internal("failed to check partner project", err)
	}

	projectID := uuid.New()
	key := apikey.Generate(projectID, s.apiKeySecret)

	mnemonic, err := generateMnemonic()
	if err != nil {
		return nil, err
	}

	privateKey, err := generatePrivateKeyFromMnemonic(mnemonic)
	if err != nil {
		return nil, apperrors.Internal("failed to generate project wallet", err)
	}

	if err = s.privateKeyRepository.WritePrivateKey(ctx, projectID, mnemonic, ed25519.PrivateKey(privateKey)); err != nil {
		return nil, apperrors.Internal("failed to store project admin key", err)
	}

	project := model.NewProject(projectID, key, partnerID)
	project.WalletAddress = privateKey.PublicKey().String()

	err = s.TransactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		if err := s.ProjectRepository.WithTx(tx).Create(ctx, project); err != nil {
			return apperrors.Internal("failed to create project", err)
		}

		if err := s.UserRepository.WithTx(tx).SetProjectID(ctx, partnerID, project.ID); err != nil {
			return apperrors.Internal("failed to set project id", err)
		}

		if err := s.StatusHistoryRepository.WithTx(tx).Insert(ctx, project.ID, model.ProjectStatusActive); err != nil {
			return apperrors.Internal("failed to record initial project status", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) GetOwnProject(ctx context.Context, partnerID uuid.UUID) (*model.Project, error) {
	project, err := s.ProjectRepository.GetByPartnerID(ctx, partnerID)
	if err != nil {
		return nil, apperrors.Internal("failed to get project", err)
	}
	if project == nil {
		return nil, apperrors.NotFound("partner project not found")
	}
	return project, nil
}

func (s *ProjectService) Edit(ctx context.Context, partnerID uuid.UUID, req *model.EditProjectReq) (*model.Project, error) {
	project, err := s.ProjectRepository.GetByPartnerID(ctx, partnerID)
	if err != nil {
		return nil, apperrors.Internal("failed to get project", err)
	}
	if project == nil {
		return nil, apperrors.NotFound("partner project not found")
	}

	if err := s.ProjectRepository.Update(ctx, project.ID, req.Name, req.SiteURL); err != nil {
		return nil, apperrors.Internal("failed to update project", err)
	}

	project.Name = req.Name
	project.SiteURL = req.SiteURL
	return project, nil
}

// Disable sets status to disabled (by partner). Only works on active projects.
// A project blocked by admin cannot be disabled by the partner.
func (s *ProjectService) Disable(ctx context.Context, partnerID uuid.UUID) error {
	project, err := s.ProjectRepository.GetByPartnerID(ctx, partnerID)
	if err != nil {
		return apperrors.Internal("failed to get project", err)
	}
	if project == nil {
		return apperrors.NotFound("partner project not found")
	}
	if project.Status == model.ProjectStatusBlocked {
		return apperrors.Forbidden("project is blocked by Duel Duck and cannot be disabled by partner")
	}
	if project.Status == model.ProjectStatusDisabled {
		return nil // already disabled, idempotent
	}

	if err := s.ProjectRepository.SetStatus(ctx, project.ID, model.ProjectStatusDisabled); err != nil {
		return apperrors.Internal("failed to disable project", err)
	}
	if err := s.StatusHistoryRepository.Insert(ctx, project.ID, model.ProjectStatusDisabled); err != nil {
		return apperrors.Internal("failed to record status change", err)
	}
	return s.BlockedProjectCache.Block(ctx, project.ID)
}

// Enable restores a partner-disabled project back to active.
// Does not affect admin-blocked projects.
func (s *ProjectService) Enable(ctx context.Context, partnerID uuid.UUID) error {
	project, err := s.ProjectRepository.GetByPartnerID(ctx, partnerID)
	if err != nil {
		return apperrors.Internal("failed to get project", err)
	}
	if project == nil {
		return apperrors.NotFound("partner project not found")
	}
	if project.Status == model.ProjectStatusBlocked {
		return apperrors.Forbidden("project is blocked by Duel Duck and cannot be enabled by partner")
	}
	if project.Status == model.ProjectStatusActive {
		return nil // already active, idempotent
	}

	if err := s.ProjectRepository.SetStatus(ctx, project.ID, model.ProjectStatusActive); err != nil {
		return apperrors.Internal("failed to enable project", err)
	}
	if err := s.StatusHistoryRepository.Insert(ctx, project.ID, model.ProjectStatusActive); err != nil {
		return apperrors.Internal("failed to record status change", err)
	}
	return s.BlockedProjectCache.Unblock(ctx, project.ID)
}

// UpdatePermissions updates user-level feature flags for the project.
// Self-resolve can only be enabled when duel creation is also enabled.
func (s *ProjectService) UpdatePermissions(ctx context.Context, partnerID uuid.UUID, req *model.UpdateProjectPermissionsReq) error {
	project, err := s.ProjectRepository.GetByPartnerID(ctx, partnerID)
	if err != nil {
		return apperrors.Internal("failed to get project", err)
	}
	if project == nil {
		return apperrors.NotFound("partner project not found")
	}

	selfResolve := req.IsSelfResolvedEnabled
	if !req.IsUsersDuelsEnabled {
		selfResolve = false
	}

	if err := s.ProjectRepository.UpdatePermissions(ctx, project.ID, req.IsUsersDuelsEnabled, selfResolve); err != nil {
		return apperrors.Internal("failed to update project permissions", err)
	}
	return nil
}

func (s *ProjectService) ParseAPIKey(key string) (uuid.UUID, bool) {
	return apikey.Parse(key, s.apiKeySecret)
}

// IsAPIKeyBlocked checks Redis first (O(1)); no DB hit on the happy path.
func (s *ProjectService) IsAPIKeyBlocked(ctx context.Context, projectID uuid.UUID) (bool, error) {
	return s.BlockedProjectCache.IsBlocked(ctx, projectID)
}

// BlockAPIKey marks the project as blocked by admin in DB and propagates to Redis.
func (s *ProjectService) BlockAPIKey(ctx context.Context, projectID uuid.UUID) error {
	if err := s.ProjectRepository.SetBlocked(ctx, projectID, true); err != nil {
		return apperrors.Internal("failed to block API key", err)
	}
	if err := s.StatusHistoryRepository.Insert(ctx, projectID, model.ProjectStatusBlocked); err != nil {
		return apperrors.Internal("failed to record block status", err)
	}
	return s.BlockedProjectCache.Block(ctx, projectID)
}

// UnblockAPIKey removes the admin block from DB and Redis.
func (s *ProjectService) UnblockAPIKey(ctx context.Context, projectID uuid.UUID) error {
	if err := s.ProjectRepository.SetBlocked(ctx, projectID, false); err != nil {
		return apperrors.Internal("failed to unblock API key", err)
	}
	if err := s.StatusHistoryRepository.Insert(ctx, projectID, model.ProjectStatusActive); err != nil {
		return apperrors.Internal("failed to record unblock status", err)
	}
	return s.BlockedProjectCache.Unblock(ctx, projectID)
}
