package sqlite

import (
	"context"

	"github.com/uptrace/bun"
)

// queryVersion queries the SQLite version using sqlite_version() function
func queryVersion(db *bun.DB) (version string, err error) {
	err = db.NewSelect().
		ColumnExpr("sqlite_version()").
		Scan(context.Background(), &version)
	return
}
