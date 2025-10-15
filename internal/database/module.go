package database

import (
	"context"

	"github.com/uptrace/bun"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/internal/log"
)

var (
	logger = log.Named("database")
	Module = fx.Module(
		"vef:database",
		fx.Provide(
			fx.Annotate(
				func(lc fx.Lifecycle, config *config.DatasourceConfig) (db *bun.DB, err error) {
					if db, err = New(config); err != nil {
						return db, err
					}

					// Get provider for StartHook validation
					provider, exists := registry.provider(config.Type)
					if !exists {
						return nil, newUnsupportedDbTypeError(config.Type)
					}

					// Register lifecycle hooks for proper startup and shutdown
					lc.Append(
						fx.StartStopHook(
							func(ctx context.Context) error {
								// Validate connection
								if err := db.PingContext(ctx); err != nil {
									return wrapPingError(provider.Type(), err)
								}

								// Log database version
								if err := logDbVersion(provider, db, logger); err != nil {
									return err
								}

								logger.Infof("Database client started successfully: %s", provider.Type())

								return nil
							},
							func() error {
								logger.Info("Closing database connection...")

								return db.Close()
							},
						),
					)

					return db, err
				},
				fx.As(new(bun.IDB)),
				fx.As(fx.Self()),
			),
		),
	)
)
