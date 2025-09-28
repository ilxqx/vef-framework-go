package orm

import (
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/sort"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// BaseAggregate defines the basic aggregate function interface with generic type support.
type BaseAggregate[T any] interface {
	// Column sets the aggregate argument using a column reference.
	Column(column string) T
	// Expr sets the aggregate argument using a raw expression.
	Expr(expr any) T
	// Filter applies a FILTER clause to the aggregate expression.
	Filter(func(ConditionBuilder)) T
}

// DistinctableAggregate defines aggregate functions that support DISTINCT operations.
type DistinctableAggregate[T any] interface {
	// Distinct marks the aggregate to operate on DISTINCT values.
	Distinct() T
}

// OrderableAggregate defines aggregate functions that support ordering.
type OrderableAggregate[T any] interface {
	// OrderBy adds ORDER BY clauses with ascending direction inside the aggregate.
	OrderBy(columns ...string) T
	// OrderByDesc adds ORDER BY clauses with descending direction inside the aggregate.
	OrderByDesc(columns ...string) T
	// OrderByExpr adds an ORDER BY clause based on a raw expression inside the aggregate.
	OrderByExpr(expr any) T
}

// NullHandlingBuilder defines aggregate functions that support NULL value handling.
type NullHandlingBuilder[T any] interface {
	// IgnoreNulls configures the aggregate to ignore NULL values.
	IgnoreNulls() T
	// RespectNulls configures the aggregate to respect NULL values.
	RespectNulls() T
}

// StatisticalAggregate defines aggregate functions that support statistical modes.
type StatisticalAggregate[T any] interface {
	// Population configures the aggregate to use population statistics (e.g., STDDEV_POP).
	Population() T
	// Sample configures the aggregate to use sample statistics (e.g., STDDEV_SAMP).
	Sample() T
}

// CountBuilder defines the COUNT aggregate function builder.
type CountBuilder interface {
	BaseAggregate[CountBuilder]
	DistinctableAggregate[CountBuilder]
	// All configures COUNT(*) semantics.
	All() CountBuilder
}

// SumBuilder defines the SUM aggregate function builder.
type SumBuilder interface {
	BaseAggregate[SumBuilder]
	DistinctableAggregate[SumBuilder]
}

// AvgBuilder defines the AVG aggregate function builder.
type AvgBuilder interface {
	BaseAggregate[AvgBuilder]
	DistinctableAggregate[AvgBuilder]
}

// MinBuilder defines the MIN aggregate function builder.
type MinBuilder interface {
	BaseAggregate[MinBuilder]
}

// MaxBuilder defines the MAX aggregate function builder.
type MaxBuilder interface {
	BaseAggregate[MaxBuilder]
}

// StringAggBuilder defines the STRING_AGG aggregate function builder.
type StringAggBuilder interface {
	BaseAggregate[StringAggBuilder]
	DistinctableAggregate[StringAggBuilder]
	OrderableAggregate[StringAggBuilder]
	NullHandlingBuilder[StringAggBuilder]

	// Separator sets the delimiter used between aggregated values.
	Separator(separator string) StringAggBuilder
}

// ArrayAggBuilder defines the ARRAY_AGG aggregate function builder.
type ArrayAggBuilder interface {
	BaseAggregate[ArrayAggBuilder]
	DistinctableAggregate[ArrayAggBuilder]
	OrderableAggregate[ArrayAggBuilder]
	NullHandlingBuilder[ArrayAggBuilder]
}

// StdDevBuilder defines the STDDEV aggregate function builder.
type StdDevBuilder interface {
	BaseAggregate[StdDevBuilder]
	StatisticalAggregate[StdDevBuilder]
}

// VarianceBuilder defines the VARIANCE aggregate function builder.
type VarianceBuilder interface {
	BaseAggregate[VarianceBuilder]
	StatisticalAggregate[VarianceBuilder]
}

// JSONObjectAggBuilder defines the JSON_OBJECT_AGG aggregate function builder.
type JSONObjectAggBuilder interface {
	BaseAggregate[JSONObjectAggBuilder]
	DistinctableAggregate[JSONObjectAggBuilder]
	OrderableAggregate[JSONObjectAggBuilder]

	// KeyColumn sets the key using a column reference.
	KeyColumn(column string) JSONObjectAggBuilder
	// KeyExpr sets the key using a raw expression.
	KeyExpr(expr any) JSONObjectAggBuilder
}

// JSONArrayAggBuilder defines the JSON_ARRAY_AGG aggregate function builder.
type JSONArrayAggBuilder interface {
	BaseAggregate[JSONArrayAggBuilder]
	DistinctableAggregate[JSONArrayAggBuilder]
	OrderableAggregate[JSONArrayAggBuilder]
}

// BitOrBuilder defines the BIT_OR aggregate function builder.
type BitOrBuilder interface {
	BaseAggregate[BitOrBuilder]
}

// BitAndBuilder defines the BIT_AND aggregate function builder.
type BitAndBuilder interface {
	BaseAggregate[BitAndBuilder]
}

// BoolOrBuilder defines the BOOL_OR aggregate function builder.
type BoolOrBuilder interface {
	BaseAggregate[BoolOrBuilder]
}

// BoolAndBuilder defines the BOOL_AND aggregate function builder.
type BoolAndBuilder interface {
	BaseAggregate[BoolAndBuilder]
}

// ========== Aggregate Function Implementations ==========

// baseAggregateExpr implements common functionality for all aggregate expressions
type baseAggregateExpr struct {
	// qb holds the current query builder
	qb QueryBuilder
	// eb provides access to expression building utilities (columns, expressions, etc.)
	eb ExprBuilder
	// funcName stores the SQL function name (e.g., "SUM", "COUNT", "JSON_AGG")
	funcName string
	// argsExpr holds the expression(s) passed as arguments to the aggregate function
	argsExpr schema.QueryAppender
	// distinct indicates whether to use DISTINCT clause in the aggregate (e.g., COUNT(DISTINCT id))
	distinct bool
	// filter stores the FILTER clause condition for conditional aggregation
	filter schema.QueryAppender
	// orderExprs contains ORDER BY clauses for ordered aggregates like STRING_AGG
	orderExprs []orderExpr
	// nullsMode controls how NULL values are handled in window functions (IGNORE/RESPECT NULLS)
	nullsMode NullsMode
	// separator stores the delimiter for string aggregation functions like STRING_AGG
	separator string
	// statisticalMode indicates population vs sample mode for statistical functions (STDDEV, VARIANCE)
	statisticalMode StatisticalMode
}

func (a *baseAggregateExpr) setFilter(builder func(ConditionBuilder)) {
	a.filter = a.qb.BuildCondition(builder)
}

func (a *baseAggregateExpr) appendOrderBy(columns ...string) {
	for _, column := range columns {
		a.orderExprs = append(a.orderExprs, orderExpr{
			builders:   a.eb,
			column:     column,
			direction:  sort.OrderAsc,
			nullsOrder: sort.NullsDefault,
		})
	}
}

func (a *baseAggregateExpr) appendOrderByDesc(columns ...string) {
	for _, column := range columns {
		a.orderExprs = append(a.orderExprs, orderExpr{
			builders:   a.eb,
			column:     column,
			direction:  sort.OrderDesc,
			nullsOrder: sort.NullsDefault,
		})
	}
}

func (a *baseAggregateExpr) appendOrderByExpr(expr any) {
	a.orderExprs = append(a.orderExprs, orderExpr{
		builders:   a.eb,
		expr:       expr,
		direction:  sort.OrderAsc,
		nullsOrder: sort.NullsDefault,
	})
}

func (a *baseAggregateExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	if a.argsExpr == nil {
		return nil, ErrAggregateMissingArgs
	}

	// Handle FILTER clause for databases without native FILTER support
	if a.filter != nil {
		var handled bool
		a.eb.RunDialect(DialectActions{
			MySQL: func() {
				b, err = a.appendCompatibleFilterQuery(fmter, b)
				handled = true
			},
			Oracle: func() {
				b, err = a.appendCompatibleFilterQuery(fmter, b)
				handled = true
			},
			SQLServer: func() {
				b, err = a.appendCompatibleFilterQuery(fmter, b)
				handled = true
			},
		})

		if handled {
			return b, err
		}
	}

	// Function name
	b = append(b, a.funcName...)
	b = append(b, constants.ByteLeftParenthesis)

	// DISTINCT keyword
	if a.distinct {
		b = append(b, "DISTINCT "...)
	}

	// Args expression
	if b, err = a.argsExpr.AppendQuery(fmter, b); err != nil {
		return
	}

	// Order by expression
	if len(a.orderExprs) > 0 {
		b = append(b, constants.ByteSpace)
		if b, err = newOrderByClause(a.orderExprs...).AppendQuery(fmter, b); err != nil {
			return
		}
	}

	b = append(b, constants.ByteRightParenthesis)

	// NULLS mode (Oracle and SQL Server only)
	if a.nullsMode != NullsDefault {
		b = append(b, constants.ByteSpace)
		b = append(b, a.nullsMode.String()...)
	}

	// FILTER clause (PostgreSQL, SQLite have native support; MySQL, Oracle, SQL Server use CASE conversion)
	if a.filter != nil {
		if b, err = newFilterClause(a.filter).AppendQuery(fmter, b); err != nil {
			return
		}
	}

	return b, nil
}

// appendCompatibleFilterQuery handles FILTER clause for MySQL, Oracle, SQL Server by converting to CASE WHEN syntax
func (a *baseAggregateExpr) appendCompatibleFilterQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	// Store original function name for comparison
	funcName := a.funcName

	// Convert function name for database compatibility (MySQL, Oracle, SQL Server)
	switch funcName {
	case "COUNT":
		// COUNT(*) FILTER (WHERE condition) → SUM(CASE WHEN condition THEN 1 ELSE 0 END)
		b = append(b, "SUM"...)
	case "AVG", "MIN", "MAX":
		// Keep the same function name but wrap column/expression in CASE WHEN
		b = append(b, funcName...)
	default:
		// For SUM, STRING_AGG, etc., keep the same function name
		b = append(b, funcName...)
	}

	b = append(b, constants.ByteLeftParenthesis)

	// DISTINCT keyword
	if a.distinct {
		b = append(b, "DISTINCT "...)
	}

	// Generate CASE WHEN expression
	if b, err = a.eb.Case(func(cb CaseBuilder) {
		when := cb.WhenExpr(a.filter)
		switch funcName {
		case "COUNT":
			when.Then(1)
		default:
			when.Then(a.argsExpr)
		}

		switch funcName {
		case "COUNT", "SUM":
			cb.Else(0)
		default:
			cb.Else(a.eb.Null())
		}
	}).AppendQuery(fmter, b); err != nil {
		return
	}

	b = append(b, constants.ByteRightParenthesis)

	return b, nil
}

// baseAggregateBuilder is a base struct for all aggregate function builders
type baseAggregateBuilder[T any] struct {
	*baseAggregateExpr
	self T
}

func (b *baseAggregateBuilder[T]) Column(column string) T {
	b.argsExpr = b.eb.Column(column)
	return b.self
}

func (b *baseAggregateBuilder[T]) Expr(expr any) T {
	b.argsExpr = b.eb.Expr("?", expr)
	return b.self
}

func (b *baseAggregateBuilder[T]) Filter(builder func(ConditionBuilder)) T {
	b.setFilter(builder)
	return b.self
}

// distinctableAggregateBuilder provides DISTINCT functionality
type distinctableAggregateBuilder[T any] struct {
	*baseAggregateBuilder[T]
}

func (b *distinctableAggregateBuilder[T]) Distinct() T {
	b.distinct = true
	return b.self
}

// orderableAggregateBuilder provides ORDER BY functionality
type orderableAggregateBuilder[T any] struct {
	*baseAggregateBuilder[T]
}

func (b *orderableAggregateBuilder[T]) OrderBy(columns ...string) T {
	b.appendOrderBy(columns...)
	return b.self
}

func (b *orderableAggregateBuilder[T]) OrderByDesc(columns ...string) T {
	b.appendOrderByDesc(columns...)
	return b.self
}

func (b *orderableAggregateBuilder[T]) OrderByExpr(expr any) T {
	b.appendOrderByExpr(expr)
	return b.self
}

// baseNullHandlingBuilder provides NULL handling functionality
type baseNullHandlingBuilder[T any] struct {
	*baseAggregateBuilder[T]
}

func (b *baseNullHandlingBuilder[T]) IgnoreNulls() T {
	b.nullsMode = NullsIgnore
	return b.self
}

func (b *baseNullHandlingBuilder[T]) RespectNulls() T {
	b.nullsMode = NullsRespect
	return b.self
}

// statisticalAggregateBuilder provides statistical mode functionality
type statisticalAggregateBuilder[T any] struct {
	*baseAggregateBuilder[T]
}

func (b *statisticalAggregateBuilder[T]) Population() T {
	b.statisticalMode = StatisticalPopulation
	return b.self
}

func (b *statisticalAggregateBuilder[T]) Sample() T {
	b.statisticalMode = StatisticalSample
	return b.self
}

// countExpr implements CountBuilder
type countExpr[T any] struct {
	*baseAggregateExpr
	*distinctableAggregateBuilder[T]
}

func (c *countExpr[T]) All() T {
	c.argsExpr = bun.Safe(columnAll)
	return c.self
}

// sumExpr implements SumBuilder
type sumExpr[T any] struct {
	*baseAggregateExpr
	*distinctableAggregateBuilder[T]
}

// avgExpr implements AvgBuilder
type avgExpr[T any] struct {
	*baseAggregateExpr
	*distinctableAggregateBuilder[T]
}

// minExpr implements MinBuilder
type minExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
}

// maxExpr implements MaxBuilder
type maxExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
}

// stringAggExpr implements StringAggBuilder
type stringAggExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
	*distinctableAggregateBuilder[T]
	*orderableAggregateBuilder[T]
	*baseNullHandlingBuilder[T]
}

func (s *stringAggExpr[T]) Separator(separator string) T {
	s.separator = separator
	return s.self
}

func (s *stringAggExpr[T]) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	var (
		originalFuncName   = s.funcName
		originalArgsExpr   = s.argsExpr
		originalDistinct   = s.distinct
		originalNullsMode  = s.nullsMode
		originalOrderExprs = s.orderExprs
	)

	defer func() {
		s.funcName = originalFuncName
		s.argsExpr = originalArgsExpr
		s.distinct = originalDistinct
		s.nullsMode = originalNullsMode
		s.orderExprs = originalOrderExprs
	}()

	if err = s.eb.RunDialectErr(DialectActionsErr{
		Postgres: func() error {
			// PostgreSQL: STRING_AGG([DISTINCT] expression, delimiter [ORDER BY ...])
			s.funcName = "STRING_AGG"

			argsExpr := s.argsExpr
			if s.nullsMode == NullsIgnore {
				// Column or expression with IGNORE NULLS emulation
				argsExpr = s.eb.Case(func(cb CaseBuilder) {
					cb.WhenExpr(s.eb.IsNotNull(s.argsExpr)).Then(s.argsExpr)
				})
			}

			s.argsExpr = s.eb.Expr(
				"?, ?",
				argsExpr,
				s.separator,
			)
			return nil
		},
		MySQL: func() error {
			// MySQL: GROUP_CONCAT([DISTINCT] expression [ORDER BY ...] [SEPARATOR delimiter])
			s.funcName = "GROUP_CONCAT"
			argsExpr := s.argsExpr
			if s.nullsMode == NullsIgnore {
				// Column or expression with IGNORE NULLS emulation
				argsExpr = s.eb.Case(func(cb CaseBuilder) {
					cb.WhenExpr(s.eb.IsNotNull(s.argsExpr)).Then(s.argsExpr)
				})
			}

			// For MySQL, we need to handle ORDER BY differently
			// GROUP_CONCAT(expression ORDER BY column SEPARATOR separator)
			if len(s.orderExprs) > 0 {
				// Build ORDER BY clause for MySQL GROUP_CONCAT
				s.argsExpr = s.eb.Expr(
					"? ? SEPARATOR ?",
					argsExpr,
					newOrderByClause(s.orderExprs...),
					s.separator,
				)
			} else {
				s.argsExpr = s.eb.Expr(
					"? SEPARATOR ?",
					argsExpr,
					s.separator,
				)
			}

			s.nullsMode = NullsDefault
			s.orderExprs = nil // Clear order expressions as they're now handled in argsExpr
			return nil
		},
		SQLite: func() error {
			// SQLite: GROUP_CONCAT([DISTINCT] expression, delimiter)
			// SQLite supports DISTINCT and basic ORDER BY in GROUP_CONCAT since version 3.44.0 (2023)
			s.funcName = "GROUP_CONCAT"

			argsExpr := s.argsExpr
			if s.nullsMode == NullsIgnore {
				// Column or expression with IGNORE NULLS emulation
				argsExpr = s.eb.Case(func(cb CaseBuilder) {
					cb.WhenExpr(s.eb.IsNotNull(s.argsExpr)).Then(s.argsExpr)
				})
			}

			// SQLite limitation: DISTINCT aggregates must have exactly one argument
			if s.distinct {
				// When using DISTINCT, we can only use the default separator (comma)
				s.argsExpr = argsExpr
			} else {
				// When not using DISTINCT, we can specify a custom separator
				s.argsExpr = s.eb.Expr(
					"?, ?",
					argsExpr,
					s.separator,
				)
			}

			s.nullsMode = NullsDefault
			// Keep DISTINCT and ORDER BY support for modern SQLite
			return nil
		},
		Default: func() error {
			return ErrDialectUnsupportedOperation
		},
	}); err != nil {
		return nil, err
	}

	return s.baseAggregateExpr.AppendQuery(fmter, b)
}

// arrayAggExpr implements ArrayAggBuilder
type arrayAggExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
	*distinctableAggregateBuilder[T]
	*orderableAggregateBuilder[T]
	*baseNullHandlingBuilder[T]
}

func (a *arrayAggExpr[T]) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	var (
		originalFuncName   = a.funcName
		originalArgsExpr   = a.argsExpr
		originalDistinct   = a.distinct
		originalNullsMode  = a.nullsMode
		originalOrderExprs = a.orderExprs
	)

	defer func() {
		a.funcName = originalFuncName
		a.argsExpr = originalArgsExpr
		a.distinct = originalDistinct
		a.nullsMode = originalNullsMode
		a.orderExprs = originalOrderExprs
	}()

	if err = a.eb.RunDialectErr(DialectActionsErr{
		Postgres: func() error {
			// PostgreSQL: ARRAY_AGG([DISTINCT] expression [ORDER BY ...])
			a.funcName = "ARRAY_AGG"
			return nil
		},
		MySQL: func() error {
			// MySQL: JSON_ARRAYAGG(expression) - MySQL 5.7.22+
			// Note: MySQL JSON_ARRAYAGG doesn't support DISTINCT or ORDER BY directly
			a.funcName = "JSON_ARRAYAGG"
			argsExpr := a.argsExpr
			if a.nullsMode == NullsIgnore {
				// Column or expression with IGNORE NULLS emulation
				argsExpr = a.eb.Case(func(cb CaseBuilder) {
					cb.WhenExpr(a.eb.IsNotNull(a.argsExpr)).Then(a.argsExpr)
				})
			}

			a.argsExpr = argsExpr
			a.nullsMode = NullsDefault
			// JSON_ARRAYAGG doesn't support DISTINCT or ORDER BY - this is correct to disable
			a.distinct = false
			a.orderExprs = nil
			return nil
		},
		SQLite: func() error {
			// SQLite: JSON_GROUP_ARRAY(expression)
			// SQLite JSON_GROUP_ARRAY supports basic features but not DISTINCT or ORDER BY
			a.funcName = "JSON_GROUP_ARRAY"
			argsExpr := a.argsExpr
			if a.nullsMode == NullsIgnore {
				// Column or expression with IGNORE NULLS emulation
				argsExpr = a.eb.Case(func(cb CaseBuilder) {
					cb.WhenExpr(a.eb.IsNotNull(a.argsExpr)).Then(a.argsExpr)
				})
			}

			a.argsExpr = argsExpr
			a.nullsMode = NullsDefault
			// JSON_GROUP_ARRAY doesn't support DISTINCT or ORDER BY - this is correct to disable
			a.distinct = false
			a.orderExprs = nil
			return nil
		},
		Default: func() error {
			return ErrDialectUnsupportedOperation
		},
	}); err != nil {
		return nil, err
	}

	return a.baseAggregateExpr.AppendQuery(fmter, b)
}

// statisticalAggExpr implements statistical aggregate functions (STDDEV, VARIANCE)
type statisticalAggExpr struct {
	*baseAggregateExpr
}

func (s *statisticalAggExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	var originalFuncName = s.funcName
	defer func() {
		s.funcName = originalFuncName
	}()

	if err = s.eb.RunDialectErr(DialectActionsErr{
		Postgres: func() error {
			s.funcName = s.funcName + constants.Underscore + lo.CoalesceOrEmpty(s.statisticalMode.String(), StatisticalPopulation.String())
			return nil
		},
		MySQL: func() error {
			switch s.statisticalMode {
			case StatisticalPopulation, StatisticalSample:
				s.funcName = s.funcName + constants.Underscore + s.statisticalMode.String()
			}
			return nil
		},
		SQLite: func() error {
			return ErrAggregateUnsupportedFunction
		},
		Default: func() error {
			return ErrDialectUnsupportedOperation
		},
	}); err != nil {
		return nil, err
	}

	return s.baseAggregateExpr.AppendQuery(fmter, b)
}

// stddevExpr implements StdDevBuilder
type stddevExpr[T any] struct {
	*statisticalAggExpr
	*statisticalAggregateBuilder[T]
}

// varianceExpr implements VarianceBuilder
type varianceExpr[T any] struct {
	*statisticalAggExpr
	*statisticalAggregateBuilder[T]
}

// jsonObjectAggExpr implements JsonObjectAggBuilder
type jsonObjectAggExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
	*distinctableAggregateBuilder[T]
	*orderableAggregateBuilder[T]

	keyExpr schema.QueryAppender
}

func (j *jsonObjectAggExpr[T]) KeyColumn(column string) T {
	j.keyExpr = j.eb.Column(column)
	return j.self
}

func (j *jsonObjectAggExpr[T]) KeyExpr(expr any) T {
	j.keyExpr = j.eb.Expr("?", expr)
	return j.self
}

func (j *jsonObjectAggExpr[T]) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	if j.keyExpr == nil {
		return nil, ErrAggregateMissingArgs
	}

	// Create combined key-value expression for the aggregate
	var (
		originalFuncName = j.funcName
		originalArgsExpr = j.argsExpr
	)
	defer func() {
		j.funcName = originalFuncName
		j.argsExpr = originalArgsExpr
	}()

	if err = j.eb.RunDialectErr(DialectActionsErr{
		Postgres: func() error {
			// PostgreSQL uses json_object_agg(key, value)
			j.funcName = "JSON_OBJECT_AGG"
			j.argsExpr = j.eb.Exprs(j.keyExpr, originalArgsExpr)
			return nil
		},
		MySQL: func() error {
			// MySQL uses JSON_OBJECTAGG(key, value)
			j.funcName = "JSON_OBJECTAGG"
			j.argsExpr = j.eb.Exprs(j.keyExpr, originalArgsExpr)
			return nil
		},
		SQLite: func() error {
			// SQLite uses json_group_object(key, value)
			j.funcName = "JSON_GROUP_OBJECT"
			j.argsExpr = j.eb.Exprs(j.keyExpr, originalArgsExpr)
			return nil
		},
		Default: func() error {
			return ErrDialectUnsupportedOperation
		},
	}); err != nil {
		return
	}

	return j.baseAggregateExpr.AppendQuery(fmter, b)
}

// jsonArrayAggExpr implements JsonArrayAggBuilder
type jsonArrayAggExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
	*distinctableAggregateBuilder[T]
	*orderableAggregateBuilder[T]
}

func (j *jsonArrayAggExpr[T]) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	var originalFuncName = j.funcName
	defer func() {
		j.funcName = originalFuncName
	}()

	if err = j.eb.RunDialectErr(DialectActionsErr{
		Postgres: func() error {
			// PostgreSQL uses json_agg(expression)
			j.funcName = "JSON_AGG"
			return nil
		},
		MySQL: func() error {
			// MySQL uses JSON_ARRAYAGG(expression)
			j.funcName = "JSON_ARRAYAGG"
			return nil
		},
		SQLite: func() error {
			// SQLite uses json_group_array(expression)
			j.funcName = "JSON_GROUP_ARRAY"
			return nil
		},
		Default: func() error {
			return ErrDialectUnsupportedOperation
		},
	}); err != nil {
		return
	}

	return j.baseAggregateExpr.AppendQuery(fmter, b)
}

// bitOrExpr implements BitOrBuilder
type bitOrExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
}

func (b *bitOrExpr[T]) AppendQuery(fmter schema.Formatter, buf []byte) (_ []byte, err error) {
	var (
		originalFuncName = b.funcName
		originalArgsExpr = b.argsExpr
	)
	defer func() {
		b.funcName = originalFuncName
		b.argsExpr = originalArgsExpr
	}()

	if err = b.eb.RunDialectErr(DialectActionsErr{
		MySQL: func() error {
			// MySQL supports BIT_OR
			b.funcName = "BIT_OR"
			return nil
		},
		Postgres: func() error {
			// PostgreSQL doesn't have BIT_OR, but we can simulate it using custom aggregation
			// For simple boolean-like bit operations, we can use BOOL_OR
			// For integer bit operations, we need a more complex approach
			b.funcName = "BOOL_OR"
			// Convert numeric values to boolean for BOOL_OR
			b.argsExpr = b.eb.Expr("? != 0", originalArgsExpr)
			return nil
		},
		SQLite: func() error {
			// SQLite doesn't have BIT_OR, simulate using MAX with CASE
			// This works for simple cases where we're doing boolean-like bit operations
			b.funcName = "MAX"
			// Convert to 1/0 and use MAX to simulate OR behavior
			b.argsExpr = b.eb.Case(func(cb CaseBuilder) {
				cb.WhenExpr(b.eb.Expr("? != 0", originalArgsExpr)).Then(1).Else(0)
			})
			return nil
		},
		Default: func() error {
			return ErrDialectUnsupportedOperation
		},
	}); err != nil {
		return
	}

	return b.baseAggregateExpr.AppendQuery(fmter, buf)
}

// bitAndExpr implements BitAndBuilder
type bitAndExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
}

func (b *bitAndExpr[T]) AppendQuery(fmter schema.Formatter, buf []byte) (_ []byte, err error) {
	var (
		originalFuncName = b.funcName
		originalArgsExpr = b.argsExpr
	)
	defer func() {
		b.funcName = originalFuncName
		b.argsExpr = originalArgsExpr
	}()

	if err = b.eb.RunDialectErr(DialectActionsErr{
		MySQL: func() error {
			// MySQL supports BIT_AND
			b.funcName = "BIT_AND"
			return nil
		},
		Postgres: func() error {
			// PostgreSQL doesn't have BIT_AND, but we can simulate it using BOOL_AND
			// For simple boolean-like bit operations
			b.funcName = "BOOL_AND"
			// Convert numeric values to boolean for BOOL_AND
			b.argsExpr = b.eb.Expr("? != 0", originalArgsExpr)
			return nil
		},
		SQLite: func() error {
			// SQLite doesn't have BIT_AND, simulate using MIN with CASE
			// This works for simple cases where we're doing boolean-like bit operations
			b.funcName = "MIN"
			// Convert to 1/0 and use MIN to simulate AND behavior
			b.argsExpr = b.eb.Case(func(cb CaseBuilder) {
				cb.WhenExpr(b.eb.Expr("? != 0", originalArgsExpr)).Then(1).Else(0)
			})
			return nil
		},
		Default: func() error {
			return ErrDialectUnsupportedOperation
		},
	}); err != nil {
		return
	}

	return b.baseAggregateExpr.AppendQuery(fmter, buf)
}

// boolOrExpr implements BoolOrBuilder
type boolOrExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
}

func (b *boolOrExpr[T]) AppendQuery(fmter schema.Formatter, buf []byte) (_ []byte, err error) {
	var (
		originalFuncName = b.funcName
		originalArgsExpr = b.argsExpr
	)
	defer func() {
		b.funcName = originalFuncName
		b.argsExpr = originalArgsExpr
	}()

	if err = b.eb.RunDialectErr(DialectActionsErr{
		Postgres: func() error {
			// PostgreSQL supports BOOL_OR
			b.funcName = "BOOL_OR"
			return nil
		},
		MySQL: func() error {
			// MySQL doesn't have BOOL_OR, simulate with MAX and CASE
			b.funcName = "MAX"
			// Convert to 0/1 integers and use MAX to simulate OR behavior
			b.argsExpr = b.eb.Case(func(cb CaseBuilder) {
				cb.WhenExpr(b.argsExpr).Then(1).Else(0)
			})
			return nil
		},
		SQLite: func() error {
			// SQLite doesn't have BOOL_OR, simulate with MAX and CASE
			b.funcName = "MAX"
			// Convert boolean values to 0/1 and use MAX to simulate OR behavior
			b.argsExpr = b.eb.Case(func(cb CaseBuilder) {
				cb.WhenExpr(b.argsExpr).Then(1).Else(0)
			})
			return nil
		},
		Default: func() error {
			return ErrDialectUnsupportedOperation
		},
	}); err != nil {
		return
	}

	return b.baseAggregateExpr.AppendQuery(fmter, buf)
}

// boolAndExpr implements BoolAndBuilder
type boolAndExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
}

func (b *boolAndExpr[T]) AppendQuery(fmter schema.Formatter, buf []byte) (_ []byte, err error) {
	var (
		originalFuncName = b.funcName
		originalArgsExpr = b.argsExpr
	)
	defer func() {
		b.funcName = originalFuncName
		b.argsExpr = originalArgsExpr
	}()

	if err = b.eb.RunDialectErr(DialectActionsErr{
		Postgres: func() error {
			// PostgreSQL supports BOOL_AND
			b.funcName = "BOOL_AND"
			return nil
		},
		MySQL: func() error {
			// MySQL doesn't have BOOL_AND, simulate with MIN and CASE
			b.funcName = "MIN"
			// Convert to 0/1 integers and use MIN to simulate AND behavior
			b.argsExpr = b.eb.Case(func(cb CaseBuilder) {
				cb.WhenExpr(b.argsExpr).Then(1).Else(0)
			})
			return nil
		},
		SQLite: func() error {
			// SQLite doesn't have BOOL_AND, simulate with MIN and CASE
			b.funcName = "MIN"
			// Convert boolean values to 0/1 and use MIN to simulate AND behavior
			b.argsExpr = b.eb.Case(func(cb CaseBuilder) {
				cb.WhenExpr(b.argsExpr).Then(1).Else(0)
			})
			return nil
		},
		Default: func() error {
			return ErrDialectUnsupportedOperation
		},
	}); err != nil {
		return
	}

	return b.baseAggregateExpr.AppendQuery(fmter, buf)
}

// ========== Factory Functions ==========

// newGenericCountExpr creates a new COUNT expression
func newGenericCountExpr[T any](self T, qb QueryBuilder) *countExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "COUNT",
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &countExpr[T]{
		baseAggregateExpr: baseExpr,
		distinctableAggregateBuilder: &distinctableAggregateBuilder[T]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = self
	return expr
}

// newCountExpr creates a new COUNT expression
func newCountExpr(qb QueryBuilder) *countExpr[CountBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "COUNT",
	}
	baseBuilder := &baseAggregateBuilder[CountBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &countExpr[CountBuilder]{
		baseAggregateExpr: baseExpr,
		distinctableAggregateBuilder: &distinctableAggregateBuilder[CountBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = expr
	return expr
}

// newGenericSumExpr creates a new SUM expression
func newGenericSumExpr[T any](self T, qb QueryBuilder) *sumExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "SUM",
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &sumExpr[T]{
		baseAggregateExpr: baseExpr,
		distinctableAggregateBuilder: &distinctableAggregateBuilder[T]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = self
	return expr
}

// newSumExpr creates a new SUM expression
func newSumExpr(qb QueryBuilder) *sumExpr[SumBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "SUM",
	}
	baseBuilder := &baseAggregateBuilder[SumBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &sumExpr[SumBuilder]{
		baseAggregateExpr: baseExpr,
		distinctableAggregateBuilder: &distinctableAggregateBuilder[SumBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = expr
	return expr
}

// newGenericAvgExpr creates a new AVG expression
func newGenericAvgExpr[T any](self T, qb QueryBuilder) *avgExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "AVG",
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &avgExpr[T]{
		baseAggregateExpr: baseExpr,
		distinctableAggregateBuilder: &distinctableAggregateBuilder[T]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = self
	return expr
}

// newAvgExpr creates a new AVG expression
func newAvgExpr(qb QueryBuilder) *avgExpr[AvgBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "AVG",
	}
	baseBuilder := &baseAggregateBuilder[AvgBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &avgExpr[AvgBuilder]{
		baseAggregateExpr: baseExpr,
		distinctableAggregateBuilder: &distinctableAggregateBuilder[AvgBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = expr
	return expr
}

// newGenericMinExpr creates a new MIN expression
func newGenericMinExpr[T any](self T, qb QueryBuilder) *minExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "MIN",
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &minExpr[T]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
	}

	baseBuilder.self = self
	return expr
}

// newMinExpr creates a new MIN expression
func newMinExpr(qb QueryBuilder) *minExpr[MinBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "MIN",
	}
	baseBuilder := &baseAggregateBuilder[MinBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &minExpr[MinBuilder]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
	}

	baseBuilder.self = expr
	return expr
}

// newGenericMaxExpr creates a new MAX expression
func newGenericMaxExpr[T any](self T, qb QueryBuilder) *maxExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "MAX",
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &maxExpr[T]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
	}

	baseBuilder.self = self
	return expr
}

// newMaxExpr creates a new MAX expression
func newMaxExpr(qb QueryBuilder) *maxExpr[MaxBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "MAX",
	}
	baseBuilder := &baseAggregateBuilder[MaxBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &maxExpr[MaxBuilder]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
	}

	baseBuilder.self = expr
	return expr
}

// newGenericStringAggExpr creates a new STRING_AGG expression
func newGenericStringAggExpr[T any](self T, qb QueryBuilder) *stringAggExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:        qb,
		eb:        qb.ExprBuilder(),
		funcName:  constants.Empty,
		separator: constants.Comma,
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &stringAggExpr[T]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
	}

	baseBuilder.self = self
	return expr
}

// newStringAggExpr creates a new STRING_AGG expression
func newStringAggExpr(qb QueryBuilder) *stringAggExpr[StringAggBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:        qb,
		eb:        qb.ExprBuilder(),
		funcName:  constants.Empty,
		separator: constants.Comma,
	}
	baseBuilder := &baseAggregateBuilder[StringAggBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &stringAggExpr[StringAggBuilder]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
		distinctableAggregateBuilder: &distinctableAggregateBuilder[StringAggBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
		orderableAggregateBuilder: &orderableAggregateBuilder[StringAggBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
		baseNullHandlingBuilder: &baseNullHandlingBuilder[StringAggBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = expr
	return expr
}

// newGenericArrayAggExpr creates a new ARRAY_AGG expression
func newGenericArrayAggExpr[T any](self T, qb QueryBuilder) *arrayAggExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: constants.Empty,
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &arrayAggExpr[T]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
		distinctableAggregateBuilder: &distinctableAggregateBuilder[T]{
			baseAggregateBuilder: baseBuilder,
		},
		orderableAggregateBuilder: &orderableAggregateBuilder[T]{
			baseAggregateBuilder: baseBuilder,
		},
		baseNullHandlingBuilder: &baseNullHandlingBuilder[T]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = self
	return expr
}

// newArrayAggExpr creates a new ARRAY_AGG expression
func newArrayAggExpr(qb QueryBuilder) *arrayAggExpr[ArrayAggBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: constants.Empty,
	}
	baseBuilder := &baseAggregateBuilder[ArrayAggBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &arrayAggExpr[ArrayAggBuilder]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
		distinctableAggregateBuilder: &distinctableAggregateBuilder[ArrayAggBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
		orderableAggregateBuilder: &orderableAggregateBuilder[ArrayAggBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
		baseNullHandlingBuilder: &baseNullHandlingBuilder[ArrayAggBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = expr
	return expr
}

// newGenericStdDevExpr creates a new STDDEV expression
func newGenericStdDevExpr[T any](self T, qb QueryBuilder) *stddevExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "STDDEV",
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &stddevExpr[T]{
		statisticalAggExpr: &statisticalAggExpr{
			baseAggregateExpr: baseExpr,
		},
		statisticalAggregateBuilder: &statisticalAggregateBuilder[T]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = self
	return expr
}

// newStdDevExpr creates a new STDDEV expression
func newStdDevExpr(qb QueryBuilder) *stddevExpr[StdDevBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "STDDEV",
	}
	baseBuilder := &baseAggregateBuilder[StdDevBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &stddevExpr[StdDevBuilder]{
		statisticalAggExpr: &statisticalAggExpr{
			baseAggregateExpr: baseExpr,
		},
		statisticalAggregateBuilder: &statisticalAggregateBuilder[StdDevBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = expr
	return expr
}

// newGenericVarianceExpr creates a new VARIANCE expression
func newGenericVarianceExpr[T any](self T, qb QueryBuilder) *varianceExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "VARIANCE",
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &varianceExpr[T]{
		statisticalAggExpr: &statisticalAggExpr{
			baseAggregateExpr: baseExpr,
		},
		statisticalAggregateBuilder: &statisticalAggregateBuilder[T]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = self
	return expr
}

// newVarianceExpr creates a new VARIANCE expression
func newVarianceExpr(qb QueryBuilder) *varianceExpr[VarianceBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "VARIANCE",
	}
	baseBuilder := &baseAggregateBuilder[VarianceBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &varianceExpr[VarianceBuilder]{
		statisticalAggExpr: &statisticalAggExpr{
			baseAggregateExpr: baseExpr,
		},
		statisticalAggregateBuilder: &statisticalAggregateBuilder[VarianceBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = expr
	return expr
}

// newGenericJsonObjectAggExpr creates a generic JSON_OBJECT_AGG expression for any builder type
func newGenericJsonObjectAggExpr[T any](self T, qb QueryBuilder) *jsonObjectAggExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: constants.Empty, // Will be set in AppendQuery based on dialect
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &jsonObjectAggExpr[T]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
		distinctableAggregateBuilder: &distinctableAggregateBuilder[T]{
			baseAggregateBuilder: baseBuilder,
		},
		orderableAggregateBuilder: &orderableAggregateBuilder[T]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = self
	return expr
}

// newJsonObjectAggExpr creates a new JSON_OBJECT_AGG expression
func newJsonObjectAggExpr(qb QueryBuilder) *jsonObjectAggExpr[JSONObjectAggBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: constants.Empty, // Will be set in AppendQuery based on dialect
	}
	baseBuilder := &baseAggregateBuilder[JSONObjectAggBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &jsonObjectAggExpr[JSONObjectAggBuilder]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
		distinctableAggregateBuilder: &distinctableAggregateBuilder[JSONObjectAggBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
		orderableAggregateBuilder: &orderableAggregateBuilder[JSONObjectAggBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = expr
	return expr
}

// newGenericJsonArrayAggExpr creates a generic JSON_ARRAY_AGG expression for any builder type
func newGenericJsonArrayAggExpr[T any](self T, qb QueryBuilder) *jsonArrayAggExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: constants.Empty, // Will be set in AppendQuery based on dialect
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &jsonArrayAggExpr[T]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
		distinctableAggregateBuilder: &distinctableAggregateBuilder[T]{
			baseAggregateBuilder: baseBuilder,
		},
		orderableAggregateBuilder: &orderableAggregateBuilder[T]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = self
	return expr
}

// newJsonArrayAggExpr creates a new JSON_ARRAY_AGG expression
func newJsonArrayAggExpr(qb QueryBuilder) *jsonArrayAggExpr[JSONArrayAggBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: constants.Empty, // Will be set in AppendQuery based on dialect
	}
	baseBuilder := &baseAggregateBuilder[JSONArrayAggBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &jsonArrayAggExpr[JSONArrayAggBuilder]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
		distinctableAggregateBuilder: &distinctableAggregateBuilder[JSONArrayAggBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
		orderableAggregateBuilder: &orderableAggregateBuilder[JSONArrayAggBuilder]{
			baseAggregateBuilder: baseBuilder,
		},
	}

	baseBuilder.self = expr
	return expr
}

// newGenericBitOrExpr creates a generic BIT_OR expression for any builder type
func newGenericBitOrExpr[T any](self T, qb QueryBuilder) *bitOrExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: constants.Empty, // Will be set in AppendQuery based on dialect
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &bitOrExpr[T]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
	}

	baseBuilder.self = self
	return expr
}

// newBitOrExpr creates a new BIT_OR expression
func newBitOrExpr(qb QueryBuilder) *bitOrExpr[BitOrBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "BIT_OR", // Default, will be adjusted in AppendQuery
	}
	baseBuilder := &baseAggregateBuilder[BitOrBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &bitOrExpr[BitOrBuilder]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
	}

	baseBuilder.self = expr
	return expr
}

// newGenericBitAndExpr creates a generic BIT_AND expression for any builder type
func newGenericBitAndExpr[T any](self T, qb QueryBuilder) *bitAndExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: constants.Empty, // Will be set in AppendQuery based on dialect
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &bitAndExpr[T]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
	}

	baseBuilder.self = self
	return expr
}

// newBitAndExpr creates a new BIT_AND expression
func newBitAndExpr(qb QueryBuilder) *bitAndExpr[BitAndBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "BIT_AND", // Default, will be adjusted in AppendQuery
	}
	baseBuilder := &baseAggregateBuilder[BitAndBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &bitAndExpr[BitAndBuilder]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
	}

	baseBuilder.self = expr
	return expr
}

// newGenericBoolOrExpr creates a generic BOOL_OR expression for any builder type
func newGenericBoolOrExpr[T any](self T, qb QueryBuilder) *boolOrExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: constants.Empty, // Will be set in AppendQuery based on dialect
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &boolOrExpr[T]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
	}

	baseBuilder.self = self
	return expr
}

// newBoolOrExpr creates a new BOOL_OR expression
func newBoolOrExpr(qb QueryBuilder) *boolOrExpr[BoolOrBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "BOOL_OR", // Default, will be adjusted in AppendQuery
	}
	baseBuilder := &baseAggregateBuilder[BoolOrBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &boolOrExpr[BoolOrBuilder]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
	}

	baseBuilder.self = expr
	return expr
}

// newGenericBoolAndExpr creates a generic BOOL_AND expression for any builder type
func newGenericBoolAndExpr[T any](self T, qb QueryBuilder) *boolAndExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: constants.Empty, // Will be set in AppendQuery based on dialect
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &boolAndExpr[T]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
	}

	baseBuilder.self = self
	return expr
}

// newBoolAndExpr creates a new BOOL_AND expression
func newBoolAndExpr(qb QueryBuilder) *boolAndExpr[BoolAndBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		funcName: "BOOL_AND", // Default, will be adjusted in AppendQuery
	}
	baseBuilder := &baseAggregateBuilder[BoolAndBuilder]{
		baseAggregateExpr: baseExpr,
	}
	expr := &boolAndExpr[BoolAndBuilder]{
		baseAggregateExpr:    baseExpr,
		baseAggregateBuilder: baseBuilder,
	}

	baseBuilder.self = expr
	return expr
}
