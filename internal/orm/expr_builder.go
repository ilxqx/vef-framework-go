package orm

import (
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/schema"
)

// QueryExprBuilder implements the ExprBuilder interface, providing methods to build various SQL expressions.
// It maintains references to the table schema and subquery builder for proper context.
type QueryExprBuilder struct {
	qb QueryBuilder
}

// Column builds a column expression with proper table alias handling.
// It supports both simple column names ("id") and qualified names ("u.id").
// For simple names, it automatically adds the table alias prefix.
func (b *QueryExprBuilder) Column(column string) schema.QueryAppender {
	dotIndex := strings.IndexByte(column, constants.ByteDot)
	if dotIndex > -1 {
		alias, name := column[:dotIndex], column[dotIndex+1:]
		if strings.IndexByte(alias, constants.ByteQuestionMark) > -1 {
			return b.Expr(alias+".?", bun.Name(name))
		} else {
			return b.Expr("?.?", bun.Name(alias), bun.Name(name))
		}
	} else {
		return b.Expr("?TableAlias.?", bun.Name(column))
	}
}

// Null returns the SQL NULL literal value.
func (*QueryExprBuilder) Null() schema.QueryAppender {
	return bun.Safe(sqlNull)
}

// IsNull checks if an expression is NULL.
func (b *QueryExprBuilder) IsNull(expr any) schema.QueryAppender {
	return b.Expr("? IS NULL", expr)
}

// IsNotNull checks if an expression is not NULL.
func (b *QueryExprBuilder) IsNotNull(expr any) schema.QueryAppender {
	return b.Expr("? IS NOT NULL", expr)
}

// Literal builds a literal expression.
func (b *QueryExprBuilder) Literal(value any) schema.QueryAppender {
	return b.Expr("?", value)
}

// Order builds an ORDER BY expression.
func (b *QueryExprBuilder) Order(builder func(OrderBuilder)) schema.QueryAppender {
	ob := newOrderExpr(b)
	builder(ob)

	return ob
}

// Case creates a CASE expression builder with access to the current table context.
// The builder function allows configuring the CASE expression with WHEN/THEN/ELSE clauses.
func (b *QueryExprBuilder) Case(builder func(CaseBuilder)) schema.QueryAppender {
	cb := newCaseExpr(b.qb)
	builder(cb)

	return cb
}

// Expr creates an expression builder for complex SQL logic.
func (b *QueryExprBuilder) Expr(expr string, args ...any) schema.QueryAppender {
	return bun.SafeQuery(expr, args...)
}

// Exprs creates an expression builder for complex SQL logic.
func (b *QueryExprBuilder) Exprs(exprs ...any) schema.QueryAppender {
	return newExpressions(constants.CommaSpace, exprs...)
}

// ExprsWS creates an expression builder for complex SQL logic with a separator.
func (b *QueryExprBuilder) ExprsWS(sep string, exprs ...any) schema.QueryAppender {
	return newExpressions(sep, exprs...)
}

// ExprByDialect creates a cross-database compatible expression.
// It selects the appropriate expression builder based on the current database dialect.
func (b *QueryExprBuilder) ExprByDialect(expr DialectExpr) schema.QueryAppender {
	switch b.qb.Dialect().Name() {
	case dialect.Oracle:
		if expr.Oracle != nil {
			return expr.Oracle()
		}
	case dialect.MSSQL:
		if expr.SQLServer != nil {
			return expr.SQLServer()
		}
	case dialect.PG:
		if expr.Postgres != nil {
			return expr.Postgres()
		}
	case dialect.MySQL:
		if expr.MySQL != nil {
			return expr.MySQL()
		}
	case dialect.SQLite:
		if expr.SQLite != nil {
			return expr.SQLite()
		}
	}

	// Fallback to default if database-specific builder is not available
	if expr.Default != nil {
		return expr.Default()
	}

	// Return NULL if no suitable builder is found
	return b.Null()
}

// RunDialect executes database-specific side-effect callbacks based on the current dialect.
func (b *QueryExprBuilder) RunDialect(actions DialectActions) {
	switch b.qb.Dialect().Name() {
	case dialect.Oracle:
		if actions.Oracle != nil {
			actions.Oracle()
			return
		}
	case dialect.MSSQL:
		if actions.SQLServer != nil {
			actions.SQLServer()
			return
		}
	case dialect.PG:
		if actions.Postgres != nil {
			actions.Postgres()
			return
		}
	case dialect.MySQL:
		if actions.MySQL != nil {
			actions.MySQL()
			return
		}
	case dialect.SQLite:
		if actions.SQLite != nil {
			actions.SQLite()
			return
		}
	}

	if actions.Default != nil {
		actions.Default()
	}
}

// RunDialectErr executes database-specific callbacks that can return an error.
func (b *QueryExprBuilder) RunDialectErr(actions DialectActionsErr) error {
	switch b.qb.Dialect().Name() {
	case dialect.Oracle:
		if actions.Oracle != nil {
			return actions.Oracle()
		}
	case dialect.MSSQL:
		if actions.SQLServer != nil {
			return actions.SQLServer()
		}
	case dialect.PG:
		if actions.Postgres != nil {
			return actions.Postgres()
		}
	case dialect.MySQL:
		if actions.MySQL != nil {
			return actions.MySQL()
		}
	case dialect.SQLite:
		if actions.SQLite != nil {
			return actions.SQLite()
		}
	}

	if actions.Default != nil {
		return actions.Default()
	}

	return ErrDialectHandlerMissing
}

// RunDialectFunc executes database-specific callbacks that return query fragments.
func (b *QueryExprBuilder) RunDialectFunc(funcs DialectFuncs) ([]byte, error) {
	switch b.qb.Dialect().Name() {
	case dialect.Oracle:
		if funcs.Oracle != nil {
			return funcs.Oracle()
		}
	case dialect.MSSQL:
		if funcs.SQLServer != nil {
			return funcs.SQLServer()
		}
	case dialect.PG:
		if funcs.Postgres != nil {
			return funcs.Postgres()
		}
	case dialect.MySQL:
		if funcs.MySQL != nil {
			return funcs.MySQL()
		}
	case dialect.SQLite:
		if funcs.SQLite != nil {
			return funcs.SQLite()
		}
	}

	if funcs.Default != nil {
		return funcs.Default()
	}

	return nil, ErrDialectHandlerMissing
}

// ========== Aggregate Functions ==========

// Count builds a COUNT aggregate expression using a builder callback.
func (b *QueryExprBuilder) Count(builder func(CountBuilder)) schema.QueryAppender {
	cb := newCountExpr(b.qb)
	builder(cb)

	return cb
}

// CountColumn builds a COUNT(column) aggregate expression.
func (b *QueryExprBuilder) CountColumn(column string, distinct ...bool) schema.QueryAppender {
	return b.Count(func(cb CountBuilder) {
		if len(distinct) > 0 && distinct[0] {
			cb.Distinct()
		}

		cb.Column(column)
	})
}

// CountAll builds a COUNT(*) aggregate expression.
func (b *QueryExprBuilder) CountAll(distinct ...bool) schema.QueryAppender {
	return b.Count(func(cb CountBuilder) {
		if len(distinct) > 0 && distinct[0] {
			cb.Distinct()
		}

		cb.All()
	})
}

// Sum builds a SUM aggregate expression using a builder callback.
func (b *QueryExprBuilder) Sum(builder func(SumBuilder)) schema.QueryAppender {
	cb := newSumExpr(b.qb)
	builder(cb)

	return cb
}

// SumColumn builds a SUM(column) aggregate expression.
func (b *QueryExprBuilder) SumColumn(column string, distinct ...bool) schema.QueryAppender {
	return b.Sum(func(cb SumBuilder) {
		if len(distinct) > 0 && distinct[0] {
			cb.Distinct()
		}

		cb.Column(column)
	})
}

// Avg builds an AVG aggregate expression using a builder callback.
func (b *QueryExprBuilder) Avg(builder func(AvgBuilder)) schema.QueryAppender {
	cb := newAvgExpr(b.qb)
	builder(cb)

	return cb
}

// AvgColumn builds an AVG(column) aggregate expression.
func (b *QueryExprBuilder) AvgColumn(column string, distinct ...bool) schema.QueryAppender {
	return b.Avg(func(cb AvgBuilder) {
		if len(distinct) > 0 && distinct[0] {
			cb.Distinct()
		}

		cb.Column(column)
	})
}

// Min builds a MIN aggregate expression using a builder callback.
func (b *QueryExprBuilder) Min(builder func(MinBuilder)) schema.QueryAppender {
	cb := newMinExpr(b.qb)
	builder(cb)

	return cb
}

// MinColumn builds a MIN(column) aggregate expression.
func (b *QueryExprBuilder) MinColumn(column string) schema.QueryAppender {
	return b.Min(func(cb MinBuilder) {
		cb.Column(column)
	})
}

// Max builds a MAX aggregate expression using a builder callback.
func (b *QueryExprBuilder) Max(builder func(MaxBuilder)) schema.QueryAppender {
	cb := newMaxExpr(b.qb)
	builder(cb)

	return cb
}

// MaxColumn builds a MAX(column) aggregate expression.
func (b *QueryExprBuilder) MaxColumn(column string) schema.QueryAppender {
	return b.Max(func(cb MaxBuilder) {
		cb.Column(column)
	})
}

func (b *QueryExprBuilder) StringAgg(builder func(StringAggBuilder)) schema.QueryAppender {
	cb := newStringAggExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) ArrayAgg(builder func(ArrayAggBuilder)) schema.QueryAppender {
	cb := newArrayAggExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) StdDev(builder func(StdDevBuilder)) schema.QueryAppender {
	cb := newStdDevExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) Variance(builder func(VarianceBuilder)) schema.QueryAppender {
	cb := newVarianceExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) JSONObjectAgg(builder func(JSONObjectAggBuilder)) schema.QueryAppender {
	cb := newJsonObjectAggExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) JSONArrayAgg(builder func(JSONArrayAggBuilder)) schema.QueryAppender {
	cb := newJsonArrayAggExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) BitOr(builder func(BitOrBuilder)) schema.QueryAppender {
	cb := newBitOrExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) BitAnd(builder func(BitAndBuilder)) schema.QueryAppender {
	cb := newBitAndExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) BoolOr(builder func(BoolOrBuilder)) schema.QueryAppender {
	cb := newBoolOrExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) BoolAnd(builder func(BoolAndBuilder)) schema.QueryAppender {
	cb := newBoolAndExpr(b.qb)
	builder(cb)

	return cb
}

// ========== Window Functions ==========

func (b *QueryExprBuilder) RowNumber(builder func(RowNumberBuilder)) schema.QueryAppender {
	cb := newRowNumberExpr(b)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) Rank(builder func(RankBuilder)) schema.QueryAppender {
	cb := newRankExpr(b)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) DenseRank(builder func(DenseRankBuilder)) schema.QueryAppender {
	cb := newDenseRankExpr(b)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) PercentRank(builder func(PercentRankBuilder)) schema.QueryAppender {
	cb := newPercentRankExpr(b)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) CumeDist(builder func(CumeDistBuilder)) schema.QueryAppender {
	cb := newCumeDistExpr(b)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) Ntile(builder func(NtileBuilder)) schema.QueryAppender {
	cb := newNtileExpr(b)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) Lag(builder func(LagBuilder)) schema.QueryAppender {
	cb := newLagExpr(b)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) Lead(builder func(LeadBuilder)) schema.QueryAppender {
	cb := newLeadExpr(b)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) FirstValue(builder func(FirstValueBuilder)) schema.QueryAppender {
	cb := newFirstValueExpr(b)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) LastValue(builder func(LastValueBuilder)) schema.QueryAppender {
	cb := newLastValueExpr(b)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) NthValue(builder func(NthValueBuilder)) schema.QueryAppender {
	cb := newNthValueExpr(b)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WCount(builder func(WindowCountBuilder)) schema.QueryAppender {
	cb := newWindowCountExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WSum(builder func(WindowSumBuilder)) schema.QueryAppender {
	cb := newWindowSumExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WAvg(builder func(WindowAvgBuilder)) schema.QueryAppender {
	cb := newWindowAvgExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WMin(builder func(WindowMinBuilder)) schema.QueryAppender {
	cb := newWindowMinExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WMax(builder func(WindowMaxBuilder)) schema.QueryAppender {
	cb := newWindowMaxExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WStringAgg(builder func(WindowStringAggBuilder)) schema.QueryAppender {
	cb := newWindowStringAggExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WArrayAgg(builder func(WindowArrayAggBuilder)) schema.QueryAppender {
	cb := newWindowArrayAggExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WStdDev(builder func(WindowStdDevBuilder)) schema.QueryAppender {
	cb := newWindowStdDevExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WVariance(builder func(WindowVarianceBuilder)) schema.QueryAppender {
	cb := newWindowVarianceExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WJSONObjectAgg(builder func(WindowJSONObjectAggBuilder)) schema.QueryAppender {
	cb := newWindowJsonObjectAggExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WJSONArrayAgg(builder func(WindowJSONArrayAggBuilder)) schema.QueryAppender {
	cb := newWindowJsonArrayAggExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WBitOr(builder func(WindowBitOrBuilder)) schema.QueryAppender {
	cb := newWindowBitOrExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WBitAnd(builder func(WindowBitAndBuilder)) schema.QueryAppender {
	cb := newWindowBitAndExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WBoolOr(builder func(WindowBoolOrBuilder)) schema.QueryAppender {
	cb := newWindowBoolOrExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WBoolAnd(builder func(WindowBoolAndBuilder)) schema.QueryAppender {
	cb := newWindowBoolAndExpr(b.qb)
	builder(cb)

	return cb
}

// ========== String Functions ==========

// Concat concatenates strings.
func (b *QueryExprBuilder) Concat(args ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		SQLite: func() schema.QueryAppender {
			// SQLite uses || operator for string concatenation
			if len(args) == 0 {
				return b.Expr("?", constants.Empty)
			}
			if len(args) == 1 {
				return b.Expr("?", args[0])
			}

			return b.ExprsWS(" || ", args...)
		},
		Default: func() schema.QueryAppender {
			// PostgreSQL and MySQL support CONCAT function
			return b.Expr("CONCAT(?)", newExpressions(constants.CommaSpace, args...))
		},
	})
}

// ConcatWS concatenates strings with a separator.
func (b *QueryExprBuilder) ConcatWS(separator string, args ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have CONCAT_WS, use string concatenation with separator
			if len(args) == 0 {
				return b.Expr("?", constants.Empty)
			}
			if len(args) == 1 {
				return b.Expr("?", args[0])
			}

			// Use group_concat or manual concatenation
			var parts []any
			for i, arg := range args {
				if i > 0 {
					parts = append(parts, separator)
				}
				parts = append(parts, arg)
			}

			return b.Concat(parts...)
		},
		Default: func() schema.QueryAppender {
			// PostgreSQL and MySQL support CONCAT_WS
			return b.Expr("CONCAT_WS(?, ?)", separator, newExpressions(constants.CommaSpace, args...))
		},
	})
}

// SubString extracts a substring from a string.
func (b *QueryExprBuilder) SubString(expr any, start int, length ...int) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		SQLite: func() schema.QueryAppender {
			// SQLite uses SUBSTR function
			if len(length) > 0 {
				return b.Expr("SUBSTR(?, ?, ?)", expr, start, length[0])
			}
			return b.Expr("SUBSTR(?, ?)", expr, start)
		},
		Default: func() schema.QueryAppender {
			// PostgreSQL and MySQL support standard SUBSTRING
			if len(length) > 0 {
				return b.Expr("SUBSTRING(?, ?, ?)", expr, start, length[0])
			}
			return b.Expr("SUBSTRING(?, ?)", expr, start)
		},
	})
}

// Upper converts string to uppercase.
func (b *QueryExprBuilder) Upper(expr any) schema.QueryAppender {
	return b.Expr("UPPER(?)", expr)
}

// Lower converts string to lowercase.
func (b *QueryExprBuilder) Lower(expr any) schema.QueryAppender {
	return b.Expr("LOWER(?)", expr)
}

// Trim removes leading and trailing whitespace.
func (b *QueryExprBuilder) Trim(expr any) schema.QueryAppender {
	return b.Expr("TRIM(?)", expr)
}

// TrimLeft removes leading whitespace.
func (b *QueryExprBuilder) TrimLeft(expr any) schema.QueryAppender {
	return b.Expr("LTRIM(?)", expr)
}

// TrimRight removes trailing whitespace.
func (b *QueryExprBuilder) TrimRight(expr any) schema.QueryAppender {
	return b.Expr("RTRIM(?)", expr)
}

// Length returns the length of a string.
func (b *QueryExprBuilder) Length(expr any) schema.QueryAppender {
	return b.Expr("LENGTH(?)", expr)
}

// CharLength returns the character length of a string.
func (b *QueryExprBuilder) CharLength(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have CHAR_LENGTH, use LENGTH
			return b.Expr("LENGTH(?)", expr)
		},
		Default: func() schema.QueryAppender {
			// PostgreSQL and MySQL support CHAR_LENGTH
			return b.Expr("CHAR_LENGTH(?)", expr)
		},
	})
}

// Position finds the position of substring in string.
func (b *QueryExprBuilder) Position(substring, str any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have POSITION IN syntax, use INSTR
			return b.Expr("INSTR(?, ?)", str, substring)
		},
		Default: func() schema.QueryAppender {
			// PostgreSQL and MySQL support POSITION IN
			return b.Expr("POSITION(? IN ?)", substring, str)
		},
	})
}

// Left returns the leftmost n characters.
func (b *QueryExprBuilder) Left(expr any, length int) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have LEFT, use SUBSTR
			return b.Expr("SUBSTR(?, 1, ?)", expr, length)
		},
		Default: func() schema.QueryAppender {
			// PostgreSQL and MySQL support LEFT
			return b.Expr("LEFT(?, ?)", expr, length)
		},
	})
}

// Right returns the rightmost n characters.
func (b *QueryExprBuilder) Right(expr any, length int) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have RIGHT, use SUBSTR with negative start
			return b.Expr("SUBSTR(?, -?)", expr, length)
		},
		Default: func() schema.QueryAppender {
			// PostgreSQL and MySQL support RIGHT
			return b.Expr("RIGHT(?, ?)", expr, length)
		},
	})
}

// Repeat repeats a string n times.
func (b *QueryExprBuilder) Repeat(expr any, count int) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have REPEAT, need to implement with REPLACE and a helper
			// Use REPLACE(SUBSTR(QUOTE(ZEROBLOB((count+1)/2)), 3, count), '0', expr)
			return b.Expr("REPLACE(SUBSTR(QUOTE(ZEROBLOB((?+1)/2)), 3, ?), '0', ?)", count, count, expr)
		},
		Default: func() schema.QueryAppender {
			// PostgreSQL and MySQL support REPEAT
			return b.Expr("REPEAT(?, ?)", expr, count)
		},
	})
}

// Replace replaces all occurrences of substring with replacement.
func (b *QueryExprBuilder) Replace(expr, search, replacement any) schema.QueryAppender {
	// REPLACE is supported by all databases with same syntax
	return b.Expr("REPLACE(?, ?, ?)", expr, search, replacement)
}

// Reverse reverses a string.
func (b *QueryExprBuilder) Reverse(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have REVERSE function, need complex implementation
			// For now, return unsupported
			return b.Expr("? /* REVERSE not supported in SQLite */", expr)
		},
		Default: func() schema.QueryAppender {
			// PostgreSQL and MySQL support REVERSE
			return b.Expr("REVERSE(?)", expr)
		},
	})
}

// ========== Date and Time Functions ==========

// CurrentDate returns the current date.
func (b *QueryExprBuilder) CurrentDate() schema.QueryAppender {
	return bun.Safe("CURRENT_DATE")
}

// CurrentTime returns the current time.
func (b *QueryExprBuilder) CurrentTime() schema.QueryAppender {
	return bun.Safe("CURRENT_TIME")
}

// CurrentTimestamp returns the current timestamp.
func (b *QueryExprBuilder) CurrentTimestamp() schema.QueryAppender {
	return bun.Safe("CURRENT_TIMESTAMP")
}

// Now returns the current timestamp.
func (b *QueryExprBuilder) Now() schema.QueryAppender {
	return bun.Safe("NOW()")
}

// ExtractYear extracts the year from a date/timestamp.
func (b *QueryExprBuilder) ExtractYear(expr any) schema.QueryAppender {
	return b.Expr("EXTRACT(YEAR FROM ?)", expr)
}

// ExtractMonth extracts the month from a date/timestamp.
func (b *QueryExprBuilder) ExtractMonth(expr any) schema.QueryAppender {
	return b.Expr("EXTRACT(MONTH FROM ?)", expr)
}

// ExtractDay extracts the day from a date/timestamp.
func (b *QueryExprBuilder) ExtractDay(expr any) schema.QueryAppender {
	return b.Expr("EXTRACT(DAY FROM ?)", expr)
}

// ExtractHour extracts the hour from a timestamp.
func (b *QueryExprBuilder) ExtractHour(expr any) schema.QueryAppender {
	return b.Expr("EXTRACT(HOUR FROM ?)", expr)
}

// ExtractMinute extracts the minute from a timestamp.
func (b *QueryExprBuilder) ExtractMinute(expr any) schema.QueryAppender {
	return b.Expr("EXTRACT(MINUTE FROM ?)", expr)
}

// ExtractSecond extracts the second from a timestamp.
func (b *QueryExprBuilder) ExtractSecond(expr any) schema.QueryAppender {
	return b.Expr("EXTRACT(SECOND FROM ?)", expr)
}

// DateTrunc truncates date/timestamp to specified precision.
func (b *QueryExprBuilder) DateTrunc(precision string, expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL has native DATE_TRUNC
			return b.Expr("DATE_TRUNC(?, ?)", precision, expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL needs different approach based on precision
			switch precision {
			case "year":
				return b.Expr("DATE_FORMAT(?, '%Y-01-01')", expr)
			case "month":
				return b.Expr("DATE_FORMAT(?, '%Y-%m-01')", expr)
			case "day":
				return b.Expr("DATE(?)", expr)
			case "hour":
				return b.Expr("DATE_FORMAT(?, '%Y-%m-%d %H:00:00')", expr)
			case "minute":
				return b.Expr("DATE_FORMAT(?, '%Y-%m-%d %H:%i:00')", expr)
			default:
				return b.Expr("DATE(?)", expr)
			}
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses strftime for truncation
			switch precision {
			case "year":
				return b.Expr("STRFTIME('%Y-01-01', ?)", expr)
			case "month":
				return b.Expr("STRFTIME('%Y-%m-01', ?)", expr)
			case "day":
				return b.Expr("DATE(?)", expr)
			case "hour":
				return b.Expr("STRFTIME('%Y-%m-%d %H:00:00', ?)", expr)
			case "minute":
				return b.Expr("STRFTIME('%Y-%m-%d %H:%M:00', ?)", expr)
			default:
				return b.Expr("DATE(?)", expr)
			}
		},
		Default: func() schema.QueryAppender {
			return b.Expr("DATE_TRUNC(?, ?)", precision, expr)
		},
	})
}

// DateAdd adds interval to date/timestamp.
func (b *QueryExprBuilder) DateAdd(expr any, interval int, unit string) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses INTERVAL syntax
			return b.Expr("? + INTERVAL '? ?'", expr, interval, bun.Safe(unit))
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses DATE_ADD with INTERVAL
			return b.Expr("DATE_ADD(?, INTERVAL ? ?)", expr, interval, bun.Safe(unit))
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses datetime with modifiers
			var modifier string
			switch unit {
			case "year", "years":
				modifier = "years"
			case "month", "months":
				modifier = "months"
			case "day", "days":
				modifier = "days"
			case "hour", "hours":
				modifier = "hours"
			case "minute", "minutes":
				modifier = "minutes"
			case "second", "seconds":
				modifier = "seconds"
			default:
				modifier = "days"
			}
			return b.Expr("DATETIME(?, '+? ?')", expr, interval, bun.Safe(modifier))
		},
		Default: func() schema.QueryAppender {
			return b.Expr("DATE_ADD(?, INTERVAL ? ?)", expr, interval, bun.Safe(unit))
		},
	})
}

// DateSubtract subtracts interval from date/timestamp.
func (b *QueryExprBuilder) DateSubtract(expr any, interval int, unit string) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses INTERVAL syntax
			return b.Expr("? - INTERVAL '? ?'", expr, interval, bun.Safe(unit))
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses DATE_SUB with INTERVAL
			return b.Expr("DATE_SUB(?, INTERVAL ? ?)", expr, interval, bun.Safe(unit))
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses datetime with negative modifiers
			var modifier string
			switch unit {
			case "year", "years":
				modifier = "years"
			case "month", "months":
				modifier = "months"
			case "day", "days":
				modifier = "days"
			case "hour", "hours":
				modifier = "hours"
			case "minute", "minutes":
				modifier = "minutes"
			case "second", "seconds":
				modifier = "seconds"
			default:
				modifier = "days"
			}
			return b.Expr("DATETIME(?, '-? ?')", expr, interval, bun.Safe(modifier))
		},
		Default: func() schema.QueryAppender {
			return b.Expr("DATE_SUB(?, INTERVAL ? ?)", expr, interval, bun.Safe(unit))
		},
	})
}

// DateDiff returns the difference between two dates in specified unit.
func (b *QueryExprBuilder) DateDiff(start, end any, unit string) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses EXTRACT with subtraction
			switch unit {
			case "day", "days":
				return b.Expr("(? - ?)::DATE", end, start)
			case "year", "years":
				return b.Expr("EXTRACT(YEAR FROM ?) - EXTRACT(YEAR FROM ?)", end, start)
			case "month", "months":
				return b.Expr("EXTRACT(YEAR FROM ?) * 12 + EXTRACT(MONTH FROM ?) - EXTRACT(YEAR FROM ?) * 12 - EXTRACT(MONTH FROM ?)", end, end, start, start)
			default:
				return b.Expr("EXTRACT(DAYS FROM ? - ?)", end, start)
			}
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has DATEDIFF for days, TIMESTAMPDIFF for other units
			switch unit {
			case "day", "days":
				return b.Expr("DATEDIFF(?, ?)", end, start)
			default:
				return b.Expr("TIMESTAMPDIFF(?, ?, ?)", bun.Safe(unit), start, end)
			}
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses julianday for date differences
			switch unit {
			case "day", "days":
				return b.Expr("JULIANDAY(?) - JULIANDAY(?)", end, start)
			case "year", "years":
				return b.Expr("(STRFTIME('%Y', ?) - STRFTIME('%Y', ?))", end, start)
			case "month", "months":
				return b.Expr("(STRFTIME('%Y', ?) - STRFTIME('%Y', ?)) * 12 + (STRFTIME('%m', ?) - STRFTIME('%m', ?))", end, start, end, start)
			default:
				return b.Expr("JULIANDAY(?) - JULIANDAY(?)", end, start)
			}
		},
		Default: func() schema.QueryAppender {
			return b.Expr("DATEDIFF(?, ?, ?)", bun.Safe(unit), end, start)
		},
	})
}

// Age returns the age (interval) between two timestamps.
func (b *QueryExprBuilder) Age(start, end any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL has native AGE function
			return b.Expr("AGE(?, ?)", end, start)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL doesn't have AGE, calculate years difference
			return b.Expr("TIMESTAMPDIFF(YEAR, ?, ?)", start, end)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have AGE, calculate years difference
			return b.Expr("STRFTIME('%Y', ?) - STRFTIME('%Y', ?)", end, start)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("AGE(?, ?)", end, start)
		},
	})
}

// ========== Math Functions ==========

// Abs returns the absolute value.
func (b *QueryExprBuilder) Abs(expr any) schema.QueryAppender {
	return b.Expr("ABS(?)", expr)
}

// Ceil returns the smallest integer greater than or equal to the value.
func (b *QueryExprBuilder) Ceil(expr any) schema.QueryAppender {
	return b.Expr("CEIL(?)", expr)
}

// Floor returns the largest integer less than or equal to the value.
func (b *QueryExprBuilder) Floor(expr any) schema.QueryAppender {
	return b.Expr("FLOOR(?)", expr)
}

// Round rounds to the nearest integer or specified decimal places.
func (b *QueryExprBuilder) Round(expr any, precision ...int) schema.QueryAppender {
	if len(precision) > 0 {
		return b.Expr("ROUND(?, ?)", expr, precision[0])
	}
	return b.Expr("ROUND(?)", expr)
}

// Trunc truncates to integer or specified decimal places.
func (b *QueryExprBuilder) Trunc(expr any, precision ...int) schema.QueryAppender {
	if len(precision) > 0 {
		return b.Expr("TRUNC(?, ?)", expr, precision[0])
	}
	return b.Expr("TRUNC(?)", expr)
}

// Power returns base raised to the power of exponent.
func (b *QueryExprBuilder) Power(base, exponent any) schema.QueryAppender {
	return b.Expr("POWER(?, ?)", base, exponent)
}

// Sqrt returns the square root.
func (b *QueryExprBuilder) Sqrt(expr any) schema.QueryAppender {
	return b.Expr("SQRT(?)", expr)
}

// Exp returns e raised to the power of the argument.
func (b *QueryExprBuilder) Exp(expr any) schema.QueryAppender {
	return b.Expr("EXP(?)", expr)
}

// Ln returns the natural logarithm.
func (b *QueryExprBuilder) Ln(expr any) schema.QueryAppender {
	return b.Expr("LN(?)", expr)
}

// Log returns the logarithm with specified base.
func (b *QueryExprBuilder) Log(expr any, base ...any) schema.QueryAppender {
	if len(base) > 0 {
		return b.Expr("LOG(?, ?)", base[0], expr)
	}
	return b.Expr("LOG(?)", expr)
}

// Sin returns the sine.
func (b *QueryExprBuilder) Sin(expr any) schema.QueryAppender {
	return b.Expr("SIN(?)", expr)
}

// Cos returns the cosine.
func (b *QueryExprBuilder) Cos(expr any) schema.QueryAppender {
	return b.Expr("COS(?)", expr)
}

// Tan returns the tangent.
func (b *QueryExprBuilder) Tan(expr any) schema.QueryAppender {
	return b.Expr("TAN(?)", expr)
}

// Asin returns the arcsine.
func (b *QueryExprBuilder) Asin(expr any) schema.QueryAppender {
	return b.Expr("ASIN(?)", expr)
}

// Acos returns the arccosine.
func (b *QueryExprBuilder) Acos(expr any) schema.QueryAppender {
	return b.Expr("ACOS(?)", expr)
}

// Atan returns the arctangent.
func (b *QueryExprBuilder) Atan(expr any) schema.QueryAppender {
	return b.Expr("ATAN(?)", expr)
}

// Pi returns the value of π.
func (b *QueryExprBuilder) Pi() schema.QueryAppender {
	return bun.Safe("PI()")
}

// Random returns a random value between 0 and 1.
func (b *QueryExprBuilder) Random() schema.QueryAppender {
	return bun.Safe("RANDOM()")
}

// Sign returns the sign of a number.
func (b *QueryExprBuilder) Sign(expr any) schema.QueryAppender {
	return b.Expr("SIGN(?)", expr)
}

// Mod returns the remainder of division.
func (b *QueryExprBuilder) Mod(dividend, divisor any) schema.QueryAppender {
	return b.Expr("MOD(?, ?)", dividend, divisor)
}

// Greatest returns the greatest value among arguments.
func (b *QueryExprBuilder) Greatest(args ...any) schema.QueryAppender {
	return b.Expr("GREATEST(?)", newExpressions(constants.CommaSpace, args...))
}

// Least returns the least value among arguments.
func (b *QueryExprBuilder) Least(args ...any) schema.QueryAppender {
	return b.Expr("LEAST(?)", newExpressions(constants.CommaSpace, args...))
}

// ========== Conditional Functions ==========

// Coalesce returns the first non-null value.
func (b *QueryExprBuilder) Coalesce(args ...any) schema.QueryAppender {
	return b.Expr("COALESCE(?)", newExpressions(constants.CommaSpace, args...))
}

// NullIf returns null if the two arguments are equal, otherwise returns the first argument.
func (b *QueryExprBuilder) NullIf(expr1, expr2 any) schema.QueryAppender {
	return b.Expr("NULLIF(?, ?)", expr1, expr2)
}

// IfNull returns the second argument if the first is null, otherwise returns the first.
func (b *QueryExprBuilder) IfNull(expr, defaultValue any) schema.QueryAppender {
	return b.Expr("IFNULL(?, ?)", expr, defaultValue)
}

// ========== Type Conversion Functions ==========

// ToString converts expression to string.
func (b *QueryExprBuilder) ToString(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses TEXT or VARCHAR
			return b.Expr("?::TEXT", expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses CHAR or VARCHAR
			return b.Expr("CAST(? AS CHAR)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have strict types, but TEXT is preferred
			return b.Expr("CAST(? AS TEXT)", expr)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("CAST(? AS VARCHAR)", expr)
		},
	})
}

// ToInteger converts expression to integer.
func (b *QueryExprBuilder) ToInteger(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses INTEGER or INT
			return b.Expr("?::INTEGER", expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses SIGNED INTEGER
			return b.Expr("CAST(? AS SIGNED INTEGER)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses INTEGER
			return b.Expr("CAST(? AS INTEGER)", expr)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("CAST(? AS INTEGER)", expr)
		},
	})
}

// ToDecimal converts expression to decimal with optional precision and scale.
func (b *QueryExprBuilder) ToDecimal(expr any, precision ...int) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses NUMERIC
			if len(precision) >= 2 {
				return b.Expr("?::NUMERIC(?, ?)", expr, precision[0], precision[1])
			} else if len(precision) == 1 {
				return b.Expr("?::NUMERIC(?)", expr, precision[0])
			}
			return b.Expr("?::NUMERIC", expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses DECIMAL
			if len(precision) >= 2 {
				return b.Expr("CAST(? AS DECIMAL(?, ?))", expr, precision[0], precision[1])
			} else if len(precision) == 1 {
				return b.Expr("CAST(? AS DECIMAL(?))", expr, precision[0])
			}
			return b.Expr("CAST(? AS DECIMAL)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses REAL for decimal numbers
			return b.Expr("CAST(? AS REAL)", expr)
		},
		Default: func() schema.QueryAppender {
			if len(precision) >= 2 {
				return b.Expr("CAST(? AS DECIMAL(?, ?))", expr, precision[0], precision[1])
			} else if len(precision) == 1 {
				return b.Expr("CAST(? AS DECIMAL(?))", expr, precision[0])
			}
			return b.Expr("CAST(? AS DECIMAL)", expr)
		},
	})
}

// ToFloat converts expression to float.
func (b *QueryExprBuilder) ToFloat(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses REAL or DOUBLE PRECISION
			return b.Expr("?::DOUBLE PRECISION", expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses DOUBLE
			return b.Expr("CAST(? AS DOUBLE)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses REAL
			return b.Expr("CAST(? AS REAL)", expr)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("CAST(? AS DOUBLE)", expr)
		},
	})
}

// ToBool converts expression to boolean.
func (b *QueryExprBuilder) ToBool(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL has native BOOLEAN type
			return b.Expr("?::BOOLEAN", expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL doesn't have BOOLEAN, use SIGNED (0/1)
			return b.Expr("CAST(? AS SIGNED) <> 0", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have native BOOLEAN, use INTEGER (0/1)
			return b.Expr("CAST(? AS INTEGER) <> 0", expr)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("CAST(? AS BOOLEAN)", expr)
		},
	})
}

// ToDate converts expression to date.
func (b *QueryExprBuilder) ToDate(expr any, format ...string) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses TO_DATE or CAST
			if len(format) > 0 {
				return b.Expr("TO_DATE(?, ?)", expr, format[0])
			}
			return b.Expr("?::DATE", expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses STR_TO_DATE or CAST
			if len(format) > 0 {
				return b.Expr("STR_TO_DATE(?, ?)", expr, format[0])
			}
			return b.Expr("CAST(? AS DATE)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses DATE function
			if len(format) > 0 {
				return b.Expr("DATE(?, ?)", expr, format[0])
			}
			return b.Expr("DATE(?)", expr)
		},
		Default: func() schema.QueryAppender {
			if len(format) > 0 {
				return b.Expr("TO_DATE(?, ?)", expr, format[0])
			}
			return b.Expr("CAST(? AS DATE)", expr)
		},
	})
}

// ToTime converts expression to time.
func (b *QueryExprBuilder) ToTime(expr any, format ...string) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses TO_TIMESTAMP or CAST
			if len(format) > 0 {
				return b.Expr("TO_TIMESTAMP(?, ?)::TIME", expr, format[0])
			}
			return b.Expr("?::TIME", expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses STR_TO_DATE or CAST
			if len(format) > 0 {
				return b.Expr("TIME(STR_TO_DATE(?, ?))", expr, format[0])
			}
			return b.Expr("CAST(? AS TIME)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses TIME function
			if len(format) > 0 {
				return b.Expr("TIME(?, ?)", expr, format[0])
			}
			return b.Expr("TIME(?)", expr)
		},
		Default: func() schema.QueryAppender {
			if len(format) > 0 {
				return b.Expr("TO_TIME(?, ?)", expr, format[0])
			}
			return b.Expr("CAST(? AS TIME)", expr)
		},
	})
}

// ToTimestamp converts expression to timestamp.
func (b *QueryExprBuilder) ToTimestamp(expr any, format ...string) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses TO_TIMESTAMP or CAST
			if len(format) > 0 {
				return b.Expr("TO_TIMESTAMP(?, ?)", expr, format[0])
			}
			return b.Expr("?::TIMESTAMP", expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses STR_TO_DATE or CAST
			if len(format) > 0 {
				return b.Expr("STR_TO_DATE(?, ?)", expr, format[0])
			}
			return b.Expr("CAST(? AS DATETIME)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses DATETIME function
			if len(format) > 0 {
				return b.Expr("DATETIME(?, ?)", expr, format[0])
			}
			return b.Expr("DATETIME(?)", expr)
		},
		Default: func() schema.QueryAppender {
			if len(format) > 0 {
				return b.Expr("TO_TIMESTAMP(?, ?)", expr, format[0])
			}
			return b.Expr("CAST(? AS TIMESTAMP)", expr)
		},
	})
}

// ToJSON converts expression to JSON.
func (b *QueryExprBuilder) ToJSON(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses ::JSON or ::JSONB
			return b.Expr("?::JSON", expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses CAST AS JSON
			return b.Expr("CAST(? AS JSON)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have explicit JSON type, but supports JSON functions
			return b.Expr("?", expr)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("CAST(? AS JSON)", expr)
		},
	})
}

// ========== JSON Functions ==========

// JSONExtract extracts value from JSON at specified path.
func (b *QueryExprBuilder) JSONExtract(json any, path string) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses -> or ->> operators
			// Convert MySQL-style "$.key" path to PostgreSQL key
			pgPath := path
			if path, ok := strings.CutPrefix(path, "$."); ok {
				pgPath = path
			}
			return b.Expr("(?->>?)", json, pgPath)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_EXTRACT
			return b.Expr("JSON_EXTRACT(?, ?)", json, path)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite has json_extract
			return b.Expr("JSON_EXTRACT(?, ?)", json, path)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_EXTRACT(?, ?)", json, path)
		},
	})
}

// JSONUnquote removes quotes from JSON string.
func (b *QueryExprBuilder) JSONUnquote(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL doesn't have JSON_UNQUOTE, use ->> with '$'
			return b.Expr("(?->>'$')", expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_UNQUOTE
			return b.Expr("JSON_UNQUOTE(?)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite has json_unquote
			return b.Expr("JSON_UNQUOTE(?)", expr)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_UNQUOTE(?)", expr)
		},
	})
}

// JSONArray creates a JSON array from arguments.
func (b *QueryExprBuilder) JSONArray(args ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses json_build_array
			if len(args) == 0 {
				return bun.Safe("'[]'::JSON")
			}
			return b.Expr("JSON_BUILD_ARRAY(?)", newExpressions(constants.CommaSpace, args...))
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_ARRAY
			if len(args) == 0 {
				return bun.Safe("JSON_ARRAY()")
			}
			return b.Expr("JSON_ARRAY(?)", newExpressions(constants.CommaSpace, args...))
		},
		SQLite: func() schema.QueryAppender {
			// SQLite has json_array
			if len(args) == 0 {
				return bun.Safe("JSON_ARRAY()")
			}
			return b.Expr("JSON_ARRAY(?)", newExpressions(constants.CommaSpace, args...))
		},
		Default: func() schema.QueryAppender {
			if len(args) == 0 {
				return bun.Safe("JSON_ARRAY()")
			}
			return b.Expr("JSON_ARRAY(?)", newExpressions(constants.CommaSpace, args...))
		},
	})
}

// JSONObject creates a JSON object from key-value pairs.
func (b *QueryExprBuilder) JSONObject(keyValues ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses json_build_object
			if len(keyValues) == 0 {
				return bun.Safe("'{}'::JSON")
			}
			return b.Expr("JSON_BUILD_OBJECT(?)", newExpressions(constants.CommaSpace, keyValues...))
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_OBJECT
			if len(keyValues) == 0 {
				return bun.Safe("JSON_OBJECT()")
			}
			return b.Expr("JSON_OBJECT(?)", newExpressions(constants.CommaSpace, keyValues...))
		},
		SQLite: func() schema.QueryAppender {
			// SQLite has json_object
			if len(keyValues) == 0 {
				return bun.Safe("JSON_OBJECT()")
			}
			return b.Expr("JSON_OBJECT(?)", newExpressions(constants.CommaSpace, keyValues...))
		},
		Default: func() schema.QueryAppender {
			if len(keyValues) == 0 {
				return bun.Safe("JSON_OBJECT()")
			}
			return b.Expr("JSON_OBJECT(?)", newExpressions(constants.CommaSpace, keyValues...))
		},
	})
}

// JSONContains checks if JSON contains a value.
func (b *QueryExprBuilder) JSONContains(json, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses @> operator for containment
			return b.Expr("? @> ?::JSONB", json, value)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_CONTAINS
			return b.Expr("JSON_CONTAINS(?, ?)", json, value)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have direct JSON_CONTAINS, use EXISTS with json_extract
			return b.Expr("EXISTS (SELECT 1 WHERE JSON_EXTRACT(?, '$[*]') = ?)", json, value)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_CONTAINS(?, ?)", json, value)
		},
	})
}

// JSONContainsPath checks if JSON contains a path.
func (b *QueryExprBuilder) JSONContainsPath(json any, path string) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_path_exists
			return b.Expr("JSONB_PATH_EXISTS(?::JSONB, ?)", json, path)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_CONTAINS_PATH
			return b.Expr("JSON_CONTAINS_PATH(?, 'one', ?)", json, path)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have direct equivalent, use json_extract IS NOT NULL
			return b.Expr("JSON_EXTRACT(?, ?) IS NOT NULL", json, path)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_CONTAINS_PATH(?, 'one', ?)", json, path)
		},
	})
}

// JSONKeys returns the keys of a JSON object.
func (b *QueryExprBuilder) JSONKeys(json any, path ...string) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_object_keys
			if len(path) > 0 {
				return b.Expr("JSONB_OBJECT_KEYS((?::JSONB)->?)", json, path[0])
			}
			return b.Expr("JSONB_OBJECT_KEYS(?::JSONB)", json)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_KEYS
			if len(path) > 0 {
				return b.Expr("JSON_KEYS(?, ?)", json, path[0])
			}
			return b.Expr("JSON_KEYS(?)", json)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have direct equivalent, need to use json_each
			if len(path) > 0 {
				return b.Expr("(SELECT JSON_GROUP_ARRAY(key) FROM JSON_EACH(JSON_EXTRACT(?, ?)))", json, path[0])
			}
			return b.Expr("(SELECT JSON_GROUP_ARRAY(key) FROM JSON_EACH(?))", json)
		},
		Default: func() schema.QueryAppender {
			if len(path) > 0 {
				return b.Expr("JSON_KEYS(?, ?)", json, path[0])
			}
			return b.Expr("JSON_KEYS(?)", json)
		},
	})
}

// JSONLength returns the length of a JSON array or object.
func (b *QueryExprBuilder) JSONLength(json any, path ...string) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_array_length for arrays, jsonb_object_keys for objects
			if len(path) > 0 {
				return b.Expr("JSONB_ARRAY_LENGTH((?::JSONB)->?)", json, path[0])
			}
			return b.Expr("JSONB_ARRAY_LENGTH(?::JSONB)", json)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_LENGTH
			if len(path) > 0 {
				return b.Expr("JSON_LENGTH(?, ?)", json, path[0])
			}
			return b.Expr("JSON_LENGTH(?)", json)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses json_array_length
			if len(path) > 0 {
				return b.Expr("JSON_ARRAY_LENGTH(?, ?)", json, path[0])
			}
			return b.Expr("JSON_ARRAY_LENGTH(?)", json)
		},
		Default: func() schema.QueryAppender {
			if len(path) > 0 {
				return b.Expr("JSON_LENGTH(?, ?)", json, path[0])
			}
			return b.Expr("JSON_LENGTH(?)", json)
		},
	})
}

// JSONType returns the type of JSON value.
func (b *QueryExprBuilder) JSONType(json any, path ...string) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_typeof
			if len(path) > 0 {
				return b.Expr("JSONB_TYPEOF((?::JSONB)->?)", json, path[0])
			}
			return b.Expr("JSONB_TYPEOF(?::JSONB)", json)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_TYPE
			if len(path) > 0 {
				return b.Expr("JSON_TYPE(?, ?)", json, path[0])
			}
			return b.Expr("JSON_TYPE(?)", json)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses json_type
			if len(path) > 0 {
				return b.Expr("JSON_TYPE(?, ?)", json, path[0])
			}
			return b.Expr("JSON_TYPE(?)", json)
		},
		Default: func() schema.QueryAppender {
			if len(path) > 0 {
				return b.Expr("JSON_TYPE(?, ?)", json, path[0])
			}
			return b.Expr("JSON_TYPE(?)", json)
		},
	})
}

// JSONValid checks if a string is valid JSON.
func (b *QueryExprBuilder) JSONValid(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL doesn't have JSON_VALID, try to cast and catch exceptions
			return b.Expr("(? IS NOT NULL AND ?::text::json IS NOT NULL)", expr, expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_VALID
			return b.Expr("JSON_VALID(?)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses json_valid
			return b.Expr("JSON_VALID(?)", expr)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_VALID(?)", expr)
		},
	})
}

// JSONSet sets value at path, creates if not exists, replaces if exists.
func (b *QueryExprBuilder) JSONSet(json any, path string, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_set
			return b.Expr("JSONB_SET(?::JSONB, ?::TEXT[], ?::JSONB, TRUE)", json, path, value)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_SET
			return b.Expr("JSON_SET(?, ?, ?)", json, path, value)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite has json_set
			return b.Expr("JSON_SET(?, ?, ?)", json, path, value)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_SET(?, ?, ?)", json, path, value)
		},
	})
}

// JSONInsert inserts value at path only if path doesn't exist.
func (b *QueryExprBuilder) JSONInsert(json any, path string, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_insert with path array format
			// Convert MySQL-style "$.key" path to PostgreSQL "{key}" format
			pgPath := path
			if key, ok := strings.CutPrefix(path, "$."); ok {
				pgPath = constants.LeftBrace + key + constants.RightBrace
			}
			return b.Expr("JSONB_INSERT(?::JSONB, ?::TEXT[], TO_JSONB(?), FALSE)", json, pgPath, value)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_INSERT
			return b.Expr("JSON_INSERT(?, ?, ?)", json, path, value)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite has json_insert
			return b.Expr("JSON_INSERT(?, ?, ?)", json, path, value)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_INSERT(?, ?, ?)", json, path, value)
		},
	})
}

// JSONReplace replaces value at path only if path exists.
func (b *QueryExprBuilder) JSONReplace(json any, path string, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_set with create_missing = false
			// Convert MySQL-style "$.key" path to PostgreSQL "{key}" format
			pgPath := path
			if key, ok := strings.CutPrefix(path, "$."); ok {
				pgPath = constants.LeftBrace + key + constants.RightBrace
			}
			return b.Expr("JSONB_SET(?::JSONB, ?::TEXT[], TO_JSONB(?), FALSE)", json, pgPath, value)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_REPLACE
			return b.Expr("JSON_REPLACE(?, ?, ?)", json, path, value)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite has json_replace
			return b.Expr("JSON_REPLACE(?, ?, ?)", json, path, value)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_REPLACE(?, ?, ?)", json, path, value)
		},
	})
}

// JSONArrayAppend appends value to JSON array at specified path.
func (b *QueryExprBuilder) JSONArrayAppend(json any, path string, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses || operator to append to array
			if path == constants.Dollar {
				// Root level array
				return b.Expr("(?::JSONB || ?::JSONB)", json, b.JSONArray(value))
			}
			// Nested array - use jsonb_set with concatenation
			return b.Expr(
				"JSONB_SET(?::JSONB, ?::TEXT[], (COALESCE(?::JSONB #> ?::TEXT[], '[]'::JSONB) || ?::JSONB))",
				json, path, json, path, b.JSONArray(value),
			)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_ARRAY_APPEND
			return b.Expr("JSON_ARRAY_APPEND(?, ?, ?)", json, path, value)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have direct JSON_ARRAY_APPEND, need to simulate
			// Get the array, add element, then set it back
			return b.Expr(
				"JSON_SET(?, ?, JSON_INSERT(COALESCE(JSON_EXTRACT(?, ?), '[]'), '$[#]', ?))",
				json, path, json, path, value,
			)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_ARRAY_APPEND(?, ?, ?)", json, path, value)
		},
	})
}

// ========== Utility Functions ==========

// Decode implements DECODE function (Oracle-style case expression).
func (b *QueryExprBuilder) Decode(args ...any) schema.QueryAppender {
	if len(args) < 3 {
		return b.Null()
	}

	return b.ExprByDialect(DialectExpr{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL doesn't have DECODE, convert to CASE WHEN
			return b.convertDecodeToCase(args...)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL doesn't have DECODE, convert to CASE WHEN
			return b.convertDecodeToCase(args...)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have DECODE, convert to CASE WHEN
			return b.convertDecodeToCase(args...)
		},
		Default: func() schema.QueryAppender {
			// For Oracle or other databases that might support DECODE
			return b.Expr("DECODE(?)", newExpressions(constants.CommaSpace, args...))
		},
	})
}

// convertDecodeToCase converts DECODE syntax to CASE WHEN expression using the existing Case builder
// DECODE(expr, search1, result1, search2, result2, ..., defaultResult)
// becomes CASE expr WHEN search1 THEN result1 WHEN search2 THEN result2 ... ELSE defaultResult END
func (b *QueryExprBuilder) convertDecodeToCase(args ...any) schema.QueryAppender {
	if len(args) < 3 {
		return b.Null()
	}

	return b.Case(func(cb CaseBuilder) {
		// Set the CASE expression (simple CASE syntax)
		cb.Case(args[0])

		// Process pairs of (search, result)
		i := 1
		for i+1 < len(args) {
			search := args[i]
			result := args[i+1]
			cb.WhenExpr(search).Then(result)
			i += 2
		}

		// Handle default value (if odd number of arguments after expr)
		if i < len(args) {
			cb.Else(args[i])
		}
	})
}
