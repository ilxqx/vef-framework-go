package orm

import (
	"context"
	"database/sql"

	"github.com/samber/lo"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

var defaultSourceAlias = "src"

// NewMergeQuery creates a new MergeQuery instance with the provided database instance.
// It initializes the query builders and sets up the table schema context for proper query building.
func NewMergeQuery(db *BunDB) *BunMergeQuery {
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

	db       *BunDB
	dialect  schema.Dialect
	eb       ExprBuilder
	query    *bun.MergeQuery
	srcAlias string
}

func (q *BunMergeQuery) DB() DB {
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

func (q *BunMergeQuery) TableFrom(model any, alias ...string) MergeQuery {
	table := q.db.TableOf(model)

	aliasToUse := table.Alias
	if len(alias) > 0 && alias[0] != constants.Empty {
		aliasToUse = alias[0]
	}

	q.query.TableExpr("? AS ?", bun.Name(table.Name), bun.Name(aliasToUse))

	return q
}

func (q *BunMergeQuery) TableExpr(builder func(ExprBuilder) any, alias ...string) MergeQuery {
	expr := builder(q.eb)
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.TableExpr("(?) AS ?", expr, bun.Name(alias[0]))
	} else {
		q.query.TableExpr("(?)", expr)
	}

	return q
}

func (q *BunMergeQuery) TableSubQuery(builder func(SelectQuery), alias ...string) MergeQuery {
	subQuery := q.BuildSubQuery(builder)
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.TableExpr("(?) AS ?", subQuery, bun.Name(alias[0]))
	} else {
		q.query.TableExpr("(?)", subQuery)
	}

	return q
}

func (q *BunMergeQuery) Using(model any, alias ...string) MergeQuery {
	table := q.db.TableOf(model)

	q.srcAlias = table.Alias
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.srcAlias = alias[0]
	}

	if q.srcAlias == constants.Empty {
		q.srcAlias = table.Name
	}

	q.query.Using("? AS ?", table.Name, bun.Name(q.srcAlias))

	return q
}

func (q *BunMergeQuery) UsingTable(table string, alias ...string) MergeQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.srcAlias = alias[0]
		q.query.Using("? AS ?", bun.Name(table), bun.Name(alias[0]))
	} else {
		q.srcAlias = table
		q.query.Using("?", bun.Name(table))
	}

	return q
}

func (q *BunMergeQuery) UsingExpr(builder func(ExprBuilder) any, alias ...string) MergeQuery {
	q.srcAlias = defaultSourceAlias
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.srcAlias = alias[0]
	}

	q.query.Using("(?) AS ?", builder(q.eb), bun.Name(q.srcAlias))

	return q
}

func (q *BunMergeQuery) UsingSubQuery(builder func(SelectQuery), alias ...string) MergeQuery {
	q.srcAlias = defaultSourceAlias
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.srcAlias = alias[0]
	}

	q.query.Using("(?) AS ?", q.BuildSubQuery(builder), bun.Name(q.srcAlias))

	return q
}

func (q *BunMergeQuery) On(builder func(ConditionBuilder)) MergeQuery {
	q.query.On("?", q.BuildCondition(builder))

	return q
}

func (q *BunMergeQuery) WhenMatched(builder ...func(ConditionBuilder)) MergeWhenBuilder {
	return newMergeWhenBuilder(q, q.srcAlias, "MATCHED", builder...)
}

func (q *BunMergeQuery) WhenNotMatched(builder ...func(ConditionBuilder)) MergeWhenBuilder {
	return newMergeWhenBuilder(q, q.srcAlias, "NOT MATCHED", builder...)
}

func (q *BunMergeQuery) WhenNotMatchedByTarget(builder ...func(ConditionBuilder)) MergeWhenBuilder {
	return newMergeWhenBuilder(q, q.srcAlias, "NOT MATCHED BY TARGET", builder...)
}

func (q *BunMergeQuery) WhenNotMatchedBySource(builder ...func(ConditionBuilder)) MergeWhenBuilder {
	return newMergeWhenBuilder(q, q.srcAlias, "NOT MATCHED BY SOURCE", builder...)
}

func (q *BunMergeQuery) Returning(columns ...string) MergeQuery {
	expr := q.eb.Exprs(
		lo.Map(columns, func(column string, _ int) any {
			return q.eb.Column(column)
		})...,
	)
	q.query.Returning("?", expr)

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
