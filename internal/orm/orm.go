package orm

import (
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/uptrace/bun"
)

// logger is the named logger instance for the ORM package.
var logger = log.Named("orm")

// New creates a new Db instance that wraps the provided bun.IDB.
// This function is used by the dependency injection system to provide Db instances.
func New(db bun.IDB) Db {
	return &bunDb{db: db}
}
