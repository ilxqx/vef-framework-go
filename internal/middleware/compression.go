package middleware

import (
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/ilxqx/vef-framework-go/internal/app"
)

// NewCompressionMiddleware returns a middleware that compresses response bodies.
// Default compression level is used; tune if needed based on performance.
func NewCompressionMiddleware() app.Middleware {
	handler := compress.New(compress.Config{
		Level: compress.LevelDefault,
	})

	return &SimpleMiddleware{
		handler: handler,
		name:    "compression",
		order:   -1000,
	}
}
