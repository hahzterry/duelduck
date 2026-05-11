package v1

import (
	"dd-prediction-api/internal/handler/middleware"
	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/service"
	"dd-prediction-api/pkg/apperrors"
	auth "dd-prediction-api/pkg/jwt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type DuelHandler struct {
	DuelQueue   *service.DuelQueue
	DuelService *service.DuelService
	FileService *service.FileService
}

func NewDuelHandler(
	duelQueue *service.DuelQueue,
	duelService *service.DuelService,
	fileService *service.FileService,

) (*DuelHandler, error) {
	authHandler := &DuelHandler{
		DuelQueue:   duelQueue,
		DuelService: duelService,
		FileService: fileService,
	}

	return authHandler, nil
}

func (h *DuelHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	duelGroup := app.Group("/duel", auth.ApiKeyMiddleware, auth.AuthMiddleware, middleware.PageViewMiddleware)
	{
		duelGroup.Get("/all", h.GetAllDuels)
		duelGroup.Get("/my", h.GetMyDuels)
		duelGroup.Get("/all-with-joined", h.GetAllDuelsAuthorized)
		duelGroup.Get("/participating", h.GetAllDuelsWhereParticipant)
		duelGroup.Get("/history", h.GetMyHistoryDuels)

		duelGroup.Get("/token-accounts", h.GetTokenAccountBalances)

		duelGroup.Post("/upload-image", h.UploadDuelLogo, auth.UploadUserImageMiddleware)

		duelGroup.Put("/self-resolve", h.ResolveCryptoDuelByOwner)

		solanaDuelGroup := duelGroup.Group("/solana", auth.RequireUserRole)
		{
			solanaDuelGroup.Post("/", h.CreateCryptoDuel)
			solanaDuelGroup.Post("/sign-tx", h.SignCreateCryptoDuelTransaction)
			solanaDuelGroup.Post("/join", h.JoinCryptoDuel)
			solanaDuelGroup.Post("/join/sign-tx", h.SignJoinCryptoDuelTransaction)
		}
	}
}

// SignCreateCryptoDuelTransaction godoc
//
//	@Summary		Sign transaction to create app crypto duel
//	@Description	Creates and signs a transaction for initializing an application-based crypto duel. Returns a signed transaction that can be submitted to the blockchain.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Param			request			body		model.CreateDuelReq		true	"Request data for creating a duel"
//	@Success		200				{object}	object{tx=string}		"Signed transaction returned successfully"
//	@Failure		400				{object}	apperrors.ErrorPublic	"Invalid request data"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized - Invalid or missing claims"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/crypto-duel/external-wallet/solana/sign-tx [post]
func (h *DuelHandler) SignCreateCryptoDuelTransaction(c fiber.Ctx) error {
	var req model.CreateDuelReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	tx, err := h.DuelService.SignCreateCryptoDuelTransaction(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"tx": tx})
}

// CreateAppCryptoDuel godoc
//
//	@Summary		Create app-based crypto duel
//	@Description	Creates an application-based crypto duel using the provided data. Requires authentication.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string						true	"Authorization Bearer token"
//	@Success		200				{object}	model.CreateCryptoDuelResp	"Duel created successfully"
//	@Failure		400				{object}	apperrors.ErrorPublic		"Invalid request data"
//	@Failure		401				{object}	apperrors.ErrorPublic		"Unauthorized - Invalid or missing claims"
//	@Failure		500				{object}	apperrors.ErrorPublic		"Internal server error"
//	@Router			/crypto-duel/external-wallet/solana [post]
func (h *DuelHandler) CreateCryptoDuel(c fiber.Ctx) error {
	var req model.CreateCryptoDuelReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	resp, err := h.DuelService.CreateCryptoDuel(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// SignJoinCryptoDuelTransaction godoc
//
//	@Summary		Sign transaction to join an app crypto duel
//	@Description	Creates a signed transaction that allows a user to join an app-based crypto duel.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Param			request			body		model.JoinDuelReq		true	"Duel join request"
//	@Success		200				{object}	object{tx=string}		"Signed transaction returned successfully"
//	@Failure		400				{object}	apperrors.ErrorPublic	"Invalid request data"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized - Invalid or missing claims"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/crypto-duel/external-wallet/solana/join/sign-tx [post]
func (h *DuelHandler) SignJoinCryptoDuelTransaction(c fiber.Ctx) error {
	var req model.JoinDuelReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	resp, err := h.DuelQueue.SignJoinCryptoDuelTransaction(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// JoinExternalWalletCryptoDuel godoc
//
//	@Summary		Join an app-based crypto duel
//	@Description	Allows an authenticated user to join an existing app-based crypto duel.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string						true	"Authorization Bearer token"
//	@Success		200				{object}	model.JoinCryptoDuelResp	"Joined the duel successfully"
//	@Failure		400				{object}	apperrors.ErrorPublic		"Invalid request data"
//	@Failure		401				{object}	apperrors.ErrorPublic		"Unauthorized - Invalid or missing claims"
//	@Failure		500				{object}	apperrors.ErrorPublic		"Internal server error"
//	@Router			/crypto-duel/external-wallet/solana/join [post]
func (h *DuelHandler) JoinCryptoDuel(c fiber.Ctx) error {
	var req model.JoinCryptoDuelReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	resp, err := h.DuelQueue.JoinCryptoDuel(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ResolveCryptoDuelByOwner godoc
//
//	@Summary		Resolve an owner-resolving duel  crypto duel
//	@Description	Allows the duel owner to resolve an owner-resolving crypto duel. Returns resulting transaction hashes.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string						true	"Authorization Bearer token"
//	@Param			request			body		model.DuelResolveReq		true	"Duel resolve request"
//	@Success		200				{object}	model.ResolveCryptoDuelResp	"Transaction hashes and duel"
//	@Failure		400				{object}	apperrors.ErrorPublic		"Invalid request data"
//	@Failure		401				{object}	apperrors.ErrorPublic		"Unauthorized - Invalid or missing claims"
//	@Failure		500				{object}	apperrors.ErrorPublic		"Internal server error"
//	@Router			/crypto-duel/self-resolve [post]
func (h *DuelHandler) ResolveCryptoDuelByOwner(c fiber.Ctx) error {
	var req model.DuelResolveReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	resp, err := h.DuelQueue.ResolveCryptoDuelByOwner(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetAllDuels godoc
//
//	@Summary		Get all duels
//	@Description	Get all duels with optional filtering, sorting and pagination
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Param			opts.pagination.page_size	query		uint64					false	"Number of items per page"		default(10)
//	@Param			opts.pagination.page_num	query		uint64					false	"Page number (starting from 1)"	default(1)
//	@Param			opts.order.order_by			query		string					false	"Field to order by"				default(created_at)
//	@Param			opts.order.order_type		query		string					false	"Order type"					Enums(desc,asc)	default("")	"Order type (asc or desc)"	Enums(asc,desc)	default(desc)
//	@Param			opts.filters[0].column		query		string					false	"First filter column name"
//	@Param			opts.filters[0].operator	query		string					false	"First filter operator"
//	@Param			opts.filters[0].value		query		string					false	"First filter value"
//	@Param			opts.filters[0].where_or	query		bool					false	"First filter OR condition"
//	@Success		200							{array}		model.DuelShow			"List of duels"
//	@Failure		400							{object}	apperrors.ErrorPublic	"Bad request"
//	@Failure		500							{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/duel/all [get]
func (h *DuelHandler) GetAllDuels(c fiber.Ctx) error {
	var req model.OptsReq
	if err := c.Bind().Query(&req); err != nil {
		return apperrors.BadRequest("invalid request params")
	}

	duels, err := h.DuelService.GetAllDuelsUnauthorized(c.Context(), &req.Opts)
	if err != nil {
		return err
	}

	return c.JSON(duels)
}

// GetAllDuelsWhereParticipant godoc
//
//	@Summary		Get all duels where the user is a participant
//	@Description	Get all duels where the authenticated user is a participant, with optional filtering, sorting and pagination.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization				header		string					true	"Authorization Bearer token"
//	@Param			opts.pagination.page_size	query		uint64					false	"Number of items per page"		default(10)
//	@Param			opts.pagination.page_num	query		uint64					false	"Page number (starting from 1)"	default(1)
//	@Param			opts.order.order_by			query		string					false	"Field to order by"				default(created_at)
//	@Param			opts.order.order_type		query		string					false	"Order type"					Enums(desc,asc)	default("")	"Order type (asc or desc)"	Enums(asc,desc)	default(desc)
//	@Param			opts.filters[0].column		query		string					false	"First filter column name"
//	@Param			opts.filters[0].operator	query		string					false	"First filter operator"
//	@Param			opts.filters[0].value		query		string					false	"First filter value"
//	@Param			opts.filters[0].where_or	query		bool					false	"First filter OR condition"
//	@Success		200							{array}		model.DuelShow			"List of duels"
//	@Failure		400							{object}	apperrors.ErrorPublic	"Bad request"
//	@Failure		401							{object}	apperrors.ErrorPublic	"Unauthorized - Invalid or missing claims"
//	@Failure		500							{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/duel/participating [get]
func (h *DuelHandler) GetAllDuelsWhereParticipant(c fiber.Ctx) error {
	var req model.OptsReq
	if err := c.Bind().Query(&req); err != nil {
		return apperrors.BadRequest("invalid request params")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	duels, err := h.DuelService.GetAllDuelsWhereParticipant(c.Context(), claims.UserID, &req.Opts)
	if err != nil {
		return err
	}

	return c.JSON(duels)
}

// GetAllDuelsAuthorized godoc
//
//	@Summary		Get all duels with user's participation status
//	@Description	Retrieves all duels and for each duel, indicates if the authenticated user is a participant. Supports optional filtering, sorting, and pagination.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization				header		string					true	"Authorization Bearer token"
//	@Param			opts.pagination.page_size	query		uint64					false	"Number of items per page"		default(10)
//	@Param			opts.pagination.page_num	query		uint64					false	"Page number (starting from 1)"	default(1)
//	@Param			opts.order.order_by			query		string					false	"Field to order by"
//	@Param			opts.order.order_type		query		string					false	"Order type"	Enums(desc,asc)	default("")	"Order type (asc or desc)"
//	@Param			opts.filters[0].column		query		string					false	"First filter column name"
//	@Param			opts.filters[0].operator	query		string					false	"First filter operator"
//	@Param			opts.filters[0].value		query		string					false	"First filter value"
//	@Param			opts.filters[0].where_or	query		bool					false	"First filter OR condition"
//	@Success		200							{array}		model.DuelShow			"List of duels with participation status"
//	@Failure		400							{object}	apperrors.ErrorPublic	"Bad request"
//	@Failure		401							{object}	apperrors.ErrorPublic	"Unauthorized - Invalid or missing claims"
//	@Failure		500							{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/duel/all-with-joined [get]
func (h *DuelHandler) GetAllDuelsAuthorized(c fiber.Ctx) error {
	var req model.OptsReq
	if err := c.Bind().Query(&req); err != nil {
		return apperrors.BadRequest("invalid request params")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	duels, err := h.DuelService.GetAllDuels(c.Context(), claims.UserID, &req.Opts)
	if err != nil {
		return err
	}

	return c.JSON(duels)
}

// GetMyHistoryDuels godoc
//
//	@Summary		Get user's duel history
//	@Description	Retrieves the authenticated user's duel history with optional filtering, sorting, and pagination.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization				header		string					true	"Authorization Bearer token"
//	@Param			opts.pagination.page_size	query		uint64					false	"Number of items per page"		default(10)
//	@Param			opts.pagination.page_num	query		uint64					false	"Page number (starting from 1)"	default(1)
//	@Param			opts.order.order_by			query		string					false	"Field to order by"
//	@Param			opts.order.order_type		query		string					false	"Order type"	Enums(desc,asc)	default("")	"Order type (asc or desc)"
//	@Param			opts.filters[0].column		query		string					false	"Filter column name"
//	@Param			opts.filters[0].operator	query		string					false	"Filter operator"
//	@Param			opts.filters[0].value		query		string					false	"Filter value"
//	@Param			opts.filters[0].where_or	query		bool					false	"Filter OR condition"
//	@Success		200							{array}		model.DuelShow			"List of user's historical duels"
//	@Failure		400							{object}	apperrors.ErrorPublic	"Invalid request parameters"
//	@Failure		401							{object}	apperrors.ErrorPublic	"Authentication required or invalid token"
//	@Failure		500							{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/duel/history [get]
func (h *DuelHandler) GetMyHistoryDuels(c fiber.Ctx) error {
	var req model.OptsReq
	if err := c.Bind().Query(&req); err != nil {
		return apperrors.BadRequest("invalid request params")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	duels, err := h.DuelService.GetMyHistoryDuels(c.Context(), claims.UserID, &req.Opts)
	if err != nil {
		return err
	}

	return c.JSON(duels)
}

// GetMyDuels godoc
//
//	@Summary		Get my duels
//	@Description	Retrieves a paginated list of duels created by or assigned to the authenticated user. Supports advanced filtering, ordering, and pagination via query parameters.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization				header		string					true	"Authorization Bearer token"
//	@Param			opts.pagination.page_size	query		uint64					false	"Number of items per page"		default(10)
//	@Param			opts.pagination.page_num	query		uint64					false	"Page number (starting from 1)"	default(1)
//	@Param			opts.order.order_by			query		string					false	"Field to order by"
//	@Param			opts.order.order_type		query		string					false	"Order type"	Enums(desc,asc)	default("")	"Order type (asc or desc)"
//	@Param			opts.filters[0].column		query		string					false	"Filter column name"
//	@Param			opts.filters[0].operator	query		string					false	"Filter operator"
//	@Param			opts.filters[0].value		query		string					false	"Filter value"
//	@Param			opts.filters[0].where_or	query		bool					false	"Filter OR condition"
//	@Success		200							{array}		model.DuelShow			"List of duels belonging to the authenticated user"
//	@Failure		400							{object}	apperrors.ErrorPublic	"Invalid request parameters"
//	@Failure		401							{object}	apperrors.ErrorPublic	"Authentication required or invalid token"
//	@Failure		500							{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/duel/my [get]
func (h *DuelHandler) GetMyDuels(c fiber.Ctx) error {
	var req model.OptsReq
	if err := c.Bind().Query(&req); err != nil {
		return apperrors.BadRequest("invalid request params")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	duel, err := h.DuelService.GetMyDuels(c.Context(), claims.UserID, &req.Opts)
	if err != nil {
		return err
	}

	return c.JSON(duel)
}

// GetDuelByIDUnauthorized godoc
//
//	@Summary		Get duel by ID (public)
//	@Description	Retrieve a duel by its ID without requiring authentication. Returns the duel details if found.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Duel ID"
//	@Success		200	{object}	model.Duel				"Duel details"
//	@Failure		400	{object}	apperrors.ErrorPublic	"Invalid duel ID"
//	@Failure		404	{object}	apperrors.ErrorPublic	"Duel not found"
//	@Failure		500	{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/duel/public/{id} [get]
func (h *DuelHandler) GetDuelByIDUnauthorized(c fiber.Ctx) error {
	duelID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid duel ID", err)
	}

	duel, err := h.DuelService.GetDuelByIDUnauthorized(c.Context(), duelID)
	if err != nil {
		return err
	}

	return c.JSON(duel)
}

// GetDuelByID godoc
//
//	@Summary		Get duel by ID (authorized)
//	@Description	Retrieve a duel by its ID for the authenticated user. Returns the duel details and the list of players. Requires authentication.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string													true	"Authorization Bearer token"
//	@Param			id				path		string													true	"Duel ID"
//	@Success		200				{object}	object{duel=model.DuelShow,players=[]model.PlayerShow}	"Duel details and players list"
//	@Failure		400				{object}	apperrors.ErrorPublic									"Invalid duel ID"
//	@Failure		401				{object}	apperrors.ErrorPublic									"Authentication required or invalid token"
//	@Failure		404				{object}	apperrors.ErrorPublic									"Duel not found"
//	@Failure		500				{object}	apperrors.ErrorPublic									"Internal server error"
//	@Router			/duel/{id} [get]
func (h *DuelHandler) GetDuelByID(c fiber.Ctx) error {
	duelID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid duel ID", err)
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	duel, players, err := h.DuelService.GetDuelByID(c.Context(), duelID, claims.UserID, claims.Role)
	if err != nil {
		return err
	}

	return c.JSON(map[string]any{
		"duel":    duel,
		"players": players,
	})
}

// GetTokenAccountBalances godoc
//
//	@Summary		Get token account balances for the user's public address
//	@Description	Retrieves the balances of all token accounts associated with the user's public blockchain address. Requires authentication. Supports filtering and pagination via query parameters.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string							true	"Authorization Bearer token"
//	@Param			type			query		string							false	"Type of token account to filter (e.g. SPL, NFT, etc.)"	default("token")
//	@Param			page_number		query		int								false	"Page number for pagination (starts from 1)"			default(1)
//	@Param			page_size		query		int								false	"Number of token accounts per page"						default(20)
//	@Param			hide_zero		query		bool							false	"Hide token accounts with zero balance"
//	@Success		200				{object}	model.SolscanTokenAccountsResp	"Token account balances retrieved successfully"
//	@Failure		400				{object}	apperrors.ErrorPublic			"Invalid request data or query parameters"
//	@Failure		401				{object}	apperrors.ErrorPublic			"Authentication required or invalid token"
//	@Failure		500				{object}	apperrors.ErrorPublic			"Internal server error"
//	@Failure		503				{object}	apperrors.ErrorPublic			"External service unavailable"
//	@Router			/crypto-duel/token-accounts [get]
func (h *DuelHandler) GetTokenAccountBalances(c fiber.Ctx) error {
	var req model.GetTokenAccountsReq
	if err := c.Bind().Query(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	resp, err := h.DuelService.GetTokenAccounts(c.Context(), &req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UploadDuelLogo godoc
//
//	@Summary		Upload a duel logo
//	@Description	Uploads and processes a duel logo image (JPEG, PNG, SVG, WebP; max 3 MB). Returns the URL of the saved logo.
//	@Tags			duel
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Param			duel_logo		formData	file					true	"Duel logo image"
//	@Success		200				{object}	map[string]string		"duel_logo_url"
//	@Failure		400				{object}	apperrors.ErrorPublic	"Invalid request data"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		429				{object}	apperrors.ErrorPublic	"Too many uploads"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/duel/upload-image [post]
func (h *DuelHandler) UploadDuelLogo(c fiber.Ctx) error {
	duelLogo, err := c.FormFile("duel_logo")
	if err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	duelLogoUrl, err := h.FileService.SaveDuelLogo(duelLogo)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"duel_logo_url": duelLogoUrl,
	})
}
