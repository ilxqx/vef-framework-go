package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
)

// buildAuthorizationMiddleware creates middleware for authorization checking.
// It validates that the user has the required permissions for the requested endpoint.
func buildAuthorizationMiddleware(manager api.Manager) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		request := contextx.APIRequest(ctx)
		definition := manager.Lookup(request.Identifier)

		if definition.RequiresPermission() {
			// TODO check permission
			logger.Infof("Authorization middleware: permission check: %v", definition.PermissionToken)
		}

		return ctx.Next()
	}
}
