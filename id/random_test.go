package id

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomIdGenerator(t *testing.T) {
	t.Run("CreateWithCustomAlphabetAndLength", func(t *testing.T) {
		alphabet := "0123456789ABCDEF"
		length := 16
		generator := NewRandomIdGenerator(alphabet, length)
		assert.NotNil(t, generator, "Generator should not be nil")

		id := generator.Generate()
		assert.NotEmpty(t, id, "ID should not be empty")
		assert.Len(t, id, length, "ID should have specified length")

		for _, char := range id {
			assert.True(t, strings.ContainsRune(alphabet, char),
				"ID should contain only alphabet characters: %c", char)
		}
	})

	t.Run("GenerateUniqueIdsWithDefaultAlphabet", func(t *testing.T) {
		alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
		length := 21
		generator := NewRandomIdGenerator(alphabet, length)

		ids := make(map[string]bool)
		iterations := 10000

		for range iterations {
			id := generator.Generate()
			assert.False(t, ids[id], "ID should be unique: %s", id)
			ids[id] = true
		}

		assert.Len(t, ids, iterations, "All IDs should be unique")
	})

	t.Run("NumericAlphabet", func(t *testing.T) {
		alphabet := "0123456789"
		length := 10
		generator := NewRandomIdGenerator(alphabet, length)

		id := generator.Generate()
		assert.Len(t, id, length, "ID should have correct length")

		for _, char := range id {
			assert.True(t, char >= '0' && char <= '9',
				"ID should contain only digits: %c", char)
		}
	})

	t.Run("AlphabeticCharacters", func(t *testing.T) {
		alphabet := "abcdefghijklmnopqrstuvwxyz"
		length := 12
		generator := NewRandomIdGenerator(alphabet, length)

		id := generator.Generate()
		assert.Len(t, id, length, "ID should have correct length")

		for _, char := range id {
			assert.True(t, char >= 'a' && char <= 'z',
				"ID should contain only lowercase letters: %c", char)
		}
	})

	t.Run("ShortIds", func(t *testing.T) {
		alphabet := "ABCDEF"
		length := 4
		generator := NewRandomIdGenerator(alphabet, length)

		id := generator.Generate()
		assert.Len(t, id, length, "ID should have correct length")

		for _, char := range id {
			assert.True(t, strings.ContainsRune(alphabet, char),
				"ID should contain only alphabet characters: %c", char)
		}
	})

	t.Run("LongIds", func(t *testing.T) {
		alphabet := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		length := 128
		generator := NewRandomIdGenerator(alphabet, length)

		id := generator.Generate()
		assert.Len(t, id, length, "ID should have correct length")

		for _, char := range id {
			assert.True(t, strings.ContainsRune(alphabet, char),
				"ID should contain only alphabet characters: %c", char)
		}
	})

	t.Run("ThreadSafe", func(t *testing.T) {
		alphabet := "0123456789abcdefghijklmnopqrstuvwxyz"
		length := 16
		generator := NewRandomIdGenerator(alphabet, length)

		const (
			numGoroutines   = 100
			idsPerGoroutine = 100
		)

		idChan := make(chan string, numGoroutines*idsPerGoroutine)

		for range numGoroutines {
			go func() {
				for range idsPerGoroutine {
					idChan <- generator.Generate()
				}
			}()
		}

		ids := make(map[string]bool)

		for range numGoroutines * idsPerGoroutine {
			id := <-idChan
			assert.Len(t, id, length, "Concurrent generation should produce correct length")
			assert.False(t, ids[id], "Concurrent generation should produce unique IDs")
			ids[id] = true
		}

		assert.Len(t, ids, numGoroutines*idsPerGoroutine, "All concurrent IDs should be unique")
	})

	t.Run("SingleCharacterAlphabet", func(t *testing.T) {
		alphabet := "A"
		length := 5
		generator := NewRandomIdGenerator(alphabet, length)

		id := generator.Generate()
		assert.Equal(t, "AAAAA", id, "Single character alphabet should repeat character")
	})

	t.Run("SpecialCharacters", func(t *testing.T) {
		alphabet := "!@#$%^&*()"
		length := 8
		generator := NewRandomIdGenerator(alphabet, length)

		id := generator.Generate()
		assert.Len(t, id, length, "ID should have correct length")

		for _, char := range id {
			assert.True(t, strings.ContainsRune(alphabet, char),
				"ID should contain only alphabet characters: %c", char)
		}
	})

	t.Run("DifferentIdsWithSameParameters", func(t *testing.T) {
		alphabet := "0123456789abcdef"
		length := 20
		generator := NewRandomIdGenerator(alphabet, length)

		ids := make([]string, 100)
		for i := range ids {
			ids[i] = generator.Generate()
		}

		uniqueIds := make(map[string]bool)
		for _, id := range ids {
			uniqueIds[id] = true
		}

		assert.GreaterOrEqual(t, len(uniqueIds), 95,
			"Should generate mostly unique IDs with sufficient entropy")
	})
}
