package middleware

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

// NewContentTypeMiddleware enforces JSON/multipart for state-changing requests to prevent accidental form submissions.
func NewContentTypeMiddleware() app.Middleware {
	return &SimpleMiddleware{
		handler: func(ctx fiber.Ctx) error {
			method := ctx.Method()
			if method != fiber.MethodPost && method != fiber.MethodPut ||
				webhelpers.IsJson(ctx) ||
				webhelpers.IsMultipart(ctx) {

				return ctx.Next()
			}

			return fiber.ErrUnsupportedMediaType
		},
		name:  "content_type",
		order: -700,
	}
}
