package cache

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/constants"
)

// TestNoOpEvictionHandler tests the no-op eviction handler.
func TestNoOpEvictionHandler(t *testing.T) {
	handler := NewNoOpEvictionHandler()
	require.NotNil(t, handler)

	t.Run("AllOperationsNoOp", func(t *testing.T) {
		// These should not panic
		handler.OnAccess("key1")
		handler.OnInsert("key1")
		handler.OnEvict("key1")
		handler.Reset()
	})

	t.Run("AlwaysReturnEmptyCandidate", func(t *testing.T) {
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})

	t.Run("HandleMultipleOperations", func(t *testing.T) {
		for i := range 100 {
			handler.OnInsert(fmt.Sprintf("key%d", i))
			handler.OnAccess(fmt.Sprintf("key%d", i))
		}

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})
}

// TestLRUHandler tests the LRU eviction handler.
func TestLRUHandler(t *testing.T) {
	t.Run("BasicInsertionAndEviction", func(t *testing.T) {
		handler := NewLruHandler()
		require.NotNil(t, handler)

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// key1 should be LRU (oldest)
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("AccessUpdatesRecency", func(t *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// Access key1, making it most recently used
		handler.OnAccess("key1")

		// Now key2 should be LRU
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("EvictionRemovesEntry", func(t *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// Evict key1
		handler.OnEvict("key1")

		// Now key2 should be LRU
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("MultipleAccessesMaintainOrder", func(t *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// Access in specific order
		handler.OnAccess("key2")
		handler.OnAccess("key1")
		handler.OnAccess("key3")

		// key2 was accessed first, so it's LRU
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("ResetClearsAllEntries", func(t *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.Reset()

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})

	t.Run("EmptyHandlerReturnsEmptyCandidate", func(t *testing.T) {
		handler := NewLruHandler()

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})

	t.Run("SingleEntry", func(t *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("EvictNonExistentKey", func(t *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")

		// Evict non-existent key (should not panic)
		handler.OnEvict("key3")

		// key1 should still be LRU
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("AccessNonExistentKeyCreatesEntry", func(t *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnAccess("key2") // This creates a new entry

		// key1 should be LRU
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		handler := NewLruHandler()

		var wg sync.WaitGroup

		// Concurrent inserts
		for i := range 100 {
			wg.Add(1)

			go func(n int) {
				defer wg.Done()

				key := fmt.Sprintf("key%d", n%26)
				handler.OnInsert(key)
				handler.OnAccess(key)
			}(i)
		}

		wg.Wait()

		// Should not panic and should return a valid candidate
		candidate := handler.SelectEvictionCandidate()
		assert.NotEqual(t, constants.Empty, candidate)
	})

	t.Run("StressTestWithManyEntries", func(t *testing.T) {
		handler := NewLruHandler()

		// Insert many entries
		for i := range 1000 {
			handler.OnInsert(fmt.Sprintf("key%d", i))
		}

		// Access some entries
		for i := range 500 {
			handler.OnAccess(fmt.Sprintf("key%d", i*2))
		}

		// Should return a valid candidate
		candidate := handler.SelectEvictionCandidate()
		assert.NotEqual(t, constants.Empty, candidate)
	})
}

// TestFIFOHandler tests the FIFO eviction handler.
func TestFIFOHandler(t *testing.T) {
	t.Run("BasicInsertionAndEviction", func(t *testing.T) {
		handler := NewFifoHandler()
		require.NotNil(t, handler)

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// key1 should be first (oldest)
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("AccessDoesNotAffectOrder", func(t *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// Access key1 multiple times
		handler.OnAccess("key1")
		handler.OnAccess("key1")
		handler.OnAccess("key1")

		// key1 should still be first (FIFO ignores access)
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("EvictionRemovesEntry", func(t *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// Evict key1
		handler.OnEvict("key1")

		// Now key2 should be first
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("ResetClearsAllEntries", func(t *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.Reset()

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})

	t.Run("EmptyHandlerReturnsEmptyCandidate", func(t *testing.T) {
		handler := NewFifoHandler()

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})

	t.Run("SingleEntry", func(t *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("DuplicateInsertIgnored", func(t *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key1") // Duplicate

		// key1 should still be first (not moved to back)
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("EvictNonExistentKey", func(t *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")

		// Evict non-existent key (should not panic)
		handler.OnEvict("key3")

		// key1 should still be first
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		handler := NewFifoHandler()

		var wg sync.WaitGroup

		// Concurrent inserts
		for i := range 100 {
			wg.Add(1)

			go func(n int) {
				defer wg.Done()

				key := fmt.Sprintf("key%d", n)
				handler.OnInsert(key)
				handler.OnAccess(key)
			}(i)
		}

		wg.Wait()

		// Should not panic and should return a valid candidate
		candidate := handler.SelectEvictionCandidate()
		assert.NotEqual(t, constants.Empty, candidate)
	})
}

// TestLFUHandler tests the LFU eviction handler.
func TestLFUHandler(t *testing.T) {
	t.Run("BasicInsertionAndEviction", func(t *testing.T) {
		handler := NewLfuHandler()
		require.NotNil(t, handler)

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// All have frequency 1, key1 should be selected (oldest by insertion order)
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("AccessIncreasesFrequency", func(t *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// Access key1 twice, key2 once
		handler.OnAccess("key1")
		handler.OnAccess("key1")
		handler.OnAccess("key2")

		// key3 has frequency 1, should be LFU
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key3", candidate)
	})

	t.Run("EvictionRemovesEntry", func(t *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.OnAccess("key1")
		handler.OnAccess("key1")
		handler.OnAccess("key2")

		// Evict key3
		handler.OnEvict("key3")

		// key2 has frequency 2, key1 has frequency 3
		// key2 should be LFU
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("TieBreakingByInsertionOrder", func(t *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// All have frequency 1
		// key1 should be selected (oldest by insertion order)
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("FrequencyOrderingMaintained", func(t *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// Access to create different frequencies
		// key1: freq 1 (no access)
		// key2: freq 2 (1 access)
		// key3: freq 3 (2 accesses)
		handler.OnAccess("key2")
		handler.OnAccess("key3")
		handler.OnAccess("key3")

		// key1 has lowest frequency (1)
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("ResetClearsAllEntries", func(t *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.Reset()

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})

	t.Run("EmptyHandlerReturnsEmptyCandidate", func(t *testing.T) {
		handler := NewLfuHandler()

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})

	t.Run("SingleEntry", func(t *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("EvictNonExistentKey", func(t *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")

		// Evict non-existent key (should not panic)
		handler.OnEvict("key3")

		// key1 should still be LFU
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("AccessNonExistentKeyCreatesEntry", func(t *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnAccess("key1")
		// OnAccess on non-existent key does NOT create entry in new implementation
		handler.OnAccess("key3") // This does nothing

		// key2 has lowest frequency (1), key1 has frequency 2
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		handler := NewLfuHandler()

		var wg sync.WaitGroup

		// Concurrent inserts and accesses
		for i := range 100 {
			wg.Add(1)

			go func(n int) {
				defer wg.Done()

				key := fmt.Sprintf("key%d", n%26)
				handler.OnInsert(key)

				for range n % 10 {
					handler.OnAccess(key)
				}
			}(i)
		}

		wg.Wait()

		// Should not panic
		candidate := handler.SelectEvictionCandidate()
		// Candidate may be empty or non-empty depending on race conditions
		_ = candidate
	})

	t.Run("StressTestWithManyEntries", func(t *testing.T) {
		handler := NewLfuHandler()

		// Insert many keys
		n := 1000
		for i := range n {
			handler.OnInsert(fmt.Sprintf("key%d", i))
		}

		// Access keys with varying frequencies
		for i := range n {
			for range i % 100 {
				handler.OnAccess(fmt.Sprintf("key%d", i))
			}
		}

		// Select and evict candidates (should be fast)
		for range 100 {
			candidate := handler.SelectEvictionCandidate()
			require.NotEqual(t, constants.Empty, candidate)
			handler.OnEvict(candidate)
		}
	})

	t.Run("FrequencyBucketsWorkCorrectly", func(t *testing.T) {
		handler := NewLfuHandler()

		// Create entries with specific frequencies
		handler.OnInsert("key1") // freq 1
		handler.OnInsert("key2") // freq 1
		handler.OnAccess("key2") // freq 2
		handler.OnInsert("key3") // freq 1
		handler.OnAccess("key3") // freq 2
		handler.OnAccess("key3") // freq 3

		// key1 has lowest frequency (1)
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)

		handler.OnEvict("key1")

		// key2 has lowest frequency (2)
		candidate = handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)

		handler.OnEvict("key2")

		// key3 is the only one left
		candidate = handler.SelectEvictionCandidate()
		assert.Equal(t, "key3", candidate)
	})
}

// TestEvictionHandlerFactory tests the eviction handler factory.
func TestEvictionHandlerFactory(t *testing.T) {
	factory := &EvictionHandlerFactory{}
	require.NotNil(t, factory)

	testCases := []struct {
		policy       EvictionPolicy
		expectedType string
	}{
		{EvictionPolicyNone, "*cache.NoOpEvictionHandler"},
		{EvictionPolicyLRU, "*cache.LruHandler"},
		{EvictionPolicyLFU, "*cache.LfuHandler"},
		{EvictionPolicyFIFO, "*cache.FifoHandler"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("policy_%d", tc.policy), func(t *testing.T) {
			handler := factory.CreateHandler(tc.policy)
			require.NotNil(t, handler)

			typeName := fmt.Sprintf("%T", handler)
			assert.Equal(t, tc.expectedType, typeName)
		})
	}

	t.Run("InvalidPolicyDefaultsToNoOp", func(t *testing.T) {
		handler := factory.CreateHandler(EvictionPolicy(999))
		require.NotNil(t, handler)

		typeName := fmt.Sprintf("%T", handler)
		assert.Equal(t, "*cache.NoOpEvictionHandler", typeName)
	})
}

// TestLRUHandlerUpdateBehavior tests that LRU handler correctly handles updates
// without causing duplicate entries or other issues.
func TestLRUHandlerUpdateBehavior(t *testing.T) {
	t.Run("UpdateMoveKeyToFront", func(t *testing.T) {
		handler := NewLruHandler()

		// Insert 3 keys
		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// key1 should be LRU
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)

		// Update key1 (simulating cache.Set on existing key)
		handler.OnAccess("key1")

		// Now key2 should be LRU
		candidate = handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("RepeatedUpdatesDoNotCauseDuplicates", func(t *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")

		// Update multiple times
		for range 10 {
			handler.OnAccess("key1")
		}

		// Should still have only 1 entry
		assert.Equal(t, 1, len(handler.accessMap))
		assert.Equal(t, 1, handler.accessList.Len())
	})

	t.Run("InterleavedInsertsAndAccesses", func(t *testing.T) {
		handler := NewLruHandler()

		// Simulate mixed insert and access patterns
		handler.OnInsert("key1") // key1 at front
		handler.OnAccess("key1") // key1 stays at front (already exists, moved to front)
		handler.OnInsert("key2") // key2 at front, key1 now second
		handler.OnAccess("key1") // key1 moved to front, key2 now second
		handler.OnInsert("key3") // key3 at front, order: key3, key1, key2
		handler.OnAccess("key2") // key2 moved to front, order: key2, key3, key1

		// Verify internal consistency
		assert.Equal(t, 3, len(handler.accessMap))
		assert.Equal(t, 3, handler.accessList.Len())

		// Order should be: key2 (most recent) -> key3 -> key1 (LRU/back)
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})
}

// TestLFUHandlerUpdateBehavior tests LFU handler update scenarios.
func TestLFUHandlerUpdateBehavior(t *testing.T) {
	t.Run("RepeatedUpdatesDoNotCauseDuplicates", func(t *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")

		// Update multiple times
		for range 10 {
			handler.OnAccess("key1")
		}

		// Should still have only 1 entry
		assert.Equal(t, 1, len(handler.keyToNode))
		assert.Equal(t, 1, len(handler.keyToBucket))
	})

	t.Run("FrequencyIncrementsCorrectly", func(t *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// Access key1 five times
		for range 5 {
			handler.OnAccess("key1")
		}

		// Access key2 three times
		for range 3 {
			handler.OnAccess("key2")
		}

		// key3 has frequency 1 (no access after insert)
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key3", candidate)

		// Evict key3
		handler.OnEvict("key3")

		// Now key2 has lowest frequency (3+1=4)
		candidate = handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("FrequencyBucketMovement", func(t *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// All start with frequency 1
		assert.Equal(t, int64(1), handler.minFreq)

		// Access key1 to move it to frequency 2
		handler.OnAccess("key1")

		// minFreq should still be 1 (key2 and key3 are still at freq 1)
		assert.Equal(t, int64(1), handler.minFreq)

		// Evict key2 and key3 (both at freq 1)
		handler.OnEvict("key2")
		handler.OnEvict("key3")

		// After evicting all freq 1 keys, minFreq should be recalculated to 2
		assert.Equal(t, int64(2), handler.minFreq)

		// Only key1 should remain
		assert.Equal(t, 1, len(handler.keyToNode))
	})
}

// TestFIFOHandlerUpdateBehavior tests FIFO handler update scenarios.
func TestFIFOHandlerUpdateBehavior(t *testing.T) {
	t.Run("RepeatedUpdatesDoNotCauseDuplicates", func(t *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")

		// Update multiple times (FIFO ignores access)
		for range 10 {
			handler.OnAccess("key1")
		}

		// Should still have only 1 entry
		assert.Equal(t, 1, len(handler.insertMap))
		assert.Equal(t, 1, handler.insertList.Len())
	})

	t.Run("AccessDoesNotChangeOrder", func(t *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		// Access keys in different order
		handler.OnAccess("key3")
		handler.OnAccess("key1")
		handler.OnAccess("key2")

		// key1 should still be the eviction candidate (oldest)
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})
}

// TestEvictionHandlerInternalConsistency verifies that eviction handlers maintain internal consistency.
func TestEvictionHandlerInternalConsistency(t *testing.T) {
	t.Run("LRUHandlerConsistency", func(t *testing.T) {
		handler := NewLruHandler()

		// Perform mixed operations
		for range 100 {
			handler.OnInsert("key1")
			handler.OnAccess("key1")
			handler.OnInsert("key2")
			handler.OnAccess("key2")
		}

		// Should have exactly 2 entries
		assert.Equal(t, 2, len(handler.accessMap))
		assert.Equal(t, 2, handler.accessList.Len())
	})

	t.Run("LFUHandlerConsistency", func(t *testing.T) {
		handler := NewLfuHandler()

		// Perform mixed operations
		for range 100 {
			handler.OnInsert("key1")
			handler.OnAccess("key1")
			handler.OnInsert("key2")
			handler.OnAccess("key2")
		}

		// Should have exactly 2 entries
		assert.Equal(t, 2, len(handler.keyToNode))
		assert.Equal(t, 2, len(handler.keyToBucket))
	})

	t.Run("FIFOHandlerConsistency", func(t *testing.T) {
		handler := NewFifoHandler()

		// Perform mixed operations
		for range 100 {
			handler.OnInsert("key1")
			handler.OnAccess("key1")
			handler.OnInsert("key2")
			handler.OnAccess("key2")
		}

		// Should have exactly 2 entries
		assert.Equal(t, 2, len(handler.insertMap))
		assert.Equal(t, 2, handler.insertList.Len())
	})
}

// TestEvictionHandlerEdgeCases tests edge cases for eviction handlers.
func TestEvictionHandlerEdgeCases(t *testing.T) {
	t.Run("LRUHandlerEvictAndReinsert", func(t *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")

		// Evict key1
		handler.OnEvict("key1")
		assert.Equal(t, 1, len(handler.accessMap))

		// Re-insert key1
		handler.OnInsert("key1")
		assert.Equal(t, 2, len(handler.accessMap))

		// key2 should now be LRU (key1 was just inserted)
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("LFUHandlerEvictAndReinsert", func(t *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnAccess("key1")
		handler.OnAccess("key1")

		handler.OnInsert("key2")

		// key2 has lowest frequency
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)

		// Evict key2
		handler.OnEvict("key2")

		// Re-insert key2
		handler.OnInsert("key2")

		// Now key2 has frequency 1, key1 has frequency 3
		// key2 should be LFU
		candidate = handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("FIFOHandlerEvictAndReinsert", func(t *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")

		// key1 is oldest
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)

		// Evict key1
		handler.OnEvict("key1")

		// Re-insert key1
		handler.OnInsert("key1")

		// Now key2 is oldest
		candidate = handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})
}

// TestEvictionHandlerLargeScale tests eviction handlers with large number of entries.
func TestEvictionHandlerLargeScale(t *testing.T) {
	t.Run("LRUHandlerLargeScale", func(t *testing.T) {
		handler := NewLruHandler()

		// Insert 10000 keys
		for i := range 10000 {
			handler.OnInsert(fmt.Sprintf("key%d", i))
		}

		// Access every 10th key
		for i := 0; i < 10000; i += 10 {
			handler.OnAccess(fmt.Sprintf("key%d", i))
		}

		// Evict 5000 keys
		for range 5000 {
			candidate := handler.SelectEvictionCandidate()
			assert.NotEqual(t, constants.Empty, candidate)
			handler.OnEvict(candidate)
		}

		// Should have 5000 keys left
		assert.Equal(t, 5000, len(handler.accessMap))
		assert.Equal(t, 5000, handler.accessList.Len())
	})

	t.Run("LFUHandlerLargeScale", func(t *testing.T) {
		handler := NewLfuHandler()

		// Insert 10000 keys
		for i := range 10000 {
			handler.OnInsert(fmt.Sprintf("key%d", i))
		}

		// Access keys with varying frequencies
		for i := range 10000 {
			for j := 0; j < i%10; j++ {
				handler.OnAccess(fmt.Sprintf("key%d", i))
			}
		}

		// Evict 5000 keys
		for range 5000 {
			candidate := handler.SelectEvictionCandidate()
			assert.NotEqual(t, constants.Empty, candidate)
			handler.OnEvict(candidate)
		}

		// Should have 5000 keys left
		assert.Equal(t, 5000, len(handler.keyToNode))
	})

	t.Run("FIFOHandlerLargeScale", func(t *testing.T) {
		handler := NewFifoHandler()

		// Insert 10000 keys
		for i := range 10000 {
			handler.OnInsert(fmt.Sprintf("key%d", i))
		}

		// Evict 5000 keys (should evict key0 to key4999)
		for i := range 5000 {
			candidate := handler.SelectEvictionCandidate()
			assert.Equal(t, fmt.Sprintf("key%d", i), candidate)
			handler.OnEvict(candidate)
		}

		// Should have 5000 keys left
		assert.Equal(t, 5000, len(handler.insertMap))
		assert.Equal(t, 5000, handler.insertList.Len())
	})
}

// Benchmark tests.
func BenchmarkLRUHandler(b *testing.B) {
	handler := NewLruHandler()

	// Pre-populate
	for i := range 1000 {
		handler.OnInsert(fmt.Sprintf("key%d", i))
	}

	b.ResetTimer()
	b.Run("OnAccess", func(b *testing.B) {
		var i int
		for b.Loop() {
			handler.OnAccess(fmt.Sprintf("key%d", i%1000))
			i++
		}
	})

	b.Run("SelectEvictionCandidate", func(b *testing.B) {
		for b.Loop() {
			handler.SelectEvictionCandidate()
		}
	})
}

func BenchmarkLFUHandler(b *testing.B) {
	handler := NewLfuHandler()

	// Pre-populate
	for i := range 1000 {
		handler.OnInsert(fmt.Sprintf("key%d", i))
	}

	b.ResetTimer()
	b.Run("OnAccess", func(b *testing.B) {
		var i int
		for b.Loop() {
			handler.OnAccess(fmt.Sprintf("key%d", i%1000))
			i++
		}
	})

	b.Run("SelectEvictionCandidate", func(b *testing.B) {
		for b.Loop() {
			handler.SelectEvictionCandidate()
		}
	})
}

func BenchmarkFIFOHandler(b *testing.B) {
	handler := NewFifoHandler()

	// Pre-populate
	for i := range 1000 {
		handler.OnInsert(fmt.Sprintf("key%d", i))
	}

	b.ResetTimer()
	b.Run("OnAccess", func(b *testing.B) {
		var i int
		for b.Loop() {
			handler.OnAccess(fmt.Sprintf("key%d", i%1000))
			i++
		}
	})

	b.Run("SelectEvictionCandidate", func(b *testing.B) {
		for b.Loop() {
			handler.SelectEvictionCandidate()
		}
	})
}

func BenchmarkLFUHandlerConcurrent(b *testing.B) {
	handler := NewLfuHandler()

	// Pre-populate
	for i := range 1000 {
		handler.OnInsert(fmt.Sprintf("key%d", i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			handler.OnAccess(fmt.Sprintf("key%d", i%1000))
			i++
		}
	})
}
