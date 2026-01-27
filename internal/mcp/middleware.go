package mcp

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/internal/app"
)

const mcpPath = "/mcp"

// McpMiddleware registers MCP routes if a handler is available.
type McpMiddleware struct {
	handler *Handler
}

// MiddlewareParams contains dependencies for creating the middleware.
type MiddlewareParams struct {
	fx.In

	Handler *Handler `optional:"true"`
}

// NewMiddleware creates a new MCP middleware.
// Returns nil if no handler is available.
func NewMiddleware(params MiddlewareParams) app.Middleware {
	if params.Handler == nil {
		return nil
	}

	return &McpMiddleware{handler: params.Handler}
}

func (*McpMiddleware) Name() string {
	return "mcp"
}

func (*McpMiddleware) Order() int {
	return 500
}

func (m *McpMiddleware) Apply(router fiber.Router) {
	if m.handler == nil {
		return
	}

	router.All(mcpPath, m.handler.FiberHandler())
	logger.Infof("MCP endpoint registered at POST %s", mcpPath)
}
