package testhelpers

import "time"

// Container images.
const (
	PostgresImage = "postgres:18-alpine"
	MySQLImage    = "mysql:lts"
	RedisImage    = "redis:8-alpine"
	MinIOImage    = "minio/minio:latest"
)

// Database credentials.
const (
	TestDatabaseName = "testdb"
	TestUsername     = "testuser"
	TestPassword     = "testpass"
)

// MinIO credentials.
const (
	TestMinIOAccessKey = "testadmin"
	TestMinIOSecretKey = "testadmin"
	TestMinioBucket    = "testbucket"
)

// Timeouts.
const (
	DefaultContainerTimeout = 30 * time.Second
)
