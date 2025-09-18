package apis

import "github.com/ilxqx/vef-framework-go/orm"

// Sorter is an interface that defines the methods for ordering a query.
// This interface provides a fluent API for building complex sort operations
// and serves as an abstraction layer between SortApplier functions and ORM queries.
//
// Key Features:
//   - Fluent interface: all methods return Sorter for method chaining
//   - Null handling: supports explicit null ordering (NULLS FIRST/LAST)
//   - Expression support: allows custom SQL expressions for complex sorting
//   - Multiple columns: supports sorting by multiple columns in a single call
//
// Usage in SortApplier:
//
//	func mySortApplier(sorter Sorter) {
//	    sorter.OrderByDesc("created_at").OrderBy("name")
//	}
//
// This design allows SortApplier functions to be database-agnostic while
// still providing full control over query ordering.
type Sorter interface {
	// OrderBy orders the query by one or more columns in ascending order.
	// Multiple columns are processed in the order specified.
	//
	// Parameters:
	//   - columns: Column names to sort by (e.g., "name", "created_at")
	//
	// Returns:
	//   - Sorter: The same sorter instance for method chaining
	//
	// Example:
	//   sorter.OrderBy("name", "email") // ORDER BY name ASC, email ASC
	OrderBy(columns ...string) Sorter

	// OrderByNullsFirst orders the query by one or more columns in ascending order,
	// with NULL values appearing first in the result set.
	//
	// Parameters:
	//   - columns: Column names to sort by
	//
	// Returns:
	//   - Sorter: The same sorter instance for method chaining
	//
	// Example:
	//   sorter.OrderByNullsFirst("optional_field") // ORDER BY optional_field ASC NULLS FIRST
	OrderByNullsFirst(columns ...string) Sorter

	// OrderByNullsLast orders the query by one or more columns in ascending order,
	// with NULL values appearing last in the result set.
	//
	// Parameters:
	//   - columns: Column names to sort by
	//
	// Returns:
	//   - Sorter: The same sorter instance for method chaining
	//
	// Example:
	//   sorter.OrderByNullsLast("optional_field") // ORDER BY optional_field ASC NULLS LAST
	OrderByNullsLast(columns ...string) Sorter

	// OrderByDesc orders the query by one or more columns in descending order.
	// Multiple columns are processed in the order specified.
	//
	// Parameters:
	//   - columns: Column names to sort by
	//
	// Returns:
	//   - Sorter: The same sorter instance for method chaining
	//
	// Example:
	//   sorter.OrderByDesc("created_at", "priority") // ORDER BY created_at DESC, priority DESC
	OrderByDesc(columns ...string) Sorter

	// OrderByDescNullsFirst orders the query by one or more columns in descending order,
	// with NULL values appearing first in the result set.
	//
	// Parameters:
	//   - columns: Column names to sort by
	//
	// Returns:
	//   - Sorter: The same sorter instance for method chaining
	//
	// Example:
	//   sorter.OrderByDescNullsFirst("score") // ORDER BY score DESC NULLS FIRST
	OrderByDescNullsFirst(columns ...string) Sorter

	// OrderByDescNullsLast orders the query by one or more columns in descending order,
	// with NULL values appearing last in the result set.
	//
	// Parameters:
	//   - columns: Column names to sort by
	//
	// Returns:
	//   - Sorter: The same sorter instance for method chaining
	//
	// Example:
	//   sorter.OrderByDescNullsLast("score") // ORDER BY score DESC NULLS LAST
	OrderByDescNullsLast(columns ...string) Sorter

	// OrderByExpr orders the query using a custom SQL expression.
	// This provides maximum flexibility for complex sorting scenarios.
	//
	// Parameters:
	//   - expr: SQL expression for ordering (e.g., "CASE WHEN status = 'urgent' THEN 0 ELSE 1 END")
	//   - args: Optional arguments to be safely interpolated into the expression
	//
	// Returns:
	//   - Sorter: The same sorter instance for method chaining
	//
	// Example:
	//   sorter.OrderByExpr("LENGTH(?)", "description") // ORDER BY LENGTH(description)
	//   sorter.OrderByExpr("RANDOM()") // ORDER BY RANDOM() for random ordering
	//
	// Security Note: Use parameter placeholders (?) for user input to prevent SQL injection
	OrderByExpr(expr string, args ...any) Sorter
}

// querySorter is the concrete implementation of the Sorter interface.
// It wraps an ORM query and delegates sorting operations to the underlying query builder.
// This implementation ensures that all sorting operations are applied to the same query instance.
type querySorter struct {
	query orm.Query // The underlying ORM query to which sorting operations are applied
}

// OrderBy implements the Sorter interface by delegating to the underlying ORM query.
func (s *querySorter) OrderBy(columns ...string) Sorter {
	s.query.OrderBy(columns...)
	return s
}

// OrderByNullsFirst implements the Sorter interface by delegating to the underlying ORM query.
func (s *querySorter) OrderByNullsFirst(columns ...string) Sorter {
	s.query.OrderByNullsFirst(columns...)
	return s
}

// OrderByNullsLast implements the Sorter interface by delegating to the underlying ORM query.
func (s *querySorter) OrderByNullsLast(columns ...string) Sorter {
	s.query.OrderByNullsLast(columns...)
	return s
}

// OrderByDesc implements the Sorter interface by delegating to the underlying ORM query.
func (s *querySorter) OrderByDesc(columns ...string) Sorter {
	s.query.OrderByDesc(columns...)
	return s
}

// OrderByDescNullsFirst implements the Sorter interface by delegating to the underlying ORM query.
func (s *querySorter) OrderByDescNullsFirst(columns ...string) Sorter {
	s.query.OrderByDescNullsFirst(columns...)
	return s
}

// OrderByDescNullsLast implements the Sorter interface by delegating to the underlying ORM query.
func (s *querySorter) OrderByDescNullsLast(columns ...string) Sorter {
	s.query.OrderByDescNullsLast(columns...)
	return s
}

// OrderByExpr implements the Sorter interface by delegating to the underlying ORM query.
func (s *querySorter) OrderByExpr(expr string, args ...any) Sorter {
	s.query.OrderByExpr(expr, args...)
	return s
}

// newSorter creates a new Sorter instance that wraps the provided ORM query.
// This factory function is used internally by the APIs package to create
// Sorter instances for use in SortApplier functions.
//
// Parameters:
//   - query: The ORM query to wrap with sorting capabilities
//
// Returns:
//   - Sorter: A new Sorter instance that delegates to the provided query
//
// Internal Usage:
//
//	sorter := newSorter(db.NewQuery().Model(&User{}))
//	sortApplier(sorter)
func newSorter(query orm.Query) Sorter {
	return &querySorter{query: query}
}
