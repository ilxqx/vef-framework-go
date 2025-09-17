package middleware

import (
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/internal/app"
)

// newRecoveryMiddleware is a middleware that recovers from panics.
func newRecoveryMiddleware() app.Middleware {
	handler := recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(ctx fiber.Ctx, err any) {
			logger := contextx.Logger(ctx)
			logger.Errorf("panic: %v\n%s", err, debug.Stack())
		},
	})

	return &SimpleMiddleware{
		handler: handler,
		name:    "recovery",
		order:   -500,
	}
}
