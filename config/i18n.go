package config

// I18nConfig defines the internationalization configuration for the application.
// It supports multiple languages through language codes and provides fallback mechanisms.
type I18nConfig struct {
	Language string `config:"language"` // Language specifies the primary language code (e.g., "zh-CN", "en"). Defaults to "zh-CN" if not specified.
}
