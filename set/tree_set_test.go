package set

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"testing"

	"github.com/emirpasic/gods/v2/sets/treeset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTreeSetNew(t *testing.T) {
	s := NewTreeSet[int]()

	assert.NotNil(t, s, "TreeSet should be initialized")
	assert.True(t, s.IsEmpty(), "New TreeSet should be empty")
	assert.Equal(t, 0, s.Size(), "Size should be 0")
}

func TestTreeSetAddMaintainsOrder(t *testing.T) {
	s := NewTreeSet[int]()

	s.Add(3, 1, 4, 1, 5, 9, 2, 6)

	values := s.Values()
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 9}, values, "Values should be sorted and unique")
}

func TestTreeSetCustomComparator(t *testing.T) {
	s := NewTreeSetWithComparator(func(a, b int) int {
		if a < b {
			return 1
		} else if a > b {
			return -1
		}

		return 0
	})

	s.Add(1, 2, 3, 4, 5)

	values := s.Values()
	assert.Equal(t, []int{5, 4, 3, 2, 1}, values, "Values should be sorted in descending order")
}

func TestTreeSetRemove(t *testing.T) {
	s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5})

	removed := s.Remove(2, 5)
	assert.True(t, removed, "Should return true when removing existing elements")
	assert.Equal(t, []int{1, 3, 4}, s.Values(), "Remaining elements should be correct")

	removed = s.Remove(8)
	assert.False(t, removed, "Should return false when removing non-existent element")
}

func TestTreeSetContains(t *testing.T) {
	s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5})

	tests := []struct {
		name     string
		method   string
		elements []int
		expected bool
	}{
		{"ContainsExisting", "Contains", []int{3}, true},
		{"ContainsNonExisting", "Contains", []int{9}, false},
		{"ContainsAllExisting", "ContainsAll", []int{1, 4, 5}, true},
		{"ContainsAllPartial", "ContainsAll", []int{1, 6}, false},
		{"ContainsAnySome", "ContainsAny", []int{0, 5, 10}, true},
		{"ContainsAnyNone", "ContainsAny", []int{-1, 6}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bool
			switch tt.method {
			case "Contains":
				result = s.Contains(tt.elements[0])
			case "ContainsAll":
				result = s.ContainsAll(tt.elements...)
			case "ContainsAny":
				result = s.ContainsAny(tt.elements...)
			}

			assert.Equal(t, tt.expected, result, "Method should return correct result")
		})
	}
}

func TestTreeSetSizeAndEmpty(t *testing.T) {
	s := NewTreeSet[int]()

	assert.True(t, s.IsEmpty(), "New set should be empty")

	s.Add(1)

	assert.False(t, s.IsEmpty(), "Set should not be empty after adding")
	assert.Equal(t, 1, s.Size(), "Size should be 1")
}

func TestTreeSetClear(t *testing.T) {
	s := NewTreeSetFromSlice([]int{1, 2, 3})

	s.Clear()

	assert.True(t, s.IsEmpty(), "Set should be empty after Clear")
}

func TestTreeSetRemoveIf(t *testing.T) {
	s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8})

	count := s.RemoveIf(func(element int) bool {
		return element%2 == 0
	})

	assert.Equal(t, 4, count, "Should remove 4 even elements")
	assert.Equal(t, []int{1, 3, 5, 7}, s.Values(), "Should contain only odd elements")
}

func TestTreeSetValues(t *testing.T) {
	s := NewTreeSetFromSlice([]int{3, 1, 2})

	values := s.Values()

	assert.Equal(t, []int{1, 2, 3}, values, "Values should be sorted")
}

func TestTreeSetSeq(t *testing.T) {
	s := NewTreeSetFromSlice([]int{3, 1, 4, 1, 5, 9, 2, 6})

	var values []int
	for element := range s.Seq() {
		values = append(values, element)
	}

	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 9}, values, "Seq should iterate in sorted order")
}

func TestTreeSetSeqWithIndex(t *testing.T) {
	s := NewTreeSetFromSlice([]int{3, 1, 4, 1, 5, 9, 2})

	var (
		positions []int
		values    []int
	)

	for index, element := range s.SeqWithIndex() {
		positions = append(positions, index)
		values = append(values, element)
	}

	assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, positions, "Indices should be sequential")
	assert.Equal(t, []int{1, 2, 3, 4, 5, 9}, values, "Values should be sorted")
}

func TestTreeSetReverseSeq(t *testing.T) {
	s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5})

	var values []int
	for element := range s.ReverseSeq() {
		values = append(values, element)
	}

	assert.Equal(t, []int{5, 4, 3, 2, 1}, values, "ReverseSeq should iterate in reverse order")
}

func TestTreeSetReverseSeqWithIndex(t *testing.T) {
	s := NewTreeSetFromSlice([]int{1, 2, 3, 4})

	var (
		positions []int
		values    []int
	)

	for index, element := range s.ReverseSeqWithIndex() {
		positions = append(positions, index)
		values = append(values, element)
	}

	assert.Equal(t, []int{0, 1, 2, 3}, positions, "Indices should be sequential")
	assert.Equal(t, []int{4, 3, 2, 1}, values, "Values should be in reverse order")
}

func TestTreeSetEach(t *testing.T) {
	t.Run("CompleteIteration", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{3, 1, 4, 1, 5, 9, 2, 6})

		var values []int
		s.Each(func(element int) bool {
			values = append(values, element)

			return true
		})

		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 9}, values, "Each should iterate in sorted order")
	})

	t.Run("EachIndexed", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{3, 1, 4, 1, 5, 9, 2, 6})

		var (
			positions []int
			values    []int
		)

		s.EachIndexed(func(index, element int) bool {
			positions = append(positions, index)
			values = append(values, element)

			return true
		})

		assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6}, positions, "Indices should be sequential")
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 9}, values, "Values should be sorted")
	})
}

func TestTreeSetUnion(t *testing.T) {
	s1 := NewTreeSetFromSlice([]int{1, 2, 3})
	s2 := NewTreeSetFromSlice([]int{3, 4, 5})

	result := s1.Union(s2)

	assert.Equal(t, []int{1, 2, 3, 4, 5}, result.Values(), "Union should contain all elements in sorted order")
}

func TestTreeSetIntersection(t *testing.T) {
	s1 := NewTreeSetFromSlice([]int{1, 2, 3, 4})
	s2 := NewTreeSetFromSlice([]int{3, 4, 5, 6})

	result := s1.Intersection(s2)

	assert.Equal(t, []int{3, 4}, result.Values(), "Intersection should contain common elements")
}

func TestTreeSetDifference(t *testing.T) {
	s1 := NewTreeSetFromSlice([]int{1, 2, 3, 4})
	s2 := NewTreeSetFromSlice([]int{3, 4, 5, 6})

	result := s1.Difference(s2)

	assert.Equal(t, []int{1, 2}, result.Values(), "Difference should contain elements only in first set")
}

func TestTreeSetSymmetricDifference(t *testing.T) {
	s1 := NewTreeSetFromSlice([]int{1, 2, 3})
	s2 := NewTreeSetFromSlice([]int{3, 4, 5})

	result := s1.SymmetricDifference(s2)

	assert.Equal(t, []int{1, 2, 4, 5}, result.Values(), "SymmetricDifference should contain non-common elements")
}

func TestTreeSetSubsetAndSuperset(t *testing.T) {
	s1 := NewTreeSetFromSlice([]int{1, 2})
	s2 := NewTreeSetFromSlice([]int{1, 2, 3, 4})

	assert.True(t, s1.IsSubset(s2), "s1 should be subset of s2")
	assert.False(t, s2.IsSubset(s1), "s2 should not be subset of s1")
	assert.True(t, s2.IsSuperset(s1), "s2 should be superset of s1")
	assert.False(t, s1.IsSuperset(s2), "s1 should not be superset of s2")
}

func TestTreeSetEqual(t *testing.T) {
	s1 := NewTreeSetFromSlice([]int{1, 2, 3})
	s2 := NewTreeSetFromSlice([]int{3, 2, 1})
	s3 := NewTreeSetFromSlice([]int{1, 2, 4})

	assert.True(t, s1.Equal(s2), "Sets with same elements should be equal")
	assert.False(t, s1.Equal(s3), "Sets with different elements should not be equal")
}

func TestTreeSetClone(t *testing.T) {
	s1 := NewTreeSetFromSlice([]int{1, 2, 3})

	s2 := s1.Clone()

	assert.Equal(t, []int{1, 2, 3}, s2.Values(), "Clone should have same values")

	s2.Add(4)

	assert.Equal(t, []int{1, 2, 3}, s1.Values(), "Original should remain unchanged")
	assert.Equal(t, []int{1, 2, 3, 4}, s2.Values(), "Clone should have new element")
}

func TestTreeSetFilter(t *testing.T) {
	s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5, 6})

	result := s.Filter(func(element int) bool {
		return element%2 == 0
	})

	assert.Equal(t, []int{2, 4, 6}, result.Values(), "Filtered set should contain only even elements")
}

func TestTreeSetMap(t *testing.T) {
	s := NewTreeSetFromSlice([]int{1, 2, 3})

	result := s.Map(func(element int) int {
		return element * 2
	})

	assert.Equal(t, []int{2, 4, 6}, result.Values(), "Mapped set should contain doubled elements")
}

func TestTreeSetAny(t *testing.T) {
	s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5})

	tests := []struct {
		name      string
		predicate func(int) bool
		expected  bool
	}{
		{
			"SomeMatch",
			func(element int) bool { return element > 3 },
			true,
		},
		{
			"NoneMatch",
			func(element int) bool { return element > 10 },
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.Any(tt.predicate)
			assert.Equal(t, tt.expected, result, "Any should return correct result")
		})
	}
}

func TestTreeSetAll(t *testing.T) {
	s := NewTreeSetFromSlice([]int{2, 4, 6, 8})

	tests := []struct {
		name      string
		predicate func(int) bool
		expected  bool
	}{
		{
			"AllMatch",
			func(element int) bool { return element%2 == 0 },
			true,
		},
		{
			"NotAllMatch",
			func(element int) bool { return element > 5 },
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.All(tt.predicate)
			assert.Equal(t, tt.expected, result, "All should return correct result")
		})
	}
}

func TestTreeSetMinAndMax(t *testing.T) {
	t.Run("NonEmptySet", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{3, 1, 4, 1, 5, 9, 2, 6})

		min, ok := s.Min()
		assert.True(t, ok, "Min should return true for non-empty set")
		assert.Equal(t, 1, min, "Min should be 1")

		max, ok := s.Max()
		assert.True(t, ok, "Max should return true for non-empty set")
		assert.Equal(t, 9, max, "Max should be 9")
	})

	t.Run("EmptySet", func(t *testing.T) {
		empty := NewTreeSet[int]()

		_, ok := empty.Min()
		assert.False(t, ok, "Min should return false for empty set")

		_, ok = empty.Max()
		assert.False(t, ok, "Max should return false for empty set")
	})
}

func TestTreeSetRange(t *testing.T) {
	s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	result := s.Range(3, 8)

	assert.Equal(t, []int{3, 4, 5, 6, 7}, result, "Range should return elements within bounds")
}

func TestTreeSetIteratorBidirectional(t *testing.T) {
	s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5})
	it := s.Iterator()

	t.Run("ForwardIteration", func(t *testing.T) {
		var forward []int
		for it.Next() {
			forward = append(forward, it.Value())
		}

		assert.Equal(t, []int{1, 2, 3, 4, 5}, forward, "Forward iteration should be in sorted order")
	})

	t.Run("BackwardIteration", func(t *testing.T) {
		var backward []int

		it.End()

		for it.Prev() {
			backward = append(backward, it.Value())
		}

		assert.Equal(t, []int{5, 4, 3, 2, 1}, backward, "Backward iteration should be in reverse order")
	})
}

func TestTreeSetJSONSerialization(t *testing.T) {
	s := NewTreeSetFromSlice([]int{1, 2, 3})

	data, err := json.Marshal(s)
	require.NoError(t, err, "Marshal should succeed")

	decoded := NewTreeSet[int]()
	require.NoError(t, json.Unmarshal(data, decoded), "Unmarshal should succeed")
	assert.True(t, s.Equal(decoded), "Decoded set should equal original")
}

func TestTreeSetGobSerialization(t *testing.T) {
	s := NewTreeSetFromSlice([]int{1, 2, 3})

	var buf bytes.Buffer
	require.NoError(t, gob.NewEncoder(&buf).Encode(s), "Gob encode should succeed")

	decoded := NewTreeSet[int]()
	require.NoError(t, gob.NewDecoder(&buf).Decode(decoded), "Gob decode should succeed")
	assert.True(t, s.Equal(decoded), "Decoded set should equal original")
}

func TestTreeSetEmptySerialization(t *testing.T) {
	s := NewTreeSet[int]()

	data, err := json.Marshal(s)
	require.NoError(t, err, "Marshal empty set should succeed")
	assert.Equal(t, "[]", string(data), "Empty set should serialize as empty array")

	decodedJSON := NewTreeSet[int]()
	require.NoError(t, json.Unmarshal(data, decodedJSON), "Unmarshal empty set should succeed")
	assert.True(t, decodedJSON.IsEmpty(), "Decoded empty set should be empty")
}

func TestTreeSetUnmarshalInvalid(t *testing.T) {
	s := NewTreeSet[int]()

	err := json.Unmarshal([]byte("not valid json"), s)
	assert.Error(t, err, "Unmarshal invalid JSON should return error")
}

func TestTreeSetSerializationWithoutComparatorInitialization(t *testing.T) {
	t.Run("UnmarshalJSONWithNilTree", func(t *testing.T) {
		var s TreeSet[int]

		assert.Panics(t, func() {
			_ = s.UnmarshalJSON([]byte("[1,2,3]"))
		}, "UnmarshalJSON should panic when tree is nil")
	})

	t.Run("GobDecodeWithNilTree", func(t *testing.T) {
		var s TreeSet[int]

		assert.Panics(t, func() {
			_ = s.GobDecode([]byte{})
		}, "GobDecode should panic when tree is nil")
	})

	t.Run("UnmarshalJSONWithNilComparator", func(t *testing.T) {
		var s TreeSet[int]

		s.tree = treeset.New[int]()
		s.cmp = nil

		assert.Panics(t, func() {
			_ = s.UnmarshalJSON([]byte("[1,2,3]"))
		}, "UnmarshalJSON should panic when comparator is nil")
	})

	t.Run("GobDecodeWithNilComparator", func(t *testing.T) {
		var s TreeSet[int]

		s.tree = treeset.New[int]()
		s.cmp = nil

		assert.Panics(t, func() {
			_ = s.GobDecode([]byte{})
		}, "GobDecode should panic when comparator is nil")
	})
}

func TestTreeSetInterfaceCompliance(t *testing.T) {
	t.Run("SetInterface", func(t *testing.T) {
		var s Set[int] = NewTreeSet[int]()
		s.Add(1, 2, 3)
		assert.Equal(t, 3, s.Size(), "Set interface should work correctly")
	})

	t.Run("OrderedSetInterface", func(t *testing.T) {
		var ordered OrderedSet[int] = NewTreeSet[int]()
		ordered.Add(3, 1, 2)

		min, _ := ordered.Min()
		max, _ := ordered.Max()

		assert.Equal(t, 1, min, "Min should be 1")
		assert.Equal(t, 3, max, "Max should be 3")
	})
}

func TestSetOperationsAcrossTypes(t *testing.T) {
	t.Run("HashSetUnionTreeSet", func(t *testing.T) {
		hashSet := NewHashSetFromSlice([]int{1, 2, 3})
		treeSet := NewTreeSetFromSlice([]int{3, 4, 5})

		result := hashSet.Union(treeSet)

		assert.Equal(t, 5, result.Size(), "Union should contain 5 elements")
		assert.True(t, result.ContainsAll(1, 2, 3, 4, 5), "Union should contain all elements")
	})

	t.Run("TreeSetIntersectionHashSet", func(t *testing.T) {
		treeSet := NewTreeSetFromSlice([]int{1, 2, 3, 4})
		hashSet := NewHashSetFromSlice([]int{3, 4, 5, 6})

		result := treeSet.Intersection(hashSet)

		assert.Equal(t, 2, result.Size(), "Intersection should contain 2 elements")
		assert.True(t, result.ContainsAll(3, 4), "Intersection should contain common elements")
	})
}
