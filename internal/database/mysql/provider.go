package mysql

import (
	"database/sql"
	"fmt"

	"github.com/samber/lo"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/schema"

	_ "github.com/go-sql-driver/mysql"

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

	dsn := p.buildDSN(config)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open mysql database: %w", err)
	}

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

func (p *provider) buildDSN(config *config.DatasourceConfig) string {
	host := lo.Ternary(config.Host != constants.Empty, config.Host, "127.0.0.1")
	port := lo.Ternary(config.Port != 0, config.Port, uint16(3306))
	user := lo.Ternary(config.User != constants.Empty, config.User, "root")
	password := config.Password
	database := config.Database

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, database)

	dsn += "?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"

	return dsn
}
