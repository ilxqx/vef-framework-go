package set

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"iter"
	"slices"
)

// HashSet is an unordered set implementation using a native Go map.
// It provides O(1) average time complexity for add, remove, and contains operations.
// HashSet is not thread-safe.
type HashSet[T comparable] struct {
	data map[T]struct{}
}

// Add adds one or more elements to the set.
func (s *HashSet[T]) Add(elements ...T) bool {
	added := false
	for _, element := range elements {
		if _, exists := s.data[element]; !exists {
			s.data[element] = struct{}{}
			added = true
		}
	}

	return added
}

// Remove removes one or more elements from the set.
func (s *HashSet[T]) Remove(elements ...T) bool {
	removed := false
	for _, element := range elements {
		if _, exists := s.data[element]; exists {
			delete(s.data, element)

			removed = true
		}
	}

	return removed
}

// Contains checks if the set contains the specified element.
func (s *HashSet[T]) Contains(element T) bool {
	_, exists := s.data[element]

	return exists
}

// ContainsAll checks if the set contains all specified elements.
func (s *HashSet[T]) ContainsAll(elements ...T) bool {
	for _, element := range elements {
		if !s.Contains(element) {
			return false
		}
	}

	return true
}

// ContainsAny checks if the set contains any of the specified elements.
func (s *HashSet[T]) ContainsAny(elements ...T) bool {
	return slices.ContainsFunc(elements, s.Contains)
}

// Size returns the number of elements in the set.
func (s *HashSet[T]) Size() int {
	return len(s.data)
}

// IsEmpty returns true if the set contains no elements.
func (s *HashSet[T]) IsEmpty() bool {
	return len(s.data) == 0
}

// Clear removes all elements from the set.
func (s *HashSet[T]) Clear() {
	s.data = make(map[T]struct{})
}

// RemoveIf removes all elements that satisfy the predicate.
// Returns the number of elements removed.
func (s *HashSet[T]) RemoveIf(predicate func(element T) bool) int {
	count := 0
	for element := range s.data {
		if predicate(element) {
			delete(s.data, element)

			count++
		}
	}

	return count
}

// Values returns all elements in the set as a slice.
// The order is random and not guaranteed to be consistent.
func (s *HashSet[T]) Values() []T {
	values := make([]T, 0, len(s.data))
	for element := range s.data {
		values = append(values, element)
	}

	return values
}

// Union returns a new set containing all elements from both sets.
func (s *HashSet[T]) Union(other Set[T]) Set[T] {
	result := NewHashSetWithCapacity[T](s.Size() + other.Size())
	result.Add(s.Values()...)
	result.Add(other.Values()...)

	return result
}

// Intersection returns a new set containing only elements present in both sets.
func (s *HashSet[T]) Intersection(other Set[T]) Set[T] {
	result := NewHashSet[T]()
	// Iterate over the smaller set for efficiency
	if other.Size() < s.Size() {
		for _, element := range other.Values() {
			if s.Contains(element) {
				result.Add(element)
			}
		}
	} else {
		for element := range s.data {
			if other.Contains(element) {
				result.Add(element)
			}
		}
	}

	return result
}

// Difference returns a new set containing elements present in this set but not in the other.
func (s *HashSet[T]) Difference(other Set[T]) Set[T] {
	result := NewHashSet[T]()
	for element := range s.data {
		if !other.Contains(element) {
			result.Add(element)
		}
	}

	return result
}

// SymmetricDifference returns a new set containing elements present in either set but not in both.
func (s *HashSet[T]) SymmetricDifference(other Set[T]) Set[T] {
	result := NewHashSet[T]()
	for element := range s.data {
		if !other.Contains(element) {
			result.Add(element)
		}
	}

	for _, element := range other.Values() {
		if !s.Contains(element) {
			result.Add(element)
		}
	}

	return result
}

// IsSubset returns true if this set is a subset of the other set.
func (s *HashSet[T]) IsSubset(other Set[T]) bool {
	if s.Size() > other.Size() {
		return false
	}

	for element := range s.data {
		if !other.Contains(element) {
			return false
		}
	}

	return true
}

// IsSuperset returns true if this set is a superset of the other set.
func (s *HashSet[T]) IsSuperset(other Set[T]) bool {
	return other.IsSubset(s)
}

// Equal returns true if both sets contain exactly the same elements.
func (s *HashSet[T]) Equal(other Set[T]) bool {
	if s.Size() != other.Size() {
		return false
	}

	for element := range s.data {
		if !other.Contains(element) {
			return false
		}
	}

	return true
}

// Clone returns a shallow copy of the set.
func (s *HashSet[T]) Clone() Set[T] {
	result := NewHashSetWithCapacity[T](s.Size())
	for element := range s.data {
		result.data[element] = struct{}{}
	}

	return result
}

// Each iterates over all elements in the set and calls the provided function.
// Iteration stops if the function returns false.
func (s *HashSet[T]) Each(fn func(element T) bool) {
	for element := range s.data {
		if !fn(element) {
			break
		}
	}
}

// Seq returns an iterator sequence over the set in unspecified order.
func (s *HashSet[T]) Seq() iter.Seq[T] {
	return func(yield func(T) bool) {
		for element := range s.data {
			if !yield(element) {
				break
			}
		}
	}
}

// Filter returns a new set containing only elements that satisfy the predicate.
func (s *HashSet[T]) Filter(predicate func(element T) bool) Set[T] {
	result := NewHashSet[T]()
	for element := range s.data {
		if predicate(element) {
			result.Add(element)
		}
	}

	return result
}

// Map transforms each element using the provided function and returns a new set.
func (s *HashSet[T]) Map(transform func(element T) T) Set[T] {
	result := NewHashSetWithCapacity[T](s.Size())
	for element := range s.data {
		result.Add(transform(element))
	}

	return result
}

// Any returns true if at least one element satisfies the predicate.
func (s *HashSet[T]) Any(predicate func(element T) bool) bool {
	for element := range s.data {
		if predicate(element) {
			return true
		}
	}

	return false
}

// All returns true if all elements satisfy the predicate.
func (s *HashSet[T]) All(predicate func(element T) bool) bool {
	for element := range s.data {
		if !predicate(element) {
			return false
		}
	}

	return true
}

func (s *HashSet[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Values())
}

func (s *HashSet[T]) UnmarshalJSON(data []byte) error {
	var elements []T
	if err := json.Unmarshal(data, &elements); err != nil {
		return err
	}

	s.Clear()
	s.Add(elements...)

	return nil
}

func (s *HashSet[T]) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(s.Values()); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *HashSet[T]) GobDecode(data []byte) error {
	var elements []T
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&elements); err != nil {
		return err
	}

	s.Clear()
	s.Add(elements...)

	return nil
}
