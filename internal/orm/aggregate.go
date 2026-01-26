package orm

import (
	"github.com/samber/lo"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/sort"
)

// BaseAggregate defines the basic aggregate function interface with generic type support.
type BaseAggregate[T any] interface {
	Column(column string) T
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

	KeyColumn(column string) JSONObjectAggBuilder
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

type dialectAggConfig struct {
	funcName        string
	argsTransformer func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender
	clearDistinct   bool
	clearOrderBy    bool
	clearNullsMode  bool
}

type dialectStrategy struct {
	postgres *dialectAggConfig
	mysql    *dialectAggConfig
	sqlite   *dialectAggConfig
	oracle   *dialectAggConfig
	sqlsrv   *dialectAggConfig
}

var bitOrStrategy = &dialectStrategy{
	postgres: &dialectAggConfig{funcName: "BIT_OR"},
	mysql:    &dialectAggConfig{funcName: "BIT_OR"},
	sqlite: &dialectAggConfig{
		funcName: "MAX",
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			return eb.Case(func(cb CaseBuilder) {
				cb.WhenExpr(eb.Expr("? != 0", state.argsExpr)).Then(1).Else(0)
			})
		},
	},
}

var bitAndStrategy = &dialectStrategy{
	postgres: &dialectAggConfig{funcName: "BIT_AND"},
	mysql:    &dialectAggConfig{funcName: "BIT_AND"},
	sqlite: &dialectAggConfig{
		funcName: "MIN",
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			return eb.Case(func(cb CaseBuilder) {
				cb.WhenExpr(eb.Expr("? != 0", state.argsExpr)).Then(1).Else(0)
			})
		},
	},
}

var boolOrStrategy = &dialectStrategy{
	postgres: &dialectAggConfig{funcName: "BOOL_OR"},
	mysql: &dialectAggConfig{
		funcName: "MAX",
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			return eb.Case(func(cb CaseBuilder) {
				cb.WhenExpr(state.argsExpr).Then(1).Else(0)
			})
		},
	},
	sqlite: &dialectAggConfig{
		funcName: "MAX",
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			return eb.Case(func(cb CaseBuilder) {
				cb.WhenExpr(state.argsExpr).Then(1).Else(0)
			})
		},
	},
}

var boolAndStrategy = &dialectStrategy{
	postgres: &dialectAggConfig{funcName: "BOOL_AND"},
	mysql: &dialectAggConfig{
		funcName: "MIN",
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			return eb.Case(func(cb CaseBuilder) {
				cb.WhenExpr(state.argsExpr).Then(1).Else(0)
			})
		},
	},
	sqlite: &dialectAggConfig{
		funcName: "MIN",
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			return eb.Case(func(cb CaseBuilder) {
				cb.WhenExpr(state.argsExpr).Then(1).Else(0)
			})
		},
	},
}

var jsonArrayAggStrategy = &dialectStrategy{
	postgres: &dialectAggConfig{funcName: "JSON_AGG"},
	mysql:    &dialectAggConfig{funcName: "JSON_ARRAYAGG"},
	sqlite:   &dialectAggConfig{funcName: "JSON_GROUP_ARRAY"},
}

var jsonObjectAggStrategy = &dialectStrategy{
	postgres: &dialectAggConfig{funcName: "JSON_OBJECT_AGG"},
	mysql:    &dialectAggConfig{funcName: "JSON_OBJECTAGG"},
	sqlite:   &dialectAggConfig{funcName: "JSON_GROUP_OBJECT"},
}

var arrayAggStrategy = &dialectStrategy{
	postgres: &dialectAggConfig{funcName: "ARRAY_AGG"},
	mysql: &dialectAggConfig{
		funcName: "JSON_ARRAYAGG",
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			if state.nullsMode == NullsIgnore {
				return eb.Case(func(cb CaseBuilder) {
					cb.WhenExpr(eb.IsNotNull(state.argsExpr)).Then(state.argsExpr)
				})
			}

			return state.argsExpr
		},
		clearDistinct:  true,
		clearOrderBy:   true,
		clearNullsMode: true,
	},
	sqlite: &dialectAggConfig{
		funcName: "JSON_GROUP_ARRAY",
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			if state.nullsMode == NullsIgnore {
				return eb.Case(func(cb CaseBuilder) {
					cb.WhenExpr(eb.IsNotNull(state.argsExpr)).Then(state.argsExpr)
				})
			}

			return state.argsExpr
		},
		clearDistinct:  true,
		clearOrderBy:   true,
		clearNullsMode: true,
	},
}

var stringAggStrategy = &dialectStrategy{
	postgres: &dialectAggConfig{
		funcName: "STRING_AGG",
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			argsExpr := state.argsExpr
			if state.nullsMode == NullsIgnore {
				argsExpr = eb.Case(func(cb CaseBuilder) {
					cb.WhenExpr(eb.IsNotNull(state.argsExpr)).Then(state.argsExpr)
				})
			}

			return eb.Expr("?, ?", argsExpr, state.separator)
		},
	},
	mysql: &dialectAggConfig{
		funcName: "GROUP_CONCAT",
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			argsExpr := state.argsExpr
			if state.nullsMode == NullsIgnore {
				argsExpr = eb.Case(func(cb CaseBuilder) {
					cb.WhenExpr(eb.IsNotNull(state.argsExpr)).Then(state.argsExpr)
				})
			}
			if len(state.orderExprs) > 0 {
				return eb.Expr("? ? SEPARATOR ?", argsExpr, newOrderByClause(state.orderExprs...), state.separator)
			}

			return eb.Expr("? SEPARATOR ?", argsExpr, state.separator)
		},
		clearOrderBy:   true,
		clearNullsMode: true,
	},
	sqlite: &dialectAggConfig{
		funcName: "GROUP_CONCAT",
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			argsExpr := state.argsExpr
			if state.nullsMode == NullsIgnore {
				argsExpr = eb.Case(func(cb CaseBuilder) {
					cb.WhenExpr(eb.IsNotNull(state.argsExpr)).Then(state.argsExpr)
				})
			}
			// SQLite DISTINCT limitation: only one argument allowed
			if state.distinct {
				return argsExpr
			}

			return eb.Expr("?, ?", argsExpr, state.separator)
		},
		clearNullsMode: true,
	},
}

var stdDevStrategy = &dialectStrategy{
	postgres: &dialectAggConfig{
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			mode := lo.CoalesceOrEmpty(state.statisticalMode.String(), StatisticalPopulation.String())
			state.funcName = "STDDEV" + constants.Underscore + mode

			return state.argsExpr
		},
	},
	mysql: &dialectAggConfig{
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			if state.statisticalMode == StatisticalPopulation || state.statisticalMode == StatisticalSample {
				state.funcName = "STDDEV" + constants.Underscore + state.statisticalMode.String()
			} else {
				state.funcName = "STDDEV"
			}

			return state.argsExpr
		},
	},
}

var varianceStrategy = &dialectStrategy{
	postgres: &dialectAggConfig{
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			mode := lo.CoalesceOrEmpty(state.statisticalMode.String(), StatisticalPopulation.String())
			state.funcName = "VAR" + constants.Underscore + mode

			return state.argsExpr
		},
	},
	mysql: &dialectAggConfig{
		argsTransformer: func(eb ExprBuilder, state *aggregateQueryState) schema.QueryAppender {
			if state.statisticalMode == StatisticalPopulation || state.statisticalMode == StatisticalSample {
				state.funcName = "VAR" + constants.Underscore + state.statisticalMode.String()
			} else {
				state.funcName = "VARIANCE"
			}

			return state.argsExpr
		},
	},
}

type aggregateQueryState struct {
	funcName        string
	argsExpr        schema.QueryAppender
	distinct        bool
	filter          schema.QueryAppender
	orderExprs      []orderExpr
	nullsMode       NullsMode
	separator       string
	statisticalMode StatisticalMode
}

type baseAggregateExpr struct {
	qb         QueryBuilder
	eb         ExprBuilder
	funcName   string
	argsExpr   schema.QueryAppender
	distinct   bool
	filter     schema.QueryAppender
	orderExprs []orderExpr
	nullsMode  NullsMode
	strategy   *dialectStrategy
}

func (a *baseAggregateExpr) getDialectConfig() *dialectAggConfig {
	if a.strategy == nil {
		return nil
	}

	var cfg *dialectAggConfig
	a.eb.ExecByDialect(DialectExecs{
		Postgres:  func() { cfg = a.strategy.postgres },
		MySQL:     func() { cfg = a.strategy.mysql },
		SQLite:    func() { cfg = a.strategy.sqlite },
		Oracle:    func() { cfg = a.strategy.oracle },
		SQLServer: func() { cfg = a.strategy.sqlsrv },
	})

	return cfg
}

func (a *baseAggregateExpr) buildQueryState() aggregateQueryState {
	return aggregateQueryState{
		funcName:   a.funcName,
		argsExpr:   a.argsExpr,
		distinct:   a.distinct,
		filter:     a.filter,
		orderExprs: a.orderExprs,
		nullsMode:  a.nullsMode,
	}
}

func (a *baseAggregateExpr) applyDialectConfig(state *aggregateQueryState, cfg *dialectAggConfig) {
	if cfg == nil {
		return
	}

	if cfg.funcName != constants.Empty {
		state.funcName = cfg.funcName
	}

	if cfg.argsTransformer != nil {
		state.argsExpr = cfg.argsTransformer(a.eb, state)
	}

	if cfg.clearDistinct {
		state.distinct = false
	}

	if cfg.clearOrderBy {
		state.orderExprs = nil
	}

	if cfg.clearNullsMode {
		state.nullsMode = NullsDefault
	}
}

func (a *baseAggregateExpr) dialectAwareAppendQuery(gen schema.QueryGen, b []byte) ([]byte, error) {
	state := a.buildQueryState()

	cfg := a.getDialectConfig()
	if cfg == nil {
		return nil, ErrDialectUnsupportedOperation
	}

	a.applyDialectConfig(&state, cfg)

	return a.appendQueryWithState(gen, b, state)
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

func (a *baseAggregateExpr) AppendQuery(gen schema.QueryGen, b []byte) ([]byte, error) {
	return a.appendQueryWithState(gen, b, a.buildQueryState())
}

func (a *baseAggregateExpr) appendQueryWithState(gen schema.QueryGen, b []byte, state aggregateQueryState) (_ []byte, err error) {
	if state.argsExpr == nil {
		return nil, ErrAggregateMissingArgs
	}

	if state.filter != nil {
		var handled bool

		a.eb.ExecByDialect(DialectExecs{
			MySQL: func() {
				b, err = a.appendCompatibleFilterQueryWithState(gen, b, state)
				handled = true
			},
			Oracle: func() {
				b, err = a.appendCompatibleFilterQueryWithState(gen, b, state)
				handled = true
			},
			SQLServer: func() {
				b, err = a.appendCompatibleFilterQueryWithState(gen, b, state)
				handled = true
			},
		})

		if handled {
			return b, err
		}
	}

	b = append(b, state.funcName...)
	b = append(b, constants.ByteLeftParenthesis)

	if state.distinct {
		b = append(b, "DISTINCT "...)
	}

	if b, err = state.argsExpr.AppendQuery(gen, b); err != nil {
		return
	}

	if len(state.orderExprs) > 0 {
		b = append(b, constants.ByteSpace)
		if b, err = newOrderByClause(state.orderExprs...).AppendQuery(gen, b); err != nil {
			return
		}
	}

	b = append(b, constants.ByteRightParenthesis)

	if state.nullsMode != NullsDefault {
		b = append(b, constants.ByteSpace)
		b = append(b, state.nullsMode.String()...)
	}

	if state.filter != nil {
		if b, err = newFilterClause(state.filter).AppendQuery(gen, b); err != nil {
			return
		}
	}

	return b, nil
}

func (a *baseAggregateExpr) appendCompatibleFilterQueryWithState(gen schema.QueryGen, b []byte, state aggregateQueryState) (_ []byte, err error) {
	funcName := state.funcName

	switch funcName {
	case "COUNT":
		b = append(b, "SUM"...)
	default:
		b = append(b, funcName...)
	}

	b = append(b, constants.ByteLeftParenthesis)

	if state.distinct {
		b = append(b, "DISTINCT "...)
	}

	if b, err = a.eb.Case(func(cb CaseBuilder) {
		when := cb.WhenExpr(state.filter)
		switch funcName {
		case "COUNT":
			when.Then(1)
		default:
			when.Then(state.argsExpr)
		}

		switch funcName {
		case "COUNT", "SUM":
			cb.Else(0)
		default:
			cb.Else(a.eb.Null())
		}
	}).AppendQuery(gen, b); err != nil {
		return
	}

	b = append(b, constants.ByteRightParenthesis)

	return b, nil
}

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

type distinctableAggregateBuilder[T any] struct {
	*baseAggregateBuilder[T]
}

func (b *distinctableAggregateBuilder[T]) Distinct() T {
	b.distinct = true

	return b.self
}

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

type statisticalAggregateBuilder[T any] struct {
	*baseAggregateBuilder[T]

	statExpr *statisticalAggExpr
}

func (b *statisticalAggregateBuilder[T]) Population() T {
	b.statExpr.statisticalMode = StatisticalPopulation

	return b.self
}

func (b *statisticalAggregateBuilder[T]) Sample() T {
	b.statExpr.statisticalMode = StatisticalSample

	return b.self
}

type countExpr[T any] struct {
	*baseAggregateExpr
	*distinctableAggregateBuilder[T]
}

func (c *countExpr[T]) All() T {
	c.argsExpr = bun.Safe(columnAll)

	return c.self
}

type sumExpr[T any] struct {
	*baseAggregateExpr
	*distinctableAggregateBuilder[T]
}

type avgExpr[T any] struct {
	*baseAggregateExpr
	*distinctableAggregateBuilder[T]
}

type minExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
}

type maxExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
}

type stringAggExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
	*distinctableAggregateBuilder[T]
	*orderableAggregateBuilder[T]
	*baseNullHandlingBuilder[T]

	separator string
}

func (s *stringAggExpr[T]) Separator(separator string) T {
	s.separator = separator

	return s.self
}

func (s *stringAggExpr[T]) AppendQuery(gen schema.QueryGen, b []byte) ([]byte, error) {
	state := s.buildQueryState()
	state.separator = s.separator

	cfg := s.getDialectConfig()
	if cfg == nil {
		return nil, ErrDialectUnsupportedOperation
	}

	s.applyDialectConfig(&state, cfg)

	return s.appendQueryWithState(gen, b, state)
}

type arrayAggExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
	*distinctableAggregateBuilder[T]
	*orderableAggregateBuilder[T]
	*baseNullHandlingBuilder[T]
}

func (a *arrayAggExpr[T]) AppendQuery(gen schema.QueryGen, b []byte) ([]byte, error) {
	return a.dialectAwareAppendQuery(gen, b)
}

type statisticalAggExpr struct {
	*baseAggregateExpr

	statisticalMode StatisticalMode
}

func (s *statisticalAggExpr) AppendQuery(gen schema.QueryGen, b []byte) ([]byte, error) {
	state := s.buildQueryState()
	state.statisticalMode = s.statisticalMode

	cfg := s.getDialectConfig()
	if cfg == nil {
		return nil, ErrDialectUnsupportedOperation
	}

	s.applyDialectConfig(&state, cfg)

	return s.appendQueryWithState(gen, b, state)
}

type stdDevExpr[T any] struct {
	*statisticalAggExpr
	*statisticalAggregateBuilder[T]
}

type varianceExpr[T any] struct {
	*statisticalAggExpr
	*statisticalAggregateBuilder[T]
}

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

func (j *jsonObjectAggExpr[T]) AppendQuery(gen schema.QueryGen, b []byte) ([]byte, error) {
	if j.keyExpr == nil {
		return nil, ErrAggregateMissingArgs
	}

	state := j.buildQueryState()

	cfg := j.getDialectConfig()
	if cfg == nil {
		return nil, ErrDialectUnsupportedOperation
	}

	j.applyDialectConfig(&state, cfg)
	state.argsExpr = j.eb.Exprs(j.keyExpr, j.argsExpr)

	return j.appendQueryWithState(gen, b, state)
}

type jsonArrayAggExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
	*distinctableAggregateBuilder[T]
	*orderableAggregateBuilder[T]
}

func (j *jsonArrayAggExpr[T]) AppendQuery(gen schema.QueryGen, b []byte) ([]byte, error) {
	return j.dialectAwareAppendQuery(gen, b)
}

type bitOrExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
}

func (b *bitOrExpr[T]) AppendQuery(gen schema.QueryGen, buf []byte) ([]byte, error) {
	return b.dialectAwareAppendQuery(gen, buf)
}

type bitAndExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
}

func (b *bitAndExpr[T]) AppendQuery(gen schema.QueryGen, buf []byte) ([]byte, error) {
	return b.dialectAwareAppendQuery(gen, buf)
}

type boolOrExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
}

func (b *boolOrExpr[T]) AppendQuery(gen schema.QueryGen, buf []byte) ([]byte, error) {
	return b.dialectAwareAppendQuery(gen, buf)
}

type boolAndExpr[T any] struct {
	*baseAggregateExpr
	*baseAggregateBuilder[T]
}

func (b *boolAndExpr[T]) AppendQuery(gen schema.QueryGen, buf []byte) ([]byte, error) {
	return b.dialectAwareAppendQuery(gen, buf)
}

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

func newGenericStringAggExpr[T any](self T, qb QueryBuilder) *stringAggExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: stringAggStrategy,
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	expr := &stringAggExpr[T]{
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
		separator: constants.Comma,
	}

	baseBuilder.self = self

	return expr
}

func newStringAggExpr(qb QueryBuilder) *stringAggExpr[StringAggBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: stringAggStrategy,
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
		separator: constants.Comma,
	}

	baseBuilder.self = expr

	return expr
}

func newGenericArrayAggExpr[T any](self T, qb QueryBuilder) *arrayAggExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: arrayAggStrategy,
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

func newArrayAggExpr(qb QueryBuilder) *arrayAggExpr[ArrayAggBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: arrayAggStrategy,
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

func newGenericStdDevExpr[T any](self T, qb QueryBuilder) *stdDevExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: stdDevStrategy,
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	statExpr := &statisticalAggExpr{
		baseAggregateExpr: baseExpr,
	}
	expr := &stdDevExpr[T]{
		statisticalAggExpr: statExpr,
		statisticalAggregateBuilder: &statisticalAggregateBuilder[T]{
			baseAggregateBuilder: baseBuilder,
			statExpr:             statExpr,
		},
	}

	baseBuilder.self = self

	return expr
}

func newStdDevExpr(qb QueryBuilder) *stdDevExpr[StdDevBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: stdDevStrategy,
	}
	baseBuilder := &baseAggregateBuilder[StdDevBuilder]{
		baseAggregateExpr: baseExpr,
	}
	statExpr := &statisticalAggExpr{
		baseAggregateExpr: baseExpr,
	}
	expr := &stdDevExpr[StdDevBuilder]{
		statisticalAggExpr: statExpr,
		statisticalAggregateBuilder: &statisticalAggregateBuilder[StdDevBuilder]{
			baseAggregateBuilder: baseBuilder,
			statExpr:             statExpr,
		},
	}

	baseBuilder.self = expr

	return expr
}

func newGenericVarianceExpr[T any](self T, qb QueryBuilder) *varianceExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: varianceStrategy,
	}
	baseBuilder := &baseAggregateBuilder[T]{
		baseAggregateExpr: baseExpr,
	}
	statExpr := &statisticalAggExpr{
		baseAggregateExpr: baseExpr,
	}
	expr := &varianceExpr[T]{
		statisticalAggExpr: statExpr,
		statisticalAggregateBuilder: &statisticalAggregateBuilder[T]{
			baseAggregateBuilder: baseBuilder,
			statExpr:             statExpr,
		},
	}

	baseBuilder.self = self

	return expr
}

func newVarianceExpr(qb QueryBuilder) *varianceExpr[VarianceBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: varianceStrategy,
	}
	baseBuilder := &baseAggregateBuilder[VarianceBuilder]{
		baseAggregateExpr: baseExpr,
	}
	statExpr := &statisticalAggExpr{
		baseAggregateExpr: baseExpr,
	}
	expr := &varianceExpr[VarianceBuilder]{
		statisticalAggExpr: statExpr,
		statisticalAggregateBuilder: &statisticalAggregateBuilder[VarianceBuilder]{
			baseAggregateBuilder: baseBuilder,
			statExpr:             statExpr,
		},
	}

	baseBuilder.self = expr

	return expr
}

func newGenericJSONObjectAggExpr[T any](self T, qb QueryBuilder) *jsonObjectAggExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: jsonObjectAggStrategy,
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

func newJSONObjectAggExpr(qb QueryBuilder) *jsonObjectAggExpr[JSONObjectAggBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: jsonObjectAggStrategy,
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

func newGenericJSONArrayAggExpr[T any](self T, qb QueryBuilder) *jsonArrayAggExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: jsonArrayAggStrategy,
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

func newJSONArrayAggExpr(qb QueryBuilder) *jsonArrayAggExpr[JSONArrayAggBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: jsonArrayAggStrategy,
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

func newGenericBitOrExpr[T any](self T, qb QueryBuilder) *bitOrExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: bitOrStrategy,
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

func newBitOrExpr(qb QueryBuilder) *bitOrExpr[BitOrBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: bitOrStrategy,
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

func newGenericBitAndExpr[T any](self T, qb QueryBuilder) *bitAndExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: bitAndStrategy,
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

func newBitAndExpr(qb QueryBuilder) *bitAndExpr[BitAndBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: bitAndStrategy,
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

func newGenericBoolOrExpr[T any](self T, qb QueryBuilder) *boolOrExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: boolOrStrategy,
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

func newBoolOrExpr(qb QueryBuilder) *boolOrExpr[BoolOrBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: boolOrStrategy,
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

func newGenericBoolAndExpr[T any](self T, qb QueryBuilder) *boolAndExpr[T] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: boolAndStrategy,
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

func newBoolAndExpr(qb QueryBuilder) *boolAndExpr[BoolAndBuilder] {
	baseExpr := &baseAggregateExpr{
		qb:       qb,
		eb:       qb.ExprBuilder(),
		strategy: boolAndStrategy,
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
