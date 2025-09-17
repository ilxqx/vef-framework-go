package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/internal/app"
)

// newHeadersMiddleware returns a middleware that sets security-related response headers.
// It adds X-Content-Type-Options, Strict-Transport-Security (for HTTPS), and a default Cache-Control if missing.
func newHeadersMiddleware() app.Middleware {
	return &SimpleMiddleware{
		handler: func(ctx fiber.Ctx) error {
			// Continue stack
			if err := ctx.Next(); err != nil {
				return err
			}

			// Set headers
			ctx.Set(fiber.HeaderXContentTypeOptions, "nosniff")
			if ctx.Protocol() == "https" {
				ctx.Set(fiber.HeaderStrictTransportSecurity, "max-age=31536000; includeSubDomains")
			}
			if len(ctx.Response().Header.Peek(fiber.HeaderCacheControl)) == 0 {
				ctx.Set(fiber.HeaderCacheControl, "no-store, no-cache, must-revalidate, max-age=0")
			}

			return nil
		},
		name:  "headers",
		order: -900,
	}
}
