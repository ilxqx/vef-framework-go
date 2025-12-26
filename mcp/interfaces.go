package mcp

// ToolProvider provides MCP tools to the server.
// Users implement this interface to register custom tools.
type ToolProvider interface {
	// Tools returns a list of tool definitions.
	Tools() []ToolDefinition
}

// ResourceProvider provides static MCP resources to the server.
// Users implement this interface to register static resources.
type ResourceProvider interface {
	// Resources returns a list of resource definitions.
	Resources() []ResourceDefinition
}

// ResourceTemplateProvider provides dynamic MCP resource templates to the server.
// Users implement this interface to register dynamic resource templates.
type ResourceTemplateProvider interface {
	// ResourceTemplates returns a list of resource template definitions.
	ResourceTemplates() []ResourceTemplateDefinition
}

// PromptProvider provides MCP prompts to the server.
// Users implement this interface to register custom prompts.
type PromptProvider interface {
	// Prompts returns a list of prompt definitions.
	Prompts() []PromptDefinition
}
