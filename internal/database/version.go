package database

import (
	"context"
	"errors"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/uptrace/bun"
)

// queryVersion queries the version of the database.
func queryVersion(dbType string, db *bun.DB) (string, error) {
	var version string // version stores the database version string

	switch dbType {
	case "sqlite":
		if err := db.NewSelect().ColumnExpr("sqlite_version()").Scan(context.Background(), &version); err != nil { // Queries SQLite version using sqlite_version() function
			return constants.Empty, err
		}

		return version, nil
	case "postgres":
		if err := db.NewSelect().ColumnExpr("version()").Scan(context.Background(), &version); err != nil { // Queries PostgreSQL version using version() function
			return constants.Empty, err
		}

		return version, nil
	case "mysql":
		if err := db.NewSelect().ColumnExpr("version()").Scan(context.Background(), &version); err != nil { // Queries MySQL version using version() function
			return constants.Empty, err
		}

		return version, nil
	default:
		return constants.Empty, errors.New("unsupported database type: " + dbType) // Returns error for unsupported database types
	}
}
