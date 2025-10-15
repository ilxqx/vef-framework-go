package storage

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/storage"
)

var logger = log.Named("storage")

// Module is the FX module for storage functionality.
var Module = fx.Module("storage",
	fx.Provide(
		fx.Annotate(
			NewStorageProvider,
			fx.OnStart(func(ctx context.Context, provider storage.Provider) error {
				if err := provider.Setup(ctx); err != nil {
					return fmt.Errorf("failed to setup storage provider: %w", err)
				}

				return nil
			}),
		),
		fx.Annotate(
			NewStorageResource,
			fx.ResultTags(`group:"vef:api:resources"`),
		),
	),
)
