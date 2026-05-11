package cron

import (
	"context"

	"dd-prediction-api/internal/service"

	rcron "github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type TokenPriceCron struct {
	Log     *zap.Logger
	Cron    *rcron.Cron
	Service *service.TokenPriceService
}

func NewTokenPriceCron(
	l *zap.Logger,
	cron *rcron.Cron,
	svc *service.TokenPriceService,
) (*TokenPriceCron, error) {
	c := &TokenPriceCron{
		Log:     l,
		Cron:    cron,
		Service: svc,
	}

	_, err := c.Cron.AddFunc(RunningEvery5Min, c.syncPrices)
	if err != nil {
		return nil, err
	}

	c.syncPrices()

	return c, nil
}

func (c *TokenPriceCron) syncPrices() {
	if err := c.Service.SyncPrices(context.Background()); err != nil {
		LogErr(c.Log, err)
		return
	}
	c.Log.Debug("token price cron: synced token prices")
}

func (c *TokenPriceCron) start(_ context.Context) error {
	c.Log.Info("token price cron started")
	c.Cron.Start()
	return nil
}

func (c *TokenPriceCron) stop(_ context.Context) error {
	c.Log.Info("token price cron stopped")
	c.Cron.Stop()
	return nil
}
