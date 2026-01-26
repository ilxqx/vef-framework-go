package config

// McpConfig defines MCP server settings.
type McpConfig struct {
	Enabled     bool `config:"enabled"`
	RequireAuth bool `config:"require_auth"`
}
