package cache

import (
	"github.com/redis/go-redis/v9"

	"github.com/ilxqx/vef-framework-go/constants"
)

const (
	cacheKeyPrefix = "vef" + constants.Colon + "cache"
)

// NewMemory constructs an in-memory cache using functional options.
func NewMemory[T any](opts ...MemoryOption) Cache[T] {
	cfg := defaultMemoryConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	return newMemoryCache[T](cfg)
}

// NewRedis constructs a Redis-backed cache with the given namespace.
// The namespace must be non-empty and is used to isolate keys.
func NewRedis[T any](client *redis.Client, namespace string, opts ...RedisOption) Cache[T] {
	if client == nil {
		panic("redis cache requires a non-nil redis client")
	}

	if namespace == constants.Empty {
		panic("cache.NewRedis requires a non-empty namespace")
	}

	cfg := defaultRedisConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	prefix := defaultKeyBuilder.Build(cacheKeyPrefix, namespace)

	return newRedisCache[T](client, NewPrefixKeyBuilder(prefix), cfg)
}
