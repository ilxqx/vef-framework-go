package config

import "github.com/ilxqx/vef-framework-go/constants"

type DatasourceConfig struct {
	Type     constants.DbType `config:"type"`
	Path     string           `config:"path"`
	Host     string           `config:"host"`
	Port     uint16           `config:"port"`
	User     string           `config:"user"`
	Password string           `config:"password"`
	Database string           `config:"database"`
	Schema   string           `config:"schema"`
}
