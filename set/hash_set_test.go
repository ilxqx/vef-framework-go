package set

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashSet(t *testing.T) {
	t.Run("NewHashSet", func(t *testing.T) {
		s := NewHashSet[int]()
		assert.NotNil(t, s)
		assert.True(t, s.IsEmpty())
		assert.Equal(t, 0, s.Size())
	})

	t.Run("Add", func(t *testing.T) {
		s := NewHashSet[int]()
		assert.True(t, s.Add(1, 2, 3))
		assert.Equal(t, 3, s.Size())
		assert.False(t, s.Add(1, 2))
		assert.Equal(t, 3, s.Size())
	})

	t.Run("Remove", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3, 4, 5})
		assert.True(t, s.Remove(2, 4))
		assert.Equal(t, 3, s.Size())
		assert.False(t, s.Remove(10))
	})

	t.Run("Contains", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3})
		assert.True(t, s.Contains(2))
		assert.False(t, s.Contains(5))
	})

	t.Run("ContainsAll", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3, 4, 5})
		assert.True(t, s.ContainsAll(1, 3, 5))
		assert.False(t, s.ContainsAll(1, 6))
	})

	t.Run("ContainsAny", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3})
		assert.True(t, s.ContainsAny(3, 4, 5))
		assert.False(t, s.ContainsAny(6, 7, 8))
	})

	t.Run("Clear", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3})
		s.Clear()
		assert.True(t, s.IsEmpty())
		assert.Equal(t, 0, s.Size())
	})

	t.Run("RemoveIf", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3, 4, 5, 6})
		count := s.RemoveIf(func(element int) bool {
			return element%2 == 0
		})
		assert.Equal(t, 3, count)
		assert.Equal(t, 3, s.Size())
		assert.True(t, s.ContainsAll(1, 3, 5))
		assert.False(t, s.ContainsAny(2, 4, 6))
	})

	t.Run("Values", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3})
		values := s.Values()
		assert.Len(t, values, 3)
		assert.ElementsMatch(t, []int{1, 2, 3}, values)
	})

	t.Run("Seq", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3})

		var values []int
		for element := range s.Seq() {
			values = append(values, element)
		}

		assert.Len(t, values, 3)
		assert.ElementsMatch(t, []int{1, 2, 3}, values)
	})

	t.Run("Union", func(t *testing.T) {
		s1 := NewHashSetFromSlice([]int{1, 2, 3})
		s2 := NewHashSetFromSlice([]int{3, 4, 5})
		result := s1.Union(s2)
		assert.Equal(t, 5, result.Size())
		assert.True(t, result.ContainsAll(1, 2, 3, 4, 5))
	})

	t.Run("Intersection", func(t *testing.T) {
		s1 := NewHashSetFromSlice([]int{1, 2, 3, 4})
		s2 := NewHashSetFromSlice([]int{3, 4, 5, 6})
		result := s1.Intersection(s2)
		assert.Equal(t, 2, result.Size())
		assert.True(t, result.ContainsAll(3, 4))
	})

	t.Run("Difference", func(t *testing.T) {
		s1 := NewHashSetFromSlice([]int{1, 2, 3, 4})
		s2 := NewHashSetFromSlice([]int{3, 4, 5, 6})
		result := s1.Difference(s2)
		assert.Equal(t, 2, result.Size())
		assert.True(t, result.ContainsAll(1, 2))
	})

	t.Run("SymmetricDifference", func(t *testing.T) {
		s1 := NewHashSetFromSlice([]int{1, 2, 3})
		s2 := NewHashSetFromSlice([]int{3, 4, 5})
		result := s1.SymmetricDifference(s2)
		assert.Equal(t, 4, result.Size())
		assert.True(t, result.ContainsAll(1, 2, 4, 5))
		assert.False(t, result.Contains(3))
	})

	t.Run("IsSubset", func(t *testing.T) {
		s1 := NewHashSetFromSlice([]int{1, 2})
		s2 := NewHashSetFromSlice([]int{1, 2, 3, 4})
		assert.True(t, s1.IsSubset(s2))
		assert.False(t, s2.IsSubset(s1))
	})

	t.Run("IsSuperset", func(t *testing.T) {
		s1 := NewHashSetFromSlice([]int{1, 2, 3, 4})
		s2 := NewHashSetFromSlice([]int{1, 2})
		assert.True(t, s1.IsSuperset(s2))
		assert.False(t, s2.IsSuperset(s1))
	})

	t.Run("Equal", func(t *testing.T) {
		s1 := NewHashSetFromSlice([]int{1, 2, 3})
		s2 := NewHashSetFromSlice([]int{3, 2, 1})
		s3 := NewHashSetFromSlice([]int{1, 2, 4})

		assert.True(t, s1.Equal(s2))
		assert.False(t, s1.Equal(s3))
	})

	t.Run("Clone", func(t *testing.T) {
		s1 := NewHashSetFromSlice([]int{1, 2, 3})
		s2 := s1.Clone()
		assert.True(t, s1.Equal(s2))
		s2.Add(4)
		assert.False(t, s1.Equal(s2))
		assert.Equal(t, 3, s1.Size())
	})

	t.Run("Each", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3, 4, 5})
		sum := 0
		s.Each(func(element int) bool {
			sum += element

			return true
		})
		assert.Equal(t, 15, sum)
	})

	t.Run("Each with early termination", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3, 4, 5})
		count := 0
		s.Each(func(element int) bool {
			count++

			return count < 3
		})
		assert.Equal(t, 3, count)
	})

	t.Run("Filter", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3, 4, 5, 6})
		result := s.Filter(func(element int) bool {
			return element%2 == 0
		})
		assert.Equal(t, 3, result.Size())
		assert.True(t, result.ContainsAll(2, 4, 6))
	})

	t.Run("Map", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3})
		result := s.Map(func(element int) int {
			return element * 2
		})
		assert.Equal(t, 3, result.Size())
		assert.True(t, result.ContainsAll(2, 4, 6))
	})

	t.Run("Any", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3, 4, 5})
		assert.True(t, s.Any(func(element int) bool {
			return element > 3
		}))
		assert.False(t, s.Any(func(element int) bool {
			return element > 10
		}))
	})

	t.Run("All", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{2, 4, 6, 8})
		assert.True(t, s.All(func(element int) bool {
			return element%2 == 0
		}))
		assert.False(t, s.All(func(element int) bool {
			return element > 5
		}))
	})

	t.Run("JSON serialization", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3})
		data, err := json.Marshal(s)
		require.NoError(t, err)

		var decoded HashSet[int]
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.True(t, s.Equal(&decoded))
	})

	t.Run("Gob serialization", func(t *testing.T) {
		s := NewHashSetFromSlice([]int{1, 2, 3})

		var buf bytes.Buffer

		require.NoError(t, gob.NewEncoder(&buf).Encode(s))

		var decoded HashSet[int]
		require.NoError(t, gob.NewDecoder(&buf).Decode(&decoded))
		assert.True(t, s.Equal(&decoded))
	})

	t.Run("Empty set serialization", func(t *testing.T) {
		s := NewHashSet[int]()

		// JSON
		data, err := json.Marshal(s)
		require.NoError(t, err)
		assert.Equal(t, "[]", string(data))

		var decodedJSON HashSet[int]
		require.NoError(t, json.Unmarshal(data, &decodedJSON))
		assert.True(t, decodedJSON.IsEmpty())
	})

	t.Run("UnmarshalJSON with invalid input", func(t *testing.T) {
		var s HashSet[int]

		err := json.Unmarshal([]byte("not valid json"), &s)
		assert.Error(t, err)
	})

	t.Run("Zero value deserialization - JSON", func(t *testing.T) {
		// HashSet supports zero value deserialization
		var s HashSet[int]
		require.NoError(t, json.Unmarshal([]byte("[1,2,3]"), &s))
		assert.Equal(t, 3, s.Size())
		assert.True(t, s.ContainsAll(1, 2, 3))
	})

	t.Run("Zero value deserialization - Gob", func(t *testing.T) {
		// First create a set and encode it
		original := NewHashSetFromSlice([]int{1, 2, 3})

		var buf bytes.Buffer
		require.NoError(t, gob.NewEncoder(&buf).Encode(original.Values()))

		// Decode into zero value HashSet
		var s HashSet[int]
		require.NoError(t, s.GobDecode(buf.Bytes()))
		assert.Equal(t, 3, s.Size())
		assert.True(t, s.ContainsAll(1, 2, 3))
	})
}

// BenchmarkHashSetSerialization compares JSON and Gob serialization performance.
func BenchmarkHashSetSerialization(b *testing.B) {
	// Create test sets of different sizes
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		// Prepare test data
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

// BenchmarkHashSetSerializationRoundtrip benchmarks full serialization round-trip.
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
