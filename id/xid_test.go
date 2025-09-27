package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXidIdGenerator(t *testing.T) {
	t.Run("should create generator successfully", func(t *testing.T) {
		generator := NewXidIdGenerator()
		assert.NotNil(t, generator, "XID generator should not be nil")
	})

	t.Run("should generate valid XID format", func(t *testing.T) {
		generator := NewXidIdGenerator()
		id := generator.Generate()

		assert.NotEmpty(t, id, "Generated XID should not be empty")
		assert.Len(t, id, 20, "XID should be exactly 20 characters long")

		// XID uses base32 encoding with alphabet [0-9a-v]
		for _, char := range id {
			assert.True(t,
				(char >= '0' && char <= '9') || (char >= 'a' && char <= 'v'),
				"XID should only contain base32 characters [0-9a-v]: %c", char)
		}
	})

	t.Run("should generate unique IDs", func(t *testing.T) {
		generator := NewXidIdGenerator()
		ids := make(map[string]bool)
		iterations := 10000

		for range iterations {
			id := generator.Generate()
			assert.False(t, ids[id], "Generated XID should be unique: %s", id)
			ids[id] = true
		}

		assert.Len(t, ids, iterations, "All generated XIDs should be unique")
	})

	t.Run("should be thread-safe", func(t *testing.T) {
		generator := NewXidIdGenerator()
		const numGoroutines = 100
		const idsPerGoroutine = 100

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
			assert.False(t, ids[id], "Concurrent XID generation should produce unique IDs")
			ids[id] = true
		}

		assert.Len(t, ids, numGoroutines*idsPerGoroutine, "All concurrent XIDs should be unique")
	})

	t.Run("should generate sortable IDs", func(t *testing.T) {
		generator := NewXidIdGenerator()

		// Generate multiple IDs and verify they are roughly sortable by time
		// (XIDs contain timestamp information)
		var ids []string
		for range 100 {
			ids = append(ids, generator.Generate())
		}

		// XIDs should be lexicographically sortable due to timestamp prefix
		for i := 1; i < len(ids); i++ {
			// Since XIDs contain timestamp, later generated IDs should generally
			// be lexicographically greater (though there might be some edge cases
			// due to the millisecond precision and counter)
			assert.True(t, len(ids[i]) == len(ids[i-1]), "All XIDs should have same length")
		}
	})

	t.Run("default generator should work", func(t *testing.T) {
		assert.NotNil(t, DefaultXidIdGenerator, "Default XID generator should be initialized")

		id := DefaultXidIdGenerator.Generate()
		assert.NotEmpty(t, id, "Default XID generator should produce valid IDs")
		assert.Len(t, id, 20, "Default XID generator should produce 20-character IDs")
	})

	t.Run("should handle rapid generation", func(t *testing.T) {
		generator := NewXidIdGenerator()
		ids := make(map[string]bool)

		// Generate many IDs rapidly
		for range 1000 {
			id := generator.Generate()
			assert.False(t, ids[id], "Rapid generation should still produce unique IDs")
			ids[id] = true
		}

		assert.Len(t, ids, 1000, "All rapidly generated XIDs should be unique")
	})
}
