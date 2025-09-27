package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/schema"
)

// provider implements databaseProvider for SQLite
type provider struct {
	dbType constants.DbType
}

// NewProvider creates a new SQLite provider
func NewProvider() *provider {
	return &provider{
		dbType: constants.DbTypeSQLite,
	}
}

// Type returns the database type
func (p *provider) Type() constants.DbType {
	return p.dbType
}

// Connect establishes a SQLite database connection
func (p *provider) Connect(config *config.DatasourceConfig) (*sql.DB, schema.Dialect, error) {
	if err := p.ValidateConfig(config); err != nil {
		return nil, nil, err
	}

	// Determine the data source name
	// If no path is specified, use in-memory SQLite
	dsn := p.buildDSN(config)

	db, err := sql.Open(sqliteshim.ShimName, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	return db, sqlitedialect.New(), nil
}

// ValidateConfig validates SQLite configuration
func (p *provider) ValidateConfig(config *config.DatasourceConfig) error {
	// SQLite is flexible - if no path is provided, we'll use in-memory mode
	// No validation errors needed
	return nil
}

// QueryVersion queries the SQLite version
func (p *provider) QueryVersion(db *bun.DB) (string, error) {
	return queryVersion(db)
}

// buildDSN constructs SQLite Data Source Name
func (p *provider) buildDSN(config *config.DatasourceConfig) string {
	// If no path is specified or path is empty, use in-memory SQLite
	if config.Path == constants.Empty {
		return ":memory:?cache=shared&mode=memory"
	}

	// Use the specified file path
	return config.Path + "?cache=shared&mode=memory"
}
