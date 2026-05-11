package cron

import (
	"context"
	"time"

	"dd-prediction-api/config"
	"dd-prediction-api/internal/service"

	rcron "github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type UserRegistrationsCron struct {
	Log     *zap.Logger
	Cron    *rcron.Cron
	Metrics *service.MetricsService
}

func NewUserRegistrationsCron(
	c *config.Config,
	l *zap.Logger,
	cron *rcron.Cron,
	metrics *service.MetricsService,
) (*UserRegistrationsCron, error) {

	cr := &UserRegistrationsCron{
		Log:     l,
		Cron:    cron,
		Metrics: metrics,
	}

	_, err := cron.AddFunc(
		c.App.MetricsUserRegistrationsSyncInterval,
		cr.sync,
	)
	if err != nil {
		return nil, err
	}

	return cr, nil
}

func (c *UserRegistrationsCron) sync() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	err := c.Metrics.SyncUserRegistrationsToClick(ctx)
	if err != nil {
		LogErr(c.Log, err)
		return
	}

	c.Log.Info("metrics cron: user registrations synced to clickhouse")
}

func (c *UserRegistrationsCron) start(_ context.Context) error {
	c.Log.Info("user registrations cron started")
	go c.sync()
	c.Cron.Start()
	return nil
}

func (c *UserRegistrationsCron) stop(_ context.Context) error {
	c.Log.Info("user registrations cron stopped")
	c.Cron.Stop()
	return nil
}
