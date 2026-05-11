package v1

import (
	"time"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/service"
	"dd-prediction-api/internal/storage/click"
	"dd-prediction-api/pkg/apperrors"
	auth "dd-prediction-api/pkg/jwt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type UserHandler struct {
	UserService    *service.UserService
	DuelService    *service.DuelService
	ProjectService *service.ProjectService
	AuditLog       *click.AdminAuditLogRepository
}

func NewUserHandler(
	userService *service.UserService,
	duelService *service.DuelService,
	projectService *service.ProjectService,
	auditLog *click.AdminAuditLogRepository,
) *UserHandler {
	return &UserHandler{
		UserService:    userService,
		DuelService:    duelService,
		ProjectService: projectService,
		AuditLog:       auditLog,
	}
}

func (h *UserHandler) RegisterRoutes(app *fiber.App, a *AuthHandler) {
	userGroup := app.Group("/user", a.ApiKeyMiddleware, a.AuthMiddleware)
	{
		userGroup.Get("/me", h.GetMe)
		userGroup.Get("/me/financial-history", h.GetFinancialHistory)

		adminGroup := userGroup.Group("/", a.CheckForAdminPermissions)
		{
			adminGroup.Get("/:id", h.GetByID)
		}
	}

	projectGroup := app.Group("/project", a.AuthMiddleware, a.CheckForAdminPermissions)
	{
		projectGroup.Put("/:id/block-key", h.BlockAPIKey)
		projectGroup.Put("/:id/unblock-key", h.UnblockAPIKey)
	}
}

// GetMe godoc
//
//	@Summary		Get current user
//	@Description	Returns the profile of the currently authenticated user.
//	@Tags			user
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Param			X-API-Key		header		string					true	"Project API key"
//	@Success		200				{object}	model.User				"Current user profile"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/user/me [get]
func (h *UserHandler) GetMe(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	user, err := h.UserService.GetByID(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.JSON(user)
}

// GetByID godoc
//
//	@Summary		Get user by ID (admin)
//	@Description	Returns the profile of any user by their UUID. Requires admin permissions. The action is recorded in the audit log.
//	@Tags			user
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Param			id				path		string					true	"User UUID"
//	@Success		200				{object}	model.User				"User profile"
//	@Failure		400				{object}	apperrors.ErrorPublic	"Invalid user ID"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Admin permission required"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/user/{id} [get]
func (h *UserHandler) GetByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid user id")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	user, err := h.UserService.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	h.AuditLog.Insert(c.Context(), model.AdminAuditLog{
		ModeratorID: claims.UserID,
		Role:        uint8(claims.Role),
		Action:      model.AuditActionGetUser,
		IP:          c.IP(),
		Metadata:    model.AuditMeta("target_user_id", id),
		CreatedAt:   time.Now().UTC(),
	})

	return c.JSON(user)
}

// GetFinancialHistory godoc
//
//	@Summary		Get financial history
//	@Description	Returns the authenticated user's on-chain financial transaction history (duel predictions, refunds, rewards, commissions).
//	@Tags			user
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Param			X-API-Key		header		string					true	"Project API key"
//	@Success		200				{array}		model.FinancialTxEntry	"List of financial transactions"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/user/me/financial-history [get]
func (h *UserHandler) GetFinancialHistory(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	history, err := h.DuelService.GetFinancialHistory(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.JSON(history)
}

// BlockAPIKey godoc
//
//	@Summary		Block a project's API key (admin)
//	@Description	Blocks the API key for the given project, preventing further use. Requires admin permissions. The action is recorded in the audit log.
//	@Tags			project
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header	string	true	"Authorization Bearer token"
//	@Param			id				path	string	true	"Project UUID"
//	@Success		204				"API key blocked successfully"
//	@Failure		400				{object}	apperrors.ErrorPublic	"Invalid project ID"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Admin permission required"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/project/{id}/block-key [put]
func (h *UserHandler) BlockAPIKey(c fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid project id")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if err := h.ProjectService.BlockAPIKey(c.Context(), projectID); err != nil {
		return err
	}

	h.AuditLog.Insert(c.Context(), model.AdminAuditLog{
		ModeratorID: claims.UserID,
		Role:        uint8(claims.Role),
		Action:      model.AuditActionBlockAPIKey,
		IP:          c.IP(),
		Metadata:    model.AuditMeta("project_id", projectID),
		CreatedAt:   time.Now().UTC(),
	})

	return c.SendStatus(fiber.StatusNoContent)
}

// UnblockAPIKey godoc
//
//	@Summary		Unblock a project's API key (admin)
//	@Description	Re-enables a previously blocked API key for the given project. Requires admin permissions. The action is recorded in the audit log.
//	@Tags			project
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header	string	true	"Authorization Bearer token"
//	@Param			id				path	string	true	"Project UUID"
//	@Success		204				"API key unblocked successfully"
//	@Failure		400				{object}	apperrors.ErrorPublic	"Invalid project ID"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Admin permission required"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/project/{id}/unblock-key [put]
func (h *UserHandler) UnblockAPIKey(c fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid project id")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if err := h.ProjectService.UnblockAPIKey(c.Context(), projectID); err != nil {
		return err
	}

	h.AuditLog.Insert(c.Context(), model.AdminAuditLog{
		ModeratorID: claims.UserID,
		Role:        uint8(claims.Role),
		Action:      model.AuditActionUnblockAPIKey,
		IP:          c.IP(),
		Metadata:    model.AuditMeta("project_id", projectID),
		CreatedAt:   time.Now().UTC(),
	})

	return c.SendStatus(fiber.StatusNoContent)
}
