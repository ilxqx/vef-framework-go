package config

type DatasourceConfig struct {
	Type     string `config:"type"`
	Path     string `config:"path"`
	Host     string `config:"host"`
	Port     uint16 `config:"port"`
	User     string `config:"user"`
	Password string `config:"password"`
	Database string `config:"database"`
	Schema   string `config:"schema"`
}
