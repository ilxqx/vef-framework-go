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
	dbType constants.DBType
}

func NewProvider() *provider {
	return &provider{
		dbType: constants.SQLite,
	}
}

func (p *provider) Type() constants.DBType {
	return p.dbType
}

func (p *provider) Connect(cfg *config.DatasourceConfig) (*sql.DB, schema.Dialect, error) {
	if err := p.ValidateConfig(cfg); err != nil {
		return nil, nil, err
	}

	db, err := sql.Open(sqliteshim.ShimName, p.buildDsn(cfg))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	return db, sqlitedialect.New(), nil
}

func (p *provider) ValidateConfig(_ *config.DatasourceConfig) error {
	return nil
}

func (p *provider) QueryVersion(db *bun.DB) (string, error) {
	return queryVersion(db)
}

// buildDsn returns the DSN for SQLite. When no path is specified, it uses
// file::memory: with shared cache to ensure multiple connections share
// the same in-memory database.
func (p *provider) buildDsn(cfg *config.DatasourceConfig) string {
	if cfg.Path == constants.Empty {
		return "file::memory:?mode=memory&cache=shared"
	}

	return "file:" + cfg.Path
}
