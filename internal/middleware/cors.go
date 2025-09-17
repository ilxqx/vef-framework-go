package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/app"
)

// newCorsMiddleware is a middleware that handles CORS.
func newCorsMiddleware(config *config.CorsConfig) app.Middleware {
	handler := cors.New(cors.Config{
		Next: func(ctx fiber.Ctx) bool {
			return !config.Enabled
		},
		AllowOrigins: config.AllowOrigins,
		AllowMethods: []string{
			fiber.MethodHead,
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodDelete,
		},
		AllowHeaders: []string{
			fiber.HeaderContentType,
			fiber.HeaderAuthorization,
			constants.HeaderXAppId,
			constants.HeaderXTimestamp,
			constants.HeaderXSignature,
		},
		AllowCredentials: false,
		ExposeHeaders:    []string{},
		MaxAge:           7200,
	})

	return &SimpleMiddleware{
		handler: handler,
		name:    "cors",
		order:   -800,
	}
}
