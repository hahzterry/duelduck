package v1

import (
	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/service"
	"dd-prediction-api/pkg/apperrors"

	"github.com/gofiber/fiber/v3"
)

type CoinHandler struct {
	CoinService *service.CoinService
}

func NewCoinHandler(coinService *service.CoinService) *CoinHandler {
	return &CoinHandler{CoinService: coinService}
}

func (h *CoinHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	coinGroup := app.Group("/coins") // todo: implement rate limiter

	{
		coinGroup.Get("/solana-tokens", h.SearchSolanaTokens)
		coinGroup.Get("/solana-tokens/price", h.GetSolanaTokenPrice)

		coinGroup.Get("/mint/:mint", h.GetTokenByMint)
		coinGroup.Get("/name/:name", h.GetTokenByName)
	}
}

// GetTokenByName godoc
//
//	@Summary		Get Solana token by name
//	@Description	Retrieve Solana token information by searching for tokens with the specified name
//	@Tags			coins
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string					true	"Token name to search for"
//	@Success		200		{array}		model.SolanaToken		"List of tokens matching the name"
//	@Failure		400		{object}	apperrors.ErrorPublic	"Invalid request data or empty token name"
//	@Failure		500		{object}	apperrors.ErrorPublic	"Internal server error during token retrieval"
//	@Router			/coins/name/{name} [get]
func (h *CoinHandler) GetTokenByName(c fiber.Ctx) error {
	token := c.Params("name")

	if token == "" {
		return apperrors.BadRequest("non-empty token query parameter expected")
	}

	tokens, err := h.CoinService.GetSolanaTokenByName(c.Context(), token)
	if err != nil {
		return err
	}

	return c.JSON(tokens)
}

// GetTokenByMint godoc
//
//	@Summary		Get Solana token by mint address
//	@Description	Retrieve Solana token information by searching for tokens with the specified mint address
//	@Tags			coins
//	@Accept			json
//	@Produce		json
//	@Param			mint	path		string					true	"Mint address to search for"
//	@Success		200		{object}	model.SolanaToken		"Token information retrieved successfully"
//	@Failure		400		{object}	apperrors.ErrorPublic	"Invalid request data or invalid mint address format"
//	@Failure		404		{object}	apperrors.ErrorPublic	"Token with provided mint address not found"
//	@Failure		500		{object}	apperrors.ErrorPublic	"Internal server error during token retrieval"
//	@Router			/coins/mint/{mint} [get]
func (h *CoinHandler) GetTokenByMint(c fiber.Ctx) error {
	mint := c.Params("mint")

	tokens, err := h.CoinService.GetSolanaTokenByMint(c.Context(), mint)
	if err != nil {
		return err
	}

	return c.JSON(tokens)
}

// SearchSolanaTokens godoc
//
//	@Summary		Search Solana tokens
//	@Description	Search Solana tokens by symbol or name (query params). Returns list of matching tokens.
//	@Tags			coins
//	@Accept			json
//	@Produce		json
//	@Param			search	query		string					true	"Search string (symbol or name)"
//	@Param			limit	query		int						false	"Max results to return"	minimum(1)	maximum(100)	default(20)
//	@Success		200		{array}		model.SolanaToken		"List of matching tokens"
//	@Failure		400		{object}	apperrors.ErrorPublic	"Invalid request params"
//	@Failure		500		{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/coins/solana-tokens [get]
func (h *CoinHandler) SearchSolanaTokens(c fiber.Ctx) error {
	var req model.SearchSolanaTokensReq
	if err := c.Bind().Query(&req); err != nil {
		return apperrors.BadRequest("invalid request params")
	}

	tokens, err := h.CoinService.FindSolanaTokensBySymbolAndName(c.Context(), req.Search, req.Limit)
	if err != nil {
		return err
	}

	return c.JSON(tokens)
}

// GetSolanaTokenPrice godoc
//
//	@Summary		Get Solana token price
//	@Description	Retrieve current token price by mint address (query param).
//	@Tags			coins
//	@Accept			json
//	@Produce		json
//	@Param			mint	query		string					true	"Token mint address"
//	@Success		200		{object}	object{price=float64}	"Token price retrieved successfully"
//	@Failure		400		{object}	apperrors.ErrorPublic	"Invalid request params"
//	@Failure		404		{object}	apperrors.ErrorPublic	"Token not found"
//	@Failure		500		{object}	apperrors.ErrorPublic	"Internal server error"
//	@Failure		503		{object}	apperrors.ErrorPublic	"External service unavailable"
//	@Router			/coins/solana-tokens/price [get]
func (h *CoinHandler) GetSolanaTokenPrice(c fiber.Ctx) error {
	var req model.SearchSolanaTokenPriceReq
	if err := c.Bind().Query(&req); err != nil {
		return apperrors.BadRequest("invalid request params")
	}

	tokenPrice, err := h.CoinService.GetSolanaTokenPrice(c.Context(), req.Mint)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"price": tokenPrice})
}
