package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateId(t *testing.T) {
	t.Run("generates unique IDs", func(t *testing.T) {
		id1 := GenerateId()
		id2 := GenerateId()

		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2)
	})

	t.Run("ID format", func(t *testing.T) {
		id := GenerateId()

		// xid generates 20-character strings
		assert.Len(t, id, 20)

		// Should contain only valid characters (base32)
		for _, c := range id {
			assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'v'))
		}
	})

	t.Run("multiple generations", func(t *testing.T) {
		ids := make(map[string]bool)
		count := 100

		for range count {
			id := GenerateId()
			assert.False(t, ids[id], "Duplicate ID generated: %s", id)
			ids[id] = true
		}

		assert.Len(t, ids, count)
	})
}

func TestGenerateRandomId(t *testing.T) {
	t.Run("default length", func(t *testing.T) {
		id := GenerateRandomId()

		assert.NotEmpty(t, id)
		// Default nanoid length is 21
		assert.Len(t, id, 21)
	})

	t.Run("custom length", func(t *testing.T) {
		lengths := []int{5, 10, 15, 32}

		for _, length := range lengths {
			id := GenerateRandomId(length)
			assert.Len(t, id, length)
			assert.NotEmpty(t, id)
		}
	})

	t.Run("generates unique IDs", func(t *testing.T) {
		id1 := GenerateRandomId(10)
		id2 := GenerateRandomId(10)

		assert.Len(t, id1, 10)
		assert.Len(t, id2, 10)
		assert.NotEqual(t, id1, id2)
	})

	t.Run("zero length", func(t *testing.T) {
		id := GenerateRandomId(0)
		assert.Empty(t, id)
	})

	t.Run("multiple generations with same length", func(t *testing.T) {
		ids := make(map[string]bool)
		count := 100
		length := 16

		for range count {
			id := GenerateRandomId(length)
			assert.Len(t, id, length)
			assert.False(t, ids[id], "Duplicate ID generated: %s", id)
			ids[id] = true
		}

		assert.Len(t, ids, count)
	})
}

func TestGenerateCustomizedRandomId(t *testing.T) {
	t.Run("custom alphabet", func(t *testing.T) {
		alphabet := "ABCDEFGHIJ"
		length := 10

		id := GenerateCustomizedRandomId(alphabet, length)

		assert.Len(t, id, length)
		for _, c := range id {
			assert.Contains(t, alphabet, string(c))
		}
	})

	t.Run("numeric alphabet", func(t *testing.T) {
		alphabet := "0123456789"
		length := 8

		id := GenerateCustomizedRandomId(alphabet, length)

		assert.Len(t, id, length)
		for _, c := range id {
			assert.True(t, c >= '0' && c <= '9')
		}
	})

	t.Run("single character alphabet", func(t *testing.T) {
		alphabet := "X"
		length := 5

		id := GenerateCustomizedRandomId(alphabet, length)

		assert.Equal(t, "XXXXX", id)
		assert.Len(t, id, length)
	})

	t.Run("binary alphabet", func(t *testing.T) {
		alphabet := "01"
		length := 16

		id := GenerateCustomizedRandomId(alphabet, length)

		assert.Len(t, id, length)
		for _, c := range id {
			assert.True(t, c == '0' || c == '1')
		}
	})

	t.Run("generates unique IDs with custom alphabet", func(t *testing.T) {
		alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		length := 12
		ids := make(map[string]bool)
		count := 50

		for range count {
			id := GenerateCustomizedRandomId(alphabet, length)
			assert.Len(t, id, length)
			assert.False(t, ids[id], "Duplicate ID generated: %s", id)
			ids[id] = true

			// Verify all characters are from alphabet
			for _, c := range id {
				assert.Contains(t, alphabet, string(c))
			}
		}

		assert.Len(t, ids, count)
	})

	t.Run("zero length", func(t *testing.T) {
		alphabet := "ABC"
		length := 0

		// This will panic, so we need to handle it
		assert.Panics(t, func() {
			GenerateCustomizedRandomId(alphabet, length)
		})
	})

	t.Run("special characters alphabet", func(t *testing.T) {
		alphabet := "!@#$%^&*()"
		length := 6

		id := GenerateCustomizedRandomId(alphabet, length)

		assert.Len(t, id, length)
		for _, c := range id {
			assert.Contains(t, alphabet, string(c))
		}
	})
}
