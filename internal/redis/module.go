package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

// Module provides Redis client functionality for the VEF framework.
// It configures and manages Redis connections with optimized settings based on runtime environment.
// The module automatically handles connection lifecycle, health checks, and proper cleanup.
var Module = fx.Module(
	"vef:redis",
	fx.Provide(
		fx.Annotate(
			NewClient,
			fx.OnStart(func(ctx context.Context, client *redis.Client) error {
				// Test connection
				if err := client.Ping(ctx).Err(); err != nil {
					return fmt.Errorf("failed to connect to redis: %w", err)
				}

				// Query and log redis server info
				if err := logRedisServerInfo(ctx, client); err != nil {
					return fmt.Errorf("failed to get redis server info: %w", err)
				}

				logger.Infof("Redis client started successfully: %s", client.Options().Addr)

				return nil
			}),
			fx.OnStop(func(client *redis.Client) error {
				logger.Info("Closing Redis client...")

				return client.Close()
			}),
		),
	),
	fx.Invoke(func() {
		logger.Info("Redis module initialized")
	}),
)
