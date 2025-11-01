package minio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/storage"
)

// MinIOService implements the storage.Service interface using MinIO.
type MinIOService struct {
	client *minio.Client
	bucket string
}

// NewMinIOService creates a new MinIO storage service.
func NewMinIOService(cfg config.MinIOConfig, appCfg *config.AppConfig) (storage.Service, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, constants.Empty),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &MinIOService{
		client: client,
		bucket: lo.CoalesceOrEmpty(cfg.Bucket, appCfg.Name, "vef-app"),
	}, nil
}

// Init initializes the MinIO provider by ensuring the bucket exists.
func (s *MinIOService) Init(ctx context.Context) error {
	// Check if bucket exists
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	// Create bucket if it doesn't exist
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", s.bucket, err)
		}

		// Set public read policy for the bucket
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, s.bucket)

		if err := s.client.SetBucketPolicy(ctx, s.bucket, policy); err != nil {
			return fmt.Errorf("failed to set public read policy for bucket %s: %w", s.bucket, err)
		}
	}

	return nil
}

// PutObject uploads an object to MinIO.
func (s *MinIOService) PutObject(ctx context.Context, opts storage.PutObjectOptions) (*storage.ObjectInfo, error) {
	uploadOpts := minio.PutObjectOptions{
		ContentType:  opts.ContentType,
		UserMetadata: opts.Metadata,
	}

	info, err := s.client.PutObject(ctx, s.bucket, opts.Key, opts.Reader, opts.Size, uploadOpts)
	if err != nil {
		return nil, s.translateError(err)
	}

	return &storage.ObjectInfo{
		Bucket:       info.Bucket,
		Key:          info.Key,
		ETag:         info.ETag,
		Size:         info.Size,
		ContentType:  opts.ContentType,
		LastModified: info.LastModified,
		Metadata:     opts.Metadata,
	}, nil
}

// GetObject retrieves an object from MinIO.
func (s *MinIOService) GetObject(ctx context.Context, opts storage.GetObjectOptions) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucket, opts.Key, minio.GetObjectOptions{})
	if err != nil {
		return nil, s.translateError(err)
	}

	// Verify the object exists by calling Stat
	if _, err = object.Stat(); err != nil {
		_ = object.Close()

		return nil, s.translateError(err)
	}

	return object, nil
}

// DeleteObject deletes a single object from MinIO.
func (s *MinIOService) DeleteObject(ctx context.Context, opts storage.DeleteObjectOptions) error {
	err := s.client.RemoveObject(ctx, s.bucket, opts.Key, minio.RemoveObjectOptions{})
	if err != nil {
		return s.translateError(err)
	}

	return nil
}

// DeleteObjects deletes multiple objects from MinIO.
func (s *MinIOService) DeleteObjects(ctx context.Context, opts storage.DeleteObjectsOptions) error {
	objectsCh := make(chan minio.ObjectInfo, len(opts.Keys))

	// Send object keys to delete
	go func() {
		defer close(objectsCh)

		for _, key := range opts.Keys {
			objectsCh <- minio.ObjectInfo{Key: key}
		}
	}()

	// Remove objects
	errorCh := s.client.RemoveObjects(ctx, s.bucket, objectsCh, minio.RemoveObjectsOptions{})

	// Check for errors
	for err := range errorCh {
		if err.Err != nil {
			return s.translateError(err.Err)
		}
	}

	return nil
}

// ListObjects lists objects in a MinIO bucket.
func (s *MinIOService) ListObjects(ctx context.Context, opts storage.ListObjectsOptions) ([]storage.ObjectInfo, error) {
	listOpts := minio.ListObjectsOptions{
		Prefix:       opts.Prefix,
		Recursive:    opts.Recursive,
		MaxKeys:      opts.MaxKeys,
		WithMetadata: true,
	}

	var objects []storage.ObjectInfo

	for object := range s.client.ListObjects(ctx, s.bucket, listOpts) {
		if object.Err != nil {
			return nil, s.translateError(object.Err)
		}

		objects = append(objects, storage.ObjectInfo{
			Bucket:       s.bucket,
			Key:          object.Key,
			ETag:         object.ETag,
			Size:         object.Size,
			ContentType:  object.ContentType,
			LastModified: object.LastModified,
			Metadata:     object.UserMetadata,
		})

		// Enforce MaxKeys limit if specified
		if opts.MaxKeys > 0 && len(objects) >= opts.MaxKeys {
			break
		}
	}

	return objects, nil
}

// GetPresignedURL generates a presigned URL for temporary access.
func (s *MinIOService) GetPresignedURL(ctx context.Context, opts storage.PresignedURLOptions) (string, error) {
	var (
		urlStr string
		err    error
	)

	switch opts.Method {
	case http.MethodGet, constants.Empty:
		u, e := s.client.PresignedGetObject(ctx, s.bucket, opts.Key, opts.Expires, nil)
		if e == nil {
			urlStr = u.String()
		}

		err = e

	case http.MethodPut:
		u, e := s.client.PresignedPutObject(ctx, s.bucket, opts.Key, opts.Expires)
		if e == nil {
			urlStr = u.String()
		}

		err = e

	default:
		return constants.Empty, fmt.Errorf("%w: %s", ErrUnsupportedHTTPMethod, opts.Method)
	}

	if err != nil {
		return constants.Empty, s.translateError(err)
	}

	return urlStr, nil
}

// CopyObject copies an object within MinIO.
func (s *MinIOService) CopyObject(ctx context.Context, opts storage.CopyObjectOptions) (*storage.ObjectInfo, error) {
	src := minio.CopySrcOptions{
		Bucket: s.bucket,
		Object: opts.SourceKey,
	}

	dst := minio.CopyDestOptions{
		Bucket: s.bucket,
		Object: opts.DestKey,
	}

	info, err := s.client.CopyObject(ctx, dst, src)
	if err != nil {
		return nil, s.translateError(err)
	}

	return &storage.ObjectInfo{
		Bucket:       info.Bucket,
		Key:          info.Key,
		ETag:         info.ETag,
		Size:         info.Size,
		LastModified: info.LastModified,
	}, nil
}

// MoveObject moves an object by copying and then deleting the source.
func (s *MinIOService) MoveObject(ctx context.Context, opts storage.MoveObjectOptions) (info *storage.ObjectInfo, err error) {
	// Copy the object
	if info, err = s.CopyObject(ctx, opts.CopyObjectOptions); err != nil {
		return info, err
	}

	// Delete the source object
	if err = s.DeleteObject(ctx, storage.DeleteObjectOptions{
		Key: opts.SourceKey,
	}); err != nil {
		return nil, fmt.Errorf("copied successfully but failed to delete source: %w", err)
	}

	return info, err
}

// StatObject retrieves metadata about an object.
func (s *MinIOService) StatObject(ctx context.Context, opts storage.StatObjectOptions) (*storage.ObjectInfo, error) {
	info, err := s.client.StatObject(ctx, s.bucket, opts.Key, minio.StatObjectOptions{})
	if err != nil {
		return nil, s.translateError(err)
	}

	return &storage.ObjectInfo{
		Bucket:       s.bucket,
		Key:          info.Key,
		ETag:         info.ETag,
		Size:         info.Size,
		ContentType:  info.ContentType,
		LastModified: info.LastModified,
		Metadata:     info.UserMetadata,
	}, nil
}

// PromoteObject moves an object from temporary storage to permanent storage.
func (s *MinIOService) PromoteObject(ctx context.Context, tempKey string) (*storage.ObjectInfo, error) {
	// Check if the key starts with temp/ prefix
	if !strings.HasPrefix(tempKey, storage.TempPrefix) {
		return nil, nil
	}

	// Remove the temp/ prefix to get the permanent key
	permanentKey := strings.TrimPrefix(tempKey, storage.TempPrefix)

	// Move the object
	return s.MoveObject(ctx, storage.MoveObjectOptions{
		CopyObjectOptions: storage.CopyObjectOptions{
			SourceKey: tempKey,
			DestKey:   permanentKey,
		},
	})
}

// translateError converts MinIO errors to storage package errors.
func (s *MinIOService) translateError(err error) error {
	if err == nil {
		return nil
	}

	// Convert minio-specific errors to storage errors
	var minioErr minio.ErrorResponse

	ok := errors.As(err, &minioErr)
	if !ok {
		return err
	}

	switch minioErr.Code {
	case "NoSuchBucket":
		return storage.ErrBucketNotFound
	case "NoSuchKey":
		return storage.ErrObjectNotFound
	case "InvalidBucketName":
		return storage.ErrInvalidBucketName
	case "AccessDenied":
		return storage.ErrAccessDenied
	default:
		return err
	}
}
