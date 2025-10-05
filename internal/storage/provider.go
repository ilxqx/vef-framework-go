package storage

import (
	"fmt"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/storage/providers/memory"
	"github.com/ilxqx/vef-framework-go/internal/storage/providers/minio"
	"github.com/ilxqx/vef-framework-go/storage"
)

// NewStorageProvider creates a storage provider based on configuration.
func NewStorageProvider(cfg *config.StorageConfig, appCfg *config.AppConfig) (storage.Provider, error) {
	provider := cfg.Provider
	if provider == constants.Empty {
		provider = constants.StorageMemory
	}

	switch provider {
	case constants.StorageMinIO:
		return minio.NewMinIOProvider(cfg.MinIO, appCfg)
	case constants.StorageMemory:
		return memory.NewMemoryProvider(), nil
		// TODO: Add other providers here
		// case constants.StorageOSS:
		//     return oss.NewOSSProvider(...)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedStorageProvider, cfg.Provider)
	}
}
