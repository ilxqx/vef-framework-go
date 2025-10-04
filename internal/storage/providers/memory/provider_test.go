package memory

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/ilxqx/vef-framework-go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryProvider(t *testing.T) {
	ctx := context.Background()
	provider := NewMemoryProvider()

	// Test Setup
	t.Run("Setup", func(t *testing.T) {
		err := provider.Setup(ctx)
		assert.NoError(t, err)
	})

	// Test PutObject
	t.Run("PutObject", func(t *testing.T) {
		data := []byte("Hello, Memory Storage!")
		reader := bytes.NewReader(data)

		info, err := provider.PutObject(ctx, storage.PutObjectOptions{
			Key:         "test.txt",
			Reader:      reader,
			Size:        int64(len(data)),
			ContentType: "text/plain",
			Metadata: map[string]string{
				"author": "test",
			},
		})

		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "test.txt", info.Key)
		assert.Equal(t, int64(len(data)), info.Size)
		assert.Equal(t, "text/plain", info.ContentType)
	})

	// Test GetObject
	t.Run("GetObject", func(t *testing.T) {
		expectedData := []byte("Hello, Memory Storage!")

		reader, err := provider.GetObject(ctx, storage.GetObjectOptions{
			Key: "test.txt",
		})

		require.NoError(t, err)
		require.NotNil(t, reader)
		defer reader.Close()

		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, expectedData, data)
	})

	// Test GetObject NotFound
	t.Run("GetObject_NotFound", func(t *testing.T) {
		reader, err := provider.GetObject(ctx, storage.GetObjectOptions{
			Key: "nonexistent.txt",
		})

		assert.Error(t, err)
		assert.Nil(t, reader)
		assert.Equal(t, storage.ErrObjectNotFound, err)
	})

	// Test StatObject
	t.Run("StatObject", func(t *testing.T) {
		info, err := provider.StatObject(ctx, storage.StatObjectOptions{
			Key: "test.txt",
		})

		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "test.txt", info.Key)
		assert.Equal(t, "text/plain", info.ContentType)
	})

	// Test CopyObject
	t.Run("CopyObject", func(t *testing.T) {
		info, err := provider.CopyObject(ctx, storage.CopyObjectOptions{
			SourceKey: "test.txt",
			DestKey:   "test-copy.txt",
		})

		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "test-copy.txt", info.Key)

		// Verify both files exist
		reader, err := provider.GetObject(ctx, storage.GetObjectOptions{
			Key: "test-copy.txt",
		})
		require.NoError(t, err)
		defer reader.Close()

		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, []byte("Hello, Memory Storage!"), data)
	})

	// Test MoveObject
	t.Run("MoveObject", func(t *testing.T) {
		info, err := provider.MoveObject(ctx, storage.MoveObjectOptions{
			CopyObjectOptions: storage.CopyObjectOptions{
				SourceKey: "test-copy.txt",
				DestKey:   "test-moved.txt",
			},
		})

		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "test-moved.txt", info.Key)

		// Verify source is deleted
		_, err = provider.GetObject(ctx, storage.GetObjectOptions{
			Key: "test-copy.txt",
		})
		assert.Error(t, err)
		assert.Equal(t, storage.ErrObjectNotFound, err)

		// Verify destination exists
		reader, err := provider.GetObject(ctx, storage.GetObjectOptions{
			Key: "test-moved.txt",
		})
		require.NoError(t, err)
		defer reader.Close()
	})

	// Test ListObjects
	t.Run("ListObjects", func(t *testing.T) {
		// Add more objects
		provider.PutObject(ctx, storage.PutObjectOptions{
			Key:    "folder/file1.txt",
			Reader: bytes.NewReader([]byte("file1")),
			Size:   5,
		})
		provider.PutObject(ctx, storage.PutObjectOptions{
			Key:    "folder/file2.txt",
			Reader: bytes.NewReader([]byte("file2")),
			Size:   5,
		})

		// List all objects
		objects, err := provider.ListObjects(ctx, storage.ListObjectsOptions{
			Recursive: true,
		})

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(objects), 3)

		// List with prefix
		objects, err = provider.ListObjects(ctx, storage.ListObjectsOptions{
			Prefix:    "folder/",
			Recursive: true,
		})

		require.NoError(t, err)
		assert.Equal(t, 2, len(objects))
	})

	// Test PromoteObject
	t.Run("PromoteObject", func(t *testing.T) {
		// Upload a temp file
		tempKey := storage.TempPrefix + "2025/01/15/test.txt"
		provider.PutObject(ctx, storage.PutObjectOptions{
			Key:    tempKey,
			Reader: bytes.NewReader([]byte("temp content")),
			Size:   12,
		})

		// Promote it
		info, err := provider.PromoteObject(ctx, tempKey)
		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "2025/01/15/test.txt", info.Key)

		// Verify temp file is deleted
		_, err = provider.GetObject(ctx, storage.GetObjectOptions{Key: tempKey})
		assert.Error(t, err)

		// Verify permanent file exists
		reader, err := provider.GetObject(ctx, storage.GetObjectOptions{
			Key: "2025/01/15/test.txt",
		})
		require.NoError(t, err)
		defer reader.Close()
	})

	// Test DeleteObject
	t.Run("DeleteObject", func(t *testing.T) {
		err := provider.DeleteObject(ctx, storage.DeleteObjectOptions{
			Key: "test.txt",
		})

		assert.NoError(t, err)

		// Verify it's deleted
		_, err = provider.GetObject(ctx, storage.GetObjectOptions{
			Key: "test.txt",
		})
		assert.Error(t, err)
	})

	// Test DeleteObjects
	t.Run("DeleteObjects", func(t *testing.T) {
		// Upload multiple objects
		keys := []string{"delete1.txt", "delete2.txt", "delete3.txt"}
		for _, key := range keys {
			provider.PutObject(ctx, storage.PutObjectOptions{
				Key:    key,
				Reader: bytes.NewReader([]byte("content")),
				Size:   7,
			})
		}

		// Delete them all
		err := provider.DeleteObjects(ctx, storage.DeleteObjectsOptions{
			Keys: keys,
		})
		assert.NoError(t, err)

		// Verify all are deleted
		for _, key := range keys {
			_, err := provider.GetObject(ctx, storage.GetObjectOptions{Key: key})
			assert.Error(t, err)
		}
	})
}
