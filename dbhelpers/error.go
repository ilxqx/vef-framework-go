package dbhelpers

import "strings"

// IsDuplicateKeyError checks if the error is a duplicate key error.
// It normalizes error message to lowercase and matches common database messages.
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())

	if strings.Contains(message, "duplicate key") { // PostgreSQL
		return true
	}

	if strings.Contains(message, "unique violation") { // PostgreSQL
		return true
	}

	if strings.Contains(message, "unique constraint") { // Generic/Orm
		return true
	}

	if strings.Contains(message, "unique constraint failed") { // SQLite
		return true
	}

	if strings.Contains(message, "duplicate entry") { // MySQL
		return true
	}

	return false
}
