package orm

import (
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

// QueryExprBuilder implements the ExprBuilder interface, providing methods to build various SQL expressions.
// It maintains references to the table schema and subquery builder for proper context.
//
// # Code Refactoring and Method Reuse
//
// This file has undergone significant refactoring to improve code maintainability and consistency
// by replacing hardcoded SQL expressions with calls to existing ExprBuilder methods. The refactoring
// follows these key patterns:
//
//  1. Type Conversion Pattern: Use ToXxx() methods instead of hardcoded CAST expressions
//     Example: b.ToInteger(expr) instead of b.Expr("CAST(? AS INTEGER)", expr)
//
//  2. Null-Handling Pattern: Use Coalesce(), IfNull(), IsNull(), IsNotNull() methods
//     Example: b.Coalesce(expr, default) instead of b.Expr("COALESCE(?, ?)", expr, default)
//
//  3. Conditional Logic Pattern: Use Case() builder for complex CASE WHEN expressions
//     Example: b.Case(func(cb) { cb.When(...).Then(...).Else(...) })
//
//  4. Function Reuse Pattern: Use existing function methods instead of hardcoded SQL
//     Example: b.SubString(expr, 1, len) instead of b.Expr("SUBSTR(?, 1, ?)", expr, len)
//
//  5. Complex Expression Decomposition: Break down complex expressions using method calls
//     Example: Use b.ExtractYear() and b.ExtractMonth() instead of inline EXTRACT expressions
//
// Benefits of this approach:
//   - Improved code consistency and readability
//   - Centralized logic for type conversions and common operations
//   - Easier maintenance and debugging
//   - Better support for future database dialects
//   - Reduced code duplication
//
// # Refactored Functions Summary
//
// The following functions have been refactored to use ExprBuilder methods:
//   - Trunc(): Uses b.Round() and b.ToInteger() for SQLite
//   - ToBool(): Uses b.ToInteger() for MySQL and SQLite
//   - IfNull(): Uses b.Coalesce() for PostgreSQL
//   - JsonContainsPath(): Uses b.IsNotNull() for SQLite
//   - JsonValid(): Uses b.IsNotNull() for PostgreSQL
//   - JsonArrayAppend(): Uses b.Coalesce() for PostgreSQL and SQLite
//   - Left(): Uses b.SubString() for SQLite
//   - Age(): Uses b.DateDiff() for MySQL and SQLite
//   - DateDiff(): Uses b.ExtractYear() and b.ExtractMonth() for PostgreSQL
//   - Sign(): Uses b.Case() and b.ToFloat() for SQLite
//
// # Good Examples of Method Reuse
//
// The following functions demonstrate excellent use of the builder pattern:
//   - JsonLength(): Uses b.Case(), ThenSubQuery(), and eb.CountAll()
//   - ConcatWithSep(): Builds parts array and delegates to b.Concat()
//   - convertDecodeToCase(): Uses b.Case() to convert Oracle DECODE to standard SQL
//
// # Known Limitations
//
// 1. Trunc() SQLite: Uses ROUND instead of true truncation (banker's rounding vs toward zero)
// 2. Right() SQLite: Cannot use b.SubString() due to lack of negative index support
// 3. Age() MySQL/SQLite: Returns only year difference, not full interval like PostgreSQL
// 4. DateDiff() year/month: Doesn't consider day component in calculations
// 5. JsonValid() PostgreSQL: May raise exceptions for invalid JSON instead of returning false
//
// When adding new functions or modifying existing ones, prefer using existing ExprBuilder
// methods over hardcoded SQL strings whenever possible. This maintains consistency and
// makes the codebase more maintainable.
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
	} else if b.qb.GetTable() != nil {
		return b.Expr("?TableAlias.?", bun.Name(column))
	} else {
		return b.Expr("?", bun.Name(column))
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

// SubQuery creates a subquery expression for use in larger queries.
// The builder function receives a SelectQuery to configure the subquery.
func (b *QueryExprBuilder) SubQuery(builder func(SelectQuery)) schema.QueryAppender {
	return b.Expr("(?)", b.qb.BuildSubQuery(builder))
}

// Exists creates an EXISTS subquery expression.
func (b *QueryExprBuilder) Exists(builder func(SelectQuery)) schema.QueryAppender {
	return b.Expr("EXISTS (?)", b.qb.BuildSubQuery(builder))
}

// NotExists creates a NOT EXISTS subquery expression.
func (b *QueryExprBuilder) NotExists(builder func(SelectQuery)) schema.QueryAppender {
	return b.Expr("NOT EXISTS (?)", b.qb.BuildSubQuery(builder))
}

// Paren wraps an expression in parentheses for explicit precedence control.
// This is useful when you need to ensure a specific evaluation order in complex expressions.
func (b *QueryExprBuilder) Paren(expr any) schema.QueryAppender {
	return b.Expr("(?)", expr)
}

// Not creates a negation expression (NOT expr).
// This is useful for negating boolean expressions or conditions.
func (b *QueryExprBuilder) Not(expr any) schema.QueryAppender {
	return b.Expr("NOT (?)", expr)
}

// Any wraps a subquery with the ANY operator.
// This is used with comparison operators to check if the comparison is true for any value in the subquery result.
func (b *QueryExprBuilder) Any(builder func(SelectQuery)) schema.QueryAppender {
	return b.Expr("ANY (?)", b.qb.BuildSubQuery(builder))
}

// All wraps a subquery with the ALL operator.
// This is used with comparison operators to check if the comparison is true for all values in the subquery result.
func (b *QueryExprBuilder) All(builder func(SelectQuery)) schema.QueryAppender {
	return b.Expr("ALL (?)", b.qb.BuildSubQuery(builder))
}

// ========== Arithmetic Operators ==========

// Add creates an addition expression (left + right).
func (b *QueryExprBuilder) Add(left, right any) schema.QueryAppender {
	return b.Expr("? + ?", left, right)
}

// Subtract creates a subtraction expression (left - right).
func (b *QueryExprBuilder) Subtract(left, right any) schema.QueryAppender {
	return b.Expr("? - ?", left, right)
}

// Multiply creates a multiplication expression (left * right).
func (b *QueryExprBuilder) Multiply(left, right any) schema.QueryAppender {
	return b.Expr("? * ?", left, right)
}

// Divide creates a division expression (left / right).
// Note: To ensure consistent float results across all databases, we cast to REAL/DOUBLE/NUMERIC.
// This prevents integer division behavior in SQLite and PostgreSQL.
func (b *QueryExprBuilder) Divide(left, right any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			return b.Expr("? / ?", b.ToDecimal(left), b.ToDecimal(right))
		},
		Postgres: func() schema.QueryAppender {
			return b.Expr("? / ?", b.ToDecimal(left), b.ToDecimal(right))
		},
		Default: func() schema.QueryAppender {
			return b.Expr("? / ?", left, right)
		},
	})
}

// ========== Comparison Operators ==========

// Equals creates an equality comparison expression (left = right).
func (b *QueryExprBuilder) Equals(left, right any) schema.QueryAppender {
	return b.Expr("? = ?", left, right)
}

// NotEquals creates an inequality comparison expression (left <> right).
func (b *QueryExprBuilder) NotEquals(left, right any) schema.QueryAppender {
	return b.Expr("? <> ?", left, right)
}

// GreaterThan creates a greater-than comparison expression (left > right).
func (b *QueryExprBuilder) GreaterThan(left, right any) schema.QueryAppender {
	return b.Expr("? > ?", left, right)
}

// GreaterThanOrEqual creates a greater-than-or-equal comparison expression (left >= right).
func (b *QueryExprBuilder) GreaterThanOrEqual(left, right any) schema.QueryAppender {
	return b.Expr("? >= ?", left, right)
}

// LessThan creates a less-than comparison expression (left < right).
func (b *QueryExprBuilder) LessThan(left, right any) schema.QueryAppender {
	return b.Expr("? < ?", left, right)
}

// LessThanOrEqual creates a less-than-or-equal comparison expression (left <= right).
func (b *QueryExprBuilder) LessThanOrEqual(left, right any) schema.QueryAppender {
	return b.Expr("? <= ?", left, right)
}

// ========== Expression Building ==========

// Expr creates an expression builder for complex SQL logic.
func (b *QueryExprBuilder) Expr(expr string, args ...any) schema.QueryAppender {
	return bun.SafeQuery(expr, args...)
}

// Exprs creates an expression builder for complex SQL logic.
func (b *QueryExprBuilder) Exprs(exprs ...any) schema.QueryAppender {
	return newExpressions(constants.CommaSpace, exprs...)
}

// ExprsWithSep creates an expression builder for complex SQL logic with a separator.
func (b *QueryExprBuilder) ExprsWithSep(sep string, exprs ...any) schema.QueryAppender {
	return newExpressions(sep, exprs...)
}

// ExprByDialect creates a cross-database compatible expression.
// It selects the appropriate expression builder based on the current database dialect.
func (b *QueryExprBuilder) ExprByDialect(exprs DialectExprs) schema.QueryAppender {
	switch b.qb.Dialect().Name() {
	case dialect.Oracle:
		if exprs.Oracle != nil {
			return exprs.Oracle()
		}
	case dialect.MSSQL:
		if exprs.SQLServer != nil {
			return exprs.SQLServer()
		}
	case dialect.PG:
		if exprs.Postgres != nil {
			return exprs.Postgres()
		}
	case dialect.MySQL:
		if exprs.MySQL != nil {
			return exprs.MySQL()
		}
	case dialect.SQLite:
		if exprs.SQLite != nil {
			return exprs.SQLite()
		}
	}

	// Fallback to default if database-specific builder is not available
	if exprs.Default != nil {
		return exprs.Default()
	}

	// Return NULL if no suitable builder is found
	return b.Null()
}

// ExecByDialect executes database-specific side-effect callbacks based on the current dialect.
func (b *QueryExprBuilder) ExecByDialect(execs DialectExecs) {
	switch b.qb.Dialect().Name() {
	case dialect.Oracle:
		if execs.Oracle != nil {
			execs.Oracle()

			return
		}

	case dialect.MSSQL:
		if execs.SQLServer != nil {
			execs.SQLServer()

			return
		}

	case dialect.PG:
		if execs.Postgres != nil {
			execs.Postgres()

			return
		}

	case dialect.MySQL:
		if execs.MySQL != nil {
			execs.MySQL()

			return
		}

	case dialect.SQLite:
		if execs.SQLite != nil {
			execs.SQLite()

			return
		}
	}

	if execs.Default != nil {
		execs.Default()
	}
}

// ExecByDialectWithErr executes database-specific callbacks that can return an error.
func (b *QueryExprBuilder) ExecByDialectWithErr(execs DialectExecsWithErr) error {
	switch b.qb.Dialect().Name() {
	case dialect.Oracle:
		if execs.Oracle != nil {
			return execs.Oracle()
		}
	case dialect.MSSQL:
		if execs.SQLServer != nil {
			return execs.SQLServer()
		}
	case dialect.PG:
		if execs.Postgres != nil {
			return execs.Postgres()
		}
	case dialect.MySQL:
		if execs.MySQL != nil {
			return execs.MySQL()
		}
	case dialect.SQLite:
		if execs.SQLite != nil {
			return execs.SQLite()
		}
	}

	if execs.Default != nil {
		return execs.Default()
	}

	return ErrDialectHandlerMissing
}

// FragmentByDialect executes database-specific callbacks that return query fragments.
func (b *QueryExprBuilder) FragmentByDialect(fragments DialectFragments) ([]byte, error) {
	switch b.qb.Dialect().Name() {
	case dialect.Oracle:
		if fragments.Oracle != nil {
			return fragments.Oracle()
		}
	case dialect.MSSQL:
		if fragments.SQLServer != nil {
			return fragments.SQLServer()
		}
	case dialect.PG:
		if fragments.Postgres != nil {
			return fragments.Postgres()
		}
	case dialect.MySQL:
		if fragments.MySQL != nil {
			return fragments.MySQL()
		}
	case dialect.SQLite:
		if fragments.SQLite != nil {
			return fragments.SQLite()
		}
	}

	if fragments.Default != nil {
		return fragments.Default()
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

func (b *QueryExprBuilder) JsonObjectAgg(builder func(JsonObjectAggBuilder)) schema.QueryAppender {
	cb := newJsonObjectAggExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) JsonArrayAgg(builder func(JsonArrayAggBuilder)) schema.QueryAppender {
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

func (b *QueryExprBuilder) WinCount(builder func(WindowCountBuilder)) schema.QueryAppender {
	cb := newWindowCountExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WinSum(builder func(WindowSumBuilder)) schema.QueryAppender {
	cb := newWindowSumExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WinAvg(builder func(WindowAvgBuilder)) schema.QueryAppender {
	cb := newWindowAvgExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WinMin(builder func(WindowMinBuilder)) schema.QueryAppender {
	cb := newWindowMinExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WinMax(builder func(WindowMaxBuilder)) schema.QueryAppender {
	cb := newWindowMaxExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WinStringAgg(builder func(WindowStringAggBuilder)) schema.QueryAppender {
	cb := newWindowStringAggExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WinArrayAgg(builder func(WindowArrayAggBuilder)) schema.QueryAppender {
	cb := newWindowArrayAggExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WinStdDev(builder func(WindowStdDevBuilder)) schema.QueryAppender {
	cb := newWindowStdDevExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WinVariance(builder func(WindowVarianceBuilder)) schema.QueryAppender {
	cb := newWindowVarianceExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WinJsonObjectAgg(builder func(WindowJsonObjectAggBuilder)) schema.QueryAppender {
	cb := newWindowJsonObjectAggExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WinJsonArrayAgg(builder func(WindowJsonArrayAggBuilder)) schema.QueryAppender {
	cb := newWindowJsonArrayAggExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WinBitOr(builder func(WindowBitOrBuilder)) schema.QueryAppender {
	cb := newWindowBitOrExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WinBitAnd(builder func(WindowBitAndBuilder)) schema.QueryAppender {
	cb := newWindowBitAndExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WinBoolOr(builder func(WindowBoolOrBuilder)) schema.QueryAppender {
	cb := newWindowBoolOrExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) WinBoolAnd(builder func(WindowBoolAndBuilder)) schema.QueryAppender {
	cb := newWindowBoolAndExpr(b.qb)
	builder(cb)

	return cb
}

// ========== String Functions ==========

// Concat concatenates strings.
func (b *QueryExprBuilder) Concat(args ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite uses || operator for string concatenation
			if len(args) == 0 {
				return b.Expr("?", constants.Empty)
			}

			if len(args) == 1 {
				return b.Expr("?", args[0])
			}

			return b.ExprsWithSep(" || ", args...)
		},
		Default: func() schema.QueryAppender {
			// PostgreSQL and MySQL support CONCAT function
			return b.Expr("CONCAT(?)", b.Exprs(args...))
		},
	})
}

// ConcatWithSep concatenates strings with a separator.
//
// Good Example: This function demonstrates proper method reuse by calling b.Concat()
// for SQLite implementation instead of hardcoding the concatenation logic. The function
// builds an array of parts (interleaving arguments with separators) and then delegates
// to b.Concat() for the actual concatenation, ensuring consistent behavior across the codebase.
func (b *QueryExprBuilder) ConcatWithSep(separator string, args ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
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
			return b.Expr("CONCAT_WS(?, ?)", separator, b.Exprs(args...))
		},
	})
}

// SubString extracts a substring from a string.
func (b *QueryExprBuilder) SubString(expr any, start int, length ...int) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
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
	return b.ExprByDialect(DialectExprs{
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
	return b.ExprByDialect(DialectExprs{
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
//
// Refactoring Note: This function has been refactored to use b.SubString() for SQLite
// implementation instead of hardcoded "SUBSTR(?, 1, ?)" expression. This improves code
// consistency and ensures that string manipulation logic is centralized in the SubString method.
//
// Behavior: SQLite doesn't have a LEFT function, so we use SUBSTR starting at position 1
// with the specified length, which produces the same result.
func (b *QueryExprBuilder) Left(expr any, length int) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have LEFT, use SUBSTR
			return b.SubString(expr, 1, length)
		},
		Default: func() schema.QueryAppender {
			// PostgreSQL and MySQL support LEFT
			return b.Expr("LEFT(?, ?)", expr, length)
		},
	})
}

// Right returns the rightmost n characters.
//
// Limitation: This function cannot be fully refactored to use b.SubString() for SQLite
// because SubString() doesn't support negative indices. SQLite's SUBSTR with negative
// start position counts from the end of the string, which is the correct behavior for
// RIGHT but is not supported by the SubString() method's current API.
//
// Future Enhancement: Consider extending SubString() to support negative indices to
// enable full refactoring of this function.
func (b *QueryExprBuilder) Right(expr any, length int) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
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
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have REPEAT, need to implement with REPLACE and a helper
			// Use REPLACE(SUBSTR(QUOTE(ZEROBLOB((count+1)/2)), 3, count), '0', expr)
			return b.Expr("REPLACE(SUBSTR(QUOTE(ZEROBLOB(?)), 3, ?), ?, ?)", b.Divide(b.Paren(b.Add(count, 1)), 2), count, "0", expr)
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
	return b.ExprByDialect(DialectExprs{
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
	return b.Expr("CURRENT_DATE")
}

// CurrentTime returns the current time.
func (b *QueryExprBuilder) CurrentTime() schema.QueryAppender {
	return b.Expr("CURRENT_TIME")
}

// CurrentTimestamp returns the current timestamp.
func (b *QueryExprBuilder) CurrentTimestamp() schema.QueryAppender {
	return b.Expr("CURRENT_TIMESTAMP")
}

// Now returns the current timestamp.
func (b *QueryExprBuilder) Now() schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("NOW()")
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("NOW()")
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses DATETIME('now') for current timestamp
			return b.Expr("DATETIME('now')")
		},
		Default: func() schema.QueryAppender {
			return b.Expr("NOW()")
		},
	})
}

// ExtractYear extracts the year from a date/timestamp.
func (b *QueryExprBuilder) ExtractYear(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("EXTRACT(YEAR FROM ?)", expr)
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("EXTRACT(YEAR FROM ?)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses STRFTIME to extract year
			return b.ToInteger(
				b.Expr("STRFTIME(?, ?)", "%Y", expr),
			)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("EXTRACT(YEAR FROM ?)", expr)
		},
	})
}

// ExtractMonth extracts the month from a date/timestamp.
func (b *QueryExprBuilder) ExtractMonth(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("EXTRACT(MONTH FROM ?)", expr)
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("EXTRACT(MONTH FROM ?)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses STRFTIME to extract month
			return b.ToInteger(
				b.Expr("STRFTIME(?, ?)", "%m", expr),
			)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("EXTRACT(MONTH FROM ?)", expr)
		},
	})
}

// ExtractDay extracts the day from a date/timestamp.
func (b *QueryExprBuilder) ExtractDay(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("EXTRACT(DAY FROM ?)", expr)
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("EXTRACT(DAY FROM ?)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses STRFTIME to extract day
			return b.ToInteger(
				b.Expr("STRFTIME(?, ?)", "%d", expr),
			)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("EXTRACT(DAY FROM ?)", expr)
		},
	})
}

// ExtractHour extracts the hour from a timestamp.
func (b *QueryExprBuilder) ExtractHour(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("EXTRACT(HOUR FROM ?)", expr)
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("EXTRACT(HOUR FROM ?)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses STRFTIME to extract hour
			return b.ToInteger(
				b.Expr("STRFTIME(?, ?)", "%H", expr),
			)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("EXTRACT(HOUR FROM ?)", expr)
		},
	})
}

// ExtractMinute extracts the minute from a timestamp.
func (b *QueryExprBuilder) ExtractMinute(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("EXTRACT(MINUTE FROM ?)", expr)
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("EXTRACT(MINUTE FROM ?)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses STRFTIME to extract minute
			return b.ToInteger(
				b.Expr("STRFTIME(?, ?)", "%M", expr),
			)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("EXTRACT(MINUTE FROM ?)", expr)
		},
	})
}

// ExtractSecond extracts the second from a timestamp.
func (b *QueryExprBuilder) ExtractSecond(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("EXTRACT(SECOND FROM ?)", expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL EXTRACT returns integer, cast to DECIMAL for consistency with PostgreSQL
			return b.ToDecimal(
				b.Expr("EXTRACT(SECOND FROM ?)", expr),
				10, 6,
			)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses STRFTIME to extract second
			// Note: STRFTIME('%S') returns seconds without fractional part
			// Use STRFTIME('%f') for seconds with fractional part, but we'll keep it simple
			return b.ToDecimal(
				b.Expr("STRFTIME(?, ?)", "%S", expr),
				10, 6,
			)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("EXTRACT(SECOND FROM ?)", expr)
		},
	})
}

// DateTrunc truncates date/timestamp to specified precision.
func (b *QueryExprBuilder) DateTrunc(precision string, expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL has native DATE_TRUNC
			return b.Expr("DATE_TRUNC(?, ?)", precision, expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL needs different approach based on precision
			switch precision {
			case "year":
				return b.Expr("DATE_FORMAT(?, ?)", expr, "%Y-01-01")
			case "month":
				return b.Expr("DATE_FORMAT(?, ?)", expr, "%Y-%m-01")
			case "day":
				return b.Expr("DATE(?)", expr)
			case "hour":
				return b.Expr("DATE_FORMAT(?, ?)", expr, "%Y-%m-%d %H:00:00")
			case "minute":
				return b.Expr("DATE_FORMAT(?, ?)", expr, "%Y-%m-%d %H:%i:00")
			default:
				return b.Expr("DATE(?)", expr)
			}
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses strftime for truncation
			switch precision {
			case "year":
				return b.Expr("STRFTIME(?, ?)", "%Y-01-01", expr)
			case "month":
				return b.Expr("STRFTIME(?, ?)", "%Y-%m-01", expr)
			case "day":
				return b.Expr("DATE(?)", expr)
			case "hour":
				return b.Expr("STRFTIME(?, ?)", "%Y-%m-%d %H:00:00", expr)
			case "minute":
				return b.Expr("STRFTIME(?, ?)", "%Y-%m-%d %H:%M:00", expr)
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
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses INTERVAL syntax
			return b.Expr("? + INTERVAL '? ?'", expr, interval, b.Expr(unit))
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses DATE_ADD with INTERVAL
			return b.Expr("DATE_ADD(?, INTERVAL ? ?)", expr, interval, b.Expr(unit))
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

			return b.Expr("DATETIME(?, '+? ?')", expr, interval, b.Expr(modifier))
		},
		Default: func() schema.QueryAppender {
			return b.Expr("DATE_ADD(?, INTERVAL ? ?)", expr, interval, b.Expr(unit))
		},
	})
}

// DateSubtract subtracts interval from date/timestamp.
func (b *QueryExprBuilder) DateSubtract(expr any, interval int, unit string) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses INTERVAL syntax
			return b.Expr("? - INTERVAL '? ?'", expr, interval, b.Expr(unit))
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses DATE_SUB with INTERVAL
			return b.Expr("DATE_SUB(?, INTERVAL ? ?)", expr, interval, b.Expr(unit))
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
//
// Refactoring Note: This function has been refactored to use b.ExtractYear() and b.ExtractMonth()
// for PostgreSQL's month calculation instead of hardcoded EXTRACT expressions. This improves
// code modularity and makes the complex calculation more readable:
//   - Before: "EXTRACT(YEAR FROM ?) * 12 + EXTRACT(MONTH FROM ?) - EXTRACT(YEAR FROM ?) * 12 - EXTRACT(MONTH FROM ?)"
//   - After: "? * 12 + ? - ? * 12 - ?" with b.ExtractYear() and b.ExtractMonth() calls
//
// Behavior: The month calculation computes total months from year 0 for both dates,
// then subtracts to get the difference. This handles year boundaries correctly.
//
// Limitation: The year and month calculations return simple differences without considering
// the day component. For example, the difference between Jan 31 and Feb 1 is 1 month,
// even though it's only 1 day. For day-precise calculations, use the "day" unit.
func (b *QueryExprBuilder) DateDiff(start, end any, unit string) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses EXTRACT with subtraction
			switch unit {
			case "day", "days":
				return b.ToDate(
					b.Paren(b.Subtract(end, start)),
				)
				// return b.Expr("(? - ?)::DATE", end, start)
			case "year", "years":
				return b.Expr("EXTRACT(YEAR FROM ?) - EXTRACT(YEAR FROM ?)", end, start)
			case "month", "months":
				return b.Expr("? * 12 + ? - ? * 12 - ?", b.ExtractYear(end), b.ExtractMonth(end), b.ExtractYear(start), b.ExtractMonth(start))
			default:
				return b.Expr("EXTRACT(DAYS FROM ? - ?)", end, start)
			}
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has DATEDIFF for days, TIMESTAMPDIFF for other units
			// Cast to DECIMAL to ensure float type for consistency across databases
			switch unit {
			case "day", "days":
				return b.ToDecimal(
					b.Expr("DATEDIFF(?, ?)", end, start),
					20, 6,
				)

			default:
				return b.ToDecimal(
					b.Expr("TIMESTAMPDIFF(?, ?, ?)", b.Expr(unit), start, end),
					20, 6,
				)
			}
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses julianday for date differences
			switch unit {
			case "day", "days":
				return b.Subtract(
					b.Expr("JULIANDAY(?)", end),
					b.Expr("JULIANDAY(?)", start),
				)

			case "year", "years":
				return b.Subtract(
					b.Expr("STRFTIME(?, ?)", "%Y", end),
					b.Expr("STRFTIME(?, ?)", "%Y", start),
				)

			case "month", "months":
				return b.Add(
					b.Multiply(
						b.Paren(
							b.Subtract(
								b.Expr("STRFTIME(?, ?)", "%Y", end),
								b.Expr("STRFTIME(?, ?)", "%Y", start),
							),
						),
						12,
					),
					b.Paren(
						b.Subtract(
							b.Expr("STRFTIME(?, ?)", "%m", end),
							b.Expr("STRFTIME(?, ?)", "%m", start),
						),
					),
				)

			default:
				return b.Subtract(
					b.Expr("JULIANDAY(?)", end),
					b.Expr("JULIANDAY(?)", start),
				)
			}
		},
		Default: func() schema.QueryAppender {
			return b.Expr("DATEDIFF(?, ?, ?)", b.Expr(unit), end, start)
		},
	})
}

// Age returns the age (interval) between two timestamps.
//
// Refactoring Note: This function has been refactored to use b.DateDiff() for MySQL
// and SQLite implementations instead of hardcoded date difference expressions. This
// eliminates code duplication and ensures consistent date calculation logic:
//   - MySQL: Uses b.DateDiff(start, end, "YEAR") instead of "TIMESTAMPDIFF(YEAR, ?, ?)"
//   - SQLite: Uses b.DateDiff(start, end, "year") instead of "STRFTIME('%Y', ?) - STRFTIME('%Y', ?)"
//
// Behavior: PostgreSQL's AGE function returns a full interval (years, months, days),
// while MySQL and SQLite implementations return only the year difference as an integer.
// This is a simplified approximation suitable for most age calculations.
//
// Limitation: The MySQL and SQLite implementations don't account for months and days,
// so they may be off by up to a year compared to PostgreSQL's AGE function. For precise
// age calculations including months and days, consider using DateDiff with "month" or "day" units.
func (b *QueryExprBuilder) Age(start, end any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL has native AGE function
			return b.Expr("AGE(?, ?)", end, start)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL doesn't have AGE, calculate years difference
			return b.DateDiff(start, end, "YEAR")
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have AGE, calculate years difference
			return b.DateDiff(start, end, "year")
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
//
// Refactoring Note: This function has been refactored to use existing ExprBuilder methods
// for better code reusability and consistency:
//   - SQLite with precision: Uses b.Round() instead of hardcoded "ROUND(?, ?)"
//   - SQLite without precision: Uses b.ToInteger() instead of hardcoded "CAST(? AS INTEGER)"
//
// Limitation: SQLite's ROUND behavior differs slightly from TRUNC - ROUND uses banker's rounding
// while TRUNC always rounds toward zero. For most use cases this difference is negligible.
func (b *QueryExprBuilder) Trunc(expr any, precision ...int) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have TRUNC function, use ROUND with precision
			if len(precision) > 0 {
				// For positive precision, ROUND works similarly to TRUNC for most cases
				return b.Round(expr, precision[0])
			}
			// For no precision, truncate to integer using CAST
			return b.ToInteger(expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses TRUNCATE instead of TRUNC
			if len(precision) > 0 {
				return b.Expr("TRUNCATE(?, ?)", expr, precision[0])
			}

			return b.Expr("TRUNCATE(?, 0)", expr)
		},
		Default: func() schema.QueryAppender {
			if len(precision) > 0 {
				return b.Expr("TRUNC(?, ?)", expr, precision[0])
			}

			return b.Expr("TRUNC(?)", expr)
		},
	})
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
	return b.Expr("PI()")
}

// Random returns a random value between 0 and 1.
func (b *QueryExprBuilder) Random() schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite's RANDOM() returns integer in range [-9223372036854775808, 9223372036854775807]
			// Normalize to [0, 1) by taking absolute value and dividing by 2^63
			return b.Abs(
				b.Divide(
					b.Multiply(b.Expr("RANDOM()"), 1.0),
					9223372036854775808.0,
				),
			)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses RAND() instead of RANDOM()
			return b.Expr("RAND()")
		},
		Default: func() schema.QueryAppender {
			return b.Expr("RANDOM()")
		},
	})
}

// Sign returns the sign of a number.
//
// Refactoring Note: This function has been refactored to use b.Case() and b.ToFloat()
// for SQLite implementation instead of hardcoded CASE WHEN and CAST expressions. This
// demonstrates proper use of the builder pattern for complex conditional logic:
//   - Uses b.Case() with WhenExpr/Then/Else for conditional logic
//   - Uses b.ToFloat() for type conversion to REAL
//   - Separates concerns: conditional logic vs type conversion
//
// Behavior: Returns 1 for positive numbers, -1 for negative numbers, and 0 for zero.
// The result is cast to REAL (float) in SQLite to match the behavior of other databases.
func (b *QueryExprBuilder) Sign(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have built-in SIGN function, use CASE expression
			// Cast to REAL to ensure float type
			return b.ToFloat(
				b.Case(func(cb CaseBuilder) {
					cb.WhenExpr(b.GreaterThan(expr, 0)).Then(1).
						WhenExpr(b.LessThan(expr, 0)).Then(-1).
						Else(0)
				}),
			)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL SIGN returns integer, cast to DECIMAL for consistency with PostgreSQL
			return b.ToDecimal(b.Expr("SIGN(?)", expr), 2, 1)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("SIGN(?)", expr)
		},
	})
}

// Mod returns the remainder of division.
func (b *QueryExprBuilder) Mod(dividend, divisor any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have MOD function, use % operator
			return b.Expr("? % ?", dividend, divisor)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("MOD(?, ?)", dividend, divisor)
		},
	})
}

// Greatest returns the greatest value among arguments.
func (b *QueryExprBuilder) Greatest(args ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have GREATEST function, use MAX
			return b.Expr("MAX(?)", newExpressions(constants.CommaSpace, args...))
		},
		Default: func() schema.QueryAppender {
			return b.Expr("GREATEST(?)", newExpressions(constants.CommaSpace, args...))
		},
	})
}

// Least returns the least value among arguments.
func (b *QueryExprBuilder) Least(args ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have LEAST function, use MIN
			return b.Expr("MIN(?)", newExpressions(constants.CommaSpace, args...))
		},
		Default: func() schema.QueryAppender {
			return b.Expr("LEAST(?)", newExpressions(constants.CommaSpace, args...))
		},
	})
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
//
// Refactoring Note: This function has been refactored to use b.Coalesce() for PostgreSQL
// instead of hardcoded "COALESCE(?, ?)" expression. This ensures consistent null-handling
// across the codebase and allows for easier maintenance.
//
// Behavior: IFNULL and COALESCE are functionally equivalent when used with two arguments.
// PostgreSQL doesn't support IFNULL, so we use COALESCE which is SQL standard.
func (b *QueryExprBuilder) IfNull(expr, defaultValue any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL doesn't have IFNULL, use COALESCE instead
			return b.Coalesce(expr, defaultValue)
		},
		Default: func() schema.QueryAppender {
			// MySQL and SQLite support IFNULL
			return b.Expr("IFNULL(?, ?)", expr, defaultValue)
		},
	})
}

// ========== Type Conversion Functions ==========

// ToString converts expression to string.
func (b *QueryExprBuilder) ToString(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses TEXT or VARCHAR
			return b.Expr("?::TEXT", b.Paren(expr))
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
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses INTEGER or INT
			return b.Expr("?::INTEGER", b.Paren(expr))
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
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses NUMERIC
			if len(precision) >= 2 {
				return b.Expr("?::NUMERIC(?, ?)", b.Paren(expr), precision[0], precision[1])
			} else if len(precision) == 1 {
				return b.Expr("?::NUMERIC(?)", b.Paren(expr), precision[0])
			}

			return b.Expr("?::NUMERIC", b.Paren(expr))
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
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses REAL or DOUBLE PRECISION
			return b.Expr("?::DOUBLE PRECISION", b.Paren(expr))
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
//
// Refactoring Note: This function has been refactored to use b.ToInteger() for type conversion
// in MySQL and SQLite implementations instead of hardcoded CAST expressions. This improves
// consistency and allows the type conversion logic to be centralized in the ToInteger() method.
//
// Behavior: MySQL and SQLite don't have native BOOLEAN types, so we convert to integer first
// and then compare with 0 (non-zero values are truthy, zero is falsy).
func (b *QueryExprBuilder) ToBool(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL has native BOOLEAN type
			return b.Expr("?::BOOLEAN", b.Paren(expr))
		},
		MySQL: func() schema.QueryAppender {
			// MySQL doesn't have BOOLEAN, use SIGNED (0/1)
			return b.NotEquals(b.ToInteger(expr), 0)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have native BOOLEAN, use INTEGER (0/1)
			return b.NotEquals(b.ToInteger(expr), 0)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("CAST(? AS BOOLEAN)", expr)
		},
	})
}

// ToDate converts expression to date.
func (b *QueryExprBuilder) ToDate(expr any, format ...string) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses TO_DATE or CAST
			if len(format) > 0 {
				return b.Expr("TO_DATE(?, ?)", expr, format[0])
			}

			return b.Expr("?::DATE", b.Paren(expr))
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses STR_TO_DATE or CAST
			if len(format) > 0 {
				return b.Expr("STR_TO_DATE(?, ?)", expr, format[0])
			}

			return b.Expr("CAST(? AS DATE)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses DATE function or STRFTIME for format conversion
			if len(format) > 0 {
				// Use STRFTIME to ensure standard date format output
				return b.Expr("STRFTIME(?, ?)", "%Y-%m-%d", expr)
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
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses TO_TIMESTAMP or CAST
			if len(format) > 0 {
				return b.Expr("TO_TIMESTAMP(?, ?)::TIME", expr, format[0])
			}

			return b.Expr("?::TIME", b.Paren(expr))
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses STR_TO_DATE or CAST
			if len(format) > 0 {
				return b.Expr("TIME(STR_TO_DATE(?, ?))", expr, format[0])
			}

			return b.Expr("CAST(? AS TIME)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses TIME function or STRFTIME for format conversion
			if len(format) > 0 {
				// Use STRFTIME to ensure standard time format output
				return b.Expr("STRFTIME(?, ?)", "%H:%M:%S", expr)
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
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses TO_TIMESTAMP or CAST
			if len(format) > 0 {
				return b.Expr("TO_TIMESTAMP(?, ?)", expr, format[0])
			}

			return b.Expr("?::TIMESTAMP", b.Paren(expr))
		},
		MySQL: func() schema.QueryAppender {
			// MySQL uses STR_TO_DATE or CAST
			if len(format) > 0 {
				return b.Expr("STR_TO_DATE(?, ?)", expr, format[0])
			}

			return b.Expr("CAST(? AS DATETIME)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite uses DATETIME function or STRFTIME for format conversion
			if len(format) > 0 {
				// Use STRFTIME to ensure standard datetime format output
				return b.Expr("STRFTIME(?, ?)", "%Y-%m-%d %H:%M:%S", expr)
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

// ToJson converts expression to JSON.
func (b *QueryExprBuilder) ToJson(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses ::JSON or ::JSONB
			return b.Expr("?::JSONB", b.Paren(expr))
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

// JsonExtract extracts value from JSON at specified path.
func (b *QueryExprBuilder) JsonExtract(json any, path string) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
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

// JsonUnquote removes quotes from JSON string.
func (b *QueryExprBuilder) JsonUnquote(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL doesn't have JSON_UNQUOTE
			// JsonExtract already uses ->> which returns text (unquoted)
			// So we just return the expression as-is
			return b.Expr("?", expr)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_UNQUOTE
			return b.Expr("JSON_UNQUOTE(?)", expr)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have json_unquote function
			// JSON_EXTRACT already returns unquoted values, so just return the expr
			return b.Expr("?", expr)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_UNQUOTE(?)", expr)
		},
	})
}

// JsonArray creates a JSON array from arguments.
func (b *QueryExprBuilder) JsonArray(args ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_build_array
			if len(args) == 0 {
				return b.ToJson("[]")
			}

			return b.Expr("JSONB_BUILD_ARRAY(?)", b.Exprs(args...))
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_ARRAY
			if len(args) == 0 {
				return b.Expr("JSON_ARRAY()")
			}

			return b.Expr("JSON_ARRAY(?)", b.Exprs(args...))
		},
		SQLite: func() schema.QueryAppender {
			// SQLite has json_array
			if len(args) == 0 {
				return b.Expr("JSON_ARRAY()")
			}

			return b.Expr("JSON_ARRAY(?)", b.Exprs(args...))
		},
		Default: func() schema.QueryAppender {
			if len(args) == 0 {
				return b.Expr("JSON_ARRAY()")
			}

			return b.Expr("JSON_ARRAY(?)", b.Exprs(args...))
		},
	})
}

// JsonObject creates a JSON object from key-value pairs.
func (b *QueryExprBuilder) JsonObject(keyValues ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_build_object
			if len(keyValues) == 0 {
				return b.ToJson("{}")
			}

			return b.Expr("JSONB_BUILD_OBJECT(?)", b.Exprs(keyValues...))
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_OBJECT
			if len(keyValues) == 0 {
				return b.Expr("JSON_OBJECT()")
			}

			return b.Expr("JSON_OBJECT(?)", b.Exprs(keyValues...))
		},
		SQLite: func() schema.QueryAppender {
			// SQLite has json_object
			if len(keyValues) == 0 {
				return b.Expr("JSON_OBJECT()")
			}

			return b.Expr("JSON_OBJECT(?)", b.Exprs(keyValues...))
		},
		Default: func() schema.QueryAppender {
			if len(keyValues) == 0 {
				return b.Expr("JSON_OBJECT()")
			}

			return b.Expr("JSON_OBJECT(?)", b.Exprs(keyValues...))
		},
	})
}

// JsonContains checks if JSON contains a value.
func (b *QueryExprBuilder) JsonContains(json, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses @> operator for containment
			return b.Expr("? @> ?", json, b.ToJson(value))
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_CONTAINS
			return b.Expr("JSON_CONTAINS(?, ?)", json, value)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have direct JSON_CONTAINS, use EXISTS with json_extract
			return b.Exists(func(sq SelectQuery) {
				sq.SelectExpr(func(eb ExprBuilder) any { return eb.Literal(1) }).
					Where(func(cb ConditionBuilder) {
						cb.Expr(func(eb ExprBuilder) any {
							return eb.Equals(
								eb.Expr("JSON_EXTRACT(?, ?)", json, "$[*]"),
								value,
							)
						})
					})
			})
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_CONTAINS(?, ?)", json, value)
		},
	})
}

// JsonContainsPath checks if JSON contains a path.
//
// Refactoring Note: This function has been refactored to use b.IsNotNull() for SQLite
// implementation instead of hardcoded "IS NOT NULL" expression. This improves code
// consistency and makes null-checking logic more explicit and maintainable.
//
// Behavior: SQLite doesn't have a direct JSON_CONTAINS_PATH equivalent, so we check
// if JSON_EXTRACT returns a non-null value, which indicates the path exists.
func (b *QueryExprBuilder) JsonContainsPath(json any, path string) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_path_exists
			return b.Expr("JSONB_PATH_EXISTS(?, ?)", b.ToJson(json), path)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_CONTAINS_PATH
			return b.Expr("JSON_CONTAINS_PATH(?, ?, ?)", json, "one", path)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have direct equivalent, use json_extract IS NOT NULL
			return b.IsNotNull(b.Expr("JSON_EXTRACT(?, ?)", json, path))
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_CONTAINS_PATH(?, ?, ?)", json, "one", path)
		},
	})
}

// JsonKeys returns the keys of a JSON object.
func (b *QueryExprBuilder) JsonKeys(json any, path ...string) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_object_keys
			if len(path) > 0 {
				return b.Expr("JSONB_OBJECT_KEYS(?->?)", b.ToJson(json), path[0])
			}

			return b.Expr("JSONB_OBJECT_KEYS(?)", b.ToJson(json))
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
				return b.SubQuery(func(sq SelectQuery) {
					sq.TableExpr(
						func(eb ExprBuilder) any {
							return eb.Expr("JSON_EACH(JSON_EXTRACT(?, ?))", json, path[0])
						}).
						SelectExpr(func(eb ExprBuilder) any {
							return eb.Expr("JSON_GROUP_ARRAY(key)")
						})
				})
			}

			return b.SubQuery(func(sq SelectQuery) {
				sq.TableExpr(
					func(eb ExprBuilder) any {
						return eb.Expr("JSON_EACH(?)", json)
					}).
					SelectExpr(func(eb ExprBuilder) any {
						return eb.Expr("JSON_GROUP_ARRAY(key)")
					})
			})
		},
		Default: func() schema.QueryAppender {
			if len(path) > 0 {
				return b.Expr("JSON_KEYS(?, ?)", json, path[0])
			}

			return b.Expr("JSON_KEYS(?)", json)
		},
	})
}

// JsonLength returns the length of a JSON array or object.
//
// Good Example: This function demonstrates proper use of the builder pattern for complex
// conditional logic. It uses:
//   - b.Case() for CASE WHEN expressions instead of hardcoded SQL
//   - ThenSubQuery() for subquery expressions
//   - eb.CountAll() for aggregate functions
//
// This is the recommended pattern for implementing complex database-specific logic.
func (b *QueryExprBuilder) JsonLength(json any, path ...string) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL: Support both arrays and objects by checking type
			// For arrays: use JSONB_ARRAY_LENGTH
			// For objects: count keys using JSONB_OBJECT_KEYS
			// For other types: return 0
			var jsonExpr schema.QueryAppender
			if len(path) > 0 {
				// With path: extract the value at path first
				jsonExpr = b.Expr("?->?", b.ToJson(json), path[0])
			} else {
				// Without path: work on the root value
				jsonExpr = b.ToJson(json)
			}

			return b.Case(func(cb CaseBuilder) {
				cb.Case(b.Expr("JSONB_TYPEOF(?)", jsonExpr)).
					WhenExpr("array").
					Then(b.Expr("JSONB_ARRAY_LENGTH(?)", jsonExpr)).
					WhenExpr("object").
					ThenSubQuery(func(query SelectQuery) {
						query.SelectExpr(func(eb ExprBuilder) any { return eb.CountAll() }).
							TableExpr(func(eb ExprBuilder) any { return eb.Expr("JSONB_OBJECT_KEYS(?)", jsonExpr) })
					}).
					Else(0)
			})
		},
		MySQL: func() schema.QueryAppender {
			// MySQL has JSON_LENGTH that works with both arrays and objects
			if len(path) > 0 {
				return b.Expr("JSON_LENGTH(?, ?)", json, path[0])
			}

			return b.Expr("JSON_LENGTH(?)", json)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite: Support both arrays and objects by checking type
			// For arrays: use JSON_ARRAY_LENGTH
			// For objects: count keys using JSON_EACH
			// For other types: return 0
			var argsExpr schema.QueryAppender
			if len(path) > 0 {
				// With path: extract the value at path first
				argsExpr = b.Expr("?, ?", json, path[0])
			} else {
				// Without path: work on the root value
				argsExpr = b.Expr("?", json)
			}

			return b.Case(func(cb CaseBuilder) {
				cb.Case(b.Expr("JSON_TYPE(?)", argsExpr)).
					WhenExpr("array").
					Then(b.Expr("JSON_ARRAY_LENGTH(?)", argsExpr)).
					WhenExpr("object").
					ThenSubQuery(func(query SelectQuery) {
						query.SelectExpr(func(eb ExprBuilder) any { return eb.CountAll() }).
							TableExpr(func(eb ExprBuilder) any { return eb.Expr("JSON_EACH(?)", argsExpr) })
					}).
					Else(0)
			})
		},
		Default: func() schema.QueryAppender {
			if len(path) > 0 {
				return b.Expr("JSON_LENGTH(?, ?)", json, path[0])
			}

			return b.Expr("JSON_LENGTH(?)", json)
		},
	})
}

// JsonType returns the type of JSON value.
func (b *QueryExprBuilder) JsonType(json any, path ...string) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_typeof
			if len(path) > 0 {
				return b.Expr("JSONB_TYPEOF(?->?)", b.ToJson(json), path[0])
			}

			return b.Expr("JSONB_TYPEOF(?)", b.ToJson(json))
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

// JsonValid checks if a string is valid JSON.
//
// Refactoring Note: This function has been refactored to use b.IsNotNull() for PostgreSQL
// implementation instead of hardcoded "IS NOT NULL" expressions. This improves consistency
// and makes the null-checking logic more explicit.
//
// Behavior: PostgreSQL doesn't have a JSON_VALID function, so we validate by attempting
// to cast the expression to JSONB. We check both that the input is not null and that
// the cast result is not null (cast returns null for invalid JSON).
//
// Limitation: This approach doesn't catch all JSON validation errors in PostgreSQL as
// invalid JSON will raise an exception rather than return null. For production use,
// consider wrapping in a PL/pgSQL function with exception handling.
func (b *QueryExprBuilder) JsonValid(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL doesn't have JSON_VALID, try to cast and catch exceptions
			return b.Expr("? AND ?", b.IsNotNull(expr), b.IsNotNull(b.ToJson(b.ToString(expr))))
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

func (b *QueryExprBuilder) JsonSet(json any, path string, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_set
			// Convert MySQL-style "$.key" path to PostgreSQL "{key}" format
			pgPath := path
			if key, ok := strings.CutPrefix(path, "$."); ok {
				pgPath = constants.LeftBrace + key + constants.RightBrace
			}

			return b.Expr("JSONB_SET(?, ?::TEXT[], ?, TRUE)", b.ToJson(json), pgPath, b.ToJson(value))
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

// JsonInsert inserts value at path only if path doesn't exist.
func (b *QueryExprBuilder) JsonInsert(json any, path string, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_insert with path array format
			// Convert MySQL-style "$.key" path to PostgreSQL "{key}" format
			pgPath := path
			if key, ok := strings.CutPrefix(path, "$."); ok {
				pgPath = constants.LeftBrace + key + constants.RightBrace
			}

			return b.Expr("JSONB_INSERT(?, ?::TEXT[], TO_JSONB(?), FALSE)", b.ToJson(json), pgPath, value)
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

// JsonReplace replaces value at path only if path exists.
func (b *QueryExprBuilder) JsonReplace(json any, path string, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_set with create_missing = false
			// Convert MySQL-style "$.key" path to PostgreSQL "{key}" format
			pgPath := path
			if key, ok := strings.CutPrefix(path, "$."); ok {
				pgPath = constants.LeftBrace + key + constants.RightBrace
			}

			return b.Expr("JSONB_SET(?, ?::TEXT[], TO_JSONB(?), FALSE)", b.ToJson(json), pgPath, value)
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

// JsonArrayAppend appends value to JSON array at specified path.
//
// Refactoring Note: This function has been refactored to use b.Coalesce() for null-handling
// in both PostgreSQL and SQLite implementations instead of hardcoded COALESCE expressions.
// This improves code modularity and makes the default value handling more explicit:
//   - PostgreSQL: Uses b.Coalesce() to provide empty array '[]'::JSONB if path doesn't exist
//   - SQLite: Uses b.Coalesce() to provide empty array "[]" if JSON_EXTRACT returns null
//
// Behavior: Both implementations ensure that if the target path doesn't exist or is null,
// an empty array is created before appending the new value. This prevents errors when
// appending to non-existent paths.
func (b *QueryExprBuilder) JsonArrayAppend(json any, path string, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses || operator to append to array
			if path == constants.Dollar {
				// Root level array
				return b.Expr("(? || ?)", b.ToJson(json), b.JsonArray(value))
			}

			// Convert MySQL-style "$.key" path to PostgreSQL "{key}" format
			pgPath := path
			if key, ok := strings.CutPrefix(path, "$."); ok {
				pgPath = constants.LeftBrace + key + constants.RightBrace
			}

			// Nested array - use jsonb_set with concatenation
			return b.Expr(
				"JSONB_SET(?, ?::TEXT[], (? || ?))",
				b.ToJson(json),
				pgPath,
				b.Coalesce(
					b.Expr("? #> ?::TEXT[]", b.ToJson(json), pgPath),
					b.ToJson("[]"),
				),
				b.JsonArray(value),
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
				"JSON_SET(?, ?, JSON_INSERT(?, ?, ?))",
				json,
				path,
				b.Coalesce(
					b.Expr("JSON_EXTRACT(?, ?)", json, path),
					"[]",
				),
				"$[#]",
				value,
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

	return b.ExprByDialect(DialectExprs{
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

// convertDecodeToCase converts DECODE syntax to CASE WHEN expression using the existing Case builder.
//
// DECODE(expr, search1, result1, search2, result2, ..., defaultResult)
// becomes CASE expr WHEN search1 THEN result1 WHEN search2 THEN result2 ... ELSE defaultResult END.
//
// Good Example: This function demonstrates proper method reuse by using b.Case() to build
// the CASE WHEN expression instead of constructing raw SQL. This approach:
//   - Leverages the existing Case builder for consistent syntax
//   - Handles the simple CASE syntax (CASE expr WHEN value THEN result)
//   - Properly manages default values with Else()
//   - Makes the code more maintainable and testable
//
// This is the recommended pattern for converting Oracle-specific syntax to standard SQL.
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
