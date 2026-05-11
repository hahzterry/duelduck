package click

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/go-clickhouse/ch"

	"dd-prediction-api/internal/model"
)

type TokenPriceRepository struct {
	db *ch.DB
}

func NewTokenPriceRepository(db *ch.DB) *TokenPriceRepository {
	return &TokenPriceRepository{db: db}
}

func (r *TokenPriceRepository) BulkInsert(ctx context.Context, items []model.TokenPrice) error {
	if len(items) == 0 {
		return nil
	}

	_, err := r.db.NewInsert().Model(&items).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert token_prices: %w", err)
	}
	return nil
}

func (r *TokenPriceRepository) GetNearestPrice(
	ctx context.Context,
	token string,
	at time.Time,
	maxDelta time.Duration,
) (*model.TokenPrice, error) {
	at = at.UTC()

	q := r.db.NewSelect().
		Model((*model.TokenPrice)(nil)).
		Where("token = ?", token).
		OrderExpr("abs(dateDiff('second', ts, ?)) ASC", at).
		Limit(1)

	if maxDelta > 0 {
		q = q.Where("ts BETWEEN ? AND ?", at.Add(-maxDelta), at.Add(maxDelta))
	}

	var price model.TokenPrice
	err := q.Scan(ctx, &price)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if price.Token == "" {
		return &model.TokenPrice{}, nil
	}

	return &price, nil
}

func (r *TokenPriceRepository) GetLastPrice(
	ctx context.Context,
	token string,
) (*model.TokenPrice, error) {
	var price model.TokenPrice

	err := r.db.NewSelect().
		Model(&price).
		Column("token", "usd_price", "block_id", "decimals", "price_change_24h").
		ColumnExpr("toDateTime(ts, 'UTC') AS ts").
		Where("token = ?", token).
		OrderExpr("ts DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &price, nil
}

func (r *TokenPriceRepository) GetLastPricesBySymbol(
	ctx context.Context,
	symbols []string,
) (map[string]float64, error) {
	var prices []model.TokenSymbolPrice

	err := r.db.NewSelect().
		Model((*model.TokenPrice)(nil)).
		Column("symbol").
		ColumnExpr("argMax(usd_price, ts) AS usd_price").
		Where("symbol IN (?)", ch.In(symbols)).
		Group("symbol").
		Scan(ctx, &prices)
	if err != nil {
		return nil, err
	}

	tokenMap := make(map[string]float64, len(prices))
	for _, p := range prices {
		tokenMap[p.Symbol] = p.USDPrice
	}

	return tokenMap, nil
}
