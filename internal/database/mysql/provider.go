package mysql

import (
	"database/sql"
	"fmt"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/schema"

	_ "github.com/go-sql-driver/mysql"
)

// provider implements databaseProvider for MySQL
type provider struct {
	dbType constants.DbType
}

// NewProvider creates a new MySQL provider
func NewProvider() *provider {
	return &provider{
		dbType: constants.DbTypeMySQL,
	}
}

// Type returns the database type
func (p *provider) Type() constants.DbType {
	return p.dbType
}

// Connect establishes a MySQL database connection
func (p *provider) Connect(config *config.DatasourceConfig) (*sql.DB, schema.Dialect, error) {
	if err := p.ValidateConfig(config); err != nil {
		return nil, nil, err
	}

	// Build MySQL DSN (Data Source Name)
	dsn := p.buildDSN(config)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open mysql database: %w", err)
	}

	return db, mysqldialect.New(), nil
}

// ValidateConfig validates MySQL configuration
func (p *provider) ValidateConfig(config *config.DatasourceConfig) error {
	// MySQL requires at least database name
	if config.Database == constants.Empty {
		return fmt.Errorf("database name is required for MySQL")
	}
	return nil
}

// QueryVersion queries the MySQL version
func (p *provider) QueryVersion(db *bun.DB) (string, error) {
	return queryVersion(db)
}

// buildDSN constructs MySQL Data Source Name
func (p *provider) buildDSN(config *config.DatasourceConfig) string {
	host := lo.Ternary(config.Host != constants.Empty, config.Host, "127.0.0.1")
	port := lo.Ternary(config.Port != 0, config.Port, uint16(3306))
	user := lo.Ternary(config.User != constants.Empty, config.User, "root")
	password := config.Password
	database := config.Database

	// Build basic DSN: user:password@tcp(host:port)/database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, database)

	// Add common MySQL parameters for better compatibility
	dsn += "?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"

	return dsn
}
