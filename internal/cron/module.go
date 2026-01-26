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
)
