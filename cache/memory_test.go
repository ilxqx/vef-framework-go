package cache

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMemoryCache[T any](maxSize int64, defaultTTL time.Duration, evictionPolicy EvictionPolicy, gcInterval time.Duration) Cache[T] {
	return NewMemory[T](
		WithMemMaxSize(maxSize),
		WithMemDefaultTTL(defaultTTL),
		WithMemEvictionPolicy(evictionPolicy),
		WithMemGCInterval(gcInterval),
	)
}

func TestNewMemoryOptions(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		cache := NewMemory[string]()
		defer cache.Close()

		mc := cache.(*memoryCache[string])

		assert.Zero(t, mc.maxSize)
		assert.Equal(t, EvictionPolicyNone, mc.evictionPolicy)
		assert.Zero(t, mc.defaultTTL)
		assert.Greater(t, mc.gcInterval, time.Duration(0))
	})

	t.Run("WithOptions", func(t *testing.T) {
		cache := NewMemory[string](
			WithMemMaxSize(128),
			WithMemDefaultTTL(2*time.Minute),
			WithMemEvictionPolicy(EvictionPolicyLFU),
			WithMemGCInterval(500*time.Millisecond),
		)
		defer cache.Close()

		mc := cache.(*memoryCache[string])

		assert.Equal(t, int64(128), mc.maxSize)
		assert.Equal(t, EvictionPolicyLFU, mc.evictionPolicy)
		assert.Equal(t, 2*time.Minute, mc.defaultTTL)
		assert.Equal(t, 500*time.Millisecond, mc.gcInterval)
	})
}

// TestMemoryCacheBasicOperations tests basic cache operations.
func TestMemoryCacheBasicOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("SetAndGet", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		err := cache.Set(ctx, "key1", "value1")
		require.NoError(t, err)

		value, found := cache.Get(ctx, "key1")
		assert.True(t, found)
		assert.Equal(t, "value1", value)
	})

	t.Run("GetNonExistentKey", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		value, found := cache.Get(ctx, "nonexistent")
		assert.False(t, found)
		assert.Equal(t, "", value)
	})

	t.Run("ContainsExistingKey", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		assert.True(t, cache.Contains(ctx, "key1"))
	})

	t.Run("ContainsNonExistentKey", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		assert.False(t, cache.Contains(ctx, "nonexistent"))
	})

	t.Run("DeleteExistingKey", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		err := cache.Delete(ctx, "key1")
		require.NoError(t, err)
		assert.False(t, cache.Contains(ctx, "key1"))
	})

	t.Run("DeleteNonExistentKey", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		err := cache.Delete(ctx, "nonexistent")
		require.NoError(t, err)
	})

	t.Run("UpdateExistingKey", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		_ = cache.Set(ctx, "key1", "value2")

		value, found := cache.Get(ctx, "key1")
		assert.True(t, found)
		assert.Equal(t, "value2", value)
	})

	t.Run("GetOrLoadUsesLoaderOnce", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		var loaderCalls atomic.Int32

		loader := func(ctx context.Context) (string, error) {
			loaderCalls.Add(1)

			return "loaded", nil
		}

		value, err := cache.GetOrLoad(ctx, "key1", loader)
		require.NoError(t, err)
		assert.Equal(t, "loaded", value)
		assert.Equal(t, int32(1), loaderCalls.Load())

		// Second call should hit cache without invoking loader again.
		value, err = cache.GetOrLoad(ctx, "key1", loader)
		require.NoError(t, err)
		assert.Equal(t, "loaded", value)
		assert.Equal(t, int32(1), loaderCalls.Load())
	})

	t.Run("GetOrLoadRequiresLoader", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_, err := cache.GetOrLoad(ctx, "key1", nil)
		assert.ErrorIs(t, err, ErrLoaderRequired)
	})
}

// TestMemoryCacheExpiration tests TTL and expiration.
func TestMemoryCacheExpiration(t *testing.T) {
	ctx := context.Background()

	t.Run("DefaultTTLExpiration", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 100*time.Millisecond, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		err := cache.Set(ctx, "key1", "value1")
		require.NoError(t, err)

		// Should exist immediately
		assert.True(t, cache.Contains(ctx, "key1"))

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should be expired
		assert.False(t, cache.Contains(ctx, "key1"))
		_, found := cache.Get(ctx, "key1")
		assert.False(t, found)
	})

	t.Run("CustomTTLExpiration", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		err := cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
		require.NoError(t, err)

		// Should exist immediately
		assert.True(t, cache.Contains(ctx, "key1"))

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should be expired
		assert.False(t, cache.Contains(ctx, "key1"))
	})

	t.Run("CustomTTLOverridesDefault", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 1*time.Second, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		// Set with shorter custom TTL
		_ = cache.Set(ctx, "key1", "value1", 100*time.Millisecond)

		time.Sleep(150 * time.Millisecond)

		// Should be expired (custom TTL applied)
		assert.False(t, cache.Contains(ctx, "key1"))
	})

	t.Run("ZeroTTLMeansNoExpiration", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")

		time.Sleep(100 * time.Millisecond)

		// Should still exist
		assert.True(t, cache.Contains(ctx, "key1"))
	})

	t.Run("NegativeTTLIgnored", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 100*time.Millisecond, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		// Negative TTL should be ignored, default TTL used
		_ = cache.Set(ctx, "key1", "value1", -1*time.Second)

		time.Sleep(150 * time.Millisecond)

		// Should be expired (default TTL applied)
		assert.False(t, cache.Contains(ctx, "key1"))
	})
}

// TestMemoryCacheSize tests size tracking.
func TestMemoryCacheSize(t *testing.T) {
	ctx := context.Background()

	t.Run("SizeIncreasesOnInsert", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		size, err := cache.Size(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), size)

		_ = cache.Set(ctx, "key1", "value1")
		size, err = cache.Size(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), size)

		_ = cache.Set(ctx, "key2", "value2")
		size, err = cache.Size(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), size)
	})

	t.Run("SizeDecreasesOnDelete", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		_ = cache.Set(ctx, "key2", "value2")

		cache.Delete(ctx, "key1")
		size, err := cache.Size(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), size)
	})

	t.Run("SizeUnchangedOnUpdate", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		size1, _ := cache.Size(ctx)

		_ = cache.Set(ctx, "key1", "value2")
		size2, _ := cache.Size(ctx)

		assert.Equal(t, size1, size2)
	})

	t.Run("SizeExcludesExpiredEntries", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 50*time.Millisecond, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		_ = cache.Set(ctx, "key2", "value2")

		time.Sleep(100 * time.Millisecond)

		// Trigger expiration check by accessing keys
		cache.Get(ctx, "key1")
		cache.Get(ctx, "key2")

		// Now expired entries should be removed and size updated
		size, err := cache.Size(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), size)
	})
}

// TestMemoryCacheEvictionPolicies tests different eviction policies.
func TestMemoryCacheEvictionPolicies(t *testing.T) {
	ctx := context.Background()

	t.Run("LRUEviction", func(t *testing.T) {
		cache := newTestMemoryCache[string](3, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		_ = cache.Set(ctx, "key2", "value2")
		_ = cache.Set(ctx, "key3", "value3")

		size, _ := cache.Size(ctx)
		assert.Equal(t, int64(3), size)

		// Adding one more should evict key1 (LRU)
		_ = cache.Set(ctx, "key4", "value4")

		size, _ = cache.Size(ctx)
		assert.Equal(t, int64(3), size)
		assert.False(t, cache.Contains(ctx, "key1"))
		assert.True(t, cache.Contains(ctx, "key4"))
	})

	t.Run("LRUWithAccessUpdates", func(t *testing.T) {
		cache := newTestMemoryCache[string](3, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		_ = cache.Set(ctx, "key2", "value2")
		_ = cache.Set(ctx, "key3", "value3")

		// Access key1 to make it recently used
		cache.Get(ctx, "key1")

		// Adding one more should evict key2 (now LRU)
		_ = cache.Set(ctx, "key4", "value4")

		assert.True(t, cache.Contains(ctx, "key1"))
		assert.False(t, cache.Contains(ctx, "key2"))
		assert.True(t, cache.Contains(ctx, "key3"))
		assert.True(t, cache.Contains(ctx, "key4"))
	})

	t.Run("LFUEviction", func(t *testing.T) {
		cache := newTestMemoryCache[string](3, 0, EvictionPolicyLFU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		_ = cache.Set(ctx, "key2", "value2")
		_ = cache.Set(ctx, "key3", "value3")

		// Access key1 and key2 multiple times
		cache.Get(ctx, "key1")
		cache.Get(ctx, "key1")
		cache.Get(ctx, "key2")

		// key3 has lowest frequency, should be evicted
		_ = cache.Set(ctx, "key4", "value4")

		assert.True(t, cache.Contains(ctx, "key1"))
		assert.True(t, cache.Contains(ctx, "key2"))
		assert.False(t, cache.Contains(ctx, "key3"))
		assert.True(t, cache.Contains(ctx, "key4"))
	})

	t.Run("FIFOEviction", func(t *testing.T) {
		cache := newTestMemoryCache[string](3, 0, EvictionPolicyFIFO, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		_ = cache.Set(ctx, "key2", "value2")
		_ = cache.Set(ctx, "key3", "value3")

		// Access key1 (shouldn't affect FIFO order)
		cache.Get(ctx, "key1")
		cache.Get(ctx, "key1")

		// key1 is oldest, should be evicted
		_ = cache.Set(ctx, "key4", "value4")

		assert.False(t, cache.Contains(ctx, "key1"))
		assert.True(t, cache.Contains(ctx, "key2"))
		assert.True(t, cache.Contains(ctx, "key3"))
		assert.True(t, cache.Contains(ctx, "key4"))
	})

	t.Run("NoOpPolicyFallsBackToLRU", func(t *testing.T) {
		// When maxSize is set with NoOp policy, it should default to LRU
		cache := newTestMemoryCache[string](2, 0, EvictionPolicyNone, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		_ = cache.Set(ctx, "key2", "value2")

		// Adding one more should trigger LRU eviction (not rejection)
		err := cache.Set(ctx, "key3", "value3")
		assert.NoError(t, err)

		// Cache size should remain at 2 (one entry was evicted)
		size, _ := cache.Size(ctx)
		assert.Equal(t, int64(2), size)

		// key3 should exist (was added successfully)
		assert.True(t, cache.Contains(ctx, "key3"))

		// Either key1 or key2 should have been evicted (LRU policy)
		count := 0
		if cache.Contains(ctx, "key1") {
			count++
		}

		if cache.Contains(ctx, "key2") {
			count++
		}

		assert.Equal(t, 1, count, "Exactly one of the old keys should remain")
	})

	t.Run("UpdateDoesNotTriggerEviction", func(t *testing.T) {
		cache := newTestMemoryCache[string](2, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		_ = cache.Set(ctx, "key2", "value2")

		// Update existing key should not trigger eviction
		err := cache.Set(ctx, "key1", "value1_updated")
		require.NoError(t, err)

		size, _ := cache.Size(ctx)
		assert.Equal(t, int64(2), size)
	})
}

// TestMemoryCacheKeys tests key listing.
func TestMemoryCacheKeys(t *testing.T) {
	ctx := context.Background()

	t.Run("ListAllKeys", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "user:1", "alice")
		_ = cache.Set(ctx, "user:2", "bob")
		_ = cache.Set(ctx, "post:1", "hello")
		_ = cache.Set(ctx, "post:2", "world")

		keys, err := cache.Keys(ctx)
		require.NoError(t, err)
		assert.Len(t, keys, 4)
	})

	t.Run("ListKeysWithPrefix", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "user:1", "alice")
		_ = cache.Set(ctx, "user:2", "bob")
		_ = cache.Set(ctx, "post:1", "hello")
		_ = cache.Set(ctx, "post:2", "world")

		userKeys, err := cache.Keys(ctx, "user:")
		require.NoError(t, err)
		assert.Len(t, userKeys, 2)

		postKeys, err := cache.Keys(ctx, "post:")
		require.NoError(t, err)
		assert.Len(t, postKeys, 2)
	})

	t.Run("EmptyCacheReturnsEmptyList", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		keys, err := cache.Keys(ctx)
		require.NoError(t, err)
		assert.Empty(t, keys)
	})

	t.Run("PrefixWithNoMatches", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "user:1", "alice")
		_ = cache.Set(ctx, "user:2", "bob")

		keys, err := cache.Keys(ctx, "post:")
		require.NoError(t, err)
		assert.Empty(t, keys)
	})

	t.Run("ExcludesExpiredKeys", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 50*time.Millisecond, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		_ = cache.Set(ctx, "key2", "value2")

		time.Sleep(100 * time.Millisecond)

		keys, err := cache.Keys(ctx)
		require.NoError(t, err)
		assert.Empty(t, keys)
	})
}

// TestMemoryCacheForEach tests iteration.
func TestMemoryCacheForEach(t *testing.T) {
	ctx := context.Background()

	t.Run("IterateAllEntries", func(t *testing.T) {
		cache := newTestMemoryCache[int](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "a", 1)
		_ = cache.Set(ctx, "b", 2)
		_ = cache.Set(ctx, "c", 3)

		sum := 0
		err := cache.ForEach(ctx, func(key string, value int) bool {
			sum += value

			return true
		})

		require.NoError(t, err)
		assert.Equal(t, 6, sum)
	})

	t.Run("IterateWithPrefixFilter", func(t *testing.T) {
		cache := newTestMemoryCache[int](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "user:1", 10)
		_ = cache.Set(ctx, "user:2", 20)
		_ = cache.Set(ctx, "post:1", 30)

		sum := 0
		err := cache.ForEach(ctx, func(key string, value int) bool {
			sum += value

			return true
		}, "user:")

		require.NoError(t, err)
		assert.Equal(t, 30, sum)
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		cache := newTestMemoryCache[int](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "a", 1)
		_ = cache.Set(ctx, "b", 2)
		_ = cache.Set(ctx, "c", 3)

		count := 0
		err := cache.ForEach(ctx, func(key string, value int) bool {
			count++

			return count < 2 // Stop after 2 iterations
		})

		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("EmptyCache", func(t *testing.T) {
		cache := newTestMemoryCache[int](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		called := false
		err := cache.ForEach(ctx, func(key string, value int) bool {
			called = true

			return true
		})

		require.NoError(t, err)
		assert.False(t, called)
	})

	t.Run("SkipsExpiredEntries", func(t *testing.T) {
		cache := newTestMemoryCache[int](0, 50*time.Millisecond, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "a", 1)
		_ = cache.Set(ctx, "b", 2)

		time.Sleep(100 * time.Millisecond)

		count := 0
		cache.ForEach(ctx, func(key string, value int) bool {
			count++

			return true
		})

		assert.Equal(t, 0, count)
	})
}

// TestMemoryCacheClear tests clearing the cache.
func TestMemoryCacheClear(t *testing.T) {
	ctx := context.Background()

	t.Run("ClearRemovesAllEntries", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		_ = cache.Set(ctx, "key2", "value2")
		_ = cache.Set(ctx, "key3", "value3")

		size, _ := cache.Size(ctx)
		assert.Equal(t, int64(3), size)

		err := cache.Clear(ctx)
		require.NoError(t, err)

		size, _ = cache.Size(ctx)
		assert.Equal(t, int64(0), size)
	})

	t.Run("ClearOnEmptyCache", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		err := cache.Clear(ctx)
		require.NoError(t, err)

		size, _ := cache.Size(ctx)
		assert.Equal(t, int64(0), size)
	})

	t.Run("CanAddEntriesAfterClear", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		cache.Clear(ctx)

		_ = cache.Set(ctx, "key2", "value2")
		assert.True(t, cache.Contains(ctx, "key2"))
	})
}

// TestMemoryCacheClose tests cache closure.
func TestMemoryCacheClose(t *testing.T) {
	ctx := context.Background()

	t.Run("CloseStopsGCGoroutine", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 100*time.Millisecond)

		_ = cache.Set(ctx, "key1", "value1")

		err := cache.Close()
		require.NoError(t, err)

		// Wait a bit to ensure GC goroutine has stopped
		time.Sleep(200 * time.Millisecond)
	})

	t.Run("OperationsAfterCloseFailGracefully", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)

		_ = cache.Set(ctx, "key1", "value1")
		cache.Close()

		// Get should return false
		_, found := cache.Get(ctx, "key1")
		assert.False(t, found)

		// Set should fail with ErrCacheClosed
		err := cache.Set(ctx, "key2", "value2")
		assert.ErrorIs(t, err, ErrCacheClosed)

		// Contains should return false
		assert.False(t, cache.Contains(ctx, "key1"))

		// Delete should not error
		err = cache.Delete(ctx, "key1")
		assert.NoError(t, err)

		// Clear should not error
		err = cache.Clear(ctx)
		assert.NoError(t, err)

		// Size should return 0
		size, err := cache.Size(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), size)

		// Keys should return nil
		keys, err := cache.Keys(ctx)
		assert.NoError(t, err)
		assert.Nil(t, keys)

		// ForEach should not error
		err = cache.ForEach(ctx, func(key, value string) bool {
			return true
		})
		assert.NoError(t, err)
	})

	t.Run("DoubleCloseIsSafe", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)

		err := cache.Close()
		require.NoError(t, err)

		err = cache.Close()
		assert.NoError(t, err)
	})
}

// TestMemoryCacheGC tests garbage collection.
func TestMemoryCacheGC(t *testing.T) {
	ctx := context.Background()

	t.Run("GCRemovesExpiredEntries", func(t *testing.T) {
		// Use short GC interval for testing
		cache := newTestMemoryCache[string](0, 50*time.Millisecond, EvictionPolicyLRU, 100*time.Millisecond)
		defer cache.Close()

		// Add entries that will expire
		for i := range 10 {
			_ = cache.Set(ctx, fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
		}

		size, _ := cache.Size(ctx)
		assert.Equal(t, int64(10), size)

		// Wait for expiration and GC
		time.Sleep(200 * time.Millisecond)

		// Size should be 0 after GC
		size, _ = cache.Size(ctx)
		assert.Equal(t, int64(0), size)
	})

	t.Run("GCDoesNotRemoveNonExpiredEntries", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 1*time.Second, EvictionPolicyLRU, 100*time.Millisecond)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")

		// Wait for GC cycle but not expiration
		time.Sleep(150 * time.Millisecond)

		// Entry should still exist
		assert.True(t, cache.Contains(ctx, "key1"))
	})
}

// TestMemoryCacheConcurrency tests concurrent operations.
func TestMemoryCacheConcurrency(t *testing.T) {
	ctx := context.Background()

	t.Run("ConcurrentWrites", func(t *testing.T) {
		cache := newTestMemoryCache[int](100, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		var wg sync.WaitGroup
		for i := range 50 {
			wg.Add(1)

			go func(n int) {
				defer wg.Done()

				key := fmt.Sprintf("key%d", n)
				_ = cache.Set(ctx, key, n)
			}(i)
		}

		wg.Wait()

		size, _ := cache.Size(ctx)
		assert.LessOrEqual(t, size, int64(100))
	})

	t.Run("ConcurrentReads", func(t *testing.T) {
		cache := newTestMemoryCache[int](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		// Pre-populate
		for i := range 50 {
			_ = cache.Set(ctx, fmt.Sprintf("key%d", i), i)
		}

		var wg sync.WaitGroup
		for i := range 50 {
			wg.Add(1)

			go func(n int) {
				defer wg.Done()

				key := fmt.Sprintf("key%d", n)
				cache.Get(ctx, key)
			}(i)
		}

		wg.Wait()
	})

	t.Run("ConcurrentMixedOperations", func(t *testing.T) {
		cache := newTestMemoryCache[int](100, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		var wg sync.WaitGroup

		// Concurrent writes
		for i := range 50 {
			wg.Add(1)

			go func(n int) {
				defer wg.Done()

				key := fmt.Sprintf("key%d", n)
				_ = cache.Set(ctx, key, n)
			}(i)
		}

		// Concurrent reads
		for i := range 50 {
			wg.Add(1)

			go func(n int) {
				defer wg.Done()

				key := fmt.Sprintf("key%d", n)
				cache.Get(ctx, key)
			}(i)
		}

		// Concurrent deletes
		for i := range 25 {
			wg.Add(1)

			go func(n int) {
				defer wg.Done()

				key := fmt.Sprintf("key%d", n)
				cache.Delete(ctx, key)
			}(i)
		}

		wg.Wait()

		// Should not panic and cache should be in consistent state
		size, _ := cache.Size(ctx)
		assert.GreaterOrEqual(t, size, int64(0))
		assert.LessOrEqual(t, size, int64(100))
	})

	t.Run("ConcurrentEvictions", func(t *testing.T) {
		cache := newTestMemoryCache[int](10, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		var wg sync.WaitGroup
		for i := range 100 {
			wg.Add(1)

			go func(n int) {
				defer wg.Done()

				key := fmt.Sprintf("key%d", n)
				_ = cache.Set(ctx, key, n)
			}(i)
		}

		wg.Wait()

		size, _ := cache.Size(ctx)
		assert.LessOrEqual(t, size, int64(10))
	})
}

// TestMemoryCacheEdgeCases tests edge cases.
func TestMemoryCacheEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("ZeroMaxSizeMeansUnlimited", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		// Should be able to add many entries
		for i := range 1000 {
			err := cache.Set(ctx, fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
			require.NoError(t, err)
		}

		size, _ := cache.Size(ctx)
		assert.Equal(t, int64(1000), size)
	})

	t.Run("UnlimitedSizeForcesNoOpEvictionPolicy", func(t *testing.T) {
		// Even if we specify LRU, it should be forced to NoOp when maxSize <= 0
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		mc := cache.(*memoryCache[string])

		// Verify that eviction policy was forced to None (NoOp)
		assert.Equal(t, EvictionPolicyNone, mc.evictionPolicy)

		// Verify that eviction handler is NoOp
		_, isNoOp := mc.evictionHandler.(*NoOpEvictionHandler)
		assert.True(t, isNoOp, "eviction handler should be NoOpEvictionHandler when maxSize is unlimited")

		// Should be able to add many entries without eviction
		for i := range 1000 {
			err := cache.Set(ctx, fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
			require.NoError(t, err)
		}

		size, _ := cache.Size(ctx)
		assert.Equal(t, int64(1000), size)
	})

	t.Run("NegativeMaxSizeForcesNoOpEvictionPolicy", func(t *testing.T) {
		cache := newTestMemoryCache[string](-1, 0, EvictionPolicyLFU, 5*time.Minute)
		defer cache.Close()

		mc := cache.(*memoryCache[string])

		// Verify that eviction policy was forced to None (NoOp)
		assert.Equal(t, EvictionPolicyNone, mc.evictionPolicy)

		// Verify that eviction handler is NoOp
		_, isNoOp := mc.evictionHandler.(*NoOpEvictionHandler)
		assert.True(t, isNoOp, "eviction handler should be NoOpEvictionHandler when maxSize is negative")
	})

	t.Run("MaxSizeOfOne", func(t *testing.T) {
		cache := newTestMemoryCache[string](1, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		_ = cache.Set(ctx, "key1", "value1")
		_ = cache.Set(ctx, "key2", "value2")

		size, _ := cache.Size(ctx)
		assert.Equal(t, int64(1), size)
		assert.False(t, cache.Contains(ctx, "key1"))
		assert.True(t, cache.Contains(ctx, "key2"))
	})

	t.Run("EmptyKey", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		err := cache.Set(ctx, "", "value")
		require.NoError(t, err)

		value, found := cache.Get(ctx, "")
		assert.True(t, found)
		assert.Equal(t, "value", value)
	})

	t.Run("EmptyValue", func(t *testing.T) {
		cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
		defer cache.Close()

		err := cache.Set(ctx, "key", "")
		require.NoError(t, err)

		value, found := cache.Get(ctx, "key")
		assert.True(t, found)
		assert.Equal(t, "", value)
	})

	t.Run("DifferentValueTypes", func(t *testing.T) {
		t.Run("IntCache", func(t *testing.T) {
			cache := newTestMemoryCache[int](0, 0, EvictionPolicyLRU, 5*time.Minute)
			defer cache.Close()

			_ = cache.Set(ctx, "key", 42)
			value, found := cache.Get(ctx, "key")
			assert.True(t, found)
			assert.Equal(t, 42, value)
		})

		t.Run("StructCache", func(t *testing.T) {
			type User struct {
				ID   int
				Name string
			}

			cache := newTestMemoryCache[User](0, 0, EvictionPolicyLRU, 5*time.Minute)
			defer cache.Close()

			user := User{ID: 1, Name: "Alice"}
			_ = cache.Set(ctx, "user:1", user)

			value, found := cache.Get(ctx, "user:1")
			assert.True(t, found)
			assert.Equal(t, user, value)
		})

		t.Run("PointerCache", func(t *testing.T) {
			cache := newTestMemoryCache[*string](0, 0, EvictionPolicyLRU, 5*time.Minute)
			defer cache.Close()

			str := "test"
			_ = cache.Set(ctx, "key", &str)

			value, found := cache.Get(ctx, "key")
			assert.True(t, found)
			assert.Equal(t, &str, value)
		})
	})
}

// Benchmark tests.
func BenchmarkMemoryCacheSet(b *testing.B) {
	cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
	defer cache.Close()

	ctx := context.Background()

	b.ResetTimer()

	var i int
	for b.Loop() {
		key := fmt.Sprintf("key%d", i%1000)
		_ = cache.Set(ctx, key, "value")
		i++
	}
}

func BenchmarkMemoryCacheGet(b *testing.B) {
	cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
	defer cache.Close()

	ctx := context.Background()

	// Pre-populate
	for i := range 1000 {
		_ = cache.Set(ctx, fmt.Sprintf("key%d", i), "value")
	}

	b.ResetTimer()

	var i int
	for b.Loop() {
		key := fmt.Sprintf("key%d", i%1000)
		cache.Get(ctx, key)

		i++
	}
}

func BenchmarkMemoryCacheSetWithEviction(b *testing.B) {
	cache := newTestMemoryCache[string](100, 0, EvictionPolicyLRU, 5*time.Minute)
	defer cache.Close()

	ctx := context.Background()

	b.ResetTimer()

	var i int
	for b.Loop() {
		key := fmt.Sprintf("key%d", i)
		_ = cache.Set(ctx, key, "value")
		i++
	}
}

func BenchmarkMemoryCacheConcurrent(b *testing.B) {
	cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
	defer cache.Close()

	ctx := context.Background()

	// Pre-populate
	for i := range 1000 {
		_ = cache.Set(ctx, fmt.Sprintf("key%d", i), "value")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i%1000)
			if i%2 == 0 {
				cache.Get(ctx, key)
			} else {
				_ = cache.Set(ctx, key, "value")
			}

			i++
		}
	})
}

func BenchmarkMemoryCacheSize(b *testing.B) {
	cache := newTestMemoryCache[string](0, 0, EvictionPolicyLRU, 5*time.Minute)
	defer cache.Close()

	ctx := context.Background()

	// Pre-populate
	for i := range 1000 {
		_ = cache.Set(ctx, fmt.Sprintf("key%d", i), "value")
	}

	b.ResetTimer()

	for b.Loop() {
		_, _ = cache.Size(ctx)
	}
}
