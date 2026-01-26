package dbhelpers

import "github.com/ilxqx/vef-framework-go/constants"

// ColumnWithAlias returns the column prefixed with the table alias if provided.
// For example: ColumnWithAlias("name", "u") returns "u.name".
func ColumnWithAlias(column string, alias ...string) string {
	if len(alias) == 0 || alias[0] == constants.Empty {
		return column
	}

	return alias[0] + constants.Dot + column
}
