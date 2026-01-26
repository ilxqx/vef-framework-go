package postgres

import (
	"database/sql"
	"fmt"

	"github.com/samber/lo"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
)

type provider struct {
	dbType constants.DBType
}

func NewProvider() *provider {
	return &provider{
		dbType: constants.Postgres,
	}
}

func (p *provider) Type() constants.DBType {
	return p.dbType
}

func (p *provider) Connect(cfg *config.DatasourceConfig) (*sql.DB, schema.Dialect, error) {
	if err := p.ValidateConfig(cfg); err != nil {
		return nil, nil, err
	}

	connector := pgdriver.NewConnector(
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(fmt.Sprintf(
			"%s:%d",
			lo.Ternary(cfg.Host != constants.Empty, cfg.Host, "127.0.0.1"),
			lo.Ternary(cfg.Port != 0, cfg.Port, uint16(5432)),
		)),
		pgdriver.WithInsecure(true),
		pgdriver.WithUser(lo.Ternary(cfg.User != constants.Empty, cfg.User, "postgres")),
		pgdriver.WithPassword(lo.Ternary(cfg.Password != constants.Empty, cfg.Password, "postgres")),
		pgdriver.WithDatabase(lo.Ternary(cfg.Database != constants.Empty, cfg.Database, "postgres")),
		pgdriver.WithApplicationName("vef"),
		pgdriver.WithConnParams(map[string]any{
			"search_path": lo.Ternary(cfg.Schema != constants.Empty, cfg.Schema, "public"),
		}),
	)

	return sql.OpenDB(connector), pgdialect.New(), nil
}

func (p *provider) ValidateConfig(_ *config.DatasourceConfig) error {
	return nil
}

func (p *provider) QueryVersion(db *bun.DB) (string, error) {
	return queryVersion(db)
}
