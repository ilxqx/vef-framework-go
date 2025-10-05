package redis

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/log"
)

var logger = log.Named("redis")

// getPoolSize calculates a reasonable default pool size based on runtime environment.
// It considers GOMAXPROCS and provides sensible bounds for different deployment scenarios.
func getPoolSize() int {
	maxProcessors := runtime.GOMAXPROCS(0)
	// Base pool size: 2x GOMAXPROCS, with reasonable bounds
	poolSize := min( // Cap maximum pool size for large deployments
		max(maxProcessors*2, 4), // Ensure minimum pool size for small deployments
		100,
	)

	return poolSize
}

// getConnectionConfig returns optimized connection settings based on pool size.
func getConnectionConfig(poolSize int) (poolTimeout, idleTimeout time.Duration, maxRetries int) {
	// Pool timeout: scale with pool size but cap at reasonable limits
	poolTimeout = min(max(time.Duration(poolSize*50)*time.Millisecond, 1*time.Second), 5*time.Second)

	// Idle timeout: reasonable default for connection reuse
	idleTimeout = 5 * time.Minute

	// Max retries: conservative default
	maxRetries = 3

	return poolTimeout, idleTimeout, maxRetries
}

// newRedisClient creates and configures a Redis client with lifecycle management.
// It establishes connection, validates server info, and registers proper cleanup hooks.
func newRedisClient(lc fx.Lifecycle, ctx context.Context, cfg *config.RedisConfig, appCfg *config.AppConfig) (*redis.Client, error) {
	client := NewClient(lo.CoalesceOrEmpty(appCfg.Name, constants.VEFName+"-app"), cfg)

	// Register lifecycle hooks for proper startup and shutdown
	lc.Append(
		fx.StartStopHook(
			func(ctx context.Context) error {
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
			},
			func() error {
				logger.Info("Closing Redis client...")

				return client.Close()
			},
		),
	)

	return client, nil
}

// logRedisServerInfo queries and logs Redis server information.
func logRedisServerInfo(ctx context.Context, client *redis.Client) error {
	info, err := client.Info(ctx, "server").Result()
	if err != nil {
		return fmt.Errorf("failed to get redis server info: %w", err)
	}

	// Parse version info from INFO command result
	version := "unknown"

	for line := range strings.SplitSeq(info, constants.CarriageReturnNewline) {
		if after, ok := strings.CutPrefix(line, "redis_version:"); ok {
			version = strings.TrimSpace(after)

			break
		}
	}

	logger.Infof("Connected to Redis server: %s, version: %s", client.Options().Addr, version)

	return nil
}

// NewClient creates a new Redis client with optimized configuration.
// It applies sensible defaults for connection pooling based on runtime environment,
// while allowing customization through the config structure.
func NewClient(name string, cfg *config.RedisConfig) *redis.Client {
	// Calculate optimal connection pool settings
	poolSize := getPoolSize()
	poolTimeout, idleTimeout, maxRetries := getConnectionConfig(poolSize)

	// Build configuration with defaults and user overrides
	options := &redis.Options{
		// Basic connection settings
		ClientName:            name,
		IdentitySuffix:        constants.VEFName,
		Protocol:              3,
		ContextTimeoutEnabled: true,
		Network:               lo.Ternary(cfg.Network != constants.Empty, cfg.Network, "tcp"),
		Addr:                  buildRedisAddr(cfg),
		Username:              cfg.User,
		Password:              cfg.Password,
		DB:                    int(cfg.Database),

		// Optimized connection pool settings
		PoolSize:    poolSize,
		PoolTimeout: poolTimeout,
		MaxRetries:  maxRetries,

		// Performance optimizations
		MinIdleConns:    poolSize / 4,     // Keep 25% of pool as idle connections
		ConnMaxLifetime: 30 * time.Minute, // Rotate connections every 30 minutes
		ConnMaxIdleTime: idleTimeout,      // Maximum idle time for connections

		// Timeout configurations
		DialTimeout:  10 * time.Second,
		ReadTimeout:  6 * time.Second,
		WriteTimeout: 6 * time.Second,
	}

	client := redis.NewClient(options)

	// Log configuration for debugging
	logger.Infof(
		"Redis client configured - Pool: %d, Timeout: %v, Idle: %v, Retries: %d",
		poolSize, poolTimeout, idleTimeout, maxRetries,
	)

	return client
}

// buildRedisAddr constructs the Redis server address from configuration.
func buildRedisAddr(cfg *config.RedisConfig) string {
	host := lo.Ternary(cfg.Host != constants.Empty, cfg.Host, "127.0.0.1")
	port := lo.Ternary(cfg.Port != 0, cfg.Port, 6379)

	return fmt.Sprintf("%s:%d", host, port)
}

// HealthCheck performs a health check on the Redis client.
// It returns an error if the Redis server is not responding.
func HealthCheck(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
