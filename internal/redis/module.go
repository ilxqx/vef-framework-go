package redis

import "go.uber.org/fx"

var Module = fx.Module(
	"vef:redis",
	fx.Provide(newRedisClient),
	fx.Invoke(func() {
		logger.Info("Redis module initialized")
	}),
)
