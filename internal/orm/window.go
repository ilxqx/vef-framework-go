package orm

import (
	"strconv"

	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/sort"
)

// WindowPartitionable defines window functions that support partitioning.
type WindowPartitionable[T any] interface {
	// Over starts configuring the OVER clause for the window function.
	Over() T
}

// BaseWindowPartitionBuilder defines the base window partition builder interface.
type BaseWindowPartitionBuilder[T any] interface {
	// PartitionBy adds PARTITION BY columns to the window definition.
	PartitionBy(columns ...string) T
	// PartitionByExpr adds a raw PARTITION BY expression to the window definition.
	PartitionByExpr(expr any) T
	// OrderBy adds ORDER BY clauses with ascending direction.
	OrderBy(columns ...string) T
	// OrderByDesc adds ORDER BY clauses with descending direction.
	OrderByDesc(columns ...string) T
	// OrderByExpr adds an ORDER BY clause using a raw expression.
	OrderByExpr(expr any) T
}

// WindowPartitionBuilder defines the window partition builder interface.
type WindowPartitionBuilder interface {
	BaseWindowPartitionBuilder[WindowPartitionBuilder]
}

// WindowFrameablePartitionBuilder defines window functions that support partitioning and frame specification.
type WindowFrameablePartitionBuilder interface {
	BaseWindowPartitionBuilder[WindowFrameablePartitionBuilder]
	// Rows configures a ROWS frame clause.
	Rows() WindowFrameBuilder
	// Range configures a RANGE frame clause.
	Range() WindowFrameBuilder
	// Groups configures a GROUPS frame clause.
	Groups() WindowFrameBuilder
}

// WindowBoundable defines window frame boundaries.
type WindowBoundable[T any] interface {
	// CurrentRow sets the boundary to CURRENT ROW.
	CurrentRow() T
	// Preceding sets the boundary to N PRECEDING.
	Preceding(n int) T
	// Following sets the boundary to N FOLLOWING.
	Following(n int) T
}

// WindowStartBoundable defines window frame start boundaries.
type WindowStartBoundable[T any] interface {
	WindowBoundable[T]

	// UnboundedPreceding sets the start boundary to UNBOUNDED PRECEDING.
	UnboundedPreceding() T
}

// WindowEndBoundable defines window frame end boundaries.
type WindowEndBoundable[T any] interface {
	WindowStartBoundable[T]

	// UnboundedFollowing sets the end boundary to UNBOUNDED FOLLOWING.
	UnboundedFollowing() T
}

// WindowFrameBuilder defines the window frame builder interface.
type WindowFrameBuilder interface {
	WindowStartBoundable[WindowFrameBuilder]

	// And switches to configuring the end boundary for BETWEEN ... AND ... syntax.
	And() WindowFrameEndBuilder
}

// WindowFrameEndBuilder defines the window frame end boundary builder interface.
type WindowFrameEndBuilder interface {
	WindowEndBoundable[WindowFrameEndBuilder]
}

// ========== Ranking Window Functions ==========

// RowNumberBuilder defines the ROW_NUMBER() window function builder.
type RowNumberBuilder interface {
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// RankBuilder defines the RANK() window function builder.
type RankBuilder interface {
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// DenseRankBuilder defines the DENSE_RANK() window function builder.
type DenseRankBuilder interface {
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// PercentRankBuilder defines the PERCENT_RANK() window function builder.
type PercentRankBuilder interface {
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// CumeDistBuilder defines the CUME_DIST() window function builder.
type CumeDistBuilder interface {
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// NtileBuilder defines the NTILE(n) window function builder.
type NtileBuilder interface {
	WindowPartitionable[WindowFrameablePartitionBuilder]

	Buckets(n int) NtileBuilder
}

// ========== Value Window Functions ==========

// LagBuilder defines the LAG() window function builder.
type LagBuilder interface {
	WindowPartitionable[WindowPartitionBuilder]

	Column(column string) LagBuilder
	Expr(expr any) LagBuilder
	Offset(offset int) LagBuilder      // Number of rows to lag (default 1)
	DefaultValue(value any) LagBuilder // Default value when no previous row exists
}

// LeadBuilder defines the LEAD() window function builder.
type LeadBuilder interface {
	WindowPartitionable[WindowPartitionBuilder]

	Column(column string) LeadBuilder
	Expr(expr any) LeadBuilder
	Offset(offset int) LeadBuilder      // Number of rows to lead (default 1)
	DefaultValue(value any) LeadBuilder // Default value when no next row exists
}

// FirstValueBuilder defines the FIRST_VALUE() window function builder.
type FirstValueBuilder interface {
	WindowPartitionable[WindowFrameablePartitionBuilder]
	NullHandlingBuilder[FirstValueBuilder]

	Column(column string) FirstValueBuilder
	Expr(expr any) FirstValueBuilder
}

// LastValueBuilder defines the LAST_VALUE() window function builder.
type LastValueBuilder interface {
	WindowPartitionable[WindowFrameablePartitionBuilder]
	NullHandlingBuilder[LastValueBuilder]

	Column(column string) LastValueBuilder
	Expr(expr any) LastValueBuilder
}

// NthValueBuilder defines the NTH_VALUE() window function builder.
type NthValueBuilder interface {
	WindowPartitionable[WindowFrameablePartitionBuilder]
	NullHandlingBuilder[NthValueBuilder]

	Column(column string) NthValueBuilder
	Expr(expr any) NthValueBuilder
	N(n int) NthValueBuilder
	FromFirst() NthValueBuilder
	FromLast() NthValueBuilder
}

// ========== Aggregate Window Functions ==========

// WindowCountBuilder defines COUNT() as window function builder.
type WindowCountBuilder interface {
	BaseAggregate[WindowCountBuilder]
	DistinctableAggregate[WindowCountBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]

	All() WindowCountBuilder
}

// WindowSumBuilder defines SUM() as window function builder.
type WindowSumBuilder interface {
	BaseAggregate[WindowSumBuilder]
	DistinctableAggregate[WindowSumBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// WindowAvgBuilder defines AVG() as window function builder.
type WindowAvgBuilder interface {
	BaseAggregate[WindowAvgBuilder]
	DistinctableAggregate[WindowAvgBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// WindowMinBuilder defines MIN() as window function builder.
type WindowMinBuilder interface {
	BaseAggregate[WindowMinBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// WindowMaxBuilder defines MAX() as window function builder.
type WindowMaxBuilder interface {
	BaseAggregate[WindowMaxBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// WindowStringAggBuilder defines STRING_AGG() as window function builder.
type WindowStringAggBuilder interface {
	DistinctableAggregate[WindowStringAggBuilder]
	OrderableAggregate[WindowStringAggBuilder]
	NullHandlingBuilder[WindowStringAggBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]

	Separator(separator string) WindowStringAggBuilder
}

// WindowArrayAggBuilder defines ARRAY_AGG() as window function builder.
type WindowArrayAggBuilder interface {
	DistinctableAggregate[WindowArrayAggBuilder]
	OrderableAggregate[WindowArrayAggBuilder]
	NullHandlingBuilder[WindowArrayAggBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// WindowStdDevBuilder defines STDDEV() as window function builder.
type WindowStdDevBuilder interface {
	BaseAggregate[WindowStdDevBuilder]
	StatisticalAggregate[WindowStdDevBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// WindowVarianceBuilder defines VARIANCE() as window function builder.
type WindowVarianceBuilder interface {
	BaseAggregate[WindowVarianceBuilder]
	StatisticalAggregate[WindowVarianceBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// WindowJSONObjectAggBuilder defines JSON_OBJECT_AGG() as window function builder.
type WindowJSONObjectAggBuilder interface {
	BaseAggregate[WindowJSONObjectAggBuilder]
	DistinctableAggregate[WindowJSONObjectAggBuilder]
	OrderableAggregate[WindowJSONObjectAggBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]

	// KeyColumn sets the key using a column reference.
	KeyColumn(column string) WindowJSONObjectAggBuilder
	// KeyExpr sets the key using a expression.
	KeyExpr(expr any) WindowJSONObjectAggBuilder
}

// WindowJSONArrayAggBuilder defines JSON_ARRAY_AGG() as window function builder.
type WindowJSONArrayAggBuilder interface {
	BaseAggregate[WindowJSONArrayAggBuilder]
	DistinctableAggregate[WindowJSONArrayAggBuilder]
	OrderableAggregate[WindowJSONArrayAggBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// WindowBitOrBuilder defines BIT_OR() as window function builder.
type WindowBitOrBuilder interface {
	BaseAggregate[WindowBitOrBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// WindowBitAndBuilder defines BIT_AND() as window function builder.
type WindowBitAndBuilder interface {
	BaseAggregate[WindowBitAndBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// WindowBoolOrBuilder defines BOOL_OR() as window function builder.
type WindowBoolOrBuilder interface {
	BaseAggregate[WindowBoolOrBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// WindowBoolAndBuilder defines BOOL_AND() as window function builder.
type WindowBoolAndBuilder interface {
	BaseAggregate[WindowBoolAndBuilder]
	WindowPartitionable[WindowFrameablePartitionBuilder]
}

// ========== Window Function Implementations ==========

// partitionExpr implements the partition expression.
type partitionExpr struct {
	builders ExprBuilder
	column   string
	expr     any
}

func (p *partitionExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	if p.column != constants.Empty {
		return p.builders.Column(p.column).AppendQuery(fmter, b)
	} else if p.expr != nil {
		return p.builders.Expr("?", p.expr).AppendQuery(fmter, b)
	}

	return b, nil
}

// baseWindowExpr implements common functionality for all window function expressions.
type baseWindowExpr struct {
	// eb provides access to expression building utilities (columns, expressions, etc.)
	eb ExprBuilder
	// funcExpr holds a pre-built aggregate function expression (used for window aggregate functions)
	funcExpr schema.QueryAppender
	// funcName stores the window function name (e.g., "ROW_NUMBER", "LAG", "SUM")
	funcName string
	// args contains the arguments passed to the window function
	args []any
	// nullsMode controls NULL handling behavior for value window functions (IGNORE/RESPECT NULLS)
	nullsMode NullsMode
	// fromDir indicates the direction for NTH_VALUE function (FROM FIRST/FROM LAST)
	fromDir FromDirection
	// partitionExprs contains PARTITION BY expressions for the OVER clause
	partitionExprs []partitionExpr
	// orderExprs contains ORDER BY expressions for the OVER clause
	orderExprs []orderExpr
	// frameType specifies the window frame type (ROWS/RANGE/GROUPS)
	frameType FrameType
	// frameStartKind defines the start boundary type of the window frame
	frameStartKind FrameBoundKind // Frame start boundary
	// frameStartN stores the numeric value for PRECEDING/FOLLOWING start boundaries
	frameStartN int
	// frameEndKind defines the end boundary type of the window frame (for BETWEEN ... AND ...)
	frameEndKind FrameBoundKind // Frame end boundary (for BETWEEN ... AND ...)
	// frameEndN stores the numeric value for PRECEDING/FOLLOWING end boundaries
	frameEndN int
}

func (w *baseWindowExpr) setArgs(args ...any) {
	w.args = args
}

func (w *baseWindowExpr) appendPartitionBy(columns ...string) {
	for _, column := range columns {
		w.partitionExprs = append(w.partitionExprs, partitionExpr{
			builders: w.eb,
			column:   column,
		})
	}
}

func (w *baseWindowExpr) appendPartitionByExpr(expr any) {
	w.partitionExprs = append(w.partitionExprs, partitionExpr{
		builders: w.eb,
		expr:     expr,
	})
}

func (w *baseWindowExpr) appendOrderBy(columns ...string) {
	for _, column := range columns {
		w.orderExprs = append(w.orderExprs, orderExpr{
			builders:   w.eb,
			column:     column,
			direction:  sort.OrderAsc,
			nullsOrder: sort.NullsDefault,
		})
	}
}

func (w *baseWindowExpr) appendOrderByDesc(columns ...string) {
	for _, column := range columns {
		w.orderExprs = append(w.orderExprs, orderExpr{
			builders:   w.eb,
			column:     column,
			direction:  sort.OrderDesc,
			nullsOrder: sort.NullsDefault,
		})
	}
}

func (w *baseWindowExpr) appendOrderByExpr(expr any) {
	w.orderExprs = append(w.orderExprs, orderExpr{
		builders:   w.eb,
		expr:       expr,
		direction:  sort.OrderAsc,
		nullsOrder: sort.NullsDefault,
	})
}

func (w *baseWindowExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	if w.funcExpr == nil {
		// Function name and arguments
		b = append(b, w.funcName...)
		b = append(b, constants.ByteLeftParenthesis)

		// Function arguments
		if len(w.args) > 0 {
			if b, err = w.eb.Exprs(w.args...).AppendQuery(fmter, b); err != nil {
				return
			}
		}

		b = append(b, constants.ByteRightParenthesis)

		// FROM DIRECTION and NULLS MODE support varies by database:
		// - Oracle: supports both FROM FIRST/LAST and IGNORE/RESPECT NULLS
		// - SQL Server: supports IGNORE/RESPECT NULLS but not FROM FIRST/LAST
		// - PostgreSQL, MySQL, SQLite: support neither
		// Use dialect-specific logic to generate appropriate SQL for each database
		if w.fromDir != FromDefault || w.nullsMode != NullsDefault {
			dialectBytes, err := w.eb.RunDialectFunc(DialectFuncs{
				// Oracle supports both FROM FIRST/LAST and IGNORE/RESPECT NULLS
				Oracle: func() ([]byte, error) {
					var dialectB []byte
					if w.fromDir != FromDefault {
						dialectB = append(dialectB, constants.ByteSpace)
						dialectB = append(dialectB, w.fromDir.String()...)
					}

					if w.nullsMode != NullsDefault {
						dialectB = append(dialectB, constants.ByteSpace)
						dialectB = append(dialectB, w.nullsMode.String()...)
					}

					return dialectB, nil
				},
				SQLServer: func() ([]byte, error) {
					var dialectB []byte
					// SQL Server doesn't support FROM FIRST/LAST clauses
					if w.nullsMode != NullsDefault {
						dialectB = append(dialectB, constants.ByteSpace)
						dialectB = append(dialectB, w.nullsMode.String()...)
					}

					return dialectB, nil
				},
				Default: func() ([]byte, error) {
					// For PostgreSQL, MySQL, SQLite: do nothing
					// These databases don't support FROM FIRST/LAST or IGNORE/RESPECT NULLS
					return nil, nil
				},
			})
			if err != nil {
				return b, err
			}

			b = append(b, dialectBytes...)
		}
	} else {
		if b, err = w.funcExpr.AppendQuery(fmter, b); err != nil {
			return
		}
	}

	// OVER clause
	b = append(b, " OVER "...)
	b = append(b, constants.ByteLeftParenthesis)

	// PARTITION BY clause
	if len(w.partitionExprs) > 0 {
		b = append(b, "PARTITION BY "...)

		for i, expr := range w.partitionExprs {
			if i > 0 {
				b = append(b, constants.CommaSpace...)
			}

			if b, err = expr.AppendQuery(fmter, b); err != nil {
				return
			}
		}
	}

	// ORDER BY clause
	if len(w.orderExprs) > 0 {
		b = append(b, constants.ByteSpace)
		if b, err = newOrderByClause(w.orderExprs...).AppendQuery(fmter, b); err != nil {
			return
		}
	}

	// Frame clause
	if w.frameType != FrameDefault {
		if len(w.partitionExprs) > 0 || len(w.orderExprs) > 0 {
			b = append(b, constants.ByteSpace)
		}

		b = append(b, w.frameType.String()...)

		b = append(b, constants.ByteSpace)
		if w.frameEndKind != FrameBoundNone {
			// Use BETWEEN syntax when both start and end bounds are present
			b = append(b, "BETWEEN "...)
			b = w.appendFrameBound(b, w.frameStartKind, w.frameStartN)
			b = append(b, " AND "...)
			b = w.appendFrameBound(b, w.frameEndKind, w.frameEndN)
		} else {
			b = w.appendFrameBound(b, w.frameStartKind, w.frameStartN)
		}
	}

	b = append(b, constants.ByteRightParenthesis)

	return b, nil
}

func (w *baseWindowExpr) appendFrameBound(b []byte, kind FrameBoundKind, n int) []byte {
	switch kind {
	case FrameBoundUnboundedPreceding, FrameBoundUnboundedFollowing, FrameBoundCurrentRow:
		return append(b, kind.String()...)
	case FrameBoundPreceding, FrameBoundFollowing:
		b = strconv.AppendInt(b, int64(n), 10)
		b = append(b, constants.ByteSpace)

		return append(b, kind.String()...)

	default:
		return b
	}
}

// baseWindowPartitionBuilder is a base struct for all window function builders.
type baseWindowPartitionBuilder[T any] struct {
	*baseWindowExpr

	self T
}

func (b *baseWindowPartitionBuilder[T]) PartitionBy(columns ...string) T {
	b.appendPartitionBy(columns...)

	return b.self
}

func (b *baseWindowPartitionBuilder[T]) PartitionByExpr(expr any) T {
	b.appendPartitionByExpr(expr)

	return b.self
}

func (b *baseWindowPartitionBuilder[T]) OrderBy(columns ...string) T {
	b.appendOrderBy(columns...)

	return b.self
}

func (b *baseWindowPartitionBuilder[T]) OrderByDesc(columns ...string) T {
	b.appendOrderByDesc(columns...)

	return b.self
}

func (b *baseWindowPartitionBuilder[T]) OrderByExpr(expr any) T {
	b.appendOrderByExpr(expr)

	return b.self
}

// baseWindowFrameablePartitionBuilder is a base struct for all window function builders.
type baseWindowFrameablePartitionBuilder[T any] struct {
	*baseWindowPartitionBuilder[T]
}

func (b *baseWindowFrameablePartitionBuilder[T]) Rows() WindowFrameBuilder {
	b.frameType = FrameRows

	return &windowFrameBuilder{baseWindowExpr: b.baseWindowExpr}
}

func (b *baseWindowFrameablePartitionBuilder[T]) Range() WindowFrameBuilder {
	b.frameType = FrameRange

	return &windowFrameBuilder{baseWindowExpr: b.baseWindowExpr}
}

func (b *baseWindowFrameablePartitionBuilder[T]) Groups() WindowFrameBuilder {
	b.frameType = FrameGroups

	return &windowFrameBuilder{baseWindowExpr: b.baseWindowExpr}
}

// windowFrameBuilder implements WindowFrameBuilder.
type windowFrameBuilder struct {
	*baseWindowExpr
}

func (b *windowFrameBuilder) UnboundedPreceding() WindowFrameBuilder {
	b.frameStartKind = FrameBoundUnboundedPreceding
	b.frameStartN = 0

	return b
}

func (b *windowFrameBuilder) CurrentRow() WindowFrameBuilder {
	b.frameStartKind = FrameBoundCurrentRow
	b.frameStartN = 0

	return b
}

func (b *windowFrameBuilder) Preceding(n int) WindowFrameBuilder {
	b.frameStartKind = FrameBoundPreceding
	b.frameStartN = n

	return b
}

func (b *windowFrameBuilder) Following(n int) WindowFrameBuilder {
	b.frameStartKind = FrameBoundFollowing
	b.frameStartN = n

	return b
}

func (b *windowFrameBuilder) And() WindowFrameEndBuilder {
	return &windowFrameEndBuilder{baseWindowExpr: b.baseWindowExpr}
}

// windowFrameEndBuilder implements WindowFrameEndBuilder.
type windowFrameEndBuilder struct {
	*baseWindowExpr
}

func (b *windowFrameEndBuilder) UnboundedPreceding() WindowFrameEndBuilder {
	b.frameEndKind = FrameBoundUnboundedPreceding
	b.frameEndN = 0

	return b
}

func (b *windowFrameEndBuilder) CurrentRow() WindowFrameEndBuilder {
	b.frameEndKind = FrameBoundCurrentRow
	b.frameEndN = 0

	return b
}

func (b *windowFrameEndBuilder) Preceding(n int) WindowFrameEndBuilder {
	b.frameEndKind = FrameBoundPreceding
	b.frameEndN = n

	return b
}

func (b *windowFrameEndBuilder) Following(n int) WindowFrameEndBuilder {
	b.frameEndKind = FrameBoundFollowing
	b.frameEndN = n

	return b
}

func (b *windowFrameEndBuilder) UnboundedFollowing() WindowFrameEndBuilder {
	b.frameEndKind = FrameBoundUnboundedFollowing
	b.frameEndN = 0

	return b
}

// baseWindowNullHandlingBuilder provides NULL handling functionality.
type baseWindowNullHandlingBuilder[T any] struct {
	*baseWindowExpr

	self T
}

func (b *baseWindowNullHandlingBuilder[T]) IgnoreNulls() T {
	b.nullsMode = NullsIgnore

	return b.self
}

func (b *baseWindowNullHandlingBuilder[T]) RespectNulls() T {
	b.nullsMode = NullsRespect

	return b.self
}

// ========== Ranking Window Functions ==========

// rowNumberExpr implements RowNumberBuilder.
type rowNumberExpr struct {
	*baseWindowExpr
}

func (rn *rowNumberExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: rn.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

// rankExpr implements RankBuilder.
type rankExpr struct {
	*baseWindowExpr
}

func (rn *rankExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: rn.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

// denseRankExpr implements DenseRankBuilder.
type denseRankExpr struct {
	*baseWindowExpr
}

func (rn *denseRankExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: rn.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

// percentRankExpr implements PercentRankBuilder.
type percentRankExpr struct {
	*baseWindowExpr
}

func (rn *percentRankExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: rn.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

// cumeDistExpr implements CumeDistBuilder.
type cumeDistExpr struct {
	*baseWindowExpr
}

func (rn *cumeDistExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: rn.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

// ntileExpr implements NtileBuilder.
type ntileExpr struct {
	*baseWindowExpr
}

func (ne *ntileExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: ne.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

func (ne *ntileExpr) Buckets(n int) NtileBuilder {
	ne.setArgs(n)

	return ne
}

// ========== Value Window Functions ==========

// lagExpr implements LagBuilder.
type lagExpr struct {
	*baseWindowExpr

	column       string
	expr         any
	offset       int
	defaultValue any
}

func (l *lagExpr) Over() WindowPartitionBuilder {
	builder := &baseWindowPartitionBuilder[WindowPartitionBuilder]{
		baseWindowExpr: l.baseWindowExpr,
	}

	builder.self = builder

	return builder
}

func (l *lagExpr) Column(column string) LagBuilder {
	l.column = column
	l.expr = nil

	return l
}

func (l *lagExpr) Expr(expr any) LagBuilder {
	l.expr = expr
	l.column = constants.Empty

	return l
}

func (l *lagExpr) Offset(offset int) LagBuilder {
	l.offset = offset

	return l
}

func (l *lagExpr) DefaultValue(value any) LagBuilder {
	l.defaultValue = value

	return l
}

func (l *lagExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	// LAG function has special syntax: LAG(column, offset, default) OVER (...)
	var args []any

	// Column or expression
	if l.column != constants.Empty {
		args = append(args, l.eb.Column(l.column))
	} else if l.expr != nil {
		args = append(args, l.expr)
	}

	if l.offset > 0 {
		args = append(args, l.offset)
	}

	if l.defaultValue != nil {
		args = append(args, l.defaultValue)
	}

	l.args = args

	return l.baseWindowExpr.AppendQuery(fmter, b)
}

// leadExpr implements LeadBuilder.
type leadExpr struct {
	*baseWindowExpr

	column       string
	expr         any
	offset       int
	defaultValue any
}

func (l *leadExpr) Over() WindowPartitionBuilder {
	builder := &baseWindowPartitionBuilder[WindowPartitionBuilder]{
		baseWindowExpr: l.baseWindowExpr,
	}

	builder.self = builder

	return builder
}

func (l *leadExpr) Column(column string) LeadBuilder {
	l.column = column
	l.expr = nil

	return l
}

func (l *leadExpr) Expr(expr any) LeadBuilder {
	l.expr = expr
	l.column = constants.Empty

	return l
}

func (l *leadExpr) Offset(offset int) LeadBuilder {
	l.offset = offset

	return l
}

func (l *leadExpr) DefaultValue(value any) LeadBuilder {
	l.defaultValue = value

	return l
}

func (l *leadExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	// LEAD function has special syntax: LEAD(column, offset, default) OVER (...)
	var args []any

	// Column or expression
	if l.column != constants.Empty {
		args = append(args, l.eb.Column(l.column))
	} else if l.expr != nil {
		args = append(args, l.expr)
	}

	if l.offset > 0 {
		args = append(args, l.offset)
	}

	if l.defaultValue != nil {
		args = append(args, l.defaultValue)
	}

	l.setArgs(args...)

	return l.baseWindowExpr.AppendQuery(fmter, b)
}

// firstValueExpr implements FirstValueBuilder.
type firstValueExpr struct {
	*baseWindowExpr
	*baseWindowNullHandlingBuilder[FirstValueBuilder]
}

func (fv *firstValueExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: fv.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

func (fv *firstValueExpr) Column(column string) FirstValueBuilder {
	fv.setArgs(fv.eb.Column(column))

	return fv
}

func (fv *firstValueExpr) Expr(expr any) FirstValueBuilder {
	fv.setArgs(expr)

	return fv
}

// lastValueExpr implements LastValueBuilder.
type lastValueExpr struct {
	*baseWindowExpr
	*baseWindowNullHandlingBuilder[LastValueBuilder]
}

func (lv *lastValueExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: lv.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

func (lv *lastValueExpr) Column(column string) LastValueBuilder {
	lv.setArgs(lv.eb.Column(column))

	return lv
}

func (lv *lastValueExpr) Expr(expr any) LastValueBuilder {
	lv.setArgs(expr)

	return lv
}

// nthValueExpr implements NthValueBuilder.
type nthValueExpr struct {
	*baseWindowExpr
	*baseWindowNullHandlingBuilder[NthValueBuilder]

	column string
	expr   any
	n      int
}

func (nv *nthValueExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: nv.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

func (nv *nthValueExpr) Column(column string) NthValueBuilder {
	nv.column = column
	nv.expr = nil

	return nv
}

func (nv *nthValueExpr) Expr(expr any) NthValueBuilder {
	nv.expr = expr
	nv.column = constants.Empty

	return nv
}

func (nv *nthValueExpr) N(nth int) NthValueBuilder {
	nv.n = nth

	return nv
}

func (nv *nthValueExpr) FromFirst() NthValueBuilder {
	nv.fromDir = FromFirst

	return nv
}

func (nv *nthValueExpr) FromLast() NthValueBuilder {
	nv.fromDir = FromLast

	return nv
}

func (nv *nthValueExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	// NTH_VALUE function has special syntax: NTH_VALUE(column, n) [FROM FIRST|FROM LAST] [IGNORE|RESPECT NULLS] OVER (...)
	var args []any
	if nv.column != constants.Empty {
		args = append(args, nv.eb.Column(nv.column))
	} else if nv.expr != nil {
		args = append(args, nv.expr)
	}

	args = append(args, nv.n)
	nv.setArgs(args...)

	return nv.baseWindowExpr.AppendQuery(fmter, b)
}

// ========== Window Aggregate Functions ==========

// windowCountExpr implements WindowCountBuilder.
type windowCountExpr struct {
	*countExpr[WindowCountBuilder]
	*baseWindowExpr
}

func (wc *windowCountExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: wc.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

func (wc *windowCountExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	wc.funcExpr = wc.countExpr

	return wc.baseWindowExpr.AppendQuery(fmter, b)
}

// windowSumExpr implements WindowSumBuilder.
type windowSumExpr struct {
	*sumExpr[WindowSumBuilder]
	*baseWindowExpr
}

func (ws *windowSumExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: ws.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

func (ws *windowSumExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	ws.funcExpr = ws.sumExpr

	return ws.baseWindowExpr.AppendQuery(fmter, b)
}

// windowAvgExpr implements WindowAvgBuilder.
type windowAvgExpr struct {
	*avgExpr[WindowAvgBuilder]
	*baseWindowExpr
}

func (wa *windowAvgExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: wa.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

func (wa *windowAvgExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	wa.funcExpr = wa.avgExpr

	return wa.baseWindowExpr.AppendQuery(fmter, b)
}

// windowMinExpr implements WindowMinBuilder.
type windowMinExpr struct {
	*minExpr[WindowMinBuilder]
	*baseWindowExpr
}

func (wm *windowMinExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: wm.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

func (wm *windowMinExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	wm.funcExpr = wm.minExpr

	return wm.baseWindowExpr.AppendQuery(fmter, b)
}

// windowMaxExpr implements WindowMaxBuilder.
type windowMaxExpr struct {
	*maxExpr[WindowMaxBuilder]
	*baseWindowExpr
}

func (wm *windowMaxExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: wm.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

func (wm *windowMaxExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	wm.funcExpr = wm.maxExpr

	return wm.baseWindowExpr.AppendQuery(fmter, b)
}

// windowStringAggExpr implements WindowStringAggBuilder.
type windowStringAggExpr struct {
	*stringAggExpr[WindowStringAggBuilder]
	*baseWindowExpr
}

func (ws *windowStringAggExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: ws.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

func (ws *windowStringAggExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	ws.funcExpr = ws.stringAggExpr

	return ws.baseWindowExpr.AppendQuery(fmter, b)
}

// windowArrayAggExpr implements WindowArrayAggBuilder.
type windowArrayAggExpr struct {
	*arrayAggExpr[WindowArrayAggBuilder]
	*baseWindowExpr
}

func (wa *windowArrayAggExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: wa.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

func (wa *windowArrayAggExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	wa.funcExpr = wa.arrayAggExpr

	return wa.baseWindowExpr.AppendQuery(fmter, b)
}

// windowStdDevExpr implements WindowStdDevBuilder.
type windowStdDevExpr struct {
	*stddevExpr[WindowStdDevBuilder]
	*baseWindowExpr
}

func (ws *windowStdDevExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: ws.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

func (ws *windowStdDevExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	ws.funcExpr = ws.stddevExpr

	return ws.baseWindowExpr.AppendQuery(fmter, b)
}

// windowVarianceExpr implements WindowVarianceBuilder.
type windowVarianceExpr struct {
	*varianceExpr[WindowVarianceBuilder]
	*baseWindowExpr
}

func (wv *windowVarianceExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: wv.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}

	baseBuilder.self = builder

	return builder
}

func (wv *windowVarianceExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	wv.funcExpr = wv.varianceExpr

	return wv.baseWindowExpr.AppendQuery(fmter, b)
}

// windowJsonObjectAggExpr implements WindowJsonObjectAggBuilder.
type windowJsonObjectAggExpr struct {
	*jsonObjectAggExpr[WindowJSONObjectAggBuilder]
	*baseWindowExpr
}

func (wj *windowJsonObjectAggExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: wj.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}
	baseBuilder.self = builder

	return builder
}

func (wj *windowJsonObjectAggExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	wj.funcExpr = wj.jsonObjectAggExpr

	return wj.baseWindowExpr.AppendQuery(fmter, b)
}

// windowJsonArrayAggExpr implements WindowJsonArrayAggBuilder.
type windowJsonArrayAggExpr struct {
	*jsonArrayAggExpr[WindowJSONArrayAggBuilder]
	*baseWindowExpr
}

func (wj *windowJsonArrayAggExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: wj.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}
	baseBuilder.self = builder

	return builder
}

func (wj *windowJsonArrayAggExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	wj.funcExpr = wj.jsonArrayAggExpr

	return wj.baseWindowExpr.AppendQuery(fmter, b)
}

// windowBitOrExpr implements WindowBitOrBuilder.
type windowBitOrExpr struct {
	*bitOrExpr[WindowBitOrBuilder]
	*baseWindowExpr
}

func (wb *windowBitOrExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: wb.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}
	baseBuilder.self = builder

	return builder
}

func (wb *windowBitOrExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	wb.funcExpr = wb.bitOrExpr

	return wb.baseWindowExpr.AppendQuery(fmter, b)
}

// windowBitAndExpr implements WindowBitAndBuilder.
type windowBitAndExpr struct {
	*bitAndExpr[WindowBitAndBuilder]
	*baseWindowExpr
}

func (wb *windowBitAndExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: wb.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}
	baseBuilder.self = builder

	return builder
}

func (wb *windowBitAndExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	wb.funcExpr = wb.bitAndExpr

	return wb.baseWindowExpr.AppendQuery(fmter, b)
}

// windowBoolOrExpr implements WindowBoolOrBuilder.
type windowBoolOrExpr struct {
	*boolOrExpr[WindowBoolOrBuilder]
	*baseWindowExpr
}

func (wb *windowBoolOrExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: wb.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}
	baseBuilder.self = builder

	return builder
}

func (wb *windowBoolOrExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	wb.funcExpr = wb.boolOrExpr

	return wb.baseWindowExpr.AppendQuery(fmter, b)
}

// windowBoolAndExpr implements WindowBoolAndBuilder.
type windowBoolAndExpr struct {
	*boolAndExpr[WindowBoolAndBuilder]
	*baseWindowExpr
}

func (wb *windowBoolAndExpr) Over() WindowFrameablePartitionBuilder {
	baseBuilder := &baseWindowPartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowExpr: wb.baseWindowExpr,
	}
	builder := &baseWindowFrameablePartitionBuilder[WindowFrameablePartitionBuilder]{
		baseWindowPartitionBuilder: baseBuilder,
	}
	baseBuilder.self = builder

	return builder
}

func (wb *windowBoolAndExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	wb.funcExpr = wb.boolAndExpr

	return wb.baseWindowExpr.AppendQuery(fmter, b)
}

// ========== Factory Functions ==========

func newRowNumberExpr(eb ExprBuilder) *rowNumberExpr {
	return &rowNumberExpr{
		baseWindowExpr: &baseWindowExpr{
			eb:       eb,
			funcName: "ROW_NUMBER",
		},
	}
}

func newRankExpr(eb ExprBuilder) *rankExpr {
	return &rankExpr{
		baseWindowExpr: &baseWindowExpr{
			eb:       eb,
			funcName: "RANK",
		},
	}
}

func newDenseRankExpr(eb ExprBuilder) *denseRankExpr {
	return &denseRankExpr{
		baseWindowExpr: &baseWindowExpr{
			eb:       eb,
			funcName: "DENSE_RANK",
		},
	}
}

func newPercentRankExpr(eb ExprBuilder) *percentRankExpr {
	return &percentRankExpr{
		baseWindowExpr: &baseWindowExpr{
			eb:       eb,
			funcName: "PERCENT_RANK",
		},
	}
}

func newCumeDistExpr(eb ExprBuilder) *cumeDistExpr {
	return &cumeDistExpr{
		baseWindowExpr: &baseWindowExpr{
			eb:       eb,
			funcName: "CUME_DIST",
		},
	}
}

func newNtileExpr(eb ExprBuilder) *ntileExpr {
	return &ntileExpr{
		baseWindowExpr: &baseWindowExpr{
			eb:       eb,
			funcName: "NTILE",
		},
	}
}

func newLagExpr(eb ExprBuilder) *lagExpr {
	return &lagExpr{
		baseWindowExpr: &baseWindowExpr{
			eb:       eb,
			funcName: "LAG",
		},
	}
}

func newLeadExpr(eb ExprBuilder) *leadExpr {
	return &leadExpr{
		baseWindowExpr: &baseWindowExpr{
			eb:       eb,
			funcName: "LEAD",
		},
		offset: 1,
	}
}

func newFirstValueExpr(eb ExprBuilder) *firstValueExpr {
	baseExpr := &baseWindowExpr{
		eb:       eb,
		funcName: "FIRST_VALUE",
	}
	baseBuilder := &baseWindowNullHandlingBuilder[FirstValueBuilder]{
		baseWindowExpr: baseExpr,
	}
	expr := &firstValueExpr{
		baseWindowExpr:                baseExpr,
		baseWindowNullHandlingBuilder: baseBuilder,
	}

	baseBuilder.self = expr

	return expr
}

func newLastValueExpr(eb ExprBuilder) *lastValueExpr {
	baseExpr := &baseWindowExpr{
		eb:       eb,
		funcName: "LAST_VALUE",
	}
	baseBuilder := &baseWindowNullHandlingBuilder[LastValueBuilder]{
		baseWindowExpr: baseExpr,
	}
	expr := &lastValueExpr{
		baseWindowExpr:                baseExpr,
		baseWindowNullHandlingBuilder: baseBuilder,
	}

	baseBuilder.self = expr

	return expr
}

func newNthValueExpr(eb ExprBuilder) *nthValueExpr {
	baseExpr := &baseWindowExpr{
		eb:       eb,
		funcName: "NTH_VALUE",
	}
	baseBuilder := &baseWindowNullHandlingBuilder[NthValueBuilder]{
		baseWindowExpr: baseExpr,
	}
	expr := &nthValueExpr{
		baseWindowExpr:                baseExpr,
		baseWindowNullHandlingBuilder: baseBuilder,
	}

	baseBuilder.self = expr

	return expr
}

func newWindowCountExpr(qb QueryBuilder) *windowCountExpr {
	expr := &windowCountExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.countExpr = newGenericCountExpr[WindowCountBuilder](expr, qb)

	return expr
}

func newWindowSumExpr(qb QueryBuilder) *windowSumExpr {
	expr := &windowSumExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.sumExpr = newGenericSumExpr[WindowSumBuilder](expr, qb)

	return expr
}

func newWindowAvgExpr(qb QueryBuilder) *windowAvgExpr {
	expr := &windowAvgExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.avgExpr = newGenericAvgExpr[WindowAvgBuilder](expr, qb)

	return expr
}

func newWindowMinExpr(qb QueryBuilder) *windowMinExpr {
	expr := &windowMinExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.minExpr = newGenericMinExpr[WindowMinBuilder](expr, qb)

	return expr
}

func newWindowMaxExpr(qb QueryBuilder) *windowMaxExpr {
	expr := &windowMaxExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.maxExpr = newGenericMaxExpr[WindowMaxBuilder](expr, qb)

	return expr
}

func newWindowStringAggExpr(qb QueryBuilder) *windowStringAggExpr {
	expr := &windowStringAggExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.stringAggExpr = newGenericStringAggExpr[WindowStringAggBuilder](expr, qb)

	return expr
}

func newWindowArrayAggExpr(qb QueryBuilder) *windowArrayAggExpr {
	expr := &windowArrayAggExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.arrayAggExpr = newGenericArrayAggExpr[WindowArrayAggBuilder](expr, qb)

	return expr
}

func newWindowStdDevExpr(qb QueryBuilder) *windowStdDevExpr {
	expr := &windowStdDevExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.stddevExpr = newGenericStdDevExpr[WindowStdDevBuilder](expr, qb)

	return expr
}

func newWindowVarianceExpr(qb QueryBuilder) *windowVarianceExpr {
	expr := &windowVarianceExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.varianceExpr = newGenericVarianceExpr[WindowVarianceBuilder](expr, qb)

	return expr
}

func newWindowJsonObjectAggExpr(qb QueryBuilder) *windowJsonObjectAggExpr {
	expr := &windowJsonObjectAggExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.jsonObjectAggExpr = newGenericJsonObjectAggExpr[WindowJSONObjectAggBuilder](expr, qb)

	return expr
}

func newWindowJsonArrayAggExpr(qb QueryBuilder) *windowJsonArrayAggExpr {
	expr := &windowJsonArrayAggExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.jsonArrayAggExpr = newGenericJsonArrayAggExpr[WindowJSONArrayAggBuilder](expr, qb)

	return expr
}

func newWindowBitOrExpr(qb QueryBuilder) *windowBitOrExpr {
	expr := &windowBitOrExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.bitOrExpr = newGenericBitOrExpr[WindowBitOrBuilder](expr, qb)

	return expr
}

func newWindowBitAndExpr(qb QueryBuilder) *windowBitAndExpr {
	expr := &windowBitAndExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.bitAndExpr = newGenericBitAndExpr[WindowBitAndBuilder](expr, qb)

	return expr
}

func newWindowBoolOrExpr(qb QueryBuilder) *windowBoolOrExpr {
	expr := &windowBoolOrExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.boolOrExpr = newGenericBoolOrExpr[WindowBoolOrBuilder](expr, qb)

	return expr
}

func newWindowBoolAndExpr(qb QueryBuilder) *windowBoolAndExpr {
	expr := &windowBoolAndExpr{
		baseWindowExpr: &baseWindowExpr{
			eb: qb.ExprBuilder(),
		},
	}

	expr.boolAndExpr = newGenericBoolAndExpr[WindowBoolAndBuilder](expr, qb)

	return expr
}
