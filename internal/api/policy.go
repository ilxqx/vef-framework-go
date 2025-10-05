package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/security"
)

// Policy defines how a specific API kind should behave.
// It encapsulates authentication, permission, limiter and path strategy.
type Policy interface {
	// Path returns the mount path for this API kind, e.g. "/api" or "/openapi".
	Path() string
	// BuildAuthenticationMiddleware returns the authentication middleware for this API kind.
	BuildAuthenticationMiddleware(manager api.Manager) fiber.Handler
}

// defaultApiPolicy is the policy for regular authenticated APIs.
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

// openApiPolicy is the policy for OpenAPI style endpoints.
type openApiPolicy struct {
	auth security.AuthManager
}

func (*openApiPolicy) Path() string { return "/openapi" }
func (p *openApiPolicy) BuildAuthenticationMiddleware(manager api.Manager) fiber.Handler {
	return buildOpenAPIAuthenticationMiddleware(manager, p.auth)
}

func NewOpenApiPolicy(auth security.AuthManager) Policy {
	return &openApiPolicy{auth: auth}
}
