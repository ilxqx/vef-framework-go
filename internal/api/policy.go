package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/security"
)

// Policy encapsulates behavior differences between API kinds (e.g., /api vs /openapi).
type Policy interface {
	Path() string
	BuildAuthenticationMiddleware(manager api.Manager) fiber.Handler
}

type defaultApiPolicy struct {
	auth security.AuthManager
}

func (*defaultApiPolicy) Path() string { return "/api" }
func (p *defaultApiPolicy) BuildAuthenticationMiddleware(manager api.Manager) fiber.Handler {
	return buildAuthenticationMiddleware(manager, p.auth)
}

func NewDefaultApiPolicy(auth security.AuthManager) Policy {
	return &defaultApiPolicy{auth: auth}
}

type openApiPolicy struct {
	auth security.AuthManager
}

func (*openApiPolicy) Path() string { return "/openapi" }
func (p *openApiPolicy) BuildAuthenticationMiddleware(manager api.Manager) fiber.Handler {
	return buildOpenApiAuthenticationMiddleware(manager, p.auth)
}

func NewOpenApiPolicy(auth security.AuthManager) Policy {
	return &openApiPolicy{auth: auth}
}
