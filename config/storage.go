package config

import "github.com/ilxqx/vef-framework-go/constants"

// StorageConfig defines storage provider settings.
type StorageConfig struct {
	Provider   constants.StorageProvider `config:"provider"`
	MinIO      MinIOConfig               `config:"minio"`
	Filesystem FilesystemConfig          `config:"filesystem"`
}

// MinIOConfig defines MinIO storage settings.
type MinIOConfig struct {
	Endpoint  string `config:"endpoint"`
	AccessKey string `config:"access_key"`
	SecretKey string `config:"secret_key"`
	Bucket    string `config:"bucket"`
	Region    string `config:"region"`
	UseSSL    bool   `config:"use_ssl"`
}

// FilesystemConfig defines filesystem storage settings.
type FilesystemConfig struct {
	Root string `config:"root"`
}
