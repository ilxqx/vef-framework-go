package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/constants"
)

// SimpleMiddleware is middleware for the app.
type SimpleMiddleware struct {
	path    string        // Path is the path of the middleware.
	handler fiber.Handler // Handler is the handler for the middleware.
	name    string        // Name is the name of the middleware.
	order   int           // Order is the order of the middleware.
}

// Name returns the name of the middleware.
func (m *SimpleMiddleware) Name() string {
	return m.name
}

// Order returns the order of the middleware.
func (m *SimpleMiddleware) Order() int {
	return m.order
}

// Apply applies the middleware to the router.
func (m *SimpleMiddleware) Apply(router fiber.Router) {
	if m.path != constants.Empty {
		router.Use(m.path, m.handler)
	} else {
		router.Use(m.handler)
	}
}
