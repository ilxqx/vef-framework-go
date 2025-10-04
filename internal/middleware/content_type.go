package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

// NewContentTypeMiddleware ensures JSON or multipart content types for state-changing requests.
// It bypasses checks for GET/HEAD and allows application/json and multipart/form-data.
func NewContentTypeMiddleware() app.Middleware {
	return &SimpleMiddleware{
		handler: func(ctx fiber.Ctx) error {
			method := ctx.Method()
			if method != fiber.MethodPost && method != fiber.MethodPut ||
				webhelpers.IsJSON(ctx) ||
				webhelpers.IsMultipart(ctx) {
				return ctx.Next()
			}

			return fiber.ErrUnsupportedMediaType
		},
		name:  "content_type",
		order: -700,
	}
}
