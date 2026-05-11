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

type AdminHandler struct {
	DuelQueue         *service.DuelQueue
	DuelService       *service.DuelService
	CommissionService *service.CommissionService
	FileService       *service.FileService
	AuditLog          *click.AdminAuditLogRepository
}

func NewAdminHandler(
	duelQueue *service.DuelQueue,
	duelService *service.DuelService,
	commissionService *service.CommissionService,
	fileService *service.FileService,
	auditLog *click.AdminAuditLogRepository,
) *AdminHandler {
	return &AdminHandler{
		DuelQueue:         duelQueue,
		DuelService:       duelService,
		CommissionService: commissionService,
		FileService:       fileService,
		AuditLog:          auditLog,
	}
}

func (h *AdminHandler) RegisterRoutes(app *fiber.App, a *AuthHandler) {
	admin := app.Group("/admin", a.AuthMiddleware, a.CheckForAdminRights)

	crypto := admin.Group("/duel")
	{
		crypto.Post("/", h.CreateCryptoDuel)
		crypto.Post("/upload-image", h.UploadDuelLogo)
		crypto.Put("/approve", h.ApproveCryptoDuel)
		crypto.Put("/resolve", h.ResolveCryptoDuel)
		crypto.Put("/cancel", h.CancelCryptoDuel)
		crypto.Put("/edit", h.EditDuel)
	}

	commission := admin.Group("/commission", a.CheckForAdminPermissions)
	{
		commission.Post("/dd-profit/claim", h.ClaimDDProfit)
		commission.Get("/dd-profit/history", h.GetDDProfitHistory)
	}
}

// ClaimDDProfit godoc
//
//	@Summary		Claim platform DD profit (admin)
//	@Description	Triggers a payout of all accumulated platform-side commission (DD profit) for completed billing periods. Requires admin permissions.
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string							true	"Authorization Bearer token"
//	@Success		200				{object}	model.ProjectCommissionClaim	"Created DD profit claim record"
//	@Failure		401				{object}	apperrors.ErrorPublic			"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic			"Admin permission required"
//	@Failure		500				{object}	apperrors.ErrorPublic			"Internal server error"
//	@Router			/admin/commission/dd-profit/claim [post]
func (h *AdminHandler) ClaimDDProfit(c fiber.Ctx) error {
	claim, err := h.CommissionService.ClaimDDProfit(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(claim)
}

// GetDDProfitHistory godoc
//
//	@Summary		Get DD profit claim history (admin)
//	@Description	Returns the full history of platform DD profit claims. Requires admin permissions.
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string							true	"Authorization Bearer token"
//	@Success		200				{array}		model.ProjectCommissionClaim	"List of DD profit claim records"
//	@Failure		401				{object}	apperrors.ErrorPublic			"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic			"Admin permission required"
//	@Failure		500				{object}	apperrors.ErrorPublic			"Internal server error"
//	@Router			/admin/commission/dd-profit/history [get]
func (h *AdminHandler) GetDDProfitHistory(c fiber.Ctx) error {
	history, err := h.CommissionService.GetDDClaimHistory(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(history)
}

// CreateCryptoDuel godoc
//
//	@Summary		Create a crypto duel (admin)
//	@Description	Creates a new crypto duel as an admin. The duel is immediately set to active status, bypassing the approval step. The action is recorded in the audit log.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Param			request			body		model.CreateDuelReq		true	"Duel creation request"
//	@Success		200				{object}	model.Duel				"Created duel"
//	@Failure		400				{object}	apperrors.ErrorPublic	"Invalid request data"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Admin permission required"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/admin/duel [post]
func (h *AdminHandler) CreateCryptoDuel(c fiber.Ctx) error {
	var req model.CreateDuelReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	duel, err := h.DuelService.CreateNewCryptoDuelAdmin(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	h.AuditLog.Insert(c.Context(), model.AdminAuditLog{
		ModeratorID: claims.UserID,
		Role:        uint8(claims.Role),
		Action:      model.AuditActionCreateCryptoDuel,
		IP:          c.IP(),
		Metadata:    model.AuditMeta("duel_id", duel.ID, "question", duel.Question),
		CreatedAt:   time.Now().UTC(),
	})

	return c.JSON(duel)
}

// ApproveCryptoDuel godoc
//
//	@Summary		Approve a crypto duel (admin)
//	@Description	Approves a pending crypto duel, making it visible and open for participants. The action is recorded in the audit log.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Param			request			body		model.DuelApproveReq	true	"Duel approve request"
//	@Success		200				{object}	model.Duel				"Approved duel"
//	@Failure		400				{object}	apperrors.ErrorPublic	"Invalid request data"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Admin permission required"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/admin/duel/approve [put]
func (h *AdminHandler) ApproveCryptoDuel(c fiber.Ctx) error {
	var req model.DuelApproveReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	resp, err := h.DuelQueue.ApproveCryptoDuel(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	h.AuditLog.Insert(c.Context(), model.AdminAuditLog{
		ModeratorID: claims.UserID,
		Role:        uint8(claims.Role),
		Action:      model.AuditActionApproveCryptoDuel,
		IP:          c.IP(),
		Metadata:    model.AuditMeta("duel_id", req.DuelID),
		CreatedAt:   time.Now().UTC(),
	})

	return c.JSON(resp)
}

// ResolveCryptoDuel godoc
//
//	@Summary		Resolve a crypto duel (admin)
//	@Description	Resolves a crypto duel by setting the winning answer and distributing rewards to winners. The action is recorded in the audit log.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string						true	"Authorization Bearer token"
//	@Param			request			body		model.DuelResolveReq		true	"Duel resolve request"
//	@Success		200				{object}	model.ResolveCryptoDuelResp	"Transaction hashes and resolved duel"
//	@Failure		400				{object}	apperrors.ErrorPublic		"Invalid request data"
//	@Failure		401				{object}	apperrors.ErrorPublic		"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic		"Admin permission required"
//	@Failure		500				{object}	apperrors.ErrorPublic		"Internal server error"
//	@Router			/admin/duel/resolve [put]
func (h *AdminHandler) ResolveCryptoDuel(c fiber.Ctx) error {
	var req model.DuelResolveReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	resp, err := h.DuelQueue.ResolveCryptoDuel(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	h.AuditLog.Insert(c.Context(), model.AdminAuditLog{
		ModeratorID: claims.UserID,
		Role:        uint8(claims.Role),
		Action:      model.AuditActionResolveCryptoDuel,
		IP:          c.IP(),
		Metadata:    model.AuditMeta("duel_id", req.DuelID, "answer", req.Answer),
		CreatedAt:   time.Now().UTC(),
	})

	return c.JSON(resp)
}

// EditDuel godoc
//
//	@Summary		Edit a duel (admin)
//	@Description	Updates editable fields of an existing duel (question, source of truth, logo URL, duel info, deadline). The action is recorded in the audit log.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header	string					true	"Authorization Bearer token"
//	@Param			request			body	model.DuelAdminEditReq	true	"Duel edit request"
//	@Success		200				"Duel updated successfully"
//	@Failure		400				{object}	apperrors.ErrorPublic	"Invalid request data"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Admin permission required"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/admin/duel/edit [put]
func (h *AdminHandler) EditDuel(c fiber.Ctx) error {
	var req model.DuelAdminEditReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if err := h.DuelService.EditDuel(c.Context(), &req); err != nil {
		return err
	}

	h.AuditLog.Insert(c.Context(), model.AdminAuditLog{
		ModeratorID: claims.UserID,
		Role:        uint8(claims.Role),
		Action:      model.AuditActionEditDuel,
		IP:          c.IP(),
		Metadata:    model.AuditMeta("duel_id", req.ID, "question", req.Question),
		CreatedAt:   time.Now().UTC(),
	})

	return c.SendStatus(fiber.StatusOK)
}

// UploadDuelLogo godoc
//
//	@Summary		Upload a duel logo (admin)
//	@Description	Uploads and processes a duel logo image. Returns the URL of the saved logo.
//	@Tags			admin
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Param			duel_logo		formData	file					true	"Duel logo image"
//	@Success		200				{object}	map[string]string		"duel_logo_url"
//	@Failure		400				{object}	apperrors.ErrorPublic	"Invalid request data"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Admin permission required"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/admin/duel/upload-image [post]
func (h *AdminHandler) UploadDuelLogo(c fiber.Ctx) error {
	duelLogo, err := c.FormFile("duel_logo")
	if err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	duelLogoURL, err := h.FileService.SaveDuelLogo(duelLogo)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"duel_logo_url": duelLogoURL,
	})
}

// CancelCryptoDuel godoc
//
//	@Summary		Cancel a crypto duel (admin)
//	@Description	Cancels a crypto duel and initiates refunds to all participants. The action is recorded in the audit log.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string						true	"Authorization Bearer token"
//	@Param			request			body		model.DuelCancelReq			true	"Duel cancel request"
//	@Success		200				{object}	model.CancelCryptoDuelResp	"Refund transaction hashes"
//	@Failure		400				{object}	apperrors.ErrorPublic		"Invalid request data"
//	@Failure		401				{object}	apperrors.ErrorPublic		"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic		"Admin permission required"
//	@Failure		500				{object}	apperrors.ErrorPublic		"Internal server error"
//	@Router			/admin/duel/cancel [put]
func (h *AdminHandler) CancelCryptoDuel(c fiber.Ctx) error {
	var req model.DuelCancelReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	resp, err := h.DuelQueue.CancelCryptoDuel(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	h.AuditLog.Insert(c.Context(), model.AdminAuditLog{
		ModeratorID: claims.UserID,
		Role:        uint8(claims.Role),
		Action:      model.AuditActionCancelCryptoDuel,
		IP:          c.IP(),
		Metadata:    model.AuditMeta("duel_id", req.DuelID, "reason", req.CancellationReason),
		CreatedAt:   time.Now().UTC(),
	})

	return c.JSON(resp)
}
