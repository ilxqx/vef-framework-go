package set

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"iter"
	"slices"
	"sync/atomic"

	"github.com/puzpuzpuz/xsync/v4"
)

// SyncHashSet is a thread-safe unordered set implementation using xsync.Map.
// It provides O(1) average time complexity for add, remove, and contains operations.
// SyncHashSet is safe for concurrent use by multiple goroutines.
type SyncHashSet[T comparable] struct {
	data *xsync.Map[T, struct{}]
	size atomic.Int64
}

// Add adds one or more elements to the set.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Add(elements ...T) bool {
	added := false
	for _, element := range elements {
		if _, loaded := s.data.LoadOrStore(element, struct{}{}); !loaded {
			s.size.Add(1)
			added = true
		}
	}

	return added
}

// Remove removes one or more elements from the set.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Remove(elements ...T) bool {
	removed := false
	for _, element := range elements {
		if _, loaded := s.data.LoadAndDelete(element); loaded {
			s.size.Add(-1)
			removed = true
		}
	}

	return removed
}

// Contains checks if the set contains the specified element.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Contains(element T) bool {
	_, exists := s.data.Load(element)

	return exists
}

// ContainsAll checks if the set contains all specified elements.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) ContainsAll(elements ...T) bool {
	for _, element := range elements {
		if !s.Contains(element) {
			return false
		}
	}

	return true
}

// ContainsAny checks if the set contains any of the specified elements.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) ContainsAny(elements ...T) bool {
	return slices.ContainsFunc(elements, s.Contains)
}

// Size returns the number of elements in the set.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Size() int {
	return int(s.size.Load())
}

// IsEmpty returns true if the set contains no elements.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) IsEmpty() bool {
	return s.size.Load() == 0
}

// Clear removes all elements from the set.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Clear() {
	s.data.Clear()
	s.size.Store(0)
}

// RemoveIf removes all elements that satisfy the predicate.
// Returns the number of elements removed.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) RemoveIf(predicate func(element T) bool) int {
	count := 0
	s.data.Range(func(element T, _ struct{}) bool {
		if predicate(element) {
			if _, loaded := s.data.LoadAndDelete(element); loaded {
				s.size.Add(-1)
				count++
			}
		}

		return true
	})

	return count
}

// Values returns all elements in the set as a slice.
// The order is random and not guaranteed to be consistent.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Values() []T {
	values := make([]T, 0, s.Size())
	s.data.Range(func(element T, _ struct{}) bool {
		values = append(values, element)
		return true
	})

	return values
}

// Union returns a new set containing all elements from both sets.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Union(other Set[T]) Set[T] {
	result := NewSyncHashSet[T]()
	result.Add(s.Values()...)
	result.Add(other.Values()...)

	return result
}

// Intersection returns a new set containing only elements present in both sets.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Intersection(other Set[T]) Set[T] {
	result := NewSyncHashSet[T]()
	// Iterate over the smaller set for efficiency
	if other.Size() < s.Size() {
		for _, element := range other.Values() {
			if s.Contains(element) {
				result.Add(element)
			}
		}
	} else {
		s.data.Range(func(element T, _ struct{}) bool {
			if other.Contains(element) {
				result.Add(element)
			}

			return true
		})
	}

	return result
}

// Difference returns a new set containing elements present in this set but not in the other.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Difference(other Set[T]) Set[T] {
	result := NewSyncHashSet[T]()
	s.data.Range(func(element T, _ struct{}) bool {
		if !other.Contains(element) {
			result.Add(element)
		}

		return true
	})

	return result
}

// SymmetricDifference returns a new set containing elements present in either set but not in both.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) SymmetricDifference(other Set[T]) Set[T] {
	result := NewSyncHashSet[T]()
	s.data.Range(func(element T, _ struct{}) bool {
		if !other.Contains(element) {
			result.Add(element)
		}

		return true
	})

	for _, element := range other.Values() {
		if !s.Contains(element) {
			result.Add(element)
		}
	}

	return result
}

// IsSubset returns true if this set is a subset of the other set.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) IsSubset(other Set[T]) bool {
	if s.Size() > other.Size() {
		return false
	}

	isSubset := true
	s.data.Range(func(element T, _ struct{}) bool {
		if !other.Contains(element) {
			isSubset = false

			return false
		}

		return true
	})

	return isSubset
}

// IsSuperset returns true if this set is a superset of the other set.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) IsSuperset(other Set[T]) bool {
	return other.IsSubset(s)
}

// Equal returns true if both sets contain exactly the same elements.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Equal(other Set[T]) bool {
	if s.Size() != other.Size() {
		return false
	}

	isEqual := true
	s.data.Range(func(element T, _ struct{}) bool {
		if !other.Contains(element) {
			isEqual = false

			return false
		}

		return true
	})

	return isEqual
}

// Clone returns a shallow copy of the set.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Clone() Set[T] {
	result := NewSyncHashSet[T]()
	s.data.Range(func(element T, _ struct{}) bool {
		result.data.Store(element, struct{}{})
		result.size.Add(1)

		return true
	})

	return result
}

// Each iterates over all elements in the set and calls the provided function.
// Iteration stops if the function returns false.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Each(fn func(element T) bool) {
	s.data.Range(func(element T, _ struct{}) bool {
		return fn(element)
	})
}

// Seq returns an iterator sequence over the set in unspecified order.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Seq() iter.Seq[T] {
	return func(yield func(T) bool) {
		s.data.Range(func(element T, _ struct{}) bool {
			return yield(element)
		})
	}
}

// Filter returns a new set containing only elements that satisfy the predicate.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Filter(predicate func(element T) bool) Set[T] {
	result := NewSyncHashSet[T]()
	s.data.Range(func(element T, _ struct{}) bool {
		if predicate(element) {
			result.Add(element)
		}

		return true
	})

	return result
}

// Map transforms each element using the provided function and returns a new set.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Map(transform func(element T) T) Set[T] {
	result := NewSyncHashSet[T]()
	s.data.Range(func(element T, _ struct{}) bool {
		result.Add(transform(element))

		return true
	})

	return result
}

// Any returns true if at least one element satisfies the predicate.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) Any(predicate func(element T) bool) bool {
	found := false
	s.data.Range(func(element T, _ struct{}) bool {
		if predicate(element) {
			found = true

			return false
		}

		return true
	})

	return found
}

// All returns true if all elements satisfy the predicate.
// Thread-safe for concurrent use.
func (s *SyncHashSet[T]) All(predicate func(element T) bool) bool {
	allMatch := true
	s.data.Range(func(element T, _ struct{}) bool {
		if !predicate(element) {
			allMatch = false

			return false
		}

		return true
	})

	return allMatch
}

func (s *SyncHashSet[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Values())
}

func (s *SyncHashSet[T]) UnmarshalJSON(data []byte) error {
	if s.data == nil {
		s.data = xsync.NewMap[T, struct{}]()
	}

	var elements []T
	if err := json.Unmarshal(data, &elements); err != nil {
		return err
	}

	s.Clear()
	s.Add(elements...)

	return nil
}

func (s *SyncHashSet[T]) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(s.Values()); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *SyncHashSet[T]) GobDecode(data []byte) error {
	if s.data == nil {
		s.data = xsync.NewMap[T, struct{}]()
	}

	var elements []T
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&elements); err != nil {
		return err
	}

	s.Clear()
	s.Add(elements...)

	return nil
}
