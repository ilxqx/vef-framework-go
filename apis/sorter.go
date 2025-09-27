package apis

import "github.com/ilxqx/vef-framework-go/orm"

// Sorter is an interface that defines the methods for ordering a query.
type Sorter interface {
	// OrderBy orders the query by one or more columns in ascending order.
	// Multiple columns are processed in the order specified.
	OrderBy(columns ...string) Sorter

	// OrderByDesc orders the query by one or more columns in descending order.
	// Multiple columns are processed in the order specified.
	OrderByDesc(columns ...string) Sorter

	// OrderByExpr orders the query using a custom SQL expression.
	// This provides maximum flexibility for complex sorting scenarios.
	OrderByExpr(func(orm.ExpressionBuilders) any) Sorter
}

// querySorter is the concrete implementation of the Sorter interface.
type querySorter struct {
	query orm.SelectQuery
}

func (s *querySorter) OrderBy(columns ...string) Sorter {
	s.query.OrderBy(columns...)
	return s
}

func (s *querySorter) OrderByDesc(columns ...string) Sorter {
	s.query.OrderByDesc(columns...)
	return s
}

func (s *querySorter) OrderByExpr(builder func(orm.ExpressionBuilders) any) Sorter {
	s.query.OrderByExpr(builder)
	return s
}

// newSorter creates a new Sorter instance that wraps the provided ORM query.
// This factory function is used internally by the APIs package to create
// Sorter instances for use in SortApplier functions.
func newSorter(query orm.SelectQuery) Sorter {
	return &querySorter{query: query}
}
