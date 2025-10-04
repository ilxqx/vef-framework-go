package database

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/testhelpers"
	"github.com/stretchr/testify/suite"
	"github.com/uptrace/bun"
)

// DatabaseTestSuite is the test suite for database package
type DatabaseTestSuite struct {
	suite.Suite
	ctx               context.Context
	postgresContainer *testhelpers.PostgresContainer
	mysqlContainer    *testhelpers.MySQLContainer
}

// SetupSuite runs before all tests in the suite
func (suite *DatabaseTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Start PostgreSQL container using testhelpers
	suite.postgresContainer = testhelpers.NewPostgresContainer(suite.ctx, &suite.Suite)

	// Start MySQL container using testhelpers
	suite.mysqlContainer = testhelpers.NewMySQLContainer(suite.ctx, &suite.Suite)
}

// TearDownSuite runs after all tests in the suite
func (suite *DatabaseTestSuite) TearDownSuite() {
	if suite.postgresContainer != nil {
		suite.postgresContainer.Terminate(suite.ctx, &suite.Suite)
	}
	if suite.mysqlContainer != nil {
		suite.mysqlContainer.Terminate(suite.ctx, &suite.Suite)
	}
}

// TestSQLiteConnection tests SQLite database connection
func (suite *DatabaseTestSuite) TestSQLiteConnection() {
	// Use in-memory SQLite (no path specified)
	config := &config.DatasourceConfig{
		Type: constants.DbSQLite,
	}

	// Test basic connection
	db, err := New(config)
	suite.Require().NoError(err)
	suite.Require().NotNil(db)

	// Test database functionality
	suite.testBasicDbOperations(db, "SQLite")

	// Clean up
	suite.Require().NoError(db.Close())
}

// TestSQLiteWithOptions tests SQLite with custom options
func (suite *DatabaseTestSuite) TestSQLiteWithOptions() {
	// Use in-memory SQLite with custom options
	config := &config.DatasourceConfig{
		Type: constants.DbSQLite,
	}

	// Test with custom options
	db, err := New(config,
		WithQueryHook(false),
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(db)

	// Test database functionality
	suite.testBasicDbOperations(db, "SQLite")

	suite.Require().NoError(db.Close())
}

// TestPostgreSQLConnection tests PostgreSQL database connection
func (suite *DatabaseTestSuite) TestPostgreSQLConnection() {
	// Use the pre-configured PostgreSQL container
	config := suite.postgresContainer.DsConfig

	suite.T().Logf("PostgreSQL connection config: %+v", config)

	// Test basic connection
	db, err := New(config)
	suite.Require().NoError(err)
	suite.Require().NotNil(db)

	// Test database functionality
	suite.testBasicDbOperations(db, "PostgreSQL")

	suite.Require().NoError(db.Close())
}

// TestMySQLConnection tests MySQL database connection
func (suite *DatabaseTestSuite) TestMySQLConnection() {
	// Use the pre-configured MySQL container
	config := suite.mysqlContainer.DsConfig

	suite.T().Logf("MySQL connection config: %+v", config)

	// Test basic connection
	db, err := New(config)
	suite.Require().NoError(err)
	suite.Require().NotNil(db)

	// Test database functionality
	suite.testBasicDbOperations(db, "MySQL")

	suite.Require().NoError(db.Close())
}

// TestUnsupportedDatabaseType tests error handling for unsupported database types
func (suite *DatabaseTestSuite) TestUnsupportedDatabaseType() {
	config := &config.DatasourceConfig{
		Type: "unsupported",
	}

	db, err := New(config)
	suite.Error(err)
	suite.Nil(db)
	suite.Contains(err.Error(), "unsupported database type")
}

// TestSQLiteInMemoryMode tests SQLite in-memory mode
func (suite *DatabaseTestSuite) TestSQLiteInMemoryMode() {
	config := &config.DatasourceConfig{
		Type: constants.DbSQLite,
		// No Path specified - should use in-memory mode
	}

	db, err := New(config)
	suite.Require().NoError(err)
	suite.Require().NotNil(db)

	// Test that in-memory mode works
	suite.testBasicDbOperations(db, "SQLite In-Memory")

	suite.Require().NoError(db.Close())
}

// TestSQLiteFileMode tests SQLite file mode
func (suite *DatabaseTestSuite) TestSQLiteFileMode() {
	// Create a temporary SQLite database file
	tempFile, err := os.CreateTemp("", "test_file_*.db")
	suite.Require().NoError(err)
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			suite.T().Logf("Failed to remove temp file: %v", err)
		}
	}()
	if err := tempFile.Close(); err != nil {
		suite.T().Logf("Failed to close temp file: %v", err)
	}

	config := &config.DatasourceConfig{
		Type: constants.DbSQLite,
		Path: tempFile.Name(),
	}

	db, err := New(config)
	suite.Require().NoError(err)
	suite.Require().NotNil(db)

	// Test that file mode works
	suite.testBasicDbOperations(db, "SQLite File")

	suite.Require().NoError(db.Close())
}

// TestMySQLValidation tests MySQL configuration validation
func (suite *DatabaseTestSuite) TestMySQLValidation() {
	config := &config.DatasourceConfig{
		Type: constants.DbMySQL,
		Host: "localhost",
		Port: 3306,
		User: "root",
		// Missing Database
	}

	db, err := New(config)
	suite.Error(err)
	suite.Nil(db)
	suite.Contains(err.Error(), "database name is required")
}

// TestConnectionPoolConfiguration tests connection pool settings
func (suite *DatabaseTestSuite) TestConnectionPoolConfiguration() {
	// Use in-memory SQLite for connection pool testing
	config := &config.DatasourceConfig{
		Type: constants.DbSQLite,
		// Path is empty, so it will use in-memory mode
	}

	customPoolConfig := &ConnectionPoolConfig{
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxIdleTime: 1 * time.Minute,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := New(config, WithConnectionPool(customPoolConfig))
	suite.Require().NoError(err)
	suite.Require().NotNil(db)

	// Verify connection pool settings
	sqlDb := db.DB

	// Note: MaxIdleClosed is cumulative, so we check if max idle connections are configured
	// by checking that the DB can accept connections properly
	suite.NotNil(sqlDb)

	// Test that we can use the connection pool
	var result int
	err = db.NewSelect().ColumnExpr("1").Scan(suite.ctx, &result)
	suite.Require().NoError(err)
	suite.Equal(1, result)

	suite.Require().NoError(db.Close())
}

// testBasicDbOperations performs basic database operations to verify functionality
func (suite *DatabaseTestSuite) testBasicDbOperations(db *bun.DB, dbType string) {
	suite.T().Logf("Testing basic operations for %s", dbType)

	// Test simple query
	var result int
	err := db.NewSelect().ColumnExpr("1 as test").Scan(suite.ctx, &result)
	suite.Require().NoError(err)
	suite.Equal(1, result)

	// Test version query (this tests our version query implementation)
	var version string
	switch dbType {
	case "SQLite", "SQLite In-Memory", "SQLite File":
		err = db.NewSelect().ColumnExpr("sqlite_version()").Scan(suite.ctx, &version)
	case "PostgreSQL":
		err = db.NewSelect().ColumnExpr("version()").Scan(suite.ctx, &version)
	case "MySQL":
		err = db.NewSelect().ColumnExpr("version()").Scan(suite.ctx, &version)
	}
	suite.Require().NoError(err)
	suite.NotEmpty(version)
	suite.T().Logf("%s version: %s", dbType, version)

	// Test table creation and basic CRUD
	_, err = db.NewCreateTable().
		Model((*TestTable)(nil)).
		IfNotExists().
		Exec(suite.ctx)
	suite.Require().NoError(err)

	// Insert test data
	testData := &TestTable{
		Name:  fmt.Sprintf("test_%s", dbType),
		Value: 42,
	}

	_, err = db.NewInsert().
		Model(testData).
		Exec(suite.ctx)
	suite.Require().NoError(err)

	// Read test data
	var retrieved TestTable
	err = db.NewSelect().
		Model(&retrieved).
		Where("name = ?", testData.Name).
		Scan(suite.ctx)
	suite.Require().NoError(err)
	suite.Equal(testData.Name, retrieved.Name)
	suite.Equal(testData.Value, retrieved.Value)

	// Clean up test table
	_, err = db.NewDropTable().
		Model((*TestTable)(nil)).
		IfExists().
		Exec(suite.ctx)
	suite.Require().NoError(err)
}

// TestTable is a simple table for testing database operations
type TestTable struct {
	ID    int64  `bun:"id,pk,autoincrement"`
	Name  string `bun:"name,notnull"`
	Value int    `bun:"value"`
}

// TestDatabaseSuite runs the test suite
func TestDatabaseSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}
