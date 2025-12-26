package mcp

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/fx"

	smcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
	ilog "github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/mcp"
)

var logger = ilog.Named("mcp")

type ServerParams struct {
	fx.In

	McpConfig         *config.McpConfig
	AppConfig         *config.AppConfig
	ToolProviders     []mcp.ToolProvider             `group:"vef:mcp:tools"`
	ResourceProviders []mcp.ResourceProvider         `group:"vef:mcp:resources"`
	TemplateProviders []mcp.ResourceTemplateProvider `group:"vef:mcp:templates"`
	PromptProviders   []mcp.PromptProvider           `group:"vef:mcp:prompts"`
	ServerInfo        *mcp.ServerInfo                `optional:"true"`
}

func NewServer(params ServerParams) *smcp.Server {
	if !params.McpConfig.Enabled {
		logger.Info("MCP is disabled by configuration")
		return nil
	}

	impl := &smcp.Implementation{
		Name:    getServerName(params),
		Version: getServerVersion(params),
	}

	opts := &smcp.ServerOptions{
		Instructions: getInstructions(params),
	}

	server := smcp.NewServer(impl, opts)

	middleware := createLoggingMiddleware()
	server.AddSendingMiddleware(middleware)
	server.AddReceivingMiddleware(middleware)

	// Register all tools
	for _, provider := range params.ToolProviders {
		for _, def := range provider.Tools() {
			server.AddTool(def.Tool, def.Handler)
			logger.Infof("Registered MCP tool: %s", def.Tool.Name)
		}
	}

	// Register all resources
	for _, provider := range params.ResourceProviders {
		for _, def := range provider.Resources() {
			server.AddResource(def.Resource, def.Handler)
			logger.Infof("Registered MCP resource: %s", def.Resource.URI)
		}
	}

	// Register all resource templates
	for _, provider := range params.TemplateProviders {
		for _, def := range provider.ResourceTemplates() {
			server.AddResourceTemplate(def.Template, def.Handler)
			logger.Infof("Registered MCP resource template: %s", def.Template.URITemplate)
		}
	}

	// Register all prompts
	for _, provider := range params.PromptProviders {
		for _, def := range provider.Prompts() {
			server.AddPrompt(def.Prompt, def.Handler)
			logger.Infof("Registered MCP prompt: %s", def.Prompt.Name)
		}
	}

	logger.Info("MCP server initialized")
	return server
}

func createLoggingMiddleware() smcp.Middleware {
	return func(next smcp.MethodHandler) smcp.MethodHandler {
		return func(ctx context.Context, method string, req smcp.Request) (smcp.Result, error) {
			start := time.Now()
			result, err := next(ctx, method, req)
			elapsed := time.Since(start)

			sessionId := req.GetSession().ID()
			params := formatParams(req.GetParams())
			latency := formatLatency(elapsed)

			if err != nil {
				logger.Errorf("Request failed: %v | method: %s | params: %s | session: %s | latency: %s", err, method, params, sessionId, latency)
			} else {
				logger.Infof("Request completed | method: %s | params: %s | session: %s | latency: %s", method, params, sessionId, latency)
			}

			return result, err
		}
	}
}

func formatLatency(elapsed time.Duration) string {
	if ms := elapsed.Milliseconds(); ms > 0 {
		return fmt.Sprintf("%dms", ms)
	}

	return fmt.Sprintf("%dμs", elapsed.Microseconds())
}

func formatParams(params smcp.Params) string {
	if params == nil {
		return "{}"
	}

	json, err := encoding.ToJson(params)
	if err != nil {
		return fmt.Sprintf("%v", params)
	}

	return json
}

func getServerName(params ServerParams) string {
	if params.ServerInfo != nil && params.ServerInfo.Name != constants.Empty {
		return params.ServerInfo.Name
	}
	if params.AppConfig != nil && params.AppConfig.Name != constants.Empty {
		return params.AppConfig.Name
	}

	return "vef-mcp-server"
}

func getServerVersion(params ServerParams) string {
	if params.ServerInfo != nil && params.ServerInfo.Version != constants.Empty {
		return params.ServerInfo.Version
	}

	return "v1.0.0"
}

func getInstructions(params ServerParams) string {
	if params.ServerInfo != nil {
		return params.ServerInfo.Instructions
	}

	return constants.Empty
}
