package cache

import (
	"context"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/ilxqx/vef-framework-go/cache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUser represents a test user struct for cache operations
type TestUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// setupRedisCache sets up a Redis-backed cache for testing
func setupRedisCache[T any](t *testing.T, cacheName string) cache.Cache[T] {
	// Skip Redis tests if REDIS_URL is not provided
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis cache tests")
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
		client.Close()
	}

	// Clean up any existing test keys for this cache
	store := createRedisStore(client, redisOptions{})
	testPattern := "vef:" + cacheName + ":*"
	keys, _ := client.Keys(ctx, testPattern).Result()
	if len(keys) > 0 {
		client.Del(ctx, keys...)
	}

	// Create cache instance
	testCache := cache.New[T](cacheName, store)

	// Add cleanup function
	t.Cleanup(func() {
		testCache.Clear(ctx)
		client.Close()
	})

	return testCache
}

func TestRedisCacheBasicOperations(t *testing.T) {
	userCache := setupRedisCache[TestUser](t, "test-users")
	ctx := context.Background()

	t.Run("Set and Get", func(t *testing.T) {
		user := TestUser{ID: 1, Name: "Alice", Age: 30}

		err := userCache.Set(ctx, "user1", user)
		require.NoError(t, err)

		result, found := userCache.Get(ctx, "user1")
		assert.True(t, found)
		assert.Equal(t, user, result)
	})

	t.Run("Contains", func(t *testing.T) {
		user := TestUser{ID: 2, Name: "Bob", Age: 25}

		err := userCache.Set(ctx, "user2", user)
		require.NoError(t, err)

		assert.True(t, userCache.Contains(ctx, "user2"))
		assert.False(t, userCache.Contains(ctx, "nonexistent"))
	})

	t.Run("Delete", func(t *testing.T) {
		user := TestUser{ID: 3, Name: "Charlie", Age: 35}

		err := userCache.Set(ctx, "user3", user)
		require.NoError(t, err)

		assert.True(t, userCache.Contains(ctx, "user3"))

		err = userCache.Delete(ctx, "user3")
		require.NoError(t, err)

		assert.False(t, userCache.Contains(ctx, "user3"))
		_, found := userCache.Get(ctx, "user3")
		assert.False(t, found)
	})

	t.Run("Update existing key", func(t *testing.T) {
		originalUser := TestUser{ID: 4, Name: "David", Age: 40}
		updatedUser := TestUser{ID: 4, Name: "David", Age: 41}

		err := userCache.Set(ctx, "user4", originalUser)
		require.NoError(t, err)

		result, found := userCache.Get(ctx, "user4")
		assert.True(t, found)
		assert.Equal(t, originalUser, result)

		err = userCache.Set(ctx, "user4", updatedUser)
		require.NoError(t, err)

		result, found = userCache.Get(ctx, "user4")
		assert.True(t, found)
		assert.Equal(t, updatedUser, result)
	})
}

func TestRedisCacheTTL(t *testing.T) {
	userCache := setupRedisCache[TestUser](t, "test-ttl-users")
	ctx := context.Background()

	t.Run("TTL expiration", func(t *testing.T) {
		user := TestUser{ID: 5, Name: "Eve", Age: 28}

		err := userCache.Set(ctx, "ttl-user", user, 100*time.Millisecond)
		require.NoError(t, err)

		// Should exist immediately
		result, found := userCache.Get(ctx, "ttl-user")
		assert.True(t, found)
		assert.Equal(t, user, result)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should be expired
		_, found = userCache.Get(ctx, "ttl-user")
		assert.False(t, found)
	})
}

func TestRedisCacheKeyPrefixIsolation(t *testing.T) {
	// Create two caches with different names
	cache1 := setupRedisCache[TestUser](t, "cache1")
	cache2 := setupRedisCache[TestUser](t, "cache2")
	ctx := context.Background()

	user1 := TestUser{ID: 1, Name: "Alice", Age: 30}
	user2 := TestUser{ID: 2, Name: "Bob", Age: 25}

	t.Run("Different cache instances should be isolated", func(t *testing.T) {
		// Set the same key in both caches with different values
		err := cache1.Set(ctx, "shared-key", user1)
		require.NoError(t, err)

		err = cache2.Set(ctx, "shared-key", user2)
		require.NoError(t, err)

		// Values should be different due to prefix isolation
		result1, found := cache1.Get(ctx, "shared-key")
		assert.True(t, found)
		assert.Equal(t, user1, result1)

		result2, found := cache2.Get(ctx, "shared-key")
		assert.True(t, found)
		assert.Equal(t, user2, result2)

		// Keys should be isolated - each cache should only see its own key
		keys1, err := cache1.Keys(ctx)
		require.NoError(t, err)
		assert.Contains(t, keys1, "vef:cache1:shared-key")
		assert.NotContains(t, keys1, "vef:cache2:shared-key")

		keys2, err := cache2.Keys(ctx)
		require.NoError(t, err)
		assert.Contains(t, keys2, "vef:cache2:shared-key")
		assert.NotContains(t, keys2, "vef:cache1:shared-key")
	})
}

func TestRedisCacheIteration(t *testing.T) {
	userCache := setupRedisCache[TestUser](t, "test-iteration")
	ctx := context.Background()

	// Setup test data
	testUsers := map[string]TestUser{
		"admin:1": {ID: 1, Name: "Admin Alice", Age: 35},
		"admin:2": {ID: 2, Name: "Admin Bob", Age: 40},
		"user:1":  {ID: 3, Name: "User Charlie", Age: 25},
		"user:2":  {ID: 4, Name: "User David", Age: 30},
		"guest:1": {ID: 5, Name: "Guest Eve", Age: 22},
	}

	for key, user := range testUsers {
		err := userCache.Set(ctx, key, user)
		require.NoError(t, err)
	}

	t.Run("Keys without prefix", func(t *testing.T) {
		keys, err := userCache.Keys(ctx)
		require.NoError(t, err)

		// Keys should include the full prefixed keys
		sort.Strings(keys)
		expectedKeys := []string{
			"vef:test-iteration:admin:1",
			"vef:test-iteration:admin:2",
			"vef:test-iteration:guest:1",
			"vef:test-iteration:user:1",
			"vef:test-iteration:user:2",
		}
		assert.Equal(t, expectedKeys, keys)
	})

	t.Run("Keys with prefix", func(t *testing.T) {
		adminKeys, err := userCache.Keys(ctx, "admin")
		require.NoError(t, err)

		sort.Strings(adminKeys)
		expectedAdminKeys := []string{
			"vef:test-iteration:admin:1",
			"vef:test-iteration:admin:2",
		}
		assert.Equal(t, expectedAdminKeys, adminKeys)

		userKeys, err := userCache.Keys(ctx, "user")
		require.NoError(t, err)

		sort.Strings(userKeys)
		expectedUserKeys := []string{
			"vef:test-iteration:user:1",
			"vef:test-iteration:user:2",
		}
		assert.Equal(t, expectedUserKeys, userKeys)
	})

	t.Run("ForEach without prefix", func(t *testing.T) {
		collected := make(map[string]TestUser)

		err := userCache.ForEach(ctx, func(key string, user TestUser) bool {
			collected[key] = user
			return true
		})
		require.NoError(t, err)

		// Should collect all users with prefixed keys
		expectedCollected := map[string]TestUser{
			"vef:test-iteration:admin:1": testUsers["admin:1"],
			"vef:test-iteration:admin:2": testUsers["admin:2"],
			"vef:test-iteration:guest:1": testUsers["guest:1"],
			"vef:test-iteration:user:1":  testUsers["user:1"],
			"vef:test-iteration:user:2":  testUsers["user:2"],
		}
		assert.Equal(t, expectedCollected, collected)
	})

	t.Run("ForEach with prefix", func(t *testing.T) {
		adminCollected := make(map[string]TestUser)

		err := userCache.ForEach(ctx, func(key string, user TestUser) bool {
			adminCollected[key] = user
			return true
		}, "admin")
		require.NoError(t, err)

		expectedAdminCollected := map[string]TestUser{
			"vef:test-iteration:admin:1": testUsers["admin:1"],
			"vef:test-iteration:admin:2": testUsers["admin:2"],
		}
		assert.Equal(t, expectedAdminCollected, adminCollected)
	})

	t.Run("ForEach early termination", func(t *testing.T) {
		collected := make(map[string]TestUser)
		count := 0

		err := userCache.ForEach(ctx, func(key string, user TestUser) bool {
			collected[key] = user
			count++
			return count < 3 // Stop after 3 items
		})
		require.NoError(t, err)

		assert.Equal(t, 3, count)
		assert.Len(t, collected, 3)
	})

	t.Run("Size", func(t *testing.T) {
		size, err := userCache.Size(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(5), size)
	})
}

func TestRedisCacheClear(t *testing.T) {
	// Create two separate caches to test isolation during clear
	cache1 := setupRedisCache[TestUser](t, "clear-test-1")
	cache2 := setupRedisCache[TestUser](t, "clear-test-2")
	ctx := context.Background()

	t.Run("Clear isolates between different cache instances", func(t *testing.T) {
		// Add data to both caches
		for i := 1; i <= 5; i++ {
			user := TestUser{ID: i, Name: fmt.Sprintf("User%d", i), Age: 20 + i}
			err := cache1.Set(ctx, fmt.Sprintf("user-%d", i), user)
			require.NoError(t, err)
		}

		user := TestUser{ID: 99, Name: "Other User", Age: 99}
		err := cache2.Set(ctx, "other-user", user)
		require.NoError(t, err)

		// Verify data exists in both caches
		size1, err := cache1.Size(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(5), size1)

		size2, err := cache2.Size(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), size2)

		// Clear cache1
		err = cache1.Clear(ctx)
		require.NoError(t, err)

		// Verify cache1 is empty
		size1, err = cache1.Size(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), size1)

		// Verify cache2 still has data
		otherUser, found := cache2.Get(ctx, "other-user")
		assert.True(t, found)
		assert.Equal(t, user, otherUser)

		size2, err = cache2.Size(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), size2)
	})
}

func TestRedisCacheStringCache(t *testing.T) {
	stringCache := setupRedisCache[string](t, "test-strings")
	ctx := context.Background()

	t.Run("String values", func(t *testing.T) {
		err := stringCache.Set(ctx, "greeting", "Hello, World!")
		require.NoError(t, err)

		result, found := stringCache.Get(ctx, "greeting")
		assert.True(t, found)
		assert.Equal(t, "Hello, World!", result)

		err = stringCache.Set(ctx, "farewell", "Goodbye!")
		require.NoError(t, err)

		keys, err := stringCache.Keys(ctx)
		require.NoError(t, err)
		assert.Len(t, keys, 2)
		assert.Contains(t, keys, "vef:test-strings:greeting")
		assert.Contains(t, keys, "vef:test-strings:farewell")
	})
}
