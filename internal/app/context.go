package app

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/security"
)

// CustomCtx is a custom Fiber context that extends the default context
// with additional framework-specific functionality like logger, principal, and database access.
type CustomCtx struct {
	fiber.DefaultCtx

	logger    log.Logger
	principal *security.Principal
	db        orm.Db
}

// Principal returns the authenticated principal (user/system/app) for the current request.
func (c *CustomCtx) Principal() *security.Principal {
	return c.principal
}

// Db returns the database connection for the current request.
func (c *CustomCtx) Db() orm.Db {
	return c.db
}

// Logger returns the logger instance for the current request.
func (c *CustomCtx) Logger() log.Logger {
	return c.logger
}

// SetLogger sets the logger instance for the current request.
func (c *CustomCtx) SetLogger(logger log.Logger) {
	c.logger = logger
}

// SetPrincipal sets the authenticated principal for the current request.
func (c *CustomCtx) SetPrincipal(principal *security.Principal) {
	c.principal = principal
}

// SetDb sets the database connection for the current request.
func (c *CustomCtx) SetDb(db orm.Db) {
	c.db = db
}
