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

func TestTreeSet(t *testing.T) {
	t.Run("NewTreeSet", func(t *testing.T) {
		s := NewTreeSet[int]()
		assert.NotNil(t, s)
		assert.True(t, s.IsEmpty())
		assert.Equal(t, 0, s.Size())
	})

	t.Run("Add maintains order", func(t *testing.T) {
		s := NewTreeSet[int]()
		s.Add(3, 1, 4, 1, 5, 9, 2, 6)
		values := s.Values()
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 9}, values)
	})

	t.Run("Custom comparator", func(t *testing.T) {
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
		assert.Equal(t, []int{5, 4, 3, 2, 1}, values)
	})

	t.Run("Remove", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5})
		assert.True(t, s.Remove(2, 5))
		assert.Equal(t, []int{1, 3, 4}, s.Values())
		assert.False(t, s.Remove(8))
	})

	t.Run("Contains variations", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5})
		assert.True(t, s.Contains(3))
		assert.False(t, s.Contains(9))
		assert.True(t, s.ContainsAll(1, 4, 5))
		assert.False(t, s.ContainsAll(1, 6))
		assert.True(t, s.ContainsAny(0, 5, 10))
		assert.False(t, s.ContainsAny(-1, 6))
	})

	t.Run("Size and empty", func(t *testing.T) {
		s := NewTreeSet[int]()
		assert.True(t, s.IsEmpty())
		s.Add(1)
		assert.False(t, s.IsEmpty())
		assert.Equal(t, 1, s.Size())
	})

	t.Run("Clear", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{1, 2, 3})
		s.Clear()
		assert.True(t, s.IsEmpty())
	})

	t.Run("RemoveIf", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8})
		count := s.RemoveIf(func(element int) bool { return element%2 == 0 })
		assert.Equal(t, 4, count)
		assert.Equal(t, []int{1, 3, 5, 7}, s.Values())
	})

	t.Run("Values", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{3, 1, 2})
		assert.Equal(t, []int{1, 2, 3}, s.Values())
	})

	t.Run("Seq", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{3, 1, 4, 1, 5, 9, 2, 6})

		var values []int
		for element := range s.Seq() {
			values = append(values, element)
		}

		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 9}, values)
	})

	t.Run("SeqWithIndex", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{3, 1, 4, 1, 5, 9, 2})

		var (
			positions []int
			values    []int
		)

		for index, element := range s.SeqWithIndex() {
			positions = append(positions, index)
			values = append(values, element)
		}

		assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, positions)
		assert.Equal(t, []int{1, 2, 3, 4, 5, 9}, values)
	})

	t.Run("ReverseSeq", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5})

		var values []int
		for element := range s.ReverseSeq() {
			values = append(values, element)
		}

		assert.Equal(t, []int{5, 4, 3, 2, 1}, values)
	})

	t.Run("ReverseSeqWithIndex", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{1, 2, 3, 4})

		var (
			positions []int
			values    []int
		)

		for index, element := range s.ReverseSeqWithIndex() {
			positions = append(positions, index)
			values = append(values, element)
		}

		assert.Equal(t, []int{0, 1, 2, 3}, positions)
		assert.Equal(t, []int{4, 3, 2, 1}, values)
	})

	t.Run("Each variations", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{3, 1, 4, 1, 5, 9, 2, 6})

		var values []int
		s.Each(func(element int) bool {
			values = append(values, element)

			return true
		})
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 9}, values)

		var positions []int

		values = nil
		s.EachIndexed(func(index, element int) bool {
			positions = append(positions, index)
			values = append(values, element)

			return true
		})
		assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6}, positions)
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 9}, values)
	})

	t.Run("Union", func(t *testing.T) {
		s1 := NewTreeSetFromSlice([]int{1, 2, 3})
		s2 := NewTreeSetFromSlice([]int{3, 4, 5})
		result := s1.Union(s2)
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result.Values())
	})

	t.Run("Intersection", func(t *testing.T) {
		s1 := NewTreeSetFromSlice([]int{1, 2, 3, 4})
		s2 := NewTreeSetFromSlice([]int{3, 4, 5, 6})
		result := s1.Intersection(s2)
		assert.Equal(t, []int{3, 4}, result.Values())
	})

	t.Run("Difference", func(t *testing.T) {
		s1 := NewTreeSetFromSlice([]int{1, 2, 3, 4})
		s2 := NewTreeSetFromSlice([]int{3, 4, 5, 6})
		result := s1.Difference(s2)
		assert.Equal(t, []int{1, 2}, result.Values())
	})

	t.Run("SymmetricDifference", func(t *testing.T) {
		s1 := NewTreeSetFromSlice([]int{1, 2, 3})
		s2 := NewTreeSetFromSlice([]int{3, 4, 5})
		result := s1.SymmetricDifference(s2)
		assert.Equal(t, []int{1, 2, 4, 5}, result.Values())
	})

	t.Run("Subset and superset", func(t *testing.T) {
		s1 := NewTreeSetFromSlice([]int{1, 2})
		s2 := NewTreeSetFromSlice([]int{1, 2, 3, 4})
		assert.True(t, s1.IsSubset(s2))
		assert.False(t, s2.IsSubset(s1))
		assert.True(t, s2.IsSuperset(s1))
		assert.False(t, s1.IsSuperset(s2))
	})

	t.Run("Equal", func(t *testing.T) {
		s1 := NewTreeSetFromSlice([]int{1, 2, 3})
		s2 := NewTreeSetFromSlice([]int{3, 2, 1})
		s3 := NewTreeSetFromSlice([]int{1, 2, 4})

		assert.True(t, s1.Equal(s2))
		assert.False(t, s1.Equal(s3))
	})

	t.Run("Clone", func(t *testing.T) {
		s1 := NewTreeSetFromSlice([]int{1, 2, 3})
		s2 := s1.Clone()
		assert.Equal(t, []int{1, 2, 3}, s2.Values())
		s2.Add(4)
		assert.Equal(t, []int{1, 2, 3}, s1.Values())
		assert.Equal(t, []int{1, 2, 3, 4}, s2.Values())
	})

	t.Run("Filter", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5, 6})
		result := s.Filter(func(element int) bool { return element%2 == 0 })
		assert.Equal(t, []int{2, 4, 6}, result.Values())
	})

	t.Run("Map", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{1, 2, 3})
		result := s.Map(func(element int) int { return element * 2 })
		assert.Equal(t, []int{2, 4, 6}, result.Values())
	})

	t.Run("Any", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5})
		assert.True(t, s.Any(func(element int) bool { return element > 3 }))
		assert.False(t, s.Any(func(element int) bool { return element > 10 }))
	})

	t.Run("All", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{2, 4, 6, 8})
		assert.True(t, s.All(func(element int) bool { return element%2 == 0 }))
		assert.False(t, s.All(func(element int) bool { return element > 5 }))
	})

	t.Run("Min and Max", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{3, 1, 4, 1, 5, 9, 2, 6})
		min, ok := s.Min()
		assert.True(t, ok)
		assert.Equal(t, 1, min)

		max, ok := s.Max()
		assert.True(t, ok)
		assert.Equal(t, 9, max)

		empty := NewTreeSet[int]()
		_, ok = empty.Min()
		assert.False(t, ok)
		_, ok = empty.Max()
		assert.False(t, ok)
	})

	t.Run("Range", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		assert.Equal(t, []int{3, 4, 5, 6, 7}, s.Range(3, 8))
	})

	t.Run("Iterator bidirectional", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{1, 2, 3, 4, 5})
		it := s.Iterator()

		var forward []int
		for it.Next() {
			forward = append(forward, it.Value())
		}

		assert.Equal(t, []int{1, 2, 3, 4, 5}, forward)

		var backward []int

		it.End()

		for it.Prev() {
			backward = append(backward, it.Value())
		}

		assert.Equal(t, []int{5, 4, 3, 2, 1}, backward)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{1, 2, 3})
		data, err := json.Marshal(s)
		require.NoError(t, err)

		decoded := NewTreeSet[int]()
		require.NoError(t, json.Unmarshal(data, decoded))
		assert.True(t, s.Equal(decoded))
	})

	t.Run("Gob serialization", func(t *testing.T) {
		s := NewTreeSetFromSlice([]int{1, 2, 3})

		var buf bytes.Buffer

		require.NoError(t, gob.NewEncoder(&buf).Encode(s))

		decoded := NewTreeSet[int]()
		require.NoError(t, gob.NewDecoder(&buf).Decode(decoded))
		assert.True(t, s.Equal(decoded))
	})

	t.Run("Empty set serialization", func(t *testing.T) {
		s := NewTreeSet[int]()

		// JSON
		data, err := json.Marshal(s)
		require.NoError(t, err)
		assert.Equal(t, "[]", string(data))

		decodedJSON := NewTreeSet[int]()
		require.NoError(t, json.Unmarshal(data, decodedJSON))
		assert.True(t, decodedJSON.IsEmpty())
	})

	t.Run("UnmarshalJSON with invalid input", func(t *testing.T) {
		s := NewTreeSet[int]()
		err := json.Unmarshal([]byte("not valid json"), s)
		assert.Error(t, err)
	})

	t.Run("Serialization without comparator initialization", func(t *testing.T) {
		// Create a TreeSet without comparator
		var s TreeSet[int]

		// UnmarshalJSON should panic with tree nil
		assert.Panics(t, func() {
			_ = s.UnmarshalJSON([]byte("[1,2,3]"))
		}, "UnmarshalJSON should panic when tree is nil")

		// GobDecode should panic with tree nil
		assert.Panics(t, func() {
			_ = s.GobDecode([]byte{})
		}, "GobDecode should panic when tree is nil")

		// Initialize tree but not comparator
		s.tree = treeset.New[int]()
		s.cmp = nil

		// UnmarshalJSON should panic with nil comparator
		assert.Panics(t, func() {
			_ = s.UnmarshalJSON([]byte("[1,2,3]"))
		}, "UnmarshalJSON should panic when comparator is nil")

		// GobDecode should panic with nil comparator
		assert.Panics(t, func() {
			_ = s.GobDecode([]byte{})
		}, "GobDecode should panic when comparator is nil")
	})
}

func TestTreeSetInterfaceCompliance(t *testing.T) {
	var s Set[int] = NewTreeSet[int]()
	s.Add(1, 2, 3)
	assert.Equal(t, 3, s.Size())

	var ordered OrderedSet[int] = NewTreeSet[int]()
	ordered.Add(3, 1, 2)
	min, _ := ordered.Min()
	max, _ := ordered.Max()

	assert.Equal(t, 1, min)
	assert.Equal(t, 3, max)
}

func TestSetOperationsAcrossTypes(t *testing.T) {
	hashSet := NewHashSetFromSlice([]int{1, 2, 3})
	treeSet := NewTreeSetFromSlice([]int{3, 4, 5})
	result := hashSet.Union(treeSet)
	assert.Equal(t, 5, result.Size())
	assert.True(t, result.ContainsAll(1, 2, 3, 4, 5))

	treeSet = NewTreeSetFromSlice([]int{1, 2, 3, 4})
	hashSet = NewHashSetFromSlice([]int{3, 4, 5, 6})
	result = treeSet.Intersection(hashSet)
	assert.Equal(t, 2, result.Size())
	assert.True(t, result.ContainsAll(3, 4))
}
