package i18n

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/internal/i18n/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// supportedLanguages defines the list of supported language codes
var supportedLanguages = []string{"zh-CN", "en"}

// newLocalizer creates a new i18n localizer with all supported languages.
// It loads message files from embedded resources and configures fallback behavior.
//
// Parameters:
//   - i18nConfig: Configuration containing the preferred language
//
// Returns:
//   - *i18n.Localizer: Configured localizer instance
//   - error: Error if localizer creation fails
func newLocalizer(i18nConfig *config.I18nConfig) (*i18n.Localizer, error) {
	// Create bundle with default language as Simplified Chinese
	bundle := i18n.NewBundle(language.SimplifiedChinese)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load all supported language files from embedded resources
	for _, lang := range supportedLanguages {
		filename := fmt.Sprintf("%s.json", lang)
		if _, err := bundle.LoadMessageFileFS(locales.EmbedLocales, filename); err != nil {
			return nil, fmt.Errorf("failed to load language file %s: %w", filename, err)
		}
	}

	// Create localizer with preferred language and fallback to Chinese
	// The order matters: first tries i18nConfig.Language, then falls back to "zh-CN"
	return i18n.NewLocalizer(bundle, i18nConfig.Language, "zh-CN"), nil
}
