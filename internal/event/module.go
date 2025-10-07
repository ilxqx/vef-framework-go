package event

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/internal/log"
)

var (
	logger = log.Named("event")
	Module = fx.Module(
		"vef:event",
		fx.Provide(
			fx.Annotate(
				func(lc fx.Lifecycle, ctx context.Context, middlewares []event.Middleware) event.Bus {
					bus := NewMemoryBus(ctx, middlewares)

					lc.Append(
						fx.StartStopHook(
							func() error {
								if err := bus.Start(); err != nil {
									return fmt.Errorf("failed to start event bus: %w", err)
								}

								logger.Infof(
									"Memory event bus started (middlewares=%d)",
									len(middlewares),
								)

								return nil
							},
							func(stopCtx context.Context) error {
								if err := bus.Shutdown(stopCtx); err != nil {
									return fmt.Errorf("failed to stop event bus: %w", err)
								}

								logger.Infof("Memory event bus stopped")

								return nil
							},
						),
					)

					return bus
				},
				fx.ParamTags(``, ``, `group:"vef:event:middlewares"`),
			),
		),
		fx.Invoke(func() {
			logger.Info("Event module initialized")
		}),
	)
)
