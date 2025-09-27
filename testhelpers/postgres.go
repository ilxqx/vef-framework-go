package testhelpers

import (
	"context"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	container *postgres.PostgresContainer
	DsConfig  *config.DatasourceConfig
}

func (c *PostgresContainer) Terminate(ctx context.Context, suite *suite.Suite) {
	if err := c.container.Terminate(ctx); err != nil {
		suite.T().Logf("Failed to terminate postgres container: %v", err)
	}
}

func NewPostgresContainer(ctx context.Context, suite *suite.Suite) *PostgresContainer {
	// Start PostgreSQL container
	postgresContainer, err := postgres.Run(
		ctx,
		PostgresImage,
		postgres.WithDatabase(TestDatabaseName),
		postgres.WithUsername(TestUsername),
		postgres.WithPassword(TestPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(DefaultContainerTimeout),
		),
	)

	suite.Require().NoError(err)
	suite.T().Log("PostgreSQL container started successfully")

	// Get container connection details
	host, err := postgresContainer.Host(ctx)
	suite.Require().NoError(err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	suite.Require().NoError(err)

	// Create database config
	dsConfig := &config.DatasourceConfig{
		Type:     "postgres",
		Host:     host,
		Port:     uint16(port.Int()),
		User:     TestUsername,
		Password: TestPassword,
		Database: TestDatabaseName,
	}

	return &PostgresContainer{
		container: postgresContainer,
		DsConfig:  dsConfig,
	}
}
