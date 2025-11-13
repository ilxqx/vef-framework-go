package set

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"iter"
	"slices"

	"github.com/emirpasic/gods/v2/sets/treeset"
)

// TreeSet is an ordered set implementation using a red-black tree from gods v2.
// Elements are kept sorted according to the provided comparator.
// It provides O(log n) time complexity for add, remove, and contains operations.
// TreeSet is not thread-safe.
type TreeSet[T comparable] struct {
	tree *treeset.Set[T]
	cmp  func(a, b T) int
}

// Add adds one or more elements to the set.
func (s *TreeSet[T]) Add(elements ...T) bool {
	added := false
	for _, element := range elements {
		if !s.tree.Contains(element) {
			s.tree.Add(element)

			added = true
		}
	}

	return added
}

// Remove removes one or more elements from the set.
func (s *TreeSet[T]) Remove(elements ...T) bool {
	removed := false
	for _, element := range elements {
		if s.tree.Contains(element) {
			s.tree.Remove(element)

			removed = true
		}
	}

	return removed
}

// Contains checks if the set contains the specified element.
func (s *TreeSet[T]) Contains(element T) bool {
	return s.tree.Contains(element)
}

// ContainsAll checks if the set contains all specified elements.
func (s *TreeSet[T]) ContainsAll(elements ...T) bool {
	for _, element := range elements {
		if !s.Contains(element) {
			return false
		}
	}

	return true
}

// ContainsAny checks if the set contains any of the specified elements.
func (s *TreeSet[T]) ContainsAny(elements ...T) bool {
	return slices.ContainsFunc(elements, s.Contains)
}

// Size returns the number of elements in the set.
func (s *TreeSet[T]) Size() int {
	return s.tree.Size()
}

// IsEmpty returns true if the set contains no elements.
func (s *TreeSet[T]) IsEmpty() bool {
	return s.tree.Empty()
}

// Clear removes all elements from the set.
func (s *TreeSet[T]) Clear() {
	s.tree.Clear()
}

// RemoveIf removes all elements that satisfy the predicate.
// Returns the number of elements removed.
func (s *TreeSet[T]) RemoveIf(predicate func(element T) bool) int {
	count := 0
	// Collect elements to remove first to avoid concurrent modification
	var toRemove []T

	it := s.tree.Iterator()
	for it.Next() {
		element := it.Value()
		if predicate(element) {
			toRemove = append(toRemove, element)
		}
	}

	for _, element := range toRemove {
		s.tree.Remove(element)

		count++
	}

	return count
}

// Values returns all elements in the set as a slice in sorted order.
func (s *TreeSet[T]) Values() []T {
	return s.tree.Values()
}

// Union returns a new set containing all elements from both sets.
func (s *TreeSet[T]) Union(other Set[T]) Set[T] {
	result := NewTreeSetWithComparator(s.cmp)
	result.Add(s.Values()...)
	result.Add(other.Values()...)

	return result
}

// Intersection returns a new set containing only elements present in both sets.
func (s *TreeSet[T]) Intersection(other Set[T]) Set[T] {
	result := NewTreeSetWithComparator(s.cmp)
	for _, element := range s.Values() {
		if other.Contains(element) {
			result.Add(element)
		}
	}

	return result
}

// Difference returns a new set containing elements present in this set but not in the other.
func (s *TreeSet[T]) Difference(other Set[T]) Set[T] {
	result := NewTreeSetWithComparator(s.cmp)
	for _, element := range s.Values() {
		if !other.Contains(element) {
			result.Add(element)
		}
	}

	return result
}

// SymmetricDifference returns a new set containing elements present in either set but not in both.
func (s *TreeSet[T]) SymmetricDifference(other Set[T]) Set[T] {
	result := NewTreeSetWithComparator(s.cmp)
	for _, element := range s.Values() {
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
func (s *TreeSet[T]) IsSubset(other Set[T]) bool {
	if s.Size() > other.Size() {
		return false
	}

	for _, element := range s.Values() {
		if !other.Contains(element) {
			return false
		}
	}

	return true
}

// IsSuperset returns true if this set is a superset of the other set.
func (s *TreeSet[T]) IsSuperset(other Set[T]) bool {
	return other.IsSubset(s)
}

// Equal returns true if both sets contain exactly the same elements.
func (s *TreeSet[T]) Equal(other Set[T]) bool {
	if s.Size() != other.Size() {
		return false
	}

	for _, element := range s.Values() {
		if !other.Contains(element) {
			return false
		}
	}

	return true
}

// Clone returns a shallow copy of the set.
func (s *TreeSet[T]) Clone() Set[T] {
	result := NewTreeSetWithComparator(s.cmp)
	result.Add(s.Values()...)

	return result
}

// Each iterates over all elements in the set in sorted order and calls the provided function.
// Iteration stops if the function returns false.
func (s *TreeSet[T]) Each(fn func(element T) bool) {
	it := s.tree.Iterator()
	for it.Next() {
		if !fn(it.Value()) {
			break
		}
	}
}

// EachIndexed iterates over elements in sorted order with their position.
func (s *TreeSet[T]) EachIndexed(fn func(index int, element T) bool) {
	it := s.tree.Iterator()

	index := 0
	for it.Next() {
		if !fn(index, it.Value()) {
			break
		}

		index++
	}
}

// Seq returns an iterator sequence over the set in ascending order.
func (s *TreeSet[T]) Seq() iter.Seq[T] {
	return func(yield func(T) bool) {
		it := s.tree.Iterator()
		for it.Next() {
			if !yield(it.Value()) {
				break
			}
		}
	}
}

// ReverseSeq returns an iterator sequence over the set in descending order.
func (s *TreeSet[T]) ReverseSeq() iter.Seq[T] {
	return func(yield func(T) bool) {
		it := s.tree.Iterator()
		it.End()

		for it.Prev() {
			if !yield(it.Value()) {
				break
			}
		}
	}
}

// SeqWithIndex returns an iterator sequence yielding index-element pairs in ascending order.
func (s *TreeSet[T]) SeqWithIndex() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		it := s.tree.Iterator()

		index := 0
		for it.Next() {
			if !yield(index, it.Value()) {
				break
			}

			index++
		}
	}
}

// ReverseSeqWithIndex returns an iterator sequence yielding index-element pairs in descending order.
func (s *TreeSet[T]) ReverseSeqWithIndex() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		it := s.tree.Iterator()
		it.End()

		index := 0
		for it.Prev() {
			if !yield(index, it.Value()) {
				break
			}

			index++
		}
	}
}

// Filter returns a new set containing only elements that satisfy the predicate.
func (s *TreeSet[T]) Filter(predicate func(element T) bool) Set[T] {
	result := NewTreeSetWithComparator(s.cmp)

	it := s.tree.Iterator()
	for it.Next() {
		element := it.Value()
		if predicate(element) {
			result.Add(element)
		}
	}

	return result
}

// Map transforms each element using the provided function and returns a new set.
func (s *TreeSet[T]) Map(transform func(element T) T) Set[T] {
	result := NewTreeSetWithComparator(s.cmp)

	it := s.tree.Iterator()
	for it.Next() {
		result.Add(transform(it.Value()))
	}

	return result
}

// Any returns true if at least one element satisfies the predicate.
func (s *TreeSet[T]) Any(predicate func(element T) bool) bool {
	it := s.tree.Iterator()
	for it.Next() {
		if predicate(it.Value()) {
			return true
		}
	}

	return false
}

// All returns true if all elements satisfy the predicate.
func (s *TreeSet[T]) All(predicate func(element T) bool) bool {
	it := s.tree.Iterator()
	for it.Next() {
		if !predicate(it.Value()) {
			return false
		}
	}

	return true
}

// Min returns the minimum element in the set.
// Returns zero value and false if the set is empty.
func (s *TreeSet[T]) Min() (T, bool) {
	var zero T
	if s.IsEmpty() {
		return zero, false
	}

	it := s.tree.Iterator()
	if it.Next() {
		return it.Value(), true
	}

	return zero, false
}

// Max returns the maximum element in the set.
// Returns zero value and false if the set is empty.
func (s *TreeSet[T]) Max() (T, bool) {
	var zero T
	if s.IsEmpty() {
		return zero, false
	}

	it := s.tree.Iterator()
	it.End()

	if it.Prev() {
		return it.Value(), true
	}

	return zero, false
}

// Range returns elements between from (inclusive) and to (exclusive).
func (s *TreeSet[T]) Range(from, to T) []T {
	var result []T

	it := s.tree.Iterator()
	for it.Next() {
		element := it.Value()
		cmpFrom := s.cmp(element, from)
		cmpTo := s.cmp(element, to)

		if cmpFrom >= 0 && cmpTo < 0 {
			result = append(result, element)
		} else if cmpTo >= 0 {
			break
		}
	}

	return result
}

// Iterator returns an iterator for traversing the set in sorted order.
func (s *TreeSet[T]) Iterator() Iterator[T] {
	return &treeSetIterator[T]{
		it: s.tree.Iterator(),
	}
}

func (s *TreeSet[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Values())
}

func (s *TreeSet[T]) UnmarshalJSON(data []byte) error {
	if s.tree == nil {
		panic("UnmarshalJSON called on uninitialized TreeSet - tree is nil. Use NewTreeSet() or NewTreeSetWithComparator() to create a TreeSet before unmarshaling")
	}

	if s.cmp == nil {
		panic("UnmarshalJSON called on TreeSet with nil comparator. Use NewTreeSet() or NewTreeSetWithComparator() to properly initialize the TreeSet")
	}

	var elements []T
	if err := json.Unmarshal(data, &elements); err != nil {
		return err
	}

	s.tree.Clear()
	s.Add(elements...)

	return nil
}

func (s *TreeSet[T]) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(s.Values()); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *TreeSet[T]) GobDecode(data []byte) error {
	if s.tree == nil {
		panic("GobDecode called on uninitialized TreeSet - tree is nil. Use NewTreeSet() or NewTreeSetWithComparator() to create a TreeSet before decoding")
	}

	if s.cmp == nil {
		panic("GobDecode called on TreeSet with nil comparator. Use NewTreeSet() or NewTreeSetWithComparator() to properly initialize the TreeSet")
	}

	var elements []T
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&elements); err != nil {
		return err
	}

	s.tree.Clear()
	s.Add(elements...)

	return nil
}

// treeSetIterator wraps the gods iterator to implement our Iterator interface.
type treeSetIterator[T comparable] struct {
	it treeset.Iterator[T]
}

// Next moves the iterator to the next element.
func (it *treeSetIterator[T]) Next() bool {
	return it.it.Next()
}

// Prev moves the iterator to the previous element.
func (it *treeSetIterator[T]) Prev() bool {
	return it.it.Prev()
}

// Value returns the current element.
func (it *treeSetIterator[T]) Value() T {
	return it.it.Value()
}

// Begin resets the iterator to the beginning (before the first element).
func (it *treeSetIterator[T]) Begin() {
	it.it.Begin()
}

// End resets the iterator to the end (after the last element).
func (it *treeSetIterator[T]) End() {
	it.it.End()
}
