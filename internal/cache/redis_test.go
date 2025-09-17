package cache

import (
	"context"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRedisClient sets up a Redis client for testing
// This assumes Redis is running on localhost:6379 for testing
func setupRedisClient(t *testing.T) *redis.Client {
	// Skip Redis tests if REDIS_URL is not provided
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis tests")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Skipf("Failed to parse REDIS_URL: %v", err)
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		t.Skipf("Redis server not available: %v", err)
	}

	// Clean up any existing test keys
	testPattern := "test:*"
	keys, _ := client.Keys(ctx, testPattern).Result()
	if len(keys) > 0 {
		client.Del(ctx, keys...)
	}

	return client
}

func TestRedisStoreBasicOperations(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	store := createRedisStore(client, redisOptions{
		DefaultTTL: 0, // No default TTL
	})

	ctx := context.Background()

	// Cleanup function to run after all subtests
	defer func() {
		_ = store.Clear(ctx, "")
	}()

	t.Run("Set and Get", func(t *testing.T) {
		testData := []byte(`{"name":"test","value":42}`)

		err := store.Set(ctx, "test-key", testData)
		require.NoError(t, err)

		result, found := store.Get(ctx, "test-key")
		assert.True(t, found)
		assert.Equal(t, testData, result)
	})

	t.Run("Contains", func(t *testing.T) {
		testData := []byte(`{"name":"exists","value":1}`)

		err := store.Set(ctx, "exists-key", testData)
		require.NoError(t, err)

		assert.True(t, store.Contains(ctx, "exists-key"))
		assert.False(t, store.Contains(ctx, "not-exists-key"))
	})

	t.Run("Delete", func(t *testing.T) {
		testData := []byte(`{"name":"delete","value":2}`)

		err := store.Set(ctx, "delete-key", testData)
		require.NoError(t, err)

		assert.True(t, store.Contains(ctx, "delete-key"))

		err = store.Delete(ctx, "delete-key")
		require.NoError(t, err)

		assert.False(t, store.Contains(ctx, "delete-key"))
	})

	t.Run("Update existing key", func(t *testing.T) {
		originalData := []byte(`{"name":"original","value":1}`)
		updatedData := []byte(`{"name":"updated","value":2}`)

		err := store.Set(ctx, "update-key", originalData)
		require.NoError(t, err)

		result, found := store.Get(ctx, "update-key")
		assert.True(t, found)
		assert.Equal(t, originalData, result)

		err = store.Set(ctx, "update-key", updatedData)
		require.NoError(t, err)

		result, found = store.Get(ctx, "update-key")
		assert.True(t, found)
		assert.Equal(t, updatedData, result)
	})
}

func TestRedisStoreTTL(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	t.Run("TTL expiration", func(t *testing.T) {
		store := createRedisStore(client, redisOptions{})

		ctx := context.Background()
		err := store.Set(ctx, "ttl-key", []byte("ttl-value"), 100*time.Millisecond)
		require.NoError(t, err)

		// Should exist immediately
		value, found := store.Get(ctx, "ttl-key")
		assert.True(t, found)
		assert.Equal(t, []byte("ttl-value"), value)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should be expired
		_, found = store.Get(ctx, "ttl-key")
		assert.False(t, found)
	})

	t.Run("Default TTL", func(t *testing.T) {
		store := createRedisStore(client, redisOptions{
			DefaultTTL: 100 * time.Millisecond,
		})

		ctx := context.Background()
		err := store.Set(ctx, "default-ttl-key", []byte("default-ttl-value"))
		require.NoError(t, err)

		// Should exist immediately
		value, found := store.Get(ctx, "default-ttl-key")
		assert.True(t, found)
		assert.Equal(t, []byte("default-ttl-value"), value)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should be expired
		_, found = store.Get(ctx, "default-ttl-key")
		assert.False(t, found)
	})
}

func TestRedisStoreIteration(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	ctx := context.Background()

	t.Run("Keys without prefix", func(t *testing.T) {
		store := createRedisStore(client, redisOptions{})

		// Setup test data for this test only
		testData := map[string][]byte{
			"user:1":    []byte("1"),
			"user:2":    []byte("2"),
			"user:3":    []byte("3"),
			"product:a": []byte("10"),
			"product:b": []byte("20"),
			"config:x":  []byte("100"),
		}

		for key, value := range testData {
			err := store.Set(ctx, key, value)
			require.NoError(t, err)
		}

		keys, err := store.Keys(ctx, "")
		require.NoError(t, err)

		sort.Strings(keys)
		expectedKeys := []string{"config:x", "product:a", "product:b", "user:1", "user:2", "user:3"}
		assert.Equal(t, expectedKeys, keys)

		// Cleanup
		_ = store.Clear(ctx, "")
	})

	t.Run("Keys with prefix", func(t *testing.T) {
		store := createRedisStore(client, redisOptions{})

		// Setup test data for this test only
		testData := map[string][]byte{
			"user:1":    []byte("1"),
			"user:2":    []byte("2"),
			"user:3":    []byte("3"),
			"product:a": []byte("10"),
			"product:b": []byte("20"),
		}

		for key, value := range testData {
			err := store.Set(ctx, key, value)
			require.NoError(t, err)
		}

		userKeys, err := store.Keys(ctx, "user:")
		require.NoError(t, err)

		sort.Strings(userKeys)
		expectedUserKeys := []string{"user:1", "user:2", "user:3"}
		assert.Equal(t, expectedUserKeys, userKeys)

		productKeys, err := store.Keys(ctx, "product:")
		require.NoError(t, err)

		sort.Strings(productKeys)
		expectedProductKeys := []string{"product:a", "product:b"}
		assert.Equal(t, expectedProductKeys, productKeys)

		// Cleanup
		_ = store.Clear(ctx, "")
	})

	t.Run("ForEach without prefix", func(t *testing.T) {
		store := createRedisStore(client, redisOptions{})

		// Setup test data for this test only
		testData := map[string][]byte{
			"user:1":    []byte("1"),
			"user:2":    []byte("2"),
			"user:3":    []byte("3"),
			"product:a": []byte("10"),
			"product:b": []byte("20"),
			"config:x":  []byte("100"),
		}

		for key, value := range testData {
			err := store.Set(ctx, key, value)
			require.NoError(t, err)
		}

		collected := make(map[string][]byte)

		err := store.ForEach(ctx, "", func(key string, value []byte) bool {
			collected[key] = value
			return true
		})
		require.NoError(t, err)

		assert.Equal(t, testData, collected)

		// Cleanup
		_ = store.Clear(ctx, "")
	})

	t.Run("ForEach with prefix", func(t *testing.T) {
		store := createRedisStore(client, redisOptions{})

		// Setup test data for this test only
		testData := map[string][]byte{
			"user:1":    []byte("1"),
			"user:2":    []byte("2"),
			"user:3":    []byte("3"),
			"product:a": []byte("10"),
			"product:b": []byte("20"),
		}

		for key, value := range testData {
			err := store.Set(ctx, key, value)
			require.NoError(t, err)
		}

		userCollected := make(map[string][]byte)

		err := store.ForEach(ctx, "user:", func(key string, value []byte) bool {
			userCollected[key] = value
			return true
		})
		require.NoError(t, err)

		expectedUserData := map[string][]byte{
			"user:1": []byte("1"),
			"user:2": []byte("2"),
			"user:3": []byte("3"),
		}
		assert.Equal(t, expectedUserData, userCollected)

		// Cleanup
		_ = store.Clear(ctx, "")
	})

	t.Run("ForEach early termination", func(t *testing.T) {
		store := createRedisStore(client, redisOptions{})

		// Setup test data for this test only
		testData := map[string][]byte{
			"item1": []byte("value1"),
			"item2": []byte("value2"),
			"item3": []byte("value3"),
			"item4": []byte("value4"),
			"item5": []byte("value5"),
		}

		for key, value := range testData {
			err := store.Set(ctx, key, value)
			require.NoError(t, err)
		}

		collected := make(map[string][]byte)
		count := 0

		err := store.ForEach(ctx, "", func(key string, value []byte) bool {
			collected[key] = value
			count++
			return count < 3 // Stop after 3 items
		})
		require.NoError(t, err)

		assert.Equal(t, 3, count)
		assert.Len(t, collected, 3)

		// Cleanup
		_ = store.Clear(ctx, "")
	})

	t.Run("Size", func(t *testing.T) {
		store := createRedisStore(client, redisOptions{})

		// Setup test data for this test only
		testData := map[string][]byte{
			"item1": []byte("value1"),
			"item2": []byte("value2"),
			"item3": []byte("value3"),
			"item4": []byte("value4"),
			"item5": []byte("value5"),
			"item6": []byte("value6"),
		}

		for key, value := range testData {
			err := store.Set(ctx, key, value)
			require.NoError(t, err)
		}

		size, err := store.Size(ctx, "")
		require.NoError(t, err)
		assert.Equal(t, int64(6), size)

		// Cleanup
		_ = store.Clear(ctx, "")
	})
}

func TestRedisStoreClear(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	ctx := context.Background()

	t.Run("Clear all keys", func(t *testing.T) {
		store := createRedisStore(client, redisOptions{})

		// Add some test data
		for i := range 10 {
			key := fmt.Sprintf("clear-test-key-%d", i)
			value := fmt.Appendf(nil, "value-%d", i)
			err := store.Set(ctx, key, value)
			require.NoError(t, err)
		}

		// Verify data exists
		size, err := store.Size(ctx, "")
		require.NoError(t, err)
		assert.Equal(t, int64(10), size)

		// Clear cache
		err = store.Clear(ctx, "")
		require.NoError(t, err)

		// Verify cache is empty
		size, err = store.Size(ctx, "")
		require.NoError(t, err)
		assert.Equal(t, int64(0), size)
	})

	t.Run("Clear with prefix", func(t *testing.T) {
		store := createRedisStore(client, redisOptions{})

		// Add test data with different prefixes
		testData := map[string][]byte{
			"user:1":    []byte("user1"),
			"user:2":    []byte("user2"),
			"user:3":    []byte("user3"),
			"product:1": []byte("product1"),
			"product:2": []byte("product2"),
			"config:1":  []byte("config1"),
		}

		for key, value := range testData {
			err := store.Set(ctx, key, value)
			require.NoError(t, err)
		}

		// Verify all data exists
		size, err := store.Size(ctx, "")
		require.NoError(t, err)
		assert.Equal(t, int64(6), size)

		// Clear only user: prefixed keys
		err = store.Clear(ctx, "user:")
		require.NoError(t, err)

		// Verify user keys are gone
		userKeys, err := store.Keys(ctx, "user:")
		require.NoError(t, err)
		assert.Empty(t, userKeys)

		// Verify other keys still exist
		productKeys, err := store.Keys(ctx, "product:")
		require.NoError(t, err)
		assert.Len(t, productKeys, 2)

		configKeys, err := store.Keys(ctx, "config:")
		require.NoError(t, err)
		assert.Len(t, configKeys, 1)

		// Cleanup remaining keys
		_ = store.Clear(ctx, "")
	})
}

func TestRedisStorePrefixFiltering(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	ctx := context.Background()

	t.Run("Prefix filtering in operations", func(t *testing.T) {
		store := createRedisStore(client, redisOptions{})

		// Add test data with different prefixes
		testData := map[string][]byte{
			"app:user:1":    []byte("user1"),
			"app:user:2":    []byte("user2"),
			"app:product:1": []byte("product1"),
			"app:product:2": []byte("product2"),
			"other:key":     []byte("other"),
		}

		for key, value := range testData {
			err := store.Set(ctx, key, value)
			require.NoError(t, err)
		}

		// Test Keys with different prefixes
		appUserKeys, err := store.Keys(ctx, "app:user:")
		require.NoError(t, err)
		sort.Strings(appUserKeys)
		assert.Equal(t, []string{"app:user:1", "app:user:2"}, appUserKeys)

		appProductKeys, err := store.Keys(ctx, "app:product:")
		require.NoError(t, err)
		sort.Strings(appProductKeys)
		assert.Equal(t, []string{"app:product:1", "app:product:2"}, appProductKeys)

		allAppKeys, err := store.Keys(ctx, "app:")
		require.NoError(t, err)
		sort.Strings(allAppKeys)
		assert.Equal(t, []string{"app:product:1", "app:product:2", "app:user:1", "app:user:2"}, allAppKeys)

		otherKeys, err := store.Keys(ctx, "other:")
		require.NoError(t, err)
		assert.Equal(t, []string{"other:key"}, otherKeys)

		// Test Size with prefix filtering
		appUserSize, err := store.Size(ctx, "app:user:")
		require.NoError(t, err)
		assert.Equal(t, int64(2), appUserSize)

		appSize, err := store.Size(ctx, "app:")
		require.NoError(t, err)
		assert.Equal(t, int64(4), appSize)

		// Test Clear with prefix
		err = store.Clear(ctx, "app:user:")
		require.NoError(t, err)

		// Verify app:user: keys are gone
		remainingAppUserKeys, err := store.Keys(ctx, "app:user:")
		require.NoError(t, err)
		assert.Empty(t, remainingAppUserKeys)

		// Verify other app: keys still exist
		remainingAppKeys, err := store.Keys(ctx, "app:")
		require.NoError(t, err)
		assert.Len(t, remainingAppKeys, 2) // product keys should remain

		// Cleanup all remaining keys
		_ = store.Clear(ctx, "")
	})
}

func TestRedisStoreClose(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	store := createRedisStore(client, redisOptions{})

	ctx := context.Background()

	// Set some data
	err := store.Set(ctx, "test-key", []byte("test-value"))
	require.NoError(t, err)

	// Close should not error and should not close the underlying Redis client
	err = store.Close(ctx)
	assert.NoError(t, err)

	// Redis client should still be functional
	_, err = client.Ping(ctx).Result()
	assert.NoError(t, err)

	// Cleanup
	client.Del(ctx, "test-key")
}
