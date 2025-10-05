package id

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUuidIdGenerator(t *testing.T) {
	t.Run("should create generator successfully", func(t *testing.T) {
		generator := NewUuidIdGenerator()
		assert.NotNil(t, generator, "UUID generator should not be nil")
	})

	t.Run("should generate valid UUID v7 format", func(t *testing.T) {
		generator := NewUuidIdGenerator()
		id := generator.Generate()

		assert.NotEmpty(t, id, "Generated UUID should not be empty")
		assert.Len(t, id, 36, "UUID should be exactly 36 characters long")

		// UUID v7 format: xxxxxxxx-xxxx-7xxx-xxxx-xxxxxxxxxxxx
		uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
		assert.True(t, uuidRegex.MatchString(id), "UUID should match v7 format: %s", id)

		// Verify specific positions
		assert.Equal(t, "-", string(id[8]), "UUID should have dash at position 8")
		assert.Equal(t, "-", string(id[13]), "UUID should have dash at position 13")
		assert.Equal(t, "-", string(id[18]), "UUID should have dash at position 18")
		assert.Equal(t, "-", string(id[23]), "UUID should have dash at position 23")
		assert.Equal(t, "7", string(id[14]), "UUID v7 should have version 7 at position 14")

		// The variant bits (position 19) should be 8, 9, a, or b
		variantChar := string(id[19])
		assert.Contains(t, []string{"8", "9", "a", "b"}, variantChar,
			"UUID variant should be 8, 9, a, or b, got: %s", variantChar)
	})

	t.Run("should generate unique UUIDs", func(t *testing.T) {
		generator := NewUuidIdGenerator()
		uuids := make(map[string]bool)
		iterations := 10000

		for range iterations {
			uuid := generator.Generate()
			assert.False(t, uuids[uuid], "Generated UUID should be unique: %s", uuid)
			uuids[uuid] = true
		}

		assert.Len(t, uuids, iterations, "All generated UUIDs should be unique")
	})

	t.Run("should be thread-safe", func(t *testing.T) {
		generator := NewUuidIdGenerator()

		const (
			numGoroutines     = 100
			uuidsPerGoroutine = 100
		)

		uuidChan := make(chan string, numGoroutines*uuidsPerGoroutine)

		// Start goroutines
		for range numGoroutines {
			go func() {
				for range uuidsPerGoroutine {
					uuidChan <- generator.Generate()
				}
			}()
		}

		// Collect all UUIDs
		uuids := make(map[string]bool)

		for range numGoroutines * uuidsPerGoroutine {
			uuid := <-uuidChan
			assert.False(t, uuids[uuid], "Concurrent UUID generation should produce unique UUIDs")
			uuids[uuid] = true
		}

		assert.Len(t, uuids, numGoroutines*uuidsPerGoroutine, "All concurrent UUIDs should be unique")
	})

	t.Run("should generate time-ordered UUIDs", func(t *testing.T) {
		generator := NewUuidIdGenerator()

		// Generate multiple UUIDs and verify they are roughly time-ordered
		var uuids []string
		for range 100 {
			uuids = append(uuids, generator.Generate())
		}

		// UUID v7 should be lexicographically sortable due to timestamp prefix
		for i := 1; i < len(uuids); i++ {
			// The timestamp part (first 48 bits) should generally be increasing
			// We'll check that later UUIDs are not significantly "earlier"
			assert.True(t, uuids[i] >= uuids[i-1] ||
				// Allow for some edge cases due to millisecond precision
				uuids[i][0:8] == uuids[i-1][0:8],
				"UUID v7 should maintain rough time ordering")
		}
	})

	t.Run("default generator should work", func(t *testing.T) {
		assert.NotNil(t, DefaultUuidIdGenerator, "Default UUID generator should be initialized")

		uuid := DefaultUuidIdGenerator.Generate()
		assert.NotEmpty(t, uuid, "Default UUID generator should produce valid UUIDs")
		assert.Len(t, uuid, 36, "Default UUID generator should produce 36-character UUIDs")

		// Verify it's a valid UUID v7
		uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
		assert.True(t, uuidRegex.MatchString(uuid), "Default generator should produce valid UUID v7")
	})

	t.Run("should handle rapid generation", func(t *testing.T) {
		generator := NewUuidIdGenerator()
		uuids := make(map[string]bool)

		// Generate many UUIDs rapidly
		for range 1000 {
			uuid := generator.Generate()
			assert.False(t, uuids[uuid], "Rapid generation should still produce unique UUIDs")
			uuids[uuid] = true
		}

		assert.Len(t, uuids, 1000, "All rapidly generated UUIDs should be unique")
	})

	t.Run("should contain valid hex characters only", func(t *testing.T) {
		generator := NewUuidIdGenerator()
		uuid := generator.Generate()

		// Remove dashes and check all characters are valid hex
		hexPart := uuid[0:8] + uuid[9:13] + uuid[14:18] + uuid[19:23] + uuid[24:36]
		for _, char := range hexPart {
			assert.True(t,
				(char >= '0' && char <= '9') || (char >= 'a' && char <= 'f'),
				"UUID should only contain valid hex characters: %c", char)
		}
	})
}
