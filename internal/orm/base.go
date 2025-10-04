package orm

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

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
	// Model selects a model.
	Model(model any) T
	// ModelTable selects a table.
	ModelTable(name string, alias ...string) T
	// Table selects a table.
	Table(name string, alias ...string) T
	// TableExpr selects a table with an expression.
	TableExpr(alias string, builder func(ExprBuilder) any) T
	// TableSubQuery selects a subquery.
	TableSubQuery(alias string, builder func(query SelectQuery)) T
}

// JoinOperations is an interface that defines the methods for joining tables in queries.
// It supports various join types including INNER, LEFT, and RIGHT joins with different source types.
type JoinOperations[T QueryExecutor] interface {
	// Join joins a table.
	Join(model any, builder func(ConditionBuilder), alias ...string) T
	// JoinTable joins a table.
	JoinTable(name string, builder func(ConditionBuilder), alias ...string) T
	// JoinSubQuery joins a subquery.
	JoinSubQuery(alias string, sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder)) T
	// JoinExpr joins an expression.
	JoinExpr(alias string, eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder)) T
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
	// Column sets a column to a value.
	Column(name string, value any) T
	// ColumnExpr sets a column to an expression.
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

// DialectExpr defines database-specific expression builders for cross-database compatibility.
// It allows users to define custom expressions that work across different database engines
// by providing database-specific implementations.
type DialectExpr struct {
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

// DialectActions defines database-specific callbacks for running side-effect logic
// without returning a SQL expression.
type DialectActions struct {
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

// DialectActionsErr defines database-specific callbacks that may return an error.
type DialectActionsErr struct {
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

// DialectFunc represents a zero-argument callback that returns a query fragment buffer.
type DialectFunc func() ([]byte, error)

// DialectFuncs defines database-specific callbacks that produce query fragments.
type DialectFuncs struct {
	// Oracle callback for Oracle database.
	Oracle DialectFunc
	// SQL Server callback for SQL Server database.
	SQLServer DialectFunc
	// Postgres callback for PostgreSQL database.
	Postgres DialectFunc
	// MySQL callback for MySQL database.
	MySQL DialectFunc
	// SQLite callback for SQLite database.
	SQLite DialectFunc
	// Default callback used when database-specific callback is not available.
	Default DialectFunc
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
// It offers a fluent API for constructing complex SQL expressions including aggregates, functions, and conditional logic.
type ExprBuilder interface {
	// Column builds a column expression with proper alias handling.
	Column(column string) schema.QueryAppender
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
	// Expr creates an expression builder for complex SQL logic.
	Expr(expr string, args ...any) schema.QueryAppender
	// Exprs creates an expression builder for complex SQL logic.
	Exprs(exprs ...any) schema.QueryAppender
	// ExprsWS creates an expression builder for complex SQL logic with a separator.
	ExprsWS(separator string, exprs ...any) schema.QueryAppender
	// ExprByDialect creates a cross-database compatible expression.
	// It selects the appropriate expression builder based on the current database dialect.
	ExprByDialect(expr DialectExpr) schema.QueryAppender
	// RunDialect selects the appropriate callback based on the current database dialect.
	RunDialect(actions DialectActions)
	// RunDialectErr runs dialect-specific callbacks and returns any error encountered.
	RunDialectErr(actions DialectActionsErr) error
	// RunDialectFunc selects the appropriate query fragment builder based on the current database dialect.
	RunDialectFunc(funcs DialectFuncs) ([]byte, error)

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
	// Ntile builds an NTILE window function expression.
	Ntile(func(NtileBuilder)) schema.QueryAppender
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
	// WCount builds a COUNT window function expression.
	WCount(func(WindowCountBuilder)) schema.QueryAppender
	// WSum builds a SUM window function expression.
	WSum(func(WindowSumBuilder)) schema.QueryAppender
	// WAvg builds an AVG window function expression.
	WAvg(func(WindowAvgBuilder)) schema.QueryAppender
	// WMin builds a MIN window function expression.
	WMin(func(WindowMinBuilder)) schema.QueryAppender
	// WMax builds a MAX window function expression.
	WMax(func(WindowMaxBuilder)) schema.QueryAppender
	// WStringAgg builds a STRING_AGG window function expression.
	WStringAgg(func(WindowStringAggBuilder)) schema.QueryAppender
	// WArrayAgg builds an ARRAY_AGG window function expression.
	WArrayAgg(func(WindowArrayAggBuilder)) schema.QueryAppender
	// WStdDev builds a STDDEV window function expression.
	WStdDev(func(WindowStdDevBuilder)) schema.QueryAppender
	// WVariance builds a VARIANCE window function expression.
	WVariance(func(WindowVarianceBuilder)) schema.QueryAppender
	// WJSONObjectAgg builds a JSON_OBJECT_AGG window function expression.
	WJSONObjectAgg(func(WindowJSONObjectAggBuilder)) schema.QueryAppender
	// WJSONArrayAgg builds a JSON_ARRAY_AGG window function expression.
	WJSONArrayAgg(func(WindowJSONArrayAggBuilder)) schema.QueryAppender
	// WBitOr builds a BIT_OR window function expression.
	WBitOr(func(WindowBitOrBuilder)) schema.QueryAppender
	// WBitAnd builds a BIT_AND window function expression.
	WBitAnd(func(WindowBitAndBuilder)) schema.QueryAppender
	// WBoolOr builds a BOOL_OR window function expression.
	WBoolOr(func(WindowBoolOrBuilder)) schema.QueryAppender
	// WBoolAnd builds a BOOL_AND window function expression.
	WBoolAnd(func(WindowBoolAndBuilder)) schema.QueryAppender

	// ========== String Functions ==========

	// Concat concatenates strings.
	Concat(args ...any) schema.QueryAppender
	// ConcatWS concatenates strings with a separator.
	ConcatWS(separator string, args ...any) schema.QueryAppender
	// SubString extracts a substring from a string.
	// start: starting position (1-based), length: optional length
	SubString(expr any, start int, length ...int) schema.QueryAppender
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
	Left(expr any, length int) schema.QueryAppender
	// Right returns the rightmost n characters.
	Right(expr any, length int) schema.QueryAppender
	// Repeat repeats a string n times.
	Repeat(expr any, count int) schema.QueryAppender
	// Replace replaces all occurrences of substring with replacement.
	Replace(expr, search, replacement any) schema.QueryAppender
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
	// precision: 'year', 'month', 'day', 'hour', 'minute', 'second'
	DateTrunc(precision string, expr any) schema.QueryAppender
	// DateAdd adds interval to date/timestamp.
	// unit: 'year', 'month', 'day', 'hour', 'minute', 'second'
	DateAdd(expr any, interval int, unit string) schema.QueryAppender
	// DateSubtract subtracts interval from date/timestamp.
	// unit: 'year', 'month', 'day', 'hour', 'minute', 'second'
	DateSubtract(expr any, interval int, unit string) schema.QueryAppender
	// DateDiff returns the difference between two dates in specified unit.
	// unit: 'year', 'month', 'day', 'hour', 'minute', 'second'
	DateDiff(start, end any, unit string) schema.QueryAppender
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
	Round(expr any, precision ...int) schema.QueryAppender
	// Trunc truncates to integer or specified decimal places.
	Trunc(expr any, precision ...int) schema.QueryAppender
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
	ToDecimal(expr any, precision ...int) schema.QueryAppender
	// ToFloat converts expression to float.
	ToFloat(expr any) schema.QueryAppender
	// ToBool converts expression to boolean.
	ToBool(expr any) schema.QueryAppender
	// ToDate converts expression to date.
	ToDate(expr any, format ...string) schema.QueryAppender
	// ToTime converts expression to time.
	ToTime(expr any, format ...string) schema.QueryAppender
	// ToTimestamp converts expression to timestamp.
	ToTimestamp(expr any, format ...string) schema.QueryAppender
	// ToJSON converts expression to JSON.
	ToJSON(expr any) schema.QueryAppender

	// ========== JSON Functions ==========

	// JSONExtract extracts value from JSON at specified path.
	JSONExtract(json any, path string) schema.QueryAppender
	// JSONUnquote removes quotes from JSON string.
	JSONUnquote(expr any) schema.QueryAppender
	// JSONArray creates a JSON array from arguments.
	JSONArray(args ...any) schema.QueryAppender
	// JSONObject creates a JSON object from key-value pairs.
	JSONObject(keyValues ...any) schema.QueryAppender
	// JSONContains checks if JSON contains a value.
	JSONContains(json, value any) schema.QueryAppender
	// JSONContainsPath checks if JSON contains a path.
	JSONContainsPath(json any, path string) schema.QueryAppender
	// JSONKeys returns the keys of a JSON object.
	JSONKeys(json any, path ...string) schema.QueryAppender
	// JSONLength returns the length of a JSON array or object.
	JSONLength(json any, path ...string) schema.QueryAppender
	// JSONType returns the type of JSON value.
	JSONType(json any, path ...string) schema.QueryAppender
	// JSONValid checks if a string is valid JSON.
	JSONValid(expr any) schema.QueryAppender
	// JSONSet sets value at path, creates if not exists, replaces if exists.
	JSONSet(json any, path string, value any) schema.QueryAppender
	// JSONInsert inserts value at path only if path doesn't exist.
	JSONInsert(json any, path string, value any) schema.QueryAppender
	// JSONReplace replaces value at path only if path exists.
	JSONReplace(json any, path string, value any) schema.QueryAppender
	// JSONArrayAppend appends value to JSON array at specified path.
	JSONArrayAppend(json any, path string, value any) schema.QueryAppender

	// ========== Utility Functions ==========

	// Decode implements DECODE function (Oracle-style case expression).
	// Usage: Decode(expr, search1, result1, search2, result2, ..., defaultResult)
	Decode(args ...any) schema.QueryAppender
}
