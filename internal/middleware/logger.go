package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/log"
)

// NewLoggerMiddleware returns a middleware that initializes a request-scoped logger.
// The logger is stored in fiber context via contextx.SetLogger and can be retrieved later.
// This middleware depends on requestid middleware to ensure request id is available.
func NewLoggerMiddleware() app.Middleware {
	return &SimpleMiddleware{
		handler: func(ctx fiber.Ctx) error {
			requestId := requestid.FromContext(ctx)
			logger := log.Named(fmt.Sprintf("request_id:%s", requestId))
			contextx.SetLogger(ctx, logger)

			ctx.SetContext(
				contextx.SetLogger(
					contextx.SetRequestId(ctx.Context(), requestId),
					logger,
				),
			)

			return ctx.Next()
		},
		name:  "logger",
		order: -600,
	}
}
