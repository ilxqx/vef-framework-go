package i18n

import (
	"fmt"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

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
		logger.Warnf("Translation failed for messageId '%s': %v", messageId, err)
		return messageId
	}

	return message
}

// TE implements the Translator interface with explicit error reporting.
// It attempts to localize the message using the underlying go-i18n library.
func (t *i18nTranslator) TE(messageId string, templateData ...map[string]any) (string, error) {
	// Validate messageId is not empty
	if messageId == constants.Empty {
		return constants.Empty, fmt.Errorf("messageId cannot be empty")
	}

	// Extract template data if provided (only use the first map)
	var data map[string]any
	if len(templateData) > 0 {
		data = templateData[0]
	}

	// Attempt to localize the message using go-i18n
	result, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageId,
		TemplateData: data,
	})

	if err != nil {
		return constants.Empty, fmt.Errorf("translation failed for messageId '%s': %w", messageId, err)
	}

	return result, nil
}

// New creates a new translator instance with the provided configuration.
// This constructor initializes the localizer from embedded locales and environment variables.
// It returns an error if the localizer cannot be created, allowing for graceful error handling.
//
// Parameters:
//   - config: I18nConfig containing embedded locale files
//
// Returns:
//   - Translator: A fully configured translator instance
//   - error: Error if initialization fails (e.g., missing locale files, invalid configuration)
func New(config I18nConfig) (Translator, error) {
	localizer, err := newLocalizer(config)
	if err != nil {
		return nil, err
	}

	return &i18nTranslator{localizer: localizer}, nil
}
