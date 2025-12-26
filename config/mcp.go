package config

// McpConfig is the MCP server configuration.
type McpConfig struct {
	Enabled     bool `config:"enabled"`
	RequireAuth bool `config:"require_auth"`
}
