package jupiter

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"dd-prediction-api/config"
	dbrepo "dd-prediction-api/internal/storage/repository"
	"dd-prediction-api/pkg/apperrors"

	"github.com/gagliardetto/solana-go"
	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
)

type Client struct {
	client       *resty.Client
	apiErrorRepo *dbrepo.APIErrorRepository
}

type Option func(*Client)

func NewClient(c *config.Config, apiErrorRepo *dbrepo.APIErrorRepository, options ...Option) *Client {
	r := resty.New().
		SetBaseURL(c.Client.JupiterBaseURL).
		SetHeader("Content-Type", "application/json")

	if c.Client.JupiterAPIKey != "" {
		r.SetHeader("x-api-key", c.Client.JupiterAPIKey)
	}

	client := &Client{
		client:       r,
		apiErrorRepo: apiErrorRepo,
	}

	for _, option := range options {
		option(client)
	}

	return client
}

type QuoteMetadata json.RawMessage

type GetQuoteParams struct {
	InputMintAddress  string
	OutputMintAddress string

	Amount         int64
	SlippageBps    int64
	PlatformFeeBps int64
}

func (c Client) GetQuote(ctx context.Context, params GetQuoteParams) (*QuoteMetadata, error) {
	const url = "swap/v1/quote"

	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("inputMint", params.InputMintAddress).
		SetQueryParam("outputMint", params.OutputMintAddress).
		SetQueryParam("amount", strconv.FormatInt(params.Amount, 10)).
		SetQueryParam("slippageBps", strconv.FormatInt(params.SlippageBps, 10)).
		SetQueryParam("maxAccounts", strconv.FormatInt(28, 10)).
		SetQueryParam("platformFeeBps", strconv.FormatInt(params.PlatformFeeBps, 10)).
		Get(url)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to get resp from jupiter: "+url, err)
	}

	if !resp.IsSuccess() {
		if c.apiErrorRepo != nil {
			c.apiErrorRepo.LogAsync("jupiter", url, resp.StatusCode(), resp.Body())
		}
		if resp.StatusCode() == http.StatusBadRequest {
			return nil, apperrors.BadRequest("bad request from jupiter: " + resp.Request.URL)
		}
		if resp.StatusCode() == http.StatusUnauthorized || resp.StatusCode() == http.StatusForbidden {
			return nil, apperrors.ServiceUnavailable("jupiter auth failed (x-api-key missing/invalid)")
		}
		return nil, apperrors.ServiceUnavailable("failed to get resp from jupiter: status is not success")
	}

	metadata := QuoteMetadata(resp.Body())

	return &metadata, nil
}

type GetSwapTransactionRequest struct {
	QuoteMetadata json.RawMessage   `json:"quoteResponse"`
	UserPublicKey solana.PublicKey  `json:"userPublicKey"`
	FeeAccount    *solana.PublicKey `json:"feeAccount,omitempty"`
}

type GetSwapTransactionResult struct {
	SwapTransaction *solana.Transaction
}
type GetSwapTransactionResponse struct {
	SwapTransaction string `json:"swapTransaction"`
}

func (c Client) GetSwapTransaction(
	ctx context.Context,
	metadata QuoteMetadata,
	wallet solana.PublicKey,
	adminPublicKey *solana.PublicKey,
) (*GetSwapTransactionResult, error) {
	const url = "swap/v1/swap"

	data, err := json.Marshal(GetSwapTransactionRequest{
		QuoteMetadata: json.RawMessage(metadata),
		UserPublicKey: wallet,
		FeeAccount:    adminPublicKey,
	})
	if err != nil {
		return nil, apperrors.Internal("failed to marshal swap transaction", err)
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(data).
		SetResult(&GetSwapTransactionResponse{}).
		Post(url)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to get swap tx from jupiter", err)
	}

	if !resp.IsSuccess() {
		if c.apiErrorRepo != nil {
			c.apiErrorRepo.LogAsync("jupiter", url, resp.StatusCode(), resp.Body())
		}
		if resp.StatusCode() == http.StatusBadRequest {
			return nil, apperrors.BadRequest("received bad request from jupiter "+url, err)
		}
		if resp.StatusCode() == http.StatusUnauthorized || resp.StatusCode() == http.StatusForbidden {
			return nil, apperrors.ServiceUnavailable("jupiter auth failed (x-api-key missing/invalid)")
		}
		return nil, apperrors.ServiceUnavailable("failed to get swap tx: non success status received", err)
	}

	res := resp.Result()
	tx, ok := res.(*GetSwapTransactionResponse)
	if !ok {
		return nil, apperrors.ServiceUnavailable("invalid data received from jupiter; cast failed")
	}

	parsedTx, err := solana.TransactionFromBase64(tx.SwapTransaction)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("decode swap transaction", err)
	}

	return &GetSwapTransactionResult{SwapTransaction: parsedTx}, nil
}

type GetPriceResp struct {
	UsdPrice       float64 `json:"usdPrice"`
	BlockId        int64   `json:"blockId"`
	Decimals       uint8   `json:"decimals"`
	PriceChange24h float64 `json:"priceChange24h"`
}

func (c Client) Price(
	ctx context.Context,
	tokens ...string,
) (map[string]GetPriceResp, error) {
	const url = "price/v3"

	result := make(map[string]GetPriceResp)
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("ids", strings.Join(tokens, ",")).
		SetResult(&result).
		Get(url)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to get token prices from jupiter", err)
	}

	if !resp.IsSuccess() {
		if c.apiErrorRepo != nil {
			c.apiErrorRepo.LogAsync("jupiter", url, resp.StatusCode(), resp.Body())
		}
		if resp.StatusCode() == http.StatusBadRequest {
			return nil, apperrors.BadRequest("received bad request from jupiter "+url, err)
		}
		if resp.StatusCode() == http.StatusUnauthorized || resp.StatusCode() == http.StatusForbidden {
			return nil, apperrors.ServiceUnavailable("jupiter auth failed (x-api-key missing/invalid)")
		}
		return nil, apperrors.ServiceUnavailable("failed to get token prices: non success status received", err)
	}

	return result, nil
}

type TokenV2 struct {
	ID           string   `json:"id"` // mint
	Name         string   `json:"name"`
	Symbol       string   `json:"symbol"`
	Icon         string   `json:"icon"`
	Decimals     int      `json:"decimals"`
	TokenProgram string   `json:"tokenProgram"`
	Tags         []string `json:"tags"`
	IsVerified   bool     `json:"isVerified"`

	UsdPrice     float64 `json:"usdPrice"`
	PriceBlockId int64   `json:"priceBlockId"`
	Liquidity    float64 `json:"liquidity"`
	Mcap         float64 `json:"mcap"`
	Fdv          float64 `json:"fdv"`

	Stats24h TokenV2Stats24h `json:"stats24h"`

	UpdatedAt string `json:"updatedAt"`
}

type TokenV2Stats24h struct {
	PriceChange     float64 `json:"priceChange"`
	HolderChange    float64 `json:"holderChange"`
	LiquidityChange float64 `json:"liquidityChange"`
	VolumeChange    float64 `json:"volumeChange"`
	BuyVolume       float64 `json:"buyVolume"`
	SellVolume      float64 `json:"sellVolume"`
	NumBuys         int64   `json:"numBuys"`
	NumSells        int64   `json:"numSells"`
	NumTraders      int64   `json:"numTraders"`
}

func (c Client) TokensV2Verified(ctx context.Context) ([]TokenV2, error) {
	const url = "tokens/v2/tag"

	var result []TokenV2
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("query", "verified").
		SetResult(&result).
		Get(url)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to get tokens v2 from jupiter", err)
	}

	if !resp.IsSuccess() {
		if c.apiErrorRepo != nil {
			c.apiErrorRepo.LogAsync("jupiter", url, resp.StatusCode(), resp.Body())
		}
		if resp.StatusCode() == http.StatusUnauthorized || resp.StatusCode() == http.StatusForbidden {
			return nil, apperrors.ServiceUnavailable("jupiter auth failed (x-api-key missing/invalid)")
		}
		return nil, apperrors.ServiceUnavailable("failed to get tokens v2: non success status received")
	}

	return result, nil
}

func (c Client) SearchV2(
	ctx context.Context,
	mints ...string,
) ([]TokenV2, error) {
	const url = "tokens/v2/search"

	var result []TokenV2
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("query", strings.Join(mints, ",")).
		SetResult(&result).
		Get(url)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to search tokens v2 from jupiter", err)
	}

	if !resp.IsSuccess() {
		if c.apiErrorRepo != nil {
			c.apiErrorRepo.LogAsync("jupiter", url, resp.StatusCode(), resp.Body())
		}
		if resp.StatusCode() == http.StatusUnauthorized || resp.StatusCode() == http.StatusForbidden {
			return nil, apperrors.ServiceUnavailable("jupiter auth failed (x-api-key missing/invalid)")
		}
		return nil, apperrors.ServiceUnavailable("failed to search tokens v2: non success status received")
	}

	return result, nil
}
