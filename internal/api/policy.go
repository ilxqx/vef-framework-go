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

type DefaultApiPolicy struct {
	auth security.AuthManager
}

func (*DefaultApiPolicy) Path() string { return "/api" }
func (p *DefaultApiPolicy) BuildAuthenticationMiddleware(manager api.Manager) fiber.Handler {
	return buildAuthenticationMiddleware(manager, p.auth)
}

func NewDefaultApiPolicy(auth security.AuthManager) Policy {
	return &DefaultApiPolicy{auth: auth}
}

type OpenApiPolicy struct {
	auth security.AuthManager
}

func (*OpenApiPolicy) Path() string { return "/openapi" }
func (p *OpenApiPolicy) BuildAuthenticationMiddleware(manager api.Manager) fiber.Handler {
	return buildOpenApiAuthenticationMiddleware(manager, p.auth)
}

func NewOpenApiPolicy(auth security.AuthManager) Policy {
	return &OpenApiPolicy{auth: auth}
}
