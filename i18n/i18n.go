package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"golang.org/x/text/language"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/i18n/locales"
	"github.com/ilxqx/vef-framework-go/internal/log"
)

var (
	logger = log.Named("i18n")
	// SupportedLanguages defines the list of supported language codes.
	supportedLanguages = []string{"zh-CN", "en"}
	// Translator is the global translator instance initialized with embedded locales
	// No longer depends on container injection - uses environment variables for configuration.
	translator Translator
)

func init() {
	var err error
	if translator, err = New(I18nConfig{
		EmbedLocales: locales.EmbedLocales,
	}); err != nil {
		panic(err)
	}
}

// I18nConfig defines the configuration for the i18n system.
// This struct contains all necessary resources for initializing the translation system.
type I18nConfig struct {
	// EmbedLocales contains the embedded locale files (JSON format)
	// These files contain the message translations for all supported languages
	EmbedLocales embed.FS
}

// newLocalizer creates a new i18n localizer with all supported languages.
func newLocalizer(config I18nConfig) (*i18n.Localizer, error) {
	bundle := i18n.NewBundle(language.SimplifiedChinese)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	for _, lang := range supportedLanguages {
		filename := fmt.Sprintf("%s.json", lang)
		if _, err := bundle.LoadMessageFileFS(config.EmbedLocales, filename); err != nil {
			logger.Errorf("Failed to load language file %s: %v", filename, err)

			return nil, fmt.Errorf("failed to load language file %s: %w", filename, err)
		}

		logger.Debugf("Successfully loaded language file: %s", filename)
	}

	preferredLanguage := lo.CoalesceOrEmpty(os.Getenv(constants.EnvI18NLanguage), constants.DefaultI18NLanguage)

	if envLang := os.Getenv(constants.EnvI18NLanguage); envLang != constants.Empty {
		logger.Infof("Using language from environment: %s", envLang)
	} else {
		logger.Infof("Using default language: %s", constants.DefaultI18NLanguage)
	}

	return i18n.NewLocalizer(bundle, preferredLanguage), nil
}

// T is a convenient global function that translates a message ID using the global translator instance.
// It provides graceful error handling - returns the messageId as fallback if translation fails.
// This is the most commonly used function for user-facing translations.
//
// Example:
//
//	welcomeMsg := i18n.T("user_welcome", map[string]any{"name": user.Name})
func T(messageId string, templateData ...map[string]any) string {
	return translator.T(messageId, templateData...)
}

// Te is a convenient global function that translates a message ID with explicit error handling.
// Use this when you need to handle translation errors programmatically.
//
// Example:
//
//	if msg, err := i18n.Te("critical_error"); err != nil {
//	    // Handle translation failure
//	}
func Te(messageId string, templateData ...map[string]any) (string, error) {
	return translator.Te(messageId, templateData...)
}

// GetSupportedLanguages returns a list of all supported language codes.
// This is useful for validation and UI language selection.
func GetSupportedLanguages() []string {
	result := make([]string, len(supportedLanguages))
	copy(result, supportedLanguages)

	return result
}

// IsLanguageSupported checks if the given language code is supported.
// This can be used to validate language selection before setting environment variables.
func IsLanguageSupported(languageCode string) bool {
	return slices.Contains(supportedLanguages, languageCode)
}

// This is primarily intended for testing scenarios where you need to verify translations
// in different languages without restarting the process.
//
// Parameters:
//   - languageCode: The language code to use (e.g., "en", "zh-CN").
//     If empty, uses the environment variable or default language.
//
// Returns:
//   - error: Error if the language code is unsupported or initialization fails
//
// Example:
//
//	// In test setup
//	if err := i18n.SetLanguage("en"); err != nil {
//	    t.Fatal(err)
//	}
//	// Run tests with English translations
//	// ...
//	// Restore to default
//	if err := i18n.SetLanguage(""); err != nil {
//	    t.Fatal(err)
//	}
func SetLanguage(languageCode string) error {
	if languageCode == constants.Empty {
		languageCode = lo.CoalesceOrEmpty(os.Getenv(constants.EnvI18NLanguage), constants.DefaultI18NLanguage)
	}

	if !IsLanguageSupported(languageCode) {
		return fmt.Errorf("%w: %s (supported: %v)", ErrUnsupportedLanguage, languageCode, supportedLanguages)
	}

	bundle := i18n.NewBundle(language.SimplifiedChinese)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	for _, lang := range supportedLanguages {
		filename := fmt.Sprintf("%s.json", lang)
		if _, err := bundle.LoadMessageFileFS(locales.EmbedLocales, filename); err != nil {
			return fmt.Errorf("failed to load language file %s: %w", filename, err)
		}
	}

	localizer := i18n.NewLocalizer(bundle, languageCode)
	translator = &i18nTranslator{localizer: localizer}

	logger.Infof("Language set to: %s", languageCode)

	return nil
}
