package cache

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/constants"
)

func TestNoOpEvictionHandler(t *testing.T) {
	handler := NewNoOpEvictionHandler()
	require.NotNil(t, handler)

	t.Run("AllOperationsNoOp", func(_ *testing.T) {
		handler.OnAccess("key1")
		handler.OnInsert("key1")
		handler.OnEvict("key1")
		handler.Reset()
	})

	t.Run("AlwaysReturnEmptyCandidate", func(_ *testing.T) {
		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})

	t.Run("HandleMultipleOperations", func(_ *testing.T) {
		for i := range 100 {
			handler.OnInsert(fmt.Sprintf("key%d", i))
			handler.OnAccess(fmt.Sprintf("key%d", i))
		}

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})
}

func TestLRUHandler(t *testing.T) {
	t.Run("BasicInsertionAndEviction", func(_ *testing.T) {
		handler := NewLruHandler()
		require.NotNil(t, handler)

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("AccessUpdatesRecency", func(_ *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.OnAccess("key1")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("EvictionRemovesEntry", func(_ *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.OnEvict("key1")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("MultipleAccessesMaintainOrder", func(_ *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.OnAccess("key2")
		handler.OnAccess("key1")
		handler.OnAccess("key3")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("ResetClearsAllEntries", func(_ *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.Reset()

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})

	t.Run("EmptyHandlerReturnsEmptyCandidate", func(_ *testing.T) {
		handler := NewLruHandler()

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})

	t.Run("SingleEntry", func(_ *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("EvictNonExistentKey", func(_ *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")

		handler.OnEvict("key3")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("AccessNonExistentKeyCreatesEntry", func(_ *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnAccess("key2")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("ConcurrentOperations", func(_ *testing.T) {
		handler := NewLruHandler()

		var wg sync.WaitGroup

		for i := range 100 {
			wg.Go(func() {
				key := fmt.Sprintf("key%d", i%26)
				handler.OnInsert(key)
				handler.OnAccess(key)
			})
		}

		wg.Wait()

		candidate := handler.SelectEvictionCandidate()
		assert.NotEqual(t, constants.Empty, candidate)
	})

	t.Run("StressTestWithManyEntries", func(_ *testing.T) {
		handler := NewLruHandler()

		for i := range 1000 {
			handler.OnInsert(fmt.Sprintf("key%d", i))
		}

		for i := range 500 {
			handler.OnAccess(fmt.Sprintf("key%d", i*2))
		}

		candidate := handler.SelectEvictionCandidate()
		assert.NotEqual(t, constants.Empty, candidate)
	})
}

func TestFIFOHandler(t *testing.T) {
	t.Run("BasicInsertionAndEviction", func(_ *testing.T) {
		handler := NewFifoHandler()
		require.NotNil(t, handler)

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("AccessDoesNotAffectOrder", func(_ *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.OnAccess("key1")
		handler.OnAccess("key1")
		handler.OnAccess("key1")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("EvictionRemovesEntry", func(_ *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.OnEvict("key1")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("ResetClearsAllEntries", func(_ *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.Reset()

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})

	t.Run("EmptyHandlerReturnsEmptyCandidate", func(_ *testing.T) {
		handler := NewFifoHandler()

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})

	t.Run("SingleEntry", func(_ *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("DuplicateInsertIgnored", func(_ *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key1")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("EvictNonExistentKey", func(_ *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")

		handler.OnEvict("key3")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("ConcurrentOperations", func(_ *testing.T) {
		handler := NewFifoHandler()

		var wg sync.WaitGroup

		// Concurrent inserts
		for i := range 100 {
			wg.Go(func() {
				key := fmt.Sprintf("key%d", i)
				handler.OnInsert(key)
				handler.OnAccess(key)
			})
		}

		wg.Wait()

		candidate := handler.SelectEvictionCandidate()
		assert.NotEqual(t, constants.Empty, candidate)
	})
}

func TestLFUHandler(t *testing.T) {
	t.Run("BasicInsertionAndEviction", func(_ *testing.T) {
		handler := NewLfuHandler()
		require.NotNil(t, handler)

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("AccessIncreasesFrequency", func(_ *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.OnAccess("key1")
		handler.OnAccess("key1")
		handler.OnAccess("key2")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key3", candidate)
	})

	t.Run("EvictionRemovesEntry", func(_ *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.OnAccess("key1")
		handler.OnAccess("key1")
		handler.OnAccess("key2")

		handler.OnEvict("key3")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("TieBreakingByInsertionOrder", func(_ *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("FrequencyOrderingMaintained", func(_ *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.OnAccess("key2")
		handler.OnAccess("key3")
		handler.OnAccess("key3")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("ResetClearsAllEntries", func(_ *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.Reset()

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})

	t.Run("EmptyHandlerReturnsEmptyCandidate", func(_ *testing.T) {
		handler := NewLfuHandler()

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, constants.Empty, candidate)
	})

	t.Run("SingleEntry", func(_ *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("EvictNonExistentKey", func(_ *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")

		handler.OnEvict("key3")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})

	t.Run("AccessNonExistentKeyCreatesEntry", func(_ *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnAccess("key1")
		handler.OnAccess("key3")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("ConcurrentOperations", func(_ *testing.T) {
		handler := NewLfuHandler()

		var wg sync.WaitGroup

		for i := range 100 {
			wg.Go(func() {
				key := fmt.Sprintf("key%d", i%26)
				handler.OnInsert(key)

				for range i % 10 {
					handler.OnAccess(key)
				}
			})
		}

		wg.Wait()

		candidate := handler.SelectEvictionCandidate()
		_ = candidate
	})

	t.Run("StressTestWithManyEntries", func(_ *testing.T) {
		handler := NewLfuHandler()

		n := 1000
		for i := range n {
			handler.OnInsert(fmt.Sprintf("key%d", i))
		}

		for i := range n {
			for range i % 100 {
				handler.OnAccess(fmt.Sprintf("key%d", i))
			}
		}

		for range 100 {
			candidate := handler.SelectEvictionCandidate()
			require.NotEqual(t, constants.Empty, candidate)
			handler.OnEvict(candidate)
		}
	})

	t.Run("FrequencyBucketsWorkCorrectly", func(_ *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnAccess("key2")
		handler.OnInsert("key3")
		handler.OnAccess("key3")
		handler.OnAccess("key3")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)

		handler.OnEvict("key1")

		candidate = handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)

		handler.OnEvict("key2")

		candidate = handler.SelectEvictionCandidate()
		assert.Equal(t, "key3", candidate)
	})
}

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
		t.Run(fmt.Sprintf("policy_%d", tc.policy), func(_ *testing.T) {
			handler := factory.CreateHandler(tc.policy)
			require.NotNil(t, handler)

			typeName := fmt.Sprintf("%T", handler)
			assert.Equal(t, tc.expectedType, typeName)
		})
	}

	t.Run("InvalidPolicyDefaultsToNoOp", func(_ *testing.T) {
		handler := factory.CreateHandler(EvictionPolicy(999))
		require.NotNil(t, handler)

		typeName := fmt.Sprintf("%T", handler)
		assert.Equal(t, "*cache.NoOpEvictionHandler", typeName)
	})
}

func TestLRUHandlerUpdateBehavior(t *testing.T) {
	t.Run("UpdateMoveKeyToFront", func(_ *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)

		handler.OnAccess("key1")

		candidate = handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("RepeatedUpdatesDoNotCauseDuplicates", func(_ *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")

		for range 10 {
			handler.OnAccess("key1")
		}

		assert.Equal(t, 1, len(handler.accessMap))
		assert.Equal(t, 1, handler.accessList.Len())
	})

	t.Run("InterleavedInsertsAndAccesses", func(_ *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnAccess("key1")
		handler.OnInsert("key2")
		handler.OnAccess("key1")
		handler.OnInsert("key3")
		handler.OnAccess("key2")

		assert.Equal(t, 3, len(handler.accessMap))
		assert.Equal(t, 3, handler.accessList.Len())

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})
}

func TestLFUHandlerUpdateBehavior(t *testing.T) {
	t.Run("RepeatedUpdatesDoNotCauseDuplicates", func(_ *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")

		for range 10 {
			handler.OnAccess("key1")
		}

		assert.Equal(t, 1, len(handler.keyToNode))
		assert.Equal(t, 1, len(handler.keyToBucket))
	})

	t.Run("FrequencyIncrementsCorrectly", func(_ *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		for range 5 {
			handler.OnAccess("key1")
		}

		for range 3 {
			handler.OnAccess("key2")
		}

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key3", candidate)

		handler.OnEvict("key3")

		candidate = handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("FrequencyBucketMovement", func(_ *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		assert.Equal(t, int64(1), handler.minFreq)

		handler.OnAccess("key1")

		assert.Equal(t, int64(1), handler.minFreq)

		handler.OnEvict("key2")
		handler.OnEvict("key3")

		assert.Equal(t, int64(2), handler.minFreq)

		assert.Equal(t, 1, len(handler.keyToNode))
	})
}

func TestFIFOHandlerUpdateBehavior(t *testing.T) {
	t.Run("RepeatedUpdatesDoNotCauseDuplicates", func(_ *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")

		for range 10 {
			handler.OnAccess("key1")
		}

		assert.Equal(t, 1, len(handler.insertMap))
		assert.Equal(t, 1, handler.insertList.Len())
	})

	t.Run("AccessDoesNotChangeOrder", func(_ *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")
		handler.OnInsert("key3")

		handler.OnAccess("key3")
		handler.OnAccess("key1")
		handler.OnAccess("key2")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)
	})
}

func TestEvictionHandlerInternalConsistency(t *testing.T) {
	t.Run("LRUHandlerConsistency", func(_ *testing.T) {
		handler := NewLruHandler()

		for range 100 {
			handler.OnInsert("key1")
			handler.OnAccess("key1")
			handler.OnInsert("key2")
			handler.OnAccess("key2")
		}

		assert.Equal(t, 2, len(handler.accessMap))
		assert.Equal(t, 2, handler.accessList.Len())
	})

	t.Run("LFUHandlerConsistency", func(_ *testing.T) {
		handler := NewLfuHandler()

		for range 100 {
			handler.OnInsert("key1")
			handler.OnAccess("key1")
			handler.OnInsert("key2")
			handler.OnAccess("key2")
		}

		assert.Equal(t, 2, len(handler.keyToNode))
		assert.Equal(t, 2, len(handler.keyToBucket))
	})

	t.Run("FIFOHandlerConsistency", func(_ *testing.T) {
		handler := NewFifoHandler()

		for range 100 {
			handler.OnInsert("key1")
			handler.OnAccess("key1")
			handler.OnInsert("key2")
			handler.OnAccess("key2")
		}

		assert.Equal(t, 2, len(handler.insertMap))
		assert.Equal(t, 2, handler.insertList.Len())
	})
}

func TestEvictionHandlerEdgeCases(t *testing.T) {
	t.Run("LRUHandlerEvictAndReinsert", func(_ *testing.T) {
		handler := NewLruHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")

		handler.OnEvict("key1")
		assert.Equal(t, 1, len(handler.accessMap))

		handler.OnInsert("key1")
		assert.Equal(t, 2, len(handler.accessMap))

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("LFUHandlerEvictAndReinsert", func(_ *testing.T) {
		handler := NewLfuHandler()

		handler.OnInsert("key1")
		handler.OnAccess("key1")
		handler.OnAccess("key1")

		handler.OnInsert("key2")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)

		handler.OnEvict("key2")

		handler.OnInsert("key2")

		candidate = handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})

	t.Run("FIFOHandlerEvictAndReinsert", func(_ *testing.T) {
		handler := NewFifoHandler()

		handler.OnInsert("key1")
		handler.OnInsert("key2")

		candidate := handler.SelectEvictionCandidate()
		assert.Equal(t, "key1", candidate)

		handler.OnEvict("key1")

		handler.OnInsert("key1")

		candidate = handler.SelectEvictionCandidate()
		assert.Equal(t, "key2", candidate)
	})
}

func TestEvictionHandlerLargeScale(t *testing.T) {
	t.Run("LRUHandlerLargeScale", func(_ *testing.T) {
		handler := NewLruHandler()

		for i := range 10000 {
			handler.OnInsert(fmt.Sprintf("key%d", i))
		}

		for i := 0; i < 10000; i += 10 {
			handler.OnAccess(fmt.Sprintf("key%d", i))
		}

		for range 5000 {
			candidate := handler.SelectEvictionCandidate()
			assert.NotEqual(t, constants.Empty, candidate)
			handler.OnEvict(candidate)
		}

		assert.Equal(t, 5000, len(handler.accessMap))
		assert.Equal(t, 5000, handler.accessList.Len())
	})

	t.Run("LFUHandlerLargeScale", func(_ *testing.T) {
		handler := NewLfuHandler()

		for i := range 10000 {
			handler.OnInsert(fmt.Sprintf("key%d", i))
		}

		for i := range 10000 {
			for j := 0; j < i%10; j++ {
				handler.OnAccess(fmt.Sprintf("key%d", i))
			}
		}

		for range 5000 {
			candidate := handler.SelectEvictionCandidate()
			assert.NotEqual(t, constants.Empty, candidate)
			handler.OnEvict(candidate)
		}

		assert.Equal(t, 5000, len(handler.keyToNode))
	})

	t.Run("FIFOHandlerLargeScale", func(_ *testing.T) {
		handler := NewFifoHandler()

		for i := range 10000 {
			handler.OnInsert(fmt.Sprintf("key%d", i))
		}

		for i := range 5000 {
			candidate := handler.SelectEvictionCandidate()
			assert.Equal(t, fmt.Sprintf("key%d", i), candidate)
			handler.OnEvict(candidate)
		}

		assert.Equal(t, 5000, len(handler.insertMap))
		assert.Equal(t, 5000, handler.insertList.Len())
	})
}

func BenchmarkLRUHandler(b *testing.B) {
	handler := NewLruHandler()

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
