package cache

import "time"

// MemoryOption configures the behavior of NewMemory caches.
type MemoryOption func(*memoryConfig)

// WithMemMaxSize sets the maximum number of entries the cache may hold.
// A value <= 0 disables size limits.
func WithMemMaxSize(size int64) MemoryOption {
	return func(cfg *memoryConfig) {
		cfg.maxSize = size
	}
}

// WithMemDefaultTTL sets a global TTL applied when Set is called without explicit ttl.
func WithMemDefaultTTL(ttl time.Duration) MemoryOption {
	return func(cfg *memoryConfig) {
		cfg.defaultTTL = ttl
	}
}

// WithMemEvictionPolicy selects the eviction strategy used when max size is enforced.
func WithMemEvictionPolicy(policy EvictionPolicy) MemoryOption {
	return func(cfg *memoryConfig) {
		cfg.evictionPolicy = policy
	}
}

// WithMemGCInterval sets the interval used by the background garbage collector.
func WithMemGCInterval(interval time.Duration) MemoryOption {
	return func(cfg *memoryConfig) {
		cfg.gcInterval = interval
	}
}

// RedisOption configures Redis-backed cache instances.
type RedisOption func(*redisConfig)

// WithRedisDefaultTTL sets a fallback TTL applied when Set is invoked without explicit duration.
func WithRedisDefaultTTL(ttl time.Duration) RedisOption {
	return func(cfg *redisConfig) {
		cfg.defaultTTL = ttl
	}
}
