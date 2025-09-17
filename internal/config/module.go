package config

import (
	"github.com/ilxqx/vef-framework-go/internal/log"
	"go.uber.org/fx"
)

var (
	logger = log.Named("config") // logger is the config module logger
	Module = fx.Module(          // Module is the fx module for configuration management
		"vef:config",
		fx.Provide(
			newConfig,
			newAppConfig,
			newDatasourceConfig,
			newCorsConfig,
			newSecurityConfig,
			newRedisConfig,
			newCacheConfig,
			newI18nConfig,
		),
		fx.Invoke(func() {
			logger.Info("Config module initialized")
		}),
	)
)
