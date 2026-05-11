package service

import (
	"dd-prediction-api/config"
	"dd-prediction-api/internal/storage/repository"
	"dd-prediction-api/pkg/sigtracker"

	"github.com/gagliardetto/solana-go/rpc"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("service",
		fx.Provide(NewAuthService),
		fx.Provide(NewUserService),
		fx.Provide(NewJWTService),
		fx.Provide(NewProjectService),
		fx.Provide(NewPriorityTracker),
		fx.Provide(func(c *config.Config, rpcClient *rpc.Client) *sigtracker.TxTracker {
			return sigtracker.NewTransactionTracker(rpcClient, c.App.SolanaWSURL)
		}),
		fx.Provide(NewTokenPriceService),
		fx.Provide(NewCoinService),
		fx.Provide(NewWalletService),
		fx.Provide(NewDuelService),
		fx.Provide(NewDuelQueue),
		fx.Provide(NewCommissionService),
		fx.Provide(NewMetricsService),
		fx.Provide(NewFileService),
		fx.Provide(func(ws *WalletService) ICommissionWalletService { return ws }),
		fx.Provide(func(cs *CoinService) ICoinService { return cs }),
		fx.Provide(func(pr *repository.ProjectRepository) repository.IProjectRepository { return pr }),
	)
}
