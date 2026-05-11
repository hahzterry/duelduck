package client

import (
	"dd-prediction-api/config"
	"dd-prediction-api/internal/client/jupiter"
	"dd-prediction-api/internal/client/solana"
	"dd-prediction-api/internal/client/solscan"
	"dd-prediction-api/internal/client/walletauthority"
	dbrepo "dd-prediction-api/internal/storage/repository"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("Clients",
		fx.Provide(func(cfg *config.Config, repo *dbrepo.APIErrorRepository) *jupiter.Client {
			return jupiter.NewClient(cfg, repo)
		}),
		fx.Provide(func(cfg *config.Config, repo *dbrepo.APIErrorRepository) *solscan.Client {
			return solscan.NewClient(cfg, repo)
		}),
		fx.Provide(solana.NewClient),
		fx.Provide(walletauthority.NewClient),
	)
}
