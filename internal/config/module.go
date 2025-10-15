package config

import (
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/internal/log"
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
			newStorageConfig,
		),
	)
)
