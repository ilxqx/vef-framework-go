package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/storage/providers/memory"
)

func TestNewStorageProvider(t *testing.T) {
	t.Run("memory provider", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Provider: constants.StorageMemory,
		}

		provider, err := NewStorageProvider(cfg, &config.AppConfig{})
		require.NoError(t, err)
		require.NotNil(t, provider)

		// Verify it's a memory provider
		_, ok := provider.(*memory.MemoryProvider)
		assert.True(t, ok, "provider should be a MemoryProvider")
	})

	t.Run("unsupported provider", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Provider: "unsupported",
		}

		provider, err := NewStorageProvider(cfg, &config.AppConfig{})
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "unsupported storage provider")
	})
}
