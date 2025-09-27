package orm

import (
	"context"
	"database/sql"

	"github.com/ilxqx/vef-framework-go/page"
	"github.com/uptrace/bun/schema"
)

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
// It provides a fluent API for constructing complex database queries with support for joins, conditions, ordering, and more.
type SelectQuery interface {
	QueryBuilder
	SelectQueryExecutor
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
	// LeftJoin joins a table.
	LeftJoin(model any, builder func(ConditionBuilder), alias ...string) SelectQuery
	// LeftJoinTable joins a table.
	LeftJoinTable(name string, builder func(ConditionBuilder), alias ...string) SelectQuery
	// LeftJoinSubQuery joins a subquery.
	LeftJoinSubQuery(alias string, sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder)) SelectQuery
	// LeftJoinExpr joins an expression.
	LeftJoinExpr(alias string, eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder)) SelectQuery
	// RightJoin joins a table.
	RightJoin(model any, builder func(ConditionBuilder), alias ...string) SelectQuery
	// RightJoinTable joins a table.
	RightJoinTable(name string, builder func(ConditionBuilder), alias ...string) SelectQuery
	// RightJoinSubQuery joins a subquery.
	RightJoinSubQuery(alias string, sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder)) SelectQuery
	// RightJoinExpr joins an expression.
	RightJoinExpr(alias string, eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder)) SelectQuery
	// ModelRelations joins a model relation.
	ModelRelations(relations ...ModelRelation) SelectQuery
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
// It supports joins, conditions, column updates, and bulk operations.
type UpdateQuery interface {
	QueryBuilder
	QueryExecutor
	CTE[UpdateQuery]
	TableSource[UpdateQuery]
	JoinOperations[UpdateQuery]
	Selectable[UpdateQuery]
	Filterable[UpdateQuery]
	Orderable[UpdateQuery]
	Limitable[UpdateQuery]
	ColumnUpdatable[UpdateQuery]
	Returnable[UpdateQuery]
	Applier[UpdateQuery]

	// Set sets a column to a value.
	Set(name string, value any) UpdateQuery
	// SetExpr sets a column to an expression.
	SetExpr(name string, builder func(ExprBuilder) any) UpdateQuery
	// OmitZero adds an omit zero clause to the query.
	OmitZero() UpdateQuery
	// Bulk adds a bulk clause to the query.
	Bulk() UpdateQuery
}

// DeleteQuery is an interface that defines the methods for building and executing DELETE queries.
// It supports conditions, ordering, limits, and soft delete operations.
type DeleteQuery interface {
	QueryBuilder
	QueryExecutor
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
	CTE[MergeQuery]
	TableSource[MergeQuery]
	Returnable[MergeQuery]
	Applier[MergeQuery]

	// Using specifies the source table or data for the merge operation.
	Using(source string, alias ...string) MergeQuery
	// UsingModel specifies a model as the source for the merge operation.
	UsingModel(model any) MergeQuery
	// UsingExpr specifies a expression as the source for the merge operation.
	UsingExpr(alias string, builder func(ExprBuilder) any) MergeQuery
	// UsingSubQuery specifies a subquery as the source for the merge operation.
	UsingSubQuery(alias string, builder func(SelectQuery)) MergeQuery
	// UsingValues specifies literal values as the source for the merge operation.
	UsingValues(model any, columns ...string) MergeQuery

	// On specifies the merge condition that determines matches between target and source.
	On(func(ConditionBuilder)) MergeQuery

	// WhenMatched starts a conditional action block for when records match.
	WhenMatched(builder ...func(ConditionBuilder)) MergeWhenBuilder
	// WhenNotMatched starts a conditional action block for when records don't match in target.
	WhenNotMatched(builder ...func(ConditionBuilder)) MergeWhenBuilder
	// WhenNotMatchedBySource starts a conditional action block for when records don't match in source.
	WhenNotMatchedBySource(builder ...func(ConditionBuilder)) MergeWhenBuilder
}

// Db is an interface that defines the methods for database operations.
// It provides factory methods for creating different types of queries and supports transactions.
type Db interface {
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
	// RunInTx runs a transaction.
	RunInTx(ctx context.Context, fn func(ctx context.Context, tx Db) error) error
	// RunInReadOnlyTx runs a read-only transaction.
	RunInReadOnlyTx(ctx context.Context, fn func(ctx context.Context, tx Db) error) error
	// WithNamedArg returns a new Db with the named arg.
	WithNamedArg(name string, value any) Db
	// ModelPKs returns the primary keys of a model.
	ModelPKs(model any) (map[string]any, error)
	// ModelPKFields returns the primary key fields of a model.
	ModelPKFields(model any) []*PKField
	// Schema returns the schema of a table.
	Schema(model any) *schema.Table
}
