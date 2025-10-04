package testhelpers

import "time"

// Test container image constants
const (
	// PostgresImage is the default PostgreSQL test container image
	PostgresImage = "postgres:17-alpine"
	// MySQLImage is the default MySQL test container image
	MySQLImage = "mysql:lts"
	// RedisImage is the default Redis test container image
	RedisImage = "redis:8-alpine"
	// MinIOImage is the default MinIO test container image
	MinIOImage = "minio/minio:latest"
)

// Test database configuration constants
const (
	// TestDatabaseName is the default test database name
	TestDatabaseName = "testdb"
	// TestUsername is the default test database username
	TestUsername = "testuser"
	// TestPassword is the default test database password
	TestPassword = "testpass"
)

// Test MinIO configuration constants
const (
	// TestMinIOAccessKey is the default MinIO access key
	TestMinIOAccessKey = "testadmin"
	// TestMinIOSecretKey is the default MinIO secret key
	TestMinIOSecretKey = "testadmin"
	// TestMinioBucket is the default test bucket name
	TestMinioBucket = "testbucket"
)

// Test timeout constants
const (
	// DefaultContainerTimeout is the default timeout for container startup
	DefaultContainerTimeout = 30 * time.Second
)
