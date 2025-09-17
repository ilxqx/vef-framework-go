package config

import "time"

// CacheConfig represents configuration for cache.
type CacheConfig struct {
	Local LocalCacheConfig `config:"local"` // Local is the configuration for local-based cache
	Redis RedisCacheConfig `config:"redis"` // Redis is the configuration for Redis-based cache
}

// LocalCacheConfig represents configuration for local-based cache.
type LocalCacheConfig struct {
	InMemory   bool          `config:"in_memory"`   // InMemory is the configuration for in-memory cache
	Directory  string        `config:"directory"`   // Directory is the configuration for directory path for persistent storage
	DefaultTTL time.Duration `config:"default_ttl"` // DefaultTTL is the configuration for default TTL
}

// RedisCacheConfig represents configuration for Redis-based cache.
type RedisCacheConfig struct {
	DefaultTTL time.Duration `config:"default_ttl"` // DefaultTTL is the configuration for default TTL
}
