package orm

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

// NewMergeQuery creates a new MergeQuery instance with the provided database instance.
// It initializes the query builders and sets up the table schema context for proper query building.
func NewMergeQuery(db *BunDb) *BunMergeQuery {
	eb := &QueryExprBuilder{}
	mq := db.db.NewMerge()
	dialect := db.db.Dialect()
	query := &BunMergeQuery{
		QueryBuilder: newQueryBuilder(db, dialect, mq, eb),

		db:      db,
		dialect: dialect,
		query:   mq,
		eb:      eb,
	}
	eb.qb = query

	return query
}

// BunMergeQuery is the concrete implementation of MergeQuery interface.
// It wraps bun.MergeQuery and provides additional functionality for expression building.
type BunMergeQuery struct {
	QueryBuilder

	db      *BunDb
	dialect schema.Dialect
	eb      ExprBuilder
	query   *bun.MergeQuery
}

func (q *BunMergeQuery) Db() Db {
	return q.db
}

func (q *BunMergeQuery) With(name string, builder func(SelectQuery)) MergeQuery {
	q.query.With(name, q.BuildSubQuery(builder))

	return q
}

func (q *BunMergeQuery) WithValues(name string, model any, withOrder ...bool) MergeQuery {
	values := q.query.NewValues(model)
	if len(withOrder) > 0 && withOrder[0] {
		values.WithOrder()
	}

	q.query.With(name, values)

	return q
}

func (q *BunMergeQuery) WithRecursive(name string, builder func(SelectQuery)) MergeQuery {
	q.query.WithRecursive(name, q.BuildSubQuery(builder))

	return q
}

func (q *BunMergeQuery) Model(model any) MergeQuery {
	q.query.Model(model)

	return q
}

func (q *BunMergeQuery) ModelTable(name string, alias ...string) MergeQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.ModelTableExpr("? AS ?", bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.ModelTableExpr("? AS ?TableAlias", bun.Name(name))
	}

	return q
}

func (q *BunMergeQuery) Table(name string, alias ...string) MergeQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.TableExpr("? AS ?", bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.Table(name)
	}

	return q
}

func (q *BunMergeQuery) TableExpr(alias string, builder func(ExprBuilder) any) MergeQuery {
	q.query.TableExpr("(?) AS ?", builder(q.eb), bun.Name(alias))

	return q
}

func (q *BunMergeQuery) TableSubQuery(alias string, builder func(SelectQuery)) MergeQuery {
	q.query.TableExpr("(?) AS ?", q.BuildSubQuery(builder), bun.Name(alias))

	return q
}

func (q *BunMergeQuery) Using(source string, alias ...string) MergeQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Using("? AS ?", bun.Name(source), bun.Name(alias[0]))
	} else {
		q.query.Using("?", bun.Name(source))
	}

	return q
}

func (q *BunMergeQuery) UsingModel(model any) MergeQuery {
	q.query.Using("?", model)

	return q
}

func (q *BunMergeQuery) UsingExpr(alias string, builder func(ExprBuilder) any) MergeQuery {
	q.query.Using("(?) AS ?", builder(q.eb), bun.Name(alias))

	return q
}

func (q *BunMergeQuery) UsingSubQuery(alias string, builder func(SelectQuery)) MergeQuery {
	q.query.Using("(?) AS ?", q.BuildSubQuery(builder), bun.Name(alias))

	return q
}

func (q *BunMergeQuery) UsingValues(model any, columns ...string) MergeQuery {
	values := q.query.NewValues(model)
	if len(columns) > 0 {
		values.Column(columns...)
	}

	q.query.Using("?", values)

	return q
}

func (q *BunMergeQuery) On(builder func(ConditionBuilder)) MergeQuery {
	q.query.On("?", q.BuildCondition(builder))

	return q
}

func (q *BunMergeQuery) WhenMatched(builder ...func(ConditionBuilder)) MergeWhenBuilder {
	return newMergeWhenBuilder(q, "MATCHED", builder...)
}

func (q *BunMergeQuery) WhenNotMatched(builder ...func(ConditionBuilder)) MergeWhenBuilder {
	return newMergeWhenBuilder(q, "NOT MATCHED", builder...)
}

func (q *BunMergeQuery) WhenNotMatchedBySource(builder ...func(ConditionBuilder)) MergeWhenBuilder {
	return newMergeWhenBuilder(q, "NOT MATCHED BY SOURCE", builder...)
}

func (q *BunMergeQuery) Returning(columns ...string) MergeQuery {
	q.query.Returning("?", Names(columns...))

	return q
}

func (q *BunMergeQuery) ReturningAll() MergeQuery {
	q.query.Returning(columnAll)

	return q
}

func (q *BunMergeQuery) ReturningNone() MergeQuery {
	q.query.Returning(sqlNull)

	return q
}

func (q *BunMergeQuery) Apply(fns ...ApplyFunc[MergeQuery]) MergeQuery {
	for _, fn := range fns {
		if fn != nil {
			fn(q)
		}
	}

	return q
}

func (q *BunMergeQuery) ApplyIf(condition bool, fns ...ApplyFunc[MergeQuery]) MergeQuery {
	if condition {
		return q.Apply(fns...)
	}

	return q
}

func (q *BunMergeQuery) Exec(ctx context.Context, dest ...any) (sql.Result, error) {
	return q.query.Exec(ctx, dest...)
}

func (q *BunMergeQuery) Scan(ctx context.Context, dest ...any) error {
	return q.query.Scan(ctx, dest...)
}

func (q *BunMergeQuery) Unwrap() *bun.MergeQuery {
	return q.query
}
