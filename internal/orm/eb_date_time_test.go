package orm

import (
	"time"

	"github.com/ilxqx/vef-framework-go/constants"
)

// DateTimeFunctionsTestSuite tests date and time manipulation methods of ExprBuilder
// including CurrentDate, CurrentTime, CurrentTimestamp, Now, date extraction functions
// (ExtractYear, ExtractMonth, ExtractDay, ExtractHour, ExtractMinute, ExtractSecond),
// DateTrunc, DateAdd, DateSubtract, DateDiff, and Age.
//
// This suite verifies cross-database compatibility for date/time functions across
// PostgreSQL, MySQL, and SQLite. Note that SQLite has limited support for many
// date/time functions and some tests are skipped accordingly.
type DateTimeFunctionsTestSuite struct {
	*OrmTestSuite
}

// TestCurrentDate tests the CurrentDate function.
// CurrentDate returns the current date without time component.
func (suite *DateTimeFunctionsTestSuite) TestCurrentDate() {
	suite.T().Logf("Testing CurrentDate function for %s", suite.dbType)

	// Test 1: Get current date using CurrentDate()
	suite.Run("GetCurrentDate", func() {
		type CurrentDateResult struct {
			CurrentDate time.Time `bun:"current_date"`
		}

		var result CurrentDateResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CurrentDate()
			}, "current_date").
			Limit(1).
			Scan(suite.ctx, &result)

		suite.NoError(err, "CurrentDate query should execute successfully")
		suite.NotZero(result.CurrentDate, "CurrentDate should return a non-zero date value")

		now := time.Now()
		suite.True(result.CurrentDate.Year() >= now.Year()-1, "CurrentDate year should be within reasonable range (current year or previous year)")

		suite.T().Logf("CurrentDate: %v", result.CurrentDate)
	})
}

// TestCurrentTime tests the CurrentTime function.
// CurrentTime returns the current time without date component.
func (suite *DateTimeFunctionsTestSuite) TestCurrentTime() {
	suite.T().Logf("Testing CurrentTime function for %s", suite.dbType)

	// Test 1: Get current time using CurrentTime()
	suite.Run("GetCurrentTime", func() {
		type CurrentTimeResult struct {
			CurrentTime time.Time `bun:"current_time"`
		}

		var result CurrentTimeResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CurrentTime()
			}, "current_time").
			Limit(1).
			Scan(suite.ctx, &result)

		suite.NoError(err, "CurrentTime query should execute successfully")
		suite.NotZero(result.CurrentTime, "CurrentTime should return a non-zero time value")

		suite.T().Logf("CurrentTime: %v", result.CurrentTime)
	})
}

// TestCurrentTimestamp tests the CurrentTimestamp function.
// CurrentTimestamp returns the current date and time.
func (suite *DateTimeFunctionsTestSuite) TestCurrentTimestamp() {
	suite.T().Logf("Testing CurrentTimestamp function for %s", suite.dbType)

	// Test 1: Get current timestamp using CurrentTimestamp()
	suite.Run("GetCurrentTimestamp", func() {
		type CurrentTimestampResult struct {
			CurrentTimestamp time.Time `bun:"current_timestamp"`
		}

		var result CurrentTimestampResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CurrentTimestamp()
			}, "current_timestamp").
			Limit(1).
			Scan(suite.ctx, &result)

		suite.NoError(err, "CurrentTimestamp query should execute successfully")
		suite.NotZero(result.CurrentTimestamp, "CurrentTimestamp should return a non-zero timestamp value")

		now := time.Now()
		suite.True(result.CurrentTimestamp.Year() >= now.Year()-1, "CurrentTimestamp year should be within reasonable range (current year or previous year)")

		suite.T().Logf("CurrentTimestamp: %v", result.CurrentTimestamp)
	})
}

// TestNow tests the Now function.
// Now returns the current timestamp (alias for CurrentTimestamp).
func (suite *DateTimeFunctionsTestSuite) TestNow() {
	suite.T().Logf("Testing Now function for %s", suite.dbType)

	// Test 1: Get current timestamp using Now()
	suite.Run("GetNow", func() {
		type NowResult struct {
			Now time.Time `bun:"now"`
		}

		var result NowResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Now()
			}, "now").
			Limit(1).
			Scan(suite.ctx, &result)

		suite.NoError(err, "Now query should execute successfully")
		suite.NotZero(result.Now, "Now should return a non-zero timestamp value")

		now := time.Now()
		suite.True(result.Now.Year() >= now.Year()-1, "Now year should be within reasonable range (current year or previous year)")

		suite.T().Logf("Now: %v", result.Now)
	})

	// Test 2: Use all current time functions together (CurrentDate, CurrentTime, CurrentTimestamp, Now)
	suite.Run("AllCurrentTimeFunctions", func() {
		type AllCurrentResult struct {
			CurrentDate      time.Time `bun:"current_date"`
			CurrentTime      time.Time `bun:"current_time"`
			CurrentTimestamp time.Time `bun:"current_timestamp"`
			Now              time.Time `bun:"now"`
		}

		var result AllCurrentResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CurrentDate()
			}, "current_date").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CurrentTime()
			}, "current_time").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CurrentTimestamp()
			}, "current_timestamp").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Now()
			}, "now").
			Limit(1).
			Scan(suite.ctx, &result)

		suite.NoError(err, "Combined current time functions query should execute successfully")
		suite.NotZero(result.CurrentDate, "CurrentDate should return a non-zero value in combined query")
		suite.NotZero(result.CurrentTime, "CurrentTime should return a non-zero value in combined query")
		suite.NotZero(result.CurrentTimestamp, "CurrentTimestamp should return a non-zero value in combined query")
		suite.NotZero(result.Now, "Now should return a non-zero value in combined query")

		suite.T().Logf("CurrentDate: %v | CurrentTime: %v | CurrentTimestamp: %v | Now: %v",
			result.CurrentDate, result.CurrentTime, result.CurrentTimestamp, result.Now)
	})
}

// TestExtractYear tests the ExtractYear function.
// ExtractYear extracts the year from a date/timestamp.
func (suite *DateTimeFunctionsTestSuite) TestExtractYear() {
	suite.T().Logf("Testing ExtractYear function for %s", suite.dbType)

	// Test 1: Extract year from created_at column
	suite.Run("ExtractYearFromCreatedAt", func() {
		type YearResult struct {
			CreatedAt time.Time `bun:"created_at"`
			Year      int64     `bun:"year"`
		}

		var results []YearResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ExtractYear(eb.Column("created_at"))
			}, "year").
			OrderBy("created_at").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "ExtractYear query should execute successfully")
		suite.True(len(results) > 0, "ExtractYear should return at least one result")

		for _, result := range results {
			suite.Equal(int64(result.CreatedAt.Year()), result.Year, "Extracted year should match the actual year from created_at timestamp")
			suite.T().Logf("CreatedAt: %v | Year: %d", result.CreatedAt, result.Year)
		}
	})
}

// TestExtractMonth tests the ExtractMonth function.
// ExtractMonth extracts the month from a date/timestamp.
func (suite *DateTimeFunctionsTestSuite) TestExtractMonth() {
	suite.T().Logf("Testing ExtractMonth function for %s", suite.dbType)

	// Test 1: Extract month from created_at column
	suite.Run("ExtractMonthFromCreatedAt", func() {
		type MonthResult struct {
			CreatedAt time.Time `bun:"created_at"`
			Month     int64     `bun:"month"`
		}

		var results []MonthResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ExtractMonth(eb.Column("created_at"))
			}, "month").
			OrderBy("created_at").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "ExtractMonth query should execute successfully")
		suite.True(len(results) > 0, "ExtractMonth should return at least one result")

		for _, result := range results {
			suite.Equal(int64(result.CreatedAt.Month()), result.Month, "Extracted month should match the actual month from created_at timestamp")
			suite.True(result.Month >= 1 && result.Month <= 12, "Extracted month should be in valid range (1-12)")
			suite.T().Logf("CreatedAt: %v | Month: %d", result.CreatedAt, result.Month)
		}
	})
}

// TestExtractDay tests the ExtractDay function.
// ExtractDay extracts the day from a date/timestamp.
func (suite *DateTimeFunctionsTestSuite) TestExtractDay() {
	suite.T().Logf("Testing ExtractDay function for %s", suite.dbType)

	// Test 1: Extract day from created_at column
	suite.Run("ExtractDayFromCreatedAt", func() {
		type DayResult struct {
			CreatedAt time.Time `bun:"created_at"`
			Day       int64     `bun:"day"`
		}

		var results []DayResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ExtractDay(eb.Column("created_at"))
			}, "day").
			OrderBy("created_at").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "ExtractDay query should execute successfully")
		suite.True(len(results) > 0, "ExtractDay should return at least one result")

		for _, result := range results {
			suite.Equal(int64(result.CreatedAt.Day()), result.Day, "Extracted day should match the actual day from created_at timestamp")
			suite.True(result.Day >= 1 && result.Day <= 31, "Extracted day should be in valid range (1-31)")
			suite.T().Logf("CreatedAt: %v | Day: %d", result.CreatedAt, result.Day)
		}
	})
}

// TestExtractHour tests the ExtractHour function.
// ExtractHour extracts the hour from a timestamp.
func (suite *DateTimeFunctionsTestSuite) TestExtractHour() {
	suite.T().Logf("Testing ExtractHour function for %s", suite.dbType)

	// Test 1: Extract hour from created_at column
	suite.Run("ExtractHourFromCreatedAt", func() {
		type HourResult struct {
			CreatedAt time.Time `bun:"created_at"`
			Hour      int64     `bun:"hour"`
		}

		var results []HourResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ExtractHour(eb.Column("created_at"))
			}, "hour").
			OrderBy("created_at").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "ExtractHour query should execute successfully")
		suite.True(len(results) > 0, "ExtractHour should return at least one result")

		for _, result := range results {
			suite.Equal(int64(result.CreatedAt.Hour()), result.Hour, "Extracted hour should match the actual hour from created_at timestamp")
			suite.True(result.Hour >= 0 && result.Hour < 24, "Extracted hour should be in valid range (0-23)")
			suite.T().Logf("CreatedAt: %v | Hour: %d", result.CreatedAt, result.Hour)
		}
	})
}

// TestExtractMinute tests the ExtractMinute function.
// ExtractMinute extracts the minute from a timestamp.
func (suite *DateTimeFunctionsTestSuite) TestExtractMinute() {
	suite.T().Logf("Testing ExtractMinute function for %s", suite.dbType)

	// Test 1: Extract minute from created_at column
	suite.Run("ExtractMinuteFromCreatedAt", func() {
		type MinuteResult struct {
			CreatedAt time.Time `bun:"created_at"`
			Minute    int64     `bun:"minute"`
		}

		var results []MinuteResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ExtractMinute(eb.Column("created_at"))
			}, "minute").
			OrderBy("created_at").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "ExtractMinute query should execute successfully")
		suite.True(len(results) > 0, "ExtractMinute should return at least one result")

		for _, result := range results {
			suite.Equal(int64(result.CreatedAt.Minute()), result.Minute, "Extracted minute should match the actual minute from created_at timestamp")
			suite.True(result.Minute >= 0 && result.Minute < 60, "Extracted minute should be in valid range (0-59)")
			suite.T().Logf("CreatedAt: %v | Minute: %d", result.CreatedAt, result.Minute)
		}
	})
}

// TestExtractSecond tests the ExtractSecond function.
// ExtractSecond extracts the second from a timestamp.
func (suite *DateTimeFunctionsTestSuite) TestExtractSecond() {
	suite.T().Logf("Testing ExtractSecond function for %s", suite.dbType)

	// Test 1: Extract second from created_at column
	suite.Run("ExtractSecondFromCreatedAt", func() {
		type SecondResult struct {
			Id        string    `bun:"id"`
			CreatedAt time.Time `bun:"created_at"`
			Second    float64   `bun:"second"`
		}

		var results []SecondResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ExtractSecond(eb.Column("created_at"))
			}, "second").
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "ExtractSecond query should execute successfully")
		suite.True(len(results) > 0, "ExtractSecond should return at least one result")

		for _, result := range results {
			suite.True(result.Second >= 0 && result.Second < 60, "Extracted second should be in valid range (0-59)")
			suite.T().Logf("ID: %s, CreatedAt: %s, Second: %.0f",
				result.Id, result.CreatedAt.Format(time.RFC3339), result.Second)
		}
	})

	// Test 2: Use all extract functions together (Year, Month, Day, Hour, Minute, Second)
	suite.Run("AllExtractFunctions", func() {
		type AllExtractResult struct {
			CreatedAt time.Time `bun:"created_at"`
			Year      int64     `bun:"year"`
			Month     int64     `bun:"month"`
			Day       int64     `bun:"day"`
			Hour      int64     `bun:"hour"`
			Minute    int64     `bun:"minute"`
			Second    float64   `bun:"second"`
		}

		var results []AllExtractResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ExtractYear(eb.Column("created_at"))
			}, "year").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ExtractMonth(eb.Column("created_at"))
			}, "month").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ExtractDay(eb.Column("created_at"))
			}, "day").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ExtractHour(eb.Column("created_at"))
			}, "hour").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ExtractMinute(eb.Column("created_at"))
			}, "minute").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ExtractSecond(eb.Column("created_at"))
			}, "second").
			OrderBy("created_at").
			Limit(3).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Combined extract functions query should execute successfully")
		suite.True(len(results) > 0, "Combined extract query should return at least one result")

		for _, result := range results {
			suite.Equal(int64(result.CreatedAt.Year()), result.Year, "Extracted year should match actual year")
			suite.Equal(int64(result.CreatedAt.Month()), result.Month, "Extracted month should match actual month")
			suite.Equal(int64(result.CreatedAt.Day()), result.Day, "Extracted day should match actual day")
			suite.Equal(int64(result.CreatedAt.Hour()), result.Hour, "Extracted hour should match actual hour")
			suite.Equal(int64(result.CreatedAt.Minute()), result.Minute, "Extracted minute should match actual minute")
			suite.True(result.Year >= 2000 && result.Year <= 3000, "Extracted year should be in reasonable range (2000-3000)")
			suite.True(result.Month >= 1 && result.Month <= 12, "Extracted month should be in valid range (1-12)")
			suite.True(result.Day >= 1 && result.Day <= 31, "Extracted day should be in valid range (1-31)")
			suite.True(result.Hour >= 0 && result.Hour < 24, "Extracted hour should be in valid range (0-23)")
			suite.True(result.Minute >= 0 && result.Minute < 60, "Extracted minute should be in valid range (0-59)")
			suite.True(result.Second >= 0 && result.Second < 60, "Extracted second should be in valid range (0-59)")

			suite.T().Logf("CreatedAt: %s, Y:%d M:%d D:%d H:%d M:%d S:%.0f",
				result.CreatedAt.Format(time.RFC3339),
				result.Year, result.Month, result.Day, result.Hour, result.Minute, result.Second)
		}
	})
}

// TestDateTrunc tests the DateTrunc function.
// DateTrunc truncates date/timestamp to specified precision.
func (suite *DateTimeFunctionsTestSuite) TestDateTrunc() {
	suite.T().Logf("Testing DateTrunc function for %s", suite.dbType)

	// Test 1: Truncate to different precisions (year, month, day, hour)
	suite.Run("TruncateToDifferentPrecisions", func() {
		type TruncResult struct {
			CreatedAt  time.Time `bun:"created_at"`
			TruncYear  time.Time `bun:"trunc_year"`
			TruncMonth time.Time `bun:"trunc_month"`
			TruncDay   time.Time `bun:"trunc_day"`
			TruncHour  time.Time `bun:"trunc_hour"`
		}

		var results []TruncResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DateTrunc("year", eb.Column("created_at"))
			}, "trunc_year").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DateTrunc("month", eb.Column("created_at"))
			}, "trunc_month").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DateTrunc("day", eb.Column("created_at"))
			}, "trunc_day").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DateTrunc("hour", eb.Column("created_at"))
			}, "trunc_hour").
			OrderBy("created_at").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "DateTrunc query should execute successfully")
		suite.True(len(results) > 0, "DateTrunc should return at least one result")

		for _, result := range results {
			// Verify truncated dates are on or before original
			suite.True(!result.TruncYear.After(result.CreatedAt), "Truncated year should not be after original timestamp")
			suite.True(!result.TruncMonth.After(result.CreatedAt), "Truncated month should not be after original timestamp")
			suite.True(!result.TruncDay.After(result.CreatedAt), "Truncated day should not be after original timestamp")
			suite.True(!result.TruncHour.After(result.CreatedAt), "Truncated hour should not be after original timestamp")

			suite.T().Logf("Original: %v | Year: %v | Month: %v | Day: %v | Hour: %v",
				result.CreatedAt, result.TruncYear, result.TruncMonth, result.TruncDay, result.TruncHour)
		}
	})
}

// TestDateAdd tests the DateAdd function.
// DateAdd adds interval to date/timestamp.
func (suite *DateTimeFunctionsTestSuite) TestDateAdd() {
	suite.T().Logf("Testing DateAdd function for %s", suite.dbType)

	// Test 1: Add different intervals (7 days, 3 months, 1 year)
	suite.Run("AddDifferentIntervals", func() {
		type DateAddResult struct {
			CreatedAt   time.Time `bun:"created_at"`
			AddedDays   time.Time `bun:"added_days"`
			AddedMonths time.Time `bun:"added_months"`
			AddedYears  time.Time `bun:"added_years"`
		}

		var results []DateAddResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DateAdd(eb.Column("created_at"), 7, "DAY")
			}, "added_days").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DateAdd(eb.Column("created_at"), 3, "MONTH")
			}, "added_months").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DateAdd(eb.Column("created_at"), 1, "YEAR")
			}, "added_years").
			OrderBy("created_at").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "DateAdd query should execute successfully")
		suite.True(len(results) > 0, "DateAdd should return at least one result")

		for _, result := range results {
			// Verify added dates are after original
			suite.True(result.AddedDays.After(result.CreatedAt), "Date with added days should be after original timestamp")
			suite.True(result.AddedMonths.After(result.CreatedAt), "Date with added months should be after original timestamp")
			suite.True(result.AddedYears.After(result.CreatedAt), "Date with added years should be after original timestamp")

			suite.T().Logf("Original: %v | +7 days: %v | +3 months: %v | +1 year: %v",
				result.CreatedAt, result.AddedDays, result.AddedMonths, result.AddedYears)
		}
	})
}

// TestDateSubtract tests the DateSubtract function.
// DateSubtract subtracts interval from date/timestamp.
func (suite *DateTimeFunctionsTestSuite) TestDateSubtract() {
	suite.T().Logf("Testing DateSubtract function for %s", suite.dbType)

	// Test 1: Subtract different intervals (5 days, 2 months)
	suite.Run("SubtractDifferentIntervals", func() {
		type DateSubtractResult struct {
			CreatedAt        time.Time `bun:"created_at"`
			SubtractedDays   time.Time `bun:"subtracted_days"`
			SubtractedMonths time.Time `bun:"subtracted_months"`
		}

		var results []DateSubtractResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DateSubtract(eb.Column("created_at"), 5, "DAY")
			}, "subtracted_days").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DateSubtract(eb.Column("created_at"), 2, "MONTH")
			}, "subtracted_months").
			OrderBy("created_at").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "DateSubtract query should execute successfully")
		suite.True(len(results) > 0, "DateSubtract should return at least one result")

		for _, result := range results {
			// Verify subtracted dates are before original
			suite.True(result.SubtractedDays.Before(result.CreatedAt), "Date with subtracted days should be before original timestamp")
			suite.True(result.SubtractedMonths.Before(result.CreatedAt), "Date with subtracted months should be before original timestamp")

			suite.T().Logf("Original: %v | -5 days: %v | -2 months: %v",
				result.CreatedAt, result.SubtractedDays, result.SubtractedMonths)
		}
	})
}

// TestDateDiff tests the DateDiff function.
// DateDiff returns the difference between two dates in specified unit.
func (suite *DateTimeFunctionsTestSuite) TestDateDiff() {
	suite.T().Logf("Testing DateDiff function for %s", suite.dbType)

	// Test 1: Calculate date differences in different units (days, hours, minutes)
	suite.Run("CalculateDateDifferences", func() {
		type DateDiffResult struct {
			CreatedAt   time.Time `bun:"created_at"`
			UpdatedAt   time.Time `bun:"updated_at"`
			DaysDiff    float64   `bun:"days_diff"`
			HoursDiff   float64   `bun:"hours_diff"`
			MinutesDiff float64   `bun:"minutes_diff"`
		}

		var results []DateDiffResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("created_at", "updated_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DateDiff(eb.Column("created_at"), eb.Column("updated_at"), "DAY")
			}, "days_diff").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DateDiff(eb.Column("created_at"), eb.Column("updated_at"), "HOUR")
			}, "hours_diff").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DateDiff(eb.Column("created_at"), eb.Column("updated_at"), "MINUTE")
			}, "minutes_diff").
			OrderBy("created_at").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "DateDiff query should execute successfully")
		suite.True(len(results) > 0, "DateDiff should return at least one result")

		for _, result := range results {
			suite.True(result.DaysDiff >= 0, "DateDiff in days should be non-negative (updated_at should not be before created_at)")
			suite.T().Logf("Created: %v | Updated: %v | Days: %.2f | Hours: %.2f | Minutes: %.2f",
				result.CreatedAt, result.UpdatedAt, result.DaysDiff, result.HoursDiff, result.MinutesDiff)
		}
	})
}

// TestAge tests the Age function.
// Age returns the age (interval) between two timestamps (PostgreSQL only).
func (suite *DateTimeFunctionsTestSuite) TestAge() {
	suite.T().Logf("Testing Age function for %s", suite.dbType)

	// Test 1: Calculate age interval between updated_at and created_at (PostgreSQL only)
	suite.Run("CalculateAgeInterval", func() {
		if suite.dbType != constants.DbPostgres {
			suite.T().Skipf("Age skipped for %s (PostgreSQL only)", suite.dbType)
		}

		type AgeResult struct {
			Id        string `bun:"id"`
			CreatedAt string `bun:"created_at"`
			UpdatedAt string `bun:"updated_at"`
			Age       string `bun:"age"`
		}

		var results []AgeResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "created_at", "updated_at").
			SelectExpr(func(eb ExprBuilder) any {
				// Calculate age between updated_at and created_at
				return eb.Age(eb.Column("updated_at"), eb.Column("created_at"))
			}, "age").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Age query should execute successfully on PostgreSQL")
		suite.True(len(results) > 0, "Age should return at least one result on PostgreSQL")

		for _, result := range results {
			suite.NotEmpty(result.Age, "Age interval should not be empty (should return PostgreSQL interval format)")
			suite.T().Logf("ID: %s, CreatedAt: %s, UpdatedAt: %s, Age: %s",
				result.Id, result.CreatedAt, result.UpdatedAt, result.Age)
		}
	})
}

// TestCombinedDateTimeFunctions tests using multiple date/time functions together.
// This verifies that date/time functions can be nested and combined.
func (suite *DateTimeFunctionsTestSuite) TestCombinedDateTimeFunctions() {
	suite.T().Logf("Testing combined date/time functions for %s", suite.dbType)

	// Test 1: Nest and combine multiple date/time functions (Extract, DateDiff, DateTrunc, Now)
	suite.Run("NestedDateTimeFunctions", func() {
		type CombinedResult struct {
			CreatedAt     time.Time `bun:"created_at"`
			Year          int64     `bun:"year"`
			MonthsFromNow float64   `bun:"months_from_now"`
			FormattedDate time.Time `bun:"formatted_date"`
		}

		var results []CombinedResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("created_at").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ExtractYear(eb.Column("created_at"))
			}, "year").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DateDiff(
					eb.DateTrunc("month", eb.Column("created_at")),
					eb.DateTrunc("month", eb.Now()),
					"MONTH",
				)
			}, "months_from_now").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DateTrunc("day", eb.Column("created_at"))
			}, "formatted_date").
			OrderBy("created_at").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Combined date/time functions query should execute successfully")
		suite.True(len(results) > 0, "Combined date/time functions should return at least one result")

		for _, result := range results {
			suite.True(result.Year > 2000, "Extracted year should be in reasonable range (after 2000)")
			suite.True(result.MonthsFromNow >= 0, "Calculated months from now should be non-negative (created_at should not be in the future)")
			suite.True(!result.FormattedDate.After(result.CreatedAt), "Truncated date should not be after original timestamp")

			suite.T().Logf("CreatedAt: %v | Year: %d | MonthsFromNow: %.0f | FormattedDate: %v",
				result.CreatedAt, result.Year, result.MonthsFromNow, result.FormattedDate)
		}
	})
}
