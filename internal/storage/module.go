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
		NewStorageProvider,
		fx.Annotate(
			NewStorageResource,
			fx.ResultTags(`group:"vef:api:resources"`),
		),
	),
	fx.Invoke(func(provider storage.Provider) error {
		if err := provider.Setup(context.Background()); err != nil {
			return fmt.Errorf("failed to setup storage provider: %w", err)
		}
		logger.Info("Storage module initialized")

		return nil
	}),
)
