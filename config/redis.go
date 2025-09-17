package config

// RedisConfig represents configuration for Redis-based cache and storage.
type RedisConfig struct {
	Host     string `config:"host"`     // Host is the Redis server host
	Port     uint16 `config:"port"`     // Port is the Redis server port
	User     string `config:"user"`     // User for Redis authentication (optional)
	Password string `config:"password"` // Password for Redis authentication (optional)
	Database uint8  `config:"database"` // Database is the Redis database number (0-15)
}
