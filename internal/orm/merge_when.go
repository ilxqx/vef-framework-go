package orm

import (
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// MergeWhenBuilder is an interface for defining actions in MERGE WHEN clauses.
// It provides methods to specify what action to take when merge conditions are met.
type MergeWhenBuilder interface {
	// ThenUpdate specifies an UPDATE action for the WHEN clause.
	ThenUpdate() MergeUpdateBuilder
	// ThenInsert specifies an INSERT action for the WHEN clause.
	ThenInsert() MergeInsertBuilder
	// ThenDelete specifies a DELETE action for the WHEN clause.
	ThenDelete() MergeQuery
	// ThenDoNothing specifies no action for the WHEN clause.
	ThenDoNothing() MergeQuery
}

// MergeUpdateBuilder is an interface for configuring UPDATE actions in MERGE queries.
// It allows setting column values and expressions for update operations.
type MergeUpdateBuilder interface {
	// Set sets a column to a specific value in the update action.
	Set(column string, value any) MergeUpdateBuilder
	// SetExpr sets a column using an expression in the update action.
	SetExpr(column string, builder func(ExprBuilder) any) MergeUpdateBuilder
	// SetColumns sets multiple columns from the source in the update action.
	SetColumns(columns ...string) MergeUpdateBuilder
	// SetExcept sets all columns except the specified ones from the source.
	SetExcept(columns ...string) MergeUpdateBuilder
	// End returns to the main merge query builder.
	End() MergeQuery
}

// MergeInsertBuilder is an interface for configuring INSERT actions in MERGE queries.
// It allows specifying which columns and values to insert.
type MergeInsertBuilder interface {
	// Value sets a column to a specific value in the insert action.
	Value(column string, value any) MergeInsertBuilder
	// ValueExpr sets a column using an expression in the insert action.
	ValueExpr(column string, builder func(ExprBuilder) any) MergeInsertBuilder
	// Values sets multiple columns from the source in the insert action.
	Values(columns ...string) MergeInsertBuilder
	// ValuesExcept sets all columns except the specified ones from the source.
	ValuesExcept(columns ...string) MergeInsertBuilder
	// End returns to the main merge query builder.
	End() MergeQuery
}

// mergeWhenBuilder implements MergeWhenBuilder interface.
type mergeWhenBuilder struct {
	query      *BunMergeQuery
	whenClause string
	condition  schema.QueryAppender
}

// newMergeWhenBuilder creates a new MergeWhenBuilder instance.
func newMergeWhenBuilder(query *BunMergeQuery, whenClause string, builder ...func(ConditionBuilder)) *mergeWhenBuilder {
	wb := &mergeWhenBuilder{
		query:      query,
		whenClause: whenClause,
	}

	// Build conditions if provided
	for len(builder) > 0 {
		wb.condition = query.BuildCondition(builder[0])
	}

	return wb
}

func (b *mergeWhenBuilder) ThenUpdate() MergeUpdateBuilder {
	return newMergeUpdateBuilder(b)
}

func (b *mergeWhenBuilder) ThenInsert() MergeInsertBuilder {
	return newMergeInsertBuilder(b)
}

func (b *mergeWhenBuilder) ThenDelete() MergeQuery {
	// Build the condition string for WhenDelete
	expr := b.whenClause
	if b.condition != nil {
		// Build additional conditions
		bs, err := b.query.eb.Expr("? AND ?", bun.Safe(expr), b.condition).AppendQuery(b.query.query.DB().Formatter(), nil)
		if err != nil {
			panic(fmt.Errorf("merge: merge condition build failed: %w", err))
		}

		expr = string(bs)
	}

	// Use bun's WhenDelete method
	b.query.query.WhenDelete(expr)

	return b.query
}

func (b *mergeWhenBuilder) ThenDoNothing() MergeQuery {
	// ThenDoNothing doesn't need to add any WHEN clause to the query
	// It's used for conditional logic without action
	return b.query
}

// mergeUpdateBuilder implements MergeUpdateBuilder interface.
type mergeUpdateBuilder struct {
	parent *mergeWhenBuilder
	sets   []schema.QueryAppender
}

// newMergeUpdateBuilder creates a new MergeUpdateBuilder instance.
func newMergeUpdateBuilder(parent *mergeWhenBuilder) *mergeUpdateBuilder {
	return &mergeUpdateBuilder{
		parent: parent,
		sets:   make([]schema.QueryAppender, 0),
	}
}

func (b *mergeUpdateBuilder) Set(column string, value any) MergeUpdateBuilder {
	eb := b.parent.query.eb
	setExpr := eb.Expr("? = ?", bun.Name(column), value)
	b.sets = append(b.sets, setExpr)

	return b
}

func (b *mergeUpdateBuilder) SetExpr(column string, builder func(ExprBuilder) any) MergeUpdateBuilder {
	eb := b.parent.query.eb
	expr := builder(eb)
	setExpr := eb.Expr("? = ?", bun.Name(column), expr)
	b.sets = append(b.sets, setExpr)

	return b
}

func (b *mergeUpdateBuilder) SetColumns(columns ...string) MergeUpdateBuilder {
	eb := b.parent.query.eb
	for _, column := range columns {
		setExpr := eb.Expr("? = SOURCE.?", bun.Name(column), bun.Name(column))
		b.sets = append(b.sets, setExpr)
	}

	return b
}

func (b *mergeUpdateBuilder) SetExcept(columns ...string) MergeUpdateBuilder {
	// Get the target table schema
	table := getTableSchemaFromQuery(b.parent.query.query)
	if table == nil {
		panic("merge: SetExcept requires a table schema - use Model() first")
	}

	// Create a set of columns to exclude
	excludeSet := make(map[string]bool)
	for _, col := range columns {
		excludeSet[col] = true
	}

	// Add all table fields except the excluded ones
	eb := b.parent.query.eb

	for _, field := range table.Fields {
		if !excludeSet[field.Name] {
			setExpr := eb.Expr("? = SOURCE.?", bun.Name(field.Name), bun.Name(field.Name))
			b.sets = append(b.sets, setExpr)
		}
	}

	return b
}

func (b *mergeUpdateBuilder) End() MergeQuery {
	// Build the condition string for WhenUpdate
	expr := b.parent.whenClause
	if b.parent.condition != nil {
		// Build additional conditions
		bs, err := b.parent.query.eb.Expr("? AND ?", bun.Safe(expr), b.parent.condition).AppendQuery(b.parent.query.query.DB().Formatter(), nil)
		if err != nil {
			panic(fmt.Errorf("merge: merge condition build failed: %w", err))
		}

		expr = string(bs)
	}

	// Use bun's WhenUpdate method with the correct signature
	b.parent.query.query.WhenUpdate(expr, func(q *bun.UpdateQuery) *bun.UpdateQuery {
		for _, set := range b.sets {
			q.Set("?", set)
		}

		return q
	})

	return b.parent.query
}

// mergeInsertBuilder implements MergeInsertBuilder interface.
type mergeInsertBuilder struct {
	parent *mergeWhenBuilder
	values []struct {
		column string
		value  schema.QueryAppender
	}
}

// newMergeInsertBuilder creates a new MergeInsertBuilder instance.
func newMergeInsertBuilder(parent *mergeWhenBuilder) *mergeInsertBuilder {
	return &mergeInsertBuilder{
		parent: parent,
		values: make([]struct {
			column string
			value  schema.QueryAppender
		}, 0),
	}
}

func (b *mergeInsertBuilder) Value(column string, value any) MergeInsertBuilder {
	eb := b.parent.query.eb
	valueExpr := eb.Expr("?", value)
	b.values = append(b.values, struct {
		column string
		value  schema.QueryAppender
	}{
		column: column,
		value:  valueExpr,
	})

	return b
}

func (b *mergeInsertBuilder) ValueExpr(column string, builder func(ExprBuilder) any) MergeInsertBuilder {
	eb := b.parent.query.eb
	expr := builder(eb)
	valueExpr := eb.Expr("?", expr)
	b.values = append(b.values, struct {
		column string
		value  schema.QueryAppender
	}{
		column: column,
		value:  valueExpr,
	})

	return b
}

func (b *mergeInsertBuilder) Values(columns ...string) MergeInsertBuilder {
	eb := b.parent.query.eb
	for _, column := range columns {
		valueExpr := eb.Expr("SOURCE.?", bun.Name(column))
		b.values = append(b.values, struct {
			column string
			value  schema.QueryAppender
		}{
			column: column,
			value:  valueExpr,
		})
	}

	return b
}

func (b *mergeInsertBuilder) ValuesExcept(columns ...string) MergeInsertBuilder {
	// Get the target table schema
	table := getTableSchemaFromQuery(b.parent.query.query)
	if table == nil {
		panic("merge: ValuesExcept requires a table schema - use Model() first")
	}

	// Create a set of columns to exclude
	excludeSet := make(map[string]bool)
	for _, col := range columns {
		excludeSet[col] = true
	}

	// Add all table fields except the excluded ones
	eb := b.parent.query.eb

	for _, field := range table.Fields {
		if !excludeSet[field.Name] {
			valueExpr := eb.Expr("SOURCE.?", bun.Name(field.Name))
			b.values = append(b.values, struct {
				column string
				value  schema.QueryAppender
			}{
				column: field.Name,
				value:  valueExpr,
			})
		}
	}

	return b
}

func (b *mergeInsertBuilder) End() MergeQuery {
	// Build the condition string for WhenInsert
	expr := b.parent.whenClause
	if b.parent.condition != nil {
		// Build additional conditions
		bs, err := b.parent.query.eb.Expr("? AND ?", bun.Safe(expr), b.parent.condition).AppendQuery(b.parent.query.query.DB().Formatter(), nil)
		if err != nil {
			panic(fmt.Errorf("merge: merge condition build failed: %w", err))
		}

		expr = string(bs)
	}

	// Use bun's WhenInsert method with the correct signature
	b.parent.query.query.WhenInsert(expr, func(q *bun.InsertQuery) *bun.InsertQuery {
		for _, value := range b.values {
			q.Value(value.column, "?", value.value)
		}

		return q
	})

	return b.parent.query
}
