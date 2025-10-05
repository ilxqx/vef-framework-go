package dbhelpers

import (
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
)

// ColumnWithAlias returns the column with alias if provided.
func ColumnWithAlias(column string, alias ...string) string {
	if len(alias) == 0 || alias[0] == constants.Empty {
		return column
	}

	var sb strings.Builder
	sb.Grow(len(alias[0]) + 1 + len(column))
	_, _ = sb.WriteString(alias[0])
	_ = sb.WriteByte(constants.ByteDot)
	_, _ = sb.WriteString(column)

	return sb.String()
}
