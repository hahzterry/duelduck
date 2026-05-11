package cron

import (
	"context"
	"time"

	"dd-prediction-api/config"
	v1 "dd-prediction-api/internal/handler/v1"
	"dd-prediction-api/internal/service"

	rcron "github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type RateLimitCron struct {
	Log         *zap.Logger
	Cron        *rcron.Cron
	UserService *service.UserService
}

const (
	RunningEvery5Min    = "*/5 * * * *"
	cleanupThresholdSec = 300 // seconds
)

func NewRateLimitCron(
	c *config.Config,
	l *zap.Logger,
	cron *rcron.Cron,
	userService *service.UserService,
) (*RateLimitCron, error) {
	rateLimitCron := &RateLimitCron{
		Log:         l,
		Cron:        cron,
		UserService: userService,
	}

	_, err := rateLimitCron.Cron.AddFunc(
		RunningEvery5Min,
		rateLimitCron.cleanRateLimitMap,
	)
	if err != nil {
		return nil, err
	}

	return rateLimitCron, nil
}

func (c *RateLimitCron) cleanRateLimitMap() {
	now := time.Now().Unix()

	v1.RateLimiterMu.Lock()
	defer v1.RateLimiterMu.Unlock()

	for key, rl := range v1.RateLimiter {
		if now-rl.LastRequest > cleanupThresholdSec {
			delete(v1.RateLimiter, key)
		}
	}

	c.Log.Debug("rate limit map cleaned")
}

func (c *RateLimitCron) start(_ context.Context) error {
	c.Log.Info("mention challenge cron started")
	c.Cron.Start()
	return nil
}

func (c *RateLimitCron) stop(_ context.Context) error {
	c.Log.Info("mention challenge cron stopped")
	c.Cron.Stop()
	return nil
}
