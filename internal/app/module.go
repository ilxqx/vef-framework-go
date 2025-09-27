package app

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"vef:app",
	fx.Provide(New),
	fx.Invoke(func() {
		logger.Info("App module initialized")
	}),
)
