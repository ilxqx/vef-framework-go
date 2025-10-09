package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/security"
)

// buildDataPermMiddleware creates middleware that resolves and applies data permissions.
func buildDataPermMiddleware(
	manager api.Manager,
	resolver security.DataPermissionResolver,
) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		// If no resolver is provided, skip data permission resolution
		if resolver == nil {
			logger.Debug("No DataPermissionResolver provided, skipping data permission middleware")

			return ctx.Next()
		}

		request := contextx.APIRequest(ctx)
		definition := manager.Lookup(request.Identifier)

		// Skip data permission for public endpoints or endpoints without permission requirements
		if !definition.RequiresPermission() {
			logger.Debugf("Endpoint %s does not require permission, skipping data permission", request.Identifier)

			return ctx.Next()
		}

		principal := contextx.Principal(ctx)

		// System principals have access to all data, skip data permission resolution
		if principal.Type == security.PrincipalTypeSystem {
			return ctx.Next()
		}

		// Resolve data scope for this principal and permission token
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

		// Log resolved data scope
		if dataScope != nil {
			logger.Debugf("Resolved data scope '%s' for principal %s",
				dataScope.Key(), principal.Id)
		} else {
			logger.Debugf("No data scope resolved for principal %s on permission %s",
				principal.Id, definition.PermToken)
		}

		// Construct request-scoped data permission applier using factory
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
