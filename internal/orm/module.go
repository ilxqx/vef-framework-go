package orm

import "go.uber.org/fx"

// Module provides the ORM functionality for the VEF framework.
// It registers the database provider and logs initialization status.
var Module = fx.Module(
	"vef:orm",
	fx.Provide(New),
	fx.Invoke(func() {
		logger.Info("ORM module initialized")
	}),
)
