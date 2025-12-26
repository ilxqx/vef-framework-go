package mysql

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
)

type provider struct {
	dbType constants.DbType
}

func NewProvider() *provider {
	return &provider{
		dbType: constants.DbMySQL,
	}
}

func (p *provider) Type() constants.DbType {
	return p.dbType
}

func (p *provider) Connect(config *config.DatasourceConfig) (*sql.DB, schema.Dialect, error) {
	if err := p.ValidateConfig(config); err != nil {
		return nil, nil, err
	}

	cfg := p.buildConfig(config)

	connector, err := mysql.NewConnector(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create mysql connector: %w", err)
	}

	db := sql.OpenDB(connector)

	return db, mysqldialect.New(), nil
}

func (p *provider) ValidateConfig(config *config.DatasourceConfig) error {
	if config.Database == constants.Empty {
		return ErrMySQLDatabaseRequired
	}

	return nil
}

func (p *provider) QueryVersion(db *bun.DB) (string, error) {
	return queryVersion(db)
}

func (p *provider) buildConfig(config *config.DatasourceConfig) *mysql.Config {
	host := lo.Ternary(config.Host != constants.Empty, config.Host, "127.0.0.1")
	port := lo.Ternary(config.Port != 0, config.Port, uint16(3306))
	user := lo.Ternary(config.User != constants.Empty, config.User, "root")
	password := config.Password
	database := config.Database

	cfg := mysql.NewConfig()
	cfg.User = user
	cfg.Passwd = password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%s:%d", host, port)
	cfg.DBName = database
	cfg.ParseTime = true
	cfg.Collation = "utf8mb4_unicode_ci"

	return cfg
}
