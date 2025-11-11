package postgres

import (
	"context"

	"github.com/uptrace/bun"
)

// queryVersion queries the PostgreSQL version using version() function.
func queryVersion(db *bun.DB) (string, error) {
	var version string

	return version, db.NewSelect().
		ColumnExpr("version()").
		Scan(context.Background(), &version)
}
