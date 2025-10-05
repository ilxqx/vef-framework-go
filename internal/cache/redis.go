package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ilxqx/vef-framework-go/cache"
	"github.com/ilxqx/vef-framework-go/constants"
)

// redisStore implements the Store interface using Redis as the storage backend.
type redisStore struct {
	client  *redis.Client
	options redisOptions
}

// redisOptions for cache configuration.
type redisOptions struct {
	DefaultTTL time.Duration // Default TTL for cache entries
}

// createRedisStore creates a new redis-based store.
func createRedisStore(client *redis.Client, opts redisOptions) cache.Store {
	return &redisStore{
		client:  client,
		options: opts,
	}
}

// Name returns the name of the store.
func (r *redisStore) Name() string {
	return "redis"
}

// Get retrieves raw bytes by key.
func (r *redisStore) Get(ctx context.Context, key string) ([]byte, bool) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Key not found, this is expected
			return nil, false
		}

		logger.Errorf("failed to get value for key %s: %v", key, err)

		return nil, false
	}

	return []byte(val), true
}

// Set stores raw bytes with the given key and optional TTL.
func (r *redisStore) Set(ctx context.Context, key string, data []byte, ttl ...time.Duration) error {
	// Determine expiration
	var expiration time.Duration
	if len(ttl) > 0 && ttl[0] > 0 {
		expiration = ttl[0]
	} else if r.options.DefaultTTL > 0 {
		expiration = r.options.DefaultTTL
	}

	// If expiration is 0, key will not expire
	err := r.client.Set(ctx, key, string(data), expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set value for key %s: %w", key, err)
	}

	return nil
}

// Contains checks if a key exists in the cache.
func (r *redisStore) Contains(ctx context.Context, key string) bool {
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false
	}

	return exists > 0
}

// Delete removes a key from the cache.
func (r *redisStore) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	return nil
}

// Clear removes all entries from the cache with the given prefix.
func (r *redisStore) Clear(ctx context.Context, prefix string) error {
	if prefix == constants.Empty {
		// If no prefix, clear entire database
		err := r.client.FlushDB(ctx).Err()
		if err != nil {
			return fmt.Errorf("failed to flush database: %w", err)
		}

		return nil
	}

	// With prefix, we need to find and delete all matching keys
	pattern := prefix + constants.Asterisk

	// Use SCAN to iterate through matching keys
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()

	var keysToDelete []string
	for iter.Next(ctx) {
		keysToDelete = append(keysToDelete, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys with pattern %s: %w", pattern, err)
	}

	// Delete keys in batches if we found any
	if len(keysToDelete) > 0 {
		err := r.client.Del(ctx, keysToDelete...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}

// Keys returns all keys in the cache, filtered by prefix.
func (r *redisStore) Keys(ctx context.Context, prefix string) ([]string, error) {
	var pattern string

	// Build the search pattern
	if prefix != constants.Empty {
		pattern = prefix + constants.Asterisk
	} else {
		pattern = constants.Asterisk
	}

	// Use SCAN for safe iteration
	var keys []string

	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan keys with pattern %s: %w", pattern, err)
	}

	return keys, nil
}

// ForEach iterates over all key-value pairs in the cache, filtered by prefix.
func (r *redisStore) ForEach(ctx context.Context, prefix string, callback func(key string, data []byte) bool) error {
	var pattern string

	// Build the search pattern
	if prefix != constants.Empty {
		pattern = prefix + constants.Asterisk
	} else {
		pattern = constants.Asterisk
	}

	// Use SCAN for safe iteration
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		redisKey := iter.Val()

		// Get the value for this key
		val, err := r.client.Get(ctx, redisKey).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				// Key might have been deleted between scan and get, skip it
				continue
			}

			return fmt.Errorf("failed to get value for key %s: %w", redisKey, err)
		}

		// Call the callback, break if it returns false
		if !callback(redisKey, []byte(val)) {
			break
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys with pattern %s: %w", pattern, err)
	}

	return nil
}

// Size returns the number of entries in the cache, filtered by prefix.
func (r *redisStore) Size(ctx context.Context, prefix string) (int64, error) {
	// If no prefix, use DBSize for efficiency
	if prefix == constants.Empty {
		size, err := r.client.DBSize(ctx).Result()
		if err != nil {
			return 0, fmt.Errorf("failed to get database size: %w", err)
		}

		return size, nil
	}

	// With prefix, we need to count matching keys
	pattern := prefix + constants.Asterisk

	var count int64

	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		count++
	}

	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to scan keys with pattern %s: %w", pattern, err)
	}

	return count, nil
}

// Close closes the cache and releases any resources.
// Note: We don't close the Redis client since it's managed by the DI container.
func (r *redisStore) Close(ctx context.Context) error {
	// Redis client is managed by fx container, so we don't close it here
	logger.Info("redis store closed (client remains managed by container)")

	return nil
}
