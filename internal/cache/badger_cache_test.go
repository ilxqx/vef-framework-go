package cache

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/cache"
)

// TestUser represents a test user struct for cache operations.
type TestUserBadger struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// setupBadgerCache sets up a Badger-backed cache for testing.
func setupBadgerCache[T any](t *testing.T, cacheName string) cache.Cache[T] {
	store, err := createBadgerStore(badgerOptions{InMemory: true})
	require.NoError(t, err, "Failed to create badger store")

	// Create cache instance
	testCache := cache.New[T](cacheName, store)

	// Add cleanup function
	t.Cleanup(func() {
		ctx := context.Background()
		if err := testCache.Clear(ctx); err != nil {
			t.Errorf("Failed to clear cache: %v", err)
		}

		if err := store.Close(ctx); err != nil {
			t.Errorf("Failed to close store: %v", err)
		}
	})

	return testCache
}

func TestBadgerCacheBasicOperations(t *testing.T) {
	userCache := setupBadgerCache[TestUserBadger](t, "test-users")
	ctx := context.Background()

	t.Run("Set and Get", func(t *testing.T) {
		user := TestUserBadger{ID: 1, Name: "Alice", Age: 30}

		err := userCache.Set(ctx, "user1", user)
		require.NoError(t, err)

		result, found := userCache.Get(ctx, "user1")
		assert.True(t, found)
		assert.Equal(t, user, result)
	})

	t.Run("Contains", func(t *testing.T) {
		user := TestUserBadger{ID: 2, Name: "Bob", Age: 25}

		err := userCache.Set(ctx, "user2", user)
		require.NoError(t, err)

		assert.True(t, userCache.Contains(ctx, "user2"))
		assert.False(t, userCache.Contains(ctx, "nonexistent"))
	})

	t.Run("Delete", func(t *testing.T) {
		user := TestUserBadger{ID: 3, Name: "Charlie", Age: 35}

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
		originalUser := TestUserBadger{ID: 4, Name: "David", Age: 40}
		updatedUser := TestUserBadger{ID: 4, Name: "David", Age: 41}

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

func TestBadgerCacheTTL(t *testing.T) {
	userCache := setupBadgerCache[TestUserBadger](t, "test-ttl-users")
	ctx := context.Background()

	t.Run("TTL expiration", func(t *testing.T) {
		user := TestUserBadger{ID: 5, Name: "Eve", Age: 28}

		err := userCache.Set(ctx, "ttl-user", user, time.Second)
		require.NoError(t, err)

		// Should exist immediately
		result, found := userCache.Get(ctx, "ttl-user")
		assert.True(t, found, "expected key 'ttl-user' to exist immediately after set with 1s TTL")
		assert.Equal(t, user, result)

		// Wait for expiration
		time.Sleep(time.Second)

		// Should be expired
		_, found = userCache.Get(ctx, "ttl-user")
		assert.False(t, found, "expected key 'ttl-user' to be expired after 1s TTL")
	})
}

func TestBadgerCacheKeyPrefixIsolation(t *testing.T) {
	// Create two caches with different names
	cache1 := setupBadgerCache[TestUserBadger](t, "cache1")
	cache2 := setupBadgerCache[TestUserBadger](t, "cache2")
	ctx := context.Background()

	user1 := TestUserBadger{ID: 1, Name: "Alice", Age: 30}
	user2 := TestUserBadger{ID: 2, Name: "Bob", Age: 25}

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

func TestBadgerCacheIteration(t *testing.T) {
	userCache := setupBadgerCache[TestUserBadger](t, "test-iteration")
	ctx := context.Background()

	// Setup test data
	testUsers := map[string]TestUserBadger{
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
		collected := make(map[string]TestUserBadger)

		err := userCache.ForEach(ctx, func(key string, user TestUserBadger) bool {
			collected[key] = user

			return true
		})
		require.NoError(t, err)

		// Should collect all users with prefixed keys
		expectedCollected := map[string]TestUserBadger{
			"vef:test-iteration:admin:1": testUsers["admin:1"],
			"vef:test-iteration:admin:2": testUsers["admin:2"],
			"vef:test-iteration:guest:1": testUsers["guest:1"],
			"vef:test-iteration:user:1":  testUsers["user:1"],
			"vef:test-iteration:user:2":  testUsers["user:2"],
		}
		assert.Equal(t, expectedCollected, collected)
	})

	t.Run("ForEach with prefix", func(t *testing.T) {
		adminCollected := make(map[string]TestUserBadger)

		err := userCache.ForEach(ctx, func(key string, user TestUserBadger) bool {
			adminCollected[key] = user

			return true
		}, "admin")
		require.NoError(t, err)

		expectedAdminCollected := map[string]TestUserBadger{
			"vef:test-iteration:admin:1": testUsers["admin:1"],
			"vef:test-iteration:admin:2": testUsers["admin:2"],
		}
		assert.Equal(t, expectedAdminCollected, adminCollected)
	})

	t.Run("ForEach early termination", func(t *testing.T) {
		collected := make(map[string]TestUserBadger)
		count := 0

		err := userCache.ForEach(ctx, func(key string, user TestUserBadger) bool {
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

func TestBadgerCacheClear(t *testing.T) {
	// Create two separate caches to test isolation during clear
	cache1 := setupBadgerCache[TestUserBadger](t, "clear-test-1")
	cache2 := setupBadgerCache[TestUserBadger](t, "clear-test-2")
	ctx := context.Background()

	t.Run("Clear isolates between different cache instances", func(t *testing.T) {
		// Add data to both caches
		for i := 1; i <= 5; i++ {
			user := TestUserBadger{ID: i, Name: fmt.Sprintf("User%d", i), Age: 20 + i}
			err := cache1.Set(ctx, fmt.Sprintf("user-%d", i), user)
			require.NoError(t, err)
		}

		user := TestUserBadger{ID: 99, Name: "Other User", Age: 99}
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

func TestBadgerCacheTypeSafety(t *testing.T) {
	ctx := context.Background()

	t.Run("String cache", func(t *testing.T) {
		stringCache := setupBadgerCache[string](t, "test-strings")

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

	t.Run("Integer cache", func(t *testing.T) {
		intCache := setupBadgerCache[int](t, "test-ints")

		err := intCache.Set(ctx, "answer", 42)
		require.NoError(t, err)

		result, found := intCache.Get(ctx, "answer")
		assert.True(t, found)
		assert.Equal(t, 42, result)
	})

	t.Run("Slice cache", func(t *testing.T) {
		sliceCache := setupBadgerCache[[]string](t, "test-slices")

		testSlice := []string{"a", "b", "c"}
		err := sliceCache.Set(ctx, "items", testSlice)
		require.NoError(t, err)

		result, found := sliceCache.Get(ctx, "items")
		assert.True(t, found)
		assert.Equal(t, testSlice, result)
	})

	t.Run("Map cache", func(t *testing.T) {
		mapCache := setupBadgerCache[map[string]int](t, "test-maps")

		testMap := map[string]int{"a": 1, "b": 2}
		err := mapCache.Set(ctx, "data", testMap)
		require.NoError(t, err)

		result, found := mapCache.Get(ctx, "data")
		assert.True(t, found)
		assert.Equal(t, testMap, result)
	})

	t.Run("Struct cache", func(t *testing.T) {
		structCache := setupBadgerCache[TestUserBadger](t, "test-structs")

		testUser := TestUserBadger{ID: 123, Name: "Test User", Age: 25}
		err := structCache.Set(ctx, "user", testUser)
		require.NoError(t, err)

		result, found := structCache.Get(ctx, "user")
		assert.True(t, found)
		assert.Equal(t, testUser, result)
	})
}

func TestBadgerCacheSerializationEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("Empty struct", func(t *testing.T) {
		type EmptyStruct struct{}

		emptyCache := setupBadgerCache[EmptyStruct](t, "test-empty")

		empty := EmptyStruct{}
		err := emptyCache.Set(ctx, "empty", empty)
		require.NoError(t, err)

		result, found := emptyCache.Get(ctx, "empty")
		assert.True(t, found)
		assert.Equal(t, empty, result)
	})

	t.Run("Pointer types", func(t *testing.T) {
		ptrCache := setupBadgerCache[*TestUserBadger](t, "test-pointers")

		user := &TestUserBadger{ID: 1, Name: "Pointer User", Age: 30}
		err := ptrCache.Set(ctx, "ptr-user", user)
		require.NoError(t, err)

		result, found := ptrCache.Get(ctx, "ptr-user")
		assert.True(t, found)
		assert.Equal(t, user, result)
	})

	t.Run("Nil pointer", func(t *testing.T) {
		ptrCache := setupBadgerCache[*TestUserBadger](t, "test-nil-pointers")

		var nilUser *TestUserBadger = nil

		// Use defer to catch panic from gob encoder
		var panicCaught bool

		func() {
			defer func() {
				if r := recover(); r != nil {
					panicCaught = true

					t.Logf("Nil pointer correctly rejected by gob serializer (panicked): %v", r)
					assert.Contains(t, fmt.Sprintf("%v", r), "cannot encode nil pointer")
				}
			}()

			err := ptrCache.Set(ctx, "nil-user", nilUser)
			if err != nil {
				t.Logf("Nil pointer correctly rejected by gob serializer (error): %v", err)
				assert.Contains(t, err.Error(), "cannot encode nil pointer")
			}
		}()

		if !panicCaught {
			t.Log("Expected panic for nil pointer was not caught")
		}
	})
}
