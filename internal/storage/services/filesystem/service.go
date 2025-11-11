package filesystem

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/storage"
)

// Service implements storage.Service using local filesystem.
type Service struct {
	root string
}

// New creates a new filesystem storage service.
func New(cfg config.FilesystemConfig) (storage.Service, error) {
	root := cfg.Root
	if root == constants.Empty {
		root = "./storage"
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create storage root directory: %w", err)
	}

	return &Service{root: root}, nil
}

func (s *Service) resolvePath(key string) string {
	return filepath.Join(s.root, filepath.FromSlash(key))
}

// PutObject stores a file on the filesystem.
func (s *Service) PutObject(ctx context.Context, opts storage.PutObjectOptions) (*storage.ObjectInfo, error) {
	path := s.resolvePath(opts.Key)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	defer func() { _ = file.Close() }()

	hasher := md5.New()
	writer := io.MultiWriter(file, hasher)

	written, err := io.Copy(writer, opts.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	etag := hex.EncodeToString(hasher.Sum(nil))

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return &storage.ObjectInfo{
		Bucket:       "filesystem",
		Key:          opts.Key,
		ETag:         etag,
		Size:         written,
		ContentType:  opts.ContentType,
		LastModified: stat.ModTime(),
		Metadata:     opts.Metadata,
	}, nil
}

// GetObject retrieves a file from the filesystem.
func (s *Service) GetObject(ctx context.Context, opts storage.GetObjectOptions) (io.ReadCloser, error) {
	path := s.resolvePath(opts.Key)

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.ErrObjectNotFound
		}

		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// DeleteObject deletes a file from the filesystem.
func (s *Service) DeleteObject(ctx context.Context, opts storage.DeleteObjectOptions) error {
	path := s.resolvePath(opts.Key)

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	s.cleanupEmptyDirs(filepath.Dir(path))

	return nil
}

// DeleteObjects deletes multiple files in batch.
func (s *Service) DeleteObjects(ctx context.Context, opts storage.DeleteObjectsOptions) error {
	for _, key := range opts.Keys {
		if err := s.DeleteObject(ctx, storage.DeleteObjectOptions{Key: key}); err != nil {
			return err
		}
	}

	return nil
}

// ListObjects lists files with optional prefix filtering.
func (s *Service) ListObjects(ctx context.Context, opts storage.ListObjectsOptions) ([]storage.ObjectInfo, error) {
	var objects []storage.ObjectInfo

	prefix := opts.Prefix
	searchPath := s.resolvePath(prefix)

	err := filepath.WalkDir(searchPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) || os.IsNotExist(err) {
				return nil
			}

			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(s.root, path)
		if err != nil {
			return err
		}

		key := filepath.ToSlash(relPath)

		if prefix != constants.Empty && !strings.HasPrefix(key, prefix) {
			return nil
		}

		if !opts.Recursive {
			relativeKey := strings.TrimPrefix(key, prefix)
			if strings.Contains(relativeKey, constants.Slash) {
				return nil
			}
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		contentType := mime.TypeByExtension(filepath.Ext(path))

		objects = append(objects, storage.ObjectInfo{
			Bucket:       "filesystem",
			Key:          key,
			ETag:         constants.Empty,
			Size:         info.Size(),
			ContentType:  contentType,
			LastModified: info.ModTime(),
		})

		if opts.MaxKeys > 0 && len(objects) >= opts.MaxKeys {
			return io.EOF
		}

		return nil
	})

	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	return objects, nil
}

// GetPresignedUrl generates a file:// Url for local filesystem.
func (s *Service) GetPresignedUrl(ctx context.Context, opts storage.PresignedURLOptions) (string, error) {
	path := s.resolvePath(opts.Key)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return fmt.Sprintf("file://%s", absPath), nil
}

// CopyObject copies a file within the filesystem.
func (s *Service) CopyObject(ctx context.Context, opts storage.CopyObjectOptions) (*storage.ObjectInfo, error) {
	srcPath := s.resolvePath(opts.SourceKey)
	destPath := s.resolvePath(opts.DestKey)

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.ErrObjectNotFound
		}

		return nil, fmt.Errorf("failed to open source file: %w", err)
	}

	defer func() { _ = src.Close() }()

	dest, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}

	defer func() { _ = dest.Close() }()

	hasher := md5.New()
	writer := io.MultiWriter(dest, hasher)

	written, err := io.Copy(writer, src)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	etag := hex.EncodeToString(hasher.Sum(nil))

	stat, err := dest.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat destination file: %w", err)
	}

	contentType := mime.TypeByExtension(filepath.Ext(destPath))

	return &storage.ObjectInfo{
		Bucket:       "filesystem",
		Key:          opts.DestKey,
		ETag:         etag,
		Size:         written,
		ContentType:  contentType,
		LastModified: stat.ModTime(),
	}, nil
}

// MoveObject moves a file by copying and deleting the source.
func (s *Service) MoveObject(ctx context.Context, opts storage.MoveObjectOptions) (*storage.ObjectInfo, error) {
	info, err := s.CopyObject(ctx, opts.CopyObjectOptions)
	if err != nil {
		return nil, err
	}

	if err := s.DeleteObject(ctx, storage.DeleteObjectOptions{Key: opts.SourceKey}); err != nil {
		return nil, fmt.Errorf("copied successfully but failed to delete source: %w", err)
	}

	return info, nil
}

// StatObject retrieves file metadata.
func (s *Service) StatObject(ctx context.Context, opts storage.StatObjectOptions) (*storage.ObjectInfo, error) {
	path := s.resolvePath(opts.Key)

	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.ErrObjectNotFound
		}

		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	etag, err := s.calculateMd5(path)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate MD5: %w", err)
	}

	contentType := mime.TypeByExtension(filepath.Ext(path))

	return &storage.ObjectInfo{
		Bucket:       "filesystem",
		Key:          opts.Key,
		ETag:         etag,
		Size:         stat.Size(),
		ContentType:  contentType,
		LastModified: stat.ModTime(),
	}, nil
}

// PromoteObject moves a file from temp/ to permanent storage.
func (s *Service) PromoteObject(ctx context.Context, tempKey string) (*storage.ObjectInfo, error) {
	if !strings.HasPrefix(tempKey, storage.TempPrefix) {
		return nil, nil
	}

	permanentKey := strings.TrimPrefix(tempKey, storage.TempPrefix)

	return s.MoveObject(ctx, storage.MoveObjectOptions{
		CopyObjectOptions: storage.CopyObjectOptions{
			SourceKey: tempKey,
			DestKey:   permanentKey,
		},
	})
}

// cleanupEmptyDirs removes empty parent directories up to root.
func (s *Service) cleanupEmptyDirs(dir string) {
	for dir != s.root && strings.HasPrefix(dir, s.root) {
		if err := os.Remove(dir); err != nil {
			break
		}

		dir = filepath.Dir(dir)
	}
}

// calculateMd5 calculates the MD5 hash of a file.
func (s *Service) calculateMd5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return constants.Empty, err
	}

	defer func() { _ = file.Close() }()

	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return constants.Empty, err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
