package i18n

import (
	"fmt"
	"runtime"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/fx"
)

// translator is a global singleton instance that holds the translator dependencies.
// It's automatically populated by fx.Populate during application startup.
// This design allows for convenient global access to translation functions.
var translator = new(translatorParams)

// translatorParams defines the dependency injection structure for the global translator.
// It uses fx.In to enable automatic dependency injection of the Translator instance.
type translatorParams struct {
	fx.In
	T Translator // The translator instance injected by the fx framework
}

// Translator returns the injected translator instance with safety checks.
// It panics with detailed information if the translator is not properly initialized,
// which indicates a serious application startup issue.
func (tp *translatorParams) Translator() Translator {
	if tp.T == nil {
		// Get caller information for better debugging
		_, file, line, ok := runtime.Caller(1)
		callerInfo := "unknown"
		if ok {
			callerInfo = fmt.Sprintf("%s:%d", file, line)
		}

		panic(fmt.Sprintf(
			"i18n translator is not initialized - this indicates a serious startup issue. "+
				"Ensure the i18n.Module is properly included in your fx application and that "+
				"all dependencies are satisfied. Called from: %s",
			callerInfo,
		))
	}

	return tp.T
}

// Translator defines the interface for message translation services.
// It provides methods to translate message IDs to localized strings with optional template data,
// supporting both error-silent and error-returning translation approaches.
type Translator interface {
	// T translates a message ID to a localized string with graceful error handling.
	// If translation fails, it returns the original messageId as a fallback and logs a warning.
	// This method is suitable for user-facing scenarios where translation failure should not break the flow.
	//
	// Parameters:
	//   - messageId: The message identifier key to translate (e.g., "user_welcome", "error_not_found")
	//   - templateData: Optional template data for message interpolation (only the first map is used)
	//
	// Returns:
	//   - string: The translated message, or the original messageId if translation fails
	//
	// Example:
	//   T("user_welcome", map[string]any{"name": "John"}) -> "Welcome, John!"
	//   T("unknown_key") -> "unknown_key" (fallback)
	T(messageId string, templateData ...map[string]any) string

	// TE translates a message ID to a localized string and returns explicit error information.
	// This method is suitable for programmatic scenarios where error handling is important.
	// Use this when you need to distinguish between successful translation and failure.
	//
	// Parameters:
	//   - messageId: The message identifier key to translate (e.g., "user_welcome", "error_not_found")
	//   - templateData: Optional template data for message interpolation (only the first map is used)
	//
	// Returns:
	//   - string: The translated message if successful
	//   - error: Error describing why translation failed (message not found, template error, etc.)
	//
	// Example:
	//   message, err := TE("user_welcome", map[string]any{"name": "John"})
	//   if err != nil { /* handle translation error */ }
	TE(messageId string, templateData ...map[string]any) (string, error)
}

// i18nTranslator is the concrete implementation of the Translator interface.
// It wraps the go-i18n localizer and provides the translation functionality.
type i18nTranslator struct {
	localizer *i18n.Localizer
}

// T implements the Translator interface with graceful error handling.
// It internally calls TE and provides fallback behavior for translation failures.
func (t *i18nTranslator) T(messageId string, templateData ...map[string]any) string {
	message, err := t.TE(messageId, templateData...)
	if err != nil {
		// Log the warning but don't fail - return the original messageId as fallback
		// This ensures the application continues to work even with missing translations
		logger.Warnf("translation failed for messageId '%s': %v", messageId, err)
		return messageId
	}

	return message
}

// TE implements the Translator interface with explicit error reporting.
// It attempts to localize the message using the underlying go-i18n library.
func (t *i18nTranslator) TE(messageId string, templateData ...map[string]any) (string, error) {
	// Extract template data if provided (only use the first map)
	var data map[string]any
	if len(templateData) > 0 {
		data = templateData[0]
	}

	// Attempt to localize the message using go-i18n
	return t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageId,
		TemplateData: data,
	})
}

// newTranslator creates a new translator instance with the provided localizer.
// This function is used by the fx dependency injection system.
func newTranslator(localizer *i18n.Localizer) Translator {
	return &i18nTranslator{localizer: localizer}
}

// T is a convenient global function that translates a message ID using the global translator instance.
// It provides graceful error handling - returns the messageId as fallback if translation fails.
// This is the most commonly used function for user-facing translations.
//
// Example:
//
//	welcomeMsg := i18n.T("user_welcome", map[string]any{"name": user.Name})
func T(messageId string, templateData ...map[string]any) string {
	return translator.Translator().T(messageId, templateData...)
}

// TE is a convenient global function that translates a message ID with explicit error handling.
// Use this when you need to handle translation errors programmatically.
//
// Example:
//
//	if msg, err := i18n.TE("critical_error"); err != nil {
//	    // Handle translation failure
//	}
func TE(messageId string, templateData ...map[string]any) (string, error) {
	return translator.Translator().TE(messageId, templateData...)
}
