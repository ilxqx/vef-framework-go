package config

// CorsConfig is the configuration for the CORS middleware.
type CorsConfig struct {
	Enabled      bool     `config:"enabled"`       // Enabled is whether to enable CORS.
	AllowOrigins []string `config:"allow_origins"` // AllowOrigins is the list of allowed origins.
}
