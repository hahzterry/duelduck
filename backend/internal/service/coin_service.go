package service

import (
	"context"
	"dd-prediction-api/config"
	"dd-prediction-api/internal/client/jupiter"
	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/storage/click"
	"dd-prediction-api/internal/storage/repository"
	"dd-prediction-api/pkg/apperrors"
	repo "dd-prediction-api/pkg/repository"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/goccy/go-json"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type ICoinService interface {
	findSolanaTokenBySymbol(ctx context.Context, symbol string) (*model.SolanaToken, error)
}

type CoinService struct {
	TokenPriceService    *TokenPriceService
	Client               *http.Client
	CoinRepository       *repository.CoinRepository
	TransactionManager   *repo.TransactionManager
	TokenPriceRepository *click.TokenPriceRepository
	JupClient            *jupiter.Client
	CMCApiKey            string
}

func NewCoinService(
	c *config.Config,
	tokenPriceService *TokenPriceService,
	coinRepository *repository.CoinRepository,
	transactionManager *repo.TransactionManager,
	tokenPriceRepository *click.TokenPriceRepository,
	jupClient *jupiter.Client,
) *CoinService {
	return &CoinService{
		TokenPriceService:    tokenPriceService,
		Client:               http.DefaultClient,
		CoinRepository:       coinRepository,
		TransactionManager:   transactionManager,
		TokenPriceRepository: tokenPriceRepository,
		JupClient:            jupClient,
		CMCApiKey:            c.App.CMCApiKey,
	}
}

func (s *CoinService) RefreshSolanaTokens(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	jupTokens, err := s.JupClient.TokensV2Verified(ctx)
	if err != nil {
		return apperrors.Internal("failed to fetch verified tokens from jupiter", err)
	}

	tokens := make([]model.SolanaToken, 0, len(jupTokens))
	for _, t := range jupTokens {
		if t.ID == "" || t.Mcap < model.MinMarketCap {
			continue
		}

		tokens = append(tokens, model.SolanaToken{
			Mint:      t.ID,
			Name:      t.Name,
			Symbol:    t.Symbol,
			Decimals:  uint8(t.Decimals),
			USDPrice:  t.UsdPrice,
			MarketCap: t.Mcap,
			ImageURL:  t.Icon,
			ProgramID: t.TokenProgram,
		})
	}

	// refresh and clear solana tokens
	err = s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {

			if err := s.CoinRepository.RefreshSolanaTokens(ctx, tokens...); err != nil {
				return apperrors.Internal("failed to refresh solana tokens", err)
			}

			if err = s.CoinRepository.ClearInvalidSolanaTokens(ctx); err != nil {
				return apperrors.Internal("failed to clear invalid solana tokens")
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *CoinService) FindSolanaTokensBySymbolAndName(
	ctx context.Context,
	search string,
	limit int,
) ([]model.SearchSolanaTokens, error) {
	if search == "" {
		return nil, nil
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	tokens, err := s.CoinRepository.FindSolanaTokensBySymbolAndName(ctx, search, limit)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *CoinService) GetSolanaTokenPrice(
	ctx context.Context,
	mint string,
) (float64, error) {
	// fetch token from CH
	tokenPrice, err := s.TokenPriceRepository.GetLastPrice(ctx, mint)
	if err != nil {
		return 0, err
	}

	now := time.Now().UTC()

	// if token found and last token fetch time is relevant - return
	if tokenPrice != nil && now.Before(tokenPrice.TS.Add(5*time.Minute)) {
		return tokenPrice.UsdPrice, nil
	}

	// if no token found or last token fetch time is outdated - fetch token info from postgres
	token, err := s.CoinRepository.FindSolanaTokenByMint(ctx, mint)
	if err != nil {
		return 0, apperrors.Internal("failed to get token by mint", err)
	}

	if token == nil {
		return 0, apperrors.NotFound("token with provided mint not found")
	}

	// force update token info & force add into CH
	if now.Unix() >= token.UpdatedAt.UTC().Add(5*time.Minute).Unix() {
		resp, err := s.JupClient.Price(ctx, token.Mint)
		if err != nil {
			return 0, err
		}

		tokenInfo, ok := resp[token.Mint]
		if !ok {
			return 0, apperrors.ServiceUnavailable("failed to get token info from jupiter")
		}

		token.USDPrice = tokenInfo.UsdPrice

		err = s.RefreshSolanaToken(ctx, token)
		if err != nil {
			zap.L().Error("failed to refresh solana token", zap.Error(err))
		}

		err = s.TokenPriceService.AddNewPricesForced(ctx, token.Mint)
		if err != nil {
			zap.L().Error("failed to add new price into ch", zap.Error(err))
		}
	}

	return token.USDPrice, nil
}

func (s *CoinService) RefreshSolanaToken(ctx context.Context, solanaToken *model.SolanaToken) error {
	if err := s.CoinRepository.RefreshSolanaTokens(ctx, *solanaToken); err != nil {
		return apperrors.Internal("failed to refresh solana tokens", err)
	}

	return nil
}

func (s *CoinService) GetSolanaTokenByName(ctx context.Context, token string) ([]model.SolanaToken, error) {
	tokenInfo, err := s.CoinRepository.GetSolanaTokenByName(ctx, token)
	if err != nil {
		return nil, apperrors.Internal("failed to find any token by name or symbol", err)
	}

	return tokenInfo, nil
}

func (s *CoinService) findSolanaTokenBySymbol(ctx context.Context, symbol string) (*model.SolanaToken, error) {
	solanaToken, err := s.CoinRepository.FindSolanaTokenBySymbol(ctx, symbol)
	if err != nil {
		return nil, apperrors.Internal("failed to find any token by name or symbol", err)
	}

	return solanaToken, nil
}

func (s *CoinService) GetSolanaTokenByMint(ctx context.Context, mint string) (*model.SolanaToken, error) {
	if _, err := solana.PublicKeyFromBase58(mint); err != nil {
		return nil, apperrors.BadRequest("failed to parse mint address", err)
	}

	token, err := s.CoinRepository.FindSolanaTokenByMint(ctx, mint)
	if err != nil {
		return nil, apperrors.Internal("failed to get token by mint", err)
	}

	if token == nil {
		return nil, apperrors.NotFound("token with provided mint not found")
	}

	return token, nil
}

func (s *CoinService) GetOpenHighLowCloseVolume(
	params *model.CMCOHLCVParams,
	ids ...model.CMCID,
) (map[model.CMCID]model.CMCOHLCVData, error) {
	const apiURL = "https://pro-api.coinmarketcap.com/v2/cryptocurrency/ohlcv/historical"
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("Failed to create request: %v", err)
	}

	queryIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		queryIDs = append(queryIDs, strconv.FormatUint(uint64(id), 10))
	}

	q := req.URL.Query()
	q.Add("id", strings.Join(queryIDs, ","))
	q.Add("time_start", params.TimeStart.Format(time.RFC3339))
	q.Add("time_end", params.TimeEnd.Format(time.RFC3339))
	q.Add("time_period", params.TimePeriod)
	q.Add("interval", params.Interval)

	req.URL.RawQuery = q.Encode()
	req.Header.Add("X-CMC_PRO_API_KEY", s.CMCApiKey)
	req.Header.Add("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("Error making the API request: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		status := strconv.FormatInt(int64(resp.StatusCode), 10)
		return nil, apperrors.ServiceUnavailable("Non-OK HTTP status from CoinMarketCap "+apiURL+": "+status, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.Internal("failed to read response", err)
	}

	var apiResponse model.CMCOHLCVMultipleTokensResp
	if err = json.Unmarshal(body, &apiResponse); err != nil {

		var apiResponse model.CMCOHLCVSingleTokenResp
		if err = json.Unmarshal(body, &apiResponse); err != nil {
			return nil, apperrors.Internal("failed to unmarshal response", err)
		}

		return map[model.CMCID]model.CMCOHLCVData{
			apiResponse.Data.ID: apiResponse.Data,
		}, nil
	}

	return apiResponse.Data, nil
}

func (s *CoinService) GetHighLowHourly(ids ...model.CMCID) (map[model.CMCID]model.CMCHighLow, error) {
	now := time.Now().UTC().Truncate(time.Hour)

	params := &model.CMCOHLCVParams{
		// if current was 17:30, TimeStart would be equal to 15:59. We want to get data for the latest hour
		// and according to CoinMarketCap doc we must pass time_start exclusively
		// https://coinmarketcap.com/api/documentation/v1/#operation/getV2CryptocurrencyOhlcvLatest
		TimeStart:  now.Add(-time.Hour - time.Minute),
		TimeEnd:    now,
		TimePeriod: "hourly",
		Interval:   "hourly",
	}

	data, err := s.GetOpenHighLowCloseVolume(params, ids...)
	if err != nil {
		return nil, err
	}

	highLow := make(map[model.CMCID]model.CMCHighLow, len(data))
	for _, ohlcv := range data {
		if len(ohlcv.Quotes) == 0 {
			zap.L().Error("no data provided for coin in CMC OHLCV resp")
			continue
		}

		quote := ohlcv.Quotes[len(ohlcv.Quotes)-1]
		hl := model.CMCHighLow{
			High: quote.Quote.USD.High,
			Low:  quote.Quote.USD.Low,
		}

		highLow[ohlcv.ID] = hl
	}

	return highLow, nil
}
