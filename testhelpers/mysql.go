package testhelpers

import (
	"context"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
)

type MySQLContainer struct {
	container *mysql.MySQLContainer
	DsConfig  *config.DatasourceConfig
}

func (c *MySQLContainer) Terminate(ctx context.Context, suite *suite.Suite) {
	if err := c.container.Terminate(ctx); err != nil {
		suite.T().Logf("Failed to terminate mysql container: %v", err)
	}
}

func NewMySQLContainer(ctx context.Context, suite *suite.Suite) *MySQLContainer {
	// Start MySQL container
	mysqlContainer, err := mysql.Run(
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
	suite.Require().NoError(err)
	suite.T().Log("MySQL container started successfully")

	// Get container connection details
	host, err := mysqlContainer.Host(ctx)
	suite.Require().NoError(err)

	port, err := mysqlContainer.MappedPort(ctx, "3306")
	suite.Require().NoError(err)

	// Create database config
	dsConfig := &config.DatasourceConfig{
		Type:     "mysql",
		Host:     host,
		Port:     uint16(port.Int()),
		User:     TestUsername,
		Password: TestPassword,
		Database: TestDatabaseName,
	}

	return &MySQLContainer{
		container: mysqlContainer,
		DsConfig:  dsConfig,
	}
}
