package dbhelpers

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun/driver/pgdriver"
)

// IsDuplicateKeyError checks if the error is a duplicate key error.
// It first checks database-specific error codes for reliability,
// then falls back to message matching for compatibility.
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	// PostgreSQL: Check pgdriver.Error with code 23505 (unique_violation)
	var pgErr pgdriver.Error
	if errors.As(err, &pgErr) {
		return pgErr.Field('C') == "23505"
	}

	// MySQL: Check MySQLError with number 1062 (ER_DUP_ENTRY) or 1169 (ER_DUP_UNIQUE)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062 || mysqlErr.Number == 1169
	}

	// Fallback: Check error message for other databases (SQLite, SQL Server, Oracle)
	message := strings.ToLower(err.Error())

	// PostgreSQL
	if strings.Contains(message, "duplicate key") || strings.Contains(message, "unique violation") {
		return true
	}

	// MySQL
	if strings.Contains(message, "duplicate entry") {
		return true
	}

	// SQLite
	if strings.Contains(message, "unique constraint failed") {
		return true
	}

	// SQL Server (Error 2627, 2601)
	if strings.Contains(message, "violation of primary key constraint") ||
		strings.Contains(message, "violation of unique key constraint") ||
		strings.Contains(message, "cannot insert duplicate key") {

		return true
	}

	// Oracle (ORA-00001)
	if strings.Contains(message, "ora-00001") {
		return true
	}

	if strings.Contains(message, "unique constraint") && strings.Contains(message, "violated") {
		return true
	}

	return false
}

// IsForeignKeyError checks if the error is a foreign key constraint error.
// It first checks database-specific error codes for reliability,
// then falls back to message matching for compatibility.
func IsForeignKeyError(err error) bool {
	if err == nil {
		return false
	}

	// PostgreSQL: Check pgdriver.Error with code 23503 (foreign_key_violation)
	var pgErr pgdriver.Error
	if errors.As(err, &pgErr) {
		return pgErr.Field('C') == "23503"
	}

	// MySQL: Check MySQLError with number 1451 (ER_ROW_IS_REFERENCED) or 1452 (ER_NO_REFERENCED_ROW)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1451 || mysqlErr.Number == 1452
	}

	// Fallback: Check error message for other databases (SQLite, SQL Server, Oracle)
	message := strings.ToLower(err.Error())

	// PostgreSQL
	if strings.Contains(message, "violates foreign key constraint") ||
		strings.Contains(message, "foreign key violation") {

		return true
	}

	// MySQL
	if strings.Contains(message, "a foreign key constraint fails") ||
		strings.Contains(message, "cannot add or update a child row") ||
		strings.Contains(message, "cannot delete or update a parent row") {

		return true
	}

	// SQLite
	if strings.Contains(message, "foreign key constraint failed") ||
		strings.Contains(message, "sqlite_constraint_foreignkey") ||
		strings.Contains(message, "foreign key mismatch") {

		return true
	}

	// SQL Server (Msg 547)
	if strings.Contains(message, "conflicted with the foreign key constraint") ||
		strings.Contains(message, "statement conflicted with the foreign key") {

		return true
	}

	// Oracle (ORA-02291, ORA-02292)
	if strings.Contains(message, "ora-02291") || strings.Contains(message, "ora-02292") {
		return true
	}

	if strings.Contains(message, "integrity constraint") && strings.Contains(message, "violated") {
		if strings.Contains(message, "parent key not found") || strings.Contains(message, "child record found") {
			return true
		}
	}

	return false
}
