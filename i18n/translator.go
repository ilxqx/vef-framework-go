package i18n

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/ilxqx/vef-framework-go/constants"
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

	// Te translates a message ID to a localized string and returns explicit error information.
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
	//   message, err := Te("user_welcome", map[string]any{"name": "John"})
	//   if err != nil { /* handle translation error */ }
	Te(messageId string, templateData ...map[string]any) (string, error)
}

// i18nTranslator is the concrete implementation of the Translator interface.
// It wraps the go-i18n localizer and provides the translation functionality.
type i18nTranslator struct {
	localizer *i18n.Localizer
}

// T implements the Translator interface with graceful error handling.
// It internally calls TE and provides fallback behavior for translation failures.
func (t *i18nTranslator) T(messageId string, templateData ...map[string]any) string {
	message, err := t.Te(messageId, templateData...)
	if err != nil {
		logger.Warnf("Translation failed for messageId %q: %v", messageId, err)
		return messageId
	}

	return message
}

// Te implements the Translator interface with explicit error reporting.
// It attempts to localize the message using the underlying go-i18n library.
func (t *i18nTranslator) Te(messageId string, templateData ...map[string]any) (string, error) {
	if messageId == constants.Empty {
		return constants.Empty, ErrMessageIdEmpty
	}

	var data map[string]any
	if len(templateData) > 0 {
		data = templateData[0]
	}

	result, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageId,
		TemplateData: data,
	})
	if err != nil {
		return constants.Empty, fmt.Errorf("translation failed for messageId %q: %w", messageId, err)
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
func New(config I18nConfig) (*i18nTranslator, error) {
	localizer, err := newLocalizer(config)
	if err != nil {
		return nil, err
	}

	return &i18nTranslator{localizer: localizer}, nil
}
