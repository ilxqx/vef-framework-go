package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/security"
)

func buildDataPermissionMiddleware(
	manager api.Manager,
	resolver security.DataPermissionResolver,
) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		// If no resolver is provided, skip data permission resolution
		if resolver == nil {
			logger.Debug("No DataPermissionResolver provided, skipping data permission middleware")

			return ctx.Next()
		}

		request := contextx.ApiRequest(ctx)
		definition := manager.Lookup(request.Identifier)

		if !definition.RequiresPermission() {
			logger.Debugf("Endpoint %s does not require permission, skipping data permission", request.Identifier)

			return ctx.Next()
		}

		principal := contextx.Principal(ctx)

		if principal.Type == security.PrincipalTypeSystem {
			return ctx.Next()
		}

		dataScope, err := resolver.ResolveDataScope(
			ctx.Context(),
			principal,
			definition.PermToken,
		)
		if err != nil {
			logger.Errorf("Failed to resolve data scope for principal %s on permission %s: %v",
				principal.Id, definition.PermToken, err)

			return fiber.ErrInternalServerError
		}

		if dataScope != nil {
			logger.Debugf("Resolved data scope %q for principal %s",
				dataScope.Key(), principal.Id)
		} else {
			logger.Debugf("No data scope resolved for principal %s on permission %s",
				principal.Id, definition.PermToken)
		}

		applier := security.NewRequestScopedDataPermApplier(
			principal,
			dataScope,
			contextx.Logger(ctx),
		)

		// Store applier in context for use by handlers
		contextx.SetDataPermApplier(ctx, applier)
		ctx.SetContext(contextx.SetDataPermApplier(ctx.Context(), applier))

		return ctx.Next()
	}
}
