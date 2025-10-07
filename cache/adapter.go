package cache

import (
	"context"
	"time"
)

// cacheAdapter adapts a Store to implement the Cache[T] interface with serialization.
type cacheAdapter[T any] struct {
	store      Store
	serializer Serializer[T]
	keyBuilder KeyBuilder
}

// Get retrieves a value by key.
func (c *cacheAdapter[T]) Get(ctx context.Context, key string) (T, bool) {
	data, found := c.store.Get(ctx, c.keyBuilder.Build(key))
	if !found {
		var zero T

		return zero, false
	}

	value, err := c.serializer.Deserialize(data)
	if err != nil {
		var zero T

		return zero, false
	}

	return value, true
}

// Set stores a value with the given key and optional TTL.
func (c *cacheAdapter[T]) Set(ctx context.Context, key string, value T, ttl ...time.Duration) error {
	data, err := c.serializer.Serialize(value)
	if err != nil {
		return err
	}

	return c.store.Set(ctx, c.keyBuilder.Build(key), data, ttl...)
}

// Contains checks if a key exists in the cache.
func (c *cacheAdapter[T]) Contains(ctx context.Context, key string) bool {
	return c.store.Contains(ctx, c.keyBuilder.Build(key))
}

// Delete removes a key from the cache.
func (c *cacheAdapter[T]) Delete(ctx context.Context, key string) error {
	return c.store.Delete(ctx, c.keyBuilder.Build(key))
}

// Clear removes all entries from the cache.
func (c *cacheAdapter[T]) Clear(ctx context.Context) error {
	return c.store.Clear(ctx, c.keyBuilder.Build())
}

// Keys returns all keys in the cache, optionally filtered by prefix.
func (c *cacheAdapter[T]) Keys(ctx context.Context, prefix ...string) ([]string, error) {
	return c.store.Keys(ctx, c.keyBuilder.Build(prefix...))
}

// ForEach iterates over all key-value pairs in the cache, optionally filtered by prefix.
func (c *cacheAdapter[T]) ForEach(ctx context.Context, callback func(key string, value T) bool, prefix ...string) error {
	return c.store.ForEach(
		ctx,
		c.keyBuilder.Build(prefix...),
		func(key string, data []byte) bool {
			value, err := c.serializer.Deserialize(data)
			if err != nil {
				// Skip invalid entries
				return true
			}

			return callback(key, value)
		},
	)
}

// Size returns the number of entries in the cache.
func (c *cacheAdapter[T]) Size(ctx context.Context) (int64, error) {
	return c.store.Size(ctx, c.keyBuilder.Build())
}
