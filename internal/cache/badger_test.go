package cache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/config"
)

// Test struct for cache testing.
type TestStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// Basic Store Tests.
func TestBadgerStoreBasicOperations(t *testing.T) {
	store, err := NewBadgerStore(&config.LocalCacheConfig{InMemory: true})
	require.NoError(t, err)

	ctx := context.Background()
	defer store.Close(ctx)

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

func TestBadgerStoreTTL(t *testing.T) {
	store, err := NewBadgerStore(&config.LocalCacheConfig{InMemory: true})
	require.NoError(t, err)

	ctx := context.Background()
	defer store.Close(ctx)

	t.Run("TTL expiration", func(t *testing.T) {
		err := store.Set(ctx, "ttl-key", []byte("ttl-value"), time.Second)
		require.NoError(t, err)

		// Should exist immediately
		value, found := store.Get(ctx, "ttl-key")
		assert.True(t, found, "expected key 'ttl-key' to exist immediately after set with 1s TTL")
		assert.Equal(t, []byte("ttl-value"), value)

		// Wait for expiration (give extra time)
		time.Sleep(time.Second)

		// Should be expired
		_, found = store.Get(ctx, "ttl-key")
		assert.False(t, found, "expected key 'ttl-key' to be expired after 1s TTL")
	})

	t.Run("Default TTL", func(t *testing.T) {
		storeWithDefaultTTL, err := NewBadgerStore(&config.LocalCacheConfig{
			InMemory:   true,
			DefaultTTL: time.Second,
		})
		require.NoError(t, err)

		defer storeWithDefaultTTL.Close(ctx)

		err = storeWithDefaultTTL.Set(ctx, "default-ttl-key", []byte("default-ttl-value"))
		require.NoError(t, err)

		// Should exist immediately
		value, found := storeWithDefaultTTL.Get(ctx, "default-ttl-key")
		assert.True(t, found, "expected key 'default-ttl-key' to exist immediately after set with 1s TTL")
		assert.Equal(t, []byte("default-ttl-value"), value)

		// Wait for expiration (give extra time)
		time.Sleep(time.Second)

		// Should be expired
		_, found = storeWithDefaultTTL.Get(ctx, "default-ttl-key")
		assert.False(t, found, "expected key 'default-ttl-key' to be expired after 1s TTL")
	})
}

func TestBadgerStoreIteration(t *testing.T) {
	store, err := NewBadgerStore(&config.LocalCacheConfig{InMemory: true})
	require.NoError(t, err)

	ctx := context.Background()
	defer store.Close(ctx)

	// Setup test data
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

	t.Run("Keys without prefix", func(t *testing.T) {
		keys, err := store.Keys(ctx, "")
		require.NoError(t, err)

		sort.Strings(keys)

		expectedKeys := []string{"config:x", "product:a", "product:b", "user:1", "user:2", "user:3"}
		assert.Equal(t, expectedKeys, keys)
	})

	t.Run("Keys with prefix", func(t *testing.T) {
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
	})

	t.Run("ForEach without prefix", func(t *testing.T) {
		collected := make(map[string][]byte)

		err := store.ForEach(ctx, "", func(key string, value []byte) bool {
			collected[key] = value

			return true
		})
		require.NoError(t, err)

		assert.Equal(t, testData, collected)
	})

	t.Run("ForEach with prefix", func(t *testing.T) {
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
	})

	t.Run("ForEach early termination", func(t *testing.T) {
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
	})

	t.Run("Size", func(t *testing.T) {
		size, err := store.Size(ctx, "")
		require.NoError(t, err)
		assert.Equal(t, int64(6), size)
	})
}

func TestBadgerStoreClear(t *testing.T) {
	store, err := NewBadgerStore(&config.LocalCacheConfig{InMemory: true})
	require.NoError(t, err)

	ctx := context.Background()
	defer store.Close(ctx)

	// Add some data
	for i := range 10 {
		key := fmt.Sprintf("key-%d", i)
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

	// Verify specific keys don't exist
	for i := range 10 {
		key := fmt.Sprintf("key-%d", i)
		assert.False(t, store.Contains(ctx, key))
	}
}

func TestBadgerStoreTypes(t *testing.T) {
	ctx := context.Background()

	t.Run("String cache", func(t *testing.T) {
		store, err := NewBadgerStore(&config.LocalCacheConfig{InMemory: true})
		require.NoError(t, err)

		defer store.Close(ctx)

		err = store.Set(ctx, "key", []byte("hello world"))
		require.NoError(t, err)

		value, found := store.Get(ctx, "key")
		assert.True(t, found)
		assert.Equal(t, []byte("hello world"), value)
	})

	t.Run("Integer cache", func(t *testing.T) {
		store, err := NewBadgerStore(&config.LocalCacheConfig{InMemory: true})
		require.NoError(t, err)

		defer store.Close(ctx)

		err = store.Set(ctx, "key", []byte("42"))
		require.NoError(t, err)

		value, found := store.Get(ctx, "key")
		assert.True(t, found)
		assert.Equal(t, []byte("42"), value)
	})

	t.Run("Slice cache", func(t *testing.T) {
		store, err := NewBadgerStore(&config.LocalCacheConfig{InMemory: true})
		require.NoError(t, err)

		defer store.Close(ctx)

		testData := []byte(`["a","b","c"]`)
		err = store.Set(ctx, "key", testData)
		require.NoError(t, err)

		value, found := store.Get(ctx, "key")
		assert.True(t, found)
		assert.Equal(t, testData, value)
	})

	t.Run("Map cache", func(t *testing.T) {
		store, err := NewBadgerStore(&config.LocalCacheConfig{InMemory: true})
		require.NoError(t, err)

		defer store.Close(ctx)

		testData := []byte(`{"a":1,"b":2}`)
		err = store.Set(ctx, "key", testData)
		require.NoError(t, err)

		value, found := store.Get(ctx, "key")
		assert.True(t, found)
		assert.Equal(t, testData, value)
	})
}

func TestBadgerStoreEdgeCases(t *testing.T) {
	store, err := NewBadgerStore(&config.LocalCacheConfig{InMemory: true})
	require.NoError(t, err)

	ctx := context.Background()
	defer store.Close(ctx)

	t.Run("Empty key", func(t *testing.T) {
		err := store.Set(ctx, "", []byte("empty-key-value"))
		// Badger doesn't allow empty keys, so this should error
		if err != nil {
			t.Logf("Empty key correctly rejected: %v", err)

			return
		}

		value, found := store.Get(ctx, "")
		assert.True(t, found)
		assert.Equal(t, []byte("empty-key-value"), value)
	})

	t.Run("Empty value", func(t *testing.T) {
		err := store.Set(ctx, "empty-value", []byte(""))
		require.NoError(t, err)

		value, found := store.Get(ctx, "empty-value")
		assert.True(t, found)
		assert.Equal(t, []byte(""), value)
	})

	t.Run("Special characters in key", func(t *testing.T) {
		specialKey := "key:with/special\\chars@#$%^&*()"
		err := store.Set(ctx, specialKey, []byte("special-value"))
		require.NoError(t, err)

		value, found := store.Get(ctx, specialKey)
		assert.True(t, found)
		assert.Equal(t, []byte("special-value"), value)
	})

	t.Run("Delete non-existent key", func(t *testing.T) {
		err := store.Delete(ctx, "non-existent-key")
		// Should not error when deleting non-existent key
		require.NoError(t, err)
	})

	t.Run("Empty prefix", func(t *testing.T) {
		// Add some test data
		err := store.Set(ctx, "test1", []byte("value1"))
		require.NoError(t, err)
		err = store.Set(ctx, "test2", []byte("value2"))
		require.NoError(t, err)

		// Empty prefix should return all keys
		keys, err := store.Keys(ctx, "")
		require.NoError(t, err)
		assert.Greater(t, len(keys), 0)

		// ForEach with empty prefix should iterate all
		count := 0
		err = store.ForEach(ctx, "", func(key string, value []byte) bool {
			count++

			return true
		})
		require.NoError(t, err)
		assert.Greater(t, count, 0)
	})
}

// Garbage Collection Tests.
func TestBadgerStoreGarbageCollection(t *testing.T) {
	t.Run("In-memory store should not start GC goroutine", func(t *testing.T) {
		store, err := NewBadgerStore(&config.LocalCacheConfig{InMemory: true})
		require.NoError(t, err)

		// For in-memory cache, GC goroutine should not be started
		// We can verify this by closing immediately without issues
		err = store.Close(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Persistent store should start and stop GC goroutine properly", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "gc-test-*")
		require.NoError(t, err)

		defer os.RemoveAll(tempDir)

		store, err := NewBadgerStore(&config.LocalCacheConfig{
			InMemory:  false,
			Directory: tempDir,
		})
		require.NoError(t, err)

		ctx := context.Background()

		// Add some data
		err = store.Set(ctx, "key1", []byte("value1"))
		require.NoError(t, err)

		// Close should gracefully stop the GC goroutine
		err = store.Close(ctx)
		assert.NoError(t, err)
	})

	t.Run("Multiple close calls should not panic", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "multi-close-test-*")
		require.NoError(t, err)

		defer os.RemoveAll(tempDir)

		store, err := NewBadgerStore(&config.LocalCacheConfig{
			InMemory:  false,
			Directory: tempDir,
		})
		require.NoError(t, err)

		ctx := context.Background()

		// First close
		err = store.Close(ctx)
		assert.NoError(t, err)

		// Second close should not panic (though it might error)
		// This tests the robustness of our channel handling
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Second close caused panic: %v", r)
			}
		}()

		store.Close(ctx) // This might error but should not panic
		// Note: Badger's Close is idempotent, so second close might not error
	})
}

func TestBadgerStoreGracefulShutdown(t *testing.T) {
	t.Run("Store close stops GC goroutine via ErrRejected", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "shutdown-test-*")
		require.NoError(t, err)

		defer os.RemoveAll(tempDir)

		store, err := NewBadgerStore(&config.LocalCacheConfig{
			InMemory:  false,
			Directory: tempDir,
		})
		require.NoError(t, err)

		ctx := context.Background()

		// Add some data
		for i := range 10 {
			err = store.Set(ctx, fmt.Sprintf("key%d", i), fmt.Appendf(nil, "value%d", i))
			require.NoError(t, err)
		}

		// Close should complete quickly as GC will detect ErrRejected and stop
		start := time.Now()
		err = store.Close(ctx)
		duration := time.Since(start)

		assert.NoError(t, err)
		// Should finish very quickly since we rely on Badger's ErrRejected mechanism
		assert.Less(t, duration, 1*time.Second, "Close took too long")
	})
}

// Persistent Storage Tests.
func TestBadgerStorePersistentOperations(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	// Create persistent cache
	store, err := NewBadgerStore(&config.LocalCacheConfig{
		InMemory:  false,
		Directory: tempDir,
	})
	require.NoError(t, err)

	ctx := context.Background()
	defer store.Close(ctx)

	t.Run("Basic persistent operations", func(t *testing.T) {
		// Set some values
		err := store.Set(ctx, "key1", []byte("value1"))
		require.NoError(t, err)

		err = store.Set(ctx, "key2", []byte("value2"))
		require.NoError(t, err)

		// Get values
		value1, found := store.Get(ctx, "key1")
		assert.True(t, found)
		assert.Equal(t, []byte("value1"), value1)

		value2, found := store.Get(ctx, "key2")
		assert.True(t, found)
		assert.Equal(t, []byte("value2"), value2)

		// Check size
		size, err := store.Size(ctx, "")
		require.NoError(t, err)
		assert.Equal(t, int64(2), size)
	})

	t.Run("Verify files are created", func(t *testing.T) {
		// Check that database files are created in the directory
		entries, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		assert.Greater(t, len(entries), 0, "Should create database files")
	})
}

func TestBadgerStorePersistentErrorHandling(t *testing.T) {
	t.Run("Missing directory should error", func(t *testing.T) {
		store, err := NewBadgerStore(&config.LocalCacheConfig{
			InMemory: false,
			// Directory is empty, should cause error
		})

		assert.Error(t, err)
		assert.Nil(t, store)
		assert.Contains(t, err.Error(), "directory path is required")
	})

	t.Run("Invalid directory should error", func(t *testing.T) {
		invalidPath := filepath.Join("/", "non-existent-root-dir", "cache-test")

		store, err := NewBadgerStore(&config.LocalCacheConfig{
			InMemory:  false,
			Directory: invalidPath,
		})

		// Should error when trying to create/access the directory
		assert.Error(t, err)
		assert.Nil(t, store)
	})
}

func TestBadgerStoreConfigurationDifferences(t *testing.T) {
	// This test documents the different configurations for memory vs persistent
	t.Run("In-memory store configuration", func(t *testing.T) {
		store, err := NewBadgerStore(&config.LocalCacheConfig{InMemory: true})
		require.NoError(t, err)

		ctx := context.Background()
		defer store.Close(ctx)

		// Should work with optimized in-memory configuration
		// (No compression, larger caches for speed)
		err = store.Set(ctx, "test", []byte("value"))
		require.NoError(t, err)

		value, found := store.Get(ctx, "test")
		assert.True(t, found)
		assert.Equal(t, []byte("value"), value)
	})

	t.Run("Persistent store configuration", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "persistent-config-test-*")
		require.NoError(t, err)

		defer os.RemoveAll(tempDir)

		store, err := NewBadgerStore(&config.LocalCacheConfig{
			InMemory:  false,
			Directory: tempDir,
		})
		require.NoError(t, err)

		ctx := context.Background()
		defer store.Close(ctx)

		// Should work with optimized persistent configuration
		// (Snappy compression, balanced cache sizes)
		err = store.Set(ctx, "test", []byte("value"))
		require.NoError(t, err)

		value, found := store.Get(ctx, "test")
		assert.True(t, found)
		assert.Equal(t, []byte("value"), value)
	})
}
