package database

import (
	"context"
	"database/sql"

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
				func(lc fx.Lifecycle, cfg *config.DatasourceConfig) (db *bun.DB, err error) {
					if db, err = New(cfg); err != nil {
						return db, err
					}

					provider, exists := registry.provider(cfg.Type)
					if !exists {
						return nil, newUnsupportedDBTypeError(cfg.Type)
					}

					lc.Append(
						fx.StartStopHook(
							func(ctx context.Context) error {
								if err := db.PingContext(ctx); err != nil {
									return wrapPingError(provider.Type(), err)
								}
								if err := logDBVersion(provider, db, logger); err != nil {
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
			func(db *bun.DB) *sql.DB {
				return db.DB
			},
		),
	)
)
