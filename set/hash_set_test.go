package set

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashSetNew(t *testing.T) {
	s := NewHashSet[int]()

	assert.NotNil(t, s, "HashSet should be initialized")
	assert.True(t, s.IsEmpty(), "New HashSet should be empty")
	assert.Equal(t, 0, s.Size(), "Size should be 0")
}

func TestHashSetAdd(t *testing.T) {
	s := NewHashSet[int]()

	added := s.Add(1, 2, 3)
	assert.True(t, added, "Should return true when adding new elements")
	assert.Equal(t, 3, s.Size(), "Size should be 3 after adding 3 elements")

	added = s.Add(1, 2)
	assert.False(t, added, "Should return false when adding duplicate elements")
	assert.Equal(t, 3, s.Size(), "Size should remain 3 after adding duplicates")
}

func TestHashSetRemove(t *testing.T) {
	s := NewHashSetFromSlice([]int{1, 2, 3, 4, 5})

	removed := s.Remove(2, 4)
	assert.True(t, removed, "Should return true when removing existing elements")
	assert.Equal(t, 3, s.Size(), "Size should be 3 after removing 2 elements")

	removed = s.Remove(10)
	assert.False(t, removed, "Should return false when removing non-existent element")
}

func TestHashSetContains(t *testing.T) {
	s := NewHashSetFromSlice([]int{1, 2, 3})

	tests := []struct {
		name     string
		element  int
		expected bool
	}{
		{"ExistingElement", 2, true},
		{"NonExistentElement", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.Contains(tt.element)
			assert.Equal(t, tt.expected, result, "Contains should return correct result")
		})
	}
}

func TestHashSetContainsAll(t *testing.T) {
	s := NewHashSetFromSlice([]int{1, 2, 3, 4, 5})

	tests := []struct {
		name     string
		elements []int
		expected bool
	}{
		{"AllExist", []int{1, 3, 5}, true},
		{"SomeDoNotExist", []int{1, 6}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.ContainsAll(tt.elements...)
			assert.Equal(t, tt.expected, result, "ContainsAll should return correct result")
		})
	}
}

func TestHashSetContainsAny(t *testing.T) {
	s := NewHashSetFromSlice([]int{1, 2, 3})

	tests := []struct {
		name     string
		elements []int
		expected bool
	}{
		{"SomeExist", []int{3, 4, 5}, true},
		{"NoneExist", []int{6, 7, 8}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.ContainsAny(tt.elements...)
			assert.Equal(t, tt.expected, result, "ContainsAny should return correct result")
		})
	}
}

func TestHashSetClear(t *testing.T) {
	s := NewHashSetFromSlice([]int{1, 2, 3})

	s.Clear()

	assert.True(t, s.IsEmpty(), "Set should be empty after Clear")
	assert.Equal(t, 0, s.Size(), "Size should be 0 after Clear")
}

func TestHashSetRemoveIf(t *testing.T) {
	s := NewHashSetFromSlice([]int{1, 2, 3, 4, 5, 6})

	count := s.RemoveIf(func(element int) bool {
		return element%2 == 0
	})

	assert.Equal(t, 3, count, "Should remove 3 even elements")
	assert.Equal(t, 3, s.Size(), "Size should be 3 after removal")
	assert.True(t, s.ContainsAll(1, 3, 5), "Should contain only odd elements")
	assert.False(t, s.ContainsAny(2, 4, 6), "Should not contain even elements")
}

func TestHashSetValues(t *testing.T) {
	s := NewHashSetFromSlice([]int{1, 2, 3})

	values := s.Values()

	assert.Len(t, values, 3, "Values should have length 3")
	assert.ElementsMatch(t, []int{1, 2, 3}, values, "Values should match set elements")
}

func TestHashSetSeq(t *testing.T) {
	s := NewHashSetFromSlice([]int{1, 2, 3})

	var values []int
	for element := range s.Seq() {
		values = append(values, element)
	}

	assert.Len(t, values, 3, "Seq should iterate over 3 elements")
	assert.ElementsMatch(t, []int{1, 2, 3}, values, "Seq should yield all elements")
}

func TestHashSetUnion(t *testing.T) {
	s1 := NewHashSetFromSlice([]int{1, 2, 3})
	s2 := NewHashSetFromSlice([]int{3, 4, 5})

	result := s1.Union(s2)

	assert.Equal(t, 5, result.Size(), "Union should contain 5 elements")
	assert.True(t, result.ContainsAll(1, 2, 3, 4, 5), "Union should contain all elements from both sets")
}

func TestHashSetIntersection(t *testing.T) {
	s1 := NewHashSetFromSlice([]int{1, 2, 3, 4})
	s2 := NewHashSetFromSlice([]int{3, 4, 5, 6})

	result := s1.Intersection(s2)

	assert.Equal(t, 2, result.Size(), "Intersection should contain 2 elements")
	assert.True(t, result.ContainsAll(3, 4), "Intersection should contain common elements")
}

func TestHashSetDifference(t *testing.T) {
	s1 := NewHashSetFromSlice([]int{1, 2, 3, 4})
	s2 := NewHashSetFromSlice([]int{3, 4, 5, 6})

	result := s1.Difference(s2)

	assert.Equal(t, 2, result.Size(), "Difference should contain 2 elements")
	assert.True(t, result.ContainsAll(1, 2), "Difference should contain elements only in first set")
}

func TestHashSetSymmetricDifference(t *testing.T) {
	s1 := NewHashSetFromSlice([]int{1, 2, 3})
	s2 := NewHashSetFromSlice([]int{3, 4, 5})

	result := s1.SymmetricDifference(s2)

	assert.Equal(t, 4, result.Size(), "SymmetricDifference should contain 4 elements")
	assert.True(t, result.ContainsAll(1, 2, 4, 5), "SymmetricDifference should contain non-common elements")
	assert.False(t, result.Contains(3), "SymmetricDifference should not contain common element")
}

func TestHashSetIsSubset(t *testing.T) {
	s1 := NewHashSetFromSlice([]int{1, 2})
	s2 := NewHashSetFromSlice([]int{1, 2, 3, 4})

	assert.True(t, s1.IsSubset(s2), "s1 should be subset of s2")
	assert.False(t, s2.IsSubset(s1), "s2 should not be subset of s1")
}

func TestHashSetIsSuperset(t *testing.T) {
	s1 := NewHashSetFromSlice([]int{1, 2, 3, 4})
	s2 := NewHashSetFromSlice([]int{1, 2})

	assert.True(t, s1.IsSuperset(s2), "s1 should be superset of s2")
	assert.False(t, s2.IsSuperset(s1), "s2 should not be superset of s1")
}

func TestHashSetEqual(t *testing.T) {
	s1 := NewHashSetFromSlice([]int{1, 2, 3})
	s2 := NewHashSetFromSlice([]int{3, 2, 1})
	s3 := NewHashSetFromSlice([]int{1, 2, 4})

	assert.True(t, s1.Equal(s2), "Sets with same elements should be equal")
	assert.False(t, s1.Equal(s3), "Sets with different elements should not be equal")
}

func TestHashSetClone(t *testing.T) {
	s1 := NewHashSetFromSlice([]int{1, 2, 3})

	s2 := s1.Clone()

	assert.True(t, s1.Equal(s2), "Clone should be equal to original")

	s2.Add(4)

	assert.False(t, s1.Equal(s2), "Modifying clone should not affect original")
	assert.Equal(t, 3, s1.Size(), "Original size should remain unchanged")
}

func TestHashSetEach(t *testing.T) {
	t.Run("CompleteIteration", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3, 4, 5})
		sum := 0

		s.Each(func(element int) bool {
			sum += element

			return true
		})

		assert.Equal(t, 15, sum, "Should iterate over all elements")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3, 4, 5})
		count := 0

		s.Each(func(element int) bool {
			count++

			return count < 3
		})

		assert.Equal(t, 3, count, "Should terminate early when returning false")
	})
}

func TestHashSetFilter(t *testing.T) {
	s := NewHashSetFromSlice([]int{1, 2, 3, 4, 5, 6})

	result := s.Filter(func(element int) bool {
		return element%2 == 0
	})

	assert.Equal(t, 3, result.Size(), "Filtered set should contain 3 elements")
	assert.True(t, result.ContainsAll(2, 4, 6), "Filtered set should contain only even elements")
}

func TestHashSetMap(t *testing.T) {
	s := NewHashSetFromSlice([]int{1, 2, 3})

	result := s.Map(func(element int) int {
		return element * 2
	})

	assert.Equal(t, 3, result.Size(), "Mapped set should contain 3 elements")
	assert.True(t, result.ContainsAll(2, 4, 6), "Mapped set should contain doubled elements")
}

func TestHashSetAny(t *testing.T) {
	s := NewHashSetFromSlice([]int{1, 2, 3, 4, 5})

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

func TestHashSetAll(t *testing.T) {
	s := NewHashSetFromSlice([]int{2, 4, 6, 8})

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

func TestHashSetJSONSerialization(t *testing.T) {
	s := NewHashSetFromSlice([]int{1, 2, 3})

	data, err := json.Marshal(s)
	require.NoError(t, err, "Marshal should succeed")

	var decoded HashSet[int]
	require.NoError(t, json.Unmarshal(data, &decoded), "Unmarshal should succeed")
	assert.True(t, s.Equal(&decoded), "Decoded set should equal original")
}

func TestHashSetGobSerialization(t *testing.T) {
	s := NewHashSetFromSlice([]int{1, 2, 3})

	var buf bytes.Buffer
	require.NoError(t, gob.NewEncoder(&buf).Encode(s), "Gob encode should succeed")

	var decoded HashSet[int]
	require.NoError(t, gob.NewDecoder(&buf).Decode(&decoded), "Gob decode should succeed")
	assert.True(t, s.Equal(&decoded), "Decoded set should equal original")
}

func TestHashSetEmptySerialization(t *testing.T) {
	s := NewHashSet[int]()

	data, err := json.Marshal(s)
	require.NoError(t, err, "Marshal empty set should succeed")
	assert.Equal(t, "[]", string(data), "Empty set should serialize as empty array")

	var decodedJSON HashSet[int]
	require.NoError(t, json.Unmarshal(data, &decodedJSON), "Unmarshal empty set should succeed")
	assert.True(t, decodedJSON.IsEmpty(), "Decoded empty set should be empty")
}

func TestHashSetUnmarshalInvalid(t *testing.T) {
	var s HashSet[int]

	err := json.Unmarshal([]byte("not valid json"), &s)
	assert.Error(t, err, "Unmarshal invalid JSON should return error")
}

func TestHashSetZeroValueDeserialization(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		var s HashSet[int]
		require.NoError(t, json.Unmarshal([]byte("[1,2,3]"), &s), "Zero value unmarshal should succeed")
		assert.Equal(t, 3, s.Size(), "Size should be 3")
		assert.True(t, s.ContainsAll(1, 2, 3), "Should contain all elements")
	})

	t.Run("Gob", func(t *testing.T) {
		original := NewHashSetFromSlice([]int{1, 2, 3})

		var buf bytes.Buffer
		require.NoError(t, gob.NewEncoder(&buf).Encode(original.Values()), "Encode should succeed")

		var s HashSet[int]
		require.NoError(t, s.GobDecode(buf.Bytes()), "Zero value GobDecode should succeed")
		assert.Equal(t, 3, s.Size(), "Size should be 3")
		assert.True(t, s.ContainsAll(1, 2, 3), "Should contain all elements")
	})
}

func BenchmarkHashSetSerialization(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		elements := make([]int, size)
		for i := range size {
			elements[i] = i
		}

		set := NewHashSetFromSlice(elements)

		b.Run(fmt.Sprintf("JSON_Marshal_%d", size), func(b *testing.B) {
			b.ResetTimer()

			for b.Loop() {
				_, err := json.Marshal(set)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("JSON_Unmarshal_%d", size), func(b *testing.B) {
			data, _ := json.Marshal(set)

			b.ResetTimer()

			for b.Loop() {
				var s HashSet[int]

				err := json.Unmarshal(data, &s)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("Gob_Encode_%d", size), func(b *testing.B) {
			b.ResetTimer()

			for b.Loop() {
				_, err := set.GobEncode()
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("Gob_Decode_%d", size), func(b *testing.B) {
			data, _ := set.GobEncode()

			b.ResetTimer()

			for b.Loop() {
				var s HashSet[int]

				err := s.GobDecode(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkHashSetSerializationRoundtrip(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		elements := make([]int, size)
		for i := range size {
			elements[i] = i
		}

		set := NewHashSetFromSlice(elements)

		b.Run(fmt.Sprintf("JSON_Roundtrip_%d", size), func(b *testing.B) {
			b.ResetTimer()

			for b.Loop() {
				data, err := json.Marshal(set)
				if err != nil {
					b.Fatal(err)
				}

				var s HashSet[int]

				err = json.Unmarshal(data, &s)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("Gob_Roundtrip_%d", size), func(b *testing.B) {
			b.ResetTimer()

			for b.Loop() {
				data, err := set.GobEncode()
				if err != nil {
					b.Fatal(err)
				}

				var s HashSet[int]

				err = s.GobDecode(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
