package v1

import (
	"time"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/service"
	"dd-prediction-api/pkg/apperrors"
	auth "dd-prediction-api/pkg/jwt"

	"github.com/gofiber/fiber/v3"
)

type CommissionHandler struct {
	CommissionService *service.CommissionService
}

func NewCommissionHandler(commissionService *service.CommissionService) *CommissionHandler {
	return &CommissionHandler{CommissionService: commissionService}
}

func (h *CommissionHandler) RegisterRoutes(app *fiber.App, a *AuthHandler) {
	partner := app.Group("/me", a.ApiKeyMiddleware, a.AuthMiddleware, a.CheckForPartnerPermissions)
	{
		partner.Get("/dashboard", h.GetDashboard)

		commission := partner.Group("/commission")
		{
			commission.Post("/claim", h.ClaimProject)
			commission.Get("/history", h.GetProjectHistory)
		}

		analytics := partner.Group("/analytics")
		{
			analytics.Get("/daily-income", h.GetDailyIncome)
			analytics.Get("/monthly-active-users", h.GetMonthlyActiveUsers)
			analytics.Get("/dd-commission-monthly", h.GetMonthlyDDCommission)
		}

		partner.Get("/transactions", h.GetProjectTransactions)
	}
}

// GetDashboard godoc
//
//	@Summary		Get partner dashboard
//	@Description	Returns the partner's project dashboard: status, total/monthly volume, commission breakdown by symbol, and net profit after DD fee.
//	@Tags			partner
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Param			X-API-Key		header		string					true	"Project API key"
//	@Success		200				{object}	model.PartnerDashboard	"Partner dashboard metrics"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Partner permission required"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/me/dashboard [get]
func (h *CommissionHandler) GetDashboard(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	dashboard, err := h.CommissionService.GetPartnerDashboard(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.JSON(dashboard)
}

// ClaimProject godoc
//
//	@Summary		Claim partner commission
//	@Description	Claims all accumulated partner commission for completed billing periods and sends it to the provided wallet address.
//	@Tags			partner
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string							true	"Authorization Bearer token"
//	@Param			request			body		model.ProjectCommissionClaimReq	true	"Wallet address to send commission to"
//	@Success		200				{object}	model.ProjectCommissionClaim	"Created commission claim record"
//	@Failure		400				{object}	apperrors.ErrorPublic			"Invalid request or nothing to claim"
//	@Failure		401				{object}	apperrors.ErrorPublic			"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic			"Partner permission required"
//	@Failure		500				{object}	apperrors.ErrorPublic			"Internal server error"
//	@Router			/me/commission/claim [post]
func (h *CommissionHandler) ClaimProject(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	var req model.ProjectCommissionClaimReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request body", err)
	}
	if req.WalletAddress == "" {
		return apperrors.BadRequest("wallet_address is required")
	}

	claim, err := h.CommissionService.ClaimProjectCommission(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.JSON(claim)
}

// GetProjectHistory godoc
//
//	@Summary		Get partner commission claim history
//	@Description	Returns the full history of partner commission claims for the authenticated user's project.
//	@Tags			partner
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string							true	"Authorization Bearer token"
//	@Success		200				{array}		model.ProjectCommissionClaim	"List of commission claim records"
//	@Failure		401				{object}	apperrors.ErrorPublic			"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic			"Partner permission required"
//	@Failure		500				{object}	apperrors.ErrorPublic			"Internal server error"
//	@Router			/me/commission/history [get]
func (h *CommissionHandler) GetProjectHistory(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	history, err := h.CommissionService.GetProjectClaimHistory(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.JSON(history)
}

// GetDailyIncome godoc
//
//	@Summary		Get daily income analytics
//	@Description	Returns daily commission income (USDC equivalent) for the partner's project within a date range. Query params: from, to (format: 2006-01-02). Defaults to current month.
//	@Tags			partner
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string						true	"Authorization Bearer token"
//	@Param			X-API-Key		header		string						true	"Project API key"
//	@Param			from			query		string						false	"Start date (YYYY-MM-DD)"
//	@Param			to				query		string						false	"End date (YYYY-MM-DD)"
//	@Success		200				{array}		model.PartnerDailyIncome	"Daily income list"
//	@Failure		400				{object}	apperrors.ErrorPublic		"Invalid date format"
//	@Failure		401				{object}	apperrors.ErrorPublic		"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic		"Partner permission required"
//	@Router			/me/analytics/daily-income [get]
func (h *CommissionHandler) GetDailyIncome(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	from, to, err := parseDateRange(c)
	if err != nil {
		return err
	}

	result, err := h.CommissionService.GetDailyIncome(c.Context(), claims.UserID, from, to)
	if err != nil {
		return err
	}

	return c.JSON(result)
}

// GetMonthlyActiveUsers godoc
//
//	@Summary		Get monthly active users analytics
//	@Description	Returns unique user counts per calendar month. A user is active if they placed at least one prediction in that month.
//	@Tags			partner
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Param			X-API-Key		header		string					true	"Project API key"
//	@Success		200				{array}		model.PartnerMonthlyMAU	"Monthly active user counts"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Partner permission required"
//	@Router			/me/analytics/monthly-active-users [get]
func (h *CommissionHandler) GetMonthlyActiveUsers(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	result, err := h.CommissionService.GetMonthlyActiveUsers(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.JSON(result)
}

// GetMonthlyDDCommission godoc
//
//	@Summary		Get DD commission breakdown by month
//	@Description	Returns per-billing-month DD commission amounts and rates, derived from commission accruals.
//	@Tags			partner
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string								true	"Authorization Bearer token"
//	@Param			X-API-Key		header		string								true	"Project API key"
//	@Success		200				{array}		model.PartnerMonthlyDDCommission	"Monthly DD commission breakdown"
//	@Failure		401				{object}	apperrors.ErrorPublic				"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic				"Partner permission required"
//	@Router			/me/analytics/dd-commission-monthly [get]
func (h *CommissionHandler) GetMonthlyDDCommission(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	result, err := h.CommissionService.GetMonthlyDDCommission(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.JSON(result)
}

// GetProjectTransactions godoc
//
//	@Summary		Get project transactions
//	@Description	Returns on-chain transactions associated with the partner's project (predictions, rewards, refunds, commissions).
//	@Tags			partner
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Param			X-API-Key		header		string					true	"Project API key"
//	@Success		200				{array}		model.DuelTransaction	"Transaction list"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		403				{object}	apperrors.ErrorPublic	"Partner permission required"
//	@Router			/me/transactions [get]
func (h *CommissionHandler) GetProjectTransactions(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	txs, err := h.CommissionService.GetProjectTransactions(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.JSON(txs)
}

// parseDateRange extracts optional from/to query params (YYYY-MM-DD).
// Defaults to current month if omitted.
func parseDateRange(c fiber.Ctx) (from, to time.Time, err error) {
	now := time.Now().UTC()
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" {
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	} else {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return time.Time{}, time.Time{}, apperrors.BadRequest("invalid 'from' date, expected YYYY-MM-DD")
		}
	}

	if toStr == "" {
		to = now.AddDate(0, 0, 1)
	} else {
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return time.Time{}, time.Time{}, apperrors.BadRequest("invalid 'to' date, expected YYYY-MM-DD")
		}
		to = to.AddDate(0, 0, 1)
	}

	return from, to, nil
}
