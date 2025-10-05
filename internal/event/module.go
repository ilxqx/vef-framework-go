package event

import (
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/internal/log"
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
