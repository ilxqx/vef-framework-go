package orm

import (
	"time"

	"github.com/ilxqx/vef-framework-go/constants"
)

// TypeConversionFunctionsTestSuite tests type conversion function methods of ExprBuilder
// including ToString, ToInteger, ToDecimal, ToFloat, ToBool, ToDate, ToTime, ToTimestamp,
// and ToJson.
//
// This suite verifies cross-database compatibility for type conversion functions across
// PostgreSQL, MySQL, and SQLite.
type TypeConversionFunctionsTestSuite struct {
	*OrmTestSuite
}

// TestToString tests the ToString function.
func (suite *TypeConversionFunctionsTestSuite) TestToString() {
	suite.T().Logf("Testing ToString function for %s", suite.DbType)

	suite.Run("ConvertNumberToString", func() {
		type ToStringResult struct {
			Id        string `bun:"id"`
			ViewCount int64  `bun:"view_count"`
			CountStr  string `bun:"count_str"`
		}

		var toStringResults []ToStringResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToString(eb.Column("view_count"))
			}, "count_str").
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &toStringResults)

		suite.NoError(err, "ToString should work")
		suite.True(len(toStringResults) > 0, "Should have ToString results")

		for _, result := range toStringResults {
			suite.NotEmpty(result.CountStr, "String representation should not be empty")
			suite.T().Logf("ID: %s, ViewCount: %d, CountStr: '%s'",
				result.Id, result.ViewCount, result.CountStr)
		}
	})
}

// TestToInteger tests the ToInteger function.
func (suite *TypeConversionFunctionsTestSuite) TestToInteger() {
	suite.T().Logf("Testing ToInteger function for %s", suite.DbType)

	suite.Run("ConvertStringToInteger", func() {
		type ToIntegerResult struct {
			Id       string `bun:"id"`
			Original string `bun:"original"`
			IntValue int64  `bun:"int_value"`
		}

		var toIntResults []ToIntegerResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToString(eb.Column("view_count"))
			}, "original").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToInteger(eb.ToString(eb.Column("view_count")))
			}, "int_value").
			Limit(5).
			Scan(suite.Ctx, &toIntResults)

		suite.NoError(err, "ToInteger should work")
		suite.True(len(toIntResults) > 0, "Should have ToInteger results")

		for _, result := range toIntResults {
			suite.True(result.IntValue >= 0, "Integer value should be non-negative")
			suite.T().Logf("ID: %s, Original: '%s', IntValue: %d",
				result.Id, result.Original, result.IntValue)
		}
	})
}

// TestToFloat tests the ToFloat function.
func (suite *TypeConversionFunctionsTestSuite) TestToFloat() {
	suite.T().Logf("Testing ToFloat function for %s", suite.DbType)

	suite.Run("ConvertNumberToFloat", func() {
		type ToFloatResult struct {
			Id         string  `bun:"id"`
			ViewCount  int64   `bun:"view_count"`
			FloatValue float64 `bun:"float_value"`
		}

		var toFloatResults []ToFloatResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToFloat(eb.Column("view_count"))
			}, "float_value").
			Limit(5).
			Scan(suite.Ctx, &toFloatResults)

		suite.NoError(err, "ToFloat should work")
		suite.True(len(toFloatResults) > 0, "Should have ToFloat results")

		for _, result := range toFloatResults {
			suite.Equal(float64(result.ViewCount), result.FloatValue, "Float value should equal view count")
			suite.T().Logf("ID: %s, ViewCount: %d, FloatValue: %.2f",
				result.Id, result.ViewCount, result.FloatValue)
		}
	})
}

// TestToDecimal tests the ToDecimal function.
func (suite *TypeConversionFunctionsTestSuite) TestToDecimal() {
	suite.T().Logf("Testing ToDecimal function for %s", suite.DbType)

	suite.Run("ConvertToDecimalWithPrecision", func() {
		type ToDecimalResult struct {
			Id           string  `bun:"id"`
			ViewCount    int64   `bun:"view_count"`
			DecimalValue float64 `bun:"decimal_value"`
		}

		var toDecimalResults []ToDecimalResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToDecimal(eb.Column("view_count"), 10, 2)
			}, "decimal_value").
			Limit(5).
			Scan(suite.Ctx, &toDecimalResults)

		suite.NoError(err, "ToDecimal should work")
		suite.True(len(toDecimalResults) > 0, "Should have ToDecimal results")

		for _, result := range toDecimalResults {
			suite.True(result.DecimalValue >= 0, "Decimal value should be non-negative")
			suite.T().Logf("ID: %s, ViewCount: %d, DecimalValue: %.2f",
				result.Id, result.ViewCount, result.DecimalValue)
		}
	})
}

// TestToBool tests the ToBool function.
func (suite *TypeConversionFunctionsTestSuite) TestToBool() {
	suite.T().Logf("Testing ToBool function for %s", suite.DbType)

	suite.Run("ConvertExpressionToBoolean", func() {
		type ToBoolResult struct {
			Id         string `bun:"id"`
			ViewCount  int64  `bun:"view_count"`
			IsPositive bool   `bun:"is_positive"`
		}

		var toBoolResults []ToBoolResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToBool(eb.Expr("CASE WHEN ? > 0 THEN 1 ELSE 0 END", eb.Column("view_count")))
			}, "is_positive").
			Limit(5).
			Scan(suite.Ctx, &toBoolResults)

		suite.NoError(err, "ToBool should work")
		suite.True(len(toBoolResults) > 0, "Should have ToBool results")

		for _, result := range toBoolResults {
			suite.T().Logf("ID: %s, ViewCount: %d, IsPositive: %v",
				result.Id, result.ViewCount, result.IsPositive)
		}
	})
}

// TestToDate tests the ToDate function.
func (suite *TypeConversionFunctionsTestSuite) TestToDate() {
	suite.T().Logf("Testing ToDate function for %s", suite.DbType)

	suite.Run("ConvertTimestampToDate", func() {
		type ToDateResult struct {
			Id        string    `bun:"id"`
			CreatedAt time.Time `bun:"created_at"`
			DateOnly  time.Time `bun:"date_only"`
		}

		var toDateResults []ToDateResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToDate(eb.Column("created_at"))
			}, "date_only").
			Limit(5).
			Scan(suite.Ctx, &toDateResults)

		suite.NoError(err, "ToDate should work")
		suite.True(len(toDateResults) > 0, "Should have ToDate results")

		for _, result := range toDateResults {
			suite.NotZero(result.DateOnly, "Date should not be zero")
			suite.T().Logf("ID: %s, CreatedAt: %s, DateOnly: %s",
				result.Id, result.CreatedAt.Format(time.RFC3339), result.DateOnly.Format(time.RFC3339))
		}
	})
}

// TestToTime tests the ToTime function.
func (suite *TypeConversionFunctionsTestSuite) TestToTime() {
	suite.T().Logf("Testing ToTime function for %s", suite.DbType)

	suite.Run("ConvertTimestampToTime", func() {
		type ToTimeResult struct {
			Id        string    `bun:"id"`
			CreatedAt time.Time `bun:"created_at"`
			TimeOnly  time.Time `bun:"time_only"`
		}

		var toTimeResults []ToTimeResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToTime(eb.Column("created_at"))
			}, "time_only").
			Limit(5).
			Scan(suite.Ctx, &toTimeResults)

		suite.NoError(err, "ToTime should work")
		suite.True(len(toTimeResults) > 0, "Should have ToTime results")

		for _, result := range toTimeResults {
			suite.NotZero(result.TimeOnly, "Time should not be zero")
			suite.T().Logf("ID: %s, CreatedAt: %s, TimeOnly: %s",
				result.Id, result.CreatedAt.Format(time.RFC3339), result.TimeOnly.Format(time.RFC3339))
		}
	})
}

// TestToTimestamp tests the ToTimestamp function.
func (suite *TypeConversionFunctionsTestSuite) TestToTimestamp() {
	suite.T().Logf("Testing ToTimestamp function for %s", suite.DbType)

	suite.Run("ConvertToTimestamp", func() {
		type ToTimestampResult struct {
			Id        string    `bun:"id"`
			CreatedAt time.Time `bun:"created_at"`
			Timestamp time.Time `bun:"timestamp"`
		}

		var toTimestampResults []ToTimestampResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToTimestamp(eb.Column("created_at"))
			}, "timestamp").
			Limit(5).
			Scan(suite.Ctx, &toTimestampResults)

		suite.NoError(err, "ToTimestamp should work")
		suite.True(len(toTimestampResults) > 0, "Should have ToTimestamp results")

		for _, result := range toTimestampResults {
			suite.NotZero(result.Timestamp, "Timestamp should not be zero")
			suite.T().Logf("ID: %s, CreatedAt: %s, Timestamp: %s",
				result.Id, result.CreatedAt.Format(time.RFC3339), result.Timestamp.Format(time.RFC3339))
		}
	})
}

// TestToJson tests the ToJson function.
func (suite *TypeConversionFunctionsTestSuite) TestToJson() {
	suite.T().Logf("Testing ToJson function for %s", suite.DbType)

	suite.Run("ConvertToJson", func() {
		type ToJsonResult struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			JsonValue string `bun:"json_value"`
		}

		var toJsonResults []ToJsonResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToJson(eb.JsonObject("title", eb.Column("title"), "id", eb.Column("id")))
			}, "json_value").
			Limit(3).
			Scan(suite.Ctx, &toJsonResults)

		suite.NoError(err, "ToJson should work on supported databases")
		suite.True(len(toJsonResults) > 0, "Should have ToJson results")

		for _, result := range toJsonResults {
			suite.NotEmpty(result.JsonValue, "Json value should not be empty")
			suite.T().Logf("ID: %s, Title: %s, JsonValue: %s",
				result.Id, result.Title, result.JsonValue)
		}
	})
}

// TestToStringNullHandling tests the ToString function with NULL values.
func (suite *TypeConversionFunctionsTestSuite) TestToStringNullHandling() {
	suite.T().Logf("Testing ToString NULL handling for %s", suite.DbType)

	suite.Run("ConvertNullToString", func() {
		type NullToStringResult struct {
			Id         string  `bun:"id"`
			Title      string  `bun:"title"`
			StringNull *string `bun:"string_null"`
		}

		var results []NullToStringResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToString(eb.Expr("NULL"))
			}, "string_null").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToString(NULL) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Nil(result.StringNull, "ToString(NULL) should return NULL")
			suite.T().Logf("ID: %s, Title: %s, StringNull: %v",
				result.Id, result.Title, result.StringNull)
		}
	})
}

// TestToIntegerNullHandling tests the ToInteger function with NULL values.
func (suite *TypeConversionFunctionsTestSuite) TestToIntegerNullHandling() {
	suite.T().Logf("Testing ToInteger NULL handling for %s", suite.DbType)

	suite.Run("ConvertNullToInteger", func() {
		type NullToIntegerResult struct {
			Id      string `bun:"id"`
			Title   string `bun:"title"`
			IntNull *int64 `bun:"int_null"`
		}

		var results []NullToIntegerResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToInteger(eb.Expr("NULL"))
			}, "int_null").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToInteger(NULL) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Nil(result.IntNull, "ToInteger(NULL) should return NULL")
			suite.T().Logf("ID: %s, Title: %s, IntNull: %v",
				result.Id, result.Title, result.IntNull)
		}
	})
}

// TestToFloatNullHandling tests the ToFloat function with NULL values.
func (suite *TypeConversionFunctionsTestSuite) TestToFloatNullHandling() {
	suite.T().Logf("Testing ToFloat NULL handling for %s", suite.DbType)

	suite.Run("ConvertNullToFloat", func() {
		type NullToFloatResult struct {
			Id        string   `bun:"id"`
			Title     string   `bun:"title"`
			FloatNull *float64 `bun:"float_null"`
		}

		var results []NullToFloatResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToFloat(eb.Expr("NULL"))
			}, "float_null").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToFloat(NULL) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Nil(result.FloatNull, "ToFloat(NULL) should return NULL")
			suite.T().Logf("ID: %s, Title: %s, FloatNull: %v",
				result.Id, result.Title, result.FloatNull)
		}
	})
}

// TestToDecimalNullHandling tests the ToDecimal function with NULL values.
func (suite *TypeConversionFunctionsTestSuite) TestToDecimalNullHandling() {
	suite.T().Logf("Testing ToDecimal NULL handling for %s", suite.DbType)

	suite.Run("ConvertNullToDecimal", func() {
		type NullToDecimalResult struct {
			Id          string   `bun:"id"`
			Title       string   `bun:"title"`
			DecimalNull *float64 `bun:"decimal_null"`
		}

		var results []NullToDecimalResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToDecimal(eb.Expr("NULL"), 10, 2)
			}, "decimal_null").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToDecimal(NULL) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Nil(result.DecimalNull, "ToDecimal(NULL) should return NULL")
			suite.T().Logf("ID: %s, Title: %s, DecimalNull: %v",
				result.Id, result.Title, result.DecimalNull)
		}
	})
}

// TestToBoolNullHandling tests the ToBool function with NULL values.
func (suite *TypeConversionFunctionsTestSuite) TestToBoolNullHandling() {
	suite.T().Logf("Testing ToBool NULL handling for %s", suite.DbType)

	suite.Run("ConvertNullToBool", func() {
		type NullToBoolResult struct {
			Id       string `bun:"id"`
			Title    string `bun:"title"`
			BoolNull *bool  `bun:"bool_null"`
		}

		var results []NullToBoolResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToBool(eb.Expr("NULL"))
			}, "bool_null").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToBool(NULL) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Nil(result.BoolNull, "ToBool(NULL) should return NULL")
			suite.T().Logf("ID: %s, Title: %s, BoolNull: %v",
				result.Id, result.Title, result.BoolNull)
		}
	})
}

// TestToDateNullHandling tests the ToDate function with NULL values.
func (suite *TypeConversionFunctionsTestSuite) TestToDateNullHandling() {
	suite.T().Logf("Testing ToDate NULL handling for %s", suite.DbType)

	suite.Run("ConvertNullToDate", func() {
		type NullToDateResult struct {
			Id       string     `bun:"id"`
			Title    string     `bun:"title"`
			DateNull *time.Time `bun:"date_null"`
		}

		var results []NullToDateResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToDate(eb.Expr("NULL"))
			}, "date_null").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToDate(NULL) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Nil(result.DateNull, "ToDate(NULL) should return NULL")
			suite.T().Logf("ID: %s, Title: %s, DateNull: %v",
				result.Id, result.Title, result.DateNull)
		}
	})
}

// TestToTimeNullHandling tests the ToTime function with NULL values.
func (suite *TypeConversionFunctionsTestSuite) TestToTimeNullHandling() {
	suite.T().Logf("Testing ToTime NULL handling for %s", suite.DbType)

	suite.Run("ConvertNullToTime", func() {
		type NullToTimeResult struct {
			Id       string     `bun:"id"`
			Title    string     `bun:"title"`
			TimeNull *time.Time `bun:"time_null"`
		}

		var results []NullToTimeResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToTime(eb.Expr("NULL"))
			}, "time_null").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToTime(NULL) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Nil(result.TimeNull, "ToTime(NULL) should return NULL")
			suite.T().Logf("ID: %s, Title: %s, TimeNull: %v",
				result.Id, result.Title, result.TimeNull)
		}
	})
}

// TestToTimestampNullHandling tests the ToTimestamp function with NULL values.
func (suite *TypeConversionFunctionsTestSuite) TestToTimestampNullHandling() {
	suite.T().Logf("Testing ToTimestamp NULL handling for %s", suite.DbType)

	suite.Run("ConvertNullToTimestamp", func() {
		type NullToTimestampResult struct {
			Id            string     `bun:"id"`
			Title         string     `bun:"title"`
			TimestampNull *time.Time `bun:"timestamp_null"`
		}

		var results []NullToTimestampResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToTimestamp(eb.Expr("NULL"))
			}, "timestamp_null").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToTimestamp(NULL) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Nil(result.TimestampNull, "ToTimestamp(NULL) should return NULL")
			suite.T().Logf("ID: %s, Title: %s, TimestampNull: %v",
				result.Id, result.Title, result.TimestampNull)
		}
	})
}

// TestToJsonNullHandling tests the ToJson function with NULL values.
func (suite *TypeConversionFunctionsTestSuite) TestToJsonNullHandling() {
	suite.T().Logf("Testing ToJson NULL handling for %s", suite.DbType)

	suite.Run("ConvertNullToJson", func() {
		type NullToJsonResult struct {
			Id       string  `bun:"id"`
			Title    string  `bun:"title"`
			JsonNull *string `bun:"json_null"`
		}

		var results []NullToJsonResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToJson(eb.Expr("NULL"))
			}, "json_null").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToJson(NULL) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Nil(result.JsonNull, "ToJson(NULL) should return NULL")
			suite.T().Logf("ID: %s, Title: %s, JsonNull: %v",
				result.Id, result.Title, result.JsonNull)
		}
	})
}

// TestToDecimalPrecisionVariations tests the ToDecimal function with different precision parameters.
func (suite *TypeConversionFunctionsTestSuite) TestToDecimalPrecisionVariations() {
	suite.T().Logf("Testing ToDecimal precision variations for %s", suite.DbType)

	suite.Run("DecimalWithPrecisionAndScale", func() {
		type DecimalPrecisionResult struct {
			Id           string  `bun:"id"`
			ViewCount    int64   `bun:"view_count"`
			DecimalValue float64 `bun:"decimal_value"`
		}

		var results []DecimalPrecisionResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToDecimal(eb.Column("view_count"), 10, 2)
			}, "decimal_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToDecimal with precision and scale should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.T().Logf("ID: %s, ViewCount: %d, DecimalValue: %.2f",
				result.Id, result.ViewCount, result.DecimalValue)
		}
	})

	suite.Run("DecimalWithPrecisionOnly", func() {
		type DecimalPrecisionOnlyResult struct {
			Id           string  `bun:"id"`
			ViewCount    int64   `bun:"view_count"`
			DecimalValue float64 `bun:"decimal_value"`
		}

		var results []DecimalPrecisionOnlyResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToDecimal(eb.Column("view_count"), 10)
			}, "decimal_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToDecimal with precision only should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.T().Logf("ID: %s, ViewCount: %d, DecimalValue: %.0f",
				result.Id, result.ViewCount, result.DecimalValue)
		}
	})

	suite.Run("DecimalWithoutParameters", func() {
		type DecimalNoParamsResult struct {
			Id           string  `bun:"id"`
			ViewCount    int64   `bun:"view_count"`
			DecimalValue float64 `bun:"decimal_value"`
		}

		var results []DecimalNoParamsResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToDecimal(eb.Column("view_count"))
			}, "decimal_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToDecimal without parameters should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.T().Logf("ID: %s, ViewCount: %d, DecimalValue: %.0f",
				result.Id, result.ViewCount, result.DecimalValue)
		}
	})
}

// TestToDateWithFormat tests the ToDate function with format parameter.
func (suite *TypeConversionFunctionsTestSuite) TestToDateWithFormat() {
	suite.T().Logf("Testing ToDate with format for %s", suite.DbType)

	suite.Run("DateWithoutFormat", func() {
		type DateNoFormatResult struct {
			Id        string    `bun:"id"`
			CreatedAt time.Time `bun:"created_at"`
			DateValue time.Time `bun:"date_value"`
		}

		var results []DateNoFormatResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToDate(eb.Column("created_at"))
			}, "date_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToDate without format should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.NotZero(result.DateValue, "Date should not be zero")
			suite.T().Logf("ID: %s, CreatedAt: %s, DateValue: %s",
				result.Id, result.CreatedAt.Format(time.RFC3339), result.DateValue.Format(time.RFC3339))
		}
	})

	suite.Run("DateWithFormat", func() {
		type DateWithFormatResult struct {
			Id        string    `bun:"id"`
			DateValue time.Time `bun:"date_value"`
		}

		var results []DateWithFormatResult

		var formatStr string
		switch suite.DbType {
		case constants.DbPostgres:
			formatStr = "YYYY-MM-DD"
		case constants.DbMySQL:
			formatStr = "%Y-%m-%d"
		default:
			formatStr = "YYYY-MM-DD"
		}

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToDate(eb.Expr("?", "2024-01-15"), formatStr)
			}, "date_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToDate with format should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.NotZero(result.DateValue, "Date should not be zero")
			suite.T().Logf("ID: %s, DateValue: %s",
				result.Id, result.DateValue.Format(time.RFC3339))
		}
	})
}

// TestToTimeWithFormat tests the ToTime function with format parameter.
func (suite *TypeConversionFunctionsTestSuite) TestToTimeWithFormat() {
	suite.T().Logf("Testing ToTime with format for %s", suite.DbType)

	suite.Run("TimeWithoutFormat", func() {
		type TimeNoFormatResult struct {
			Id        string    `bun:"id"`
			CreatedAt time.Time `bun:"created_at"`
			TimeValue time.Time `bun:"time_value"`
		}

		var results []TimeNoFormatResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToTime(eb.Column("created_at"))
			}, "time_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToTime without format should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.NotZero(result.TimeValue, "Time should not be zero")
			suite.T().Logf("ID: %s, CreatedAt: %s, TimeValue: %s",
				result.Id, result.CreatedAt.Format(time.RFC3339), result.TimeValue.Format(time.RFC3339))
		}
	})

	suite.Run("TimeWithFormat", func() {
		type TimeWithFormatResult struct {
			Id        string    `bun:"id"`
			TimeValue time.Time `bun:"time_value"`
		}

		var results []TimeWithFormatResult

		var formatStr string
		switch suite.DbType {
		case constants.DbPostgres:
			formatStr = "HH24:MI:SS"
		case constants.DbMySQL:
			formatStr = "%H:%i:%s"
		default:
			formatStr = "HH24:MI:SS"
		}

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToTime(eb.Expr("?", "14:30:00"), formatStr)
			}, "time_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToTime with format should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.NotZero(result.TimeValue, "Time should not be zero")
			suite.T().Logf("ID: %s, TimeValue: %s",
				result.Id, result.TimeValue.Format(time.RFC3339))
		}
	})
}

// TestToTimestampWithFormat tests the ToTimestamp function with format parameter.
func (suite *TypeConversionFunctionsTestSuite) TestToTimestampWithFormat() {
	suite.T().Logf("Testing ToTimestamp with format for %s", suite.DbType)

	suite.Run("TimestampWithoutFormat", func() {
		type TimestampNoFormatResult struct {
			Id             string    `bun:"id"`
			CreatedAt      time.Time `bun:"created_at"`
			TimestampValue time.Time `bun:"timestamp_value"`
		}

		var results []TimestampNoFormatResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToTimestamp(eb.Column("created_at"))
			}, "timestamp_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToTimestamp without format should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.NotZero(result.TimestampValue, "Timestamp should not be zero")
			suite.T().Logf("ID: %s, CreatedAt: %s, TimestampValue: %s",
				result.Id, result.CreatedAt.Format(time.RFC3339), result.TimestampValue.Format(time.RFC3339))
		}
	})

	suite.Run("TimestampWithFormat", func() {
		type TimestampWithFormatResult struct {
			Id             string    `bun:"id"`
			TimestampValue time.Time `bun:"timestamp_value"`
		}

		var results []TimestampWithFormatResult

		var formatStr string
		switch suite.DbType {
		case constants.DbPostgres:
			formatStr = "YYYY-MM-DD HH24:MI:SS"
		case constants.DbMySQL:
			formatStr = "%Y-%m-%d %H:%i:%s"
		default:
			formatStr = "YYYY-MM-DD HH24:MI:SS"
		}

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToTimestamp(eb.Expr("?", "2024-01-15 14:30:00"), formatStr)
			}, "timestamp_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToTimestamp with format should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.NotZero(result.TimestampValue, "Timestamp should not be zero")
			suite.T().Logf("ID: %s, TimestampValue: %s",
				result.Id, result.TimestampValue.Format(time.RFC3339))
		}
	})
}

// TestToStringFromDifferentTypes tests the ToString function with different source types.
func (suite *TypeConversionFunctionsTestSuite) TestToStringFromDifferentTypes() {
	suite.T().Logf("Testing ToString from different types for %s", suite.DbType)

	suite.Run("ConvertBooleanToString", func() {
		type BoolToStringResult struct {
			Id           string `bun:"id"`
			IsActive     bool   `bun:"is_active"`
			ActiveString string `bun:"active_string"`
		}

		var results []BoolToStringResult

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "is_active").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToString(eb.Column("is_active"))
			}, "active_string").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToString(boolean) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.NotEmpty(result.ActiveString, "Boolean string should not be empty")
			suite.T().Logf("ID: %s, IsActive: %v, ActiveString: '%s'",
				result.Id, result.IsActive, result.ActiveString)
		}
	})

	suite.Run("ConvertDateToString", func() {
		type DateToStringResult struct {
			Id         string    `bun:"id"`
			CreatedAt  time.Time `bun:"created_at"`
			DateString string    `bun:"date_string"`
		}

		var results []DateToStringResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToString(eb.ToDate(eb.Column("created_at")))
			}, "date_string").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToString(date) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.NotEmpty(result.DateString, "Date string should not be empty")
			suite.T().Logf("ID: %s, CreatedAt: %s, DateString: '%s'",
				result.Id, result.CreatedAt.Format(time.RFC3339), result.DateString)
		}
	})
}

// TestToIntegerFromStrings tests the ToInteger function with string sources.
func (suite *TypeConversionFunctionsTestSuite) TestToIntegerFromStrings() {
	suite.T().Logf("Testing ToInteger from strings for %s", suite.DbType)

	suite.Run("ConvertNegativeStringToInteger", func() {
		type NegativeIntResult struct {
			Id       string `bun:"id"`
			IntValue int64  `bun:"int_value"`
		}

		var results []NegativeIntResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToInteger(eb.Expr("?", "-123"))
			}, "int_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToInteger(negative string) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Equal(int64(-123), result.IntValue, "Should convert negative string correctly")
			suite.T().Logf("ID: %s, IntValue: %d", result.Id, result.IntValue)
		}
	})

	suite.Run("ConvertZeroStringToInteger", func() {
		type ZeroIntResult struct {
			Id       string `bun:"id"`
			IntValue int64  `bun:"int_value"`
		}

		var results []ZeroIntResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToInteger(eb.Expr("?", "0"))
			}, "int_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToInteger('0') should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Equal(int64(0), result.IntValue, "Should convert '0' correctly")
			suite.T().Logf("ID: %s, IntValue: %d", result.Id, result.IntValue)
		}
	})
}

// TestToFloatFromStrings tests the ToFloat function with string sources.
func (suite *TypeConversionFunctionsTestSuite) TestToFloatFromStrings() {
	suite.T().Logf("Testing ToFloat from strings for %s", suite.DbType)

	suite.Run("ConvertDecimalStringToFloat", func() {
		type DecimalStringResult struct {
			Id         string  `bun:"id"`
			FloatValue float64 `bun:"float_value"`
		}

		var results []DecimalStringResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToFloat(eb.Expr("?", "3.14159"))
			}, "float_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToFloat('3.14159') should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.InDelta(3.14159, result.FloatValue, 0.00001, "Should convert decimal string correctly")
			suite.T().Logf("ID: %s, FloatValue: %.5f", result.Id, result.FloatValue)
		}
	})

	suite.Run("ConvertNegativeFloatString", func() {
		type NegativeFloatResult struct {
			Id         string  `bun:"id"`
			FloatValue float64 `bun:"float_value"`
		}

		var results []NegativeFloatResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToFloat(eb.Expr("?", "-99.99"))
			}, "float_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToFloat('-99.99') should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.InDelta(-99.99, result.FloatValue, 0.01, "Should convert negative float string correctly")
			suite.T().Logf("ID: %s, FloatValue: %.2f", result.Id, result.FloatValue)
		}
	})
}

// TestToBoolDirectConversion tests the ToBool function with direct numeric conversion.
func (suite *TypeConversionFunctionsTestSuite) TestToBoolDirectConversion() {
	suite.T().Logf("Testing ToBool direct conversion for %s", suite.DbType)

	suite.Run("ConvertPositiveIntegerToBool", func() {
		type PositiveIntBoolResult struct {
			Id        string `bun:"id"`
			BoolValue bool   `bun:"bool_value"`
		}

		var results []PositiveIntBoolResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToBool(eb.Expr("?", 1))
			}, "bool_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToBool(1) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.True(result.BoolValue, "ToBool(1) should return true")
			suite.T().Logf("ID: %s, BoolValue: %v", result.Id, result.BoolValue)
		}
	})

	suite.Run("ConvertZeroToBool", func() {
		type ZeroBoolResult struct {
			Id        string `bun:"id"`
			BoolValue bool   `bun:"bool_value"`
		}

		var results []ZeroBoolResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToBool(eb.Expr("?", 0))
			}, "bool_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToBool(0) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.False(result.BoolValue, "ToBool(0) should return false")
			suite.T().Logf("ID: %s, BoolValue: %v", result.Id, result.BoolValue)
		}
	})

	suite.Run("ConvertNegativeIntegerToBool", func() {
		type NegativeIntBoolResult struct {
			Id        string `bun:"id"`
			BoolValue bool   `bun:"bool_value"`
		}

		var results []NegativeIntBoolResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToBool(eb.Expr("?", -1))
			}, "bool_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToBool(-1) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.True(result.BoolValue, "ToBool(-1) should return true (non-zero)")
			suite.T().Logf("ID: %s, BoolValue: %v", result.Id, result.BoolValue)
		}
	})
}

// TestToIntegerBoundaryConditions tests the ToInteger function with boundary values.
func (suite *TypeConversionFunctionsTestSuite) TestToIntegerBoundaryConditions() {
	suite.T().Logf("Testing ToInteger boundary conditions for %s", suite.DbType)

	suite.Run("ConvertLargePositiveInteger", func() {
		type LargePositiveResult struct {
			Id       string `bun:"id"`
			IntValue int64  `bun:"int_value"`
		}

		var results []LargePositiveResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToInteger(eb.Expr("?", 2147483647))
			}, "int_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToInteger(large positive) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Equal(int64(2147483647), result.IntValue, "Should handle large positive integer")
			suite.T().Logf("ID: %s, IntValue: %d", result.Id, result.IntValue)
		}
	})

	suite.Run("ConvertLargeNegativeInteger", func() {
		type LargeNegativeResult struct {
			Id       string `bun:"id"`
			IntValue int64  `bun:"int_value"`
		}

		var results []LargeNegativeResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToInteger(eb.Expr("?", -2147483647))
			}, "int_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToInteger(large negative) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Equal(int64(-2147483647), result.IntValue, "Should handle large negative integer")
			suite.T().Logf("ID: %s, IntValue: %d", result.Id, result.IntValue)
		}
	})
}

// TestToFloatPrecisionAndBoundaries tests the ToFloat function with precision and boundary values.
func (suite *TypeConversionFunctionsTestSuite) TestToFloatPrecisionAndBoundaries() {
	suite.T().Logf("Testing ToFloat precision and boundaries for %s", suite.DbType)

	suite.Run("ConvertVerySmallFloat", func() {
		type VerySmallFloatResult struct {
			Id         string  `bun:"id"`
			FloatValue float64 `bun:"float_value"`
		}

		var results []VerySmallFloatResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToFloat(eb.Expr("?", 0.000001))
			}, "float_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToFloat(very small) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.InDelta(0.000001, result.FloatValue, 0.0000001, "Should handle very small float")
			suite.T().Logf("ID: %s, FloatValue: %.7f", result.Id, result.FloatValue)
		}
	})

	suite.Run("ConvertVeryLargeFloat", func() {
		type VeryLargeFloatResult struct {
			Id         string  `bun:"id"`
			FloatValue float64 `bun:"float_value"`
		}

		var results []VeryLargeFloatResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToFloat(eb.Expr("?", 999999999.99))
			}, "float_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToFloat(very large) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.InDelta(999999999.99, result.FloatValue, 0.01, "Should handle very large float")
			suite.T().Logf("ID: %s, FloatValue: %.2f", result.Id, result.FloatValue)
		}
	})

	suite.Run("ConvertZeroFloat", func() {
		type ZeroFloatResult struct {
			Id         string  `bun:"id"`
			FloatValue float64 `bun:"float_value"`
		}

		var results []ZeroFloatResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToFloat(eb.Expr("?", 0.0))
			}, "float_value").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToFloat(0.0) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Equal(float64(0.0), result.FloatValue, "Should handle zero float")
			suite.T().Logf("ID: %s, FloatValue: %.1f", result.Id, result.FloatValue)
		}
	})
}

// TestToBoolDatabaseSpecificBehavior tests ToBool function behavior across databases.
func (suite *TypeConversionFunctionsTestSuite) TestToBoolDatabaseSpecificBehavior() {
	suite.T().Logf("Testing ToBool database-specific behavior for %s", suite.DbType)

	suite.Run("VerifyBooleanRepresentation", func() {
		type BoolRepresentationResult struct {
			Id         string `bun:"id"`
			IsActive   bool   `bun:"is_active"`
			BoolAsBool bool   `bun:"bool_as_bool"`
		}

		var results []BoolRepresentationResult

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "is_active").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToBool(eb.Column("is_active"))
			}, "bool_as_bool").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ToBool(column) should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Equal(result.IsActive, result.BoolAsBool, "ToBool should preserve boolean value")
			suite.T().Logf("ID: %s, IsActive: %v, BoolAsBool: %v (DB: %s)",
				result.Id, result.IsActive, result.BoolAsBool, suite.DbType)
		}
	})
}
