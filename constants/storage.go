package constants

type StorageType string

const (
	StorageMinIO      StorageType = "minio"
	StorageMemory     StorageType = "memory"
	StorageFilesystem StorageType = "filesystem"
)
