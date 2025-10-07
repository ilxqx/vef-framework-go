package set

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncHashSet(t *testing.T) {
	t.Run("NewSyncHashSet", func(t *testing.T) {
		s := NewSyncHashSet[int]()
		assert.NotNil(t, s)
		assert.True(t, s.IsEmpty())
		assert.Equal(t, 0, s.Size())
	})

	t.Run("Add", func(t *testing.T) {
		s := NewSyncHashSet[int]()
		assert.True(t, s.Add(1, 2, 3))
		assert.Equal(t, 3, s.Size())
		assert.False(t, s.Add(1, 2))
		assert.Equal(t, 3, s.Size())
	})

	t.Run("Remove", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3, 4, 5})
		assert.True(t, s.Remove(2, 4))
		assert.Equal(t, 3, s.Size())
		assert.False(t, s.Remove(10))
	})

	t.Run("Contains", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3})
		assert.True(t, s.Contains(2))
		assert.False(t, s.Contains(5))
	})

	t.Run("ContainsAll", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3, 4, 5})
		assert.True(t, s.ContainsAll(1, 3, 5))
		assert.False(t, s.ContainsAll(1, 6))
	})

	t.Run("ContainsAny", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3})
		assert.True(t, s.ContainsAny(3, 4, 5))
		assert.False(t, s.ContainsAny(6, 7, 8))
	})

	t.Run("Clear", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3})
		s.Clear()
		assert.True(t, s.IsEmpty())
		assert.Equal(t, 0, s.Size())
	})

	t.Run("RemoveIf", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3, 4, 5, 6})
		count := s.RemoveIf(func(element int) bool {
			return element%2 == 0
		})
		assert.Equal(t, 3, count)
		assert.Equal(t, 3, s.Size())
		assert.True(t, s.ContainsAll(1, 3, 5))
		assert.False(t, s.ContainsAny(2, 4, 6))
	})

	t.Run("Values", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3})
		values := s.Values()
		assert.Len(t, values, 3)
		assert.ElementsMatch(t, []int{1, 2, 3}, values)
	})

	t.Run("Seq", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3})

		var values []int
		for element := range s.Seq() {
			values = append(values, element)
		}

		assert.Len(t, values, 3)
		assert.ElementsMatch(t, []int{1, 2, 3}, values)
	})

	t.Run("Union", func(t *testing.T) {
		s1 := NewSyncHashSetFromSlice([]int{1, 2, 3})
		s2 := NewSyncHashSetFromSlice([]int{3, 4, 5})
		result := s1.Union(s2)
		assert.Equal(t, 5, result.Size())
		assert.True(t, result.ContainsAll(1, 2, 3, 4, 5))
	})

	t.Run("Intersection", func(t *testing.T) {
		s1 := NewSyncHashSetFromSlice([]int{1, 2, 3, 4})
		s2 := NewSyncHashSetFromSlice([]int{3, 4, 5, 6})
		result := s1.Intersection(s2)
		assert.Equal(t, 2, result.Size())
		assert.True(t, result.ContainsAll(3, 4))
	})

	t.Run("Difference", func(t *testing.T) {
		s1 := NewSyncHashSetFromSlice([]int{1, 2, 3, 4})
		s2 := NewSyncHashSetFromSlice([]int{3, 4, 5, 6})
		result := s1.Difference(s2)
		assert.Equal(t, 2, result.Size())
		assert.True(t, result.ContainsAll(1, 2))
	})

	t.Run("SymmetricDifference", func(t *testing.T) {
		s1 := NewSyncHashSetFromSlice([]int{1, 2, 3})
		s2 := NewSyncHashSetFromSlice([]int{3, 4, 5})
		result := s1.SymmetricDifference(s2)
		assert.Equal(t, 4, result.Size())
		assert.True(t, result.ContainsAll(1, 2, 4, 5))
		assert.False(t, result.Contains(3))
	})

	t.Run("IsSubset", func(t *testing.T) {
		s1 := NewSyncHashSetFromSlice([]int{1, 2})
		s2 := NewSyncHashSetFromSlice([]int{1, 2, 3, 4})
		assert.True(t, s1.IsSubset(s2))
		assert.False(t, s2.IsSubset(s1))
	})

	t.Run("IsSuperset", func(t *testing.T) {
		s1 := NewSyncHashSetFromSlice([]int{1, 2, 3, 4})
		s2 := NewSyncHashSetFromSlice([]int{1, 2})
		assert.True(t, s1.IsSuperset(s2))
		assert.False(t, s2.IsSuperset(s1))
	})

	t.Run("Equal", func(t *testing.T) {
		s1 := NewSyncHashSetFromSlice([]int{1, 2, 3})
		s2 := NewSyncHashSetFromSlice([]int{3, 2, 1})
		s3 := NewSyncHashSetFromSlice([]int{1, 2, 4})

		assert.True(t, s1.Equal(s2))
		assert.False(t, s1.Equal(s3))
	})

	t.Run("Clone", func(t *testing.T) {
		s1 := NewSyncHashSetFromSlice([]int{1, 2, 3})
		s2 := s1.Clone()
		assert.True(t, s1.Equal(s2))
		s2.Add(4)
		assert.False(t, s1.Equal(s2))
		assert.Equal(t, 3, s1.Size())
	})

	t.Run("Each", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3, 4, 5})
		sum := 0
		s.Each(func(element int) bool {
			sum += element

			return true
		})
		assert.Equal(t, 15, sum)
	})

	t.Run("Each with early termination", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3, 4, 5})
		count := 0
		s.Each(func(element int) bool {
			count++

			return count < 3
		})
		assert.Equal(t, 3, count)
	})

	t.Run("Filter", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3, 4, 5, 6})
		result := s.Filter(func(element int) bool {
			return element%2 == 0
		})
		assert.Equal(t, 3, result.Size())
		assert.True(t, result.ContainsAll(2, 4, 6))
	})

	t.Run("Map", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3})
		result := s.Map(func(element int) int {
			return element * 2
		})
		assert.Equal(t, 3, result.Size())
		assert.True(t, result.ContainsAll(2, 4, 6))
	})

	t.Run("Any", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3, 4, 5})
		assert.True(t, s.Any(func(element int) bool {
			return element > 3
		}))
		assert.False(t, s.Any(func(element int) bool {
			return element > 10
		}))
	})

	t.Run("All", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{2, 4, 6, 8})
		assert.True(t, s.All(func(element int) bool {
			return element%2 == 0
		}))
		assert.False(t, s.All(func(element int) bool {
			return element > 5
		}))
	})

	t.Run("JSON serialization", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3})
		data, err := json.Marshal(s)
		require.NoError(t, err)

		decoded := NewSyncHashSet[int]()
		require.NoError(t, json.Unmarshal(data, decoded))
		assert.True(t, s.Equal(decoded))
	})

	t.Run("Gob serialization", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3})

		var buf bytes.Buffer

		require.NoError(t, gob.NewEncoder(&buf).Encode(s))

		decoded := NewSyncHashSet[int]()
		require.NoError(t, gob.NewDecoder(&buf).Decode(decoded))
		assert.True(t, s.Equal(decoded))
	})

	t.Run("Empty set serialization", func(t *testing.T) {
		s := NewSyncHashSet[int]()

		// JSON
		data, err := json.Marshal(s)
		require.NoError(t, err)
		assert.Equal(t, "[]", string(data))

		decodedJSON := NewSyncHashSet[int]()
		require.NoError(t, json.Unmarshal(data, decodedJSON))
		assert.True(t, decodedJSON.IsEmpty())
	})

	t.Run("UnmarshalJSON with invalid input", func(t *testing.T) {
		s := NewSyncHashSet[int]()
		err := json.Unmarshal([]byte("not valid json"), s)
		assert.Error(t, err)
	})

	t.Run("Zero value deserialization - JSON", func(t *testing.T) {
		// SyncHashSet supports zero value deserialization
		var s SyncHashSet[int]
		require.NoError(t, json.Unmarshal([]byte("[1,2,3]"), &s))
		assert.Equal(t, 3, s.Size())
		assert.True(t, s.ContainsAll(1, 2, 3))
	})

	t.Run("Zero value deserialization - Gob", func(t *testing.T) {
		// First create a set and encode it
		original := NewSyncHashSetFromSlice([]int{1, 2, 3})

		var buf bytes.Buffer
		require.NoError(t, gob.NewEncoder(&buf).Encode(original.Values()))

		// Decode into zero value SyncHashSet
		var s SyncHashSet[int]
		require.NoError(t, s.GobDecode(buf.Bytes()))
		assert.Equal(t, 3, s.Size())
		assert.True(t, s.ContainsAll(1, 2, 3))
	})
}

// TestSyncHashSetConcurrency tests thread safety of SyncHashSet.
func TestSyncHashSetConcurrency(t *testing.T) {
	t.Run("Concurrent Add", func(t *testing.T) {
		s := NewSyncHashSet[int]()

		var wg sync.WaitGroup

		goroutines := 100
		itemsPerGoroutine := 100

		for i := range goroutines {
			wg.Go(func() {
				for j := range itemsPerGoroutine {
					s.Add(i*itemsPerGoroutine + j)
				}
			})
		}

		wg.Wait()
		assert.Equal(t, goroutines*itemsPerGoroutine, s.Size())
	})

	t.Run("Concurrent Add and Remove", func(t *testing.T) {
		s := NewSyncHashSet[int]()

		var wg sync.WaitGroup

		goroutines := 50

		// Pre-populate with some elements
		for i := range 1000 {
			s.Add(i)
		}

		// Half goroutines add, half remove
		for i := range goroutines {
			if i%2 == 0 {
				wg.Go(func() {
					for j := range 100 {
						s.Add(i*1000 + j)
					}
				})
			} else {
				wg.Go(func() {
					for j := range 100 {
						s.Remove(j)
					}
				})
			}
		}

		wg.Wait()
		// Just verify no panic occurred and size is reasonable
		assert.True(t, s.Size() >= 0)
	})

	t.Run("Concurrent Contains", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

		var wg sync.WaitGroup

		goroutines := 100

		for range goroutines {
			wg.Go(func() {
				for j := 1; j <= 10; j++ {
					s.Contains(j)
				}
			})
		}

		wg.Wait()
		assert.Equal(t, 10, s.Size())
	})

	t.Run("Concurrent Values", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3, 4, 5})

		var wg sync.WaitGroup

		goroutines := 50

		for range goroutines {
			wg.Go(func() {
				values := s.Values()
				assert.NotNil(t, values)
				assert.True(t, len(values) >= 0)
			})
		}

		wg.Wait()
	})

	t.Run("Concurrent Clear and Add", func(t *testing.T) {
		s := NewSyncHashSet[int]()

		var wg sync.WaitGroup

		goroutines := 10

		for i := range goroutines {
			wg.Go(func() {
				if i%2 == 0 {
					s.Clear()
				} else {
					for j := range 100 {
						s.Add(j)
					}
				}
			})
		}

		wg.Wait()
		// Just verify no panic occurred
		assert.True(t, s.Size() >= 0)
	})

	t.Run("Concurrent Each", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

		var wg sync.WaitGroup

		goroutines := 50

		for range goroutines {
			wg.Go(func() {
				count := 0
				s.Each(func(element int) bool {
					count++

					return true
				})
				assert.True(t, count >= 0)
			})
		}

		wg.Wait()
	})

	t.Run("Concurrent RemoveIf", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

		var wg sync.WaitGroup

		goroutines := 10

		for i := range goroutines {
			wg.Go(func() {
				s.RemoveIf(func(element int) bool {
					return element%2 == i%2
				})
			})
		}

		wg.Wait()
		// Just verify no panic occurred
		assert.True(t, s.Size() >= 0)
	})

	t.Run("Concurrent Clone", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3, 4, 5})

		var wg sync.WaitGroup

		goroutines := 50

		for range goroutines {
			wg.Go(func() {
				cloned := s.Clone()
				assert.NotNil(t, cloned)
				assert.True(t, cloned.Size() >= 0)
			})
		}

		wg.Wait()
	})

	t.Run("Concurrent Set Operations", func(t *testing.T) {
		s1 := NewSyncHashSetFromSlice([]int{1, 2, 3, 4, 5})
		s2 := NewSyncHashSetFromSlice([]int{4, 5, 6, 7, 8})

		var wg sync.WaitGroup

		goroutines := 20

		for i := range goroutines {
			wg.Go(func() {
				switch i % 4 {
				case 0:
					s1.Union(s2)
				case 1:
					s1.Intersection(s2)
				case 2:
					s1.Difference(s2)
				case 3:
					s1.SymmetricDifference(s2)
				}
			})
		}

		wg.Wait()
		// Verify original sets are unchanged
		assert.Equal(t, 5, s1.Size())
		assert.Equal(t, 5, s2.Size())
	})

	t.Run("Concurrent Serialization", func(t *testing.T) {
		s := NewSyncHashSetFromSlice([]int{1, 2, 3, 4, 5})

		var wg sync.WaitGroup

		goroutines := 50

		for i := range goroutines {
			wg.Go(func() {
				switch i % 2 {
				case 0:
					_, _ = json.Marshal(s)
				case 1:
					_, _ = s.GobEncode()
				}
			})
		}

		wg.Wait()
		assert.Equal(t, 5, s.Size())
	})
}
