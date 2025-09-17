package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"go.uber.org/fx"
)

var logger = log.Named("redis")

func newRedisClient(lc fx.Lifecycle, ctx context.Context, appConfig *config.AppConfig, redisConfig *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		ClientName:            appConfig.Name,
		IdentitySuffix:        "vef",
		Protocol:              3,
		ContextTimeoutEnabled: true,
		Network:               "tcp",
		Addr: fmt.Sprintf(
			"%s:%d",
			lo.Ternary(redisConfig.Host != constants.Empty, redisConfig.Host, "127.0.0.1"),
			lo.Ternary(redisConfig.Port != 0, redisConfig.Port, 6379),
		),
		Username: redisConfig.User,
		Password: redisConfig.Password,
		DB:       int(redisConfig.Database),
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", err)
	}

	// Query redis server info
	info, err := client.Info(ctx, "server").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get redis server info: %v", err)
	} else {
		// Parse version info from INFO command result
		for line := range strings.SplitSeq(info, constants.CarriageReturnNewline) {
			if after, ok := strings.CutPrefix(line, "redis_version:"); ok {
				version := after
				logger.Infof("successfully connected to redis: %s, version: %s", client.Options().Addr, version)
				break
			}
		}
	}

	lc.Append(fx.StartHook(client.Close))

	return client, nil
}
