package mcp

import "github.com/modelcontextprotocol/go-sdk/mcp"

// Type aliases for MCP SDK types - users don't need to import the SDK directly.
type (
	Server         = mcp.Server
	ServerOptions  = mcp.ServerOptions
	ServerSession  = mcp.ServerSession
	Implementation = mcp.Implementation

	Tool            = mcp.Tool
	ToolHandler     = mcp.ToolHandler
	CallToolRequest = mcp.CallToolRequest
	CallToolResult  = mcp.CallToolResult

	Resource            = mcp.Resource
	ResourceTemplate    = mcp.ResourceTemplate
	ResourceHandler     = mcp.ResourceHandler
	ReadResourceRequest = mcp.ReadResourceRequest
	ReadResourceResult  = mcp.ReadResourceResult

	Prompt           = mcp.Prompt
	PromptHandler    = mcp.PromptHandler
	GetPromptRequest = mcp.GetPromptRequest
	GetPromptParams  = mcp.GetPromptParams
	GetPromptResult  = mcp.GetPromptResult
	PromptMessage    = mcp.PromptMessage
	PromptArgument   = mcp.PromptArgument

	Content      = mcp.Content
	TextContent  = mcp.TextContent
	ImageContent = mcp.ImageContent
	AudioContent = mcp.AudioContent

	Role        = mcp.Role
	Annotations = mcp.Annotations
)

// Function aliases.
var (
	ResourceNotFoundError = mcp.ResourceNotFoundError
)

// NewToolResultText creates a CallToolResult with text content.
func NewToolResultText(text string) *CallToolResult {
	return &CallToolResult{
		Content: []Content{&TextContent{Text: text}},
	}
}

// NewToolResultError creates a CallToolResult indicating an error.
func NewToolResultError(errMsg string) *CallToolResult {
	return &CallToolResult{
		Content: []Content{&TextContent{Text: errMsg}},
		IsError: true,
	}
}
