package api

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/security"
)

func buildAuthorizationMiddleware(manager api.Manager, checker security.PermissionChecker) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		request := contextx.ApiRequest(ctx)
		definition := manager.Lookup(request.Identifier)

		if definition.RequiresPermission() {
			principal := contextx.Principal(ctx)

			if principal.Type == security.PrincipalTypeSystem {
				return ctx.Next()
			}

			if checker == nil {
				return fmt.Errorf(
					"%w: no 'PermissionChecker' provided, denying access to permission %q",
					fiber.ErrForbidden,
					definition.PermToken,
				)
			}

			hasPermission, err := checker.HasPermission(ctx.Context(), principal, definition.PermToken)
			if err != nil {
				return fmt.Errorf(
					"permission check failed for principal %q on permission %q: %w",
					principal.Id,
					definition.PermToken,
					err,
				)
			}

			if !hasPermission {
				return fmt.Errorf(
					"%w: principal %q (type=%s) does not have permission %q",
					fiber.ErrForbidden,
					principal.Id,
					principal.Type,
					definition.PermToken,
				)
			}
		}

		return ctx.Next()
	}
}
