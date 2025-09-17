package middleware

import (
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/ilxqx/vef-framework-go/internal/app"
)

// newCompressionMiddleware returns a middleware that compresses response bodies.
// Default compression level is used; tune if needed based on performance.
func newCompressionMiddleware() app.Middleware {
	handler := compress.New(compress.Config{
		Level: compress.LevelDefault,
	})

	return &SimpleMiddleware{
		handler: handler,
		name:    "compression",
		order:   -1000,
	}
}
