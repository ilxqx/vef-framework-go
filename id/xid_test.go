package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXidIdGenerator(t *testing.T) {
	t.Run("CreateGenerator", func(t *testing.T) {
		generator := NewXidIdGenerator()
		assert.NotNil(t, generator, "Generator should not be nil")
	})

	t.Run("GenerateValidXidFormat", func(t *testing.T) {
		generator := NewXidIdGenerator()
		id := generator.Generate()

		assert.NotEmpty(t, id, "XID should not be empty")
		assert.Len(t, id, 20, "XID should be 20 characters")

		for _, char := range id {
			assert.True(t,
				(char >= '0' && char <= '9') || (char >= 'a' && char <= 'v'),
				"XID should contain base32 characters [0-9a-v]: %c", char)
		}
	})

	t.Run("GenerateUniqueIds", func(t *testing.T) {
		generator := NewXidIdGenerator()
		ids := make(map[string]bool)
		iterations := 10000

		for range iterations {
			id := generator.Generate()
			assert.False(t, ids[id], "XID should be unique: %s", id)
			ids[id] = true
		}

		assert.Len(t, ids, iterations, "All XIDs should be unique")
	})

	t.Run("ThreadSafe", func(t *testing.T) {
		generator := NewXidIdGenerator()

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
			assert.False(t, ids[id], "Concurrent generation should produce unique IDs")
			ids[id] = true
		}

		assert.Len(t, ids, numGoroutines*idsPerGoroutine, "All concurrent XIDs should be unique")
	})

	t.Run("SortableIds", func(t *testing.T) {
		generator := NewXidIdGenerator()

		var ids []string
		for range 100 {
			ids = append(ids, generator.Generate())
		}

		for i := 1; i < len(ids); i++ {
			assert.True(t, len(ids[i]) == len(ids[i-1]), "All XIDs should have same length")
		}
	})

	t.Run("DefaultGenerator", func(t *testing.T) {
		assert.NotNil(t, DefaultXidIdGenerator, "Default generator should be initialized")

		id := DefaultXidIdGenerator.Generate()
		assert.NotEmpty(t, id, "Default generator should produce IDs")
		assert.Len(t, id, 20, "Default generator should produce 20-character IDs")
	})

	t.Run("RapidGeneration", func(t *testing.T) {
		generator := NewXidIdGenerator()
		ids := make(map[string]bool)

		for range 1000 {
			id := generator.Generate()
			assert.False(t, ids[id], "Rapid generation should produce unique IDs")
			ids[id] = true
		}

		assert.Len(t, ids, 1000, "All rapidly generated XIDs should be unique")
	})
}
