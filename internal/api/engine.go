package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/security"
)

type Engine interface {
	Connect(router fiber.Router)
}

func NewEngine(
	manager api.Manager,
	policy Policy,
	checker security.PermissionChecker,
	resolver security.DataPermissionResolver,
	db orm.Db,
	publisher event.Publisher,
) Engine {
	return &DefaultEngine{
		manager:   manager,
		policy:    policy,
		checker:   checker,
		resolver:  resolver,
		db:        db,
		publisher: publisher,
	}
}

type DefaultEngine struct {
	manager   api.Manager
	policy    Policy
	checker   security.PermissionChecker
	resolver  security.DataPermissionResolver
	db        orm.Db
	publisher event.Publisher
}

func (e *DefaultEngine) Connect(router fiber.Router) {
	middlewares := e.buildMiddlewares()
	middlewares = append(middlewares, e.dispatch)

	router.Post(
		e.policy.Path(),
		middlewares[0],
		middlewares[1:]...,
	)
}

// dispatch relies on requestMiddleware having already validated the identifier exists.
func (e *DefaultEngine) dispatch(ctx fiber.Ctx) error {
	request := contextx.ApiRequest(ctx)
	definition := e.manager.Lookup(request.Identifier)

	return definition.Handler(ctx)
}

// buildMiddlewares constructs the middleware chain.
// Ordering matters: request parsing → auth → context → authorization → data permission → rate limiting → audit.
func (e *DefaultEngine) buildMiddlewares() []any {
	return []any{
		requestMiddleware(e.manager),
		e.policy.BuildAuthenticationMiddleware(e.manager),
		buildContextMiddleware(e.db),
		buildAuthorizationMiddleware(e.manager, e.checker),
		buildDataPermissionMiddleware(e.manager, e.resolver),
		buildRateLimiterMiddleware(e.manager),
		buildAuditMiddleware(e.manager, e.publisher),
	}
}
