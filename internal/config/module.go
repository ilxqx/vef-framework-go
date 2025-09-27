package config

import (
	"github.com/ilxqx/vef-framework-go/internal/log"
	"go.uber.org/fx"
)

var (
	logger = log.Named("config")
	Module = fx.Module(
		"vef:config",
		fx.Provide(
			newConfig,
			newAppConfig,
			newDatasourceConfig,
			newCorsConfig,
			newSecurityConfig,
			newRedisConfig,
			newCacheConfig,
		),
		fx.Invoke(func() {
			logger.Info("Config module initialized")
		}),
	)
)
