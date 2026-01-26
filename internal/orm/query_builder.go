package orm

import (
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// BaseQueryBuilder provides a common implementation for QueryBuilder interface.
// It can be embedded in concrete query types to reduce code duplication.
type BaseQueryBuilder struct {
	db      *BunDB
	dialect schema.Dialect
	query   interface {
		bun.Query
		fmt.Stringer

		NewSelect() *bun.SelectQuery
	}
	eb ExprBuilder
}

// Dialect returns the dialect of the current database connection.
func (b *BaseQueryBuilder) Dialect() schema.Dialect {
	return b.dialect
}

// GetTable returns the table information for the current query.
func (b *BaseQueryBuilder) GetTable() *schema.Table {
	return getTableSchemaFromQuery(b.query)
}

// Query returns the query of the current query instance.
func (b *BaseQueryBuilder) Query() bun.Query {
	return b.query
}

// ExprBuilder returns the expression builder for this query.
func (b *BaseQueryBuilder) ExprBuilder() ExprBuilder {
	return b.eb
}

// CreateSubQuery creates a new subquery from the given bun.SelectQuery.
func (b *BaseQueryBuilder) CreateSubQuery(subQuery *bun.SelectQuery) SelectQuery {
	eb := &QueryExprBuilder{
		qb: b,
	}
	queryBuilder := newQueryBuilder(b.db, b.dialect, subQuery, eb)
	query := &BunSelectQuery{
		QueryBuilder: queryBuilder,

		db:         b.db,
		dialect:    b.dialect,
		query:      subQuery,
		eb:         eb,
		isSubQuery: true,
	}
	eb.qb = query

	return query
}

// BuildSubQuery constructs a subquery using a builder function.
func (b *BaseQueryBuilder) BuildSubQuery(builder func(query SelectQuery)) *bun.SelectQuery {
	subQuery := b.query.NewSelect()
	wrappedQuery := b.CreateSubQuery(subQuery)
	builder(wrappedQuery)

	// Apply deferred select state before returning the subquery
	// This ensures that select operations in the subquery are properly applied
	if sq, ok := wrappedQuery.(*BunSelectQuery); ok {
		sq.applySelectState()
	}

	return subQuery
}

// BuildCondition creates a condition builder for WHERE clauses.
func (b *BaseQueryBuilder) BuildCondition(builder func(ConditionBuilder)) interface {
	schema.QueryAppender
	ConditionBuilder
} {
	cb := newConditionBuilder(b)
	builder(cb)

	return cb
}

// String returns the SQL query string.
func (b *BaseQueryBuilder) String() string {
	return b.query.String()
}

// newQueryBuilder creates a new query builder.
func newQueryBuilder(db *BunDB, dialect schema.Dialect, query interface {
	bun.Query
	fmt.Stringer

	NewSelect() *bun.SelectQuery
}, eb ExprBuilder,
) *BaseQueryBuilder {
	return &BaseQueryBuilder{
		db:      db,
		dialect: dialect,
		query:   query,
		eb:      eb,
	}
}
