package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ilxqx/vef-framework-go/storage"
)

// mockStorageService is a mock implementation of storage.Service for testing
type mockStorageService struct {
	mock.Mock
}

func (m *mockStorageService) PutObject(_ context.Context, _ storage.PutObjectOptions) (*storage.ObjectInfo, error) {
	return nil, nil
}

func (m *mockStorageService) GetObject(_ context.Context, opts storage.GetObjectOptions) (io.ReadCloser, error) {
	args := m.Called(opts)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *mockStorageService) DeleteObject(_ context.Context, _ storage.DeleteObjectOptions) error {
	return nil
}

func (m *mockStorageService) DeleteObjects(_ context.Context, _ storage.DeleteObjectsOptions) error {
	return nil
}

func (m *mockStorageService) ListObjects(_ context.Context, _ storage.ListObjectsOptions) ([]storage.ObjectInfo, error) {
	return nil, nil
}

func (m *mockStorageService) GetPresignedUrl(_ context.Context, _ storage.PresignedURLOptions) (string, error) {
	return "", nil
}

func (m *mockStorageService) CopyObject(_ context.Context, _ storage.CopyObjectOptions) (*storage.ObjectInfo, error) {
	return nil, nil
}

func (m *mockStorageService) MoveObject(_ context.Context, _ storage.MoveObjectOptions) (*storage.ObjectInfo, error) {
	return nil, nil
}

func (m *mockStorageService) StatObject(_ context.Context, opts storage.StatObjectOptions) (*storage.ObjectInfo, error) {
	args := m.Called(opts)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.ObjectInfo), args.Error(1)
}

func (m *mockStorageService) PromoteObject(_ context.Context, _ string) (*storage.ObjectInfo, error) {
	return nil, nil
}

func TestProxyMiddleware(t *testing.T) {
	// Helper function to create a configured Fiber app with error handler
	createApp := func() *fiber.App {
		return fiber.New(fiber.Config{
			ErrorHandler: func(_ fiber.Ctx, _ error) error {
				// Return 200 for business errors (matching framework behavior)
				return nil
			},
		})
	}

	t.Run("successful file download", func(t *testing.T) {
		mockService := new(mockStorageService)
		fileContent := []byte("test file content")

		// Setup expectations
		mockService.On("GetObject", storage.GetObjectOptions{
			Key: "temp/2025/01/15/test.jpg",
		}).Return(io.NopCloser(bytes.NewReader(fileContent)), nil)

		mockService.On("StatObject", storage.StatObjectOptions{
			Key: "temp/2025/01/15/test.jpg",
		}).Return(&storage.ObjectInfo{
			ContentType: "image/jpeg",
			ETag:        "etag123",
		}, nil)

		app := createApp()
		middleware := NewProxyMiddleware(mockService)
		middleware.Apply(app)

		req := httptest.NewRequest(http.MethodGet, "/files/temp/2025/01/15/test.jpg", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "image/jpeg", resp.Header.Get("Content-Type"))

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, fileContent, body)

		mockService.AssertExpectations(t)
	})

	t.Run("file not found", func(t *testing.T) {
		mockService := new(mockStorageService)

		// Setup expectations
		mockService.On("GetObject", storage.GetObjectOptions{
			Key: "nonexistent.jpg",
		}).Return(nil, storage.ErrObjectNotFound)

		app := createApp()
		middleware := NewProxyMiddleware(mockService)
		middleware.Apply(app)

		req := httptest.NewRequest(http.MethodGet, "/files/nonexistent.jpg", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockService.AssertExpectations(t)
	})

	t.Run("empty file key", func(t *testing.T) {
		app := createApp()
		middleware := NewProxyMiddleware(nil)
		middleware.Apply(app)

		req := httptest.NewRequest(http.MethodGet, "/files/", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("URL encoded file key", func(t *testing.T) {
		mockService := new(mockStorageService)
		fileContent := []byte("test content")

		// Setup expectations - the key should be decoded
		mockService.On("GetObject", storage.GetObjectOptions{
			Key: "temp/测试文件.jpg",
		}).Return(io.NopCloser(bytes.NewReader(fileContent)), nil)

		mockService.On("StatObject", storage.StatObjectOptions{
			Key: "temp/测试文件.jpg",
		}).Return(&storage.ObjectInfo{
			ContentType: "image/jpeg",
		}, nil)

		app := createApp()
		middleware := NewProxyMiddleware(mockService)
		middleware.Apply(app)

		// URL encode the Chinese characters
		req := httptest.NewRequest(http.MethodGet, "/files/temp/%E6%B5%8B%E8%AF%95%E6%96%87%E4%BB%B6.jpg", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockService.AssertExpectations(t)
	})

	t.Run("storage error", func(t *testing.T) {
		mockService := new(mockStorageService)

		// Setup expectations
		mockService.On("GetObject", storage.GetObjectOptions{
			Key: "error.jpg",
		}).Return(nil, errors.New("storage error"))

		app := createApp()
		middleware := NewProxyMiddleware(mockService)
		middleware.Apply(app)

		req := httptest.NewRequest(http.MethodGet, "/files/error.jpg", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockService.AssertExpectations(t)
	})

	t.Run("content type fallback from extension when stat fails", func(t *testing.T) {
		mockService := new(mockStorageService)
		fileContent := []byte("test content")

		mockService.On("GetObject", storage.GetObjectOptions{
			Key: "test.png",
		}).Return(io.NopCloser(bytes.NewReader(fileContent)), nil)

		// StatObject fails - should fallback to extension-based detection
		mockService.On("StatObject", storage.StatObjectOptions{
			Key: "test.png",
		}).Return(nil, errors.New("stat failed"))

		app := createApp()
		middleware := NewProxyMiddleware(mockService)
		middleware.Apply(app)

		req := httptest.NewRequest(http.MethodGet, "/files/test.png", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "image/png", resp.Header.Get("Content-Type"))

		mockService.AssertExpectations(t)
	})

	t.Run("content type fallback from extension when content type empty", func(t *testing.T) {
		mockService := new(mockStorageService)
		fileContent := []byte("test content")

		mockService.On("GetObject", storage.GetObjectOptions{
			Key: "document.pdf",
		}).Return(io.NopCloser(bytes.NewReader(fileContent)), nil)

		// StatObject succeeds but ContentType is empty
		mockService.On("StatObject", storage.StatObjectOptions{
			Key: "document.pdf",
		}).Return(&storage.ObjectInfo{
			ContentType: "",
			ETag:        "etag456",
		}, nil)

		app := createApp()
		middleware := NewProxyMiddleware(mockService)
		middleware.Apply(app)

		req := httptest.NewRequest(http.MethodGet, "/files/document.pdf", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/pdf", resp.Header.Get("Content-Type"))
		assert.Equal(t, "etag456", resp.Header.Get("ETag"))

		mockService.AssertExpectations(t)
	})
}
