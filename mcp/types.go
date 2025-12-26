package mcp

// ToolDefinition defines a tool and its handler.
type ToolDefinition struct {
	// Tool is the MCP tool metadata.
	Tool *Tool
	// Handler is the function that handles tool calls.
	Handler ToolHandler
}

// ResourceDefinition defines a static resource and its handler.
type ResourceDefinition struct {
	// Resource is the MCP resource metadata.
	Resource *Resource
	// Handler is the function that handles resource reads.
	Handler ResourceHandler
}

// ResourceTemplateDefinition defines a dynamic resource template and its handler.
type ResourceTemplateDefinition struct {
	// Template is the MCP resource template metadata.
	Template *ResourceTemplate
	// Handler is the function that handles resource reads.
	Handler ResourceHandler
}

// PromptDefinition defines a prompt and its handler.
type PromptDefinition struct {
	// Prompt is the MCP prompt metadata.
	Prompt *Prompt
	// Handler is the function that handles prompt requests.
	Handler PromptHandler
}

// ServerInfo configures MCP server identification.
type ServerInfo struct {
	// Name is the server name for MCP initialization.
	Name string
	// Version is the server version string.
	Version string
	// Instructions are optional instructions for connected clients.
	Instructions string
}
