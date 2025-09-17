package event

import (
	"github.com/ilxqx/vef-framework-go/internal/log"
	"go.uber.org/fx"
)

var (
	logger = log.Named("event")
	Module = fx.Module(
		"vef:event",
		fx.Provide(
			fx.Annotate(
				newMemoryEventBus,
				fx.ParamTags(`group:"vef:event:middlewares"`),
			),
		),
		fx.Invoke(func() {
			logger.Info("Event module initialized")
		}),
	)
)
