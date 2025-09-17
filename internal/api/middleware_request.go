package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
)

// requestMiddleware parses the request body and validates the API definition exists.
// It stores the parsed request in the context for use by subsequent middlewares.
func requestMiddleware(manager api.Manager) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		var request api.Request
		if err := ctx.Bind().Body(&request); err != nil {
			return err
		}

		definition := manager.Lookup(request.Identifier)
		if definition == nil {
			return fiber.ErrNotFound
		}

		contextx.SetAPIRequest(ctx, &request)
		return ctx.Next()
	}
}
