package testhelpers

import (
	"context"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ilxqx/vef-framework-go/config"
)

type RedisContainer struct {
	container *redis.RedisContainer
	RdsConfig *config.RedisConfig
}

func (c *RedisContainer) Terminate(ctx context.Context, s *suite.Suite) {
	if err := c.container.Terminate(ctx); err != nil {
		s.T().Logf("Failed to terminate redis container: %v", err)
	}
}

func NewRedisContainer(ctx context.Context, s *suite.Suite) *RedisContainer {
	container, err := redis.Run(
		ctx,
		RedisImage,
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(DefaultContainerTimeout),
		),
	)
	s.Require().NoError(err)
	s.T().Log("Redis container started successfully")

	host, err := container.Host(ctx)
	s.Require().NoError(err)

	port, err := container.MappedPort(ctx, "6379")
	s.Require().NoError(err)

	return &RedisContainer{
		container: container,
		RdsConfig: &config.RedisConfig{
			Host:     host,
			Port:     uint16(port.Int()),
			Database: 0,
		},
	}
}
