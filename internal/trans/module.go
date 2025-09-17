package trans

import (
	"github.com/ilxqx/vef-framework-go/internal/log"
	"go.uber.org/fx"
)

// logger is the shared logger instance for the transformer module
var logger = log.Named("transformer")

// Module defines the fx module for the transformer package
// It provides dependency injection configuration for all transformer components
var Module = fx.Module(
	"vef:transformer",
	fx.Provide(
		// newTransformer constructor with grouped dependencies
		// Collects all field transformers, struct transformers, and interceptors
		fx.Annotate(
			newTransformer,
			fx.ParamTags(`group:"vef:trans:field_modifiers"`, `group:"vef:trans:struct_modifiers"`, `group:"vef:trans:interceptors"`),
		),
		// Built-in data dictionary transformer
		newDataDictTransformer,
	),
	// Initialize the transformer module
	fx.Invoke(func() {
		logger.Info("Transformer module initialized")
	}),
)
