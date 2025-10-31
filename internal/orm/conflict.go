package orm

import (
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

// ConflictBuilder is used to configure INSERT conflict handling (UPSERT) target in a dialect-aware way.
// This is the first stage that defines the conflict target (columns, constraints, conditions).
type ConflictBuilder interface {
	Columns(columns ...string) ConflictBuilder
	Constraint(name string) ConflictBuilder
	Where(func(ConditionBuilder)) ConflictBuilder

	// DoNothing performs DO NOTHING on conflict and finalizes the conflict handling.
	DoNothing()
	// DoUpdate performs DO UPDATE on conflict and returns a builder for update operations.
	DoUpdate() ConflictUpdateBuilder
}

// ConflictUpdateBuilder is used to configure the UPDATE part of conflict handling.
// This is the second stage that defines what to update when conflicts occur.
type ConflictUpdateBuilder interface {
	// Set adds an assignment in DO UPDATE clause. If no value provided, uses excluded/VALUES value when supported.
	Set(column string, value ...any) ConflictUpdateBuilder
	// SetExpr adds an expression assignment in DO UPDATE clause.
	SetExpr(column string, builder func(ExprBuilder) any) ConflictUpdateBuilder
	// Where adds a predicate to DO UPDATE (PostgreSQL/SQLite). Ignored on MySQL.
	Where(func(ConditionBuilder)) ConflictUpdateBuilder
}

type InsertQueryConflictBuilder struct {
	qb QueryBuilder
	eb ExprBuilder
	// target
	columns     []string
	constraint  string
	targetWhere schema.QueryAppender
	// action
	action ConflictAction
	// updates
	sets        []schema.QueryAppender
	updateWhere schema.QueryAppender
}

func newConflictBuilder(qb QueryBuilder) *InsertQueryConflictBuilder {
	return &InsertQueryConflictBuilder{
		qb: qb,
		eb: qb.ExprBuilder(),
	}
}

func (b *InsertQueryConflictBuilder) Columns(columns ...string) ConflictBuilder {
	b.columns = append(b.columns, columns...)

	return b
}

func (b *InsertQueryConflictBuilder) Constraint(name string) ConflictBuilder {
	b.constraint = name

	return b
}

func (b *InsertQueryConflictBuilder) Where(builder func(ConditionBuilder)) ConflictBuilder {
	b.targetWhere = b.qb.BuildCondition(builder)

	return b
}

func (b *InsertQueryConflictBuilder) DoNothing() {
	b.action = ConflictDoNothing
}

func (b *InsertQueryConflictBuilder) DoUpdate() ConflictUpdateBuilder {
	b.action = ConflictDoUpdate

	return &InsertQueryConflictUpdateBuilder{parent: b}
}

// InsertQueryConflictUpdateBuilder implements ConflictUpdateBuilder interface.
type InsertQueryConflictUpdateBuilder struct {
	parent *InsertQueryConflictBuilder
}

func (b *InsertQueryConflictUpdateBuilder) Set(column string, value ...any) ConflictUpdateBuilder {
	var valueExpr any
	if len(value) > 0 {
		valueExpr = value[0]
	} else {
		b.parent.eb.ExecByDialect(DialectExecs{
			Postgres: func() {
				valueExpr = b.parent.eb.Expr("EXCLUDED.?", bun.Name(column))
			},
			MySQL: func() {
				valueExpr = bun.Name(column)
			},
			SQLite: func() {
				valueExpr = b.parent.eb.Expr("EXCLUDED.?", bun.Name(column))
			},
		})
	}

	setExpr := b.parent.eb.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.parent.eb.Expr("? = ?", bun.Name(column), valueExpr)
		},
		MySQL: func() schema.QueryAppender {
			return bun.SafeQuery("? = ?", bun.Name(column), valueExpr)
		},
		SQLite: func() schema.QueryAppender {
			return bun.SafeQuery("? = ?", bun.Name(column), valueExpr)
		},
	})

	b.parent.sets = append(b.parent.sets, setExpr)

	return b
}

func (b *InsertQueryConflictUpdateBuilder) SetExpr(column string, builder func(ExprBuilder) any) ConflictUpdateBuilder {
	valueExpr := builder(b.parent.eb)
	setExpr := b.parent.eb.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return b.parent.eb.Expr("? = ?", bun.Name(column), valueExpr)
		},
		MySQL: func() schema.QueryAppender {
			return b.parent.eb.Expr("? = ?", bun.Name(column), valueExpr)
		},
		SQLite: func() schema.QueryAppender {
			return b.parent.eb.Expr("? = ?", bun.Name(column), valueExpr)
		},
	})

	b.parent.sets = append(b.parent.sets, setExpr)

	return b
}

func (b *InsertQueryConflictUpdateBuilder) Where(builder func(ConditionBuilder)) ConflictUpdateBuilder {
	b.parent.updateWhere = b.parent.qb.BuildCondition(builder)

	return b
}

// build applies the configured conflict handling to the underlying bun.InsertQuery.
func (b *InsertQueryConflictBuilder) build(query *bun.InsertQuery) {
	// Dialect specific handling
	b.eb.ExecByDialect(DialectExecs{
		MySQL: func() {
			// MySQL: ON DUPLICATE KEY UPDATE ... or INSERT IGNORE for do-nothing
			if b.action == ConflictDoNothing {
				query.Ignore()

				return
			}

			// Otherwise treat as DO UPDATE
			query.On("DUPLICATE KEY UPDATE")

			for _, set := range b.sets {
				query.Set("?", set)
			}
			// MySQL has no DO UPDATE WHERE; ignore updateWhere/targetWhere
		},
		Default: func() {
			// PostgreSQL/SQLite
			// Build conflict target
			var (
				target  schema.QueryAppender
				builder strings.Builder
				args    []any
			)

			if b.constraint != constants.Empty {
				target = b.eb.Expr("CONSTRAINT ?", bun.Name(b.constraint))
			} else if len(b.columns) > 0 {
				target = b.eb.Expr("(?)", Names(b.columns...))
			}

			_, _ = builder.WriteString("CONFLICT ")
			if target == nil {
				// PostgreSQL requires a conflict target (columns or constraint) for DO UPDATE.
				// When no target is specified, fallback to DO NOTHING to avoid syntax errors.
				_, _ = builder.WriteString(ConflictDoNothing.String())
				query.On(builder.String())

				return
			}

			_, _ = builder.WriteString("? ")

			args = append(args, target)

			if b.targetWhere != nil {
				// Target WHERE (partial index)
				_, _ = builder.WriteString("WHERE ? ")

				args = append(args, b.targetWhere)
			}

			_, _ = builder.WriteString(b.action.String())

			query.On(builder.String(), args...)

			for _, set := range b.sets {
				query.Set("?", set)
			}

			if b.updateWhere != nil {
				query.Where("?", b.updateWhere)
			}
		},
	})
}
