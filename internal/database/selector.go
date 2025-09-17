package database

import (
	"database/sql"
	"fmt"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/schema"

	"github.com/samber/lo"
)

// selectDbDriver selects the database driver based on the database type.
func selectDbDriver(dbConfig *config.DatasourceConfig) (*sql.DB, schema.Dialect, error) {
	switch dbConfig.Type {
	case "sqlite":
		// db opens SQLite database connection
		db, err := sql.Open(sqliteshim.ShimName, dbConfig.Path)
		if err != nil {
			logger.Panicf("failed to open sqlite database: %s", err)
		}

		return db, sqlitedialect.New(), nil
	case "postgres":
		// connector creates PostgreSQL connection configuration
		connector := pgdriver.NewConnector(
			pgdriver.WithNetwork("tcp"),
			pgdriver.WithAddr(
				fmt.Sprintf(
					"%s:%d",
					lo.Ternary(dbConfig.Host != constants.Empty, dbConfig.Host, "127.0.0.1"),
					lo.Ternary(dbConfig.Port != 0, dbConfig.Port, 5432),
				),
			),
			pgdriver.WithInsecure(true),
			pgdriver.WithUser(lo.Ternary(dbConfig.User != constants.Empty, dbConfig.User, "postgres")),
			pgdriver.WithPassword(lo.Ternary(dbConfig.Password != constants.Empty, dbConfig.Password, "postgres")),
			pgdriver.WithDatabase(lo.Ternary(dbConfig.Database != constants.Empty, dbConfig.Database, "postgres")),
			pgdriver.WithApplicationName("vef"),
			// pgdriver.WithTimeout(5*time.Second),
			// pgdriver.WithDialTimeout(5*time.Second),
			// pgdriver.WithReadTimeout(5*time.Second),
			// pgdriver.WithWriteTimeout(5*time.Second),
			pgdriver.WithConnParams(map[string]any{
				"search_path": lo.Ternary(dbConfig.Schema != constants.Empty, dbConfig.Schema, "public"), // Uses configured schema or defaults to public
			}),
		)

		return sql.OpenDB(connector), pgdialect.New(), nil
	case "none":
		return nil, nil, nil
	default:
		return nil, nil, fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}
}
