package storage

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/internal/contract"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/storage"
)

var logger = log.Named("storage")

// Module is the FX module for storage functionality.
var Module = fx.Module(
	"vef:storage",
	fx.Provide(
		fx.Annotate(
			NewService,
			fx.OnStart(func(ctx context.Context, service storage.Service) error {
				if initializer, ok := service.(contract.Initializer); ok {
					if err := initializer.Init(ctx); err != nil {
						return fmt.Errorf("failed to initialize storage service: %w", err)
					}
				}

				return nil
			}),
		),
		fx.Annotate(
			NewResource,
			fx.ResultTags(`group:"vef:api:resources"`),
		),
	),
)
