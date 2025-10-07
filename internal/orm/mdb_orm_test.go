package orm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/database"
	"github.com/ilxqx/vef-framework-go/testhelpers"
)

// MultiDatabaseORMTestSuite manages multiple database containers and runs ORMTestSuite against each.
// This is the top-level test suite that orchestrates testing across PostgreSQL, MySQL, and SQLite.
type MultiDatabaseORMTestSuite struct {
	suite.Suite

	ctx               context.Context
	postgresContainer *testhelpers.PostgresContainer
	mysqlContainer    *testhelpers.MySQLContainer
}

// SetupSuite initializes database containers.
func (suite *MultiDatabaseORMTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Start PostgreSQL container
	suite.postgresContainer = testhelpers.NewPostgresContainer(suite.ctx, &suite.Suite)

	// Start MySQL container
	suite.mysqlContainer = testhelpers.NewMySQLContainer(suite.ctx, &suite.Suite)
}

// TearDownSuite cleans up database containers.
func (suite *MultiDatabaseORMTestSuite) TearDownSuite() {
	if suite.postgresContainer != nil {
		suite.postgresContainer.Terminate(suite.ctx, &suite.Suite)
	}

	if suite.mysqlContainer != nil {
		suite.mysqlContainer.Terminate(suite.ctx, &suite.Suite)
	}
}

// TestPostgre runs all ORM tests against PostgreSQL.
func (suite *MultiDatabaseORMTestSuite) TestPostgre() {
	suite.runORMTests(suite.postgresContainer.DsConfig)
}

// TestMySQL runs all ORM tests against MySQL.
func (suite *MultiDatabaseORMTestSuite) TestMySQL() {
	suite.runORMTests(suite.mysqlContainer.DsConfig)
}

// TestSQLite runs all ORM tests against SQLite (in-memory).
func (suite *MultiDatabaseORMTestSuite) TestSQLite() {
	// Create SQLite in-memory database config
	dsConfig := &config.DatasourceConfig{
		Type: constants.DbSQLite,
	}

	suite.runORMTests(dsConfig)
}

// runORMTests executes all ORM test methods on the given suite.
func (st *MultiDatabaseORMTestSuite) runORMTests(dsConfig *config.DatasourceConfig) {
	// Create database connection
	db, err := database.CreateDb(dsConfig)
	st.Require().NoError(err)

	defer func() {
		// Close the database connection after all tests are completed
		if err := db.Close(); err != nil {
			st.T().Logf("Error closing database connection for %s: %v", dsConfig.Type, err)
		}

		st.T().Logf("All ORM tests completed for %s", dsConfig.Type)
	}()

	// Create Select Suite
	selectSuite := &SelectTestSuite{
		ORMTestSuite: &ORMTestSuite{
			ctx:    st.ctx,
			dbType: dsConfig.Type,
			db:     New(db),
		},
	}

	// Create Condition Suite
	conditionSuite := &ConditionTestSuite{
		ORMTestSuite: &ORMTestSuite{
			ctx:    st.ctx,
			dbType: dsConfig.Type,
			db:     New(db),
		},
	}

	// Create Insert Suite
	insertSuite := &InsertTestSuite{
		ORMTestSuite: &ORMTestSuite{
			ctx:    st.ctx,
			dbType: dsConfig.Type,
			db:     New(db),
		},
	}

	// Create Update Suite
	updateSuite := &UpdateTestSuite{
		ORMTestSuite: &ORMTestSuite{
			ctx:    st.ctx,
			dbType: dsConfig.Type,
			db:     New(db),
		},
	}

	// Create Delete Suite
	deleteSuite := &DeleteTestSuite{
		ORMTestSuite: &ORMTestSuite{
			ctx:    st.ctx,
			dbType: dsConfig.Type,
			db:     New(db),
		},
	}

	// Create Merge Suite
	// mergeSuite := &MergeTestSuite{
	// 	ORMTestSuite: &ORMTestSuite{
	// 		ctx:    st.ctx,
	// 		dbType: dsConfig.Type,
	// 		db:     New(db),
	// 	},
	// }

	st.Run("TestSelect", func() {
		suite.Run(st.T(), selectSuite)
	})

	st.Run("TestCondition", func() {
		suite.Run(st.T(), conditionSuite)
	})

	st.Run("TestInsert", func() {
		suite.Run(st.T(), insertSuite)
	})

	st.Run("TestUpdate", func() {
		suite.Run(st.T(), updateSuite)
	})

	st.Run("TestDelete", func() {
		suite.Run(st.T(), deleteSuite)
	})

	// st.Run("TestMerge", func() {
	// 	suite.Run(st.T(), mergeSuite)
	// })
}

// TestMultiDatabaseORM runs the complete ORM test suite against PostgreSQL, MySQL, and SQLite.
// This is the main entry point for testing the ORM's cross-database compatibility.
func TestMultiDatabaseORM(t *testing.T) {
	suite.Run(t, new(MultiDatabaseORMTestSuite))
}
