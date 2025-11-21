package orm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/database"
	"github.com/ilxqx/vef-framework-go/internal/testhelpers"
)

// runAllOrmTests executes all Orm test suites on the given database configuration.
func runAllOrmTests(t *testing.T, ctx context.Context, dsConfig *config.DatasourceConfig) {
	// Create database connection
	db, err := database.New(dsConfig)
	require.NoError(t, err)

	defer func() {
		// Close the database connection after all tests are completed
		if err := db.Close(); err != nil {
			t.Logf("Error closing database connection for %s: %v", dsConfig.Type, err)
		}
	}()

	ormDb := New(db)

	// Create Select Suite
	selectSuite := &SelectTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dsConfig.Type,
			db:     ormDb,
		},
	}

	// Create Insert Suite
	insertSuite := &InsertTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dsConfig.Type,
			db:     ormDb,
		},
	}

	// Create Update Suite
	updateSuite := &UpdateTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dsConfig.Type,
			db:     ormDb,
		},
	}

	// Create Delete Suite
	deleteSuite := &DeleteTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dsConfig.Type,
			db:     ormDb,
		},
	}

	// Create Merge Suite
	mergeSuite := &MergeTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dsConfig.Type,
			db:     ormDb,
		},
	}

	t.Run("TestSelect", func(t *testing.T) {
		suite.Run(t, selectSuite)
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

	t.Run("TestMerge", func(t *testing.T) {
		suite.Run(t, mergeSuite)
	})

	t.Run("TestConditionBuilder", func(t *testing.T) {
		runAllConditionBuilderTests(t, ctx, dsConfig.Type, ormDb)
	})

	t.Run("TestExprBuilder", func(t *testing.T) {
		runAllExprBuilderTests(t, ctx, dsConfig.Type, ormDb)
	})
}

// runAllConditionBuilderTests executes all ConditionBuilder test suites on the given database.
func runAllConditionBuilderTests(t *testing.T, ctx context.Context, dbType constants.DbType, db Db) {
	// Create base suite configuration
	baseSuite := &ConditionBuilderTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dbType,
			db:     db,
		},
	}

	// Create all test suites
	basicComparisonSuite := &BasicComparisonTestSuite{
		ConditionBuilderTestSuite: baseSuite,
	}

	rangeSetOperationsSuite := &RangeSetOperationsTestSuite{
		ConditionBuilderTestSuite: baseSuite,
	}

	subqueryOperationsSuite := &SubqueryOperationsTestSuite{
		ConditionBuilderTestSuite: baseSuite,
	}

	expressionOperationsSuite := &ExpressionOperationsTestSuite{
		ConditionBuilderTestSuite: baseSuite,
	}

	nullBooleanChecksSuite := &NullBooleanChecksTestSuite{
		ConditionBuilderTestSuite: baseSuite,
	}

	stringOperationsSuite := &StringOperationsTestSuite{
		ConditionBuilderTestSuite: baseSuite,
	}

	auditConditionsSuite := &AuditConditionsTestSuite{
		ConditionBuilderTestSuite: baseSuite,
	}

	primaryKeyConditionsSuite := &PrimaryKeyConditionsTestSuite{
		ConditionBuilderTestSuite: baseSuite,
	}

	logicalGroupingSuite := &LogicalGroupingTestSuite{
		ConditionBuilderTestSuite: baseSuite,
	}

	conditionComprehensiveSuite := &ConditionComprehensiveTestSuite{
		ConditionBuilderTestSuite: baseSuite,
	}

	// Run all test suites
	t.Run("TestBasicComparison", func(t *testing.T) {
		suite.Run(t, basicComparisonSuite)
	})

	t.Run("TestRangeSetOperations", func(t *testing.T) {
		suite.Run(t, rangeSetOperationsSuite)
	})

	t.Run("TestSubqueryOperations", func(t *testing.T) {
		suite.Run(t, subqueryOperationsSuite)
	})

	t.Run("TestExpressionOperations", func(t *testing.T) {
		suite.Run(t, expressionOperationsSuite)
	})

	t.Run("TestNullBooleanChecks", func(t *testing.T) {
		suite.Run(t, nullBooleanChecksSuite)
	})

	t.Run("TestStringOperations", func(t *testing.T) {
		suite.Run(t, stringOperationsSuite)
	})

	t.Run("TestAuditConditions", func(t *testing.T) {
		suite.Run(t, auditConditionsSuite)
	})

	t.Run("TestPrimaryKeyConditions", func(t *testing.T) {
		suite.Run(t, primaryKeyConditionsSuite)
	})

	t.Run("TestLogicalGrouping", func(t *testing.T) {
		suite.Run(t, logicalGroupingSuite)
	})

	t.Run("TestConditionComprehensive", func(t *testing.T) {
		suite.Run(t, conditionComprehensiveSuite)
	})
}

// runAllExprBuilderTests executes all ExprBuilder test suites on the given database.
// This function is exported so it can be called from the parent orm package test runner.
func runAllExprBuilderTests(t *testing.T, ctx context.Context, dbType constants.DbType, db Db) {
	// Create test suites
	basicExpressionsSuite := &BasicExpressionsTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dbType,
			db:     db,
		},
	}

	aggregationFunctionsSuite := &AggregationFunctionsTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dbType,
			db:     db,
		},
	}

	windowFunctionsSuite := &WindowFunctionsTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dbType,
			db:     db,
		},
	}

	stringFunctionsSuite := &StringFunctionsTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dbType,
			db:     db,
		},
	}

	dateTimeFunctionsSuite := &DateTimeFunctionsTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dbType,
			db:     db,
		},
	}

	mathFunctionsSuite := &MathFunctionsTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dbType,
			db:     db,
		},
	}

	conditionalFunctionsSuite := &ConditionalFunctionsTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dbType,
			db:     db,
		},
	}

	typeConversionFunctionsSuite := &TypeConversionFunctionsTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dbType,
			db:     db,
		},
	}

	jsonFunctionsSuite := &JsonFunctionsTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dbType,
			db:     db,
		},
	}

	utilityFunctionsSuite := &UtilityFunctionsTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dbType,
			db:     db,
		},
	}

	comparisonExpressionsSuite := &ComparisonExpressionsTestSuite{
		OrmTestSuite: &OrmTestSuite{
			ctx:    ctx,
			dbType: dbType,
			db:     db,
		},
	}

	// Run all test suites
	t.Run("TestBasicExpressions", func(t *testing.T) {
		suite.Run(t, basicExpressionsSuite)
	})

	t.Run("TestComparisonExpressions", func(t *testing.T) {
		suite.Run(t, comparisonExpressionsSuite)
	})

	t.Run("TestAggregationFunctions", func(t *testing.T) {
		suite.Run(t, aggregationFunctionsSuite)
	})

	t.Run("TestStringFunctions", func(t *testing.T) {
		suite.Run(t, stringFunctionsSuite)
	})

	t.Run("TestDateTimeFunctions", func(t *testing.T) {
		suite.Run(t, dateTimeFunctionsSuite)
	})

	t.Run("TestMathFunctions", func(t *testing.T) {
		suite.Run(t, mathFunctionsSuite)
	})

	t.Run("TestConditionalFunctions", func(t *testing.T) {
		suite.Run(t, conditionalFunctionsSuite)
	})

	t.Run("TestTypeConversionFunctions", func(t *testing.T) {
		suite.Run(t, typeConversionFunctionsSuite)
	})

	t.Run("TestJsonFunctions", func(t *testing.T) {
		suite.Run(t, jsonFunctionsSuite)
	})

	t.Run("TestUtilityFunctions", func(t *testing.T) {
		suite.Run(t, utilityFunctionsSuite)
	})

	t.Run("TestWindowFunctions", func(t *testing.T) {
		suite.Run(t, windowFunctionsSuite)
	})
}

// TestPostgres runs all Orm tests against PostgreSQL.
func TestPostgres(t *testing.T) {
	ctx := context.Background()

	// Create a dummy suite for container management
	dummySuite := &suite.Suite{}
	dummySuite.SetT(t)

	// Start PostgreSQL container
	postgresContainer := testhelpers.NewPostgresContainer(ctx, dummySuite)
	defer postgresContainer.Terminate(ctx, dummySuite)

	// Run all Orm tests
	runAllOrmTests(t, ctx, postgresContainer.DsConfig)
}

// TestMySQL runs all Orm tests against MySQL.
func TestMySQL(t *testing.T) {
	ctx := context.Background()

	// Create a dummy suite for container management
	dummySuite := &suite.Suite{}
	dummySuite.SetT(t)

	// Start MySQL container
	mysqlContainer := testhelpers.NewMySQLContainer(ctx, dummySuite)
	defer mysqlContainer.Terminate(ctx, dummySuite)

	// Run all Orm tests
	runAllOrmTests(t, ctx, mysqlContainer.DsConfig)
}

// TestSQLite runs all Orm tests against SQLite (in-memory).
func TestSQLite(t *testing.T) {
	ctx := context.Background()

	// Create SQLite in-memory database config
	dsConfig := &config.DatasourceConfig{
		Type: constants.DbSQLite,
	}

	// Run all Orm tests
	runAllOrmTests(t, ctx, dsConfig)
}
