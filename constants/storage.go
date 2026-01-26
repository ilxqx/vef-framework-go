package constants

// StorageProvider represents supported storage backend types.
type StorageProvider string

// Supported storage providers.
const (
	StorageMinIO      StorageProvider = "minio"
	StorageMemory     StorageProvider = "memory"
	StorageFilesystem StorageProvider = "filesystem"
)
