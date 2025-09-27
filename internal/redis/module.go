package redis

import "go.uber.org/fx"

// Module provides Redis client functionality for the VEF framework.
// It configures and manages Redis connections with optimized settings based on runtime environment.
// The module automatically handles connection lifecycle, health checks, and proper cleanup.
var Module = fx.Module(
	"vef:redis",
	fx.Provide(newRedisClient),
	fx.Invoke(func() {
		logger.Info("Redis module initialized")
	}),
)
