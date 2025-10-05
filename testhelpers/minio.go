package testhelpers

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
)

type MinIOContainer struct {
	container *minio.MinioContainer
	Config    *config.MinIOConfig
}

func (c *MinIOContainer) Terminate(ctx context.Context, suite *suite.Suite) {
	if err := c.container.Terminate(ctx); err != nil {
		suite.T().Logf("Failed to terminate MinIO container: %v", err)
	}
}

func NewMinIOContainer(ctx context.Context, suite *suite.Suite) *MinIOContainer {
	// Start MinIO container
	minioContainer, err := minio.Run(
		ctx,
		MinIOImage,
		minio.WithUsername(TestMinIOAccessKey),
		minio.WithPassword(TestMinIOSecretKey),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/minio/health/live").
				WithPort("9000/tcp").
				WithStartupTimeout(DefaultContainerTimeout),
		),
	)

	suite.Require().NoError(err)
	suite.T().Log("MinIO container started successfully")

	// Get container connection details
	host, err := minioContainer.Host(ctx)
	suite.Require().NoError(err)

	port, err := minioContainer.MappedPort(ctx, "9000")
	suite.Require().NoError(err)

	// Create MinIO config
	minioConfig := &config.MinIOConfig{
		Endpoint:  fmt.Sprintf("%s:%s", host, port.Port()),
		AccessKey: TestMinIOAccessKey,
		SecretKey: TestMinIOSecretKey,
		UseSSL:    false,
		Region:    constants.Empty,
		Bucket:    TestMinioBucket,
	}

	return &MinIOContainer{
		container: minioContainer,
		Config:    minioConfig,
	}
}
