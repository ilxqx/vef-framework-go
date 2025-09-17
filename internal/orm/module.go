package orm

import "go.uber.org/fx"

var Module = fx.Module(
	"vef:orm",
	fx.Provide(newDb),
	fx.Invoke(func() {
		logger.Info("ORM module initialized")
	}),
)
