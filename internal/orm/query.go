package orm

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/mo"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/uptrace/bun"
)

var (
	// errSubQuery is returned when attempting to execute a sub-query directly
	errSubQuery = errors.New("cannot execute a sub-query directly, sub-queries must be used as part of a parent query")
)

// NewQuery returns a new Query instance.
func NewQuery(db bun.IDB) orm.Query {
	return &bunQuery{
		query: db.NewSelect(),
	}
}

type bunQuery struct {
	query      *bun.SelectQuery
	isSubQuery bool
}

func (q *bunQuery) subQuery(subQuery *bun.SelectQuery) orm.Query {
	return &bunQuery{
		query:      subQuery,
		isSubQuery: true,
	}
}

func (q *bunQuery) buildSubQuery(builder func(query orm.Query)) *bun.SelectQuery {
	subQuery := q.query.NewSelect()
	builder(q.subQuery(subQuery))

	return subQuery
}

func (q *bunQuery) buildCondition(builder func(orm.ConditionBuilder)) orm.ConditionBuilder {
	cb := newCommonConditionBuilder(getTableSchemaFromQuery(q.query), q.buildSubQuery)
	builder(cb)
	return cb
}

func (q *bunQuery) With(name string, builder func(query orm.Query)) orm.Query {
	q.query.With(name, q.buildSubQuery(builder))
	return q
}

func (q *bunQuery) WithValues(name string, model any, withOrder ...bool) orm.Query {
	values := q.query.NewValues(model)
	if len(withOrder) > 0 && withOrder[0] {
		values.WithOrder()
	}

	q.query.With(name, values)
	return q
}

func (q *bunQuery) WithRecursive(name string, builder func(query orm.Query)) orm.Query {
	q.query.WithRecursive(name, q.buildSubQuery(builder))
	return q
}

func (q *bunQuery) SelectAll() orm.Query {
	q.query.Column(orm.ColumnAll)
	return q
}

func (q *bunQuery) Select(columns ...string) orm.Query {
	for _, column := range columns {
		q.query.ColumnExpr("?", parseColumnExpr(column))
	}
	return q
}

func (q *bunQuery) SelectAs(column string, alias string) orm.Query {
	q.query.ColumnExpr("? AS ?", parseColumnExpr(column), bun.Name(alias))
	return q
}

func (q *bunQuery) SelectExpr(expr string, args ...any) orm.Query {
	q.query.ColumnExpr(expr, args...)
	return q
}

func (q *bunQuery) SelectModelColumns() orm.Query {
	q.query.ColumnExpr(orm.ExprTableColumns)
	return q
}

func (q *bunQuery) SelectModelPKs() orm.Query {
	q.query.ColumnExpr(orm.ExprTablePKs)
	return q
}

func (q *bunQuery) SelectExprAs(expr string, alias string, args ...any) orm.Query {
	q.query.ColumnExpr("? AS ?", bun.SafeQuery(expr, args...), bun.Name(alias))
	return q
}

func (q *bunQuery) Exclude(columns ...string) orm.Query {
	q.query.ExcludeColumn(columns...)
	return q
}

func (q *bunQuery) ExcludeAll() orm.Query {
	q.query.ExcludeColumn(orm.ColumnAll)
	return q
}

func (q *bunQuery) Distinct() orm.Query {
	q.query.Distinct()
	return q
}

func (q *bunQuery) DistinctOn(expr string, args ...any) orm.Query {
	q.query.DistinctOn(expr, args...)
	return q
}

func (q *bunQuery) Model(model any) orm.Query {
	q.query.Model(model)
	return q
}

func (q *bunQuery) ModelTable(table string) orm.Query {
	q.query.ModelTableExpr("? AS ?TableAlias", bun.Name(table))
	return q
}

func (q *bunQuery) Table(name string) orm.Query {
	q.query.Table(name)
	return q
}

func (q *bunQuery) TableAs(name string, alias string) orm.Query {
	q.query.TableExpr("? AS ?", bun.Name(name), bun.Name(alias))
	return q
}

func (q *bunQuery) TableExpr(expr string, args ...any) orm.Query {
	q.query.TableExpr(expr, args...)
	return q
}

func (q *bunQuery) TableExprAs(expr string, alias string, args ...any) orm.Query {
	q.query.TableExpr("? AS ?", bun.SafeQuery(expr, args...), bun.Name(alias))
	return q
}

func (q *bunQuery) TableSubQuery(builder func(query orm.Query)) orm.Query {
	q.query.TableExpr("(?)", q.buildSubQuery(builder))
	return q
}

func (q *bunQuery) TableSubQueryAs(builder func(query orm.Query), alias string) orm.Query {
	q.query.TableExpr("(?) AS ?", q.buildSubQuery(builder), bun.Name(alias))
	return q
}

func (q *bunQuery) Join(model any, builder func(orm.ConditionBuilder)) orm.Query {
	table := getTableSchema(model, q.query.DB())
	q.query.Join("JOIN ? AS ?", bun.Name(table.Name), bun.Name(table.Alias))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) JoinAs(model any, alias string, builder func(orm.ConditionBuilder)) orm.Query {
	table := getTableSchema(model, q.query.DB())
	q.query.Join("JOIN ? AS ?", bun.Name(table.Name), bun.Name(alias))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) JoinTable(name string, builder func(orm.ConditionBuilder)) orm.Query {
	q.query.Join("JOIN ?", bun.Name(name))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) JoinTableAs(name string, alias string, builder func(orm.ConditionBuilder)) orm.Query {
	q.query.Join("JOIN ? AS ?", bun.Name(name), bun.Name(alias))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) JoinSubQuery(builder func(query orm.Query), conditionBuilder func(orm.ConditionBuilder)) orm.Query {
	q.query.Join("JOIN (?)", q.buildSubQuery(builder))
	q.query.JoinOn("?", q.buildCondition(conditionBuilder))
	return q
}

func (q *bunQuery) JoinSubQueryAs(builder func(query orm.Query), alias string, conditionBuilder func(orm.ConditionBuilder)) orm.Query {
	q.query.Join("JOIN (?) AS ?", q.buildSubQuery(builder), bun.Name(alias))
	q.query.JoinOn("?", q.buildCondition(conditionBuilder))
	return q
}

func (q *bunQuery) JoinExpr(expr string, builder func(cb orm.ConditionBuilder), args ...any) orm.Query {
	q.query.Join("JOIN ?", bun.SafeQuery(expr, args...))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) JoinExprAs(expr, alias string, builder func(cb orm.ConditionBuilder), args ...any) orm.Query {
	q.query.Join("JOIN ? AS ?", bun.SafeQuery(expr, args...), bun.Name(alias))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) LeftJoin(model any, builder func(orm.ConditionBuilder)) orm.Query {
	table := getTableSchema(model, q.query.DB())
	q.query.Join("LEFT JOIN ? AS ?", bun.Name(table.Name), bun.Name(table.Alias))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) LeftJoinAs(model any, alias string, builder func(orm.ConditionBuilder)) orm.Query {
	table := getTableSchema(model, q.query.DB())
	q.query.Join("LEFT JOIN ? AS ?", bun.Name(table.Name), bun.Name(alias))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) LeftJoinTable(name string, builder func(orm.ConditionBuilder)) orm.Query {
	q.query.Join("LEFT JOIN ?", bun.Name(name))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) LeftJoinTableAs(name string, alias string, builder func(orm.ConditionBuilder)) orm.Query {
	q.query.Join("LEFT JOIN ? AS ?", bun.Name(name), bun.Name(alias))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) LeftJoinSubQuery(builder func(query orm.Query), conditionBuilder func(orm.ConditionBuilder)) orm.Query {
	q.query.Join("LEFT JOIN (?)", q.buildSubQuery(builder))
	q.query.JoinOn("?", q.buildCondition(conditionBuilder))
	return q
}

func (q *bunQuery) LeftJoinSubQueryAs(builder func(query orm.Query), alias string, conditionBuilder func(orm.ConditionBuilder)) orm.Query {
	q.query.Join("LEFT JOIN (?) AS ?", q.buildSubQuery(builder), bun.Name(alias))
	q.query.JoinOn("?", q.buildCondition(conditionBuilder))
	return q
}

func (q *bunQuery) LeftJoinExpr(expr string, builder func(cb orm.ConditionBuilder), args ...any) orm.Query {
	q.query.Join("LEFT JOIN ?", bun.SafeQuery(expr, args...))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) LeftJoinExprAs(expr, alias string, builder func(cb orm.ConditionBuilder), args ...any) orm.Query {
	q.query.Join("LEFT JOIN ? AS ?", bun.SafeQuery(expr, args...), bun.Name(alias))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) RightJoin(model any, builder func(orm.ConditionBuilder)) orm.Query {
	table := getTableSchema(model, q.query.DB())
	q.query.Join("RIGHT JOIN ? AS ?", bun.Name(table.Name), bun.Name(table.Alias))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) RightJoinAs(model any, alias string, builder func(orm.ConditionBuilder)) orm.Query {
	table := getTableSchema(model, q.query.DB())
	q.query.Join("RIGHT JOIN ? AS ?", bun.Name(table.Name), bun.Name(alias))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) RightJoinTable(name string, builder func(orm.ConditionBuilder)) orm.Query {
	q.query.Join("RIGHT JOIN ?", bun.Name(name))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) RightJoinTableAs(name string, alias string, builder func(orm.ConditionBuilder)) orm.Query {
	q.query.Join("RIGHT JOIN ? AS ?", bun.Name(name), bun.Name(alias))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) RightJoinSubQuery(builder func(query orm.Query), conditionBuilder func(orm.ConditionBuilder)) orm.Query {
	q.query.Join("RIGHT JOIN (?)", q.buildSubQuery(builder))
	q.query.JoinOn("?", q.buildCondition(conditionBuilder))
	return q
}

func (q *bunQuery) RightJoinSubQueryAs(builder func(query orm.Query), alias string, conditionBuilder func(orm.ConditionBuilder)) orm.Query {
	q.query.Join("RIGHT JOIN (?) AS ?", q.buildSubQuery(builder), bun.Name(alias))
	q.query.JoinOn("?", q.buildCondition(conditionBuilder))
	return q
}

func (q *bunQuery) RightJoinExpr(expr string, builder func(cb orm.ConditionBuilder), args ...any) orm.Query {
	q.query.Join("RIGHT JOIN ?", bun.SafeQuery(expr, args...))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) RightJoinExprAs(expr, alias string, builder func(cb orm.ConditionBuilder), args ...any) orm.Query {
	q.query.Join("RIGHT JOIN ? AS ?", bun.SafeQuery(expr, args...), bun.Name(alias))
	q.query.JoinOn("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) ModelRelation(relation ...orm.ModelRelation) orm.Query {
	for _, r := range relation {
		applyModelRelation(r, q)
	}

	return q
}

func (q *bunQuery) Relation(name string, apply ...func(query orm.Query)) orm.Query {
	if len(apply) == 0 {
		q.query.Relation(name)
	} else {
		q.query.Relation(name, func(query *bun.SelectQuery) *bun.SelectQuery {
			subQuery := q.subQuery(query)
			for _, apply := range apply {
				apply(subQuery)
			}

			return query
		})
	}

	return q
}

func (q *bunQuery) Where(builder func(orm.ConditionBuilder)) orm.Query {
	cb := newQueryConditionBuilder(getTableSchemaFromQuery(q.query), q.query.QueryBuilder(), q.buildSubQuery)
	builder(cb)
	return q
}

func (q *bunQuery) WherePK(columns ...string) orm.Query {
	q.query.WherePK(columns...)
	return q
}

func (q *bunQuery) WhereDeleted() orm.Query {
	q.query.WhereDeleted()
	return q
}

func (q *bunQuery) WhereAllWithDeleted() orm.Query {
	q.query.WhereAllWithDeleted()
	return q
}

func (q *bunQuery) GroupBy(columns ...string) orm.Query {
	for _, column := range columns {
		q.query.GroupExpr("?", parseColumnExpr(column))
	}
	return q
}

func (q *bunQuery) GroupByExpr(expr string, args ...any) orm.Query {
	q.query.GroupExpr(expr, args...)
	return q
}

func (q *bunQuery) Having(builder func(orm.ConditionBuilder)) orm.Query {
	q.query.Having("?", q.buildCondition(builder))
	return q
}

func (q *bunQuery) OrderBy(columns ...string) orm.Query {
	for _, column := range columns {
		q.query.OrderExpr("?", parseColumnExpr(column))
	}
	return q
}

func (q *bunQuery) OrderByNullsFirst(columns ...string) orm.Query {
	for _, column := range columns {
		q.query.OrderExpr("? NULLS FIRST", parseColumnExpr(column))
	}

	return q
}

func (q *bunQuery) OrderByNullsLast(columns ...string) orm.Query {
	for _, column := range columns {
		q.query.OrderExpr("? NULLS LAST", parseColumnExpr(column))
	}

	return q
}

func (q *bunQuery) OrderByDesc(columns ...string) orm.Query {
	for _, column := range columns {
		q.query.OrderExpr("? DESC", parseColumnExpr(column))
	}

	return q
}

func (q *bunQuery) OrderByDescNullsFirst(columns ...string) orm.Query {
	for _, column := range columns {
		q.query.OrderExpr("? DESC NULLS FIRST", parseColumnExpr(column))
	}

	return q
}

func (q *bunQuery) OrderByDescNullsLast(columns ...string) orm.Query {
	for _, column := range columns {
		q.query.OrderExpr("? DESC NULLS LAST", parseColumnExpr(column))
	}

	return q
}

func (q *bunQuery) OrderByExpr(expr string, args ...any) orm.Query {
	q.query.OrderExpr(expr, args...)
	return q
}

func (q *bunQuery) Limit(limit int) orm.Query {
	q.query.Limit(limit)
	return q
}

func (q *bunQuery) Offset(offset int) orm.Query {
	q.query.Offset(offset)
	return q
}

func (q *bunQuery) Paginate(pageable mo.Pageable, defaultAlias ...string) orm.Query {
	if pageable.Sort != constants.Empty {
		NewSorter(pageable.Sort).Apply(q, defaultAlias...)
	}

	p := pageable.Page
	if p < 1 {
		p = mo.DefaultPageNumber
	}
	size := min(pageable.Size, mo.MaxPageSize)
	if size < 1 {
		size = mo.DefaultPageSize
	}

	offset := (p - 1) * size
	return q.Offset(offset).Limit(size)
}

func (q *bunQuery) ForShare(tables ...string) orm.Query {
	if len(tables) == 0 {
		q.query.For("SHARE")
	} else {
		q.query.For("SHARE OF ?", Names(tables...))
	}
	return q
}

func (q *bunQuery) ForShareNoWait(tables ...string) orm.Query {
	if len(tables) == 0 {
		q.query.For("SHARE NO WAIT")
	} else {
		q.query.For("SHARE OF ? NO WAIT", Names(tables...))
	}
	return q
}

func (q *bunQuery) ForShareSkipLocked(tables ...string) orm.Query {
	if len(tables) == 0 {
		q.query.For("SHARE SKIP LOCKED")
	} else {
		q.query.For("SHARE OF ? SKIP LOCKED", Names(tables...))
	}
	return q
}

func (q *bunQuery) ForUpdate(tables ...string) orm.Query {
	if len(tables) == 0 {
		q.query.For("UPDATE")
	} else {
		q.query.For("UPDATE OF ?", Names(tables...))
	}
	return q
}

func (q *bunQuery) ForUpdateNoWait(tables ...string) orm.Query {
	if len(tables) == 0 {
		q.query.For("UPDATE NO WAIT")
	} else {
		q.query.For("UPDATE OF ? NO WAIT", Names(tables...))
	}
	return q
}

func (q *bunQuery) ForUpdateSkipLocked(tables ...string) orm.Query {
	if len(tables) == 0 {
		q.query.For("UPDATE SKIP LOCKED")
	} else {
		q.query.For("UPDATE OF ? SKIP LOCKED", Names(tables...))
	}
	return q
}

func (q *bunQuery) Union(builder func(query orm.Query)) orm.Query {
	q.query.Union(q.buildSubQuery(builder))
	return q
}

func (q *bunQuery) UnionAll(builder func(query orm.Query)) orm.Query {
	q.query.UnionAll(q.buildSubQuery(builder))
	return q
}

func (q *bunQuery) Intersect(builder func(query orm.Query)) orm.Query {
	q.query.Intersect(q.buildSubQuery(builder))
	return q
}

func (q *bunQuery) IntersectAll(builder func(query orm.Query)) orm.Query {
	q.query.IntersectAll(q.buildSubQuery(builder))
	return q
}

func (q *bunQuery) Except(builder func(query orm.Query)) orm.Query {
	q.query.Except(q.buildSubQuery(builder))
	return q
}

func (q *bunQuery) ExceptAll(builder func(query orm.Query)) orm.Query {
	q.query.ExceptAll(q.buildSubQuery(builder))
	return q
}

func (q *bunQuery) Apply(fns ...orm.ApplyFunc[orm.Query]) orm.Query {
	for _, fn := range fns {
		if fn != nil {
			fn(q)
		}
	}

	return q
}

func (q *bunQuery) ApplyIf(condition bool, fns ...orm.ApplyFunc[orm.Query]) orm.Query {
	if condition {
		return q.Apply(fns...)
	}
	return q
}

func (q *bunQuery) Exec(ctx context.Context, dest ...any) (sql.Result, error) {
	if q.isSubQuery {
		return nil, errSubQuery
	}

	r, err := q.query.Exec(ctx, dest...)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, result.ErrRecordNotFound
	}

	return r, err
}

func (q *bunQuery) Scan(ctx context.Context, dest ...any) error {
	if q.isSubQuery {
		return errSubQuery
	}

	if err := q.query.Scan(ctx, dest...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result.ErrRecordNotFound
		}

		return err
	}

	return nil
}

func (q *bunQuery) Rows(ctx context.Context) (*sql.Rows, error) {
	if q.isSubQuery {
		return nil, errSubQuery
	}

	rows, err := q.query.Rows(ctx)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, result.ErrRecordNotFound
	}

	return rows, err
}

func (q *bunQuery) ScanAndCount(ctx context.Context, dest ...any) (int64, error) {
	if q.isSubQuery {
		return 0, errSubQuery
	}

	total, err := q.query.ScanAndCount(ctx, dest...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, result.ErrRecordNotFound
		}

		return 0, err
	}

	return int64(total), nil
}

func (q *bunQuery) Count(ctx context.Context) (int64, error) {
	if q.isSubQuery {
		return 0, errSubQuery
	}

	total, err := q.query.Count(ctx)
	return int64(total), err
}

func (q *bunQuery) Exists(ctx context.Context) (bool, error) {
	if q.isSubQuery {
		return false, errSubQuery
	}

	return q.query.Exists(ctx)
}
