package mcp

// ToolDefinition defines a tool and its handler.
type ToolDefinition struct {
	Tool    *Tool
	Handler ToolHandler
}

// ResourceDefinition defines a static resource and its handler.
type ResourceDefinition struct {
	Resource *Resource
	Handler  ResourceHandler
}

// ResourceTemplateDefinition defines a dynamic resource template and its handler.
type ResourceTemplateDefinition struct {
	Template *ResourceTemplate
	Handler  ResourceHandler
}

// PromptDefinition defines a prompt and its handler.
type PromptDefinition struct {
	Prompt  *Prompt
	Handler PromptHandler
}

// ServerInfo configures MCP server identification.
type ServerInfo struct {
	Name         string
	Version      string
	Instructions string
}
