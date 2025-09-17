package cache

import (
	"github.com/ilxqx/vef-framework-go/cache"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var logger = log.Named("cache")

func newBadgerStore(lc fx.Lifecycle, cacheConfig *config.CacheConfig) (cache.Store, error) {
	store, err := createBadgerStore(badgerOptions{
		InMemory:   cacheConfig.Local.InMemory,
		Directory:  cacheConfig.Local.Directory,
		DefaultTTL: cacheConfig.Local.DefaultTTL,
	})

	if err != nil {
		return nil, err
	}

	lc.Append(fx.StartHook(store.Close))

	return store, nil
}

func newRedisStore(redisClient *redis.Client, cacheConfig *config.CacheConfig) cache.Store {
	return createRedisStore(redisClient, redisOptions{
		DefaultTTL: cacheConfig.Redis.DefaultTTL,
	})
}
