package utils

import (
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
)

// ColumnWithAlias returns the column with alias if provided.
func ColumnWithAlias(column string, alias ...string) string {
	var sb strings.Builder
	if len(alias) > 0 && alias[0] != constants.Empty {
		// Write alias to string builder
		_, _ = sb.WriteString(alias[0])
		// Write dot separator
		_ = sb.WriteByte(constants.ByteDot)
	}
	// Write column name
	_, _ = sb.WriteString(column)

	return sb.String()
}

// IsDuplicateKeyError checks if the error is a duplicate key error.
func IsDuplicateKeyError(err error) bool {
	message := err.Error()
	return strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "unique constraint")
}
