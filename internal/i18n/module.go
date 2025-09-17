package i18n

import (
	"github.com/ilxqx/vef-framework-go/internal/log"
	"go.uber.org/fx"
)

// logger is the shared logger instance for the i18n module
var logger = log.Named("i18n")

// Module defines the fx module for the internationalization system.
var Module = fx.Module(
	"vef:i18n",
	// Provides the i18n.Localizer and Translator instance
	fx.Provide(newLocalizer, newTranslator),
	// Populates the global translator with dependencies
	fx.Populate(translator),
	fx.Invoke(func() {
		logger.Info("I18n module initialized")
	}),
)
