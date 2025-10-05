package id

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomIdGenerator(t *testing.T) {
	t.Run("should create generator with custom alphabet and length", func(t *testing.T) {
		alphabet := "0123456789ABCDEF"
		length := 16
		generator := NewRandomIdGenerator(alphabet, length)
		assert.NotNil(t, generator, "Random ID generator should not be nil")

		id := generator.Generate()
		assert.NotEmpty(t, id, "Generated random ID should not be empty")
		assert.Len(t, id, length, "Generated ID should have specified length")

		// Verify all characters are from the specified alphabet
		for _, char := range id {
			assert.True(t, strings.ContainsRune(alphabet, char),
				"Generated ID should only contain characters from alphabet: %c", char)
		}
	})

	t.Run("should generate unique IDs with default alphabet", func(t *testing.T) {
		// Use nanoid default alphabet
		alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
		length := 21
		generator := NewRandomIdGenerator(alphabet, length)

		ids := make(map[string]bool)
		iterations := 10000

		for range iterations {
			id := generator.Generate()
			assert.False(t, ids[id], "Generated random ID should be unique: %s", id)
			ids[id] = true
		}

		assert.Len(t, ids, iterations, "All generated random IDs should be unique")
	})

	t.Run("should work with numeric alphabet", func(t *testing.T) {
		alphabet := "0123456789"
		length := 10
		generator := NewRandomIdGenerator(alphabet, length)

		id := generator.Generate()
		assert.Len(t, id, length, "Numeric ID should have correct length")

		// Verify all characters are digits
		for _, char := range id {
			assert.True(t, char >= '0' && char <= '9',
				"Numeric ID should only contain digits: %c", char)
		}
	})

	t.Run("should work with alphabetic characters", func(t *testing.T) {
		alphabet := "abcdefghijklmnopqrstuvwxyz"
		length := 12
		generator := NewRandomIdGenerator(alphabet, length)

		id := generator.Generate()
		assert.Len(t, id, length, "Alphabetic ID should have correct length")

		// Verify all characters are lowercase letters
		for _, char := range id {
			assert.True(t, char >= 'a' && char <= 'z',
				"Alphabetic ID should only contain lowercase letters: %c", char)
		}
	})

	t.Run("should work with short IDs", func(t *testing.T) {
		alphabet := "ABCDEF"
		length := 4
		generator := NewRandomIdGenerator(alphabet, length)

		id := generator.Generate()
		assert.Len(t, id, length, "Short ID should have correct length")

		for _, char := range id {
			assert.True(t, strings.ContainsRune(alphabet, char),
				"Short ID should only contain characters from alphabet: %c", char)
		}
	})

	t.Run("should work with long IDs", func(t *testing.T) {
		alphabet := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		length := 128
		generator := NewRandomIdGenerator(alphabet, length)

		id := generator.Generate()
		assert.Len(t, id, length, "Long ID should have correct length")

		for _, char := range id {
			assert.True(t, strings.ContainsRune(alphabet, char),
				"Long ID should only contain characters from alphabet: %c", char)
		}
	})

	t.Run("should be thread-safe", func(t *testing.T) {
		alphabet := "0123456789abcdefghijklmnopqrstuvwxyz"
		length := 16
		generator := NewRandomIdGenerator(alphabet, length)

		const (
			numGoroutines   = 100
			idsPerGoroutine = 100
		)

		idChan := make(chan string, numGoroutines*idsPerGoroutine)

		// Start goroutines
		for range numGoroutines {
			go func() {
				for range idsPerGoroutine {
					idChan <- generator.Generate()
				}
			}()
		}

		// Collect all IDs
		ids := make(map[string]bool)

		for range numGoroutines * idsPerGoroutine {
			id := <-idChan
			assert.Len(t, id, length, "Concurrent generation should produce correct length")
			assert.False(t, ids[id], "Concurrent generation should produce unique IDs")
			ids[id] = true
		}

		assert.Len(t, ids, numGoroutines*idsPerGoroutine, "All concurrent random IDs should be unique")
	})

	t.Run("should handle single character alphabet", func(t *testing.T) {
		alphabet := "A"
		length := 5
		generator := NewRandomIdGenerator(alphabet, length)

		id := generator.Generate()
		assert.Equal(t, "AAAAA", id, "Single character alphabet should generate repeated character")
	})

	t.Run("should handle special characters", func(t *testing.T) {
		alphabet := "!@#$%^&*()"
		length := 8
		generator := NewRandomIdGenerator(alphabet, length)

		id := generator.Generate()
		assert.Len(t, id, length, "Special character ID should have correct length")

		for _, char := range id {
			assert.True(t, strings.ContainsRune(alphabet, char),
				"Special character ID should only contain characters from alphabet: %c", char)
		}
	})

	t.Run("should generate different IDs with same parameters", func(t *testing.T) {
		alphabet := "0123456789abcdef"
		length := 20
		generator := NewRandomIdGenerator(alphabet, length)

		// Generate multiple IDs and ensure they're different
		// (with high probability for this length and alphabet size)
		ids := make([]string, 100)
		for i := range ids {
			ids[i] = generator.Generate()
		}

		// Check that most IDs are different (allowing for tiny probability of collision)
		uniqueIds := make(map[string]bool)
		for _, id := range ids {
			uniqueIds[id] = true
		}

		// With 16^20 possible combinations, we should have near 100% uniqueness
		assert.GreaterOrEqual(t, len(uniqueIds), 95,
			"Should generate mostly unique IDs with sufficient entropy")
	})
}
