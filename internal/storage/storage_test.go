package storage_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/app"
	appTest "github.com/ilxqx/vef-framework-go/internal/app/test"
	"github.com/ilxqx/vef-framework-go/result"
	storagePkg "github.com/ilxqx/vef-framework-go/storage"
	"github.com/ilxqx/vef-framework-go/testhelpers"
)

// StorageResourceTestSuite is the test suite for StorageResource.
type StorageResourceTestSuite struct {
	suite.Suite

	ctx            context.Context
	app            *app.App
	stop           func()
	minioContainer *testhelpers.MinIOContainer
	provider       storagePkg.Provider

	testBucketName  string
	testObjectKey   string
	testObjectData  []byte
	testContentType string
}

// SetupSuite runs once before all tests in the suite.
func (suite *StorageResourceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.testBucketName = testhelpers.TestMinioBucket
	suite.testObjectKey = "test-upload.txt"
	suite.testObjectData = []byte("Hello, Storage Api Test!")
	suite.testContentType = "text/plain"

	// Start MinIO container
	suite.minioContainer = testhelpers.NewMinIOContainer(suite.ctx, &suite.Suite)

	// Setup test app with storage module
	suite.setupTestApp()

	// Upload a test object for read operations
	reader := bytes.NewReader(suite.testObjectData)
	_, err := suite.provider.PutObject(suite.ctx, storagePkg.PutObjectOptions{
		Key:         suite.testObjectKey,
		Reader:      reader,
		Size:        int64(len(suite.testObjectData)),
		ContentType: suite.testContentType,
		Metadata: map[string]string{
			storagePkg.MetadataKeyOriginalFilename: "test.txt",
		},
	})
	suite.Require().NoError(err)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *StorageResourceTestSuite) TearDownSuite() {
	if suite.stop != nil {
		suite.stop()
	}

	if suite.minioContainer != nil {
		suite.minioContainer.Terminate(suite.ctx, &suite.Suite)
	}
}

func (suite *StorageResourceTestSuite) setupTestApp() {
	// Create MinIO config with bucket
	minioConfig := *suite.minioContainer.Config

	suite.app, suite.stop = appTest.NewTestApp(
		suite.T(),
		// Replace storage config with test values
		fx.Replace(
			&config.DatasourceConfig{
				Type: "sqlite",
			},
			&config.StorageConfig{
				Provider: "minio",
				MinIO:    minioConfig,
			},
		),
		fx.Populate(&suite.provider),
	)
}

// Helper methods

func (suite *StorageResourceTestSuite) makeApiRequest(body api.Request) *http.Response {
	jsonBody, err := encoding.ToJSON(body)
	suite.Require().NoError(err)

	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(jsonBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	return resp
}

func (suite *StorageResourceTestSuite) makeMultipartRequest(params map[string]string, fieldName, fileName string, fileContent []byte) *http.Response {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add params
	for key, value := range params {
		_ = writer.WriteField(key, value)
	}

	// Add file
	if fieldName != "" && fileName != "" {
		part, err := writer.CreateFormFile(fieldName, fileName)
		suite.Require().NoError(err)
		_, err = part.Write(fileContent)
		suite.Require().NoError(err)
	}

	err := writer.Close()
	suite.Require().NoError(err)

	req := httptest.NewRequest(fiber.MethodPost, "/api", body)
	req.Header.Set(fiber.HeaderContentType, writer.FormDataContentType())

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	return resp
}

func (suite *StorageResourceTestSuite) readBody(resp *http.Response) result.Result {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	suite.Require().NoError(err)
	res, err := encoding.FromJSON[result.Result](string(body))
	suite.Require().NoError(err)

	return *res
}

func (suite *StorageResourceTestSuite) readDataAsMap(data any) map[string]any {
	m, ok := data.(map[string]any)
	suite.Require().True(ok, "Expected data to be a map")

	return m
}

// Test Cases

func (suite *StorageResourceTestSuite) TestUploadSuccess() {
	uploadData := []byte("Uploaded via Api")

	params := map[string]string{
		"resource": "base/storage",
		"action":   "upload",
		"version":  "v1",
	}

	resp := suite.makeMultipartRequest(params, "file", "test.txt", uploadData)

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful upload")
	suite.Equal(i18n.T(result.OkMessage), body.Message)

	// Verify upload info
	data := suite.readDataAsMap(body.Data)
	suite.Equal(suite.testBucketName, data["bucket"])
	suite.NotEmpty(data["key"])
	suite.Contains(data["key"], ".txt", "Key should preserve file extension")
	suite.NotEmpty(data["eTag"])
	suite.NotZero(data["size"])

	// Verify the key format (should be temp/YYYY/MM/DD/{uuid}.txt)
	key := data["key"].(string)
	parts := strings.Split(key, "/")
	suite.GreaterOrEqual(len(parts), 4, "Key should have date-based path structure")
	suite.True(strings.HasSuffix(key, ".txt"), "Key should end with .txt")

	// Verify file was actually uploaded
	reader, err := suite.provider.GetObject(suite.ctx, storagePkg.GetObjectOptions{
		Key: key,
	})
	suite.Require().NoError(err)

	defer reader.Close()

	content, err := io.ReadAll(reader)
	suite.Require().NoError(err)
	suite.Equal(uploadData, content)

	// Verify original filename is automatically added to metadata
	info, err := suite.provider.StatObject(suite.ctx, storagePkg.StatObjectOptions{
		Key: key,
	})
	suite.Require().NoError(err)
	suite.NotNil(info.Metadata)
	suite.Equal("test.txt", info.Metadata[storagePkg.MetadataKeyOriginalFilename])
}

func (suite *StorageResourceTestSuite) TestUploadMissingFile() {
	params := map[string]string{
		"resource": "base/storage",
		"action":   "upload",
		"version":  "v1",
	}

	// No file provided
	resp := suite.makeMultipartRequest(params, "", "", nil)

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Expected upload to fail without file")
}

func (suite *StorageResourceTestSuite) TestUploadWithJSON() {
	// Upload requires multipart form, should fail with JSON
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "base/storage",
			Action:   "upload",
			Version:  "v1",
		},
		Params: map[string]any{
			"key": "test.txt",
		},
	})

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Expected upload to fail with JSON request")
}

func (suite *StorageResourceTestSuite) TestGetPresignedUrlForDownload() {
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "base/storage",
			Action:   "get_presigned_url",
			Version:  "v1",
		},
		Params: map[string]any{
			"key":     suite.testObjectKey,
			"expires": 3600,
			"method":  "GET",
		},
	})

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful presigned URL generation")

	data := suite.readDataAsMap(body.Data)
	url, ok := data["url"].(string)
	suite.True(ok)
	suite.NotEmpty(url)
	suite.Contains(url, suite.testBucketName)
	suite.Contains(url, suite.testObjectKey)

	// Verify we can download using the presigned URL
	downloadReq, err := http.NewRequestWithContext(suite.ctx, http.MethodGet, url, nil)
	suite.Require().NoError(err)

	downloadResp, err := http.DefaultClient.Do(downloadReq)
	suite.Require().NoError(err)

	defer downloadResp.Body.Close()

	suite.Equal(http.StatusOK, downloadResp.StatusCode)
	content, err := io.ReadAll(downloadResp.Body)
	suite.Require().NoError(err)
	suite.Equal(suite.testObjectData, content)
}

func (suite *StorageResourceTestSuite) TestGetPresignedUrlForUpload() {
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "base/storage",
			Action:   "get_presigned_url",
			Version:  "v1",
		},
		Params: map[string]any{
			"key":     "presigned-upload.txt",
			"expires": 3600,
			"method":  "PUT",
		},
	})

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful presigned URL generation")

	data := suite.readDataAsMap(body.Data)
	url, ok := data["url"].(string)
	suite.True(ok)
	suite.NotEmpty(url)
	suite.Contains(url, suite.testBucketName)

	// Upload using the presigned URL
	uploadData := []byte("Uploaded via presigned URL")
	uploadReq, err := http.NewRequestWithContext(suite.ctx, http.MethodPut, url, bytes.NewReader(uploadData))
	suite.Require().NoError(err)

	uploadResp, err := http.DefaultClient.Do(uploadReq)
	suite.Require().NoError(err)

	defer uploadResp.Body.Close()

	suite.Equal(http.StatusOK, uploadResp.StatusCode)
}

func (suite *StorageResourceTestSuite) TestGetPresignedUrlDefaultExpires() {
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "base/storage",
			Action:   "get_presigned_url",
			Version:  "v1",
		},
		Params: map[string]any{
			"key": suite.testObjectKey,
			// No expires specified, should default to 3600 seconds
		},
	})

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful presigned URL generation with default expires")

	data := suite.readDataAsMap(body.Data)
	suite.Contains(data, "url")
	suite.NotEmpty(data["url"])
}

func (suite *StorageResourceTestSuite) TestStatObjectSuccess() {
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "base/storage",
			Action:   "stat",
			Version:  "v1",
		},
		Params: map[string]any{
			"key": suite.testObjectKey,
		},
	})

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful stat")

	data := suite.readDataAsMap(body.Data)
	suite.Equal(suite.testBucketName, data["bucket"])
	suite.Equal(suite.testObjectKey, data["key"])
	suite.NotEmpty(data["eTag"])
	suite.NotZero(data["size"])
	suite.Equal(suite.testContentType, data["contentType"])
	suite.NotZero(data["lastModified"])
	suite.Equal("test.txt", suite.readDataAsMap(data["metadata"])[storagePkg.MetadataKeyOriginalFilename])
}

func (suite *StorageResourceTestSuite) TestStatObjectNotFound() {
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "base/storage",
			Action:   "stat",
			Version:  "v1",
		},
		Params: map[string]any{
			"key": "non-existent-key.txt",
		},
	})

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Expected stat to fail for non-existent object")
}

func (suite *StorageResourceTestSuite) TestListObjectsSuccess() {
	// Upload multiple test objects
	objects := map[string][]byte{
		"folder1/file1.txt": []byte("content1"),
		"folder1/file2.txt": []byte("content2"),
		"folder2/file3.txt": []byte("content3"),
	}

	for key, content := range objects {
		reader := bytes.NewReader(content)
		_, err := suite.provider.PutObject(suite.ctx, storagePkg.PutObjectOptions{
			Key:         key,
			Reader:      reader,
			Size:        int64(len(content)),
			ContentType: "text/plain",
		})
		suite.Require().NoError(err)
	}

	// List all objects
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "base/storage",
			Action:   "list",
			Version:  "v1",
		},
		Params: map[string]any{
			"recursive": true,
		},
	})

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful list")

	// Verify we have at least 4 objects: 3 from this test + 1 from setup
	// (other tests may have uploaded additional files)
	dataSlice, ok := body.Data.([]any)
	suite.True(ok, "Expected data to be a slice")
	suite.GreaterOrEqual(len(dataSlice), 4, "Should have at least 4 objects: 3 uploaded in this test + 1 from setup")
}

func (suite *StorageResourceTestSuite) TestListObjectsWithPrefix() {
	// Upload test objects with different prefixes
	objects := map[string][]byte{
		"prefix-test/file1.txt": []byte("content1"),
		"prefix-test/file2.txt": []byte("content2"),
		"other/file3.txt":       []byte("content3"),
	}

	for key, content := range objects {
		reader := bytes.NewReader(content)
		_, err := suite.provider.PutObject(suite.ctx, storagePkg.PutObjectOptions{
			Key:         key,
			Reader:      reader,
			Size:        int64(len(content)),
			ContentType: "text/plain",
		})
		suite.Require().NoError(err)
	}

	// List with prefix
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "base/storage",
			Action:   "list",
			Version:  "v1",
		},
		Params: map[string]any{
			"prefix":    "prefix-test/",
			"recursive": true,
		},
	})

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful list with prefix")

	dataSlice, ok := body.Data.([]any)
	suite.True(ok)
	suite.GreaterOrEqual(len(dataSlice), 2)

	// Verify all returned objects have the prefix
	for _, item := range dataSlice {
		obj := item.(map[string]any)
		key := obj["key"].(string)
		suite.Contains(key, "prefix-test/")
	}
}

func (suite *StorageResourceTestSuite) TestListObjectsWithMaxKeys() {
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "base/storage",
			Action:   "list",
			Version:  "v1",
		},
		Params: map[string]any{
			"recursive": true,
			"maxKeys":   1,
		},
	})

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful list with maxKeys")

	dataSlice, ok := body.Data.([]any)
	suite.True(ok)
	suite.Equal(1, len(dataSlice), "MaxKeys should limit results to 1")
}

func (suite *StorageResourceTestSuite) TestUploadWithMetadata() {
	uploadData := []byte("Test with metadata")

	params := map[string]string{
		"resource": "base/storage",
		"action":   "upload",
		"version":  "v1",
		"params":   `{"metadata":{"author":"test-suite","version":"1.0"}}`,
	}

	resp := suite.makeMultipartRequest(params, "file", "test-metadata.txt", uploadData)

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful upload with metadata")

	// Get the uploaded key from response
	data := suite.readDataAsMap(body.Data)
	uploadKey := data["key"].(string)

	// Verify metadata is stored (including user metadata and original filename)
	info, err := suite.provider.StatObject(suite.ctx, storagePkg.StatObjectOptions{
		Key: uploadKey,
	})
	suite.Require().NoError(err)
	suite.NotNil(info.Metadata)

	// Verify user-provided metadata (MinIO canonicalizes keys to Title-Case)
	suite.Equal("test-suite", info.Metadata["Author"])
	suite.Equal("1.0", info.Metadata["Version"])

	// Verify original filename is automatically added
	suite.Equal("test-metadata.txt", info.Metadata[storagePkg.MetadataKeyOriginalFilename])
}

func (suite *StorageResourceTestSuite) TestUploadWithContentType() {
	uploadData := []byte(`{"test": "data"}`)

	params := map[string]string{
		"resource": "base/storage",
		"action":   "upload",
		"version":  "v1",
		"params":   `{"contentType":"application/json"}`,
	}

	resp := suite.makeMultipartRequest(params, "file", "test.json", uploadData)

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful upload with content type")

	// Get the uploaded key from response
	data := suite.readDataAsMap(body.Data)
	uploadKey := data["key"].(string)

	// Verify content type
	info, err := suite.provider.StatObject(suite.ctx, storagePkg.StatObjectOptions{
		Key: uploadKey,
	})
	suite.Require().NoError(err)
	suite.Equal("application/json", info.ContentType)
}

func (suite *StorageResourceTestSuite) TestGetPresignedUrlExpiration() {
	// Test with custom expiration
	customExpires := 7200 // 2 hours

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "base/storage",
			Action:   "get_presigned_url",
			Version:  "v1",
		},
		Params: map[string]any{
			"key":     suite.testObjectKey,
			"expires": float64(customExpires),
		},
	})

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful presigned URL generation with custom expiration")

	data := suite.readDataAsMap(body.Data)
	url := data["url"].(string)
	suite.NotEmpty(url)

	// Verify the URL contains expiration parameter
	suite.Contains(url, "X-Amz-Expires")
}

func (suite *StorageResourceTestSuite) TestConcurrentUploads() {
	// Test concurrent uploads to verify thread safety
	numUploads := 5
	done := make(chan bool, numUploads)

	for i := range numUploads {
		go func(index int) {
			defer func() { done <- true }()

			uploadData := fmt.Appendf(nil, "Concurrent upload %d", index)
			params := map[string]string{
				"resource": "base/storage",
				"action":   "upload",
				"version":  "v1",
			}

			resp := suite.makeMultipartRequest(params, "file", fmt.Sprintf("test%d.txt", index), uploadData)
			suite.Equal(200, resp.StatusCode)

			body := suite.readBody(resp)
			suite.True(body.IsOk())
		}(i)
	}

	// Wait for all uploads to complete
	timeout := time.After(30 * time.Second)

	for range numUploads {
		select {
		case <-done:
			// Upload completed
		case <-timeout:
			suite.Fail("Concurrent upload test timed out")

			return
		}
	}
}

// TestStorageResourceSuite runs the test suite.
func TestStorageResourceSuite(t *testing.T) {
	suite.Run(t, new(StorageResourceTestSuite))
}
