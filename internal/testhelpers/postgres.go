package testhelpers

import (
	"context"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ilxqx/vef-framework-go/config"
)

type PostgresContainer struct {
	container *postgres.PostgresContainer
	DsConfig  *config.DatasourceConfig
}

func (c *PostgresContainer) Terminate(ctx context.Context, s *suite.Suite) {
	if err := c.container.Terminate(ctx); err != nil {
		s.T().Logf("Failed to terminate postgres container: %v", err)
	}
}

func NewPostgresContainer(ctx context.Context, s *suite.Suite) *PostgresContainer {
	container, err := postgres.Run(
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
	s.Require().NoError(err)
	s.T().Log("PostgreSQL container started successfully")

	host, err := container.Host(ctx)
	s.Require().NoError(err)

	port, err := container.MappedPort(ctx, "5432")
	s.Require().NoError(err)

	return &PostgresContainer{
		container: container,
		DsConfig: &config.DatasourceConfig{
			Type:     "postgres",
			Host:     host,
			Port:     uint16(port.Int()),
			User:     TestUsername,
			Password: TestPassword,
			Database: TestDatabaseName,
		},
	}
}
