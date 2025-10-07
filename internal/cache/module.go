package cache

import (
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/cache"
	"github.com/ilxqx/vef-framework-go/config"
)

var Module = fx.Module(
	"vef:cache",
	fx.Provide(
		fx.Annotate(
			func(lc fx.Lifecycle, cfg *config.CacheConfig) (cache.Store, error) {
				store, err := NewBadgerStore(&cfg.Local)
				if err != nil {
					return nil, err
				}

				lc.Append(fx.StopHook(store.Close))

				return store, nil
			},
			fx.ResultTags(`name:"vef:cache:badger"`),
		),
		fx.Annotate(
			func(cfg *config.CacheConfig, client *redis.Client) cache.Store {
				return NewRedisStore(client, &cfg.Redis)
			},
			fx.ResultTags(`name:"vef:cache:redis"`),
		),
	),
	fx.Invoke(func() {
		logger.Info("Cache module initialized")
	}),
)
