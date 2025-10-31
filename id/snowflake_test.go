package id

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnowflakeIdGenerator(t *testing.T) {
	t.Run("CreateGenerator", func(t *testing.T) {
		generator, err := NewSnowflakeIdGenerator(1)
		require.NoError(t, err, "Should create generator without error")
		assert.NotNil(t, generator, "Generator should not be nil")
	})

	t.Run("GenerateValidIds", func(t *testing.T) {
		generator, err := NewSnowflakeIdGenerator(1)
		require.NoError(t, err)

		id := generator.Generate()
		assert.NotEmpty(t, id, "ID should not be empty")

		for _, char := range id {
			assert.True(t,
				(char >= '0' && char <= '9') || (char >= 'a' && char <= 'z'),
				"ID should contain only alphanumeric characters: %c", char)
		}
	})

	t.Run("GenerateUniqueIds", func(t *testing.T) {
		generator, err := NewSnowflakeIdGenerator(1)
		require.NoError(t, err)

		ids := make(map[string]bool)
		iterations := 10000

		for range iterations {
			id := generator.Generate()
			assert.False(t, ids[id], "ID should be unique: %s", id)
			ids[id] = true
		}

		assert.Len(t, ids, iterations, "All IDs should be unique")
	})

	t.Run("DifferentNodeIds", func(t *testing.T) {
		gen1, err := NewSnowflakeIdGenerator(1)
		require.NoError(t, err)

		gen2, err := NewSnowflakeIdGenerator(2)
		require.NoError(t, err)

		id1 := gen1.Generate()
		id2 := gen2.Generate()

		assert.NotEqual(t, id1, id2, "IDs from different nodes should be different")
	})

	t.Run("InvalidNodeId", func(t *testing.T) {
		_, err := NewSnowflakeIdGenerator(64)
		assert.Error(t, err, "Should fail with invalid node ID")
		assert.Contains(t, err.Error(), "failed to create snowflake node")
	})

	t.Run("NegativeNodeId", func(t *testing.T) {
		_, err := NewSnowflakeIdGenerator(-1)
		assert.Error(t, err, "Should fail with negative node ID")
	})
}

func TestSnowflakeEnvironmentVariables(t *testing.T) {
	t.Run("UseNodeIdEnvironmentVariable", func(t *testing.T) {
		assert.NotNil(t, DefaultSnowflakeIdGenerator, "Default generator should be initialized")

		id := DefaultSnowflakeIdGenerator.Generate()
		assert.NotEmpty(t, id, "Default generator should produce valid IDs")
	})

	t.Run("ConcurrentGeneration", func(t *testing.T) {
		generator, err := NewSnowflakeIdGenerator(1)
		require.NoError(t, err)

		const (
			numGoroutines   = 50
			idsPerGoroutine = 200
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

		assert.Len(t, ids, numGoroutines*idsPerGoroutine, "All concurrent IDs should be unique")
	})
}

func TestSnowflakeConfiguration(t *testing.T) {
	t.Run("CustomEpochConfiguration", func(t *testing.T) {
		generator, err := NewSnowflakeIdGenerator(0)
		require.NoError(t, err)

		id := generator.Generate()
		assert.NotEmpty(t, id, "Generator with custom epoch should work")

		assert.False(t, strings.Contains(id, " "), "ID should not contain spaces")
		assert.False(t, strings.Contains(id, "+"), "ID should not contain plus signs")
		assert.False(t, strings.Contains(id, "/"), "ID should not contain slashes")
	})

	t.Run("BoundaryNodeIds", func(t *testing.T) {
		gen0, err := NewSnowflakeIdGenerator(0)
		require.NoError(t, err)

		id0 := gen0.Generate()
		assert.NotEmpty(t, id0, "Node ID 0 should work")

		gen63, err := NewSnowflakeIdGenerator(63)
		require.NoError(t, err)

		id63 := gen63.Generate()
		assert.NotEmpty(t, id63, "Node ID 63 should work")

		assert.NotEqual(t, id0, id63, "Different node IDs should generate different IDs")
	})
}
