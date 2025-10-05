package memory

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cast"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/storage"
)

// MemoryProvider implements the storage.Provider interface using in-memory storage.
// This provider is intended for testing purposes only.
type MemoryProvider struct {
	mu      sync.RWMutex
	objects map[string]*objectData
}

type objectData struct {
	data         []byte
	contentType  string
	metadata     map[string]string
	lastModified time.Time
}

// NewMemoryProvider creates a new in-memory storage provider.
func NewMemoryProvider() storage.Provider {
	return &MemoryProvider{
		objects: make(map[string]*objectData),
	}
}

// Setup initializes the memory provider (no-op for in-memory storage).
func (p *MemoryProvider) Setup(ctx context.Context) error {
	return nil
}

// PutObject stores an object in memory.
func (p *MemoryProvider) PutObject(ctx context.Context, opts storage.PutObjectOptions) (*storage.ObjectInfo, error) {
	data, err := io.ReadAll(opts.Reader)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	p.objects[opts.Key] = &objectData{
		data:         data,
		contentType:  opts.ContentType,
		metadata:     opts.Metadata,
		lastModified: now,
	}

	return &storage.ObjectInfo{
		Bucket:       "memory",
		Key:          opts.Key,
		ETag:         cast.ToString(now.UnixNano()),
		Size:         int64(len(data)),
		ContentType:  opts.ContentType,
		LastModified: now,
		Metadata:     opts.Metadata,
	}, nil
}

// GetObject retrieves an object from memory.
func (p *MemoryProvider) GetObject(ctx context.Context, opts storage.GetObjectOptions) (io.ReadCloser, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	obj, exists := p.objects[opts.Key]
	if !exists {
		return nil, storage.ErrObjectNotFound
	}

	return io.NopCloser(bytes.NewReader(obj.data)), nil
}

// DeleteObject deletes a single object from memory.
func (p *MemoryProvider) DeleteObject(ctx context.Context, opts storage.DeleteObjectOptions) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.objects, opts.Key)

	return nil
}

// DeleteObjects deletes multiple objects from memory.
func (p *MemoryProvider) DeleteObjects(ctx context.Context, opts storage.DeleteObjectsOptions) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, key := range opts.Keys {
		delete(p.objects, key)
	}

	return nil
}

// ListObjects lists objects in memory.
func (p *MemoryProvider) ListObjects(ctx context.Context, opts storage.ListObjectsOptions) ([]storage.ObjectInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var objects []storage.ObjectInfo

	for key, obj := range p.objects {
		// Filter by prefix if specified
		if opts.Prefix != constants.Empty && !strings.HasPrefix(key, opts.Prefix) {
			continue
		}

		// Check recursive option
		if !opts.Recursive {
			// Remove prefix from key
			relativeKey := strings.TrimPrefix(key, opts.Prefix)
			// If there's a slash in the relative key, skip (it's in a subfolder)
			if strings.Contains(relativeKey, "/") {
				continue
			}
		}

		objects = append(objects, storage.ObjectInfo{
			Bucket:       "memory",
			Key:          key,
			ETag:         cast.ToString(obj.lastModified.UnixNano()),
			Size:         int64(len(obj.data)),
			ContentType:  obj.contentType,
			LastModified: obj.lastModified,
			Metadata:     obj.metadata,
		})

		// Enforce MaxKeys limit if specified
		if opts.MaxKeys > 0 && len(objects) >= opts.MaxKeys {
			break
		}
	}

	return objects, nil
}

// GetPresignedURL generates a mock presigned URL for in-memory storage.
func (p *MemoryProvider) GetPresignedURL(ctx context.Context, opts storage.PresignedURLOptions) (string, error) {
	// For memory provider, just return a mock URL
	return fmt.Sprintf("memory://%s?method=%s&expires=%d", opts.Key, opts.Method, opts.Expires), nil
}

// CopyObject copies an object within memory storage.
func (p *MemoryProvider) CopyObject(ctx context.Context, opts storage.CopyObjectOptions) (*storage.ObjectInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	source, exists := p.objects[opts.SourceKey]
	if !exists {
		return nil, storage.ErrObjectNotFound
	}

	// Copy the data
	dataCopy := make([]byte, len(source.data))
	copy(dataCopy, source.data)

	// Copy metadata
	metadataCopy := make(map[string]string, len(source.metadata))
	maps.Copy(metadataCopy, source.metadata)

	now := time.Now()
	p.objects[opts.DestKey] = &objectData{
		data:         dataCopy,
		contentType:  source.contentType,
		metadata:     metadataCopy,
		lastModified: now,
	}

	return &storage.ObjectInfo{
		Bucket:       "memory",
		Key:          opts.DestKey,
		ETag:         cast.ToString(now.UnixNano()),
		Size:         int64(len(dataCopy)),
		ContentType:  source.contentType,
		LastModified: now,
		Metadata:     metadataCopy,
	}, nil
}

// MoveObject moves an object by copying and then deleting the source.
func (p *MemoryProvider) MoveObject(ctx context.Context, opts storage.MoveObjectOptions) (info *storage.ObjectInfo, err error) {
	// Copy the object
	if info, err = p.CopyObject(ctx, opts.CopyObjectOptions); err != nil {
		return info, err
	}

	// Delete the source object
	if err = p.DeleteObject(ctx, storage.DeleteObjectOptions{
		Key: opts.SourceKey,
	}); err != nil {
		return nil, fmt.Errorf("copied successfully but failed to delete source: %w", err)
	}

	return info, err
}

// StatObject retrieves metadata about an object.
func (p *MemoryProvider) StatObject(ctx context.Context, opts storage.StatObjectOptions) (*storage.ObjectInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	obj, exists := p.objects[opts.Key]
	if !exists {
		return nil, storage.ErrObjectNotFound
	}

	return &storage.ObjectInfo{
		Bucket:       "memory",
		Key:          opts.Key,
		ETag:         cast.ToString(obj.lastModified.UnixNano()),
		Size:         int64(len(obj.data)),
		ContentType:  obj.contentType,
		LastModified: obj.lastModified,
		Metadata:     obj.metadata,
	}, nil
}

// PromoteObject moves an object from temporary storage to permanent storage.
func (p *MemoryProvider) PromoteObject(ctx context.Context, tempKey string) (*storage.ObjectInfo, error) {
	// Check if the key starts with temp/ prefix
	if !strings.HasPrefix(tempKey, storage.TempPrefix) {
		return nil, nil
	}

	// Remove the temp/ prefix to get the permanent key
	permanentKey := strings.TrimPrefix(tempKey, storage.TempPrefix)

	// Move the object
	return p.MoveObject(ctx, storage.MoveObjectOptions{
		CopyObjectOptions: storage.CopyObjectOptions{
			SourceKey: tempKey,
			DestKey:   permanentKey,
		},
	})
}
