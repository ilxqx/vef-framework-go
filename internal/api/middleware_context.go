package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/security"
)

// buildContextMiddleware creates middleware that sets up contextual database and logger.
// It injects the current user's ID into the database context and creates a named logger.
func buildContextMiddleware(db orm.Db, transformer mold.Transformer) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		principal := contextx.Principal(ctx)
		if principal == nil {
			principal = security.PrincipalAnonymous
		}

		contextualDb := db.WithNamedArg(constants.PlaceholderKeyOperator, principal.Id)
		contextx.SetDb(ctx, contextualDb)
		ctx.SetContext(
			contextx.SetDb(ctx.Context(), contextualDb),
		)

		request := contextx.APIRequest(ctx)
		logger := contextx.Logger(ctx)
		contextx.SetLogger(
			ctx,
			logger.Named(request.Resource+constants.Colon+request.Action+constants.At+request.Version).
				Named(string(principal.Type)+constants.Colon+principal.Id+constants.At+principal.Name),
		)
		ctx.SetContext(
			contextx.SetLogger(
				ctx.Context(),
				logger.Named(request.Resource+constants.Colon+request.Action+constants.At+request.Version).
					Named(string(principal.Type)+constants.Colon+principal.Id+constants.At+principal.Name),
			),
		)

		contextx.SetTransformer(ctx, transformer)
		ctx.SetContext(
			contextx.SetTransformer(ctx.Context(), transformer),
		)

		return ctx.Next()
	}
}
