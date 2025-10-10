package orm

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

// NewDeleteQuery creates a new DeleteQuery instance with the provided database instance.
// It initializes the query builders and sets up the table schema context for proper query building.
func NewDeleteQuery(db *BunDb) *BunDeleteQuery {
	eb := &QueryExprBuilder{}
	dq := db.db.NewDelete()
	dialect := db.db.Dialect()
	query := &BunDeleteQuery{
		QueryBuilder: newQueryBuilder(dialect, dq, eb),

		db:      db,
		dialect: dialect,
		query:   dq,
		eb:      eb,
	}
	eb.qb = query

	return query
}

// BunDeleteQuery is the concrete implementation of DeleteQuery interface.
// It wraps bun.DeleteQuery and provides additional functionality for expression building.
type BunDeleteQuery struct {
	QueryBuilder

	db      *BunDb
	dialect schema.Dialect
	eb      ExprBuilder
	query   *bun.DeleteQuery
}

func (q *BunDeleteQuery) Db() Db {
	return q.db
}

func (q *BunDeleteQuery) With(name string, builder func(SelectQuery)) DeleteQuery {
	q.query.With(name, q.BuildSubQuery(builder))

	return q
}

func (q *BunDeleteQuery) WithValues(name string, model any, withOrder ...bool) DeleteQuery {
	values := q.query.NewValues(model)
	if len(withOrder) > 0 && withOrder[0] {
		values.WithOrder()
	}

	q.query.With(name, values)

	return q
}

func (q *BunDeleteQuery) WithRecursive(name string, builder func(SelectQuery)) DeleteQuery {
	q.query.WithRecursive(name, q.BuildSubQuery(builder))

	return q
}

func (q *BunDeleteQuery) Model(model any) DeleteQuery {
	q.query.Model(model)

	return q
}

func (q *BunDeleteQuery) ModelTable(name string, alias ...string) DeleteQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.ModelTableExpr("? AS ?", bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.ModelTableExpr("? AS ?TableAlias", bun.Name(name))
	}

	return q
}

func (q *BunDeleteQuery) Table(name string, alias ...string) DeleteQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.TableExpr("? AS ?", bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.Table(name)
	}

	return q
}

func (q *BunDeleteQuery) TableExpr(alias string, builder func(ExprBuilder) any) DeleteQuery {
	q.query.TableExpr("(?) AS ?", builder(q.eb), bun.Name(alias))

	return q
}

func (q *BunDeleteQuery) TableSubQuery(alias string, builder func(SelectQuery)) DeleteQuery {
	q.query.TableExpr("(?) AS ?", q.BuildSubQuery(builder), bun.Name(alias))

	return q
}

func (q *BunDeleteQuery) Where(builder func(ConditionBuilder)) DeleteQuery {
	cb := newQueryConditionBuilder(q.query.QueryBuilder(), q)
	builder(cb)

	return q
}

func (q *BunDeleteQuery) WherePK(columns ...string) DeleteQuery {
	q.query.WherePK(columns...)

	return q
}

func (q *BunDeleteQuery) WhereDeleted() DeleteQuery {
	q.query.WhereDeleted()

	return q
}

func (q *BunDeleteQuery) IncludeDeleted() DeleteQuery {
	q.query.WhereAllWithDeleted()

	return q
}

func (q *BunDeleteQuery) OrderBy(columns ...string) DeleteQuery {
	q.query.Order(columns...)

	return q
}

func (q *BunDeleteQuery) OrderByDesc(columns ...string) DeleteQuery {
	for _, column := range columns {
		q.query.OrderExpr("? DESC", bun.Ident(column))
	}

	return q
}

func (q *BunDeleteQuery) OrderByExpr(builder func(ExprBuilder) any) DeleteQuery {
	q.query.OrderExpr("?", builder(q.eb))

	return q
}

func (q *BunDeleteQuery) ForceDelete() DeleteQuery {
	q.query.ForceDelete()

	return q
}

func (q *BunDeleteQuery) Limit(limit int) DeleteQuery {
	q.query.Limit(limit)

	return q
}

func (q *BunDeleteQuery) Returning(columns ...string) DeleteQuery {
	q.query.Returning("?", Names(columns...))

	return q
}

func (q *BunDeleteQuery) ReturningAll() DeleteQuery {
	q.query.Returning(columnAll)

	return q
}

func (q *BunDeleteQuery) ReturningNone() DeleteQuery {
	q.query.Returning(sqlNull)

	return q
}

func (q *BunDeleteQuery) Apply(fns ...ApplyFunc[DeleteQuery]) DeleteQuery {
	for _, fn := range fns {
		if fn != nil {
			fn(q)
		}
	}

	return q
}

func (q *BunDeleteQuery) ApplyIf(condition bool, fns ...ApplyFunc[DeleteQuery]) DeleteQuery {
	if condition {
		return q.Apply(fns...)
	}

	return q
}

func (q *BunDeleteQuery) Exec(ctx context.Context, dest ...any) (sql.Result, error) {
	return q.query.Exec(ctx, dest...)
}

func (q *BunDeleteQuery) Scan(ctx context.Context, dest ...any) error {
	return q.query.Scan(ctx, dest...)
}

func (q *BunDeleteQuery) Unwrap() *bun.DeleteQuery {
	return q.query
}
