package config

type AppConfig struct {
	Name      string         `config:"name"`
	BodyLimit string         `config:"body_limit"`
	Port      uint16         `config:"port"`
	Security  SecurityConfig `config:"security"`
}
