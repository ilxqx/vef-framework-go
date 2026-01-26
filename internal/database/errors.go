package database

import (
	"errors"
	"fmt"

	"github.com/ilxqx/vef-framework-go/constants"
)

var (
	ErrUnsupportedDBType  = errors.New("unsupported database type")
	errPingFailed         = errors.New("database ping failed")
	errVersionQueryFailed = errors.New("database version query failed")
)

type DatabaseError struct {
	Type    constants.DBType
	Op      string
	Err     error
	Context map[string]any
}

func (e *DatabaseError) Error() string {
	if len(e.Context) > 0 {
		return fmt.Sprintf("database error [%s] during %s: %v (context: %+v)", e.Type, e.Op, e.Err, e.Context)
	}

	return fmt.Sprintf("database error [%s] during %s: %v", e.Type, e.Op, e.Err)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}

func newDatabaseError(dbType constants.DBType, operation string, err error, context map[string]any) *DatabaseError {
	return &DatabaseError{
		Type:    dbType,
		Op:      operation,
		Err:     err,
		Context: context,
	}
}

func wrapPingError(dbType constants.DBType, err error) error {
	return newDatabaseError(dbType, "ping", fmt.Errorf("%w: %w", errPingFailed, err), nil)
}

func wrapVersionQueryError(dbType constants.DBType, err error) error {
	return newDatabaseError(dbType, "version_query", fmt.Errorf("%w: %w", errVersionQueryFailed, err), nil)
}

func newUnsupportedDBTypeError(dbType constants.DBType) error {
	return newDatabaseError(dbType, "validation", ErrUnsupportedDBType, map[string]any{
		"supported_types": []constants.DBType{constants.SQLite, constants.Postgres, constants.MySQL},
	})
}
