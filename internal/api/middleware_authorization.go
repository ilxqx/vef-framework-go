package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/security"
)

// buildAuthorizationMiddleware creates middleware for authorization checking.
// It validates that the user has the required permissions for the requested endpoint.
func buildAuthorizationMiddleware(manager api.Manager, checker security.PermissionChecker) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		request := contextx.ApiRequest(ctx)
		definition := manager.Lookup(request.Identifier)

		if definition.RequiresPermission() {
			principal := contextx.Principal(ctx)

			// System principal has all permissions
			if principal.Type == security.PrincipalTypeSystem {
				return ctx.Next()
			}

			// If no permission checker is provided, deny access to protected endpoints
			if checker == nil {
				logger.Warnf("No PermissionChecker provided, denying access to permission: %s", definition.PermToken)

				return fiber.ErrForbidden
			}

			// Check if the principal has the required permission
			hasPermission, err := checker.HasPermission(ctx.Context(), principal, definition.PermToken)
			if err != nil {
				logger.Errorf("Permission check failed for principal %s on permission %s: %v",
					principal.Id, definition.PermToken, err)

				return fiber.ErrInternalServerError
			}

			if !hasPermission {
				logger.Infof("Permission denied for principal %s (type=%s) on permission %s",
					principal.Id, principal.Type, definition.PermToken)

				return fiber.ErrForbidden
			}

			logger.Debugf("Permission granted for principal %s on permission %s",
				principal.Id, definition.PermToken)
		}

		return ctx.Next()
	}
}
