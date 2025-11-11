package config

import "github.com/ilxqx/vef-framework-go/constants"

// StorageConfig represents the configuration for storage providers.
type StorageConfig struct {
	// Provider specifies which storage provider to use
	Provider constants.StorageProvider `config:"provider"`
	// MinIO contains MinIO-specific configuration
	MinIO MinIOConfig `config:"minio"`
	// Filesystem contains filesystem-specific configuration
	Filesystem FilesystemConfig `config:"filesystem"`
}

// MinIOConfig contains configuration for MinIO storage provider.
type MinIOConfig struct {
	// Endpoint is the MinIO server endpoint (e.g., "localhost:9000")
	Endpoint string `config:"endpoint"`
	// AccessKey is the access key for authentication
	AccessKey string `config:"access_key"`
	// SecretKey is the secret key for authentication
	SecretKey string `config:"secret_key"`
	// UseSSL determines whether to use HTTPS
	UseSSL bool `config:"use_ssl"`
	// Region is the region name (optional)
	Region string `config:"region"`
	// Bucket is the default bucket for all storage operations
	Bucket string `config:"bucket"`
}

// FilesystemConfig contains configuration for filesystem storage provider.
type FilesystemConfig struct {
	// Root is the root directory for file storage (e.g., "./storage", "/var/app/files")
	Root string `config:"root"`
}
