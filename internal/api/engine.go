package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/security"
)

// Engine defines the interface for API engines that can connect to a router.
// It provides the ability to register API endpoints with a Fiber router.
type Engine interface {
	// Connect registers the API engine with the given router.
	Connect(router fiber.Router)
}

// NewEngine creates an Engine with the given policy.
func NewEngine(
	manager api.Manager,
	policy Policy,
	checker security.PermissionChecker,
	resolver security.DataPermissionResolver,
	db orm.Db,
	transformer mold.Transformer,
	publisher event.Publisher,
) Engine {
	return &DefaultEngine{
		manager:     manager,
		policy:      policy,
		checker:     checker,
		resolver:    resolver,
		db:          db,
		transformer: transformer,
		publisher:   publisher,
	}
}

type DefaultEngine struct {
	manager     api.Manager
	policy      Policy
	checker     security.PermissionChecker
	resolver    security.DataPermissionResolver
	db          orm.Db
	transformer mold.Transformer
	publisher   event.Publisher
}

// Connect registers the API engine with the given router.
// It sets up the middleware chain and registers the API endpoint.
func (e *DefaultEngine) Connect(router fiber.Router) {
	middlewares := e.buildMiddlewares()
	middlewares = append(middlewares, e.dispatch)

	router.Post(
		e.policy.Path(),
		middlewares[0],
		middlewares[1:]...,
	)
}

// dispatch handles the API request by looking up the definition and calling its handler.
// The definition lookup is guaranteed to succeed as requestMiddleware already validates it exists.
func (e *DefaultEngine) dispatch(ctx fiber.Ctx) error {
	request := contextx.APIRequest(ctx)
	definition := e.manager.Lookup(request.Identifier)

	return definition.Handler(ctx)
}

// buildMiddlewares constructs the middleware chain for the API engine.
// The middleware order is important: request parsing, authentication, context setup, authorization, data permission, rate limiting, and audit.
func (e *DefaultEngine) buildMiddlewares() []fiber.Handler {
	return []fiber.Handler{
		requestMiddleware(e.manager),
		e.policy.BuildAuthenticationMiddleware(e.manager),
		buildContextMiddleware(e.db, e.transformer),
		buildAuthorizationMiddleware(e.manager, e.checker),
		buildDataPermMiddleware(e.manager, e.resolver),
		buildRateLimiterMiddleware(e.manager),
		buildAuditMiddleware(e.manager, e.publisher),
	}
}
