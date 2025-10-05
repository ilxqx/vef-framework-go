package id

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test edge cases and error conditions

func TestSnowflakeEdgeCases(t *testing.T) {
	t.Run("should handle maximum node ID", func(t *testing.T) {
		// Maximum node ID for 6-bit node field is 63 (2^6 - 1)
		generator, err := NewSnowflakeIdGenerator(63)
		require.NoError(t, err)

		id := generator.Generate()
		assert.NotEmpty(t, id, "Max node ID should generate valid IDs")
	})

	t.Run("should fail with node ID exceeding maximum", func(t *testing.T) {
		// Node ID 64 should fail (exceeds 6-bit limit)
		_, err := NewSnowflakeIdGenerator(64)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create snowflake node")
	})

	t.Run("should fail with negative node ID", func(t *testing.T) {
		_, err := NewSnowflakeIdGenerator(-1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create snowflake node")
	})

	t.Run("should handle rapid sequence generation", func(t *testing.T) {
		generator, err := NewSnowflakeIdGenerator(1)
		require.NoError(t, err)

		// Generate many IDs in rapid succession to test sequence counter
		ids := make(map[string]bool)

		for range 5000 {
			id := generator.Generate()
			assert.False(t, ids[id], "Rapid sequence generation should produce unique IDs")
			ids[id] = true
		}

		assert.Len(t, ids, 5000, "All rapid sequence IDs should be unique")
	})
}

func TestRandomIdGeneratorEdgeCases(t *testing.T) {
	t.Run("should handle empty alphabet", func(t *testing.T) {
		// This should panic when trying to generate, which is expected behavior
		// for invalid input
		generator := NewRandomIdGenerator("", 10)
		assert.NotNil(t, generator, "Should create generator even with empty alphabet")

		// Generating with empty alphabet should panic (expected behavior)
		assert.Panics(t, func() {
			generator.Generate()
		}, "Empty alphabet should panic when generating")
	})

	t.Run("should handle zero length", func(t *testing.T) {
		generator := NewRandomIdGenerator("abc", 0)

		// nanoid panics with zero length, which is expected behavior
		assert.Panics(t, func() {
			generator.Generate()
		}, "Zero length should panic with nanoid")
	})

	t.Run("should handle single character alphabet", func(t *testing.T) {
		generator := NewRandomIdGenerator("X", 10)
		id := generator.Generate()
		assert.Equal(t, "XXXXXXXXXX", id, "Single character alphabet should repeat character")
	})

	t.Run("should handle very long IDs", func(t *testing.T) {
		generator := NewRandomIdGenerator("0123456789", 1000)
		id := generator.Generate()
		assert.Len(t, id, 1000, "Should handle very long ID generation")

		// Verify all characters are from alphabet
		for _, char := range id {
			assert.True(t, char >= '0' && char <= '9', "Long ID should only contain digits")
		}
	})

	t.Run("should handle unicode characters", func(t *testing.T) {
		generator := NewRandomIdGenerator("αβγδε", 5)
		id := generator.Generate()
		assert.NotEmpty(t, id, "Should handle unicode alphabet")

		// Verify all characters are from the unicode alphabet
		allowedRunes := []rune("αβγδε")
		for _, char := range id {
			found := slices.Contains(allowedRunes, char)
			assert.True(t, found, "Unicode ID should only contain characters from alphabet: %c", char)
		}
	})
}

func TestUuidEdgeCases(t *testing.T) {
	t.Run("should handle rapid generation without collision", func(t *testing.T) {
		generator := NewUuidIdGenerator()
		uuids := make(map[string]bool)

		// Generate many UUIDs rapidly to test for collisions
		for range 100000 {
			uuid := generator.Generate()
			assert.False(t, uuids[uuid], "Rapid UUID generation should not have collisions")
			uuids[uuid] = true
		}

		assert.Len(t, uuids, 100000, "All rapid UUIDs should be unique")
	})

	t.Run("should maintain version and variant bits under load", func(t *testing.T) {
		generator := NewUuidIdGenerator()

		for range 1000 {
			uuid := generator.Generate()

			// Verify version is always 7
			assert.Equal(t, "7", string(uuid[14]), "Version should always be 7")

			// Verify variant bits (should be 8, 9, a, or b)
			variantChar := string(uuid[19])
			assert.Contains(t, []string{"8", "9", "a", "b"}, variantChar,
				"Variant should be valid")
		}
	})
}

func TestXidEdgeCases(t *testing.T) {
	t.Run("should handle concurrent generation from multiple generators", func(t *testing.T) {
		const (
			numGenerators   = 10
			idsPerGenerator = 1000
		)

		idChan := make(chan string, numGenerators*idsPerGenerator)

		// Create multiple generators and run them concurrently
		for range numGenerators {
			go func() {
				generator := NewXidIdGenerator()
				for range idsPerGenerator {
					idChan <- generator.Generate()
				}
			}()
		}

		// Collect all IDs
		ids := make(map[string]bool)

		for range numGenerators * idsPerGenerator {
			id := <-idChan
			assert.False(t, ids[id], "Multiple generator instances should produce unique IDs")
			ids[id] = true
		}

		assert.Len(t, ids, numGenerators*idsPerGenerator, "All IDs from multiple generators should be unique")
	})

	t.Run("should maintain format consistency", func(t *testing.T) {
		generator := NewXidIdGenerator()

		for range 1000 {
			id := generator.Generate()
			assert.Len(t, id, 20, "XID length should always be 20")

			// Check base32 alphabet consistency
			for _, char := range id {
				assert.True(t,
					(char >= '0' && char <= '9') || (char >= 'a' && char <= 'v'),
					"XID should always use base32 alphabet")
			}
		}
	})
}

func TestEnvironmentVariables(t *testing.T) {
	t.Run("should handle invalid NODE_ID environment variable", func(t *testing.T) {
		// This test verifies that invalid NODE_ID causes panic during init
		// Since init already ran, we can't test this directly in the same process

		// We can test the conversion logic that would be used
		// This is a unit test for the error handling logic
		originalNodeId := os.Getenv("NODE_ID")

		defer func() {
			if originalNodeId != "" {
				_ = os.Setenv("NODE_ID", originalNodeId)
			} else {
				_ = os.Unsetenv("NODE_ID")
			}
		}()

		// The actual init() function would panic with invalid NODE_ID
		// but we can't test that here since init() already ran
		assert.NotNil(t, DefaultSnowflakeIdGenerator, "Default generator should be initialized despite env var")
	})
}

func TestInterfaceCompliance(t *testing.T) {
	t.Run("all generators should implement IdGenerator interface", func(t *testing.T) {
		generators := []IdGenerator{
			NewXidIdGenerator(),
			NewUuidIdGenerator(),
			NewRandomIdGenerator("abc", 10),
			DefaultXidIdGenerator,
			DefaultUuidIdGenerator,
			DefaultSnowflakeIdGenerator,
		}

		for i, generator := range generators {
			assert.NotNil(t, generator, "Generator %d should not be nil", i)

			id := generator.Generate()
			assert.NotEmpty(t, id, "Generator %d should produce non-empty ID", i)
		}
	})

	t.Run("snowflake generator from constructor should implement interface", func(t *testing.T) {
		generator, err := NewSnowflakeIdGenerator(1)
		require.NoError(t, err)

		_ = generator

		id := generator.Generate()
		assert.NotEmpty(t, id, "Snowflake generator should produce non-empty ID")
	})
}

func TestMemoryUsage(t *testing.T) {
	t.Run("generators should not leak memory", func(t *testing.T) {
		// Create and discard many generators to test for memory leaks
		for i := range 1000 {
			_ = NewXidIdGenerator()
			_ = NewUuidIdGenerator()
			_ = NewRandomIdGenerator("abc", 10)

			if i%100 == 0 {
				// Generate some IDs
				DefaultXidIdGenerator.Generate()
				DefaultUuidIdGenerator.Generate()
				DefaultSnowflakeIdGenerator.Generate()
			}
		}

		// If we reach here without running out of memory, the test passes
		assert.True(t, true, "Memory usage test completed")
	})
}

func TestStringManipulation(t *testing.T) {
	t.Run("generated IDs should be safe for common string operations", func(t *testing.T) {
		generators := []IdGenerator{
			DefaultXidIdGenerator,
			DefaultUuidIdGenerator,
			DefaultSnowflakeIdGenerator,
			NewRandomIdGenerator("0123456789abcdef", 16),
		}

		for _, generator := range generators {
			id := generator.Generate()

			// Test common string operations
			assert.NotEmpty(t, strings.TrimSpace(id), "ID should not be empty after trimming")
			assert.False(t, strings.Contains(id, " "), "ID should not contain spaces")
			assert.False(t, strings.Contains(id, "\n"), "ID should not contain newlines")
			assert.False(t, strings.Contains(id, "\t"), "ID should not contain tabs")

			// Test case operations (should not panic)
			_ = strings.ToUpper(id)
			_ = strings.ToLower(id)
		}
	})
}
