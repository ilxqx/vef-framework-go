package orm

import (
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

// QueryExprBuilder implements the ExprBuilder interface, providing methods to build various SQL expressions.
type QueryExprBuilder struct {
	qb QueryBuilder
}

// Column builds a column expression with proper table alias handling.
func (b *QueryExprBuilder) Column(column string) schema.QueryAppender {
	dotIndex := strings.IndexByte(column, constants.ByteDot)
	if dotIndex > -1 {
		alias, name := column[:dotIndex], column[dotIndex+1:]
		if strings.IndexByte(alias, constants.ByteQuestionMark) == 0 {
			var sb strings.Builder
			sb.Grow(len(alias) + 2)
			_, _ = sb.WriteString(alias)
			_, _ = sb.WriteString(".?")

			return b.Expr(sb.String(), bun.Name(name))
		}

		return b.Expr("?.?", bun.Name(alias), bun.Name(name))
	}

	if b.qb.GetTable() != nil {
		return b.Expr("?TableAlias.?", bun.Name(column))
	}

	return b.Expr("?", bun.Name(column))
}

func (*QueryExprBuilder) Null() schema.QueryAppender {
	return bun.Safe(sqlNull)
}

func (b *QueryExprBuilder) IsNull(expr any) schema.QueryAppender {
	return b.Expr("? IS NULL", expr)
}

func (b *QueryExprBuilder) IsNotNull(expr any) schema.QueryAppender {
	return b.Expr("? IS NOT NULL", expr)
}

func (b *QueryExprBuilder) Literal(value any) schema.QueryAppender {
	return b.Expr("?", value)
}

func (b *QueryExprBuilder) Order(builder func(OrderBuilder)) schema.QueryAppender {
	ob := newOrderExpr(b)
	builder(ob)

	return ob
}

func (b *QueryExprBuilder) Case(builder func(CaseBuilder)) schema.QueryAppender {
	cb := newCaseExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) SubQuery(builder func(SelectQuery)) schema.QueryAppender {
	return b.Expr("(?)", b.qb.BuildSubQuery(builder))
}

func (b *QueryExprBuilder) Exists(builder func(SelectQuery)) schema.QueryAppender {
	return b.Expr("EXISTS (?)", b.qb.BuildSubQuery(builder))
}

func (b *QueryExprBuilder) NotExists(builder func(SelectQuery)) schema.QueryAppender {
	return b.Expr("NOT EXISTS (?)", b.qb.BuildSubQuery(builder))
}

func (b *QueryExprBuilder) Paren(expr any) schema.QueryAppender {
	return b.Expr("(?)", expr)
}

func (b *QueryExprBuilder) Not(expr any) schema.QueryAppender {
	return b.Expr("NOT (?)", expr)
}

func (b *QueryExprBuilder) Any(builder func(SelectQuery)) schema.QueryAppender {
	return b.Expr("ANY (?)", b.qb.BuildSubQuery(builder))
}

func (b *QueryExprBuilder) All(builder func(SelectQuery)) schema.QueryAppender {
	return b.Expr("ALL (?)", b.qb.BuildSubQuery(builder))
}

// ========== Arithmetic Operators ==========

func (b *QueryExprBuilder) Add(left, right any) schema.QueryAppender {
	return b.Expr("? + ?", left, right)
}

func (b *QueryExprBuilder) Subtract(left, right any) schema.QueryAppender {
	return b.Expr("? - ?", left, right)
}

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

func (b *QueryExprBuilder) Equals(left, right any) schema.QueryAppender {
	return b.Expr("? = ?", left, right)
}

func (b *QueryExprBuilder) NotEquals(left, right any) schema.QueryAppender {
	return b.Expr("? <> ?", left, right)
}

func (b *QueryExprBuilder) GreaterThan(left, right any) schema.QueryAppender {
	return b.Expr("? > ?", left, right)
}

func (b *QueryExprBuilder) GreaterThanOrEqual(left, right any) schema.QueryAppender {
	return b.Expr("? >= ?", left, right)
}

func (b *QueryExprBuilder) LessThan(left, right any) schema.QueryAppender {
	return b.Expr("? < ?", left, right)
}

func (b *QueryExprBuilder) LessThanOrEqual(left, right any) schema.QueryAppender {
	return b.Expr("? <= ?", left, right)
}

func (b *QueryExprBuilder) Between(expr, lower, upper any) schema.QueryAppender {
	return b.Expr("? BETWEEN ? AND ?", expr, lower, upper)
}

func (b *QueryExprBuilder) NotBetween(expr, lower, upper any) schema.QueryAppender {
	return b.Expr("? NOT BETWEEN ? AND ?", expr, lower, upper)
}

func (b *QueryExprBuilder) In(expr any, values ...any) schema.QueryAppender {
	return b.Expr("? IN (?)", expr, bun.In(values))
}

func (b *QueryExprBuilder) NotIn(expr any, values ...any) schema.QueryAppender {
	return b.Expr("? NOT IN (?)", expr, bun.In(values))
}

func (b *QueryExprBuilder) IsTrue(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite uses integers for booleans, check if equals 1
			return b.Equals(expr, 1)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("? IS TRUE", expr)
		},
	})
}

func (b *QueryExprBuilder) IsFalse(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite uses integers for booleans, check if equals 0
			return b.Equals(expr, 0)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("? IS FALSE", expr)
		},
	})
}

// ========== Expression Building ==========

func (b *QueryExprBuilder) Expr(expr string, args ...any) schema.QueryAppender {
	return bun.SafeQuery(expr, args...)
}

func (b *QueryExprBuilder) Exprs(exprs ...any) schema.QueryAppender {
	return newExpressions(constants.CommaSpace, exprs...)
}

func (b *QueryExprBuilder) ExprsWithSep(sep any, exprs ...any) schema.QueryAppender {
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

func (b *QueryExprBuilder) Count(builder func(CountBuilder)) schema.QueryAppender {
	cb := newCountExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) CountColumn(column string, distinct ...bool) schema.QueryAppender {
	return b.Count(func(cb CountBuilder) {
		if len(distinct) > 0 && distinct[0] {
			cb.Distinct()
		}

		cb.Column(column)
	})
}

func (b *QueryExprBuilder) CountAll(distinct ...bool) schema.QueryAppender {
	return b.Count(func(cb CountBuilder) {
		if len(distinct) > 0 && distinct[0] {
			cb.Distinct()
		}

		cb.All()
	})
}

func (b *QueryExprBuilder) Sum(builder func(SumBuilder)) schema.QueryAppender {
	cb := newSumExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) SumColumn(column string, distinct ...bool) schema.QueryAppender {
	return b.Sum(func(cb SumBuilder) {
		if len(distinct) > 0 && distinct[0] {
			cb.Distinct()
		}

		cb.Column(column)
	})
}

func (b *QueryExprBuilder) Avg(builder func(AvgBuilder)) schema.QueryAppender {
	cb := newAvgExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) AvgColumn(column string, distinct ...bool) schema.QueryAppender {
	return b.Avg(func(cb AvgBuilder) {
		if len(distinct) > 0 && distinct[0] {
			cb.Distinct()
		}

		cb.Column(column)
	})
}

func (b *QueryExprBuilder) Min(builder func(MinBuilder)) schema.QueryAppender {
	cb := newMinExpr(b.qb)
	builder(cb)

	return cb
}

func (b *QueryExprBuilder) MinColumn(column string) schema.QueryAppender {
	return b.Min(func(cb MinBuilder) {
		cb.Column(column)
	})
}

func (b *QueryExprBuilder) Max(builder func(MaxBuilder)) schema.QueryAppender {
	cb := newMaxExpr(b.qb)
	builder(cb)

	return cb
}

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

func (b *QueryExprBuilder) NTile(builder func(NTileBuilder)) schema.QueryAppender {
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

func (b *QueryExprBuilder) Concat(args ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			if len(args) == 0 {
				return b.Expr("?", constants.Empty)
			}

			if len(args) == 1 {
				return b.Expr("?", args[0])
			}

			return b.ExprsWithSep(" || ", args...)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("CONCAT(?)", b.Exprs(args...))
		},
	})
}

func (b *QueryExprBuilder) ConcatWithSep(separator any, args ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have CONCAT_WS
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
			return b.Expr("CONCAT_WS(?, ?)", separator, b.Exprs(args...))
		},
	})
}

func (b *QueryExprBuilder) SubString(expr, start any, length ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			if len(length) > 0 {
				return b.Expr("SUBSTR(?, ?, ?)", expr, start, length[0])
			}

			return b.Expr("SUBSTR(?, ?)", expr, start)
		},
		Default: func() schema.QueryAppender {
			if len(length) > 0 {
				return b.Expr("SUBSTRING(?, ?, ?)", expr, start, length[0])
			}

			return b.Expr("SUBSTRING(?, ?)", expr, start)
		},
	})
}

func (b *QueryExprBuilder) Upper(expr any) schema.QueryAppender {
	return b.Expr("UPPER(?)", expr)
}

func (b *QueryExprBuilder) Lower(expr any) schema.QueryAppender {
	return b.Expr("LOWER(?)", expr)
}

func (b *QueryExprBuilder) Trim(expr any) schema.QueryAppender {
	return b.Expr("TRIM(?)", expr)
}

func (b *QueryExprBuilder) TrimLeft(expr any) schema.QueryAppender {
	return b.Expr("LTRIM(?)", expr)
}

func (b *QueryExprBuilder) TrimRight(expr any) schema.QueryAppender {
	return b.Expr("RTRIM(?)", expr)
}

func (b *QueryExprBuilder) Length(expr any) schema.QueryAppender {
	return b.Expr("LENGTH(?)", expr)
}

func (b *QueryExprBuilder) CharLength(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have CHAR_LENGTH
			return b.Expr("LENGTH(?)", expr)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("CHAR_LENGTH(?)", expr)
		},
	})
}

func (b *QueryExprBuilder) Position(substring, str any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have POSITION IN
			return b.Expr("INSTR(?, ?)", str, substring)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("POSITION(? IN ?)", substring, str)
		},
	})
}

func (b *QueryExprBuilder) Left(expr, length any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have LEFT
			return b.SubString(expr, 1, length)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("LEFT(?, ?)", expr, length)
		},
	})
}

func (b *QueryExprBuilder) Right(expr, length any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have RIGHT
			return b.Expr("SUBSTR(?, -?)", expr, length)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("RIGHT(?, ?)", expr, length)
		},
	})
}

func (b *QueryExprBuilder) Repeat(expr, count any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have REPEAT, need to implement with REPLACE and a helper
			// Use REPLACE(SUBSTR(QUOTE(ZEROBLOB((count+1)/2)), 3, count), '0', expr)
			return b.Expr("REPLACE(SUBSTR(QUOTE(ZEROBLOB(?)), 3, ?), ?, ?)", b.Divide(b.Paren(b.Add(count, 1)), 2), count, "0", expr)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("REPEAT(?, ?)", expr, count)
		},
	})
}

func (b *QueryExprBuilder) Replace(expr, search, replacement any) schema.QueryAppender {
	return b.Expr("REPLACE(?, ?, ?)", expr, search, replacement)
}

// fuzzyMatch builds a fuzzy LIKE expression based on FuzzyKind.
// It optimizes for string literals by pre-building the LIKE pattern in Go code,
// avoiding runtime string concatenation in the database.
func (b *QueryExprBuilder) fuzzyMatch(expr, pattern any, kind FuzzyKind, ignoreCase bool) schema.QueryAppender {
	// Optimize for string literals: build pattern in Go
	if strPattern, ok := pattern.(string); ok {
		likePattern := kind.BuildPattern(strPattern)
		if ignoreCase {
			return b.ExprByDialect(DialectExprs{
				Postgres: func() schema.QueryAppender {
					return b.Expr("? ILIKE ?", expr, likePattern)
				},
				Default: func() schema.QueryAppender {
					return b.Expr("? LIKE ?", b.Lower(expr), strings.ToLower(likePattern))
				},
			})
		}

		return b.Expr("? LIKE ?", expr, likePattern)
	}

	// For dynamic expressions, build pattern using b.Concat()
	var likePattern any
	switch kind {
	case FuzzyContains:
		likePattern = b.Concat("%", pattern, "%")
	case FuzzyStarts:
		likePattern = b.Concat(pattern, "%")
	case FuzzyEnds:
		likePattern = b.Concat("%", pattern)
	default:
		likePattern = pattern
	}

	if ignoreCase {
		return b.ExprByDialect(DialectExprs{
			Postgres: func() schema.QueryAppender {
				return b.Expr("? ILIKE ?", expr, likePattern)
			},
			Default: func() schema.QueryAppender {
				return b.Expr("? LIKE ?", b.Lower(expr), b.Lower(likePattern))
			},
		})
	}

	return b.Expr("? LIKE ?", expr, likePattern)
}

func (b *QueryExprBuilder) Contains(expr, substr any) schema.QueryAppender {
	return b.fuzzyMatch(expr, substr, FuzzyContains, false)
}

func (b *QueryExprBuilder) StartsWith(expr, prefix any) schema.QueryAppender {
	return b.fuzzyMatch(expr, prefix, FuzzyStarts, false)
}

func (b *QueryExprBuilder) EndsWith(expr, suffix any) schema.QueryAppender {
	return b.fuzzyMatch(expr, suffix, FuzzyEnds, false)
}

func (b *QueryExprBuilder) ContainsIgnoreCase(expr, substr any) schema.QueryAppender {
	return b.fuzzyMatch(expr, substr, FuzzyContains, true)
}

func (b *QueryExprBuilder) StartsWithIgnoreCase(expr, prefix any) schema.QueryAppender {
	return b.fuzzyMatch(expr, prefix, FuzzyStarts, true)
}

func (b *QueryExprBuilder) EndsWithIgnoreCase(expr, suffix any) schema.QueryAppender {
	return b.fuzzyMatch(expr, suffix, FuzzyEnds, true)
}

func (b *QueryExprBuilder) Reverse(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have REVERSE function, need complex implementation
			// For now, return unsupported
			return b.Expr("?", expr)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("REVERSE(?)", expr)
		},
	})
}

// ========== Date and Time Functions ==========

func (b *QueryExprBuilder) CurrentDate() schema.QueryAppender {
	return b.Expr("CURRENT_DATE")
}

func (b *QueryExprBuilder) CurrentTime() schema.QueryAppender {
	return b.Expr("CURRENT_TIME")
}

func (b *QueryExprBuilder) CurrentTimestamp() schema.QueryAppender {
	return b.Expr("CURRENT_TIMESTAMP")
}

func (b *QueryExprBuilder) Now() schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("NOW()")
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("NOW()")
		},
		SQLite: func() schema.QueryAppender {
			return b.Expr("DATETIME('now')")
		},
		Default: func() schema.QueryAppender {
			return b.Expr("NOW()")
		},
	})
}

func (b *QueryExprBuilder) ExtractYear(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("EXTRACT(YEAR FROM ?)", expr)
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("EXTRACT(YEAR FROM ?)", expr)
		},
		SQLite: func() schema.QueryAppender {
			return b.ToInteger(
				b.Expr("STRFTIME(?, ?)", "%Y", expr),
			)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("EXTRACT(YEAR FROM ?)", expr)
		},
	})
}

func (b *QueryExprBuilder) ExtractMonth(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("EXTRACT(MONTH FROM ?)", expr)
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("EXTRACT(MONTH FROM ?)", expr)
		},
		SQLite: func() schema.QueryAppender {
			return b.ToInteger(
				b.Expr("STRFTIME(?, ?)", "%m", expr),
			)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("EXTRACT(MONTH FROM ?)", expr)
		},
	})
}

func (b *QueryExprBuilder) ExtractDay(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("EXTRACT(DAY FROM ?)", expr)
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("EXTRACT(DAY FROM ?)", expr)
		},
		SQLite: func() schema.QueryAppender {
			return b.ToInteger(
				b.Expr("STRFTIME(?, ?)", "%d", expr),
			)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("EXTRACT(DAY FROM ?)", expr)
		},
	})
}

func (b *QueryExprBuilder) ExtractHour(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("EXTRACT(HOUR FROM ?)", expr)
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("EXTRACT(HOUR FROM ?)", expr)
		},
		SQLite: func() schema.QueryAppender {
			return b.ToInteger(
				b.Expr("STRFTIME(?, ?)", "%H", expr),
			)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("EXTRACT(HOUR FROM ?)", expr)
		},
	})
}

func (b *QueryExprBuilder) ExtractMinute(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("EXTRACT(MINUTE FROM ?)", expr)
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("EXTRACT(MINUTE FROM ?)", expr)
		},
		SQLite: func() schema.QueryAppender {
			return b.ToInteger(
				b.Expr("STRFTIME(?, ?)", "%M", expr),
			)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("EXTRACT(MINUTE FROM ?)", expr)
		},
	})
}

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

func (b *QueryExprBuilder) DateTrunc(unit DateTimeUnit, expr any) schema.QueryAppender {
	precision := unit.ForDateTrunc()

	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("DATE_TRUNC(?, ?)", precision, expr)
		},
		MySQL: func() schema.QueryAppender {
			switch unit {
			case UnitYear:
				return b.Expr("DATE_FORMAT(?, ?)", expr, "%Y-01-01")
			case UnitMonth:
				return b.Expr("DATE_FORMAT(?, ?)", expr, "%Y-%m-01")
			case UnitDay:
				return b.Expr("DATE(?)", expr)
			case UnitHour:
				return b.Expr("DATE_FORMAT(?, ?)", expr, "%Y-%m-%d %H:00:00")
			case UnitMinute:
				return b.Expr("DATE_FORMAT(?, ?)", expr, "%Y-%m-%d %H:%i:00")
			default:
				return b.Expr("DATE(?)", expr)
			}
		},
		SQLite: func() schema.QueryAppender {
			switch unit {
			case UnitYear:
				return b.Expr("STRFTIME(?, ?)", "%Y-01-01", expr)
			case UnitMonth:
				return b.Expr("STRFTIME(?, ?)", "%Y-%m-01", expr)
			case UnitDay:
				return b.Expr("DATE(?)", expr)
			case UnitHour:
				return b.Expr("STRFTIME(?, ?)", "%Y-%m-%d %H:00:00", expr)
			case UnitMinute:
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

func (b *QueryExprBuilder) DateAdd(expr, interval any, unit DateTimeUnit) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("? + INTERVAL '? ?'", expr, interval, b.Expr(unit.ForPostgres()))
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("DATE_ADD(?, INTERVAL ? ?)", expr, interval, b.Expr(unit.ForMySQL()))
		},
		SQLite: func() schema.QueryAppender {
			return b.Expr("DATETIME(?, '+? ?')", expr, interval, b.Expr(unit.ForSQLite()))
		},
		Default: func() schema.QueryAppender {
			return b.Expr("DATE_ADD(?, INTERVAL ? ?)", expr, interval, b.Expr(unit.String()))
		},
	})
}

func (b *QueryExprBuilder) DateSubtract(expr, interval any, unit DateTimeUnit) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("? - INTERVAL '? ?'", expr, interval, b.Expr(unit.ForPostgres()))
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("DATE_SUB(?, INTERVAL ? ?)", expr, interval, b.Expr(unit.ForMySQL()))
		},
		SQLite: func() schema.QueryAppender {
			return b.Expr("DATETIME(?, '-? ?')", expr, interval, b.Expr(unit.ForSQLite()))
		},
		Default: func() schema.QueryAppender {
			return b.Expr("DATE_SUB(?, INTERVAL ? ?)", expr, interval, b.Expr(unit.String()))
		},
	})
}

func (b *QueryExprBuilder) DateDiff(start, end any, unit DateTimeUnit) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL: use EXTRACT(EPOCH FROM ...) for accurate calculations
			// EPOCH returns total seconds, then divide by unit
			epochDiff := b.Expr("EXTRACT(EPOCH FROM ?)", b.Subtract(end, start))

			switch unit {
			case UnitSecond:
				return epochDiff
			case UnitMinute:
				return b.Divide(epochDiff, 60)
			case UnitHour:
				return b.Divide(epochDiff, 3600)
			case UnitDay:
				return b.Divide(epochDiff, 86400)
			case UnitMonth:
				// Approximate: 30.44 days per month on average
				return b.Divide(epochDiff, 2629800)
			case UnitYear:
				// Approximate: 365.25 days per year on average
				return b.Divide(epochDiff, 31557600)
			default:
				return b.Divide(epochDiff, 86400) // default to days
			}
		},
		MySQL: func() schema.QueryAppender {
			switch unit {
			case UnitDay:
				return b.ToDecimal(b.Expr("DATEDIFF(?, ?)", end, start), 20, 6)
			default:
				return b.ToDecimal(b.Expr("TIMESTAMPDIFF(?, ?, ?)", b.Expr(unit.ForMySQL()), start, end), 20, 6)
			}
		},
		SQLite: func() schema.QueryAppender {
			switch unit {
			case UnitSecond:
				// (JULIANDAY difference) * 86400 seconds per day
				return b.Multiply(
					b.Paren(b.Subtract(b.Expr("JULIANDAY(?)", end), b.Expr("JULIANDAY(?)", start))),
					86400,
				)

			case UnitMinute:
				// (JULIANDAY difference) * 1440 minutes per day
				return b.Multiply(
					b.Paren(b.Subtract(b.Expr("JULIANDAY(?)", end), b.Expr("JULIANDAY(?)", start))),
					1440,
				)

			case UnitHour:
				// (JULIANDAY difference) * 24 hours per day
				return b.Multiply(
					b.Paren(b.Subtract(b.Expr("JULIANDAY(?)", end), b.Expr("JULIANDAY(?)", start))),
					24,
				)

			case UnitDay:
				return b.Subtract(b.Expr("JULIANDAY(?)", end), b.Expr("JULIANDAY(?)", start))
			case UnitMonth:
				// Approximate: (JULIANDAY difference) / 30.44 days per month
				return b.Divide(
					b.Paren(b.Subtract(b.Expr("JULIANDAY(?)", end), b.Expr("JULIANDAY(?)", start))),
					30.44,
				)

			case UnitYear:
				// Approximate: (JULIANDAY difference) / 365.25 days per year
				return b.Divide(
					b.Paren(b.Subtract(b.Expr("JULIANDAY(?)", end), b.Expr("JULIANDAY(?)", start))),
					365.25,
				)

			default:
				return b.Subtract(b.Expr("JULIANDAY(?)", end), b.Expr("JULIANDAY(?)", start))
			}
		},
		Default: func() schema.QueryAppender {
			return b.Expr("DATEDIFF(?, ?, ?)", b.Expr(unit.String()), end, start)
		},
	})
}

// Age returns the age (interval) between two timestamps.
// Returns a PostgreSQL-compatible interval string in format: "X years Y mons Z days"
// This provides a symbolic result using field-by-field subtraction with adjustments.
func (b *QueryExprBuilder) Age(start, end any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL's AGE() returns simplified format for small intervals (e.g., "1 mon" instead of "0 years 1 mons 0 days")
			// We need to ensure consistent formatting by extracting years, months, and days components explicitly
			ageInterval := b.Expr("AGE(?, ?)", end, start)

			years := b.ExtractYear(ageInterval)
			months := b.ExtractMonth(ageInterval)
			days := b.ExtractDay(ageInterval)

			return b.Concat(
				b.ToString(years), " years ",
				b.ToString(months), " mons ",
				b.ToString(days), " days",
			)
		},
		MySQL: func() schema.QueryAppender {
			years := b.ToInteger(b.DateDiff(start, end, UnitYear))

			// Dynamic value, can't use DateAdd
			startPlusYears := b.DateAdd(start, years, UnitYear)

			// Calculate remaining months
			months := b.ToInteger(b.DateDiff(startPlusYears, end, UnitMonth))

			// Dynamic value, can't use DateAdd
			startPlusYearsMonths := b.DateAdd(startPlusYears, months, UnitMonth)

			// Calculate remaining days
			days := b.ToInteger(b.DateDiff(startPlusYearsMonths, end, UnitDay))

			return b.Concat(
				years, " years ",
				months, " mons ",
				days, " days",
			)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have AGE function, so we emulate it
			approxYears := b.Subtract(
				b.ToInteger(b.ExtractYear(end)),
				b.ToInteger(b.ExtractYear(start)),
			)

			// Dynamic value requires raw expression
			startPlusYears := b.Expr(`DATE(?, '+' || CAST(? AS TEXT) || ' years')`, start, approxYears)
			actualYears := b.Expr(`CASE WHEN ? > ? THEN (?) - 1 ELSE ? END`, startPlusYears, end, approxYears, approxYears)
			startPlusActualYears := b.Expr(`DATE(?, '+' || CAST(? AS TEXT) || ' years')`, start, actualYears)

			approxMonths := b.Add(
				b.Multiply(
					b.Paren(b.Subtract(
						b.ToInteger(b.ExtractYear(end)),
						b.ToInteger(b.ExtractYear(startPlusActualYears)),
					)),
					12,
				),
				b.Subtract(
					b.ToInteger(b.ExtractMonth(end)),
					b.ToInteger(b.ExtractMonth(startPlusActualYears)),
				),
			)

			startPlusYearsMonths := b.Expr(`DATE(?, '+' || CAST(? AS TEXT) || ' months')`, startPlusActualYears, approxMonths)
			actualMonths := b.Expr(`CASE WHEN ? > ? THEN (?) - 1 ELSE ? END`, startPlusYearsMonths, end, approxMonths, approxMonths)
			startPlusActualYearsMonths := b.Expr(`DATE(?, '+' || CAST(? AS TEXT) || ' years', '+' || CAST(? AS TEXT) || ' months')`,
				start, actualYears, actualMonths)

			days := b.ToInteger(
				b.Subtract(
					b.Expr("JULIANDAY(?)", end),
					b.Expr("JULIANDAY(?)", startPlusActualYearsMonths),
				),
			)

			return b.Concat(
				b.ToString(actualYears), " years ",
				b.ToString(actualMonths), " mons ",
				b.ToString(days), " days",
			)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("AGE(?, ?)", end, start)
		},
	})
}

// ========== Math Functions ==========

func (b *QueryExprBuilder) Abs(expr any) schema.QueryAppender {
	return b.Expr("ABS(?)", expr)
}

func (b *QueryExprBuilder) Ceil(expr any) schema.QueryAppender {
	return b.Expr("CEIL(?)", expr)
}

func (b *QueryExprBuilder) Floor(expr any) schema.QueryAppender {
	return b.Expr("FLOOR(?)", expr)
}

func (b *QueryExprBuilder) Round(expr any, precision ...any) schema.QueryAppender {
	if len(precision) > 0 {
		return b.Expr("ROUND(?, ?)", expr, precision[0])
	}

	return b.Expr("ROUND(?)", expr)
}

func (b *QueryExprBuilder) Trunc(expr any, precision ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			if len(precision) > 0 {
				// For positive precision, ROUND works similarly to TRUNC for most cases
				return b.Round(expr, precision[0])
			}

			return b.ToInteger(expr)
		},
		MySQL: func() schema.QueryAppender {
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

func (b *QueryExprBuilder) Power(base, exponent any) schema.QueryAppender {
	return b.Expr("POWER(?, ?)", base, exponent)
}

func (b *QueryExprBuilder) Sqrt(expr any) schema.QueryAppender {
	return b.Expr("SQRT(?)", expr)
}

func (b *QueryExprBuilder) Exp(expr any) schema.QueryAppender {
	return b.Expr("EXP(?)", expr)
}

func (b *QueryExprBuilder) Ln(expr any) schema.QueryAppender {
	return b.Expr("LN(?)", expr)
}

func (b *QueryExprBuilder) Log(expr any, base ...any) schema.QueryAppender {
	if len(base) > 0 {
		return b.Expr("LOG(?, ?)", base[0], expr)
	}

	return b.Expr("LOG(?)", expr)
}

func (b *QueryExprBuilder) Sin(expr any) schema.QueryAppender {
	return b.Expr("SIN(?)", expr)
}

func (b *QueryExprBuilder) Cos(expr any) schema.QueryAppender {
	return b.Expr("COS(?)", expr)
}

func (b *QueryExprBuilder) Tan(expr any) schema.QueryAppender {
	return b.Expr("TAN(?)", expr)
}

func (b *QueryExprBuilder) Asin(expr any) schema.QueryAppender {
	return b.Expr("ASIN(?)", expr)
}

func (b *QueryExprBuilder) Acos(expr any) schema.QueryAppender {
	return b.Expr("ACOS(?)", expr)
}

func (b *QueryExprBuilder) Atan(expr any) schema.QueryAppender {
	return b.Expr("ATAN(?)", expr)
}

func (b *QueryExprBuilder) Pi() schema.QueryAppender {
	return b.Expr("PI()")
}

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
			return b.Expr("RAND()")
		},
		Default: func() schema.QueryAppender {
			return b.Expr("RANDOM()")
		},
	})
}

func (b *QueryExprBuilder) Sign(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
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

func (b *QueryExprBuilder) Mod(dividend, divisor any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			return b.Expr("? % ?", dividend, divisor)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("MOD(?, ?)", dividend, divisor)
		},
	})
}

func (b *QueryExprBuilder) Greatest(args ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			return b.Expr("MAX(?)", newExpressions(constants.CommaSpace, args...))
		},
		Default: func() schema.QueryAppender {
			return b.Expr("GREATEST(?)", newExpressions(constants.CommaSpace, args...))
		},
	})
}

func (b *QueryExprBuilder) Least(args ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		SQLite: func() schema.QueryAppender {
			return b.Expr("MIN(?)", newExpressions(constants.CommaSpace, args...))
		},
		Default: func() schema.QueryAppender {
			return b.Expr("LEAST(?)", newExpressions(constants.CommaSpace, args...))
		},
	})
}

// ========== Conditional Functions ==========

func (b *QueryExprBuilder) Coalesce(args ...any) schema.QueryAppender {
	return b.Expr("COALESCE(?)", newExpressions(constants.CommaSpace, args...))
}

func (b *QueryExprBuilder) NullIf(expr1, expr2 any) schema.QueryAppender {
	return b.Expr("NULLIF(?, ?)", expr1, expr2)
}

func (b *QueryExprBuilder) IfNull(expr, defaultValue any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL doesn't have IFNULL, use COALESCE instead
			return b.Coalesce(expr, defaultValue)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("IFNULL(?, ?)", expr, defaultValue)
		},
	})
}

// ========== Type Conversion Functions ==========

func (b *QueryExprBuilder) ToString(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("?::TEXT", b.Paren(expr))
		},
		MySQL: func() schema.QueryAppender {
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

func (b *QueryExprBuilder) ToInteger(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("?::INTEGER", b.Paren(expr))
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("CAST(? AS SIGNED INTEGER)", expr)
		},
		SQLite: func() schema.QueryAppender {
			return b.Expr("CAST(? AS INTEGER)", expr)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("CAST(? AS INTEGER)", expr)
		},
	})
}

func (b *QueryExprBuilder) ToDecimal(expr any, precision ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			if len(precision) >= 2 {
				return b.Expr("?::NUMERIC(?, ?)", b.Paren(expr), precision[0], precision[1])
			} else if len(precision) == 1 {
				return b.Expr("?::NUMERIC(?)", b.Paren(expr), precision[0])
			}

			return b.Expr("?::NUMERIC", b.Paren(expr))
		},
		MySQL: func() schema.QueryAppender {
			if len(precision) >= 2 {
				return b.Expr("CAST(? AS DECIMAL(?, ?))", expr, precision[0], precision[1])
			} else if len(precision) == 1 {
				return b.Expr("CAST(? AS DECIMAL(?))", expr, precision[0])
			}

			return b.Expr("CAST(? AS DECIMAL)", expr)
		},
		SQLite: func() schema.QueryAppender {
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

func (b *QueryExprBuilder) ToFloat(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("?::DOUBLE PRECISION", b.Paren(expr))
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("CAST(? AS DOUBLE)", expr)
		},
		SQLite: func() schema.QueryAppender {
			return b.Expr("CAST(? AS REAL)", expr)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("CAST(? AS DOUBLE)", expr)
		},
	})
}

func (b *QueryExprBuilder) ToBool(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("?::BOOLEAN", b.Paren(expr))
		},
		MySQL: func() schema.QueryAppender {
			return b.NotEquals(b.ToInteger(expr), 0)
		},
		SQLite: func() schema.QueryAppender {
			return b.NotEquals(b.ToInteger(expr), 0)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("CAST(? AS BOOLEAN)", expr)
		},
	})
}

func (b *QueryExprBuilder) ToDate(expr any, format ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			if len(format) > 0 {
				return b.Expr("TO_DATE(?, ?)", expr, format[0])
			}

			return b.Expr("?::DATE", b.Paren(expr))
		},
		MySQL: func() schema.QueryAppender {
			if len(format) > 0 {
				return b.Expr("STR_TO_DATE(?, ?)", expr, format[0])
			}

			return b.Expr("CAST(? AS DATE)", expr)
		},
		SQLite: func() schema.QueryAppender {
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

func (b *QueryExprBuilder) ToTime(expr any, format ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			if len(format) > 0 {
				return b.Expr("TO_TIMESTAMP(?, ?)::TIME", expr, format[0])
			}

			return b.Expr("?::TIME", b.Paren(expr))
		},
		MySQL: func() schema.QueryAppender {
			if len(format) > 0 {
				return b.Expr("TIME(STR_TO_DATE(?, ?))", expr, format[0])
			}

			return b.Expr("CAST(? AS TIME)", expr)
		},
		SQLite: func() schema.QueryAppender {
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

func (b *QueryExprBuilder) ToTimestamp(expr any, format ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			if len(format) > 0 {
				return b.Expr("TO_TIMESTAMP(?, ?)", expr, format[0])
			}

			return b.Expr("?::TIMESTAMP", b.Paren(expr))
		},
		MySQL: func() schema.QueryAppender {
			if len(format) > 0 {
				return b.Expr("STR_TO_DATE(?, ?)", expr, format[0])
			}

			return b.Expr("CAST(? AS DATETIME)", expr)
		},
		SQLite: func() schema.QueryAppender {
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

func (b *QueryExprBuilder) ToJson(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("?::JSONB", b.Paren(expr))
		},
		MySQL: func() schema.QueryAppender {
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

// processJsonPath formats the JSON path according to the dialect requirements.
// Input path is expected to be in "key1.key2" format (dot-separated keys).
func (b *QueryExprBuilder) processJsonPath(path any, dialectName dialect.Name) any {
	switch dialectName {
	case dialect.PG:
		if pathStr, ok := path.(string); ok {
			// Split by dot and join with comma
			parts := strings.Split(pathStr, constants.Dot)

			return constants.LeftBrace + strings.Join(parts, constants.Comma) + constants.RightBrace
		}
		// For expressions, we replace . with , and wrap in {}
		// Note: This assumes the expression evaluates to a dot-separated string
		return b.Concat(constants.LeftBrace, b.Replace(path, constants.Dot, constants.Comma), constants.RightBrace)

	default: // MySQL, SQLite
		if pathStr, ok := path.(string); ok {
			if len(pathStr) == 0 {
				return constants.Dollar
			}

			return constants.Dollar + constants.Dot + pathStr
		}

		return b.Concat(constants.Dollar+constants.Dot, path)
	}
}

// JsonExtract extracts value from JSON at specified path.
//
// Note: For PostgreSQL, this uses the #>> operator which returns the value as text (unquoted).
// This is compatible with b.ToJson() which will correctly cast the text back to JSONB
// if the result is used in subsequent JSON functions (e.g. chaining JsonExtract).
func (b *QueryExprBuilder) JsonExtract(json, path any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("? #>> ?::text[]", b.ToJson(json), b.processJsonPath(path, dialect.PG))
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("JSON_EXTRACT(?, ?)", json, b.processJsonPath(path, dialect.MySQL))
		},
		SQLite: func() schema.QueryAppender {
			return b.Expr("JSON_EXTRACT(?, ?)", json, b.processJsonPath(path, dialect.SQLite))
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_EXTRACT(?, ?)", json, b.processJsonPath(path, dialect.MySQL))
		},
	})
}

func (b *QueryExprBuilder) JsonUnquote(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL doesn't have JSON_UNQUOTE
			// JsonExtract already uses ->> which returns text (unquoted)
			// So we just return the expression as-is
			return b.Expr("?", expr)
		},
		MySQL: func() schema.QueryAppender {
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

func (b *QueryExprBuilder) JsonArray(args ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			if len(args) == 0 {
				return b.ToJson("[]")
			}

			return b.Expr("JSONB_BUILD_ARRAY(?)", b.Exprs(args...))
		},
		MySQL: func() schema.QueryAppender {
			if len(args) == 0 {
				return b.Expr("JSON_ARRAY()")
			}

			return b.Expr("JSON_ARRAY(?)", b.Exprs(args...))
		},
		SQLite: func() schema.QueryAppender {
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

func (b *QueryExprBuilder) JsonObject(keyValues ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			if len(keyValues) == 0 {
				return b.ToJson("{}")
			}

			return b.Expr("JSONB_BUILD_OBJECT(?)", b.Exprs(keyValues...))
		},
		MySQL: func() schema.QueryAppender {
			if len(keyValues) == 0 {
				return b.Expr("JSON_OBJECT()")
			}

			return b.Expr("JSON_OBJECT(?)", b.Exprs(keyValues...))
		},
		SQLite: func() schema.QueryAppender {
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

func (b *QueryExprBuilder) JsonContains(json, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("? @> ?", json, b.ToJson(value))
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("JSON_CONTAINS(?, ?)", json, value)
		},
		SQLite: func() schema.QueryAppender {
			return b.Exists(func(sq SelectQuery) {
				sq.SelectExpr(func(eb ExprBuilder) any { return eb.Literal(1) }).
					Where(func(cb ConditionBuilder) {
						cb.Expr(func(eb ExprBuilder) any {
							return eb.Equals(
								eb.JsonExtract(json, "$[*]"),
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

func (b *QueryExprBuilder) JsonContainsPath(json, path any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL uses jsonb_path_exists which requires JSONPath syntax ($.key)
			// We use the same path formatting as MySQL
			return b.Expr("JSONB_PATH_EXISTS(?, ?::jsonpath)", b.ToJson(json), b.processJsonPath(path, dialect.MySQL))
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("JSON_CONTAINS_PATH(?, ?, ?)", json, "one", b.processJsonPath(path, dialect.MySQL))
		},
		SQLite: func() schema.QueryAppender {
			return b.IsNotNull(b.JsonExtract(json, path))
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_CONTAINS_PATH(?, ?, ?)", json, "one", b.processJsonPath(path, dialect.MySQL))
		},
	})
}

func (b *QueryExprBuilder) JsonKeys(json any, path ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			if len(path) > 0 {
				return b.Expr("JSONB_OBJECT_KEYS(? #> ?::text[])", b.ToJson(json), b.processJsonPath(path[0], dialect.PG))
			}

			return b.Expr("JSONB_OBJECT_KEYS(?)", b.ToJson(json))
		},
		MySQL: func() schema.QueryAppender {
			if len(path) > 0 {
				return b.Expr("JSON_KEYS(?, ?)", json, b.processJsonPath(path[0], dialect.MySQL))
			}

			return b.Expr("JSON_KEYS(?)", json)
		},
		SQLite: func() schema.QueryAppender {
			buildKeyArray := func(source any) schema.QueryAppender {
				return b.SubQuery(func(sq SelectQuery) {
					sq.TableExpr(
						func(eb ExprBuilder) any {
							return eb.Expr("JSON_EACH(?)", source)
						}).
						SelectExpr(func(eb ExprBuilder) any {
							return eb.Expr("JSON_GROUP_ARRAY(key)")
						})
				})
			}

			if len(path) > 0 {
				return buildKeyArray(b.JsonExtract(json, path[0]))
			}

			return buildKeyArray(json)
		},
		Default: func() schema.QueryAppender {
			if len(path) > 0 {
				return b.Expr("JSON_KEYS(?, ?)", json, b.processJsonPath(path[0], dialect.MySQL))
			}

			return b.Expr("JSON_KEYS(?)", json)
		},
	})
}

func (b *QueryExprBuilder) JsonLength(json any, path ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL: Support both arrays and objects by checking type
			// For arrays: use JSONB_ARRAY_LENGTH
			// For objects: count keys using JSONB_OBJECT_KEYS
			// For other types: return 0
			var jsonExpr schema.QueryAppender
			if len(path) > 0 {
				// With path: extract the value at path first
				jsonExpr = b.Expr("? #> ?::text[]", b.ToJson(json), b.processJsonPath(path[0], dialect.PG))
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
				return b.Expr("JSON_LENGTH(?, ?)", json, b.processJsonPath(path[0], dialect.MySQL))
			}

			return b.Expr("JSON_LENGTH(?)", json)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite: Support both arrays and objects by checking type
			// For arrays: use JSON_ARRAY_LENGTH
			// For objects: count keys using JSON_EACH
			// For other types: return 0
			valueExpr := json
			if len(path) > 0 {
				valueExpr = b.JsonExtract(json, path[0])
			}

			return b.Case(func(cb CaseBuilder) {
				cb.Case(b.JsonType(json, path...)).
					WhenExpr("array").
					Then(b.Expr("JSON_ARRAY_LENGTH(?)", valueExpr)).
					WhenExpr("object").
					ThenSubQuery(func(query SelectQuery) {
						query.SelectExpr(func(eb ExprBuilder) any { return eb.CountAll() }).
							TableExpr(func(eb ExprBuilder) any {
								return eb.Expr("JSON_EACH(?)", valueExpr)
							})
					}).
					Else(0)
			})
		},
		Default: func() schema.QueryAppender {
			if len(path) > 0 {
				return b.Expr("JSON_LENGTH(?, ?)", json, b.processJsonPath(path[0], dialect.MySQL))
			}

			return b.Expr("JSON_LENGTH(?)", json)
		},
	})
}

func (b *QueryExprBuilder) JsonType(json any, path ...any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			if len(path) > 0 {
				return b.Expr("JSONB_TYPEOF(? #> ?::text[])", b.ToJson(json), b.processJsonPath(path[0], dialect.PG))
			}

			return b.Expr("JSONB_TYPEOF(?)", b.ToJson(json))
		},
		MySQL: func() schema.QueryAppender {
			if len(path) > 0 {
				return b.Expr("JSON_TYPE(?, ?)", json, b.processJsonPath(path[0], dialect.MySQL))
			}

			return b.Expr("JSON_TYPE(?)", json)
		},
		SQLite: func() schema.QueryAppender {
			if len(path) > 0 {
				return b.Expr("JSON_TYPE(?, ?)", json, b.processJsonPath(path[0], dialect.SQLite))
			}

			return b.Expr("JSON_TYPE(?)", json)
		},
		Default: func() schema.QueryAppender {
			if len(path) > 0 {
				return b.Expr("JSON_TYPE(?, ?)", json, b.processJsonPath(path[0], dialect.MySQL))
			}

			return b.Expr("JSON_TYPE(?)", json)
		},
	})
}

func (b *QueryExprBuilder) JsonValid(expr any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("? AND ?", b.IsNotNull(expr), b.IsNotNull(b.ToJson(b.ToString(expr))))
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("JSON_VALID(?)", expr)
		},
		SQLite: func() schema.QueryAppender {
			return b.Expr("JSON_VALID(?)", expr)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_VALID(?)", expr)
		},
	})
}

func (b *QueryExprBuilder) JsonSet(json, path, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("JSONB_SET(?, ?::text[], ?, TRUE)", b.ToJson(json), b.processJsonPath(path, dialect.PG), b.ToJson(value))
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("JSON_SET(?, ?, ?)", json, b.processJsonPath(path, dialect.MySQL), value)
		},
		SQLite: func() schema.QueryAppender {
			return b.Expr("JSON_SET(?, ?, ?)", json, b.processJsonPath(path, dialect.SQLite), value)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_SET(?, ?, ?)", json, b.processJsonPath(path, dialect.MySQL), value)
		},
	})
}

func (b *QueryExprBuilder) JsonInsert(json, path, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("JSONB_INSERT(?, ?::text[], TO_JSONB(?), FALSE)", b.ToJson(json), b.processJsonPath(path, dialect.PG), value)
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("JSON_INSERT(?, ?, ?)", json, b.processJsonPath(path, dialect.MySQL), value)
		},
		SQLite: func() schema.QueryAppender {
			return b.Expr("JSON_INSERT(?, ?, ?)", json, b.processJsonPath(path, dialect.SQLite), value)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_INSERT(?, ?, ?)", json, b.processJsonPath(path, dialect.MySQL), value)
		},
	})
}

func (b *QueryExprBuilder) JsonReplace(json, path, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.Expr("JSONB_SET(?, ?::text[], TO_JSONB(?), FALSE)", b.ToJson(json), b.processJsonPath(path, dialect.PG), value)
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("JSON_REPLACE(?, ?, ?)", json, b.processJsonPath(path, dialect.MySQL), value)
		},
		SQLite: func() schema.QueryAppender {
			return b.Expr("JSON_REPLACE(?, ?, ?)", json, b.processJsonPath(path, dialect.SQLite), value)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_REPLACE(?, ?, ?)", json, b.processJsonPath(path, dialect.MySQL), value)
		},
	})
}

func (b *QueryExprBuilder) JsonArrayAppend(json, path, value any) schema.QueryAppender {
	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			if pathStr, ok := path.(string); ok && (pathStr == "$" || pathStr == constants.Dollar) {
				// Root level array
				return b.Expr("(? || ?)", b.ToJson(json), b.JsonArray(value))
			}

			pgPath := b.processJsonPath(path, dialect.PG)

			// Nested array - use jsonb_set with concatenation
			return b.Expr(
				"JSONB_SET(?, ?::text[], (? || ?))",
				b.ToJson(json),
				pgPath,
				b.Coalesce(
					b.Expr("? #> ?::text[]", b.ToJson(json), pgPath),
					b.ToJson("[]"),
				),
				b.JsonArray(value),
			)
		},
		MySQL: func() schema.QueryAppender {
			return b.Expr("JSON_ARRAY_APPEND(?, ?, ?)", json, b.processJsonPath(path, dialect.MySQL), value)
		},
		SQLite: func() schema.QueryAppender {
			// Use helper composition to reuse JSON path logic
			return b.JsonSet(
				json,
				path,
				b.JsonInsert(
					b.Coalesce(
						b.JsonExtract(json, path),
						"[]",
					),
					"$[#]",
					value,
				),
			)
		},
		Default: func() schema.QueryAppender {
			return b.Expr("JSON_ARRAY_APPEND(?, ?, ?)", json, b.processJsonPath(path, dialect.MySQL), value)
		},
	})
}

// ========== Utility Functions ==========

func (b *QueryExprBuilder) Decode(args ...any) schema.QueryAppender {
	if len(args) < 3 {
		return b.Null()
	}

	return b.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			// PostgreSQL doesn't have DECODE
			return b.convertDecodeToCase(args...)
		},
		MySQL: func() schema.QueryAppender {
			// MySQL doesn't have DECODE
			return b.convertDecodeToCase(args...)
		},
		SQLite: func() schema.QueryAppender {
			// SQLite doesn't have DECODE
			return b.convertDecodeToCase(args...)
		},
		Default: func() schema.QueryAppender {
			// For Oracle or other databases that might support DECODE
			return b.Expr("DECODE(?)", newExpressions(constants.CommaSpace, args...))
		},
	})
}

// convertDecodeToCase converts DECODE syntax to CASE WHEN expression using the existing Case builder.
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
