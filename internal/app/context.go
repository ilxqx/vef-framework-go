package app

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/security"
)

type CustomCtx struct {
	fiber.DefaultCtx
	logger    log.Logger         // logger is the request-scoped logger instance
	principal security.Principal // principal contains the authenticated user information
	db        orm.Db             // db is the database connection for the request
}

func (c *CustomCtx) Principal() security.Principal {
	return c.principal
}

func (c *CustomCtx) Db() orm.Db {
	return c.db
}

func (c *CustomCtx) Logger() log.Logger {
	return c.logger
}

func (c *CustomCtx) SetLogger(logger log.Logger) {
	c.logger = logger
}

func (c *CustomCtx) SetPrincipal(principal security.Principal) {
	c.principal = principal
}

func (c *CustomCtx) SetDb(db orm.Db) {
	c.db = db
}
