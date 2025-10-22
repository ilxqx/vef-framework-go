package storage

import (
	"errors"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/id"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/storage"
	"github.com/ilxqx/vef-framework-go/testhelpers"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

const (
	templateDatePath = "2006/01/02"
	defaultExtension = ".bin"
)

// NewStorageResource creates a new storage resource with the provided storage provider.
func NewStorageResource(provider storage.Provider) api.Resource {
	// In test environment, make all Apis public (no authentication required)
	isPublic := testhelpers.IsTestEnv()

	return &StorageResource{
		provider: provider,
		Resource: api.NewResource(
			"base/storage",
			api.WithApis(
				api.Spec{Action: "upload", Public: isPublic},
				api.Spec{Action: "get_presigned_url", Public: isPublic},
				api.Spec{Action: "stat", Public: isPublic},
				api.Spec{Action: "list", Public: isPublic},
			),
		),
	}
}

// StorageResource handles storage-related Api endpoints.
type StorageResource struct {
	api.Resource

	provider storage.Provider
}

// UploadParams represents the request parameters for file upload.
type UploadParams struct {
	api.P

	// File is the file to upload
	File *multipart.FileHeader

	// ContentType specifies the MIME type of the object (optional, auto-detected if not provided)
	ContentType string `json:"contentType"`
	// Metadata contains custom key-value pairs to associate with the object
	Metadata map[string]string `json:"metadata"`
}

// Upload uploads a file to storage with auto-generated key.
// Key generation format: temp/YYYY/MM/DD/{uuid}{extension}
// Example: temp/2025/01/15/550e8400-e29b-41d4-a716-446655440000.jpg
//
// The file should be uploaded via multipart form with field name "file".
func (r *StorageResource) Upload(ctx fiber.Ctx, params UploadParams) error {
	if webhelpers.IsJson(ctx) {
		return result.Err(i18n.T("upload_requires_multipart"))
	}

	if params.File == nil {
		return result.Err(i18n.T("upload_requires_file"))
	}

	// Generate unique key with date-based partitioning
	key := r.generateObjectKey(params.File.Filename)

	// Open the uploaded file
	file, err := params.File.Open()
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			logger.Errorf("failed to close file: %v", closeErr)
		}
	}()

	// Determine content type if not provided
	contentType := params.ContentType
	if contentType == constants.Empty {
		contentType = params.File.Header.Get(fiber.HeaderContentType)
	}

	// Merge user metadata with original filename
	metadata := params.Metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}

	metadata[storage.MetadataKeyOriginalFilename] = params.File.Filename

	// Upload to storage provider
	info, err := r.provider.PutObject(ctx.Context(), storage.PutObjectOptions{
		Key:         key,
		Reader:      file,
		Size:        params.File.Size,
		ContentType: contentType,
		Metadata:    metadata,
	})
	if err != nil {
		return err
	}

	return result.Ok(info).Response(ctx)
}

// generateObjectKey generates a unique object key with date-based partitioning.
// Format: temp/YYYY/MM/DD/{uuid}{extension}.
func (r *StorageResource) generateObjectKey(filename string) string {
	// Get current date for partitioning
	now := time.Now()
	datePath := now.Format(templateDatePath)

	// Generate UUID for uniqueness
	id := id.GenerateUuid()

	// Extract file extension (including the dot)
	ext := filepath.Ext(filename)
	if ext == constants.Empty {
		ext = defaultExtension
	}

	// Build the key
	var keyBuilder strings.Builder

	_, _ = keyBuilder.WriteString(storage.TempPrefix)

	// Add date path
	_, _ = keyBuilder.WriteString(datePath)
	_ = keyBuilder.WriteByte(constants.ByteSlash)

	// Add UUID and extension
	_, _ = keyBuilder.WriteString(id)
	_, _ = keyBuilder.WriteString(ext)

	return keyBuilder.String()
}

// GetPresignedUrlParams represents the request parameters for getting presigned URL.
type GetPresignedUrlParams struct {
	api.P

	// Key is the unique identifier of the object
	Key string `json:"key" validate:"required"`
	// Expires specifies URL validity duration in seconds (default: 3600)
	Expires int `json:"expires"`
	// Method specifies the HTTP method (GET for download, PUT for upload)
	Method string `json:"method"`
}

// GetPresignedUrl generates a presigned URL for temporary access to an object.
func (r *StorageResource) GetPresignedUrl(ctx fiber.Ctx, params GetPresignedUrlParams) error {
	// Default values
	expires := params.Expires
	if expires <= 0 {
		expires = 3600 // 1 hour default
	}

	method := params.Method
	if method == constants.Empty {
		method = http.MethodGet
	}

	// Generate presigned URL
	url, err := r.provider.GetPresignedURL(ctx.Context(), storage.PresignedURLOptions{
		Key:     params.Key,
		Expires: time.Duration(expires) * time.Second,
		Method:  method,
	})
	if err != nil {
		return err
	}

	return result.Ok(fiber.Map{"url": url}).Response(ctx)
}

// DeleteParams represents the request parameters for deleting an object.
type DeleteParams struct {
	api.P

	// Key is the unique identifier of the object to delete
	Key string `json:"key" validate:"required"`
}

// Delete deletes a single object from storage.
func (r *StorageResource) Delete(ctx fiber.Ctx, params DeleteParams) error {
	err := r.provider.DeleteObject(ctx.Context(), storage.DeleteObjectOptions{
		Key: params.Key,
	})
	if err != nil {
		return err
	}

	return result.Ok().Response(ctx)
}

// DeleteManyParams represents the request parameters for batch deleting objects.
type DeleteManyParams struct {
	api.P

	// Keys is the list of object identifiers to delete
	Keys []string `json:"keys" validate:"required,min=1"`
}

// DeleteMany deletes multiple objects from storage in a batch operation.
func (r *StorageResource) DeleteMany(ctx fiber.Ctx, params DeleteManyParams) error {
	err := r.provider.DeleteObjects(ctx.Context(), storage.DeleteObjectsOptions{
		Keys: params.Keys,
	})
	if err != nil {
		return err
	}

	return result.Ok().Response(ctx)
}

// ListParams represents the request parameters for listing objects.
type ListParams struct {
	api.P

	// Prefix filters objects by key prefix
	Prefix string `json:"prefix"`
	// Recursive determines whether to list objects recursively
	Recursive bool `json:"recursive"`
	// MaxKeys limits the maximum number of objects to return
	MaxKeys int `json:"maxKeys"`
}

// List lists objects in a bucket with optional filtering.
func (r *StorageResource) List(ctx fiber.Ctx, params ListParams) error {
	objects, err := r.provider.ListObjects(ctx.Context(), storage.ListObjectsOptions{
		Prefix:    params.Prefix,
		Recursive: params.Recursive,
		MaxKeys:   params.MaxKeys,
	})
	if err != nil {
		return err
	}

	return result.Ok(objects).Response(ctx)
}

// CopyParams represents the request parameters for copying an object.
type CopyParams struct {
	api.P

	// SourceKey is the identifier of the source object
	SourceKey string `json:"sourceKey" validate:"required"`
	// DestKey is the identifier for the copied object
	DestKey string `json:"destKey" validate:"required"`
}

// Copy copies an object from source to destination.
func (r *StorageResource) Copy(ctx fiber.Ctx, params CopyParams) error {
	info, err := r.provider.CopyObject(ctx.Context(), storage.CopyObjectOptions{
		SourceKey: params.SourceKey,
		DestKey:   params.DestKey,
	})
	if err != nil {
		return err
	}

	return result.Ok(info).Response(ctx)
}

// MoveParams represents the request parameters for moving an object.
type MoveParams struct {
	api.P

	// SourceKey is the identifier of the source object
	SourceKey string `json:"sourceKey" validate:"required"`
	// DestKey is the identifier for the moved object
	DestKey string `json:"destKey" validate:"required"`
}

// Move moves an object from source to destination (implemented as Copy + Delete).
func (r *StorageResource) Move(ctx fiber.Ctx, params MoveParams) error {
	info, err := r.provider.MoveObject(ctx.Context(), storage.MoveObjectOptions{
		CopyObjectOptions: storage.CopyObjectOptions{
			SourceKey: params.SourceKey,
			DestKey:   params.DestKey,
		},
	})
	if err != nil {
		return err
	}

	return result.Ok(info).Response(ctx)
}

// StatParams represents the request parameters for getting object metadata.
type StatParams struct {
	api.P

	// Key is the unique identifier of the object
	Key string `json:"key" validate:"required"`
}

// Stat retrieves metadata information about an object.
func (r *StorageResource) Stat(ctx fiber.Ctx, params StatParams) error {
	info, err := r.provider.StatObject(ctx.Context(), storage.StatObjectOptions{
		Key: params.Key,
	})
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return result.Err(i18n.T("object_not_found"))
		}

		return err
	}

	return result.Ok(info).Response(ctx)
}
