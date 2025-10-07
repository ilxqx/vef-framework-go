package cache

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"

	"github.com/ilxqx/vef-framework-go/cache"
	"github.com/ilxqx/vef-framework-go/config"
)

// TestUser represents a test user struct for cache operations.
type TestUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// RedisCacheTestSuite is the test suite for Redis cache functionality.
type RedisCacheTestSuite struct {
	suite.Suite

	ctx    context.Context
	client *redis.Client
}

// SetupTest runs before each individual test.
func (suite *RedisCacheTestSuite) SetupTest() {
	// Clean up any existing test keys before each test
	keys, _ := suite.client.Keys(suite.ctx, "*").Result()
	if len(keys) > 0 {
		suite.client.Del(suite.ctx, keys...)
	}
}

// setupRedisCache creates a Redis-backed cache for testing.
func (suite *RedisCacheTestSuite) setupRedisCache(cacheName string) cache.Cache[TestUser] {
	store := NewRedisStore(suite.client, &config.RedisCacheConfig{})

	return cache.New[TestUser](cacheName, store)
}

// setupStringCache creates a Redis-backed string cache for testing.
func (suite *RedisCacheTestSuite) setupStringCache(cacheName string) cache.Cache[string] {
	store := NewRedisStore(suite.client, &config.RedisCacheConfig{})

	return cache.New[string](cacheName, store)
}

func (suite *RedisCacheTestSuite) TestRedisCacheBasicOperations() {
	userCache := suite.setupRedisCache("test-users")

	suite.Run("Set and Get", func() {
		user := TestUser{ID: 1, Name: "Alice", Age: 30}

		err := userCache.Set(suite.ctx, "user1", user)
		suite.Require().NoError(err)

		result, found := userCache.Get(suite.ctx, "user1")
		suite.True(found)
		suite.Equal(user, result)
	})

	suite.Run("Contains", func() {
		user := TestUser{ID: 2, Name: "Bob", Age: 25}

		err := userCache.Set(suite.ctx, "user2", user)
		suite.Require().NoError(err)

		suite.True(userCache.Contains(suite.ctx, "user2"))
		suite.False(userCache.Contains(suite.ctx, "nonexistent"))
	})

	suite.Run("Delete", func() {
		user := TestUser{ID: 3, Name: "Charlie", Age: 35}

		err := userCache.Set(suite.ctx, "user3", user)
		suite.Require().NoError(err)

		suite.True(userCache.Contains(suite.ctx, "user3"))

		err = userCache.Delete(suite.ctx, "user3")
		suite.Require().NoError(err)

		suite.False(userCache.Contains(suite.ctx, "user3"))
		_, found := userCache.Get(suite.ctx, "user3")
		suite.False(found)
	})

	suite.Run("Update existing key", func() {
		originalUser := TestUser{ID: 4, Name: "David", Age: 40}
		updatedUser := TestUser{ID: 4, Name: "David", Age: 41}

		err := userCache.Set(suite.ctx, "user4", originalUser)
		suite.Require().NoError(err)

		result, found := userCache.Get(suite.ctx, "user4")
		suite.True(found)
		suite.Equal(originalUser, result)

		err = userCache.Set(suite.ctx, "user4", updatedUser)
		suite.Require().NoError(err)

		result, found = userCache.Get(suite.ctx, "user4")
		suite.True(found)
		suite.Equal(updatedUser, result)
	})
}

func (suite *RedisCacheTestSuite) TestRedisCacheTTL() {
	userCache := suite.setupRedisCache("test-ttl-users")

	suite.Run("TTL expiration", func() {
		user := TestUser{ID: 5, Name: "Eve", Age: 28}

		err := userCache.Set(suite.ctx, "ttl-user", user, 100*time.Millisecond)
		suite.Require().NoError(err)

		// Should exist immediately
		result, found := userCache.Get(suite.ctx, "ttl-user")
		suite.True(found)
		suite.Equal(user, result)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should be expired
		_, found = userCache.Get(suite.ctx, "ttl-user")
		suite.False(found)
	})
}

func (suite *RedisCacheTestSuite) TestRedisCacheKeyPrefixIsolation() {
	// Create two caches with different names
	cache1 := suite.setupRedisCache("cache1")
	cache2 := suite.setupRedisCache("cache2")

	user1 := TestUser{ID: 1, Name: "Alice", Age: 30}
	user2 := TestUser{ID: 2, Name: "Bob", Age: 25}

	suite.Run("Different cache instances should be isolated", func() {
		// Set the same key in both caches with different values
		err := cache1.Set(suite.ctx, "shared-key", user1)
		suite.Require().NoError(err)

		err = cache2.Set(suite.ctx, "shared-key", user2)
		suite.Require().NoError(err)

		// Values should be different due to prefix isolation
		result1, found := cache1.Get(suite.ctx, "shared-key")
		suite.True(found)
		suite.Equal(user1, result1)

		result2, found := cache2.Get(suite.ctx, "shared-key")
		suite.True(found)
		suite.Equal(user2, result2)

		// Keys should be isolated - each cache should only see its own key
		keys1, err := cache1.Keys(suite.ctx)
		suite.Require().NoError(err)
		suite.Contains(keys1, "vef:cache1:shared-key")
		suite.NotContains(keys1, "vef:cache2:shared-key")

		keys2, err := cache2.Keys(suite.ctx)
		suite.Require().NoError(err)
		suite.Contains(keys2, "vef:cache2:shared-key")
		suite.NotContains(keys2, "vef:cache1:shared-key")
	})
}

func (suite *RedisCacheTestSuite) TestRedisCacheIteration() {
	userCache := suite.setupRedisCache("test-iteration")

	// Setup test data
	testUsers := map[string]TestUser{
		"admin:1": {ID: 1, Name: "Admin Alice", Age: 35},
		"admin:2": {ID: 2, Name: "Admin Bob", Age: 40},
		"user:1":  {ID: 3, Name: "User Charlie", Age: 25},
		"user:2":  {ID: 4, Name: "User David", Age: 30},
		"guest:1": {ID: 5, Name: "Guest Eve", Age: 22},
	}

	for key, user := range testUsers {
		err := userCache.Set(suite.ctx, key, user)
		suite.Require().NoError(err)
	}

	suite.Run("Keys without prefix", func() {
		keys, err := userCache.Keys(suite.ctx)
		suite.Require().NoError(err)

		// Keys should include the full prefixed keys
		sort.Strings(keys)

		expectedKeys := []string{
			"vef:test-iteration:admin:1",
			"vef:test-iteration:admin:2",
			"vef:test-iteration:guest:1",
			"vef:test-iteration:user:1",
			"vef:test-iteration:user:2",
		}
		suite.Equal(expectedKeys, keys)
	})

	suite.Run("Keys with prefix", func() {
		adminKeys, err := userCache.Keys(suite.ctx, "admin")
		suite.Require().NoError(err)

		sort.Strings(adminKeys)

		expectedAdminKeys := []string{
			"vef:test-iteration:admin:1",
			"vef:test-iteration:admin:2",
		}
		suite.Equal(expectedAdminKeys, adminKeys)

		userKeys, err := userCache.Keys(suite.ctx, "user")
		suite.Require().NoError(err)

		sort.Strings(userKeys)

		expectedUserKeys := []string{
			"vef:test-iteration:user:1",
			"vef:test-iteration:user:2",
		}
		suite.Equal(expectedUserKeys, userKeys)
	})

	suite.Run("ForEach without prefix", func() {
		collected := make(map[string]TestUser)

		err := userCache.ForEach(suite.ctx, func(key string, user TestUser) bool {
			collected[key] = user

			return true
		})
		suite.Require().NoError(err)

		// Should collect all users with prefixed keys
		expectedCollected := map[string]TestUser{
			"vef:test-iteration:admin:1": testUsers["admin:1"],
			"vef:test-iteration:admin:2": testUsers["admin:2"],
			"vef:test-iteration:guest:1": testUsers["guest:1"],
			"vef:test-iteration:user:1":  testUsers["user:1"],
			"vef:test-iteration:user:2":  testUsers["user:2"],
		}
		suite.Equal(expectedCollected, collected)
	})

	suite.Run("ForEach with prefix", func() {
		adminCollected := make(map[string]TestUser)

		err := userCache.ForEach(suite.ctx, func(key string, user TestUser) bool {
			adminCollected[key] = user

			return true
		}, "admin")
		suite.Require().NoError(err)

		expectedAdminCollected := map[string]TestUser{
			"vef:test-iteration:admin:1": testUsers["admin:1"],
			"vef:test-iteration:admin:2": testUsers["admin:2"],
		}
		suite.Equal(expectedAdminCollected, adminCollected)
	})

	suite.Run("ForEach early termination", func() {
		collected := make(map[string]TestUser)
		count := 0

		err := userCache.ForEach(suite.ctx, func(key string, user TestUser) bool {
			collected[key] = user
			count++

			return count < 3 // Stop after 3 items
		})
		suite.Require().NoError(err)

		suite.Equal(3, count)
		suite.Len(collected, 3)
	})

	suite.Run("Size", func() {
		size, err := userCache.Size(suite.ctx)
		suite.Require().NoError(err)
		suite.Equal(int64(5), size)
	})
}

func (suite *RedisCacheTestSuite) TestRedisCacheClear() {
	// Create two separate caches to test isolation during clear
	cache1 := suite.setupRedisCache("clear-test-1")
	cache2 := suite.setupRedisCache("clear-test-2")

	suite.Run("Clear isolates between different cache instances", func() {
		// Add data to both caches
		for i := 1; i <= 5; i++ {
			user := TestUser{ID: i, Name: fmt.Sprintf("User%d", i), Age: 20 + i}
			err := cache1.Set(suite.ctx, fmt.Sprintf("user-%d", i), user)
			suite.Require().NoError(err)
		}

		user := TestUser{ID: 99, Name: "Other User", Age: 99}
		err := cache2.Set(suite.ctx, "other-user", user)
		suite.Require().NoError(err)

		// Verify data exists in both caches
		size1, err := cache1.Size(suite.ctx)
		suite.Require().NoError(err)
		suite.Equal(int64(5), size1)

		size2, err := cache2.Size(suite.ctx)
		suite.Require().NoError(err)
		suite.Equal(int64(1), size2)

		// Clear cache1
		err = cache1.Clear(suite.ctx)
		suite.Require().NoError(err)

		// Verify cache1 is empty
		size1, err = cache1.Size(suite.ctx)
		suite.Require().NoError(err)
		suite.Equal(int64(0), size1)

		// Verify cache2 still has data
		otherUser, found := cache2.Get(suite.ctx, "other-user")
		suite.True(found)
		suite.Equal(user, otherUser)

		size2, err = cache2.Size(suite.ctx)
		suite.Require().NoError(err)
		suite.Equal(int64(1), size2)
	})
}

func (suite *RedisCacheTestSuite) TestRedisCacheStringCache() {
	stringCache := suite.setupStringCache("test-strings")

	suite.Run("String values", func() {
		err := stringCache.Set(suite.ctx, "greeting", "Hello, World!")
		suite.Require().NoError(err)

		result, found := stringCache.Get(suite.ctx, "greeting")
		suite.True(found)
		suite.Equal("Hello, World!", result)

		err = stringCache.Set(suite.ctx, "farewell", "Goodbye!")
		suite.Require().NoError(err)

		keys, err := stringCache.Keys(suite.ctx)
		suite.Require().NoError(err)
		suite.Len(keys, 2)
		suite.Contains(keys, "vef:test-strings:greeting")
		suite.Contains(keys, "vef:test-strings:farewell")
	})
}
