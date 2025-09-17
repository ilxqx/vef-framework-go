package cache

import "go.uber.org/fx"

var Module = fx.Module(
	"vef:cache",
	fx.Provide(
		fx.Annotate(
			newBadgerStore,
			fx.ResultTags(`name:"vef:cache:badger"`),
		),
		fx.Annotate(
			newRedisStore,
			fx.ResultTags(`name:"vef:cache:redis"`),
		),
	),
	fx.Invoke(func() {
		logger.Info("Cache module initialized")
	}),
)
