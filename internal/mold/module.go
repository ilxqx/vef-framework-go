package mold

import (
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/mold"
)

var logger = log.Named("mold")

// Module defines the fx module for the transformer package
// It provides dependency injection configuration for all transformer components.
var Module = fx.Module(
	"vef:mold",
	fx.Provide(
		// NewTransformer constructor with grouped dependencies
		// Collects all field transformers, struct transformers, and interceptors
		fx.Annotate(
			NewTransformer,
			fx.ParamTags(`group:"vef:mold:field_transformers"`, `group:"vef:mold:struct_transformers"`, `group:"vef:mold:interceptors"`),
		),
		// Built-in translation transformer
		fx.Annotate(
			NewTranslateTransformer,
			fx.ParamTags(`group:"vef:mold:translators"`),
			fx.ResultTags(`group:"vef:mold:field_transformers"`),
		),
		// Built-in data dictionary translator
		fx.Annotate(
			NewDataDictTranslator,
			fx.ParamTags(`optional:"true"`),
			fx.ResultTags(`group:"vef:mold:translators"`),
		),
		// Built-in cached data dictionary resolver
		fx.Annotate(
			func(loader mold.DataDictLoader, bus event.Subscriber) mold.DataDictResolver {
				if loader == nil {
					return nil
				}

				return mold.NewCachedDataDictResolver(loader, bus)
			},
			fx.ParamTags(`optional:"true"`),
		),
	),
	// Initialize the mold module
	fx.Invoke(func() {
		logger.Info("Mold module initialized")
	}),
)
