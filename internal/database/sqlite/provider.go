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

type Provider struct {
	dbType constants.DBType
}

func NewProvider() *Provider {
	return &Provider{
		dbType: constants.SQLite,
	}
}

func (p *Provider) Type() constants.DBType {
	return p.dbType
}

func (p *Provider) Connect(cfg *config.DatasourceConfig) (*sql.DB, schema.Dialect, error) {
	if err := p.ValidateConfig(cfg); err != nil {
		return nil, nil, err
	}

	db, err := sql.Open(sqliteshim.ShimName, p.buildDsn(cfg))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	return db, sqlitedialect.New(), nil
}

func (*Provider) ValidateConfig(_ *config.DatasourceConfig) error {
	return nil
}

func (*Provider) QueryVersion(db *bun.DB) (string, error) {
	return queryVersion(db)
}

// buildDsn returns the DSN for SQLite. When no path is specified, it uses
// file::memory: with shared cache to ensure multiple connections share
// the same in-memory database.
func (*Provider) buildDsn(cfg *config.DatasourceConfig) string {
	if cfg.Path == constants.Empty {
		return "file::memory:?mode=memory&cache=shared"
	}

	return "file:" + cfg.Path
}
