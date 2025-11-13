package database

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/uptrace/bun"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/testhelpers"
)

// DatabaseTestSuite tests database connection and configuration for PostgreSQL, MySQL, and SQLite.
type DatabaseTestSuite struct {
	suite.Suite

	ctx               context.Context
	postgresContainer *testhelpers.PostgresContainer
	mysqlContainer    *testhelpers.MySQLContainer
}

func (suite *DatabaseTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	suite.postgresContainer = testhelpers.NewPostgresContainer(suite.ctx, &suite.Suite)
	suite.mysqlContainer = testhelpers.NewMySQLContainer(suite.ctx, &suite.Suite)
}

func (suite *DatabaseTestSuite) TearDownSuite() {
	if suite.postgresContainer != nil {
		suite.postgresContainer.Terminate(suite.ctx, &suite.Suite)
	}

	if suite.mysqlContainer != nil {
		suite.mysqlContainer.Terminate(suite.ctx, &suite.Suite)
	}
}

// TestSQLiteConnection tests SQLite in-memory database connection and basic operations.
func (suite *DatabaseTestSuite) TestSQLiteConnection() {
	config := &config.DatasourceConfig{
		Type: constants.DbSQLite,
	}

	db, err := New(config)
	suite.Require().NoError(err, "SQLite connection should succeed")
	suite.Require().NotNil(db, "Database instance should not be nil")

	suite.testBasicDbOperations(db, "SQLite")

	suite.Require().NoError(db.Close(), "Database should close without error")
}

// TestSQLiteWithOptions tests SQLite with custom configuration options.
func (suite *DatabaseTestSuite) TestSQLiteWithOptions() {
	config := &config.DatasourceConfig{
		Type: constants.DbSQLite,
	}

	db, err := New(config, DisableQueryHook())
	suite.Require().NoError(err, "SQLite with custom options should succeed")
	suite.Require().NotNil(db, "Database instance should not be nil")

	suite.testBasicDbOperations(db, "SQLite")

	suite.Require().NoError(db.Close(), "Database should close without error")
}

// TestPostgreSQLConnection tests PostgreSQL database connection via Testcontainers.
func (suite *DatabaseTestSuite) TestPostgreSQLConnection() {
	config := suite.postgresContainer.DsConfig

	suite.T().Logf("Testing PostgreSQL connection with config: %+v", config)

	db, err := New(config)
	suite.Require().NoError(err, "PostgreSQL connection should succeed")
	suite.Require().NotNil(db, "Database instance should not be nil")

	suite.testBasicDbOperations(db, "PostgreSQL")

	suite.Require().NoError(db.Close(), "Database should close without error")
}

// TestMySQLConnection tests MySQL database connection via Testcontainers.
func (suite *DatabaseTestSuite) TestMySQLConnection() {
	config := suite.mysqlContainer.DsConfig

	suite.T().Logf("Testing MySQL connection with config: %+v", config)

	db, err := New(config)
	suite.Require().NoError(err, "MySQL connection should succeed")
	suite.Require().NotNil(db, "Database instance should not be nil")

	suite.testBasicDbOperations(db, "MySQL")

	suite.Require().NoError(db.Close(), "Database should close without error")
}

// TestUnsupportedDatabaseType tests error handling for unsupported database types.
func (suite *DatabaseTestSuite) TestUnsupportedDatabaseType() {
	config := &config.DatasourceConfig{
		Type: "unsupported",
	}

	db, err := New(config)
	suite.Error(err, "Should return error for unsupported database type")
	suite.Nil(db, "Database instance should be nil on error")
	suite.Contains(err.Error(), "unsupported database type", "Error message should mention unsupported type")
}

// TestSQLiteInMemoryMode tests SQLite in-memory mode explicitly.
func (suite *DatabaseTestSuite) TestSQLiteInMemoryMode() {
	config := &config.DatasourceConfig{
		Type: constants.DbSQLite,
	}

	db, err := New(config)
	suite.Require().NoError(err, "In-memory SQLite connection should succeed")
	suite.Require().NotNil(db, "Database instance should not be nil")

	suite.testBasicDbOperations(db, "SQLite In-Memory")

	suite.Require().NoError(db.Close(), "Database should close without error")
}

// TestSQLiteFileMode tests SQLite file-based database mode.
func (suite *DatabaseTestSuite) TestSQLiteFileMode() {
	tempFile, err := os.CreateTemp("", "test_file_*.db")
	suite.Require().NoError(err, "Temporary file creation should succeed")

	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			suite.T().Logf("Failed to remove temp file: %v", err)
		}
	}()

	suite.Require().NoError(tempFile.Close(), "Temporary file should close successfully")

	config := &config.DatasourceConfig{
		Type: constants.DbSQLite,
		Path: tempFile.Name(),
	}

	db, err := New(config)
	suite.Require().NoError(err, "File-based SQLite connection should succeed")
	suite.Require().NotNil(db, "Database instance should not be nil")

	suite.testBasicDbOperations(db, "SQLite File")

	suite.Require().NoError(db.Close(), "Database should close without error")
}

// TestMySQLValidation tests MySQL configuration validation for missing required fields.
func (suite *DatabaseTestSuite) TestMySQLValidation() {
	config := &config.DatasourceConfig{
		Type: constants.DbMySQL,
		Host: "localhost",
		Port: 3306,
		User: "root",
	}

	db, err := New(config)
	suite.Error(err, "Should return error when database name is missing")
	suite.Nil(db, "Database instance should be nil on validation error")
	suite.Contains(err.Error(), "database name is required", "Error message should mention missing database name")
}

// TestConnectionPoolConfiguration tests custom connection pool configuration.
func (suite *DatabaseTestSuite) TestConnectionPoolConfiguration() {
	config := &config.DatasourceConfig{
		Type: constants.DbSQLite,
	}

	customPoolConfig := &ConnectionPoolConfig{
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxIdleTime: 1 * time.Minute,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := New(config, WithConnectionPool(customPoolConfig))
	suite.Require().NoError(err, "Connection with custom pool config should succeed")
	suite.Require().NotNil(db, "Database instance should not be nil")

	sqlDb := db.DB
	suite.NotNil(sqlDb, "Underlying SQL DB should not be nil")

	var result int

	err = db.NewSelect().ColumnExpr("1").Scan(suite.ctx, &result)
	suite.Require().NoError(err, "Query should succeed with connection pool")
	suite.Equal(1, result, "Query result should be 1")

	suite.Require().NoError(db.Close(), "Database should close without error")
}

func (suite *DatabaseTestSuite) testBasicDbOperations(db *bun.DB, dbType string) {
	suite.T().Logf("Testing basic operations for %s", dbType)

	var result int

	err := db.NewSelect().ColumnExpr("1 as test").Scan(suite.ctx, &result)
	suite.Require().NoError(err, "Simple query should succeed")
	suite.Equal(1, result, "Query result should be 1")

	var version string
	switch dbType {
	case "SQLite", "SQLite In-Memory", "SQLite File":
		err = db.NewSelect().ColumnExpr("sqlite_version()").Scan(suite.ctx, &version)
	case "PostgreSQL":
		err = db.NewSelect().ColumnExpr("version()").Scan(suite.ctx, &version)
	case "MySQL":
		err = db.NewSelect().ColumnExpr("version()").Scan(suite.ctx, &version)
	}

	suite.Require().NoError(err, "Version query should succeed")
	suite.NotEmpty(version, "Version should not be empty")
	suite.T().Logf("%s version: %s", dbType, version)

	_, err = db.NewCreateTable().
		Model((*TestTable)(nil)).
		IfNotExists().
		Exec(suite.ctx)
	suite.Require().NoError(err, "Table creation should succeed")

	testData := &TestTable{
		Name:  fmt.Sprintf("test_%s", dbType),
		Value: 42,
	}

	_, err = db.NewInsert().
		Model(testData).
		Exec(suite.ctx)
	suite.Require().NoError(err, "Insert should succeed")

	var retrieved TestTable

	err = db.NewSelect().
		Model(&retrieved).
		Where("name = ?", testData.Name).
		Scan(suite.ctx)
	suite.Require().NoError(err, "Select should succeed")
	suite.Equal(testData.Name, retrieved.Name, "Retrieved name should match")
	suite.Equal(testData.Value, retrieved.Value, "Retrieved value should match")

	_, err = db.NewDropTable().
		Model((*TestTable)(nil)).
		IfExists().
		Exec(suite.ctx)
	suite.Require().NoError(err, "Table cleanup should succeed")
}

type TestTable struct {
	ID    int64  `bun:"id,pk,autoincrement"`
	Name  string `bun:"name,notnull"`
	Value int    `bun:"value"`
}

func TestDatabaseSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}
