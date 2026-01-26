package mcp

// ToolProvider provides MCP tools to the server.
type ToolProvider interface {
	Tools() []ToolDefinition
}

// ResourceProvider provides static MCP resources to the server.
type ResourceProvider interface {
	Resources() []ResourceDefinition
}

// ResourceTemplateProvider provides dynamic MCP resource templates to the server.
type ResourceTemplateProvider interface {
	ResourceTemplates() []ResourceTemplateDefinition
}

// PromptProvider provides MCP prompts to the server.
type PromptProvider interface {
	Prompts() []PromptDefinition
}
