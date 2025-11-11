package storage

import (
	"fmt"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/storage/services/filesystem"
	"github.com/ilxqx/vef-framework-go/internal/storage/services/memory"
	"github.com/ilxqx/vef-framework-go/internal/storage/services/minio"
	"github.com/ilxqx/vef-framework-go/storage"
)

// NewService creates a storage service based on configuration.
func NewService(cfg *config.StorageConfig, appCfg *config.AppConfig) (storage.Service, error) {
	provider := cfg.Provider
	if provider == constants.Empty {
		provider = constants.StorageMemory
	}

	switch provider {
	case constants.StorageMinIO:
		return minio.New(cfg.MinIO, appCfg)
	case constants.StorageMemory:
		return memory.New(), nil
	case constants.StorageFilesystem:
		return filesystem.New(cfg.Filesystem)
		// TODO: Add other services here
		// case constants.StorageOSS:
		//     return oss.New(...)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedStorageProvider, cfg.Provider)
	}
}
