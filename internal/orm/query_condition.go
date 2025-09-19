package orm

import (
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/uptrace/bun"
)

// queryConditionBuilder is a builder for building query conditions.
type queryConditionBuilder struct {
	*richConditionBuilder
}

// newQueryConditionBuilder creates a new query builder.
func newQueryConditionBuilder(table *orm.Table, builder bun.QueryBuilder, subQueryBuilder func(builder func(query orm.Query)) *bun.SelectQuery) *queryConditionBuilder {
	return &queryConditionBuilder{
		richConditionBuilder: &richConditionBuilder{
			table:           table,
			subQueryBuilder: subQueryBuilder,
			and: func(query string, args ...any) {
				builder.Where(query, args...)
			},
			or: func(query string, args ...any) {
				builder.WhereOr(query, args...)
			},
			group: func(sep string, cb func(orm.ConditionBuilder)) {
				builder.WhereGroup(
					sep,
					func(qb bun.QueryBuilder) bun.QueryBuilder {
						cb(newQueryConditionBuilder(table, qb, subQueryBuilder))
						return qb
					},
				)
			},
		},
	}
}
