package testhelpers

import (
	"context"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ilxqx/vef-framework-go/config"
)

type MySQLContainer struct {
	container *mysql.MySQLContainer
	DsConfig  *config.DatasourceConfig
}

func (c *MySQLContainer) Terminate(ctx context.Context, s *suite.Suite) {
	if err := c.container.Terminate(ctx); err != nil {
		s.T().Logf("Failed to terminate mysql container: %v", err)
	}
}

func NewMySQLContainer(ctx context.Context, s *suite.Suite) *MySQLContainer {
	container, err := mysql.Run(
		ctx,
		MySQLImage,
		mysql.WithDatabase(TestDatabaseName),
		mysql.WithUsername(TestUsername),
		mysql.WithPassword(TestPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("port: 3306  MySQL Community Server - GPL").
				WithStartupTimeout(DefaultContainerTimeout),
		),
	)
	s.Require().NoError(err)
	s.T().Log("MySQL container started successfully")

	host, err := container.Host(ctx)
	s.Require().NoError(err)

	port, err := container.MappedPort(ctx, "3306")
	s.Require().NoError(err)

	return &MySQLContainer{
		container: container,
		DsConfig: &config.DatasourceConfig{
			Type:     "mysql",
			Host:     host,
			Port:     uint16(port.Int()),
			User:     TestUsername,
			Password: TestPassword,
			Database: TestDatabaseName,
		},
	}
}
