package database

import (
	"errors"
	"fmt"

	"github.com/ilxqx/vef-framework-go/constants"
)

// Database error types.
var (
	ErrUnsupportedDBType  = errors.New("unsupported database type")
	errPingFailed         = errors.New("database ping failed")
	errVersionQueryFailed = errors.New("database version query failed")
)

// DatabaseError represents a database-specific error with additional context.
type DatabaseError struct {
	Type    constants.DbType
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

// newDatabaseError creates a new DatabaseError.
func newDatabaseError(dbType constants.DbType, operation string, err error, context map[string]any) *DatabaseError {
	return &DatabaseError{
		Type:    dbType,
		Op:      operation,
		Err:     err,
		Context: context,
	}
}

// Helper functions for common error scenarios.
func wrapPingError(dbType constants.DbType, err error) error {
	return newDatabaseError(dbType, "ping", fmt.Errorf("%w: %w", errPingFailed, err), nil)
}

func wrapVersionQueryError(dbType constants.DbType, err error) error {
	return newDatabaseError(dbType, "version_query", fmt.Errorf("%w: %w", errVersionQueryFailed, err), nil)
}

func newUnsupportedDbTypeError(dbType constants.DbType) error {
	return newDatabaseError(dbType, "validation", ErrUnsupportedDBType, map[string]any{
		"supported_types": []constants.DbType{constants.DbSQLite, constants.DbPostgres, constants.DbMySQL},
	})
}
