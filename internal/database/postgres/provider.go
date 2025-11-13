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
	dbType constants.DbType
}

func NewProvider() *provider {
	return &provider{
		dbType: constants.DbPostgres,
	}
}

func (p *provider) Type() constants.DbType {
	return p.dbType
}

func (p *provider) Connect(config *config.DatasourceConfig) (*sql.DB, schema.Dialect, error) {
	if err := p.ValidateConfig(config); err != nil {
		return nil, nil, err
	}

	connector := pgdriver.NewConnector(
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(
			fmt.Sprintf(
				"%s:%d",
				lo.Ternary(config.Host != constants.Empty, config.Host, "127.0.0.1"),
				lo.Ternary(config.Port != 0, config.Port, uint16(5432)),
			),
		),
		pgdriver.WithInsecure(true),
		pgdriver.WithUser(lo.Ternary(config.User != constants.Empty, config.User, "postgres")),
		pgdriver.WithPassword(lo.Ternary(config.Password != constants.Empty, config.Password, "postgres")),
		pgdriver.WithDatabase(lo.Ternary(config.Database != constants.Empty, config.Database, "postgres")),
		pgdriver.WithApplicationName("vef"),
		pgdriver.WithConnParams(map[string]any{
			"search_path": lo.Ternary(config.Schema != constants.Empty, config.Schema, "public"),
		}),
	)

	return sql.OpenDB(connector), pgdialect.New(), nil
}

// PostgreSQL allows defaults for all connection parameters and validates during connection.
func (p *provider) ValidateConfig(config *config.DatasourceConfig) error {
	return nil
}

func (p *provider) QueryVersion(db *bun.DB) (string, error) {
	return queryVersion(db)
}
