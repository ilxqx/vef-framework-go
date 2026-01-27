package middleware

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

// NewContentTypeMiddleware enforces JSON/multipart for POST/PUT requests.
func NewContentTypeMiddleware() app.Middleware {
	return &SimpleMiddleware{
		handler: func(ctx fiber.Ctx) error {
			method := ctx.Method()

			isStateChanging := method == fiber.MethodPost || method == fiber.MethodPut
			if !isStateChanging || webhelpers.IsJSON(ctx) || webhelpers.IsMultipart(ctx) {
				return ctx.Next()
			}

			return fiber.ErrUnsupportedMediaType
		},
		name:  "content_type",
		order: -700,
	}
}
