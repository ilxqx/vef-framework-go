package test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/database"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/testhelpers"
)

// MultiDatabaseAPIsTestSuite manages multiple database containers and runs API test suites against each.
// This is the top-level test suite that orchestrates testing across PostgreSQL, MySQL, and SQLite.
type MultiDatabaseAPIsTestSuite struct {
	suite.Suite

	ctx               context.Context
	postgresContainer *testhelpers.PostgresContainer
	mysqlContainer    *testhelpers.MySQLContainer
}

// SetupSuite initializes database containers.
func (suite *MultiDatabaseAPIsTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Start PostgreSQL container
	suite.postgresContainer = testhelpers.NewPostgresContainer(suite.ctx, &suite.Suite)

	// Start MySQL container
	suite.mysqlContainer = testhelpers.NewMySQLContainer(suite.ctx, &suite.Suite)
}

// TearDownSuite cleans up database containers.
func (suite *MultiDatabaseAPIsTestSuite) TearDownSuite() {
	if suite.postgresContainer != nil {
		suite.postgresContainer.Terminate(suite.ctx, &suite.Suite)
	}

	if suite.mysqlContainer != nil {
		suite.mysqlContainer.Terminate(suite.ctx, &suite.Suite)
	}
}

// TestPostgres runs all API tests against PostgreSQL.
func (suite *MultiDatabaseAPIsTestSuite) TestPostgres() {
	suite.runAPITests(suite.postgresContainer.DsConfig)
}

// TestMySQL runs all API tests against MySQL.
func (suite *MultiDatabaseAPIsTestSuite) TestMySQL() {
	suite.runAPITests(suite.mysqlContainer.DsConfig)
}

// TestSQLite runs all API tests against SQLite (in-memory).
func (suite *MultiDatabaseAPIsTestSuite) TestSQLite() {
	// Create SQLite in-memory database config
	dsConfig := &config.DatasourceConfig{
		Type: constants.DbSQLite,
	}

	suite.runAPITests(dsConfig)
}

// runAPITests executes all API test suites on the given database configuration.
func (st *MultiDatabaseAPIsTestSuite) runAPITests(dsConfig *config.DatasourceConfig) {
	// Create database connection
	db, err := database.New(dsConfig)
	st.Require().NoError(err)

	defer func() {
		// Close the database connection after all tests are completed
		if err := db.Close(); err != nil {
			st.T().Logf("Error closing database connection for %s: %v", dsConfig.Type, err)
		}

		st.T().Logf("All API tests completed for %s", dsConfig.Type)
	}()

	// Setup test data using fixtures
	st.setupTestFixtures(db, dsConfig.Type)

	ormDb := orm.New(db)

	// Create FindAll Suite
	findAllSuite := &FindAllTestSuite{
		BaseSuite{
			ctx:    st.ctx,
			db:     ormDb,
			dbType: dsConfig.Type,
		},
	}

	// Create FindPage Suite
	findPageSuite := &FindPageTestSuite{
		BaseSuite{
			ctx:    st.ctx,
			db:     ormDb,
			dbType: dsConfig.Type,
		},
	}

	// Create FindOne Suite
	findOneSuite := &FindOneTestSuite{
		BaseSuite{
			ctx:    st.ctx,
			db:     ormDb,
			dbType: dsConfig.Type,
		},
	}

	// Create FindOptions Suite
	findOptionsSuite := &FindOptionsTestSuite{
		BaseSuite{
			ctx:    st.ctx,
			db:     ormDb,
			dbType: dsConfig.Type,
		},
	}

	// Create FindTree Suite
	findTreeSuite := &FindTreeTestSuite{
		BaseSuite{
			ctx:    st.ctx,
			db:     ormDb,
			dbType: dsConfig.Type,
		},
	}

	// Create FindTreeOptions Suite
	findTreeOptionsSuite := &FindTreeOptionsTestSuite{
		BaseSuite{
			ctx:    st.ctx,
			db:     ormDb,
			dbType: dsConfig.Type,
		},
	}

	// Create Suite
	createSuite := &CreateTestSuite{
		BaseSuite{
			ctx:    st.ctx,
			db:     ormDb,
			dbType: dsConfig.Type,
		},
	}
	createManySuite := &CreateManyTestSuite{
		BaseSuite{
			ctx:    st.ctx,
			db:     ormDb,
			dbType: dsConfig.Type,
		},
	}

	// Create Update Suite
	updateSuite := &UpdateTestSuite{
		BaseSuite{
			ctx:    st.ctx,
			db:     ormDb,
			dbType: dsConfig.Type,
		},
	}
	updateManySuite := &UpdateManyTestSuite{
		BaseSuite{
			ctx:    st.ctx,
			db:     ormDb,
			dbType: dsConfig.Type,
		},
	}

	// Create Delete Suite
	deleteSuite := &DeleteTestSuite{
		BaseSuite{
			ctx:    st.ctx,
			db:     ormDb,
			dbType: dsConfig.Type,
		},
	}
	deleteManySuite := &DeleteManyTestSuite{
		BaseSuite{
			ctx:    st.ctx,
			db:     ormDb,
			dbType: dsConfig.Type,
		},
	}

	// Create Export Suite
	exportSuite := &ExportTestSuite{
		BaseSuite{
			ctx:    st.ctx,
			db:     ormDb,
			dbType: dsConfig.Type,
		},
	}

	// Create Import Suite
	importSuite := &ImportTestSuite{
		BaseSuite{
			ctx:    st.ctx,
			db:     ormDb,
			dbType: dsConfig.Type,
		},
	}

	st.Run("TestFindAll", func() {
		suite.Run(st.T(), findAllSuite)
	})

	st.Run("TestFindPage", func() {
		suite.Run(st.T(), findPageSuite)
	})

	st.Run("TestFindOne", func() {
		suite.Run(st.T(), findOneSuite)
	})

	st.Run("TestFindOptions", func() {
		suite.Run(st.T(), findOptionsSuite)
	})

	// TODO: SQLite doesn't support recursive CTE with parentheses generated by Bun framework
	// This is a known issue with Bun's SQL generation. Skip these tests for SQLite until Bun fixes it.
	// See: https://github.com/uptrace/bun/issues/xxx (placeholder for issue link)
	if dsConfig.Type != constants.DbSQLite {
		st.Run("TestFindTree", func() {
			suite.Run(st.T(), findTreeSuite)
		})

		st.Run("TestFindTreeOptions", func() {
			suite.Run(st.T(), findTreeOptionsSuite)
		})
	} else {
		st.T().Logf("Skipping FindTree and FindTreeOptions tests for SQLite due to Bun recursive CTE syntax issue")
	}

	st.Run("TestCreate", func() {
		suite.Run(st.T(), createSuite)
	})

	st.Run("TestCreateMany", func() {
		suite.Run(st.T(), createManySuite)
	})

	st.Run("TestUpdate", func() {
		suite.Run(st.T(), updateSuite)
	})

	st.Run("TestUpdateMany", func() {
		suite.Run(st.T(), updateManySuite)
	})

	st.Run("TestDelete", func() {
		suite.Run(st.T(), deleteSuite)
	})

	st.Run("TestDeleteMany", func() {
		suite.Run(st.T(), deleteManySuite)
	})

	st.Run("TestExport", func() {
		suite.Run(st.T(), exportSuite)
	})

	st.Run("TestImport", func() {
		suite.Run(st.T(), importSuite)
	})
}

// setupTestFixtures loads test data from fixture files using dbfixture.
func (st *MultiDatabaseAPIsTestSuite) setupTestFixtures(db bun.IDB, dbType constants.DbType) {
	st.T().Logf("Setting up test fixtures for %s", dbType)

	bunDb, ok := db.(*bun.DB)
	if !ok {
		st.Require().Fail("Could not convert to *bun.DB")
	}

	// Register models
	bunDb.RegisterModel(
		(*TestUser)(nil),
		(*TestCategory)(nil),
		(*TestCompositePKItem)(nil),
		(*ExportUser)(nil),
		(*ImportUser)(nil),
	)

	// Create fixture loader with template functions
	fixture := dbfixture.New(
		bunDb,
		dbfixture.WithRecreateTables(),
	)

	// Load fixtures from testdata directory
	err := fixture.Load(st.ctx, os.DirFS("testdata"), "fixture.yaml")
	st.Require().NoError(err, "Failed to load fixtures for %s", dbType)

	st.T().Logf("Test fixtures loaded for %s database", dbType)
}

// TestMultiDatabaseAPIs runs the complete API test suite against PostgreSQL, MySQL, and SQLite.
// This is the main entry point for testing the APIs' cross-database compatibility.
func TestMultiDatabaseAPIs(t *testing.T) {
	suite.Run(t, new(MultiDatabaseAPIsTestSuite))
}
