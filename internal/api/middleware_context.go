package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/security"
)

func buildContextMiddleware(db orm.Db) fiber.Handler {
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

		request := contextx.ApiRequest(ctx)
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

		return ctx.Next()
	}
}
