package minio

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	minioPkg "github.com/minio/minio-go/v7"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/storage"
	"github.com/ilxqx/vef-framework-go/testhelpers"
)

// MinIOProviderTestSuite is the test suite for MinIO provider.
type MinIOProviderTestSuite struct {
	suite.Suite

	ctx            context.Context
	minioContainer *testhelpers.MinIOContainer
	provider       storage.Provider
	minioClient    *minioPkg.Client

	testBucketName  string
	testObjectKey   string
	testObjectData  []byte
	testContentType string
}

// SetupSuite runs before all tests in the suite.
func (suite *MinIOProviderTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.testBucketName = testhelpers.TestMinioBucket
	suite.testObjectKey = "test-file.txt"
	suite.testObjectData = []byte("Hello, MinIO Test!")
	suite.testContentType = "text/plain"

	// Start MinIO container
	suite.minioContainer = testhelpers.NewMinIOContainer(suite.ctx, &suite.Suite)

	// Create MinIO provider
	provider, err := NewMinIOProvider(*suite.minioContainer.Config, &config.AppConfig{})
	suite.Require().NoError(err)
	suite.provider = provider

	// Get MinIO client for setup operations
	suite.minioClient = suite.provider.(*MinIOProvider).client

	// Create test bucket
	err = suite.minioClient.MakeBucket(suite.ctx, suite.testBucketName, minioPkg.MakeBucketOptions{})
	suite.Require().NoError(err)
	suite.T().Logf("Created test bucket: %s", suite.testBucketName)
}

// TearDownSuite runs after all tests in the suite.
func (suite *MinIOProviderTestSuite) TearDownSuite() {
	// Clean up test bucket (remove all objects first)
	objectsCh := suite.minioClient.ListObjects(suite.ctx, suite.testBucketName, minioPkg.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectsCh {
		if object.Err != nil {
			continue
		}

		_ = suite.minioClient.RemoveObject(suite.ctx, suite.testBucketName, object.Key, minioPkg.RemoveObjectOptions{})
	}

	// Remove bucket
	_ = suite.minioClient.RemoveBucket(suite.ctx, suite.testBucketName)

	// Terminate container
	if suite.minioContainer != nil {
		suite.minioContainer.Terminate(suite.ctx, &suite.Suite)
	}
}

// SetupTest runs before each test.
func (suite *MinIOProviderTestSuite) SetupTest() {
	// Clean up any objects from previous tests
	objectsCh := suite.minioClient.ListObjects(suite.ctx, suite.testBucketName, minioPkg.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectsCh {
		if object.Err != nil {
			continue
		}

		_ = suite.minioClient.RemoveObject(suite.ctx, suite.testBucketName, object.Key, minioPkg.RemoveObjectOptions{})
	}
}

// Test Cases

func (suite *MinIOProviderTestSuite) TestPutObjectSuccess() {
	reader := bytes.NewReader(suite.testObjectData)

	info, err := suite.provider.PutObject(suite.ctx, storage.PutObjectOptions{
		Key:         suite.testObjectKey,
		Reader:      reader,
		Size:        int64(len(suite.testObjectData)),
		ContentType: suite.testContentType,
		Metadata: map[string]string{
			"author": "test-suite",
		},
	})

	suite.Require().NoError(err)
	suite.NotNil(info)
	suite.Equal(suite.testBucketName, info.Bucket)
	suite.Equal(suite.testObjectKey, info.Key)
	suite.NotEmpty(info.ETag)
	suite.Equal(int64(len(suite.testObjectData)), info.Size)
	suite.Equal(suite.testContentType, info.ContentType)
}

// TestPutObjectInvalidBucket is no longer relevant since bucket is configured at provider level
// Removed as part of bucket configuration refactoring

func (suite *MinIOProviderTestSuite) TestGetObjectSuccess() {
	// First upload an object
	suite.uploadTestObject()

	// Get the object
	reader, err := suite.provider.GetObject(suite.ctx, storage.GetObjectOptions{
		Key: suite.testObjectKey,
	})

	suite.Require().NoError(err)

	suite.NotNil(reader)
	defer reader.Close()

	// Read and verify content
	data, err := io.ReadAll(reader)
	suite.Require().NoError(err)
	suite.Equal(suite.testObjectData, data)
}

func (suite *MinIOProviderTestSuite) TestGetObjectNotFound() {
	reader, err := suite.provider.GetObject(suite.ctx, storage.GetObjectOptions{
		Key: "non-existent-key.txt",
	})

	suite.Error(err)
	suite.Nil(reader)
	suite.Equal(storage.ErrObjectNotFound, err)
}

func (suite *MinIOProviderTestSuite) TestDeleteObjectSuccess() {
	// First upload an object
	suite.uploadTestObject()

	// Delete the object
	err := suite.provider.DeleteObject(suite.ctx, storage.DeleteObjectOptions{
		Key: suite.testObjectKey,
	})

	suite.NoError(err)

	// Verify it's deleted
	_, err = suite.provider.GetObject(suite.ctx, storage.GetObjectOptions{
		Key: suite.testObjectKey,
	})
	suite.Error(err)
}

func (suite *MinIOProviderTestSuite) TestDeleteObjectNotFound() {
	// Deleting a non-existent object should not return an error in MinIO
	err := suite.provider.DeleteObject(suite.ctx, storage.DeleteObjectOptions{
		Key: "non-existent-key.txt",
	})

	suite.NoError(err)
}

func (suite *MinIOProviderTestSuite) TestDeleteObjectsSuccess() {
	// Upload multiple objects
	keys := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, key := range keys {
		suite.uploadObject(key, []byte("test content"))
	}

	// Delete all objects
	err := suite.provider.DeleteObjects(suite.ctx, storage.DeleteObjectsOptions{
		Keys: keys,
	})

	suite.NoError(err)

	// Verify all are deleted
	for _, key := range keys {
		_, err := suite.provider.GetObject(suite.ctx, storage.GetObjectOptions{
			Key: key,
		})
		suite.Error(err)
	}
}

func (suite *MinIOProviderTestSuite) TestListObjectsSuccess() {
	// Upload multiple objects with different prefixes
	objects := map[string][]byte{
		"folder1/file1.txt": []byte("content1"),
		"folder1/file2.txt": []byte("content2"),
		"folder2/file3.txt": []byte("content3"),
		"root.txt":          []byte("root content"),
	}

	for key, data := range objects {
		suite.uploadObject(key, data)
	}

	// List all objects
	suite.Run("ListAll", func() {
		result, err := suite.provider.ListObjects(suite.ctx, storage.ListObjectsOptions{
			Recursive: true,
		})

		suite.NoError(err)
		suite.Len(result, 4)
	})

	// List with prefix
	suite.Run("ListWithPrefix", func() {
		result, err := suite.provider.ListObjects(suite.ctx, storage.ListObjectsOptions{
			Prefix:    "folder1/",
			Recursive: true,
		})

		suite.NoError(err)
		suite.Len(result, 2)

		for _, obj := range result {
			suite.Contains(obj.Key, "folder1/")
		}
	})

	// List with max keys
	suite.Run("ListWithMaxKeys", func() {
		result, err := suite.provider.ListObjects(suite.ctx, storage.ListObjectsOptions{
			Recursive: true,
			MaxKeys:   2,
		})

		suite.NoError(err)
		suite.Equal(2, len(result))
	})
}

func (suite *MinIOProviderTestSuite) TestGetPresignedURLForGet() {
	// Upload an object first
	suite.uploadTestObject()

	// Get presigned URL
	url, err := suite.provider.GetPresignedURL(suite.ctx, storage.PresignedURLOptions{
		Key:     suite.testObjectKey,
		Expires: 1 * time.Hour,
		Method:  http.MethodGet,
	})

	suite.NoError(err)
	suite.NotEmpty(url)
	suite.Contains(url, suite.testBucketName)
	suite.Contains(url, suite.testObjectKey)

	// Verify we can download using the presigned URL
	downloadReq, err := http.NewRequestWithContext(suite.ctx, http.MethodGet, url, nil)
	suite.Require().NoError(err)

	resp, err := http.DefaultClient.Do(downloadReq)
	suite.Require().NoError(err)

	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)
	data, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	suite.Equal(suite.testObjectData, data)
}

func (suite *MinIOProviderTestSuite) TestGetPresignedURLForPut() {
	// Get presigned URL for upload
	url, err := suite.provider.GetPresignedURL(suite.ctx, storage.PresignedURLOptions{
		Key:     "presigned-upload.txt",
		Expires: 1 * time.Hour,
		Method:  http.MethodPut,
	})

	suite.NoError(err)
	suite.NotEmpty(url)
	suite.Contains(url, suite.testBucketName)

	// Upload using the presigned URL
	uploadData := []byte("Uploaded via presigned URL")
	req, err := http.NewRequestWithContext(suite.ctx, http.MethodPut, url, bytes.NewReader(uploadData))
	suite.Require().NoError(err)

	resp, err := http.DefaultClient.Do(req)
	suite.Require().NoError(err)

	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	// Verify the object was uploaded
	reader, err := suite.provider.GetObject(suite.ctx, storage.GetObjectOptions{
		Key: "presigned-upload.txt",
	})
	suite.Require().NoError(err)

	defer reader.Close()

	data, err := io.ReadAll(reader)
	suite.Require().NoError(err)
	suite.Equal(uploadData, data)
}

func (suite *MinIOProviderTestSuite) TestCopyObjectSuccess() {
	// Upload source object
	suite.uploadTestObject()

	// Copy the object
	destKey := "copied-file.txt"
	info, err := suite.provider.CopyObject(suite.ctx, storage.CopyObjectOptions{
		SourceKey: suite.testObjectKey,
		DestKey:   destKey,
	})

	suite.NoError(err)
	suite.NotNil(info)
	suite.Equal(suite.testBucketName, info.Bucket)
	suite.Equal(destKey, info.Key)
	suite.NotEmpty(info.ETag)

	// Verify both source and destination exist
	reader, err := suite.provider.GetObject(suite.ctx, storage.GetObjectOptions{
		Key: destKey,
	})
	suite.Require().NoError(err)

	defer reader.Close()

	data, err := io.ReadAll(reader)
	suite.Require().NoError(err)
	suite.Equal(suite.testObjectData, data)
}

func (suite *MinIOProviderTestSuite) TestCopyObjectNotFound() {
	_, err := suite.provider.CopyObject(suite.ctx, storage.CopyObjectOptions{
		SourceKey: "non-existent.txt",
		DestKey:   "destination.txt",
	})

	suite.Error(err)
	suite.Equal(storage.ErrObjectNotFound, err)
}

func (suite *MinIOProviderTestSuite) TestMoveObjectSuccess() {
	// Upload source object
	suite.uploadTestObject()

	// Move the object
	destKey := "moved-file.txt"
	info, err := suite.provider.MoveObject(suite.ctx, storage.MoveObjectOptions{
		CopyObjectOptions: storage.CopyObjectOptions{
			SourceKey: suite.testObjectKey,
			DestKey:   destKey,
		},
	})

	suite.NoError(err)
	suite.NotNil(info)
	suite.Equal(suite.testBucketName, info.Bucket)
	suite.Equal(destKey, info.Key)

	// Verify destination exists
	reader, err := suite.provider.GetObject(suite.ctx, storage.GetObjectOptions{
		Key: destKey,
	})
	suite.Require().NoError(err)

	defer reader.Close()

	data, err := io.ReadAll(reader)
	suite.Require().NoError(err)
	suite.Equal(suite.testObjectData, data)

	// Verify source is deleted
	_, err = suite.provider.GetObject(suite.ctx, storage.GetObjectOptions{
		Key: suite.testObjectKey,
	})
	suite.Error(err)
	suite.Equal(storage.ErrObjectNotFound, err)
}

func (suite *MinIOProviderTestSuite) TestStatObjectSuccess() {
	// Upload object with metadata
	suite.uploadTestObject()

	// Get object stats
	info, err := suite.provider.StatObject(suite.ctx, storage.StatObjectOptions{
		Key: suite.testObjectKey,
	})

	suite.NoError(err)
	suite.NotNil(info)
	suite.Equal(suite.testBucketName, info.Bucket)
	suite.Equal(suite.testObjectKey, info.Key)
	suite.NotEmpty(info.ETag)
	suite.Equal(int64(len(suite.testObjectData)), info.Size)
	suite.Equal(suite.testContentType, info.ContentType)
	suite.NotZero(info.LastModified)
}

func (suite *MinIOProviderTestSuite) TestStatObjectNotFound() {
	_, err := suite.provider.StatObject(suite.ctx, storage.StatObjectOptions{
		Key: "non-existent.txt",
	})

	suite.Error(err)
	suite.Equal(storage.ErrObjectNotFound, err)
}

func (suite *MinIOProviderTestSuite) TestPromoteObjectSuccess() {
	// Upload a temporary file
	tempKey := storage.TempPrefix + "2025/01/15/test-promote.txt"
	content := []byte("Content to be promoted")
	suite.uploadObject(tempKey, content)

	// Promote the object
	info, err := suite.provider.PromoteObject(suite.ctx, tempKey)
	suite.Require().NoError(err)
	suite.NotNil(info)

	// Verify the permanent key (without temp/ prefix)
	expectedKey := "2025/01/15/test-promote.txt"
	suite.Equal(expectedKey, info.Key)
	suite.Equal(suite.testBucketName, info.Bucket)

	// Verify the temporary file no longer exists
	_, err = suite.provider.StatObject(suite.ctx, storage.StatObjectOptions{Key: tempKey})
	suite.Error(err)
	suite.Equal(storage.ErrObjectNotFound, err)

	// Verify the permanent file exists
	permanentInfo, err := suite.provider.StatObject(suite.ctx, storage.StatObjectOptions{Key: expectedKey})
	suite.NoError(err)
	suite.Equal(expectedKey, permanentInfo.Key)
	suite.Equal(int64(len(content)), permanentInfo.Size)
}

func (suite *MinIOProviderTestSuite) TestPromoteObjectNonTempKey() {
	// Try to promote a non-temp key (should do nothing and return nil)
	normalKey := "normal/file.txt"
	content := []byte("Normal file content")
	suite.uploadObject(normalKey, content)

	// Try to promote (should return nil since it's not a temp key)
	info, err := suite.provider.PromoteObject(suite.ctx, normalKey)
	suite.NoError(err)
	suite.Nil(info, "PromoteObject should return nil for non-temp keys")

	// Verify the original file still exists
	originalInfo, err := suite.provider.StatObject(suite.ctx, storage.StatObjectOptions{Key: normalKey})
	suite.NoError(err)
	suite.Equal(normalKey, originalInfo.Key)
}

func (suite *MinIOProviderTestSuite) TestPromoteObjectNotFound() {
	// Try to promote a non-existent temp file
	tempKey := storage.TempPrefix + "non-existent.txt"

	info, err := suite.provider.PromoteObject(suite.ctx, tempKey)
	suite.Error(err)
	suite.Nil(info)
}

// Helper methods

func (suite *MinIOProviderTestSuite) uploadTestObject() {
	suite.uploadObject(suite.testObjectKey, suite.testObjectData)
}

func (suite *MinIOProviderTestSuite) uploadObject(key string, data []byte) {
	reader := bytes.NewReader(data)
	_, err := suite.provider.PutObject(suite.ctx, storage.PutObjectOptions{
		Key:         key,
		Reader:      reader,
		Size:        int64(len(data)),
		ContentType: suite.testContentType,
	})
	suite.Require().NoError(err)
}

// TestMinIOProviderSuite runs the test suite.
func TestMinIOProviderSuite(t *testing.T) {
	suite.Run(t, new(MinIOProviderTestSuite))
}
