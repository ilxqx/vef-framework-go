package orm

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/page"
	"github.com/ilxqx/vef-framework-go/result"
)

// NewSelectQuery creates a new SelectQuery instance with the provided database instance.
// It initializes the query builders and sets up the table schema context for proper query building.
func NewSelectQuery(db *BunDB) *BunSelectQuery {
	eb := &QueryExprBuilder{}
	sq := db.db.NewSelect()
	dialect := db.db.Dialect()
	query := &BunSelectQuery{
		QueryBuilder: newQueryBuilder(db, dialect, sq, eb),

		db:      db,
		dialect: dialect,
		eb:      eb,
		query:   sq,
	}
	eb.qb = query

	return query
}

// BunSelectQuery is the concrete implementation of SelectQuery interface.
// It wraps bun.SelectQuery and provides additional functionality for expression building.
type BunSelectQuery struct {
	QueryBuilder

	db         *BunDB
	dialect    schema.Dialect
	eb         ExprBuilder
	query      *bun.SelectQuery
	isSubQuery bool

	// State tracking for deferred select operations
	hasSelectAll          bool
	hasSelectModelColumns bool
	hasSelectModelPKs     bool
	hasExplicitSelect     bool
	explicitSelects       []func()
	exprSelects           []func()
	selectStateApplied    bool
}

func (q *BunSelectQuery) DB() DB {
	return q.db
}

func (q *BunSelectQuery) With(name string, builder func(query SelectQuery)) SelectQuery {
	q.query.With(name, q.BuildSubQuery(builder))

	return q
}

func (q *BunSelectQuery) WithValues(name string, model any, withOrder ...bool) SelectQuery {
	values := q.query.NewValues(model)
	if len(withOrder) > 0 && withOrder[0] {
		values.WithOrder()
	}

	q.query.With(name, values)

	return q
}

func (q *BunSelectQuery) WithRecursive(name string, builder func(query SelectQuery)) SelectQuery {
	q.query.WithRecursive(name, q.BuildSubQuery(builder))

	return q
}

func (q *BunSelectQuery) SelectAll() SelectQuery {
	q.clearSelectState()
	q.hasSelectAll = true

	return q
}

func (q *BunSelectQuery) Select(columns ...string) SelectQuery {
	if q.hasSelectAll {
		q.hasSelectAll = false
	}

	q.hasExplicitSelect = true

	for _, column := range columns {
		q.explicitSelects = append(q.explicitSelects, func() {
			q.query.ColumnExpr("?", q.eb.Column(column))
		})
	}

	return q
}

func (q *BunSelectQuery) SelectAs(column, alias string) SelectQuery {
	if q.hasSelectAll {
		q.hasSelectAll = false
	}

	q.hasExplicitSelect = true

	q.explicitSelects = append(q.explicitSelects, func() {
		q.query.ColumnExpr("? AS ?", q.eb.Column(column), bun.Name(alias))
	})

	return q
}

func (q *BunSelectQuery) SelectExpr(builder func(ExprBuilder) any, alias ...string) SelectQuery {
	var (
		expr       = builder(q.eb)
		aliasToUse string
	)
	if len(alias) > 0 && alias[0] != constants.Empty {
		aliasToUse = alias[0]
	}

	q.exprSelects = append(q.exprSelects, func() {
		if aliasToUse != constants.Empty {
			q.query.ColumnExpr("? AS ?", expr, bun.Name(aliasToUse))
		} else {
			q.query.ColumnExpr("?", expr)
		}
	})

	return q
}

func (q *BunSelectQuery) SelectModelColumns() SelectQuery {
	if q.hasSelectAll {
		q.hasSelectAll = false
	}

	if q.hasSelectModelPKs {
		q.hasSelectModelPKs = false
	}

	q.hasSelectModelColumns = true

	return q
}

func (q *BunSelectQuery) SelectModelPKs() SelectQuery {
	if q.hasSelectAll {
		q.hasSelectAll = false
	}

	if q.hasSelectModelColumns {
		q.hasSelectModelColumns = false
	}

	q.hasSelectModelPKs = true

	return q
}

func (q *BunSelectQuery) Exclude(columns ...string) SelectQuery {
	q.query.ExcludeColumn(columns...)

	return q
}

func (q *BunSelectQuery) ExcludeAll() SelectQuery {
	q.query.ExcludeColumn(columnAll)

	return q
}

func (q *BunSelectQuery) Distinct() SelectQuery {
	q.query.Distinct()

	return q
}

func (q *BunSelectQuery) DistinctOnColumns(columns ...string) SelectQuery {
	for _, column := range columns {
		q.query.DistinctOn("?", q.eb.Column(column))
	}

	return q
}

func (q *BunSelectQuery) DistinctOnExpr(builder func(ExprBuilder) any) SelectQuery {
	expr := builder(q.eb)
	q.query.DistinctOn("?", expr)

	return q
}

func (q *BunSelectQuery) Model(model any) SelectQuery {
	q.query.Model(model)

	return q
}

func (q *BunSelectQuery) ModelTable(name string, alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.ModelTableExpr("? AS ?", bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.ModelTableExpr("? AS ?TableAlias", bun.Name(name))
	}

	return q
}

func (q *BunSelectQuery) Table(name string, alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.TableExpr("? AS ?", bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.Table(name)
	}

	return q
}

func (q *BunSelectQuery) TableFrom(model any, alias ...string) SelectQuery {
	table := q.db.TableOf(model)

	aliasToUse := table.Alias
	if len(alias) > 0 && alias[0] != constants.Empty {
		aliasToUse = alias[0]
	}

	q.query.TableExpr("? AS ?", bun.Name(table.Name), bun.Name(aliasToUse))

	return q
}

func (q *BunSelectQuery) TableExpr(builder func(ExprBuilder) any, alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.TableExpr("? AS ?", builder(q.eb), bun.Name(alias[0]))
	} else {
		q.query.TableExpr("?", builder(q.eb))
	}

	return q
}

func (q *BunSelectQuery) TableSubQuery(builder func(query SelectQuery), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.TableExpr("(?) AS ?", q.BuildSubQuery(builder), bun.Name(alias[0]))
	} else {
		q.query.TableExpr("(?)", q.BuildSubQuery(builder))
	}

	return q
}

func (q *BunSelectQuery) Join(model any, builder func(ConditionBuilder), alias ...string) SelectQuery {
	table := q.db.TableOf(model)

	aliasToUse := table.Alias
	if len(alias) > 0 && alias[0] != constants.Empty {
		aliasToUse = alias[0]
	}

	q.query.Join(
		"? ? AS ?",
		bun.Safe(JoinInner.String()),
		bun.Name(table.Name),
		bun.Name(aliasToUse),
	)
	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) JoinTable(name string, builder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? ? AS ?", bun.Safe(JoinInner.String()), bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.Join("? ?", bun.Safe(JoinInner.String()), bun.Name(name))
	}

	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) JoinSubQuery(sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? (?) AS ?", bun.Safe(JoinInner.String()), q.BuildSubQuery(sqBuilder), bun.Name(alias[0]))
	} else {
		q.query.Join("? (?)", bun.Safe(JoinInner.String()), q.BuildSubQuery(sqBuilder))
	}

	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunSelectQuery) JoinExpr(eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? (?) AS ?", bun.Safe(JoinInner.String()), eBuilder(q.eb), bun.Name(alias[0]))
	} else {
		q.query.Join("? (?)", bun.Safe(JoinInner.String()), eBuilder(q.eb))
	}

	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunSelectQuery) LeftJoin(model any, builder func(ConditionBuilder), alias ...string) SelectQuery {
	table := q.db.TableOf(model)

	aliasToUse := table.Alias
	if len(alias) > 0 && alias[0] != constants.Empty {
		aliasToUse = alias[0]
	}

	q.query.Join(
		"? ? AS ?",
		bun.Safe(JoinLeft.String()),
		bun.Name(table.Name),
		bun.Name(aliasToUse),
	)
	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) LeftJoinTable(name string, builder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? ? AS ?", bun.Safe(JoinLeft.String()), bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.Join("? ?", bun.Safe(JoinLeft.String()), bun.Name(name))
	}

	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) LeftJoinSubQuery(sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? (?) AS ?", bun.Safe(JoinLeft.String()), q.BuildSubQuery(sqBuilder), bun.Name(alias[0]))
	} else {
		q.query.Join("? (?)", bun.Safe(JoinLeft.String()), q.BuildSubQuery(sqBuilder))
	}

	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunSelectQuery) LeftJoinExpr(eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? (?) AS ?", bun.Safe(JoinLeft.String()), eBuilder(q.eb), bun.Name(alias[0]))
	} else {
		q.query.Join("? (?)", bun.Safe(JoinLeft.String()), eBuilder(q.eb))
	}

	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunSelectQuery) RightJoin(model any, builder func(ConditionBuilder), alias ...string) SelectQuery {
	table := q.db.TableOf(model)

	aliasToUse := table.Alias
	if len(alias) > 0 && alias[0] != constants.Empty {
		aliasToUse = alias[0]
	}

	q.query.Join(
		"? ? AS ?",
		bun.Safe(JoinRight.String()),
		bun.Name(table.Name),
		bun.Name(aliasToUse),
	)
	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) RightJoinTable(name string, builder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? ? AS ?", bun.Safe(JoinRight.String()), bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.Join("? ?", bun.Safe(JoinRight.String()), bun.Name(name))
	}

	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) RightJoinSubQuery(sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? (?) AS ?", bun.Safe(JoinRight.String()), q.BuildSubQuery(sqBuilder), bun.Name(alias[0]))
	} else {
		q.query.Join("? (?)", bun.Safe(JoinRight.String()), q.BuildSubQuery(sqBuilder))
	}

	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunSelectQuery) RightJoinExpr(eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? (?) AS ?", bun.Safe(JoinRight.String()), eBuilder(q.eb), bun.Name(alias[0]))
	} else {
		q.query.Join("? (?)", bun.Safe(JoinRight.String()), eBuilder(q.eb))
	}

	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunSelectQuery) FullJoin(model any, builder func(ConditionBuilder), alias ...string) SelectQuery {
	table := q.db.TableOf(model)

	aliasToUse := table.Alias
	if len(alias) > 0 && alias[0] != constants.Empty {
		aliasToUse = alias[0]
	}

	q.query.Join(
		"? ? AS ?",
		bun.Safe(JoinFull.String()),
		bun.Name(table.Name),
		bun.Name(aliasToUse),
	)
	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) FullJoinTable(name string, builder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? ? AS ?", bun.Safe(JoinFull.String()), bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.Join("? ?", bun.Safe(JoinFull.String()), bun.Name(name))
	}

	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) FullJoinSubQuery(sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? (?) AS ?", bun.Safe(JoinFull.String()), q.BuildSubQuery(sqBuilder), bun.Name(alias[0]))
	} else {
		q.query.Join("? (?)", bun.Safe(JoinFull.String()), q.BuildSubQuery(sqBuilder))
	}

	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunSelectQuery) FullJoinExpr(eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? (?) AS ?", bun.Safe(JoinFull.String()), eBuilder(q.eb), bun.Name(alias[0]))
	} else {
		q.query.Join("? (?)", bun.Safe(JoinFull.String()), eBuilder(q.eb))
	}

	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunSelectQuery) CrossJoin(model any, alias ...string) SelectQuery {
	table := q.db.TableOf(model)

	aliasToUse := table.Alias
	if len(alias) > 0 && alias[0] != constants.Empty {
		aliasToUse = alias[0]
	}

	q.query.Join(
		"? ? AS ?",
		bun.Safe(JoinCross.String()),
		bun.Name(table.Name),
		bun.Name(aliasToUse),
	)

	return q
}

func (q *BunSelectQuery) CrossJoinTable(name string, alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? ? AS ?", bun.Safe(JoinCross.String()), bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.Join("? ?", bun.Safe(JoinCross.String()), bun.Name(name))
	}

	return q
}

func (q *BunSelectQuery) CrossJoinSubQuery(sqBuilder func(query SelectQuery), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? (?) AS ?", bun.Safe(JoinCross.String()), q.BuildSubQuery(sqBuilder), bun.Name(alias[0]))
	} else {
		q.query.Join("? (?)", bun.Safe(JoinCross.String()), q.BuildSubQuery(sqBuilder))
	}

	return q
}

func (q *BunSelectQuery) CrossJoinExpr(eBuilder func(ExprBuilder) any, alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? (?) AS ?", bun.Safe(JoinCross.String()), eBuilder(q.eb), bun.Name(alias[0]))
	} else {
		q.query.Join("? (?)", bun.Safe(JoinCross.String()), eBuilder(q.eb))
	}

	return q
}

func (q *BunSelectQuery) JoinRelations(specs ...*RelationSpec) SelectQuery {
	for _, spec := range specs {
		applyRelationSpec(spec, q)
	}

	return q
}

func (q *BunSelectQuery) Relation(name string, apply ...func(query SelectQuery)) SelectQuery {
	if len(apply) == 0 {
		q.query.Relation(name)
	} else {
		q.query.Relation(name, func(query *bun.SelectQuery) *bun.SelectQuery {
			subQuery := q.CreateSubQuery(query)
			for _, apply := range apply {
				apply(subQuery)
			}

			return query
		})
	}

	return q
}

func (q *BunSelectQuery) Where(builder func(ConditionBuilder)) SelectQuery {
	cb := newQueryConditionBuilder(q.query.QueryBuilder(), q)
	builder(cb)

	return q
}

func (q *BunSelectQuery) WherePK(columns ...string) SelectQuery {
	q.query.WherePK(columns...)

	return q
}

func (q *BunSelectQuery) WhereDeleted() SelectQuery {
	q.query.WhereDeleted()

	return q
}

func (q *BunSelectQuery) IncludeDeleted() SelectQuery {
	q.query.WhereAllWithDeleted()

	return q
}

func (q *BunSelectQuery) GroupBy(columns ...string) SelectQuery {
	for _, column := range columns {
		q.query.GroupExpr("?", q.eb.Column(column))
	}

	return q
}

func (q *BunSelectQuery) GroupByExpr(builder func(ExprBuilder) any) SelectQuery {
	expr := builder(q.eb)
	q.query.GroupExpr("?", expr)

	return q
}

func (q *BunSelectQuery) Having(builder func(ConditionBuilder)) SelectQuery {
	q.query.Having("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) OrderBy(columns ...string) SelectQuery {
	for _, column := range columns {
		q.query.OrderExpr("? ASC", q.eb.Column(column))
	}

	return q
}

func (q *BunSelectQuery) OrderByDesc(columns ...string) SelectQuery {
	for _, column := range columns {
		q.query.OrderExpr("? DESC", q.eb.Column(column))
	}

	return q
}

func (q *BunSelectQuery) OrderByExpr(builder func(ExprBuilder) any) SelectQuery {
	expr := builder(q.eb)
	q.query.OrderExpr("?", expr)

	return q
}

func (q *BunSelectQuery) Limit(limit int) SelectQuery {
	q.query.Limit(limit)

	return q
}

func (q *BunSelectQuery) Offset(offset int) SelectQuery {
	q.query.Offset(offset)

	return q
}

func (q *BunSelectQuery) Paginate(pageable page.Pageable) SelectQuery {
	pageable.Normalize()

	return q.Offset(pageable.Offset()).Limit(pageable.Size)
}

func (q *BunSelectQuery) ForShare(tables ...string) SelectQuery {
	if len(tables) == 0 {
		q.query.For("SHARE")
	} else {
		q.query.For("SHARE OF ?", Names(tables...))
	}

	return q
}

func (q *BunSelectQuery) ForShareNoWait(tables ...string) SelectQuery {
	if len(tables) == 0 {
		q.query.For("SHARE NO WAIT")
	} else {
		q.query.For("SHARE OF ? NO WAIT", Names(tables...))
	}

	return q
}

func (q *BunSelectQuery) ForShareSkipLocked(tables ...string) SelectQuery {
	if len(tables) == 0 {
		q.query.For("SHARE SKIP LOCKED")
	} else {
		q.query.For("SHARE OF ? SKIP LOCKED", Names(tables...))
	}

	return q
}

func (q *BunSelectQuery) ForUpdate(tables ...string) SelectQuery {
	if len(tables) == 0 {
		q.query.For("UPDATE")
	} else {
		q.query.For("UPDATE OF ?", Names(tables...))
	}

	return q
}

func (q *BunSelectQuery) ForUpdateNoWait(tables ...string) SelectQuery {
	if len(tables) == 0 {
		q.query.For("UPDATE NOWAIT")
	} else {
		q.query.For("UPDATE OF ? NOWAIT", Names(tables...))
	}

	return q
}

func (q *BunSelectQuery) ForUpdateSkipLocked(tables ...string) SelectQuery {
	if len(tables) == 0 {
		q.query.For("UPDATE SKIP LOCKED")
	} else {
		q.query.For("UPDATE OF ? SKIP LOCKED", Names(tables...))
	}

	return q
}

func (q *BunSelectQuery) Union(builder func(query SelectQuery)) SelectQuery {
	q.query.Union(q.BuildSubQuery(builder))

	return q
}

func (q *BunSelectQuery) UnionAll(builder func(query SelectQuery)) SelectQuery {
	q.query.UnionAll(q.BuildSubQuery(builder))

	return q
}

func (q *BunSelectQuery) Intersect(builder func(query SelectQuery)) SelectQuery {
	q.query.Intersect(q.BuildSubQuery(builder))

	return q
}

func (q *BunSelectQuery) IntersectAll(builder func(query SelectQuery)) SelectQuery {
	q.query.IntersectAll(q.BuildSubQuery(builder))

	return q
}

func (q *BunSelectQuery) Except(builder func(query SelectQuery)) SelectQuery {
	q.query.Except(q.BuildSubQuery(builder))

	return q
}

func (q *BunSelectQuery) ExceptAll(builder func(query SelectQuery)) SelectQuery {
	q.query.ExceptAll(q.BuildSubQuery(builder))

	return q
}

func (q *BunSelectQuery) Apply(fns ...ApplyFunc[SelectQuery]) SelectQuery {
	for _, fn := range fns {
		if fn != nil {
			fn(q)
		}
	}

	return q
}

func (q *BunSelectQuery) ApplyIf(condition bool, fns ...ApplyFunc[SelectQuery]) SelectQuery {
	if condition {
		return q.Apply(fns...)
	}

	return q
}

// clearSelectState clears base column selection state flags.
// Note: exprSelects are NOT cleared as SelectExpr is cumulative and can combine with any mode.
func (q *BunSelectQuery) clearSelectState() {
	q.hasSelectAll = false
	q.hasSelectModelColumns = false
	q.hasSelectModelPKs = false
	q.hasExplicitSelect = false
	q.explicitSelects = nil
}

// applySelectState applies deferred select state before query execution.
func (q *BunSelectQuery) applySelectState() {
	if q.selectStateApplied {
		return
	}

	if q.hasSelectAll {
		q.query.Column(columnAll)
	} else {
		if q.hasSelectModelColumns {
			q.query.ColumnExpr(constants.ExprTableColumns)
		} else if q.hasSelectModelPKs {
			q.query.ColumnExpr(constants.ExprTablePKs)
		}

		if q.hasExplicitSelect {
			for _, selectFn := range q.explicitSelects {
				selectFn()
			}
		}
	}

	for _, exprFn := range q.exprSelects {
		exprFn()
	}

	q.selectStateApplied = true
}

func (q *BunSelectQuery) Exec(ctx context.Context, dest ...any) (res sql.Result, err error) {
	if q.isSubQuery {
		return nil, ErrSubQuery
	}

	q.applySelectState()

	if res, err = q.query.Exec(ctx, dest...); err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, result.ErrRecordNotFound
	}

	return res, err
}

func (q *BunSelectQuery) Scan(ctx context.Context, dest ...any) (err error) {
	if q.isSubQuery {
		return ErrSubQuery
	}

	q.applySelectState()

	if err = q.query.Scan(ctx, dest...); err != nil && errors.Is(err, sql.ErrNoRows) {
		return result.ErrRecordNotFound
	}

	return err
}

func (q *BunSelectQuery) Rows(ctx context.Context) (rows *sql.Rows, err error) {
	if q.isSubQuery {
		return nil, ErrSubQuery
	}

	q.applySelectState()

	if rows, err = q.query.Rows(ctx); err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, result.ErrRecordNotFound
	}

	return rows, err
}

func (q *BunSelectQuery) ScanAndCount(ctx context.Context, dest ...any) (int64, error) {
	if q.isSubQuery {
		return 0, ErrSubQuery
	}

	q.applySelectState()

	total, err := q.query.ScanAndCount(ctx, dest...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, result.ErrRecordNotFound
		}

		return 0, err
	}

	return int64(total), nil
}

func (q *BunSelectQuery) Count(ctx context.Context) (int64, error) {
	if q.isSubQuery {
		return 0, ErrSubQuery
	}

	q.applySelectState()

	total, err := q.query.Count(ctx)

	return int64(total), err
}

func (q *BunSelectQuery) Exists(ctx context.Context) (bool, error) {
	if q.isSubQuery {
		return false, ErrSubQuery
	}

	q.applySelectState()

	return q.query.Exists(ctx)
}

func (q *BunSelectQuery) Unwrap() *bun.SelectQuery {
	return q.query
}
