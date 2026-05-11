package service

import (
	"context"
	"dd-prediction-api/internal/client/jupiter"
	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/storage/click"
	"dd-prediction-api/internal/storage/repository"
	"dd-prediction-api/pkg/apperrors"
	"time"
)

type TokenPriceService struct {
	DuelRepository *repository.DuelRepository
	Repo           *click.TokenPriceRepository
	Jupiter        *jupiter.Client

	BaseTokens []string
}

func NewTokenPriceService(
	duelRepository *repository.DuelRepository,
	repo *click.TokenPriceRepository,
	jup *jupiter.Client,
) *TokenPriceService {
	baseTokens := []string{
		"So11111111111111111111111111111111111111112",  // SOL
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
	}
	return &TokenPriceService{
		DuelRepository: duelRepository,
		Repo:           repo,
		Jupiter:        jup,
		BaseTokens:     baseTokens,
	}
}

func (s *TokenPriceService) SyncPrices(ctx context.Context) error {
	activeMints, err := s.DuelRepository.GetActiveDuelTokenMints(ctx, s.BaseTokens)
	if err != nil {
		return apperrors.Internal("failed to get active duel token mints", err)
	}

	tokens := make([]string, 0, len(s.BaseTokens)+len(activeMints))

	tokens = append(tokens, s.BaseTokens...)
	tokens = append(tokens, activeMints...)

	var tokensInfo = make([]jupiter.TokenV2, 0, len(tokens))

	// cause Jupiter Search can handle only 100 items per once
	for i := 0; i < len(tokens); i += 100 {
		end := i + 100
		if end > len(tokens) {
			end = len(tokens)
		}

		part := tokens[i:end]

		partTokensInfo, err := s.Jupiter.SearchV2(ctx, part...)
		if err != nil {
			return err
		}

		tokensInfo = append(tokensInfo, partTokensInfo...)
	}

	now := time.Now().UTC()
	items := make([]model.TokenPrice, 0, len(tokensInfo))
	for _, i := range tokensInfo {
		items = append(items, model.TokenPrice{
			Token:          i.ID,
			Symbol:         i.Symbol,
			TS:             now,
			UsdPrice:       i.UsdPrice,
			BlockID:        i.PriceBlockId,
			Decimals:       uint8(i.Decimals),
			PriceChange24h: i.Stats24h.PriceChange,
		})
	}

	return s.Repo.BulkInsert(ctx, items)
}

func (s *TokenPriceService) AddNewPricesForced(ctx context.Context, mints ...string) error {
	var tokensInfo = make([]jupiter.TokenV2, 0, len(mints))

	// cause Jupiter Search can handle only 100 items per once
	for i := 0; i < len(mints); i += 100 {
		end := i + 100
		if end > len(mints) {
			end = len(mints)
		}

		part := mints[i:end]

		partTokensInfo, err := s.Jupiter.SearchV2(ctx, part...)
		if err != nil {
			return err
		}

		tokensInfo = append(tokensInfo, partTokensInfo...)
	}

	now := time.Now().UTC()
	items := make([]model.TokenPrice, 0, len(tokensInfo))
	for _, i := range tokensInfo {
		items = append(items, model.TokenPrice{
			Token:          i.ID,
			Symbol:         i.Symbol,
			TS:             now,
			UsdPrice:       i.UsdPrice,
			BlockID:        i.PriceBlockId,
			Decimals:       uint8(i.Decimals),
			PriceChange24h: i.Stats24h.PriceChange,
		})
	}

	return s.Repo.BulkInsert(ctx, items)
}
