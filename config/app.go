package config

// AppConfig defines core application settings.
type AppConfig struct {
	Name      string `config:"name"`
	Port      uint16 `config:"port"`
	BodyLimit string `config:"body_limit"`
}
