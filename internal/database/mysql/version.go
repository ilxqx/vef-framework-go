package mysql

import (
	"context"

	"github.com/uptrace/bun"
)

// queryVersion queries the MySQL version using version() function
func queryVersion(db *bun.DB) (version string, err error) {
	err = db.NewSelect().
		ColumnExpr("version()").
		Scan(context.Background(), &version)
	return
}
