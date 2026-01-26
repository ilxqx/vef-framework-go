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
	"github.com/ilxqx/vef-framework-go/internal/apptest"
	"github.com/ilxqx/vef-framework-go/internal/testhelpers"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/storage"
)

// StorageResourceTestSuite tests the storage API resource functionality.
// Tests cover file upload, download, presigned URLs, object metadata, and listing operations.
type StorageResourceTestSuite struct {
	suite.Suite

	ctx            context.Context
	app            *app.App
	stop           func()
	minioContainer *testhelpers.MinIOContainer
	service        storage.Service

	testBucketName  string
	testObjectKey   string
	testObjectData  []byte
	testContentType string
}

// SetupSuite runs once before all tests in the suite.
func (suite *StorageResourceTestSuite) SetupSuite() {
	suite.T().Log("Setting up StorageResourceTestSuite - starting MinIO container and test app")

	suite.ctx = context.Background()
	suite.testBucketName = testhelpers.TestMinioBucket
	suite.testObjectKey = "test-upload.txt"
	suite.testObjectData = []byte("Hello, Storage Api Test!")
	suite.testContentType = "text/plain"

	suite.minioContainer = testhelpers.NewMinIOContainer(suite.ctx, &suite.Suite)

	suite.setupTestApp()

	reader := bytes.NewReader(suite.testObjectData)
	_, err := suite.service.PutObject(suite.ctx, storage.PutObjectOptions{
		Key:         suite.testObjectKey,
		Reader:      reader,
		Size:        int64(len(suite.testObjectData)),
		ContentType: suite.testContentType,
		Metadata: map[string]string{
			storage.MetadataKeyOriginalFilename: "test.txt",
		},
	})
	suite.Require().NoError(err, "Should upload test object for read operations")

	suite.T().Log("StorageResourceTestSuite setup complete - MinIO and test app ready")
}

// TearDownSuite runs once after all tests in the suite.
func (suite *StorageResourceTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down StorageResourceTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	if suite.minioContainer != nil {
		suite.minioContainer.Terminate(suite.ctx, &suite.Suite)
	}

	suite.T().Log("StorageResourceTestSuite teardown complete")
}

func (suite *StorageResourceTestSuite) setupTestApp() {
	// Create MinIO config with bucket
	minioConfig := *suite.minioContainer.MinIOConfig

	suite.app, suite.stop = apptest.NewTestApp(
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
		fx.Populate(&suite.service),
	)
}

// Helper methods for making API requests and reading responses

func (suite *StorageResourceTestSuite) makeApiRequest(body api.Request) *http.Response {
	jsonBody, err := encoding.ToJSON(body)
	suite.Require().NoError(err, "Should encode request to JSON")

	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(jsonBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *StorageResourceTestSuite) makeMultipartRequest(params map[string]string, fieldName, fileName string, fileContent []byte) *http.Response {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, value := range params {
		_ = writer.WriteField(key, value)
	}

	if fieldName != "" && fileName != "" {
		part, err := writer.CreateFormFile(fieldName, fileName)
		suite.Require().NoError(err, "Should create form file")
		_, err = part.Write(fileContent)
		suite.Require().NoError(err, "Should write file content")
	}

	err := writer.Close()
	suite.Require().NoError(err, "Should close multipart writer")

	req := httptest.NewRequest(fiber.MethodPost, "/api", body)
	req.Header.Set(fiber.HeaderContentType, writer.FormDataContentType())

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *StorageResourceTestSuite) readBody(resp *http.Response) result.Result {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	suite.Require().NoError(err, "Should read response body")
	res, err := encoding.FromJSON[result.Result](string(body))
	suite.Require().NoError(err, "Should decode response JSON")

	return *res
}

func (suite *StorageResourceTestSuite) readDataAsMap(data any) map[string]any {
	m, ok := data.(map[string]any)
	suite.Require().True(ok, "Data should be a map")

	return m
}

// Test Cases

// TestUpload tests file upload functionality.
func (suite *StorageResourceTestSuite) TestUpload() {
	suite.T().Log("Testing file upload functionality")

	suite.Run("Success", func() {
		uploadData := []byte("Uploaded via Api")

		params := map[string]string{
			"resource": "sys/storage",
			"action":   "upload",
			"version":  "v1",
		}

		resp := suite.makeMultipartRequest(params, "file", "test.txt", uploadData)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Upload should succeed")
		suite.Equal(i18n.T(result.OkMessage), body.Message, "Should return success message")

		data := suite.readDataAsMap(body.Data)
		suite.Equal(suite.testBucketName, data["bucket"], "Bucket should match test bucket")
		suite.NotEmpty(data["key"], "Key should not be empty")
		suite.Contains(data["key"], ".txt", "Key should preserve file extension")
		suite.NotEmpty(data["eTag"], "ETag should not be empty")
		suite.NotZero(data["size"], "Size should not be zero")

		key := data["key"].(string)
		parts := strings.Split(key, "/")
		suite.GreaterOrEqual(len(parts), 4, "Key should have date-based path structure")
		suite.True(strings.HasSuffix(key, ".txt"), "Key should end with .txt")

		reader, err := suite.service.GetObject(suite.ctx, storage.GetObjectOptions{
			Key: key,
		})
		suite.Require().NoError(err, "Should retrieve uploaded file")

		defer reader.Close()

		content, err := io.ReadAll(reader)
		suite.Require().NoError(err, "Should read file content")
		suite.Equal(uploadData, content, "File content should match uploaded data")

		info, err := suite.service.StatObject(suite.ctx, storage.StatObjectOptions{
			Key: key,
		})
		suite.Require().NoError(err, "Should get file metadata")
		suite.NotNil(info.Metadata, "Metadata should not be nil")
		suite.Equal("test.txt", info.Metadata[storage.MetadataKeyOriginalFilename], "Original filename should be preserved in metadata")
	})

	suite.Run("MissingFile", func() {
		params := map[string]string{
			"resource": "sys/storage",
			"action":   "upload",
			"version":  "v1",
		}

		resp := suite.makeMultipartRequest(params, "", "", nil)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Upload should fail without file")
	})

	suite.Run("WithJSON", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/storage",
				Action:   "upload",
				Version:  "v1",
			},
			Params: map[string]any{
				"key": "test.txt",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Upload should fail with JSON request")
	})
}

// TestGetPresignedURL tests presigned URL generation for various scenarios.
func (suite *StorageResourceTestSuite) TestGetPresignedURL() {
	suite.T().Log("Testing presigned URL generation")

	suite.Run("ForDownload", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/storage",
				Action:   "get_presigned_url",
				Version:  "v1",
			},
			Params: map[string]any{
				"key":     suite.testObjectKey,
				"expires": 3600,
				"method":  "GET",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Presigned URL generation should succeed")

		data := suite.readDataAsMap(body.Data)
		url, ok := data["url"].(string)
		suite.True(ok, "URL should be a string")
		suite.NotEmpty(url, "URL should not be empty")
		suite.Contains(url, suite.testBucketName, "URL should contain bucket name")
		suite.Contains(url, suite.testObjectKey, "URL should contain object key")

		downloadReq, err := http.NewRequestWithContext(suite.ctx, http.MethodGet, url, nil)
		suite.Require().NoError(err, "Should create download request")

		downloadResp, err := http.DefaultClient.Do(downloadReq)
		suite.Require().NoError(err, "Should execute download request")

		defer downloadResp.Body.Close()

		suite.Equal(http.StatusOK, downloadResp.StatusCode, "Download should succeed")
		content, err := io.ReadAll(downloadResp.Body)
		suite.Require().NoError(err, "Should read downloaded content")
		suite.Equal(suite.testObjectData, content, "Downloaded content should match original data")
	})

	suite.Run("ForUpload", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/storage",
				Action:   "get_presigned_url",
				Version:  "v1",
			},
			Params: map[string]any{
				"key":     "presigned-upload.txt",
				"expires": 3600,
				"method":  "PUT",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Presigned URL generation should succeed")

		data := suite.readDataAsMap(body.Data)
		url, ok := data["url"].(string)
		suite.True(ok, "URL should be a string")
		suite.NotEmpty(url, "URL should not be empty")
		suite.Contains(url, suite.testBucketName, "URL should contain bucket name")

		uploadData := []byte("Uploaded via presigned URL")
		uploadReq, err := http.NewRequestWithContext(suite.ctx, http.MethodPut, url, bytes.NewReader(uploadData))
		suite.Require().NoError(err, "Should create upload request")

		uploadResp, err := http.DefaultClient.Do(uploadReq)
		suite.Require().NoError(err, "Should execute upload request")

		defer uploadResp.Body.Close()

		suite.Equal(http.StatusOK, uploadResp.StatusCode, "Upload should succeed")
	})

	suite.Run("DefaultExpires", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/storage",
				Action:   "get_presigned_url",
				Version:  "v1",
			},
			Params: map[string]any{
				"key": suite.testObjectKey,
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Presigned URL generation should succeed with default expiration")

		data := suite.readDataAsMap(body.Data)
		suite.Contains(data, "url", "Response should contain URL")
		suite.NotEmpty(data["url"], "URL should not be empty")
	})

	suite.Run("CustomExpiration", func() {
		customExpires := 7200

		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/storage",
				Action:   "get_presigned_url",
				Version:  "v1",
			},
			Params: map[string]any{
				"key":     suite.testObjectKey,
				"expires": float64(customExpires),
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Presigned URL generation with custom expiration should succeed")

		data := suite.readDataAsMap(body.Data)
		url := data["url"].(string)
		suite.NotEmpty(url, "URL should not be empty")
		suite.Contains(url, "X-Amz-Expires", "URL should contain expiration parameter")
	})
}

// TestStatObject tests getting object metadata.
func (suite *StorageResourceTestSuite) TestStatObject() {
	suite.T().Log("Testing stat object functionality")

	suite.Run("Success", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/storage",
				Action:   "stat",
				Version:  "v1",
			},
			Params: map[string]any{
				"key": suite.testObjectKey,
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Stat should succeed")

		data := suite.readDataAsMap(body.Data)
		suite.Equal(suite.testBucketName, data["bucket"], "Bucket should match")
		suite.Equal(suite.testObjectKey, data["key"], "Key should match")
		suite.NotEmpty(data["eTag"], "ETag should not be empty")
		suite.NotZero(data["size"], "Size should not be zero")
		suite.Equal(suite.testContentType, data["contentType"], "Content type should match")
		suite.NotZero(data["lastModified"], "Last modified should not be zero")
		suite.Equal("test.txt", suite.readDataAsMap(data["metadata"])[storage.MetadataKeyOriginalFilename], "Original filename should be in metadata")
	})

	suite.Run("NotFound", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/storage",
				Action:   "stat",
				Version:  "v1",
			},
			Params: map[string]any{
				"key": "non-existent-key.txt",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Stat should fail for non-existent object")
	})
}

// TestListObjects tests listing objects with various filters.
func (suite *StorageResourceTestSuite) TestListObjects() {
	suite.T().Log("Testing list objects functionality")

	suite.Run("Success", func() {
		objects := map[string][]byte{
			"folder1/file1.txt": []byte("content1"),
			"folder1/file2.txt": []byte("content2"),
			"folder2/file3.txt": []byte("content3"),
		}

		for key, content := range objects {
			reader := bytes.NewReader(content)
			_, err := suite.service.PutObject(suite.ctx, storage.PutObjectOptions{
				Key:         key,
				Reader:      reader,
				Size:        int64(len(content)),
				ContentType: "text/plain",
			})
			suite.Require().NoError(err, "Should upload test object")
		}

		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/storage",
				Action:   "list",
				Version:  "v1",
			},
			Params: map[string]any{
				"recursive": true,
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "List should succeed")

		dataSlice, ok := body.Data.([]any)
		suite.True(ok, "Data should be a slice")
		suite.GreaterOrEqual(len(dataSlice), 4, "Should have at least 4 objects")
	})

	suite.Run("WithPrefix", func() {
		objects := map[string][]byte{
			"prefix-test/file1.txt": []byte("content1"),
			"prefix-test/file2.txt": []byte("content2"),
			"other/file3.txt":       []byte("content3"),
		}

		for key, content := range objects {
			reader := bytes.NewReader(content)
			_, err := suite.service.PutObject(suite.ctx, storage.PutObjectOptions{
				Key:         key,
				Reader:      reader,
				Size:        int64(len(content)),
				ContentType: "text/plain",
			})
			suite.Require().NoError(err, "Should upload test object")
		}

		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/storage",
				Action:   "list",
				Version:  "v1",
			},
			Params: map[string]any{
				"prefix":    "prefix-test/",
				"recursive": true,
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "List with prefix should succeed")

		dataSlice, ok := body.Data.([]any)
		suite.True(ok, "Data should be a slice")
		suite.GreaterOrEqual(len(dataSlice), 2, "Should have at least 2 objects with prefix")

		for _, item := range dataSlice {
			obj := item.(map[string]any)
			key := obj["key"].(string)
			suite.Contains(key, "prefix-test/", "All keys should contain prefix")
		}
	})

	suite.Run("WithMaxKeys", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/storage",
				Action:   "list",
				Version:  "v1",
			},
			Params: map[string]any{
				"recursive": true,
				"maxKeys":   1,
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "List with maxKeys should succeed")

		dataSlice, ok := body.Data.([]any)
		suite.True(ok, "Data should be a slice")
		suite.Equal(1, len(dataSlice), "MaxKeys should limit results to 1")
	})
}

// TestUploadWithMetadata tests uploading file with custom metadata.
func (suite *StorageResourceTestSuite) TestUploadWithMetadata() {
	suite.T().Log("Testing upload with custom metadata")

	uploadData := []byte("Test with metadata")

	params := map[string]string{
		"resource": "sys/storage",
		"action":   "upload",
		"version":  "v1",
		"params":   `{"metadata":{"author":"test-suite","version":"1.0"}}`,
	}

	resp := suite.makeMultipartRequest(params, "file", "test-metadata.txt", uploadData)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Upload with metadata should succeed")

	data := suite.readDataAsMap(body.Data)
	uploadKey := data["key"].(string)

	info, err := suite.service.StatObject(suite.ctx, storage.StatObjectOptions{
		Key: uploadKey,
	})
	suite.Require().NoError(err, "Should get object metadata")
	suite.NotNil(info.Metadata, "Metadata should not be nil")

	suite.Equal("test-suite", info.Metadata["Author"], "Author metadata should match")
	suite.Equal("1.0", info.Metadata["Version"], "Version metadata should match")
	suite.Equal("test-metadata.txt", info.Metadata[storage.MetadataKeyOriginalFilename], "Original filename should be preserved")
}

// TestUploadWithContentType tests uploading file with custom content type.
func (suite *StorageResourceTestSuite) TestUploadWithContentType() {
	suite.T().Log("Testing upload with custom content type")

	uploadData := []byte(`{"test": "data"}`)

	params := map[string]string{
		"resource": "sys/storage",
		"action":   "upload",
		"version":  "v1",
		"params":   `{"contentType":"application/json"}`,
	}

	resp := suite.makeMultipartRequest(params, "file", "test.json", uploadData)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Upload with content type should succeed")

	data := suite.readDataAsMap(body.Data)
	uploadKey := data["key"].(string)

	info, err := suite.service.StatObject(suite.ctx, storage.StatObjectOptions{
		Key: uploadKey,
	})
	suite.Require().NoError(err, "Should get object metadata")
	suite.Equal("application/json", info.ContentType, "Content type should match")
}

// TestDeleteTemp tests the delete_temp action for removing temporary uploaded files.
// Tests cover successful deletion, rejection of non-temp keys, idempotent deletion, and missing parameter validation.
func (suite *StorageResourceTestSuite) TestDeleteTemp() {
	suite.T().Log("Testing delete temporary file functionality")

	suite.Run("Success", func() {
		uploadData := []byte("Temporary file to delete")
		params := map[string]string{
			"resource": "sys/storage",
			"action":   "upload",
			"version":  "v1",
		}

		uploadResp := suite.makeMultipartRequest(params, "file", "temp.txt", uploadData)
		suite.Equal(200, uploadResp.StatusCode, "Upload should return 200 OK")

		uploadBody := suite.readBody(uploadResp)
		suite.True(uploadBody.IsOk(), "Upload should succeed")

		uploadResult := suite.readDataAsMap(uploadBody.Data)
		tempKey := uploadResult["key"].(string)
		suite.True(strings.HasPrefix(tempKey, "temp/"), "Uploaded key should have temp/ prefix")
		suite.T().Logf("Uploaded temp file: %s", tempKey)

		_, err := suite.service.StatObject(suite.ctx, storage.StatObjectOptions{
			Key: tempKey,
		})
		suite.Require().NoError(err, "Uploaded file should exist")

		deleteResp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/storage",
				Action:   "delete_temp",
				Version:  "v1",
			},
			Params: map[string]any{
				"key": tempKey,
			},
		})

		suite.Equal(200, deleteResp.StatusCode, "Should return 200 OK")

		deleteBody := suite.readBody(deleteResp)
		suite.True(deleteBody.IsOk(), "Delete temp should succeed")
		suite.Equal(i18n.T(result.OkMessage), deleteBody.Message, "Should return success message")
		suite.T().Logf("Deleted temp file: %s", tempKey)

		_, err = suite.service.StatObject(suite.ctx, storage.StatObjectOptions{
			Key: tempKey,
		})
		suite.Error(err, "File should not exist after deletion")
	})

	suite.Run("NonTempKeyRejected", func() {
		nonTempKey := "permanent/file.txt"
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/storage",
				Action:   "delete_temp",
				Version:  "v1",
			},
			Params: map[string]any{
				"key": nonTempKey,
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Delete temp should fail for non-temp key")
		suite.Equal(body.Message, i18n.T("invalid_temp_key"), "Error message should indicate temp file restriction")
		suite.T().Logf("Rejected non-temp key: %s (message: %s)", nonTempKey, body.Message)
	})

	suite.Run("NonExistentFile", func() {
		nonExistentKey := "temp/non-existent-file.txt"
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/storage",
				Action:   "delete_temp",
				Version:  "v1",
			},
			Params: map[string]any{
				"key": nonExistentKey,
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Delete temp should succeed even for non-existent file")
		suite.T().Logf("Idempotent deletion for non-existent key: %s", nonExistentKey)
	})

	suite.Run("MissingKey", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/storage",
				Action:   "delete_temp",
				Version:  "v1",
			},
			Params: map[string]any{},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Delete temp should fail without key")
		suite.T().Logf("Rejected request without key (message: %s)", body.Message)
	})
}

// TestConcurrentUploads tests concurrent file uploads for thread safety.
func (suite *StorageResourceTestSuite) TestConcurrentUploads() {
	suite.T().Log("Testing concurrent uploads")

	numUploads := 5
	done := make(chan bool, numUploads)

	for i := range numUploads {
		go func(index int) {
			defer func() { done <- true }()

			uploadData := fmt.Appendf(nil, "Concurrent upload %d", index)
			params := map[string]string{
				"resource": "sys/storage",
				"action":   "upload",
				"version":  "v1",
			}

			resp := suite.makeMultipartRequest(params, "file", fmt.Sprintf("test%d.txt", index), uploadData)
			suite.Equal(200, resp.StatusCode, "Concurrent upload should return 200 OK")

			body := suite.readBody(resp)
			suite.True(body.IsOk(), "Concurrent upload should succeed")
		}(i)
	}

	timeout := time.After(30 * time.Second)

	for range numUploads {
		select {
		case <-done:
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
