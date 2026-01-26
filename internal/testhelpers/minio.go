package testhelpers

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ilxqx/vef-framework-go/config"
)

type MinIOContainer struct {
	container   *minio.MinioContainer
	MinIOConfig *config.MinIOConfig
}

func (c *MinIOContainer) Terminate(ctx context.Context, s *suite.Suite) {
	if err := c.container.Terminate(ctx); err != nil {
		s.T().Logf("Failed to terminate MinIO container: %v", err)
	}
}

func NewMinIOContainer(ctx context.Context, s *suite.Suite) *MinIOContainer {
	container, err := minio.Run(
		ctx,
		MinIOImage,
		minio.WithUsername(TestMinIOAccessKey),
		minio.WithPassword(TestMinIOSecretKey),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForListeningPort("9000/tcp"),
				wait.ForHTTP("/minio/health/live").WithPort("9000/tcp"),
				wait.ForHTTP("/minio/health/ready").WithPort("9000/tcp"),
			).WithDeadline(DefaultContainerTimeout),
		),
	)
	s.Require().NoError(err)
	s.T().Log("MinIO container started successfully")

	host, err := container.Host(ctx)
	s.Require().NoError(err)

	port, err := container.MappedPort(ctx, "9000")
	s.Require().NoError(err)

	return &MinIOContainer{
		container: container,
		MinIOConfig: &config.MinIOConfig{
			Endpoint:  fmt.Sprintf("%s:%s", host, port.Port()),
			AccessKey: TestMinIOAccessKey,
			SecretKey: TestMinIOSecretKey,
			UseSSL:    false,
			Bucket:    TestMinioBucket,
		},
	}
}
