package cron

import (
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/cron"
)

// Module provides dependency injection configuration for the cron scheduler.
// It registers the scheduler provider and initializes the cron subsystem.
var Module = fx.Module(
	"vef:cron",
	fx.Provide(newScheduler),
	fx.Provide(cron.NewScheduler),
)
