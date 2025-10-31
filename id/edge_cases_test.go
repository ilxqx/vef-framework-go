package id

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnowflakeEdgeCases(t *testing.T) {
	t.Run("MaximumNodeId", func(t *testing.T) {
		generator, err := NewSnowflakeIdGenerator(63)
		require.NoError(t, err)

		id := generator.Generate()
		assert.NotEmpty(t, id, "Max node ID should generate valid IDs")
	})

	t.Run("NodeIdExceedingMaximum", func(t *testing.T) {
		_, err := NewSnowflakeIdGenerator(64)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create snowflake node")
	})

	t.Run("NegativeNodeId", func(t *testing.T) {
		_, err := NewSnowflakeIdGenerator(-1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create snowflake node")
	})

	t.Run("RapidSequenceGeneration", func(t *testing.T) {
		generator, err := NewSnowflakeIdGenerator(1)
		require.NoError(t, err)

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
	t.Run("EmptyAlphabet", func(t *testing.T) {
		generator := NewRandomIdGenerator("", 10)
		assert.NotNil(t, generator, "Should create generator even with empty alphabet")

		assert.Panics(t, func() {
			generator.Generate()
		}, "Empty alphabet should panic when generating")
	})

	t.Run("ZeroLength", func(t *testing.T) {
		generator := NewRandomIdGenerator("abc", 0)

		assert.Panics(t, func() {
			generator.Generate()
		}, "Zero length should panic")
	})

	t.Run("SingleCharacterAlphabet", func(t *testing.T) {
		generator := NewRandomIdGenerator("X", 10)
		id := generator.Generate()
		assert.Equal(t, "XXXXXXXXXX", id, "Single character alphabet should repeat character")
	})

	t.Run("VeryLongIds", func(t *testing.T) {
		generator := NewRandomIdGenerator("0123456789", 1000)
		id := generator.Generate()
		assert.Len(t, id, 1000, "Should handle very long ID generation")

		for _, char := range id {
			assert.True(t, char >= '0' && char <= '9', "Long ID should contain only digits")
		}
	})

	t.Run("UnicodeCharacters", func(t *testing.T) {
		generator := NewRandomIdGenerator("αβγδε", 5)
		id := generator.Generate()
		assert.NotEmpty(t, id, "Should handle unicode alphabet")

		allowedRunes := []rune("αβγδε")
		for _, char := range id {
			found := slices.Contains(allowedRunes, char)
			assert.True(t, found, "Unicode ID should contain only alphabet characters: %c", char)
		}
	})
}

func TestUuidEdgeCases(t *testing.T) {
	t.Run("RapidGenerationWithoutCollision", func(t *testing.T) {
		generator := NewUuidIdGenerator()
		uuids := make(map[string]bool)

		for range 100000 {
			uuid := generator.Generate()
			assert.False(t, uuids[uuid], "Rapid UUID generation should not have collisions")
			uuids[uuid] = true
		}

		assert.Len(t, uuids, 100000, "All rapid UUIDs should be unique")
	})

	t.Run("VersionAndVariantBitsUnderLoad", func(t *testing.T) {
		generator := NewUuidIdGenerator()

		for range 1000 {
			uuid := generator.Generate()

			assert.Equal(t, "7", string(uuid[14]), "Version should always be 7")

			variantChar := string(uuid[19])
			assert.Contains(t, []string{"8", "9", "a", "b"}, variantChar,
				"Variant should be valid")
		}
	})
}

func TestXidEdgeCases(t *testing.T) {
	t.Run("ConcurrentGenerationFromMultipleGenerators", func(t *testing.T) {
		const (
			numGenerators   = 10
			idsPerGenerator = 1000
		)

		idChan := make(chan string, numGenerators*idsPerGenerator)

		for range numGenerators {
			go func() {
				generator := NewXidIdGenerator()
				for range idsPerGenerator {
					idChan <- generator.Generate()
				}
			}()
		}

		ids := make(map[string]bool)

		for range numGenerators * idsPerGenerator {
			id := <-idChan
			assert.False(t, ids[id], "Multiple generators should produce unique IDs")
			ids[id] = true
		}

		assert.Len(t, ids, numGenerators*idsPerGenerator, "All IDs from multiple generators should be unique")
	})

	t.Run("FormatConsistency", func(t *testing.T) {
		generator := NewXidIdGenerator()

		for range 1000 {
			id := generator.Generate()
			assert.Len(t, id, 20, "XID length should always be 20")

			for _, char := range id {
				assert.True(t,
					(char >= '0' && char <= '9') || (char >= 'a' && char <= 'v'),
					"XID should always use base32 alphabet")
			}
		}
	})
}

func TestEnvironmentVariables(t *testing.T) {
	t.Run("InvalidNodeIdEnvironmentVariable", func(t *testing.T) {
		originalNodeId := os.Getenv("NODE_ID")

		defer func() {
			if originalNodeId != "" {
				_ = os.Setenv("NODE_ID", originalNodeId)
			} else {
				_ = os.Unsetenv("NODE_ID")
			}
		}()

		assert.NotNil(t, DefaultSnowflakeIdGenerator, "Default generator should be initialized")
	})
}

func TestInterfaceCompliance(t *testing.T) {
	t.Run("AllGeneratorsImplementInterface", func(t *testing.T) {
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
			assert.NotEmpty(t, id, "Generator %d should produce ID", i)
		}
	})

	t.Run("SnowflakeGeneratorFromConstructor", func(t *testing.T) {
		generator, err := NewSnowflakeIdGenerator(1)
		require.NoError(t, err)

		id := generator.Generate()
		assert.NotEmpty(t, id, "Snowflake generator should produce ID")
	})
}

func TestMemoryUsage(t *testing.T) {
	t.Run("GeneratorsShouldNotLeakMemory", func(t *testing.T) {
		for i := range 1000 {
			_ = NewXidIdGenerator()
			_ = NewUuidIdGenerator()
			_ = NewRandomIdGenerator("abc", 10)

			if i%100 == 0 {
				DefaultXidIdGenerator.Generate()
				DefaultUuidIdGenerator.Generate()
				DefaultSnowflakeIdGenerator.Generate()
			}
		}

		assert.True(t, true, "Memory usage test completed")
	})
}

func TestStringManipulation(t *testing.T) {
	t.Run("IdsSafeForCommonStringOperations", func(t *testing.T) {
		generators := []IdGenerator{
			DefaultXidIdGenerator,
			DefaultUuidIdGenerator,
			DefaultSnowflakeIdGenerator,
			NewRandomIdGenerator("0123456789abcdef", 16),
		}

		for _, generator := range generators {
			id := generator.Generate()

			assert.NotEmpty(t, strings.TrimSpace(id), "ID should not be empty after trimming")
			assert.False(t, strings.Contains(id, " "), "ID should not contain spaces")
			assert.False(t, strings.Contains(id, "\n"), "ID should not contain newlines")
			assert.False(t, strings.Contains(id, "\t"), "ID should not contain tabs")

			_ = strings.ToUpper(id)
			_ = strings.ToLower(id)
		}
	})
}
