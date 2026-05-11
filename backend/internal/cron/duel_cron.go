package cron

import (
	"context"
	"time"

	"dd-prediction-api/config"
	"dd-prediction-api/internal/service"

	rcron "github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type DuelCron struct {
	Log                       *zap.Logger
	Cron                      *rcron.Cron
	DuelQueue                 *service.DuelQueue
	DuelService               *service.DuelService
	CoinService               *service.CoinService
	AutoDuelsConf             *config.Duels
	duelsResolveJoinNotBefore time.Duration
}

const (
	RunningEveryHour = "@hourly"
)

func NewDuelCron(
	c *config.Config,
	l *zap.Logger,
	cron *rcron.Cron,
	duelQueue *service.DuelQueue,
	duelService *service.DuelService,
	coinService *service.CoinService,
) (*DuelCron, error) {
	duelCron := &DuelCron{
		Log:                       l,
		Cron:                      cron,
		DuelQueue:                 duelQueue,
		DuelService:               duelService,
		CoinService:               coinService,
		AutoDuelsConf:             &c.Duels,
		duelsResolveJoinNotBefore: c.Duels.ResolveJoinNotBefore,
	}

	_, err := duelCron.Cron.AddFunc("@every 2m", duelCron.cleanupQueueWorkers)
	if err != nil {
		return nil, err
	}

	return duelCron, nil
}

func (c *DuelCron) cleanupQueueWorkers() {
	removed := c.DuelQueue.CleanupWorkers(3 * time.Minute)

	if removed > 0 {
		c.Log.Info("cleaned duel workers",
			zap.Int("count", removed))
	}
}

func (c *DuelCron) start(_ context.Context) error {
	c.Log.Info("duel cron started")
	c.Cron.Start()
	return nil
}

func (c *DuelCron) stop(_ context.Context) error {
	c.Log.Info("duel cron stopped")
	c.Cron.Stop()
	return nil
}
