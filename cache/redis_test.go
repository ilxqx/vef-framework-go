package cache

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/internal/redis"
	"github.com/ilxqx/vef-framework-go/testhelpers"
)

type TestUser struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type RedisCacheTestSuite struct {
	suite.Suite

	ctx            context.Context
	redisContainer *testhelpers.RedisContainer
	client         *goredis.Client
}

func (suite *RedisCacheTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	container := testhelpers.NewRedisContainer(suite.ctx, &suite.Suite)
	suite.redisContainer = container

	suite.client = redis.NewClient(container.RdsConfig, &config.AppConfig{Name: "test-app"})
	err := suite.client.Ping(suite.ctx).Err()
	suite.Require().NoError(err, "failed to ping redis client")
}

func (suite *RedisCacheTestSuite) TearDownSuite() {
	if suite.client != nil {
		if err := suite.client.Close(); err != nil {
			suite.T().Logf("failed to close redis client: %v", err)
		}
	}

	if suite.redisContainer != nil {
		suite.redisContainer.Terminate(suite.ctx, &suite.Suite)
	}
}

func (suite *RedisCacheTestSuite) SetupTest() {
	keys, _ := suite.client.Keys(suite.ctx, "*").Result()
	if len(keys) > 0 {
		suite.client.Del(suite.ctx, keys...)
	}
}

func (suite *RedisCacheTestSuite) setupRedisCache(namespace string, opts ...RedisOption) Cache[TestUser] {
	c := NewRedis[TestUser](suite.client, namespace, opts...)
	suite.Require().NotNil(c)

	return c
}

func (suite *RedisCacheTestSuite) setupStringCache(namespace string) Cache[string] {
	c := NewRedis[string](suite.client, namespace)
	suite.Require().NotNil(c)

	return c
}

func (suite *RedisCacheTestSuite) TestRedisCacheBasicOperations() {
	userCache := suite.setupRedisCache("test-users")
	defer userCache.Close()

	suite.Run("SetAndGet", func() {
		user := TestUser{Id: 1, Name: "Alice", Age: 30}

		err := userCache.Set(suite.ctx, "user1", user)
		suite.Require().NoError(err)

		result, found := userCache.Get(suite.ctx, "user1")
		suite.True(found)
		suite.Equal(user, result)
	})

	suite.Run("Contains", func() {
		user := TestUser{Id: 2, Name: "Bob", Age: 25}

		err := userCache.Set(suite.ctx, "user2", user)
		suite.Require().NoError(err)

		suite.True(userCache.Contains(suite.ctx, "user2"))
		suite.False(userCache.Contains(suite.ctx, "nonexistent"))
	})

	suite.Run("Delete", func() {
		user := TestUser{Id: 3, Name: "Charlie", Age: 35}

		err := userCache.Set(suite.ctx, "user3", user)
		suite.Require().NoError(err)

		suite.True(userCache.Contains(suite.ctx, "user3"))

		err = userCache.Delete(suite.ctx, "user3")
		suite.Require().NoError(err)

		suite.False(userCache.Contains(suite.ctx, "user3"))
		_, found := userCache.Get(suite.ctx, "user3")
		suite.False(found)
	})

	suite.Run("UpdateExistingKey", func() {
		originalUser := TestUser{Id: 4, Name: "David", Age: 40}
		updatedUser := TestUser{Id: 4, Name: "David", Age: 41}

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
	defer userCache.Close()

	suite.Run("CustomTTLExpiration", func() {
		user := TestUser{Id: 5, Name: "Eve", Age: 28}

		err := userCache.Set(suite.ctx, "ttl-user", user, 100*time.Millisecond)
		suite.Require().NoError(err)

		result, found := userCache.Get(suite.ctx, "ttl-user")
		suite.True(found)
		suite.Equal(user, result)

		time.Sleep(150 * time.Millisecond)

		_, found = userCache.Get(suite.ctx, "ttl-user")
		suite.False(found)
	})

	suite.Run("DefaultTTL", func() {
		cacheWithDefaultTTL := suite.setupRedisCache("test-default-ttl", WithRdsDefaultTTL(100*time.Millisecond))
		defer cacheWithDefaultTTL.Close()

		user := TestUser{Id: 6, Name: "Frank", Age: 32}

		err := cacheWithDefaultTTL.Set(suite.ctx, "default-ttl-user", user)
		suite.Require().NoError(err)

		result, found := cacheWithDefaultTTL.Get(suite.ctx, "default-ttl-user")
		suite.True(found)
		suite.Equal(user, result)

		time.Sleep(150 * time.Millisecond)

		_, found = cacheWithDefaultTTL.Get(suite.ctx, "default-ttl-user")
		suite.False(found)
	})
}

func (suite *RedisCacheTestSuite) TestRedisCacheGetOrLoad() {
	userCache := suite.setupRedisCache("test-getorload")
	defer userCache.Close()

	user := TestUser{Id: 7, Name: "Grace", Age: 26}

	var loadCount atomic.Int32

	loader := func(context.Context) (TestUser, error) {
		loadCount.Add(1)

		return user, nil
	}

	suite.Run("SingleLoad", func() {
		result, err := userCache.GetOrLoad(suite.ctx, "user7", loader)
		suite.Require().NoError(err)
		suite.Equal(user, result)
		suite.Equal(int32(1), loadCount.Load())
	})

	suite.Run("CachedValue", func() {
		result, err := userCache.GetOrLoad(suite.ctx, "user7", loader)
		suite.Require().NoError(err)
		suite.Equal(user, result)
		suite.Equal(int32(1), loadCount.Load(), "loader should not be invoked again for cached value")
	})

	suite.Run("ConcurrentRequests", func() {
		loadCount.Store(0)

		var wg sync.WaitGroup

		const goroutines = 20

		for range goroutines {
			wg.Go(func() {
				_, err := userCache.GetOrLoad(suite.ctx, "concurrent", loader)
				suite.Require().NoError(err)
			})
		}

		wg.Wait()
		suite.Equal(int32(1), loadCount.Load(), "loader should execute exactly once under contention")
	})
}

func (suite *RedisCacheTestSuite) TestRedisCacheKeyPrefixIsolation() {
	cache1 := suite.setupRedisCache("cache1")
	defer cache1.Close()

	cache2 := suite.setupRedisCache("cache2")
	defer cache2.Close()

	user1 := TestUser{Id: 1, Name: "Alice", Age: 30}
	user2 := TestUser{Id: 2, Name: "Bob", Age: 25}

	err := cache1.Set(suite.ctx, "shared-key", user1)
	suite.Require().NoError(err)

	err = cache2.Set(suite.ctx, "shared-key", user2)
	suite.Require().NoError(err)

	result1, found := cache1.Get(suite.ctx, "shared-key")
	suite.True(found)
	suite.Equal(user1, result1)

	result2, found := cache2.Get(suite.ctx, "shared-key")
	suite.True(found)
	suite.Equal(user2, result2)

	keys1, err := cache1.Keys(suite.ctx)
	suite.Require().NoError(err)
	suite.Contains(keys1, "shared-key")
	suite.Len(keys1, 1)

	keys2, err := cache2.Keys(suite.ctx)
	suite.Require().NoError(err)
	suite.Contains(keys2, "shared-key")
	suite.Len(keys2, 1)
}

func (suite *RedisCacheTestSuite) TestRedisCacheIteration() {
	userCache := suite.setupRedisCache("test-iteration")
	defer userCache.Close()

	testUsers := map[string]TestUser{
		"admin:1": {Id: 1, Name: "Admin Alice", Age: 35},
		"admin:2": {Id: 2, Name: "Admin Bob", Age: 40},
		"user:1":  {Id: 3, Name: "User Charlie", Age: 25},
		"user:2":  {Id: 4, Name: "User David", Age: 30},
		"guest:1": {Id: 5, Name: "Guest Eve", Age: 22},
	}

	for key, user := range testUsers {
		err := userCache.Set(suite.ctx, key, user)
		suite.Require().NoError(err)
	}

	suite.Run("KeysWithoutPrefix", func() {
		keys, err := userCache.Keys(suite.ctx)
		suite.Require().NoError(err)

		sort.Strings(keys)

		expectedKeys := []string{
			"admin:1",
			"admin:2",
			"guest:1",
			"user:1",
			"user:2",
		}
		suite.Equal(expectedKeys, keys)
	})

	suite.Run("KeysWithPrefix", func() {
		adminKeys, err := userCache.Keys(suite.ctx, "admin")
		suite.Require().NoError(err)
		sort.Strings(adminKeys)

		expectedAdmin := []string{
			"admin:1",
			"admin:2",
		}
		suite.Equal(expectedAdmin, adminKeys)

		userKeys, err := userCache.Keys(suite.ctx, "user")
		suite.Require().NoError(err)
		sort.Strings(userKeys)

		expectedUser := []string{
			"user:1",
			"user:2",
		}
		suite.Equal(expectedUser, userKeys)
	})

	suite.Run("ForEachWithoutPrefix", func() {
		collected := make(map[string]TestUser)

		err := userCache.ForEach(suite.ctx, func(key string, user TestUser) bool {
			collected[key] = user

			return true
		})
		suite.Require().NoError(err)

		expected := map[string]TestUser{
			"admin:1": testUsers["admin:1"],
			"admin:2": testUsers["admin:2"],
			"guest:1": testUsers["guest:1"],
			"user:1":  testUsers["user:1"],
			"user:2":  testUsers["user:2"],
		}
		suite.Equal(expected, collected)
	})

	suite.Run("ForEachWithPrefix", func() {
		collected := make(map[string]TestUser)

		err := userCache.ForEach(suite.ctx, func(key string, user TestUser) bool {
			collected[key] = user

			return true
		}, "admin")
		suite.Require().NoError(err)

		expected := map[string]TestUser{
			"admin:1": testUsers["admin:1"],
			"admin:2": testUsers["admin:2"],
		}
		suite.Equal(expected, collected)
	})

	suite.Run("ForEachEarlyTermination", func() {
		var count int

		err := userCache.ForEach(suite.ctx, func(key string, user TestUser) bool {
			count++

			return count < 3
		})
		suite.Require().NoError(err)
		suite.Equal(3, count)
	})

	suite.Run("Size", func() {
		size, err := userCache.Size(suite.ctx)
		suite.Require().NoError(err)
		suite.Equal(int64(5), size)
	})
}

func (suite *RedisCacheTestSuite) TestRedisCacheClear() {
	cache1 := suite.setupRedisCache("clear-test-1")
	defer cache1.Close()

	cache2 := suite.setupRedisCache("clear-test-2")
	defer cache2.Close()

	for i := 1; i <= 5; i++ {
		user := TestUser{Id: i, Name: fmt.Sprintf("User%d", i), Age: 20 + i}
		err := cache1.Set(suite.ctx, fmt.Sprintf("user-%d", i), user)
		suite.Require().NoError(err)
	}

	user := TestUser{Id: 99, Name: "Other User", Age: 99}
	err := cache2.Set(suite.ctx, "other-user", user)
	suite.Require().NoError(err)

	size1, err := cache1.Size(suite.ctx)
	suite.Require().NoError(err)
	suite.Equal(int64(5), size1)

	size2, err := cache2.Size(suite.ctx)
	suite.Require().NoError(err)
	suite.Equal(int64(1), size2)

	err = cache1.Clear(suite.ctx)
	suite.Require().NoError(err)

	size1, err = cache1.Size(suite.ctx)
	suite.Require().NoError(err)
	suite.Equal(int64(0), size1)

	retrieved, found := cache2.Get(suite.ctx, "other-user")
	suite.True(found)
	suite.Equal(user, retrieved)

	size2, err = cache2.Size(suite.ctx)
	suite.Require().NoError(err)
	suite.Equal(int64(1), size2)
}

func (suite *RedisCacheTestSuite) TestRedisCacheStringValues() {
	stringCache := suite.setupStringCache("test-strings")
	defer stringCache.Close()

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
	suite.Contains(keys, "greeting")
	suite.Contains(keys, "farewell")
}

func (suite *RedisCacheTestSuite) TestRedisCacheClose() {
	cache := suite.setupRedisCache("close-behavior")

	ctx := suite.ctx

	err := cache.Close()
	suite.Require().NoError(err)

	err = cache.Close()
	suite.Require().NoError(err)

	err = cache.Set(ctx, "key", TestUser{Id: 1})
	suite.Require().ErrorIs(err, ErrCacheClosed)

	_, found := cache.Get(ctx, "key")
	suite.False(found)
	suite.False(cache.Contains(ctx, "key"))

	suite.Require().NoError(cache.Delete(ctx, "key"))
	suite.Require().NoError(cache.Clear(ctx))

	keys, err := cache.Keys(ctx)
	suite.Require().NoError(err)
	suite.Nil(keys)

	called := false
	err = cache.ForEach(ctx, func(key string, value TestUser) bool {
		called = true

		return true
	})
	suite.Require().NoError(err)
	suite.False(called, "callback should not be invoked after cache is closed")
}

func (suite *RedisCacheTestSuite) TestRedisCacheKeyStripping() {
	userCache := suite.setupRedisCache("key-stripping-test")
	defer userCache.Close()

	suite.Run("KeysReturnUserOriginalKeys", func() {
		testData := map[string]TestUser{
			"user:1":    {Id: 1, Name: "Alice", Age: 30},
			"user:2":    {Id: 2, Name: "Bob", Age: 25},
			"admin:100": {Id: 100, Name: "Admin", Age: 40},
		}

		for key, user := range testData {
			err := userCache.Set(suite.ctx, key, user)
			suite.Require().NoError(err)
		}

		// Get all keys
		keys, err := userCache.Keys(suite.ctx)
		suite.Require().NoError(err)
		suite.Require().Len(keys, 3)

		suite.Contains(keys, "user:1")
		suite.Contains(keys, "user:2")
		suite.Contains(keys, "admin:100")

		for _, key := range keys {
			suite.NotContains(key, "vef:cache:")
			suite.NotContains(key, "key-stripping-test:")
		}
	})

	suite.Run("KeysWithPrefixReturnUserOriginalKeys", func() {
		keys, err := userCache.Keys(suite.ctx, "user")
		suite.Require().NoError(err)
		suite.Require().Len(keys, 2)

		suite.Contains(keys, "user:1")
		suite.Contains(keys, "user:2")

		suite.NotContains(keys, "admin:100")

		for _, key := range keys {
			suite.NotContains(key, "vef:cache:")
		}
	})

	suite.Run("ForEachReturnsUserOriginalKeys", func() {
		collected := make(map[string]TestUser)

		err := userCache.ForEach(suite.ctx, func(key string, user TestUser) bool {
			collected[key] = user

			return true
		})
		suite.Require().NoError(err)

		suite.Contains(collected, "user:1")
		suite.Contains(collected, "user:2")
		suite.Contains(collected, "admin:100")

		for key := range collected {
			suite.NotContains(key, "vef:cache:")
			suite.NotContains(key, "key-stripping-test:")
		}
	})

	suite.Run("ForEachWithPrefixReturnsUserOriginalKeys", func() {
		collected := make(map[string]TestUser)

		err := userCache.ForEach(suite.ctx, func(key string, user TestUser) bool {
			collected[key] = user

			return true
		}, "admin")
		suite.Require().NoError(err)

		suite.Require().Len(collected, 1)
		suite.Contains(collected, "admin:100")

		for key := range collected {
			suite.NotContains(key, "vef:cache:")
		}
	})
}

func (suite *RedisCacheTestSuite) TestRedisCacheKeyStrippingEdgeCases() {
	suite.Run("SingleLevelNamespace", func() {
		cache := suite.setupRedisCache("simple")
		defer cache.Close()

		user := TestUser{Id: 1, Name: "Alice", Age: 30}
		err := cache.Set(suite.ctx, "test-key", user)
		suite.Require().NoError(err)

		keys, err := cache.Keys(suite.ctx)
		suite.Require().NoError(err)
		suite.Require().Len(keys, 1)

		suite.Equal("test-key", keys[0])
		suite.NotContains(keys[0], "vef:cache:")
		suite.NotContains(keys[0], "simple:")
	})

	suite.Run("KeysWithColonInUserKey", func() {
		cache := suite.setupRedisCache("colon-test")
		defer cache.Close()

		complexKey := "namespace:subnamespace:item:123"
		user := TestUser{Id: 123, Name: "Complex", Age: 35}

		err := cache.Set(suite.ctx, complexKey, user)
		suite.Require().NoError(err)

		keys, err := cache.Keys(suite.ctx)
		suite.Require().NoError(err)
		suite.Require().Len(keys, 1)

		suite.Equal(complexKey, keys[0])
	})

	suite.Run("KeysWithSpecialCharacters", func() {
		cache := suite.setupRedisCache("special-chars")
		defer cache.Close()

		specialKeys := []string{
			"user@domain.com",
			"path/to/resource",
			"key-with-dashes",
			"key_with_underscores",
			"key.with.dots",
		}

		for i, key := range specialKeys {
			user := TestUser{Id: i + 1, Name: key, Age: 20 + i}
			err := cache.Set(suite.ctx, key, user)
			suite.Require().NoError(err)
		}

		keys, err := cache.Keys(suite.ctx)
		suite.Require().NoError(err)
		suite.Require().Len(keys, len(specialKeys))

		for _, expectedKey := range specialKeys {
			suite.Contains(keys, expectedKey)
		}
	})

	suite.Run("KeyStrippingDoesNotAffectGetAndSet", func() {
		cache := suite.setupRedisCache("get-set-test")
		defer cache.Close()

		userKey := "my-user-key"
		user := TestUser{Id: 999, Name: "TestUser", Age: 50}

		err := cache.Set(suite.ctx, userKey, user)
		suite.Require().NoError(err)

		retrieved, found := cache.Get(suite.ctx, userKey)
		suite.True(found)
		suite.Equal(user, retrieved)

		keys, err := cache.Keys(suite.ctx)
		suite.Require().NoError(err)
		suite.Require().Len(keys, 1)
		suite.Equal(userKey, keys[0])
	})
}

func TestRedisCacheSuite(t *testing.T) {
	suite.Run(t, new(RedisCacheTestSuite))
}
