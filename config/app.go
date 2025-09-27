package config

// AppConfig represents the application configuration.
type AppConfig struct {
	Name      string `config:"name"`       // Name is the name of the application
	BodyLimit string `config:"body_limit"` // BodyLimit is the body limit of the application
	Port      uint16 `config:"port"`       // Port is the port of the application
}
