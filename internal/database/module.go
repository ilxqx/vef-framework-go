package database

import (
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"vef:db",
	fx.Provide(
		fx.Annotate(
			newDb,
			fx.As(new(bun.IDB)),
			fx.As(fx.Self()),
		),
	),
	fx.Invoke(func() {
		logger.Info("Database module initialized")
	}),
)
