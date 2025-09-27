package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	t.Run("should generate non-empty ID", func(t *testing.T) {
		id := Generate()
		assert.NotEmpty(t, id, "Generated ID should not be empty")
	})

	t.Run("should generate unique IDs", func(t *testing.T) {
		ids := make(map[string]bool)
		iterations := 1000

		for range iterations {
			id := Generate()
			assert.False(t, ids[id], "Generated ID should be unique: %s", id)
			ids[id] = true
		}

		assert.Len(t, ids, iterations, "All generated IDs should be unique")
	})

	t.Run("should use XID generator by default", func(t *testing.T) {
		id := Generate()

		// XID format: 12 bytes encoded as 20-character string
		assert.Len(t, id, 20, "XID should be 20 characters long")

		// XID uses base32 encoding (0-9, a-v)
		for _, char := range id {
			assert.True(t,
				(char >= '0' && char <= '9') || (char >= 'a' && char <= 'v'),
				"XID should only contain base32 characters (0-9, a-v): %c", char)
		}
	})
}

func TestGenerateUuid(t *testing.T) {
	t.Run("should generate valid UUID v7", func(t *testing.T) {
		uuid := GenerateUuid()
		assert.NotEmpty(t, uuid, "Generated UUID should not be empty")

		// UUID v7 format: xxxxxxxx-xxxx-7xxx-xxxx-xxxxxxxxxxxx
		assert.Len(t, uuid, 36, "UUID should be 36 characters long")
		assert.Equal(t, "-", string(uuid[8]), "UUID should have dash at position 8")
		assert.Equal(t, "-", string(uuid[13]), "UUID should have dash at position 13")
		assert.Equal(t, "-", string(uuid[18]), "UUID should have dash at position 18")
		assert.Equal(t, "-", string(uuid[23]), "UUID should have dash at position 23")
		assert.Equal(t, "7", string(uuid[14]), "UUID v7 should have version 7 at position 14")
	})

	t.Run("should generate unique UUIDs", func(t *testing.T) {
		uuids := make(map[string]bool)
		iterations := 1000

		for range iterations {
			uuid := GenerateUuid()
			assert.False(t, uuids[uuid], "Generated UUID should be unique: %s", uuid)
			uuids[uuid] = true
		}

		assert.Len(t, uuids, iterations, "All generated UUIDs should be unique")
	})
}

func TestDefaultGenerators(t *testing.T) {
	t.Run("default generators should be initialized", func(t *testing.T) {
		assert.NotNil(t, DefaultXidIdGenerator, "DefaultXidIdGenerator should be initialized")
		assert.NotNil(t, DefaultUuidIdGenerator, "DefaultUuidIdGenerator should be initialized")
		assert.NotNil(t, DefaultSnowflakeIdGenerator, "DefaultSnowflakeIdGenerator should be initialized")
	})

	t.Run("default generators should work correctly", func(t *testing.T) {
		xid := DefaultXidIdGenerator.Generate()
		assert.NotEmpty(t, xid, "XID generator should produce non-empty ID")

		uuid := DefaultUuidIdGenerator.Generate()
		assert.NotEmpty(t, uuid, "UUID generator should produce non-empty ID")

		snowflake := DefaultSnowflakeIdGenerator.Generate()
		assert.NotEmpty(t, snowflake, "Snowflake generator should produce non-empty ID")
	})
}

func TestConcurrentGeneration(t *testing.T) {
	t.Run("concurrent generation should be safe", func(t *testing.T) {
		const numGoroutines = 100
		const idsPerGoroutine = 100

		idChan := make(chan string, numGoroutines*idsPerGoroutine)

		// Start goroutines
		for range numGoroutines {
			go func() {
				for range idsPerGoroutine {
					idChan <- Generate()
				}
			}()
		}

		// Collect all IDs
		ids := make(map[string]bool)
		for range numGoroutines * idsPerGoroutine {
			id := <-idChan
			assert.False(t, ids[id], "Concurrent generation should produce unique IDs")
			ids[id] = true
		}

		assert.Len(t, ids, numGoroutines*idsPerGoroutine, "All concurrent IDs should be unique")
	})
}
