package constants

type StorageProvider string

const (
	StorageMinIO      StorageProvider = "minio"
	StorageMemory     StorageProvider = "memory"
	StorageFilesystem StorageProvider = "filesystem"
)
