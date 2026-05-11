package cron

import (
	"context"

	"dd-prediction-api/internal/service"

	rcron "github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type CoinCron struct {
	Log         *zap.Logger
	Cron        *rcron.Cron
	CoinService *service.CoinService
}

const (
	RunningDailyAt11PM = "0 23 * * *"
)

func NewCoinCron(
	l *zap.Logger,
	cron *rcron.Cron,
	coinService *service.CoinService,
) (*CoinCron, error) {
	coinCron := &CoinCron{
		Log:         l,
		Cron:        cron,
		CoinService: coinService,
	}

	_, err := coinCron.Cron.AddFunc(RunningDailyAt11PM, coinCron.refreshSolanaTokens)
	if err != nil {
		return nil, err
	}

	coinCron.refreshSolanaTokens()

	return coinCron, nil
}

func (c *CoinCron) refreshSolanaTokens() {
	err := c.CoinService.RefreshSolanaTokens(context.Background())
	if err != nil {
		LogErr(c.Log, err)
	} else {
		c.Log.Debug("coin cron: successfully refreshed solana tokens")
	}
}

func (c *CoinCron) start(_ context.Context) error {
	c.Log.Info("coin cron started")
	c.Cron.Start()
	return nil
}

func (c *CoinCron) stop(_ context.Context) error {
	c.Log.Info("coin cron stopped")
	c.Cron.Stop()
	return nil
}
