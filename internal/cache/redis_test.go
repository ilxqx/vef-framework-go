package cache

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/internal/redis"
	"github.com/ilxqx/vef-framework-go/testhelpers"
)

// RedisStoreTestSuite is the test suite for Redis store functionality.
type RedisStoreTestSuite struct {
	suite.Suite

	ctx            context.Context
	redisContainer *testhelpers.RedisContainer
	client         *goredis.Client
}

// SetupSuite runs before all tests in the suite.
func (suite *RedisStoreTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Start Redis container
	redisContainer := testhelpers.NewRedisContainer(suite.ctx, &suite.Suite)
	suite.redisContainer = redisContainer

	// Create Redis client
	suite.client = redis.NewClient(redisContainer.RdsConfig, &config.AppConfig{
		Name: "test-app",
	})

	// Test connection
	err := suite.client.Ping(suite.ctx).Err()
	suite.Require().NoError(err, "Failed to ping redis client")
}

// TearDownSuite runs after all tests in the suite.
func (suite *RedisStoreTestSuite) TearDownSuite() {
	if suite.client != nil {
		if err := suite.client.Close(); err != nil {
			suite.T().Logf("Failed to close redis client: %v", err)
		}
	}

	if suite.redisContainer != nil {
		suite.redisContainer.Terminate(suite.ctx, &suite.Suite)
	}
}

// SetupTest runs before each individual test method (not sub-tests).
func (suite *RedisStoreTestSuite) SetupTest() {
	// Clean up any existing test keys before each test method
	keys, _ := suite.client.Keys(suite.ctx, "*").Result()
	if len(keys) > 0 {
		suite.client.Del(suite.ctx, keys...)
	}
}

func (suite *RedisStoreTestSuite) TestRedisStoreBasicOperations() {
	store := NewRedisStore(suite.client, &config.RedisCacheConfig{
		DefaultTTL: 0, // No default TTL
	})

	suite.Run("Set and Get", func() {
		testData := []byte(`{"name":"test","value":42}`)

		err := store.Set(suite.ctx, "test-key", testData)
		suite.Require().NoError(err)

		result, found := store.Get(suite.ctx, "test-key")
		suite.True(found)
		suite.Equal(testData, result)
	})

	suite.Run("Contains", func() {
		testData := []byte(`{"name":"exists","value":1}`)

		err := store.Set(suite.ctx, "exists-key", testData)
		suite.Require().NoError(err)

		suite.True(store.Contains(suite.ctx, "exists-key"))
		suite.False(store.Contains(suite.ctx, "not-exists-key"))
	})

	suite.Run("Delete", func() {
		testData := []byte(`{"name":"delete","value":2}`)

		err := store.Set(suite.ctx, "delete-key", testData)
		suite.Require().NoError(err)

		suite.True(store.Contains(suite.ctx, "delete-key"))

		err = store.Delete(suite.ctx, "delete-key")
		suite.Require().NoError(err)

		suite.False(store.Contains(suite.ctx, "delete-key"))
	})

	suite.Run("Update existing key", func() {
		originalData := []byte(`{"name":"original","value":1}`)
		updatedData := []byte(`{"name":"updated","value":2}`)

		err := store.Set(suite.ctx, "update-key", originalData)
		suite.Require().NoError(err)

		result, found := store.Get(suite.ctx, "update-key")
		suite.True(found)
		suite.Equal(originalData, result)

		err = store.Set(suite.ctx, "update-key", updatedData)
		suite.Require().NoError(err)

		result, found = store.Get(suite.ctx, "update-key")
		suite.True(found)
		suite.Equal(updatedData, result)
	})
}

func (suite *RedisStoreTestSuite) TestRedisStoreTTL() {
	suite.Run("TTL expiration", func() {
		store := NewRedisStore(suite.client, &config.RedisCacheConfig{})

		err := store.Set(suite.ctx, "ttl-key", []byte("ttl-value"), 100*time.Millisecond)
		suite.Require().NoError(err)

		// Should exist immediately
		value, found := store.Get(suite.ctx, "ttl-key")
		suite.True(found)
		suite.Equal([]byte("ttl-value"), value)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should be expired
		_, found = store.Get(suite.ctx, "ttl-key")
		suite.False(found)
	})

	suite.Run("Default TTL", func() {
		store := NewRedisStore(suite.client, &config.RedisCacheConfig{
			DefaultTTL: 100 * time.Millisecond,
		})

		err := store.Set(suite.ctx, "default-ttl-key", []byte("default-ttl-value"))
		suite.Require().NoError(err)

		// Should exist immediately
		value, found := store.Get(suite.ctx, "default-ttl-key")
		suite.True(found)
		suite.Equal([]byte("default-ttl-value"), value)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should be expired
		_, found = store.Get(suite.ctx, "default-ttl-key")
		suite.False(found)
	})
}

func (suite *RedisStoreTestSuite) TestRedisStoreIteration() {
	suite.Run("Keys without prefix", func() {
		// Clean up before this sub-test
		keys, _ := suite.client.Keys(suite.ctx, "*").Result()
		if len(keys) > 0 {
			suite.client.Del(suite.ctx, keys...)
		}

		store := NewRedisStore(suite.client, &config.RedisCacheConfig{})

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
			err := store.Set(suite.ctx, key, value)
			suite.Require().NoError(err)
		}

		keys, err := store.Keys(suite.ctx, "")
		suite.Require().NoError(err)

		sort.Strings(keys)

		expectedKeys := []string{"config:x", "product:a", "product:b", "user:1", "user:2", "user:3"}
		suite.Equal(expectedKeys, keys)
	})

	suite.Run("Keys with prefix", func() {
		// Clean up before this sub-test
		keys, _ := suite.client.Keys(suite.ctx, "*").Result()
		if len(keys) > 0 {
			suite.client.Del(suite.ctx, keys...)
		}

		store := NewRedisStore(suite.client, &config.RedisCacheConfig{})

		// Setup test data for this test only
		testData := map[string][]byte{
			"user:1":    []byte("1"),
			"user:2":    []byte("2"),
			"user:3":    []byte("3"),
			"product:a": []byte("10"),
			"product:b": []byte("20"),
		}

		for key, value := range testData {
			err := store.Set(suite.ctx, key, value)
			suite.Require().NoError(err)
		}

		userKeys, err := store.Keys(suite.ctx, "user:")
		suite.Require().NoError(err)

		sort.Strings(userKeys)

		expectedUserKeys := []string{"user:1", "user:2", "user:3"}
		suite.Equal(expectedUserKeys, userKeys)

		productKeys, err := store.Keys(suite.ctx, "product:")
		suite.Require().NoError(err)

		sort.Strings(productKeys)

		expectedProductKeys := []string{"product:a", "product:b"}
		suite.Equal(expectedProductKeys, productKeys)
	})

	suite.Run("ForEach without prefix", func() {
		// Clean up before this sub-test
		keys, _ := suite.client.Keys(suite.ctx, "*").Result()
		if len(keys) > 0 {
			suite.client.Del(suite.ctx, keys...)
		}

		store := NewRedisStore(suite.client, &config.RedisCacheConfig{})

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
			err := store.Set(suite.ctx, key, value)
			suite.Require().NoError(err)
		}

		collected := make(map[string][]byte)

		err := store.ForEach(suite.ctx, "", func(key string, value []byte) bool {
			collected[key] = value

			return true
		})
		suite.Require().NoError(err)

		suite.Equal(testData, collected)
	})

	suite.Run("ForEach with prefix", func() {
		// Clean up before this sub-test
		keys, _ := suite.client.Keys(suite.ctx, "*").Result()
		if len(keys) > 0 {
			suite.client.Del(suite.ctx, keys...)
		}

		store := NewRedisStore(suite.client, &config.RedisCacheConfig{})

		// Setup test data for this test only
		testData := map[string][]byte{
			"user:1":    []byte("1"),
			"user:2":    []byte("2"),
			"user:3":    []byte("3"),
			"product:a": []byte("10"),
			"product:b": []byte("20"),
		}

		for key, value := range testData {
			err := store.Set(suite.ctx, key, value)
			suite.Require().NoError(err)
		}

		userCollected := make(map[string][]byte)

		err := store.ForEach(suite.ctx, "user:", func(key string, value []byte) bool {
			userCollected[key] = value

			return true
		})
		suite.Require().NoError(err)

		expectedUserData := map[string][]byte{
			"user:1": []byte("1"),
			"user:2": []byte("2"),
			"user:3": []byte("3"),
		}
		suite.Equal(expectedUserData, userCollected)
	})

	suite.Run("ForEach early termination", func() {
		// Clean up before this sub-test
		keys, _ := suite.client.Keys(suite.ctx, "*").Result()
		if len(keys) > 0 {
			suite.client.Del(suite.ctx, keys...)
		}

		store := NewRedisStore(suite.client, &config.RedisCacheConfig{})

		// Setup test data for this test only
		testData := map[string][]byte{
			"item1": []byte("value1"),
			"item2": []byte("value2"),
			"item3": []byte("value3"),
			"item4": []byte("value4"),
			"item5": []byte("value5"),
		}

		for key, value := range testData {
			err := store.Set(suite.ctx, key, value)
			suite.Require().NoError(err)
		}

		collected := make(map[string][]byte)
		count := 0

		err := store.ForEach(suite.ctx, "", func(key string, value []byte) bool {
			collected[key] = value
			count++

			return count < 3 // Stop after 3 items
		})
		suite.Require().NoError(err)

		suite.Equal(3, count)
		suite.Len(collected, 3)
	})

	suite.Run("Size", func() {
		// Clean up before this sub-test
		keys, _ := suite.client.Keys(suite.ctx, "*").Result()
		if len(keys) > 0 {
			suite.client.Del(suite.ctx, keys...)
		}

		store := NewRedisStore(suite.client, &config.RedisCacheConfig{})

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
			err := store.Set(suite.ctx, key, value)
			suite.Require().NoError(err)
		}

		size, err := store.Size(suite.ctx, "")
		suite.Require().NoError(err)
		suite.Equal(int64(6), size)
	})
}

func (suite *RedisStoreTestSuite) TestRedisStoreClear() {
	suite.Run("Clear all keys", func() {
		store := NewRedisStore(suite.client, &config.RedisCacheConfig{})

		// Add some test data
		for i := range 10 {
			key := fmt.Sprintf("clear-test-key-%d", i)
			value := fmt.Appendf(nil, "value-%d", i)
			err := store.Set(suite.ctx, key, value)
			suite.Require().NoError(err)
		}

		// Verify data exists
		size, err := store.Size(suite.ctx, "")
		suite.Require().NoError(err)
		suite.Equal(int64(10), size)

		// Clear cache
		err = store.Clear(suite.ctx, "")
		suite.Require().NoError(err)

		// Verify cache is empty
		size, err = store.Size(suite.ctx, "")
		suite.Require().NoError(err)
		suite.Equal(int64(0), size)
	})

	suite.Run("Clear with prefix", func() {
		store := NewRedisStore(suite.client, &config.RedisCacheConfig{})

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
			err := store.Set(suite.ctx, key, value)
			suite.Require().NoError(err)
		}

		// Verify all data exists
		size, err := store.Size(suite.ctx, "")
		suite.Require().NoError(err)
		suite.Equal(int64(6), size)

		// Clear only user: prefixed keys
		err = store.Clear(suite.ctx, "user:")
		suite.Require().NoError(err)

		// Verify user keys are gone
		userKeys, err := store.Keys(suite.ctx, "user:")
		suite.Require().NoError(err)
		suite.Empty(userKeys)

		// Verify other keys still exist
		productKeys, err := store.Keys(suite.ctx, "product:")
		suite.Require().NoError(err)
		suite.Len(productKeys, 2)

		configKeys, err := store.Keys(suite.ctx, "config:")
		suite.Require().NoError(err)
		suite.Len(configKeys, 1)
	})
}

func (suite *RedisStoreTestSuite) TestRedisStorePrefixFiltering() {
	suite.Run("Prefix filtering in operations", func() {
		store := NewRedisStore(suite.client, &config.RedisCacheConfig{})

		// Add test data with different prefixes
		testData := map[string][]byte{
			"app:user:1":    []byte("user1"),
			"app:user:2":    []byte("user2"),
			"app:product:1": []byte("product1"),
			"app:product:2": []byte("product2"),
			"other:key":     []byte("other"),
		}

		for key, value := range testData {
			err := store.Set(suite.ctx, key, value)
			suite.Require().NoError(err)
		}

		// Test Keys with different prefixes
		appUserKeys, err := store.Keys(suite.ctx, "app:user:")
		suite.Require().NoError(err)
		sort.Strings(appUserKeys)
		suite.Equal([]string{"app:user:1", "app:user:2"}, appUserKeys)

		appProductKeys, err := store.Keys(suite.ctx, "app:product:")
		suite.Require().NoError(err)
		sort.Strings(appProductKeys)
		suite.Equal([]string{"app:product:1", "app:product:2"}, appProductKeys)

		allAppKeys, err := store.Keys(suite.ctx, "app:")
		suite.Require().NoError(err)
		sort.Strings(allAppKeys)
		suite.Equal([]string{"app:product:1", "app:product:2", "app:user:1", "app:user:2"}, allAppKeys)

		otherKeys, err := store.Keys(suite.ctx, "other:")
		suite.Require().NoError(err)
		suite.Equal([]string{"other:key"}, otherKeys)

		// Test Size with prefix filtering
		appUserSize, err := store.Size(suite.ctx, "app:user:")
		suite.Require().NoError(err)
		suite.Equal(int64(2), appUserSize)

		appSize, err := store.Size(suite.ctx, "app:")
		suite.Require().NoError(err)
		suite.Equal(int64(4), appSize)

		// Test Clear with prefix
		err = store.Clear(suite.ctx, "app:user:")
		suite.Require().NoError(err)

		// Verify app:user: keys are gone
		remainingAppUserKeys, err := store.Keys(suite.ctx, "app:user:")
		suite.Require().NoError(err)
		suite.Empty(remainingAppUserKeys)

		// Verify other app: keys still exist
		remainingAppKeys, err := store.Keys(suite.ctx, "app:")
		suite.Require().NoError(err)
		suite.Len(remainingAppKeys, 2) // product keys should remain
	})
}

func (suite *RedisStoreTestSuite) TestRedisStoreClose() {
	store := NewRedisStore(suite.client, &config.RedisCacheConfig{})

	// Set some data
	err := store.Set(suite.ctx, "test-key", []byte("test-value"))
	suite.Require().NoError(err)

	// Close should not error and should not close the underlying Redis client
	err = store.Close(suite.ctx)
	suite.NoError(err)

	// Redis client should still be functional
	_, err = suite.client.Ping(suite.ctx).Result()
	suite.NoError(err)

	// Cleanup
	suite.client.Del(suite.ctx, "test-key")
}

func (st *RedisStoreTestSuite) TestRedisCacheSuite() {
	suite.Run(st.T(), &RedisCacheTestSuite{
		ctx:    st.ctx,
		client: st.client,
	})
}

// TestRedisStoreSuite runs the test suite.
func TestRedisStoreSuite(t *testing.T) {
	suite.Run(t, new(RedisStoreTestSuite))
}
