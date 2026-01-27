package schema_test

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/suite"
	"github.com/uptrace/bun"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/apptest"
	"github.com/ilxqx/vef-framework-go/internal/testhelpers"
	"github.com/ilxqx/vef-framework-go/result"
)

// SchemaResourceTestSuite tests the schema API resource functionality.
type SchemaResourceTestSuite struct {
	suite.Suite

	ctx               context.Context
	postgresContainer *testhelpers.PostgresContainer
	mysqlContainer    *testhelpers.MySQLContainer
}

func (suite *SchemaResourceTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	suite.postgresContainer = testhelpers.NewPostgresContainer(suite.ctx, &suite.Suite)
	suite.mysqlContainer = testhelpers.NewMySQLContainer(suite.ctx, &suite.Suite)
}

func (suite *SchemaResourceTestSuite) TearDownSuite() {
	if suite.postgresContainer != nil {
		suite.postgresContainer.Terminate(suite.ctx, &suite.Suite)
	}

	if suite.mysqlContainer != nil {
		suite.mysqlContainer.Terminate(suite.ctx, &suite.Suite)
	}
}

func (suite *SchemaResourceTestSuite) TestPostgresResource() {
	suite.T().Log("Testing Schema Resource for PostgreSQL")
	suite.runResourceTests(suite.postgresContainer.DsConfig, "PostgreSQL")
}

func (suite *SchemaResourceTestSuite) TestMySQLResource() {
	suite.T().Log("Testing Schema Resource for MySQL")
	suite.runResourceTests(suite.mysqlContainer.DsConfig, "MySQL")
}

func (suite *SchemaResourceTestSuite) TestSQLiteResource() {
	suite.T().Log("Testing Schema Resource for SQLite")

	dsConfig := &config.DatasourceConfig{
		Type: constants.SQLite,
	}

	suite.runResourceTests(dsConfig, "SQLite")
}

func (suite *SchemaResourceTestSuite) runResourceTests(dsConfig *config.DatasourceConfig, dbType string) {
	var (
		bunDB   *bun.DB
		testApp *app.App
	)

	testApp, stop := apptest.NewTestApp(
		suite.T(),
		fx.Replace(dsConfig),
		fx.Populate(&bunDB),
	)

	defer stop()

	suite.setupTestTables(bunDB.DB, dsConfig.Type)
	defer suite.cleanupTestTables(bunDB.DB, dsConfig.Type)

	suite.Run("ListTables", func() {
		resp := suite.makeApiRequest(testApp, api.Request{
			Identifier: api.Identifier{
				Resource: "sys/schema",
				Action:   "list_tables",
				Version:  "v1",
			},
		})

		suite.Equal(http.StatusOK, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "list_tables should succeed")

		tables, ok := body.Data.([]any)
		suite.True(ok, "Data should be an array")

		tableNames := make([]string, 0, len(tables))
		for _, t := range tables {
			tableMap, ok := t.(map[string]any)
			if ok {
				if name, exists := tableMap["name"]; exists {
					tableNames = append(tableNames, name.(string))
				}
			}
		}

		suite.T().Logf("%s tables found via API: %v", dbType, tableNames)
		suite.Contains(tableNames, "resource_test_orders", "Should find resource_test_orders table")
		suite.Contains(tableNames, "resource_test_items", "Should find resource_test_items table")
	})

	suite.Run("GetTableSchemaSuccess", func() {
		resp := suite.makeApiRequest(testApp, api.Request{
			Identifier: api.Identifier{
				Resource: "sys/schema",
				Action:   "get_table_schema",
				Version:  "v1",
			},
			Params: map[string]any{
				"name": "resource_test_orders",
			},
		})

		suite.Equal(http.StatusOK, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "get_table_schema should succeed")

		tableSchema, ok := body.Data.(map[string]any)
		suite.True(ok, "Data should be a map")

		suite.Equal("resource_test_orders", tableSchema["name"], "Table name should match")

		columns, ok := tableSchema["columns"].([]any)
		suite.True(ok, "Columns should be an array")
		suite.NotEmpty(columns, "Columns should not be empty")

		columnNames := make([]string, 0, len(columns))
		for _, col := range columns {
			colMap, ok := col.(map[string]any)
			if ok {
				if name, exists := colMap["name"]; exists {
					columnNames = append(columnNames, name.(string))
				}
			}
		}

		suite.T().Logf("%s resource_test_orders columns via API: %v", dbType, columnNames)
		suite.Contains(columnNames, "id", "Should have id column")
		suite.Contains(columnNames, "customer_name", "Should have customer_name column")
		suite.Contains(columnNames, "total_amount", "Should have total_amount column")
	})

	suite.Run("GetTableSchemaWithPrimaryKey", func() {
		resp := suite.makeApiRequest(testApp, api.Request{
			Identifier: api.Identifier{
				Resource: "sys/schema",
				Action:   "get_table_schema",
				Version:  "v1",
			},
			Params: map[string]any{
				"name": "resource_test_orders",
			},
		})

		body := suite.readBody(resp)
		tableSchema := body.Data.(map[string]any)

		pk, hasPK := tableSchema["primaryKey"]
		suite.True(hasPK, "Should have primaryKey")
		suite.NotNil(pk, "PrimaryKey should not be nil")

		pkMap, ok := pk.(map[string]any)
		suite.True(ok, "PrimaryKey should be a map")

		pkColumns, ok := pkMap["columns"].([]any)
		suite.True(ok, "PrimaryKey columns should be an array")
		suite.NotEmpty(pkColumns, "PrimaryKey columns should not be empty")

		suite.T().Logf("%s resource_test_orders primary key via API: %v", dbType, pkColumns)
	})

	suite.Run("GetTableSchemaNotFound", func() {
		resp := suite.makeApiRequest(testApp, api.Request{
			Identifier: api.Identifier{
				Resource: "sys/schema",
				Action:   "get_table_schema",
				Version:  "v1",
			},
			Params: map[string]any{
				"name": "nonexistent_table_xyz",
			},
		})

		suite.Equal(http.StatusOK, resp.StatusCode, "Should return 200 OK (error in body)")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "get_table_schema should fail for nonexistent table")
		suite.Equal(result.ErrCodeSchemaTableNotFound, body.Code, "Error code should be ErrCodeSchemaTableNotFound")

		suite.T().Logf("%s table not found error: code=%d, message=%s", dbType, body.Code, body.Message)
	})

	suite.Run("GetTableSchemaValidationError", func() {
		resp := suite.makeApiRequest(testApp, api.Request{
			Identifier: api.Identifier{
				Resource: "sys/schema",
				Action:   "get_table_schema",
				Version:  "v1",
			},
			Params: map[string]any{
				// Missing required "name" parameter
			},
		})

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "get_table_schema should fail without name parameter")

		suite.T().Logf("%s validation error: code=%d, message=%s", dbType, body.Code, body.Message)
	})
}

func (suite *SchemaResourceTestSuite) makeApiRequest(testApp *app.App, body api.Request) *http.Response {
	jsonBody, err := encoding.ToJSON(body)
	suite.Require().NoError(err, "Should encode request to JSON")

	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(jsonBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := testApp.Test(req)
	suite.Require().NoError(err, "API request should not fail")

	return resp
}

func (suite *SchemaResourceTestSuite) readBody(resp *http.Response) result.Result {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	suite.Require().NoError(err, "Should read response body")

	res, err := encoding.FromJSON[result.Result](string(body))
	suite.Require().NoError(err, "Should decode response JSON")

	return *res
}

func (suite *SchemaResourceTestSuite) setupTestTables(db *sql.DB, dbType constants.DBType) {
	var ordersSQL, itemsSQL string

	switch dbType {
	case constants.Postgres:
		ordersSQL = `
			CREATE TABLE IF NOT EXISTS resource_test_orders (
				id SERIAL PRIMARY KEY,
				customer_name VARCHAR(100) NOT NULL,
				total_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
				status VARCHAR(20) DEFAULT 'pending',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`
		itemsSQL = `
			CREATE TABLE IF NOT EXISTS resource_test_items (
				id SERIAL PRIMARY KEY,
				order_id INTEGER NOT NULL REFERENCES resource_test_orders(id) ON DELETE CASCADE,
				product_name VARCHAR(200) NOT NULL,
				quantity INTEGER NOT NULL DEFAULT 1,
				unit_price DECIMAL(10, 2) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`

	case constants.MySQL:
		ordersSQL = `
			CREATE TABLE IF NOT EXISTS resource_test_orders (
				id INT AUTO_INCREMENT PRIMARY KEY,
				customer_name VARCHAR(100) NOT NULL,
				total_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
				status VARCHAR(20) DEFAULT 'pending',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`
		itemsSQL = `
			CREATE TABLE IF NOT EXISTS resource_test_items (
				id INT AUTO_INCREMENT PRIMARY KEY,
				order_id INT NOT NULL,
				product_name VARCHAR(200) NOT NULL,
				quantity INT NOT NULL DEFAULT 1,
				unit_price DECIMAL(10, 2) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				CONSTRAINT fk_items_order FOREIGN KEY (order_id) REFERENCES resource_test_orders(id) ON DELETE CASCADE,
				INDEX idx_items_order (order_id)
			)`

	case constants.SQLite:
		ordersSQL = `
			CREATE TABLE IF NOT EXISTS resource_test_orders (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				customer_name TEXT NOT NULL,
				total_amount REAL NOT NULL DEFAULT 0,
				status TEXT DEFAULT 'pending',
				created_at TEXT DEFAULT CURRENT_TIMESTAMP
			)`
		itemsSQL = `
			CREATE TABLE IF NOT EXISTS resource_test_items (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				order_id INTEGER NOT NULL REFERENCES resource_test_orders(id) ON DELETE CASCADE,
				product_name TEXT NOT NULL,
				quantity INTEGER NOT NULL DEFAULT 1,
				unit_price REAL NOT NULL,
				created_at TEXT DEFAULT CURRENT_TIMESTAMP
			)`
	}

	_, err := db.ExecContext(suite.ctx, ordersSQL)
	suite.Require().NoError(err, "Creating resource_test_orders table should succeed")

	_, err = db.ExecContext(suite.ctx, itemsSQL)
	suite.Require().NoError(err, "Creating resource_test_items table should succeed")
}

func (suite *SchemaResourceTestSuite) cleanupTestTables(db *sql.DB, _ constants.DBType) {
	_, _ = db.ExecContext(suite.ctx, "DROP TABLE IF EXISTS resource_test_items")
	_, _ = db.ExecContext(suite.ctx, "DROP TABLE IF EXISTS resource_test_orders")
}

func TestSchemaResourceSuite(t *testing.T) {
	suite.Run(t, new(SchemaResourceTestSuite))
}
