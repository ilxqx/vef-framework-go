package config

import "github.com/ilxqx/vef-framework-go/constants"

// DatasourceConfig defines database connection settings.
type DatasourceConfig struct {
	Type           constants.DBType `config:"type"`
	Host           string           `config:"host"`
	Port           uint16           `config:"port"`
	User           string           `config:"user"`
	Password       string           `config:"password"`
	Database       string           `config:"database"`
	Schema         string           `config:"schema"`
	Path           string           `config:"path"`
	EnableSQLGuard bool             `config:"enable_sql_guard"`
}
