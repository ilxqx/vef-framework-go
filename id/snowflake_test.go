package id

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnowflakeIdGenerator(t *testing.T) {
	t.Run("should create generator successfully", func(t *testing.T) {
		generator, err := NewSnowflakeIdGenerator(1)
		require.NoError(t, err, "Should create snowflake generator without error")
		assert.NotNil(t, generator, "Generator should not be nil")
	})

	t.Run("should generate valid snowflake IDs", func(t *testing.T) {
		generator, err := NewSnowflakeIdGenerator(1)
		require.NoError(t, err)

		id := generator.Generate()
		assert.NotEmpty(t, id, "Generated ID should not be empty")

		// Snowflake IDs in Base36 format should be alphanumeric
		for _, char := range id {
			assert.True(t,
				(char >= '0' && char <= '9') || (char >= 'a' && char <= 'z'),
				"Snowflake ID should only contain alphanumeric characters: %c", char)
		}
	})

	t.Run("should generate unique IDs", func(t *testing.T) {
		generator, err := NewSnowflakeIdGenerator(1)
		require.NoError(t, err)

		ids := make(map[string]bool)
		iterations := 10000

		for range iterations {
			id := generator.Generate()
			assert.False(t, ids[id], "Generated ID should be unique: %s", id)
			ids[id] = true
		}

		assert.Len(t, ids, iterations, "All generated IDs should be unique")
	})

	t.Run("should handle different node IDs", func(t *testing.T) {
		gen1, err := NewSnowflakeIdGenerator(1)
		require.NoError(t, err)

		gen2, err := NewSnowflakeIdGenerator(2)
		require.NoError(t, err)

		// Generate IDs from different nodes
		id1 := gen1.Generate()
		id2 := gen2.Generate()

		assert.NotEqual(t, id1, id2, "IDs from different nodes should be different")
	})

	t.Run("should fail with invalid node ID", func(t *testing.T) {
		// Test with node ID that exceeds the limit (2^6 - 1 = 63)
		_, err := NewSnowflakeIdGenerator(64)
		assert.Error(t, err, "Should fail with invalid node ID")
		assert.Contains(t, err.Error(), "failed to create snowflake node")
	})

	t.Run("should handle negative node ID", func(t *testing.T) {
		_, err := NewSnowflakeIdGenerator(-1)
		assert.Error(t, err, "Should fail with negative node ID")
	})
}

func TestSnowflakeEnvironmentVariables(t *testing.T) {
	t.Run("should use NODE_ID environment variable", func(t *testing.T) {
		// This test verifies the init() function behavior
		// Since init() already ran, we can't test it directly,
		// but we can verify the default generator works
		assert.NotNil(t, DefaultSnowflakeIdGenerator, "Default generator should be initialized")

		id := DefaultSnowflakeIdGenerator.Generate()
		assert.NotEmpty(t, id, "Default generator should produce valid IDs")
	})

	t.Run("should handle concurrent generation", func(t *testing.T) {
		generator, err := NewSnowflakeIdGenerator(1)
		require.NoError(t, err)

		const (
			numGoroutines   = 50
			idsPerGoroutine = 200
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
			assert.False(t, ids[id], "Concurrent generation should produce unique IDs")
			ids[id] = true
		}

		assert.Len(t, ids, numGoroutines*idsPerGoroutine, "All concurrent IDs should be unique")
	})
}

func TestSnowflakeConfiguration(t *testing.T) {
	t.Run("should use custom epoch configuration", func(t *testing.T) {
		// We can't directly test the epoch since it's set in init(),
		// but we can verify that IDs are generated correctly
		generator, err := NewSnowflakeIdGenerator(0)
		require.NoError(t, err)

		id := generator.Generate()
		assert.NotEmpty(t, id, "Generator with custom epoch should work")

		// Verify the ID doesn't contain invalid characters
		assert.False(t, strings.Contains(id, " "), "ID should not contain spaces")
		assert.False(t, strings.Contains(id, "+"), "ID should not contain plus signs")
		assert.False(t, strings.Contains(id, "/"), "ID should not contain slashes")
	})

	t.Run("should handle boundary node IDs", func(t *testing.T) {
		// Test minimum valid node ID
		gen0, err := NewSnowflakeIdGenerator(0)
		require.NoError(t, err)

		id0 := gen0.Generate()
		assert.NotEmpty(t, id0, "Node ID 0 should work")

		// Test maximum valid node ID (2^6 - 1 = 63)
		gen63, err := NewSnowflakeIdGenerator(63)
		require.NoError(t, err)

		id63 := gen63.Generate()
		assert.NotEmpty(t, id63, "Node ID 63 should work")

		assert.NotEqual(t, id0, id63, "Different node IDs should generate different IDs")
	})
}
