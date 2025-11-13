package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
)

type provider struct {
	dbType constants.DbType
}

func NewProvider() *provider {
	return &provider{
		dbType: constants.DbSQLite,
	}
}

func (p *provider) Type() constants.DbType {
	return p.dbType
}

func (p *provider) Connect(config *config.DatasourceConfig) (*sql.DB, schema.Dialect, error) {
	if err := p.ValidateConfig(config); err != nil {
		return nil, nil, err
	}

	dsn := p.buildDSN(config)

	db, err := sql.Open(sqliteshim.ShimName, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	return db, sqlitedialect.New(), nil
}

func (p *provider) ValidateConfig(config *config.DatasourceConfig) error {
	return nil
}

func (p *provider) QueryVersion(db *bun.DB) (string, error) {
	return queryVersion(db)
}

// buildDSN uses file::memory: with shared cache to ensure multiple connections
// share the same in-memory database when no path is specified.
func (p *provider) buildDSN(config *config.DatasourceConfig) string {
	if config.Path == constants.Empty {
		return "file::memory:?mode=memory&cache=shared"
	}

	return "file:" + config.Path
}
