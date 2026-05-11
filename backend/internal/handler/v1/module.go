package v1

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("v1",
		fx.Provide(NewAuthHandler),
		fx.Provide(NewProjectHandler),
		fx.Provide(NewDuelHandler),
		fx.Provide(NewUserHandler),
		fx.Provide(NewAdminHandler),
		fx.Provide(NewCommissionHandler),
		fx.Provide(NewCoinHandler),
		fx.Invoke(func(app *fiber.App, authHandler *AuthHandler) {
			authHandler.RegisterRoutes(app)
		}),
		fx.Invoke(func(app *fiber.App, authHandler *AuthHandler, projectHandler *ProjectHandler) {
			projectHandler.RegisterRoutes(app, authHandler)
		}),
		fx.Invoke(func(app *fiber.App, authHandler *AuthHandler, duelHandler *DuelHandler) {
			duelHandler.RegisterRoutes(app, authHandler)
		}),
		fx.Invoke(func(app *fiber.App, authHandler *AuthHandler, userHandler *UserHandler) {
			userHandler.RegisterRoutes(app, authHandler)
		}),
		fx.Invoke(func(app *fiber.App, authHandler *AuthHandler, adminHandler *AdminHandler) {
			adminHandler.RegisterRoutes(app, authHandler)
		}),
		fx.Invoke(func(app *fiber.App, authHandler *AuthHandler, commissionHandler *CommissionHandler) {
			commissionHandler.RegisterRoutes(app, authHandler)
		}),
		fx.Invoke(func(app *fiber.App, authHandler *AuthHandler, coinHandler *CoinHandler) {
			coinHandler.RegisterRoutes(app, authHandler)
		}),
	)
}
