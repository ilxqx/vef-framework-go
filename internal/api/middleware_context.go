package api

import (
	"strings"

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

		resourceLoggerName := buildRequestLoggerName(request.Resource, request.Action, request.Version)
		principalLoggerName := buildPrincipalLoggerName(principal)
		scopedLogger := logger.Named(resourceLoggerName).Named(principalLoggerName)

		contextx.SetLogger(ctx, scopedLogger)
		ctx.SetContext(contextx.SetLogger(ctx.Context(), scopedLogger))

		return ctx.Next()
	}
}

func buildRequestLoggerName(resource, action, version string) string {
	var sb strings.Builder

	_, _ = sb.WriteString(resource)
	_ = sb.WriteByte(constants.ByteColon)
	_, _ = sb.WriteString(action)
	_ = sb.WriteByte(constants.ByteAt)
	_, _ = sb.WriteString(version)

	return sb.String()
}

func buildPrincipalLoggerName(principal *security.Principal) string {
	var sb strings.Builder

	_, _ = sb.WriteString(string(principal.Type))
	_ = sb.WriteByte(constants.ByteColon)
	_, _ = sb.WriteString(principal.Id)
	_ = sb.WriteByte(constants.ByteAt)
	_, _ = sb.WriteString(principal.Name)

	return sb.String()
}
