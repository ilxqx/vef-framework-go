package config

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"vef:config",
	fx.Provide(
		newConfig,
		newAppConfig,
		newDatasourceConfig,
		newCorsConfig,
		newSecurityConfig,
		newRedisConfig,
		newStorageConfig,
		newMonitorConfig,
	),
)
