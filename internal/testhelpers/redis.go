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

func (c *RedisContainer) Terminate(ctx context.Context, suite *suite.Suite) {
	if err := c.container.Terminate(ctx); err != nil {
		suite.T().Logf("Failed to terminate redis container: %v", err)
	}
}

func NewRedisContainer(ctx context.Context, suite *suite.Suite) *RedisContainer {
	// Start Redis container
	redisContainer, err := redis.Run(
		ctx,
		RedisImage,
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(DefaultContainerTimeout),
		),
	)
	suite.Require().NoError(err)
	suite.T().Log("Redis container started successfully")

	host, err := redisContainer.Host(ctx)
	suite.Require().NoError(err)

	port, err := redisContainer.MappedPort(ctx, "6379")
	suite.Require().NoError(err)

	rdsConfig := &config.RedisConfig{
		Host:     host,
		Port:     uint16(port.Int()),
		Database: 0,
	}

	return &RedisContainer{
		container: redisContainer,
		RdsConfig: rdsConfig,
	}
}
