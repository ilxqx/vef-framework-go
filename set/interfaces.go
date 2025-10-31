package set

import "iter"

// Set defines the interface for set data structures.
type Set[T comparable] interface {
	// Add adds one or more elements to the set.
	// Returns true if at least one element was added (wasn't already present).
	Add(elements ...T) bool
	// Remove removes one or more elements from the set.
	// Returns true if at least one element was removed (was present).
	Remove(elements ...T) bool
	// Contains checks if the set contains the specified element.
	Contains(element T) bool
	// ContainsAll checks if the set contains all specified elements.
	ContainsAll(elements ...T) bool
	// ContainsAny checks if the set contains any of the specified elements.
	ContainsAny(elements ...T) bool
	// Size returns the number of elements in the set.
	Size() int
	// IsEmpty returns true if the set contains no elements.
	IsEmpty() bool
	// Clear removes all elements from the set.
	Clear()
	// RemoveIf removes all elements that satisfy the predicate.
	// Returns the number of elements removed.
	RemoveIf(predicate func(element T) bool) int
	// Values returns all elements in the set as a slice.
	// For unordered sets (HashSet), the order is random.
	// For ordered sets (TreeSet), the order is sorted.
	Values() []T
	// Union returns a new set containing all elements from both sets.
	Union(other Set[T]) Set[T]
	// Intersection returns a new set containing only elements present in both sets.
	Intersection(other Set[T]) Set[T]
	// Difference returns a new set containing elements present in this set but not in the other.
	Difference(other Set[T]) Set[T]
	// SymmetricDifference returns a new set containing elements present in either set but not in both.
	SymmetricDifference(other Set[T]) Set[T]
	// IsSubset returns true if this set is a subset of the other set.
	IsSubset(other Set[T]) bool
	// IsSuperset returns true if this set is a superset of the other set.
	IsSuperset(other Set[T]) bool
	// Equal returns true if both sets contain exactly the same elements.
	Equal(other Set[T]) bool
	// Clone returns a shallow copy of the set.
	Clone() Set[T]
	// Each iterates over all elements in the set and calls the provided function.
	// Iteration stops if the function returns false.
	Each(func(element T) bool)
	// Seq returns an iterator sequence over the set.
	// Hash-based sets yield elements without a defined order, ordered sets respect their ordering.
	Seq() iter.Seq[T]
	// Filter returns a new set containing only elements that satisfy the predicate.
	Filter(predicate func(element T) bool) Set[T]
	// Map transforms each element using the provided function and returns a new set.
	// Note: The transform function must return a comparable type.
	Map(transform func(element T) T) Set[T]
	// Any returns true if at least one element satisfies the predicate.
	Any(predicate func(element T) bool) bool
	// All returns true if all elements satisfy the predicate.
	All(predicate func(element T) bool) bool
}

// OrderedSet extends Set with additional methods for ordered sets (like TreeSet).
type OrderedSet[T comparable] interface {
	Set[T]

	// Min returns the minimum element in the set.
	// Returns zero value and false if the set is empty.
	Min() (T, bool)
	// Max returns the maximum element in the set.
	// Returns zero value and false if the set is empty.
	Max() (T, bool)
	// Range returns elements between from (inclusive) and to (exclusive).
	Range(from, to T) []T
	// Iterator returns an iterator for traversing the set in order.
	Iterator() Iterator[T]
	// ReverseSeq returns a reverse-order iterator sequence over the set.
	ReverseSeq() iter.Seq[T]
	// EachIndexed iterates with the element index in sorted order.
	EachIndexed(func(index int, element T) bool)
	// SeqWithIndex returns a sequence yielding index-element pairs in ascending order.
	SeqWithIndex() iter.Seq2[int, T]
	// ReverseSeqWithIndex returns a sequence yielding index-element pairs in descending order.
	ReverseSeqWithIndex() iter.Seq2[int, T]
}

// Iterator defines the interface for iterating over set elements.
type Iterator[T any] interface {
	// Next moves the iterator to the next element.
	// Returns true if there is a next element, false otherwise.
	Next() bool
	// Prev moves the iterator to the previous element.
	// Returns true if there is a previous element, false otherwise.
	Prev() bool
	// Value returns the current element.
	// Should only be called after Next() or Prev() returns true.
	Value() T
	// Begin resets the iterator to the beginning (before the first element).
	// The next call to Next() will return the first element.
	Begin()
	// End resets the iterator to the end (after the last element).
	// The next call to Prev() will return the last element.
	End()
}
