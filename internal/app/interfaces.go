package app

import (
	"github.com/gofiber/fiber/v3"
)

// Middleware is a middleware for the app.
type Middleware interface {
	// Name returns the name of the middleware.
	Name() string
	// Order returns the order of the middleware.
	Order() int
	// Apply applies the middleware to the router.
	Apply(router fiber.Router)
}
