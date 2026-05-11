package repository

import (
	"context"
	"database/sql"
	"dd-prediction-api/internal/model"
	"dd-prediction-api/pkg/repository"
	"errors"

	"github.com/uptrace/bun"
)

type CoinRepository struct {
	repository.Generic[model.Coin, uint64]
}

func NewCoinRepository(
	genericRepository repository.Generic[model.Coin, uint64],
) *CoinRepository {
	return &CoinRepository{
		Generic: genericRepository,
	}
}

func (r *CoinRepository) ExistsWhereID(
	ctx context.Context,
	coinID int64,
) (bool, error) {
	coin := new(model.Coin)

	ok, err := r.DB.NewSelect().
		Model(coin).
		Where("id = ?", coinID).
		Exists(ctx)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (r *CoinRepository) GetAllCoins(
	ctx context.Context,
	options *repository.Options,
) ([]model.Coin, error) {
	coins := make([]model.Coin, 0)

	q := r.DB.NewSelect().
		Model(&coins)
	q = options.Apply(q)

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return coins, nil
}

func (r *CoinRepository) GetCoinsByCMCIDs(
	ctx context.Context,
	cmcIDs []uint64,
) ([]model.Coin, error) {
	coins := make([]model.Coin, 0, len(cmcIDs))

	err := r.DB.NewSelect().
		Model(&coins).
		Where("coins.id in (?)", bun.In(cmcIDs)).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return coins, nil
}

func (r *CoinRepository) RefreshSolanaTokens(
	ctx context.Context,
	tokens ...model.SolanaToken,
) error {
	_, err := r.DB.NewInsert().
		Model(&tokens).
		On(`CONFLICT (mint) DO UPDATE SET
    		name = EXCLUDED.name,
    		symbol = EXCLUDED.symbol,
    		decimals = EXCLUDED.decimals,
    		image_url = EXCLUDED.image_url,
			usd_price = EXCLUDED.usd_price,
			market_cap = EXCLUDED.market_cap,
			program_id = EXCLUDED.program_id,
			updated_at = NOW()`).
		Exec(ctx)

	return err
}

func (r *CoinRepository) ClearInvalidSolanaTokens(
	ctx context.Context,
) error {

	// 2 days cause solana tokens has cron with 1 day interval
	_, err := r.DB.NewDelete().
		Model((*model.SolanaToken)(nil)).
		Where("updated_at < NOW() - INTERVAL '2 days'").
		Exec(ctx)

	return err
}

func (r *CoinRepository) FindSolanaTokensBySymbolAndName(
	ctx context.Context,
	search string,
	limit int,
) ([]model.SearchSolanaTokens, error) {
	tokens := make([]model.SearchSolanaTokens, 0)

	pattern := "%" + search + "%"

	err := r.DB.NewSelect().
		Model(&tokens).
		Column("mint").
		Column("name").
		Column("symbol").
		Column("image_url").
		WhereGroup(" AND ", func(qb *bun.SelectQuery) *bun.SelectQuery {
			return qb.
				Where("symbol ILIKE ?", pattern).
				WhereOr("name ILIKE ?", pattern).
				WhereOr("mint ILIKE ?", pattern)
		}).
		Order("market_cap desc").
		Limit(limit).
		Scan(ctx)

	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *CoinRepository) GetSolanaTokenByName(
	ctx context.Context,
	token string,
) ([]model.SolanaToken, error) {
	tokenParam := "%" + token + "%"
	tokenInfo := make([]model.SolanaToken, 0)

	err := r.DB.NewSelect().
		Model(&tokenInfo).
		WhereOr("name ILIKE ?", tokenParam).
		WhereOr("symbol ILIKE ?", tokenParam).
		Order("market_cap DESC").
		Limit(10).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return tokenInfo, nil
}

func (r *CoinRepository) FindSolanaTokenBySymbol(
	ctx context.Context,
	symbol string,
) (*model.SolanaToken, error) {
	var solanaToken model.SolanaToken

	err := r.DB.NewSelect().
		Model(&solanaToken).
		Where("symbol = ?", symbol).
		Order("market_cap DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &solanaToken, nil
}

func (r *CoinRepository) FindSolanaTokenByMint(
	ctx context.Context,
	mint string,
) (*model.SolanaToken, error) {
	token := new(model.SolanaToken)

	err := r.DB.NewSelect().
		Model(token).
		WhereOr("mint = ?", mint).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return token, nil
}

func (r *CoinRepository) GetTokenUSDPricesForTVL(ctx context.Context) (map[string]float64, error) {
	var prices []model.TokenSymbolPrice
	err := r.DB.NewSelect().
		TableExpr("solana_tokens").
		Column("symbol").
		ColumnExpr("usd_price::float8 AS usd_price").
		Scan(ctx, &prices)
	if err != nil {
		return nil, err
	}
	out := make(map[string]float64, len(prices)+2)
	for _, p := range prices {
		out[p.Symbol] = p.USDPrice
	}
	out[model.USDCSymbol] = 1.0
	return out, nil
}

// GetTokenUSDPricesBySymbols returns USD prices only for the given symbols (from solana_tokens).
// USDC is always included (1.0). Safe to pass nil or empty symbols.
func (r *CoinRepository) GetTokenUSDPricesBySymbols(ctx context.Context, symbols []string) (map[string]float64, error) {
	out := make(map[string]float64, len(symbols)+2)
	out[model.USDCSymbol] = 1.0
	if len(symbols) == 0 {
		return out, nil
	}
	var prices []model.TokenSymbolPrice
	err := r.DB.NewSelect().
		TableExpr("solana_tokens").
		Column("symbol").
		ColumnExpr("usd_price::float8 AS usd_price").
		Where("symbol IN (?)", bun.In(symbols)).
		Scan(ctx, &prices)
	if err != nil {
		return nil, err
	}
	for _, p := range prices {
		out[p.Symbol] = p.USDPrice
	}
	return out, nil
}
