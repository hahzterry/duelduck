package v1

import (
	"time"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/service"
	"dd-prediction-api/internal/storage/click"
	"dd-prediction-api/pkg/apperrors"
	auth "dd-prediction-api/pkg/jwt"

	"github.com/gofiber/fiber/v3"
)

type ProjectHandler struct {
	ProjectService *service.ProjectService
	AuditLog       *click.AdminAuditLogRepository
}

func NewProjectHandler(
	projectService *service.ProjectService,
	auditLog *click.AdminAuditLogRepository,
) *ProjectHandler {
	return &ProjectHandler{
		ProjectService: projectService,
		AuditLog:       auditLog,
	}
}

func (h *ProjectHandler) RegisterRoutes(
	app *fiber.App,
	a *AuthHandler,
) {
	partner := app.Group("/me", a.AuthMiddleware, a.CheckForPartnerPermissions)
	{
		partner.Post("/project", h.CreateProject)
		partner.Get("/project", h.GetProject)
		partner.Put("/project", h.EditProject)
		partner.Put("/project/disable", h.DisableProject)
		partner.Put("/project/enable", h.EnableProject)
		partner.Put("/project/permissions", h.UpdatePermissions)
	}
}

// CreateProject godoc
//
//	@Summary		Create a partner project
//	@Description	Creates a new project for the authenticated partner. Returns the project with its generated API key. One project per partner.
//	@Tags			project
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Success		200				{object}	model.Project			"Created project"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Partner permission required"
//	@Failure		409				{object}	apperrors.ErrorPublic	"Project already exists"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/me/project [post]
func (h *ProjectHandler) CreateProject(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	project, err := h.ProjectService.Create(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	h.AuditLog.Insert(c.Context(), model.AdminAuditLog{
		ModeratorID: claims.UserID,
		Role:        uint8(claims.Role),
		Action:      model.AuditActionCreateProject,
		IP:          c.IP(),
		Metadata:    model.AuditMeta("project_id", project.ID),
		CreatedAt:   time.Now().UTC(),
	})

	return c.JSON(project)
}

// GetProject godoc
//
//	@Summary		Get own project
//	@Description	Returns the partner's project including name, site URL, status, and status history.
//	@Tags			project
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Success		200				{object}	model.Project			"Project details"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Partner permission required"
//	@Failure		404				{object}	apperrors.ErrorPublic	"Project not found"
//	@Router			/me/project [get]
func (h *ProjectHandler) GetProject(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	project, err := h.ProjectService.GetOwnProject(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.JSON(project)
}

// EditProject godoc
//
//	@Summary		Edit partner project
//	@Description	Updates the project name and/or site URL.
//	@Tags			project
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Param			request			body		model.EditProjectReq	true	"Edit request"
//	@Success		200				{object}	model.Project			"Updated project"
//	@Failure		400				{object}	apperrors.ErrorPublic	"Invalid request"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Partner permission required"
//	@Failure		404				{object}	apperrors.ErrorPublic	"Project not found"
//	@Router			/me/project [put]
func (h *ProjectHandler) EditProject(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	var req model.EditProjectReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request body", err)
	}

	project, err := h.ProjectService.Edit(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.JSON(project)
}

// DisableProject godoc
//
//	@Summary		Disable partner project
//	@Description	Sets project status to disabled. API key will be rejected until the project is re-enabled. Cannot be used on admin-blocked projects.
//	@Tags			project
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header	string	true	"Authorization Bearer token"
//	@Success		200				"Project disabled"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Forbidden — project is admin-blocked"
//	@Failure		404				{object}	apperrors.ErrorPublic	"Project not found"
//	@Router			/me/project/disable [put]
func (h *ProjectHandler) DisableProject(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if err := h.ProjectService.Disable(c.Context(), claims.UserID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// EnableProject godoc
//
//	@Summary		Re-enable partner project
//	@Description	Restores a partner-disabled project to active status. Does not affect admin-blocked projects.
//	@Tags			project
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header	string	true	"Authorization Bearer token"
//	@Success		200				"Project enabled"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Forbidden — project is admin-blocked"
//	@Failure		404				{object}	apperrors.ErrorPublic	"Project not found"
//	@Router			/me/project/enable [put]
func (h *ProjectHandler) EnableProject(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if err := h.ProjectService.Enable(c.Context(), claims.UserID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// UpdatePermissions godoc
//
//	@Summary		Update user permission flags
//	@Description	Enables or disables duel creation and self-resolve for users of this project's site. Self-resolve is automatically disabled when duel creation is disabled.
//	@Tags			project
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header	string								true	"Authorization Bearer token"
//	@Param			request			body	model.UpdateProjectPermissionsReq	true	"Permission flags"
//	@Success		200				"Permissions updated"
//	@Failure		400				{object}	apperrors.ErrorPublic	"Invalid request"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Partner permission required"
//	@Failure		404				{object}	apperrors.ErrorPublic	"Project not found"
//	@Router			/me/project/permissions [put]
func (h *ProjectHandler) UpdatePermissions(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	var req model.UpdateProjectPermissionsReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request body", err)
	}

	if err := h.ProjectService.UpdatePermissions(c.Context(), claims.UserID, &req); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}
