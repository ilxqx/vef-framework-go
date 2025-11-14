package api

import (
	"fmt"

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
		if resolver == nil {
			logger.Debug("No DataPermissionResolver provided, skipping data permission middleware")

			return ctx.Next()
		}

		request := contextx.ApiRequest(ctx)
		definition := manager.Lookup(request.Identifier)

		if !definition.RequiresPermission() {
			logger.Debugf("Endpoint %q does not require permission, skipping data permission", request.Identifier)

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
			return fmt.Errorf(
				"failed to resolve data scope for principal %q on permission %q: %w",
				principal.Id,
				definition.PermToken,
				err,
			)
		}

		if dataScope != nil {
			logger.Debugf("Resolved data scope %q for principal %q",
				dataScope.Key(), principal.Id)
		} else {
			logger.Debugf("No data scope resolved for principal %q on permission %q",
				principal.Id, definition.PermToken)
		}

		applier := security.NewRequestScopedDataPermApplier(
			principal,
			dataScope,
			contextx.Logger(ctx),
		)

		contextx.SetDataPermApplier(ctx, applier)
		ctx.SetContext(contextx.SetDataPermApplier(ctx.Context(), applier))

		return ctx.Next()
	}
}
