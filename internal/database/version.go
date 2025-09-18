package database

import (
	"context"
	"errors"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/uptrace/bun"
)

// queryVersion queries the version of the database.
func queryVersion(dbType string, db *bun.DB) (string, error) {
	var version string

	switch dbType {
	case "sqlite":
		// Queries SQLite version using sqlite_version() function
		if err := db.NewSelect().ColumnExpr("sqlite_version()").Scan(context.Background(), &version); err != nil {
			return constants.Empty, err
		}

		return version, nil
	case "postgres":
		// Queries PostgreSQL version using version() function
		if err := db.NewSelect().ColumnExpr("version()").Scan(context.Background(), &version); err != nil {
			return constants.Empty, err
		}

		return version, nil
	case "mysql":
		// Queries MySQL version using version() function
		if err := db.NewSelect().ColumnExpr("version()").Scan(context.Background(), &version); err != nil {
			return constants.Empty, err
		}

		return version, nil
	default:
		// Returns error for unsupported database types
		return constants.Empty, errors.New("unsupported database type: " + dbType)
	}
}
