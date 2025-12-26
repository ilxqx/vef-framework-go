package mcp

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/security"
)

type Handler struct {
	httpHandler http.Handler
}

type HandlerParams struct {
	fx.In

	McpConfig   *config.McpConfig
	Server      *mcp.Server `optional:"true"`
	AuthManager security.AuthManager
}

func NewHandler(params HandlerParams) *Handler {
	if params.Server == nil {
		return nil
	}

	server := params.Server
	var httpHandler http.Handler = mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server {
			return server
		},
		&mcp.StreamableHTTPOptions{},
	)

	if params.McpConfig.RequireAuth {
		verifier := CreateTokenVerifier(params.AuthManager)
		authMiddleware := auth.RequireBearerToken(verifier, nil)
		httpHandler = authMiddleware(httpHandler)
	}

	return &Handler{
		httpHandler: httpHandler,
	}
}

func (h *Handler) FiberHandler() fiber.Handler {
	return adaptor.HTTPHandler(h.httpHandler)
}
