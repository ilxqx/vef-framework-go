package cron

import (
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/cron"
)

// Module provides dependency injection configuration for the cron scheduler.
var Module = fx.Module(
	"vef:cron",
	fx.Provide(newScheduler, fx.Private),
	fx.Provide(cron.NewScheduler),
	fx.Provide(
		fx.Annotate(
			NewSchedulerHandlerParamResolver,
			fx.ResultTags(`group:"vef:api:handler_param_resolvers"`),
		),
		fx.Annotate(
			NewSchedulerFactoryParamResolver,
			fx.ResultTags(`group:"vef:api:factory_param_resolvers"`),
		),
	),
)
