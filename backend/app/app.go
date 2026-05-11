package app

import (
	"dd-prediction-api/config"
	"dd-prediction-api/internal/client"
	"dd-prediction-api/internal/cron"
	"dd-prediction-api/internal/handler/server"
	v1 "dd-prediction-api/internal/handler/v1"
	"dd-prediction-api/internal/service"
	"dd-prediction-api/internal/storage/cache"
	"dd-prediction-api/internal/storage/click"
	"dd-prediction-api/internal/storage/cypher"
	"dd-prediction-api/internal/storage/repository"
	auth "dd-prediction-api/pkg/jwt"
	"dd-prediction-api/pkg/logger"
	"dd-prediction-api/pkg/mailer"

	"go.uber.org/fx"
)

func Build() *fx.App {
	return fx.New(
		fx.Options(
			config.Module,
			logger.Module,
			mailer.Module,
		),
		auth.Module(),

		repository.Module(),
		cache.Module(),
		cypher.Module(),
		client.Module(),
		cron.Module(),

		service.Module(),
		server.Module(),

		v1.Module(),

		click.Module(),
	)
}
