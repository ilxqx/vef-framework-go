package orm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/database"
	"github.com/ilxqx/vef-framework-go/testhelpers"
)

// runAllORMTests executes all ORM test suites on the given database configuration.
func runAllORMTests(t *testing.T, ctx context.Context, dsConfig *config.DatasourceConfig) {
	// Create database connection
	db, err := database.New(dsConfig)
	require.NoError(t, err)

	defer func() {
		// Close the database connection after all tests are completed
		if err := db.Close(); err != nil {
			t.Logf("Error closing database connection for %s: %v", dsConfig.Type, err)
		}

		t.Logf("All ORM tests completed for %s", dsConfig.Type)
	}()

	ormDb := New(db)

	// Create Select Suite
	selectSuite := &SelectTestSuite{
		ORMTestSuite: &ORMTestSuite{
			ctx:    ctx,
			dbType: dsConfig.Type,
			db:     ormDb,
		},
	}

	// Create Condition Suite
	conditionSuite := &ConditionTestSuite{
		ORMTestSuite: &ORMTestSuite{
			ctx:    ctx,
			dbType: dsConfig.Type,
			db:     ormDb,
		},
	}

	// Create Insert Suite
	insertSuite := &InsertTestSuite{
		ORMTestSuite: &ORMTestSuite{
			ctx:    ctx,
			dbType: dsConfig.Type,
			db:     ormDb,
		},
	}

	// Create Update Suite
	updateSuite := &UpdateTestSuite{
		ORMTestSuite: &ORMTestSuite{
			ctx:    ctx,
			dbType: dsConfig.Type,
			db:     ormDb,
		},
	}

	// Create Delete Suite
	deleteSuite := &DeleteTestSuite{
		ORMTestSuite: &ORMTestSuite{
			ctx:    ctx,
			dbType: dsConfig.Type,
			db:     ormDb,
		},
	}

	// Create Merge Suite
	// mergeSuite := &MergeTestSuite{
	// 	ORMTestSuite: &ORMTestSuite{
	// 		ctx:    ctx,
	// 		dbType: dsConfig.Type,
	// 		db:     ormDb,
	// 	},
	// }

	t.Run("TestSelect", func(t *testing.T) {
		suite.Run(t, selectSuite)
	})

	t.Run("TestCondition", func(t *testing.T) {
		suite.Run(t, conditionSuite)
	})

	t.Run("TestInsert", func(t *testing.T) {
		suite.Run(t, insertSuite)
	})

	t.Run("TestUpdate", func(t *testing.T) {
		suite.Run(t, updateSuite)
	})

	t.Run("TestDelete", func(t *testing.T) {
		suite.Run(t, deleteSuite)
	})

	// t.Run("TestMerge", func(t *testing.T) {
	// 	suite.Run(t, mergeSuite)
	// })
}

// TestPostgres runs all ORM tests against PostgreSQL.
func TestPostgres(t *testing.T) {
	ctx := context.Background()

	// Create a dummy suite for container management
	dummySuite := &suite.Suite{}
	dummySuite.SetT(t)

	// Start PostgreSQL container
	postgresContainer := testhelpers.NewPostgresContainer(ctx, dummySuite)
	defer postgresContainer.Terminate(ctx, dummySuite)

	// Run all ORM tests
	runAllORMTests(t, ctx, postgresContainer.DsConfig)
}

// TestMySQL runs all ORM tests against MySQL.
func TestMySQL(t *testing.T) {
	ctx := context.Background()

	// Create a dummy suite for container management
	dummySuite := &suite.Suite{}
	dummySuite.SetT(t)

	// Start MySQL container
	mysqlContainer := testhelpers.NewMySQLContainer(ctx, dummySuite)
	defer mysqlContainer.Terminate(ctx, dummySuite)

	// Run all ORM tests
	runAllORMTests(t, ctx, mysqlContainer.DsConfig)
}

// TestSQLite runs all ORM tests against SQLite (in-memory).
func TestSQLite(t *testing.T) {
	ctx := context.Background()

	// Create SQLite in-memory database config
	dsConfig := &config.DatasourceConfig{
		Type: constants.DbSQLite,
	}

	// Run all ORM tests
	runAllORMTests(t, ctx, dsConfig)
}
