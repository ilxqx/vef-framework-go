package orm

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/page"
)

// DBAccessor provides access to the underlying DB instance.
type DBAccessor interface {
	// DB returns the underlying DB instance.
	DB() DB
}

// QueryExecutor is an interface that defines the methods for executing database queries.
// It provides the basic execution methods that all query types must implement.
type QueryExecutor interface {
	// Exec executes a query and returns the result.
	Exec(ctx context.Context, dest ...any) (sql.Result, error)
	// Scan scans the result into a slice of any type.
	Scan(ctx context.Context, dest ...any) error
}

// CTE is an interface that defines the methods for creating Common Table Expressions (CTEs).
// CTEs allow you to define temporary result sets that exist only for the duration of a single query.
type CTE[T QueryExecutor] interface {
	// With creates a common table expression.
	With(name string, builder func(query SelectQuery)) T
	// WithValues creates a common table expression with values.
	WithValues(name string, model any, withOrder ...bool) T
	// WithRecursive creates a recursive common table expression.
	WithRecursive(name string, builder func(query SelectQuery)) T
}

// Selectable is an interface that defines the methods for column selection in queries.
// It provides methods to specify which columns to include or exclude from the result set.
type Selectable[T QueryExecutor] interface {
	// SelectAll selects all columns.
	SelectAll() T
	// Select selects specific columns.
	Select(columns ...string) T
	// Exclude excludes specific columns.
	Exclude(columns ...string) T
	// ExcludeAll excludes all columns.
	ExcludeAll() T
}

// TableSource is an interface that defines the methods for specifying table sources in queries.
// It supports both model-based and raw table references with optional aliases.
type TableSource[T QueryExecutor] interface {
	// Model sets the primary table with automatic table name and alias resolution from model structure.
	Model(model any) T
	// ModelTable overrides the table name and alias auto-resolved by Model method.
	// It must be called after Model to manually specify the table name and alias.
	ModelTable(name string, alias ...string) T
	// Table sets a table using string name and optional alias directly.
	Table(name string, alias ...string) T
	// TableFrom sets a table by auto-resolving table name and alias from model.
	// The provided alias parameter takes precedence over the model's default alias.
	TableFrom(model any, alias ...string) T
	// TableExpr sets a table using a custom expression builder with optional alias.
	TableExpr(builder func(ExprBuilder) any, alias ...string) T
	// TableSubQuery sets a table using a subquery with optional alias.
	TableSubQuery(builder func(query SelectQuery), alias ...string) T
}

// JoinOperations is an interface that defines the methods for joining tables in queries.
// It supports all standard SQL join types including INNER, LEFT, RIGHT, FULL, and CROSS joins with different source types.
type JoinOperations[T any] interface {
	// Join performs an INNER JOIN with a model.
	Join(model any, builder func(ConditionBuilder), alias ...string) T
	// JoinTable performs an INNER JOIN with a table name.
	JoinTable(name string, builder func(ConditionBuilder), alias ...string) T
	// JoinSubQuery performs an INNER JOIN with a subquery.
	JoinSubQuery(sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder), alias ...string) T
	// JoinExpr performs an INNER JOIN with a custom expression.
	JoinExpr(eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder), alias ...string) T

	// LeftJoin performs a LEFT OUTER JOIN with a model.
	LeftJoin(model any, builder func(ConditionBuilder), alias ...string) T
	// LeftJoinTable performs a LEFT OUTER JOIN with a table name.
	LeftJoinTable(name string, builder func(ConditionBuilder), alias ...string) T
	// LeftJoinSubQuery performs a LEFT OUTER JOIN with a subquery.
	LeftJoinSubQuery(sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder), alias ...string) T
	// LeftJoinExpr performs a LEFT OUTER JOIN with a custom expression.
	LeftJoinExpr(eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder), alias ...string) T

	// RightJoin performs a RIGHT OUTER JOIN with a model.
	RightJoin(model any, builder func(ConditionBuilder), alias ...string) T
	// RightJoinTable performs a RIGHT OUTER JOIN with a table name.
	RightJoinTable(name string, builder func(ConditionBuilder), alias ...string) T
	// RightJoinSubQuery performs a RIGHT OUTER JOIN with a subquery.
	RightJoinSubQuery(sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder), alias ...string) T
	// RightJoinExpr performs a RIGHT OUTER JOIN with a custom expression.
	RightJoinExpr(eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder), alias ...string) T

	// FullJoin performs a FULL OUTER JOIN with a model.
	FullJoin(model any, builder func(ConditionBuilder), alias ...string) T
	// FullJoinTable performs a FULL OUTER JOIN with a table name.
	FullJoinTable(name string, builder func(ConditionBuilder), alias ...string) T
	// FullJoinSubQuery performs a FULL OUTER JOIN with a subquery.
	FullJoinSubQuery(sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder), alias ...string) T
	// FullJoinExpr performs a FULL OUTER JOIN with a custom expression.
	FullJoinExpr(eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder), alias ...string) T

	// CrossJoin performs a CROSS JOIN with a model.
	CrossJoin(model any, alias ...string) T
	// CrossJoinTable performs a CROSS JOIN with a table name.
	CrossJoinTable(name string, alias ...string) T
	// CrossJoinSubQuery performs a CROSS JOIN with a subquery.
	CrossJoinSubQuery(sqBuilder func(query SelectQuery), alias ...string) T
	// CrossJoinExpr performs a CROSS JOIN with a custom expression.
	CrossJoinExpr(eBuilder func(ExprBuilder) any, alias ...string) T
}

// Filterable is an interface that defines the methods for adding WHERE clauses to queries.
// It provides methods for filtering results based on conditions and supports soft delete operations.
type Filterable[T QueryExecutor] interface {
	// Where adds a where clause to the query.
	Where(func(ConditionBuilder)) T
	// WherePK adds a where clause to the query using the primary key.
	WherePK(columns ...string) T
	// WhereDeleted adds a where clause to the query using the deleted column.
	WhereDeleted() T
	// IncludeDeleted includes soft-deleted records in the query results.
	IncludeDeleted() T
}

// Orderable is an interface that defines the methods for ordering query results.
// It supports ordering by columns and expressions in ascending or descending order.
type Orderable[T QueryExecutor] interface {
	// OrderBy orders the query by a column.
	OrderBy(columns ...string) T
	// OrderByDesc orders the query by a column in descending order.
	OrderByDesc(columns ...string) T
	// OrderByExpr orders the query by an expression.
	OrderByExpr(func(ExprBuilder) any) T
}

// Limitable is an interface that defines the methods for limiting the number of rows returned by a query.
// It provides the LIMIT clause functionality for result set size control.
type Limitable[T QueryExecutor] interface {
	// Limit limits the number of rows returned by the query.
	Limit(limit int) T
}

// ColumnUpdatable is an interface that defines the methods for setting column values in queries.
// It supports both direct value assignment and expression-based column updates.
type ColumnUpdatable[T QueryExecutor] interface {
	// Column sets a column to a specific value.
	Column(name string, value any) T
	// ColumnExpr sets a column using an expression builder.
	ColumnExpr(name string, builder func(ExprBuilder) any) T
}

// Returnable is an interface that defines the methods for specifying RETURNING clauses in queries.
// It allows queries to return data after INSERT, UPDATE, or DELETE operations.
type Returnable[T QueryExecutor] interface {
	// Returning returns the query with the specified columns.
	Returning(columns ...string) T
	// ReturningAll returns the query with all columns.
	ReturningAll() T
	// ReturningNone returns the query with no columns.
	ReturningNone() T
}

// Unwrapper is an interface that defines the method for unwrapping the underlying query object.
// It provides access to the original wrapped query implementation for advanced use cases.
type Unwrapper[T any] interface {
	// Unwrap returns the underlying query object.
	Unwrap() T
}

// ApplyFunc is a function type that applies a shared operation to a query.
// It enables reusable query modifications that can be applied to different query types.
type ApplyFunc[T any] func(T)

// Applier is an interface that defines the methods for applying shared operations to queries.
// It enables reusable query modifications and conditional application of operations.
type Applier[T any] interface {
	// Apply applies shared operations.
	Apply(fns ...ApplyFunc[T]) T
	// ApplyIf applies shared operations if the condition is true.
	ApplyIf(condition bool, fns ...ApplyFunc[T]) T
}

// DialectExprBuilder represents a zero-argument callback that returns a QueryAppender.
type DialectExprBuilder func() schema.QueryAppender

// DialectExprs defines database-specific expression builders for cross-database compatibility.
// It allows users to define custom expressions that work across different database engines
// by providing database-specific implementations.
type DialectExprs struct {
	// Oracle expression builder for Oracle database.
	Oracle DialectExprBuilder
	// SQL Server expression builder for SQL Server database.
	SQLServer DialectExprBuilder
	// Postgres expression builder for PostgreSQL database.
	Postgres DialectExprBuilder
	// MySQL expression builder for MySQL database.
	MySQL DialectExprBuilder
	// SQLite expression builder for SQLite database.
	SQLite DialectExprBuilder
	// Default expression builder used when database-specific builder is not available.
	Default DialectExprBuilder
}

// DialectAction represents a zero-argument callback.
type DialectAction func()

// DialectExecs defines database-specific callbacks for running side-effect logic
// without returning a SQL expression.
type DialectExecs struct {
	// Oracle callback for Oracle database.
	Oracle DialectAction
	// SQL Server callback for SQL Server database.
	SQLServer DialectAction
	// Postgres callback for PostgreSQL database.
	Postgres DialectAction
	// MySQL callback for MySQL database.
	MySQL DialectAction
	// SQLite callback for SQLite database.
	SQLite DialectAction
	// Default callback used when database-specific callback is not available.
	Default DialectAction
}

// DialectActionErr represents a zero-argument callback that can return an error.
type DialectActionErr func() error

// DialectExecsWithErr defines database-specific callbacks that may return an error.
type DialectExecsWithErr struct {
	// Oracle callback for Oracle database.
	Oracle DialectActionErr
	// SQL Server callback for SQL Server database.
	SQLServer DialectActionErr
	// Postgres callback for PostgreSQL database.
	Postgres DialectActionErr
	// MySQL callback for MySQL database.
	MySQL DialectActionErr
	// SQLite callback for SQLite database.
	SQLite DialectActionErr
	// Default callback used when database-specific callback is not available.
	Default DialectActionErr
}

// DialectFragmentBuilder represents a zero-argument callback that returns a query fragment buffer.
type DialectFragmentBuilder func() ([]byte, error)

// DialectFragments defines database-specific callbacks that produce query fragments.
type DialectFragments struct {
	// Oracle callback for Oracle database.
	Oracle DialectFragmentBuilder
	// SQL Server callback for SQL Server database.
	SQLServer DialectFragmentBuilder
	// Postgres callback for PostgreSQL database.
	Postgres DialectFragmentBuilder
	// MySQL callback for MySQL database.
	MySQL DialectFragmentBuilder
	// SQLite callback for SQLite database.
	SQLite DialectFragmentBuilder
	// Default callback used when database-specific callback is not available.
	Default DialectFragmentBuilder
}

// QueryBuilder defines the common interface for building subqueries and conditions.
// It provides a unified way to create subqueries and condition builders across different query types.
type QueryBuilder interface {
	fmt.Stringer

	// Dialect returns the database dialect for cross-database compatibility.
	Dialect() schema.Dialect
	// GetTable returns the table information for the current query.
	GetTable() *schema.Table
	// Query returns the underlying bun query instance.
	Query() bun.Query
	// ExprBuilder returns the expression builder for this query.
	ExprBuilder() ExprBuilder
	// CreateSubQuery creates a new subquery from the given bun.SelectQuery.
	// It returns a SelectQuery that can be used to build complex nested queries.
	CreateSubQuery(subQuery *bun.SelectQuery) SelectQuery
	// BuildSubQuery constructs a subquery using a builder function.
	// The builder function receives a SelectQuery to configure the subquery.
	// Returns the configured bun.SelectQuery for use in parent queries.
	BuildSubQuery(builder func(query SelectQuery)) *bun.SelectQuery
	// BuildCondition creates a condition builder for WHERE clauses.
	// The builder function receives a ConditionBuilder to configure conditions.
	// Returns the configured ConditionBuilder for use in query filtering.
	BuildCondition(builder func(ConditionBuilder)) interface {
		schema.QueryAppender
		ConditionBuilder
	}
}

// ExprBuilder provides methods for building various SQL expressions and operations.
// It offers a fluent Api for constructing complex SQL expressions including aggregates, functions, and conditional logic.
type ExprBuilder interface {
	// Column builds a column expression with proper alias handling.
	// If withTableAlias is false, skips automatic table alias addition even when table exists.
	Column(column string, withTableAlias ...bool) schema.QueryAppender
	// TableColumns returns a table columns expression (?TableColumns or ?Columns).
	TableColumns(withTableAlias ...bool) schema.QueryAppender
	// AllColumns returns a wildcard column expression for all columns.
	AllColumns(tableAlias ...string) schema.QueryAppender
	// Null returns the NULL SQL literal.
	Null() schema.QueryAppender
	// IsNull checks if an expression is NULL.
	IsNull(expr any) schema.QueryAppender
	// IsNotNull checks if an expression is not NULL.
	IsNotNull(expr any) schema.QueryAppender
	// Literal builds a literal expression.
	Literal(value any) schema.QueryAppender
	// Order builds an ORDER BY expression.
	Order(func(OrderBuilder)) schema.QueryAppender
	// Case creates a CASE expression builder for conditional logic.
	Case(func(CaseBuilder)) schema.QueryAppender
	// SubQuery creates a subquery expression for use in larger queries.
	SubQuery(func(SelectQuery)) schema.QueryAppender
	// Exists creates an EXISTS subquery expression.
	Exists(func(SelectQuery)) schema.QueryAppender
	// NotExists creates a NOT EXISTS subquery expression.
	NotExists(func(SelectQuery)) schema.QueryAppender
	// Paren wraps an expression in parentheses.
	Paren(expr any) schema.QueryAppender
	// Not creates a negation expression (NOT expr).
	Not(expr any) schema.QueryAppender
	// Any wraps a subquery with the ANY operator.
	Any(func(SelectQuery)) schema.QueryAppender
	// All wraps a subquery with the ALL operator.
	All(func(SelectQuery)) schema.QueryAppender

	// ========== Arithmetic Operators ==========

	// Add creates an addition expression (left + right).
	Add(left, right any) schema.QueryAppender
	// Subtract creates a subtraction expression (left - right).
	Subtract(left, right any) schema.QueryAppender
	// Multiply creates a multiplication expression (left * right).
	Multiply(left, right any) schema.QueryAppender
	// Divide creates a division expression (left / right).
	Divide(left, right any) schema.QueryAppender

	// ========== Comparison Operators ==========

	// Equals creates an equality comparison expression (left = right).
	Equals(left, right any) schema.QueryAppender
	// NotEquals creates an inequality comparison expression (left <> right).
	NotEquals(left, right any) schema.QueryAppender
	// GreaterThan creates a greater-than comparison expression (left > right).
	GreaterThan(left, right any) schema.QueryAppender
	// GreaterThanOrEqual creates a greater-than-or-equal comparison expression (left >= right).
	GreaterThanOrEqual(left, right any) schema.QueryAppender
	// LessThan creates a less-than comparison expression (left < right).
	LessThan(left, right any) schema.QueryAppender
	// LessThanOrEqual creates a less-than-or-equal comparison expression (left <= right).
	LessThanOrEqual(left, right any) schema.QueryAppender
	// Between creates a between comparison expression (expr BETWEEN lower AND upper).
	Between(expr, lower, upper any) schema.QueryAppender
	// NotBetween creates a not between comparison expression (expr NOT BETWEEN lower AND upper).
	NotBetween(expr, lower, upper any) schema.QueryAppender
	// In creates an IN comparison expression (expr IN (values...)).
	In(expr any, values ...any) schema.QueryAppender
	// NotIn creates a NOT IN comparison expression (expr NOT IN (values...)).
	NotIn(expr any, values ...any) schema.QueryAppender
	// IsTrue checks if a boolean expression is TRUE.
	IsTrue(expr any) schema.QueryAppender
	// IsFalse checks if a boolean expression is FALSE.
	IsFalse(expr any) schema.QueryAppender

	// ========== Expression Building ==========

	// Expr creates an expression builder for complex SQL logic.
	Expr(expr string, args ...any) schema.QueryAppender
	// Exprs creates an expression builder for complex SQL logic.
	Exprs(exprs ...any) schema.QueryAppender
	// ExprsWithSep creates an expression builder for complex SQL logic with a separator.
	ExprsWithSep(separator any, exprs ...any) schema.QueryAppender
	// ExprByDialect creates a cross-database compatible expression.
	// It selects the appropriate expression builder based on the current database dialect.
	ExprByDialect(exprs DialectExprs) schema.QueryAppender
	// ExecByDialect selects the appropriate callback based on the current database dialect.
	ExecByDialect(execs DialectExecs)
	// ExecByDialectWithErr runs dialect-specific callbacks and returns any error encountered.
	ExecByDialectWithErr(execs DialectExecsWithErr) error
	// FragmentByDialect selects the appropriate query fragment builder based on the current database dialect.
	FragmentByDialect(fragments DialectFragments) ([]byte, error)

	// ========== Aggregate Functions ==========

	// Count builds a COUNT aggregate expression using a builder callback.
	Count(func(CountBuilder)) schema.QueryAppender
	// CountColumn builds a COUNT(column) aggregate expression.
	CountColumn(column string, distinct ...bool) schema.QueryAppender
	// CountAll builds a COUNT(*) aggregate expression.
	CountAll(distinct ...bool) schema.QueryAppender
	// Sum builds a SUM aggregate expression using a builder callback.
	Sum(func(SumBuilder)) schema.QueryAppender
	// SumColumn builds a SUM(column) aggregate expression.
	SumColumn(column string, distinct ...bool) schema.QueryAppender
	// Avg builds an AVG aggregate expression using a builder callback.
	Avg(func(AvgBuilder)) schema.QueryAppender
	// AvgColumn builds an AVG(column) aggregate expression.
	AvgColumn(column string, distinct ...bool) schema.QueryAppender
	// Min builds a MIN aggregate expression using a builder callback.
	Min(func(MinBuilder)) schema.QueryAppender
	// MinColumn builds a MIN(column) aggregate expression.
	MinColumn(column string) schema.QueryAppender
	// Max builds a MAX aggregate expression using a builder callback.
	Max(func(MaxBuilder)) schema.QueryAppender
	// MaxColumn builds a MAX(column) aggregate expression.
	MaxColumn(column string) schema.QueryAppender
	// StringAgg builds a STRING_AGG aggregate expression using a builder callback.
	StringAgg(func(StringAggBuilder)) schema.QueryAppender
	// ArrayAgg builds an ARRAY_AGG aggregate expression using a builder callback.
	ArrayAgg(func(ArrayAggBuilder)) schema.QueryAppender
	// JSONObjectAgg builds a JSON_OBJECT_AGG aggregate expression using a builder callback.
	JSONObjectAgg(func(JSONObjectAggBuilder)) schema.QueryAppender
	// JSONArrayAgg builds a JSON_ARRAY_AGG aggregate expression using a builder callback.
	JSONArrayAgg(func(JSONArrayAggBuilder)) schema.QueryAppender
	// BitOr builds a BIT_OR aggregate expression using a builder callback.
	BitOr(func(BitOrBuilder)) schema.QueryAppender
	// BitAnd builds a BIT_AND aggregate expression using a builder callback.
	BitAnd(func(BitAndBuilder)) schema.QueryAppender
	// BoolOr builds a BOOL_OR aggregate expression using a builder callback.
	BoolOr(func(BoolOrBuilder)) schema.QueryAppender
	// BoolAnd builds a BOOL_AND aggregate expression using a builder callback.
	BoolAnd(func(BoolAndBuilder)) schema.QueryAppender
	// StdDev builds a STDDEV aggregate expression using a builder callback.
	StdDev(func(StdDevBuilder)) schema.QueryAppender
	// Variance builds a VARIANCE aggregate expression using a builder callback.
	Variance(func(VarianceBuilder)) schema.QueryAppender

	// ========== Window Functions ==========

	// RowNumber builds a ROW_NUMBER window function expression.
	RowNumber(func(RowNumberBuilder)) schema.QueryAppender
	// Rank builds a RANK window function expression.
	Rank(func(RankBuilder)) schema.QueryAppender
	// DenseRank builds a DENSE_RANK window function expression.
	DenseRank(func(DenseRankBuilder)) schema.QueryAppender
	// PercentRank builds a PERCENT_RANK window function expression.
	PercentRank(func(PercentRankBuilder)) schema.QueryAppender
	// CumeDist builds a CUME_DIST window function expression.
	CumeDist(func(CumeDistBuilder)) schema.QueryAppender
	// NTile builds an NTILE window function expression.
	NTile(func(NTileBuilder)) schema.QueryAppender
	// Lag builds a LAG window function expression.
	Lag(func(LagBuilder)) schema.QueryAppender
	// Lead builds a LEAD window function expression.
	Lead(func(LeadBuilder)) schema.QueryAppender
	// FirstValue builds a FIRST_VALUE window function expression.
	FirstValue(func(FirstValueBuilder)) schema.QueryAppender
	// LastValue builds a LAST_VALUE window function expression.
	LastValue(func(LastValueBuilder)) schema.QueryAppender
	// NthValue builds an NTH_VALUE window function expression.
	NthValue(func(NthValueBuilder)) schema.QueryAppender
	// WinCount builds a COUNT window function expression.
	WinCount(func(WindowCountBuilder)) schema.QueryAppender
	// WinSum builds a SUM window function expression.
	WinSum(func(WindowSumBuilder)) schema.QueryAppender
	// WinAvg builds an AVG window function expression.
	WinAvg(func(WindowAvgBuilder)) schema.QueryAppender
	// WinMin builds a MIN window function expression.
	WinMin(func(WindowMinBuilder)) schema.QueryAppender
	// WinMax builds a MAX window function expression.
	WinMax(func(WindowMaxBuilder)) schema.QueryAppender
	// WinStringAgg builds a STRING_AGG window function expression.
	WinStringAgg(func(WindowStringAggBuilder)) schema.QueryAppender
	// WinArrayAgg builds an ARRAY_AGG window function expression.
	WinArrayAgg(func(WindowArrayAggBuilder)) schema.QueryAppender
	// WinStdDev builds a STDDEV window function expression.
	WinStdDev(func(WindowStdDevBuilder)) schema.QueryAppender
	// WinVariance builds a VARIANCE window function expression.
	WinVariance(func(WindowVarianceBuilder)) schema.QueryAppender
	// WinJsonObjectAgg builds a JSON_OBJECT_AGG window function expression.
	WinJsonObjectAgg(func(WindowJSONObjectAggBuilder)) schema.QueryAppender
	// WinJsonArrayAgg builds a JSON_ARRAY_AGG window function expression.
	WinJsonArrayAgg(func(WindowJSONArrayAggBuilder)) schema.QueryAppender
	// WinBitOr builds a BIT_OR window function expression.
	WinBitOr(func(WindowBitOrBuilder)) schema.QueryAppender
	// WinBitAnd builds a BIT_AND window function expression.
	WinBitAnd(func(WindowBitAndBuilder)) schema.QueryAppender
	// WinBoolOr builds a BOOL_OR window function expression.
	WinBoolOr(func(WindowBoolOrBuilder)) schema.QueryAppender
	// WinBoolAnd builds a BOOL_AND window function expression.
	WinBoolAnd(func(WindowBoolAndBuilder)) schema.QueryAppender

	// ========== String Functions ==========

	// Concat concatenates strings.
	Concat(args ...any) schema.QueryAppender
	// ConcatWithSep concatenates strings with a separator.
	ConcatWithSep(separator any, args ...any) schema.QueryAppender
	// SubString extracts a substring from a string.
	// start: starting position (1-based), length: optional length
	SubString(expr, start any, length ...any) schema.QueryAppender
	// Upper converts string to uppercase.
	Upper(expr any) schema.QueryAppender
	// Lower converts string to lowercase.
	Lower(expr any) schema.QueryAppender
	// Trim removes leading and trailing whitespace.
	Trim(expr any) schema.QueryAppender
	// TrimLeft removes leading whitespace.
	TrimLeft(expr any) schema.QueryAppender
	// TrimRight removes trailing whitespace.
	TrimRight(expr any) schema.QueryAppender
	// Length returns the length of a string.
	Length(expr any) schema.QueryAppender
	// CharLength returns the character length of a string.
	CharLength(expr any) schema.QueryAppender
	// Position finds the position of substring in string (1-based, 0 if not found).
	Position(substring, str any) schema.QueryAppender
	// Left returns the leftmost n characters.
	Left(expr, length any) schema.QueryAppender
	// Right returns the rightmost n characters.
	Right(expr, length any) schema.QueryAppender
	// Repeat repeats a string n times.
	Repeat(expr, count any) schema.QueryAppender
	// Replace replaces all occurrences of substring with replacement.
	Replace(expr, search, replacement any) schema.QueryAppender
	// Contains checks if a string contains a substring (case-sensitive).
	Contains(expr, substr any) schema.QueryAppender
	// StartsWith checks if a string starts with a prefix (case-sensitive).
	StartsWith(expr, prefix any) schema.QueryAppender
	// EndsWith checks if a string ends with a suffix (case-sensitive).
	EndsWith(expr, suffix any) schema.QueryAppender
	// ContainsIgnoreCase checks if a string contains a substring (case-insensitive).
	ContainsIgnoreCase(expr, substr any) schema.QueryAppender
	// StartsWithIgnoreCase checks if a string starts with a prefix (case-insensitive).
	StartsWithIgnoreCase(expr, prefix any) schema.QueryAppender
	// EndsWithIgnoreCase checks if a string ends with a suffix (case-insensitive).
	EndsWithIgnoreCase(expr, suffix any) schema.QueryAppender
	// Reverse reverses a string.
	Reverse(expr any) schema.QueryAppender

	// ========== Date and Time Functions ==========

	// CurrentDate returns the current date.
	CurrentDate() schema.QueryAppender
	// CurrentTime returns the current time.
	CurrentTime() schema.QueryAppender
	// CurrentTimestamp returns the current timestamp.
	CurrentTimestamp() schema.QueryAppender
	// Now returns the current timestamp (alias for CurrentTimestamp).
	Now() schema.QueryAppender
	// ExtractYear extracts the year from a date/timestamp.
	ExtractYear(expr any) schema.QueryAppender
	// ExtractMonth extracts the month from a date/timestamp.
	ExtractMonth(expr any) schema.QueryAppender
	// ExtractDay extracts the day from a date/timestamp.
	ExtractDay(expr any) schema.QueryAppender
	// ExtractHour extracts the hour from a timestamp.
	ExtractHour(expr any) schema.QueryAppender
	// ExtractMinute extracts the minute from a timestamp.
	ExtractMinute(expr any) schema.QueryAppender
	// ExtractSecond extracts the second from a timestamp.
	ExtractSecond(expr any) schema.QueryAppender
	// DateTrunc truncates date/timestamp to specified precision.
	DateTrunc(unit DateTimeUnit, expr any) schema.QueryAppender
	// DateAdd adds interval to date/timestamp.
	DateAdd(expr, interval any, unit DateTimeUnit) schema.QueryAppender
	// DateSubtract subtracts interval from date/timestamp.
	DateSubtract(expr, interval any, unit DateTimeUnit) schema.QueryAppender
	// DateDiff returns the difference between two dates in specified unit.
	DateDiff(start, end any, unit DateTimeUnit) schema.QueryAppender
	// Age returns the age (interval) between two timestamps.
	Age(start, end any) schema.QueryAppender

	// ========== Math Functions ==========

	// Abs returns the absolute value.
	Abs(expr any) schema.QueryAppender
	// Ceil returns the smallest integer greater than or equal to the value.
	Ceil(expr any) schema.QueryAppender
	// Floor returns the largest integer less than or equal to the value.
	Floor(expr any) schema.QueryAppender
	// Round rounds to the nearest integer or specified decimal places.
	Round(expr any, precision ...any) schema.QueryAppender
	// Trunc truncates to integer or specified decimal places.
	Trunc(expr any, precision ...any) schema.QueryAppender
	// Power returns base raised to the power of exponent.
	Power(base, exponent any) schema.QueryAppender
	// Sqrt returns the square root.
	Sqrt(expr any) schema.QueryAppender
	// Exp returns e raised to the power of the argument.
	Exp(expr any) schema.QueryAppender
	// Ln returns the natural logarithm.
	Ln(expr any) schema.QueryAppender
	// Log returns the logarithm with specified base (default base 10).
	Log(expr any, base ...any) schema.QueryAppender
	// Sin returns the sine.
	Sin(expr any) schema.QueryAppender
	// Cos returns the cosine.
	Cos(expr any) schema.QueryAppender
	// Tan returns the tangent.
	Tan(expr any) schema.QueryAppender
	// Asin returns the arcsine.
	Asin(expr any) schema.QueryAppender
	// Acos returns the arccosine.
	Acos(expr any) schema.QueryAppender
	// Atan returns the arctangent.
	Atan(expr any) schema.QueryAppender
	// Pi returns the value of π.
	Pi() schema.QueryAppender
	// Random returns a random value between 0 and 1.
	Random() schema.QueryAppender
	// Sign returns the sign of a number (-1, 0, or 1).
	Sign(expr any) schema.QueryAppender
	// Mod returns the remainder of division.
	Mod(dividend, divisor any) schema.QueryAppender
	// Greatest returns the greatest value among arguments.
	Greatest(args ...any) schema.QueryAppender
	// Least returns the least value among arguments.
	Least(args ...any) schema.QueryAppender

	// ========== Conditional Functions ==========

	// Coalesce returns the first non-null value.
	Coalesce(args ...any) schema.QueryAppender
	// NullIf returns null if the two arguments are equal, otherwise returns the first argument.
	NullIf(expr1, expr2 any) schema.QueryAppender
	// IfNull returns the second argument if the first is null, otherwise returns the first.
	IfNull(expr, defaultValue any) schema.QueryAppender

	// ========== Type Conversion Functions ==========

	// ToString converts expression to string.
	ToString(expr any) schema.QueryAppender
	// ToInteger converts expression to integer.
	ToInteger(expr any) schema.QueryAppender
	// ToDecimal converts expression to decimal with optional precision and scale.
	ToDecimal(expr any, precision ...any) schema.QueryAppender
	// ToFloat converts expression to float.
	ToFloat(expr any) schema.QueryAppender
	// ToBool converts expression to boolean.
	ToBool(expr any) schema.QueryAppender
	// ToDate converts expression to date.
	ToDate(expr any, format ...any) schema.QueryAppender
	// ToTime converts expression to time.
	ToTime(expr any, format ...any) schema.QueryAppender
	// ToTimestamp converts expression to timestamp.
	ToTimestamp(expr any, format ...any) schema.QueryAppender
	// ToJSON converts expression to JSON.
	ToJSON(expr any) schema.QueryAppender

	// ========== JSON Functions ==========

	// JSONExtract extracts value from JSON at specified path.
	JSONExtract(json, path any) schema.QueryAppender
	// JSONUnquote removes quotes from JSON string.
	JSONUnquote(expr any) schema.QueryAppender
	// JSONArray creates a JSON array from arguments.
	JSONArray(args ...any) schema.QueryAppender
	// JSONObject creates a JSON object from key-value pairs.
	JSONObject(keyValues ...any) schema.QueryAppender
	// JSONContains checks if JSON contains a value.
	JSONContains(json, value any) schema.QueryAppender
	// JSONContainsPath checks if JSON contains a path.
	JSONContainsPath(json, path any) schema.QueryAppender
	// JSONKeys returns the keys of a JSON object.
	JSONKeys(json any, path ...any) schema.QueryAppender
	// JSONLength returns the length of a JSON array or object.
	JSONLength(json any, path ...any) schema.QueryAppender
	// JSONType returns the type of JSON value.
	JSONType(json any, path ...any) schema.QueryAppender
	// JSONValid checks if a string is valid JSON.
	JSONValid(expr any) schema.QueryAppender
	// JSONSet sets value at path in JSON (insert or update).
	JSONSet(json, path, value any) schema.QueryAppender
	// JSONInsert inserts value at path only if path doesn't exist.
	JSONInsert(json, path, value any) schema.QueryAppender
	// JSONReplace replaces value at path only if path exists.
	JSONReplace(json, path, value any) schema.QueryAppender
	// JSONArrayAppend appends value to JSON array at specified path.
	JSONArrayAppend(json, path, value any) schema.QueryAppender

	// ========== Utility Functions ==========

	// Decode implements DECODE function (Oracle-style case expression).
	// Usage: Decode(expr, search1, result1, search2, result2, ..., defaultResult)
	Decode(args ...any) schema.QueryAppender
}

// SelectQueryExecutor is an interface that defines the methods for executing SELECT queries.
// It extends QueryExecutor with additional methods specific to SELECT operations.
type SelectQueryExecutor interface {
	QueryExecutor
	// Rows returns the result as a sql.Rows.
	Rows(ctx context.Context) (*sql.Rows, error)
	// ScanAndCount scans the result into a slice of any type and returns the count of the result.
	ScanAndCount(ctx context.Context, dest ...any) (int64, error)
	// Count returns the count of the result.
	Count(ctx context.Context) (int64, error)
	// Exists returns true if the result exists.
	Exists(ctx context.Context) (bool, error)
}

// SelectQuery is an interface that defines the methods for building and executing SELECT queries.
// It provides a fluent Api for constructing complex database queries with support for joins, conditions, ordering, and more.
type SelectQuery interface {
	QueryBuilder
	SelectQueryExecutor
	DBAccessor
	CTE[SelectQuery]
	Selectable[SelectQuery]
	TableSource[SelectQuery]
	JoinOperations[SelectQuery]
	Filterable[SelectQuery]
	Orderable[SelectQuery]
	Limitable[SelectQuery]
	Applier[SelectQuery]

	// SelectAs selects a column with an alias.
	SelectAs(column, alias string) SelectQuery
	// SelectModelColumns selects the columns of a model.
	// By default, all columns of the model are selected if no select-related methods are called.
	SelectModelColumns() SelectQuery
	// SelectModelPKs selects the primary keys of a model.
	SelectModelPKs() SelectQuery
	// SelectExpr selects a column with an expression.
	SelectExpr(builder func(ExprBuilder) any, alias ...string) SelectQuery
	// Distinct returns a distinct query.
	Distinct() SelectQuery
	// DistinctOnColumns returns a distinct query on columns.
	DistinctOnColumns(columns ...string) SelectQuery
	// DistinctOnExpr returns a distinct query on an expression.
	DistinctOnExpr(builder func(ExprBuilder) any) SelectQuery
	// JoinRelations applies RelationSpec configurations to perform JOIN operations with automatic column resolution.
	// It provides a declarative way to join related models with minimal configuration.
	JoinRelations(specs ...*RelationSpec) SelectQuery
	// Relation joins a relation.
	Relation(name string, apply ...func(query SelectQuery)) SelectQuery
	// GroupBy groups the query by a column.
	GroupBy(columns ...string) SelectQuery
	// GroupByExpr groups the query by an expression.
	GroupByExpr(func(ExprBuilder) any) SelectQuery
	// Having adds a having clause to the query.
	Having(func(ConditionBuilder)) SelectQuery
	// Offset adds an offset to the query.
	Offset(offset int) SelectQuery
	// Paginate paginates the query.
	Paginate(pageable page.Pageable) SelectQuery
	// ForShare adds a for share lock to the query.
	ForShare(tables ...string) SelectQuery
	// ForShareNoWait adds a for share no wait lock to the query.
	ForShareNoWait(tables ...string) SelectQuery
	// ForShareSkipLocked adds a for share skip locked lock to the query.
	ForShareSkipLocked(tables ...string) SelectQuery
	// ForUpdate adds a for update lock to the query.
	ForUpdate(tables ...string) SelectQuery
	// ForUpdateNoWait adds a for update no wait lock to the query.
	ForUpdateNoWait(tables ...string) SelectQuery
	// ForUpdateSkipLocked adds a for update skip locked lock to the query.
	ForUpdateSkipLocked(tables ...string) SelectQuery
	// Union combines the result of this query with another query.
	Union(func(query SelectQuery)) SelectQuery
	// UnionAll combines the result of this query with another query, including duplicates.
	UnionAll(func(query SelectQuery)) SelectQuery
	// Intersect returns only rows that exist in both this query and another query.
	Intersect(func(query SelectQuery)) SelectQuery
	// IntersectAll returns only rows that exist in both queries, including duplicates.
	IntersectAll(func(query SelectQuery)) SelectQuery
	// Except returns rows that exist in this query but not in another query.
	Except(func(query SelectQuery)) SelectQuery
	// ExceptAll returns rows that exist in this query but not in another query, including duplicates.
	ExceptAll(func(query SelectQuery)) SelectQuery
}

// RawQuery is an interface that defines the methods for executing raw SQL queries.
// It allows direct SQL execution with parameter binding for cases where the query builder is insufficient.
type RawQuery interface {
	QueryExecutor
}

// InsertQuery is an interface that defines the methods for building and executing INSERT queries.
// It supports conflict resolution, column selection, and expression-based values.
type InsertQuery interface {
	QueryBuilder
	QueryExecutor
	DBAccessor
	CTE[InsertQuery]
	TableSource[InsertQuery]
	Selectable[InsertQuery]
	ColumnUpdatable[InsertQuery]
	Returnable[InsertQuery]
	Applier[InsertQuery]

	// OnConflict configures conflict handling (UPSERT) using a builder.
	OnConflict(func(ConflictBuilder)) InsertQuery
}

// UpdateQuery is an interface that defines the methods for building and executing UPDATE queries.
// It supports FROM clause for joining additional tables, conditions, column updates, and bulk operations.
// Note: UPDATE does not inherit JoinOperations directly; use From methods for PostgreSQL-style FROM clause.
type UpdateQuery interface {
	QueryBuilder
	QueryExecutor
	DBAccessor
	CTE[UpdateQuery]
	TableSource[UpdateQuery]
	Selectable[UpdateQuery]
	Filterable[UpdateQuery]
	Orderable[UpdateQuery]
	Limitable[UpdateQuery]
	ColumnUpdatable[UpdateQuery]
	Returnable[UpdateQuery]
	Applier[UpdateQuery]

	// Set sets a column to a specific value (alias for Column).
	Set(name string, value any) UpdateQuery
	// SetExpr sets a column using an expression builder (alias for ColumnExpr).
	SetExpr(name string, builder func(ExprBuilder) any) UpdateQuery
	// OmitZero adds an omit zero clause to the query.
	OmitZero() UpdateQuery
	// Bulk adds a bulk clause to the query.
	Bulk() UpdateQuery
}

// DeleteQuery is an interface that defines the methods for building and executing DELETE queries.
// It supports USING clause for joining additional tables, conditions, ordering, limits, and soft delete operations.
type DeleteQuery interface {
	QueryBuilder
	QueryExecutor
	DBAccessor
	CTE[DeleteQuery]
	TableSource[DeleteQuery]
	Filterable[DeleteQuery]
	Orderable[DeleteQuery]
	Limitable[DeleteQuery]
	Returnable[DeleteQuery]
	Applier[DeleteQuery]

	// ForceDelete adds a force delete clause to the query.
	ForceDelete() DeleteQuery
}

// MergeQuery is an interface that defines the methods for building and executing MERGE queries.
// It supports complex merge operations with conditional actions based on match/no-match scenarios.
type MergeQuery interface {
	QueryBuilder
	QueryExecutor
	DBAccessor
	CTE[MergeQuery]
	TableSource[MergeQuery]
	Returnable[MergeQuery]
	Applier[MergeQuery]

	// Using specifies a model as the source for the merge operation.
	Using(model any, alias ...string) MergeQuery
	// UsingTable specifies the source table for the merge operation.
	UsingTable(table string, alias ...string) MergeQuery
	// UsingExpr specifies a expression as the source for the merge operation.
	UsingExpr(builder func(ExprBuilder) any, alias ...string) MergeQuery
	// UsingSubQuery specifies a subquery as the source for the merge operation.
	UsingSubQuery(builder func(SelectQuery), alias ...string) MergeQuery

	// On specifies the merge condition that determines matches between target and source.
	On(func(ConditionBuilder)) MergeQuery

	// WhenMatched starts a conditional action block for when records match.
	WhenMatched(builder ...func(ConditionBuilder)) MergeWhenBuilder
	// WhenNotMatched starts a conditional action block for when records don't match in target.
	WhenNotMatched(builder ...func(ConditionBuilder)) MergeWhenBuilder
	// WhenNotMatchedByTarget starts a conditional action block for when records don't match in target (explicit form).
	WhenNotMatchedByTarget(builder ...func(ConditionBuilder)) MergeWhenBuilder
	// WhenNotMatchedBySource starts a conditional action block for when records don't match in source.
	WhenNotMatchedBySource(builder ...func(ConditionBuilder)) MergeWhenBuilder
}

// DB is an interface that defines the methods for database operations.
// It provides factory methods for creating different types of queries and supports transactions.
type DB interface {
	// NewSelect creates a new select query.
	NewSelect() SelectQuery
	// NewInsert creates a new insert.
	NewInsert() InsertQuery
	// NewUpdate creates a new update.
	NewUpdate() UpdateQuery
	// NewDelete creates a new delete.
	NewDelete() DeleteQuery
	// NewMerge creates a new merge query.
	NewMerge() MergeQuery
	// NewRaw creates a new raw query.
	NewRaw(query string, args ...any) RawQuery
	// RunInTX runs a transaction.
	RunInTX(ctx context.Context, fn func(ctx context.Context, tx DB) error) error
	// RunInReadOnlyTX runs a read-only transaction.
	RunInReadOnlyTX(ctx context.Context, fn func(ctx context.Context, tx DB) error) error
	// WithNamedArg returns a new DB with the named arg.
	WithNamedArg(name string, value any) DB
	// ModelPKs returns the primary keys of a model.
	ModelPKs(model any) (map[string]any, error)
	// ModelPKFields returns the primary key fields of a model.
	ModelPKFields(model any) []*PKField
	// TableOf returns the table information for a model.
	TableOf(model any) *schema.Table
}
