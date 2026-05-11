package click

import (
	"time"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("click",
		fx.Provide(
			CreateClickHouseConnection,
			NewRetentionBadRatesRepository,
			NewUserActivityRepository,
			NewTokenPriceRepository,
			NewUserRegistrationRepository,
			fx.Annotate(
				NewDuelTransactionRepository,
				fx.As(new(IDuelTransactionRepository)),
			),
			NewAdminAuditLogRepository,
			NewPartnerAnalyticsRepository,
		),
		fx.Invoke(
			fx.Annotate(
				func(repo *UserActivityRepository) {
					cfg := Config{
						BufferSize:    1000,
						FlushInterval: 15 * time.Minute,
					}
					InitActivityBuffer(repo, cfg)
				},
			),
		),
	)
}
