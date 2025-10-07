package set

import (
	"cmp"

	"github.com/puzpuzpuz/xsync/v4"

	godstreeset "github.com/emirpasic/gods/v2/sets/treeset"
)

// NewHashSet creates a new empty HashSet.
func NewHashSet[T comparable]() *HashSet[T] {
	return &HashSet[T]{
		data: make(map[T]struct{}),
	}
}

// NewHashSetWithCapacity creates a new HashSet with the specified initial capacity.
func NewHashSetWithCapacity[T comparable](capacity int) *HashSet[T] {
	return &HashSet[T]{
		data: make(map[T]struct{}, capacity),
	}
}

// NewHashSetFromSlice creates a new HashSet from the given slice.
func NewHashSetFromSlice[T comparable](elements []T) *HashSet[T] {
	s := NewHashSetWithCapacity[T](len(elements))
	s.Add(elements...)

	return s
}

// NewTreeSet creates a new empty TreeSet using the default comparator for ordered types.
// T must satisfy cmp.Ordered constraint (implements < operator).
func NewTreeSet[T cmp.Ordered]() *TreeSet[T] {
	return &TreeSet[T]{
		tree: godstreeset.New[T](),
		cmp:  cmp.Compare[T],
	}
}

// NewTreeSetWithComparator creates a new empty TreeSet with a custom comparator.
// The comparator should return:
//   - negative value if a < b
//   - zero if a == b
//   - positive value if a > b
func NewTreeSetWithComparator[T comparable](comparator func(a, b T) int) *TreeSet[T] {
	return &TreeSet[T]{
		tree: godstreeset.NewWith(comparator),
		cmp:  comparator,
	}
}

// NewTreeSetFromSlice creates a new TreeSet from the given slice using the default comparator.
func NewTreeSetFromSlice[T cmp.Ordered](elements []T) *TreeSet[T] {
	s := NewTreeSet[T]()
	s.Add(elements...)

	return s
}

// NewTreeSetFromSliceWithComparator creates a new TreeSet from the given slice with a custom comparator.
func NewTreeSetFromSliceWithComparator[T comparable](elements []T, comparator func(a, b T) int) *TreeSet[T] {
	s := NewTreeSetWithComparator(comparator)
	s.Add(elements...)

	return s
}

// NewSyncHashSet creates a new empty thread-safe SyncHashSet.
func NewSyncHashSet[T comparable]() *SyncHashSet[T] {
	return &SyncHashSet[T]{
		data: xsync.NewMap[T, struct{}](),
	}
}

// NewSyncHashSetFromSlice creates a new thread-safe SyncHashSet from the given slice.
func NewSyncHashSetFromSlice[T comparable](elements []T) *SyncHashSet[T] {
	s := NewSyncHashSet[T]()
	s.Add(elements...)

	return s
}
