package storage

import (
	"context"
	"io"
)

// Service defines the core interface for object storage operations.
// Implementations should support various cloud storage services like MinIO, S3, Aliyun OSS, Tencent COS, etc.
type Service interface {
	// PutObject uploads an object to storage
	PutObject(ctx context.Context, opts PutObjectOptions) (*ObjectInfo, error)
	// GetObject retrieves an object from storage
	GetObject(ctx context.Context, opts GetObjectOptions) (io.ReadCloser, error)
	// DeleteObject deletes a single object from storage
	DeleteObject(ctx context.Context, opts DeleteObjectOptions) error
	// DeleteObjects deletes multiple objects from storage in a batch operation
	DeleteObjects(ctx context.Context, opts DeleteObjectsOptions) error
	// ListObjects lists objects in a bucket with optional filtering
	ListObjects(ctx context.Context, opts ListObjectsOptions) ([]ObjectInfo, error)
	// GetPresignedUrl generates a presigned Url for temporary access to an object
	GetPresignedUrl(ctx context.Context, opts PresignedURLOptions) (string, error)
	// CopyObject copies an object from source to destination
	CopyObject(ctx context.Context, opts CopyObjectOptions) (*ObjectInfo, error)
	// MoveObject moves an object from source to destination (implemented as Copy + Delete)
	MoveObject(ctx context.Context, opts MoveObjectOptions) (*ObjectInfo, error)
	// StatObject retrieves metadata information about an object
	StatObject(ctx context.Context, opts StatObjectOptions) (*ObjectInfo, error)

	// PromoteObject moves an object from temporary storage (temp/ prefix) to permanent storage.
	// It removes the "temp/" prefix from the object key, effectively promoting a temporary upload
	// to a permanent file.
	// This is useful for handling multistep upload workflows where files
	// are initially uploaded to temp/ and only moved to permanent storage after business logic commits.
	// If the key does not start with "temp/", this method does nothing and returns nil.
	PromoteObject(ctx context.Context, tempKey string) (*ObjectInfo, error)
}

// Promoter defines the interface for automatic file field promotion and cleanup.
// It supports three types of meta information fields:
// - uploaded_file: Direct file fields (string, *string, null.String, []string)
// - richtext: Rich text fields (string, *string, null.String), automatically extracts and processes resource references in HTML
// - markdown: Markdown fields (string, *string, null.String), automatically extracts and processes resource references in Markdown.
type Promoter[T any] interface {
	// Promote handles file promotion and cleanup based on the scenario:
	// - newModel != nil && oldModel == nil: Create (promote new files)
	// - newModel != nil && oldModel != nil: Update (promote new files + cleanup replaced files)
	// - newModel == nil && oldModel != nil: Delete (cleanup all files)
	Promote(ctx context.Context, newModel, oldModel *T) error
}
