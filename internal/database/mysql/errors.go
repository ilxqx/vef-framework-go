package mysql

import "errors"

// ErrMySQLDatabaseRequired indicates database name is required for MySQL.
var ErrMySQLDatabaseRequired = errors.New("database name is required for MySQL")
