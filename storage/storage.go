package storage

import "time"

// ObjectInfo represents metadata information about a stored object.
type ObjectInfo struct {
	// Bucket is the name of the storage bucket
	Bucket string `json:"bucket"`

	// Key is the unique identifier of the object within the bucket
	Key string `json:"key"`

	// ETag is the entity tag, typically an MD5 hash used for versioning and cache validation
	ETag string `json:"eTag"`

	// Size is the object size in bytes
	Size int64 `json:"size"`

	// ContentType is the MIME type of the object
	ContentType string `json:"contentType"`

	// LastModified is the timestamp when the object was last modified
	LastModified time.Time `json:"lastModified"`

	// Metadata contains custom key-value pairs associated with the object
	Metadata map[string]string `json:"metadata,omitempty"`
}
