package config

// AppConfig represents the application configuration.
type AppConfig struct {
	Name      string `config:"name"`
	BodyLimit string `config:"body_limit"`
	Port      uint16 `config:"port"`
}
