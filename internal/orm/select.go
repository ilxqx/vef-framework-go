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

// NewSelectQuery creates a new SelectQuery instance with the provided database connection.
// It initializes the query builders and sets up the table schema context for proper query building.
func NewSelectQuery(db bun.IDB) *BunSelectQuery {
	eb := &QueryExprBuilder{}
	sq := db.NewSelect()
	dialect := db.Dialect()
	query := &BunSelectQuery{
		QueryBuilder: newQueryBuilder(dialect, sq, eb),

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

	dialect    schema.Dialect
	eb         ExprBuilder
	query      *bun.SelectQuery
	isSubQuery bool
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
	q.query.Column(columnAll)

	return q
}

func (q *BunSelectQuery) Select(columns ...string) SelectQuery {
	for _, column := range columns {
		q.query.ColumnExpr("?", q.eb.Column(column))
	}

	return q
}

func (q *BunSelectQuery) SelectAs(column, alias string) SelectQuery {
	q.query.ColumnExpr("? AS ?", q.eb.Column(column), bun.Name(alias))

	return q
}

func (q *BunSelectQuery) SelectExpr(builder func(ExprBuilder) any, alias ...string) SelectQuery {
	expr := builder(q.eb)
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.ColumnExpr("? AS ?", expr, bun.Name(alias[0]))
	} else {
		q.query.ColumnExpr("?", expr)
	}

	return q
}

func (q *BunSelectQuery) SelectModelColumns() SelectQuery {
	q.query.ColumnExpr(constants.ExprTableColumns)

	return q
}

func (q *BunSelectQuery) SelectModelPKs() SelectQuery {
	q.query.ColumnExpr(constants.ExprTablePKs)

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

func (q *BunSelectQuery) TableExpr(alias string, builder func(ExprBuilder) any) SelectQuery {
	q.query.TableExpr("? AS ?", builder(q.eb), bun.Name(alias))

	return q
}

func (q *BunSelectQuery) TableSubQuery(alias string, builder func(query SelectQuery)) SelectQuery {
	q.query.TableExpr("(?) AS ?", q.BuildSubQuery(builder), bun.Name(alias))

	return q
}

func (q *BunSelectQuery) Join(model any, builder func(ConditionBuilder), alias ...string) SelectQuery {
	table := getTableSchema(model, q.query.DB())

	aliasToUse := table.Alias
	if len(alias) > 0 && alias[0] != constants.Empty {
		aliasToUse = alias[0]
	}

	q.query.Join(
		"JOIN ? AS ?",
		bun.Name(table.Name),
		bun.Name(aliasToUse),
	)
	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) JoinTable(name string, builder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("JOIN ? AS ?", bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.Join("JOIN ?", bun.Name(name))
	}

	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) JoinSubQuery(alias string, sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder)) SelectQuery {
	q.query.Join("JOIN (?) AS ?", q.BuildSubQuery(sqBuilder), bun.Name(alias))
	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunSelectQuery) JoinExpr(alias string, eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder)) SelectQuery {
	q.query.Join("JOIN (?) AS ?", eBuilder(q.eb), bun.Name(alias))
	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunSelectQuery) LeftJoin(model any, builder func(ConditionBuilder), alias ...string) SelectQuery {
	table := getTableSchema(model, q.query.DB())

	aliasToUse := table.Alias
	if len(alias) > 0 && alias[0] != constants.Empty {
		aliasToUse = alias[0]
	}

	q.query.Join(
		"LEFT JOIN ? AS ?",
		bun.Name(table.Name),
		bun.Name(aliasToUse),
	)
	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) LeftJoinTable(name string, builder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("LEFT JOIN ? AS ?", bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.Join("LEFT JOIN ?", bun.Name(name))
	}

	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) LeftJoinSubQuery(alias string, sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder)) SelectQuery {
	q.query.Join("LEFT JOIN (?) AS ?", q.BuildSubQuery(sqBuilder), bun.Name(alias))
	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunSelectQuery) LeftJoinExpr(alias string, eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder)) SelectQuery {
	q.query.Join("LEFT JOIN (?) AS ?", eBuilder(q.eb), bun.Name(alias))
	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunSelectQuery) RightJoin(model any, builder func(ConditionBuilder), alias ...string) SelectQuery {
	table := getTableSchema(model, q.query.DB())

	aliasToUse := table.Alias
	if len(alias) > 0 && alias[0] != constants.Empty {
		aliasToUse = alias[0]
	}

	q.query.Join(
		"RIGHT JOIN ? AS ?",
		bun.Name(table.Name),
		bun.Name(aliasToUse),
	)
	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) RightJoinTable(name string, builder func(ConditionBuilder), alias ...string) SelectQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("RIGHT JOIN ? AS ?", bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.Join("RIGHT JOIN ?", bun.Name(name))
	}

	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunSelectQuery) RightJoinSubQuery(alias string, sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder)) SelectQuery {
	q.query.Join("RIGHT JOIN (?) AS ?", q.BuildSubQuery(sqBuilder), bun.Name(alias))
	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunSelectQuery) RightJoinExpr(alias string, eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder)) SelectQuery {
	q.query.Join("RIGHT JOIN (?) AS ?", eBuilder(q.eb), bun.Name(alias))
	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunSelectQuery) ModelRelations(relations ...ModelRelation) SelectQuery {
	for _, r := range relations {
		applyModelRelation(r, q)
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
	if len(pageable.Sort) > 0 {
		applySort(pageable.Sort, q)
	}

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
		q.query.For("UPDATE NO WAIT")
	} else {
		q.query.For("UPDATE OF ? NO WAIT", Names(tables...))
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

func (q *BunSelectQuery) Exec(ctx context.Context, dest ...any) (sql.Result, error) {
	if q.isSubQuery {
		return nil, ErrSubQuery
	}

	r, err := q.query.Exec(ctx, dest...)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, result.ErrRecordNotFound
	}

	return r, err
}

func (q *BunSelectQuery) Scan(ctx context.Context, dest ...any) error {
	if q.isSubQuery {
		return ErrSubQuery
	}

	if err := q.query.Scan(ctx, dest...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result.ErrRecordNotFound
		}

		return err
	}

	return nil
}

func (q *BunSelectQuery) Rows(ctx context.Context) (*sql.Rows, error) {
	if q.isSubQuery {
		return nil, ErrSubQuery
	}

	rows, err := q.query.Rows(ctx)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, result.ErrRecordNotFound
	}

	return rows, err
}

func (q *BunSelectQuery) ScanAndCount(ctx context.Context, dest ...any) (int64, error) {
	if q.isSubQuery {
		return 0, ErrSubQuery
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

func (q *BunSelectQuery) Count(ctx context.Context) (int64, error) {
	if q.isSubQuery {
		return 0, ErrSubQuery
	}

	total, err := q.query.Count(ctx)

	return int64(total), err
}

func (q *BunSelectQuery) Exists(ctx context.Context) (bool, error) {
	if q.isSubQuery {
		return false, ErrSubQuery
	}

	return q.query.Exists(ctx)
}

func (q *BunSelectQuery) Unwrap() *bun.SelectQuery {
	return q.query
}
