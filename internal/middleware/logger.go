package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/log"
)

// NewLoggerMiddleware creates request-scoped loggers to correlate all log entries within a request.
func NewLoggerMiddleware() app.Middleware {
	return &SimpleMiddleware{
		handler: func(ctx fiber.Ctx) error {
			requestId := requestid.FromContext(ctx)
			logger := log.Named(fmt.Sprintf("request_id:%s", requestId))
			contextx.SetLogger(ctx, logger)
			contextx.SetRequestId(ctx, requestId)

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
