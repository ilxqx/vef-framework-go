package orm

import (
	"github.com/uptrace/bun"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/log"
)

// logger is the named logger instance for the Orm package.
var logger = log.Named("orm")

// New creates a new Db instance that wraps the provided bun.IDB.
// This function is used by the dependency injection system to provide Db instances.
func New(db bun.IDB) Db {
	inst := &BunDb{db: db}

	return inst.WithNamedArg(constants.ExprOperator, constants.OperatorSystem)
}
