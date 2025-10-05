package orm

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/dbhelpers"
	"github.com/ilxqx/vef-framework-go/result"
)

// NewInsertQuery creates a new InsertQuery instance with the provided database connection.
// It initializes the query builders and sets up the table schema context for proper query building.
func NewInsertQuery(db bun.IDB) *BunInsertQuery {
	eb := &QueryExprBuilder{}
	iq := db.NewInsert()
	dialect := db.Dialect()
	query := &BunInsertQuery{
		QueryBuilder: newQueryBuilder(dialect, iq, eb),

		dialect: dialect,
		eb:      eb,
		query:   iq,
	}
	eb.qb = query

	return query
}

// BunInsertQuery is the concrete implementation of InsertQuery interface.
// It wraps bun.InsertQuery and provides additional functionality for expression building.
type BunInsertQuery struct {
	QueryBuilder

	dialect schema.Dialect
	eb      ExprBuilder
	query   *bun.InsertQuery
}

func (q *BunInsertQuery) With(name string, builder func(SelectQuery)) InsertQuery {
	q.query.With(name, q.BuildSubQuery(builder))

	return q
}

func (q *BunInsertQuery) WithValues(name string, model any, withOrder ...bool) InsertQuery {
	values := q.query.NewValues(model)
	if len(withOrder) > 0 && withOrder[0] {
		values.WithOrder()
	}

	q.query.With(name, values)

	return q
}

func (q *BunInsertQuery) WithRecursive(name string, builder func(SelectQuery)) InsertQuery {
	q.query.WithRecursive(name, q.BuildSubQuery(builder))

	return q
}

func (q *BunInsertQuery) Model(model any) InsertQuery {
	q.query.Model(model)

	return q
}

func (q *BunInsertQuery) ModelTable(name string, alias ...string) InsertQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.ModelTableExpr("? AS ?", bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.ModelTableExpr("?", bun.Name(name))
	}

	return q
}

func (q *BunInsertQuery) Table(name string, alias ...string) InsertQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.TableExpr("? AS ?", bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.Table(name)
	}

	return q
}

func (q *BunInsertQuery) TableExpr(alias string, builder func(ExprBuilder) any) InsertQuery {
	q.query.TableExpr("(?) AS ?", builder(q.eb), bun.Name(alias))

	return q
}

func (q *BunInsertQuery) TableSubQuery(alias string, builder func(SelectQuery)) InsertQuery {
	q.query.TableExpr("(?) AS ?", q.BuildSubQuery(builder), bun.Name(alias))

	return q
}

// OnConflict configures conflict handling via a dialect-aware builder.
func (q *BunInsertQuery) OnConflict(builder func(ConflictBuilder)) InsertQuery {
	cb := newConflictBuilder(q)
	builder(cb)
	cb.build(q.query)

	return q
}

func (q *BunInsertQuery) SelectAll() InsertQuery {
	q.query.Column(columnAll)

	return q
}

func (q *BunInsertQuery) Select(columns ...string) InsertQuery {
	q.query.Column(columns...)

	return q
}

func (q *BunInsertQuery) Exclude(columns ...string) InsertQuery {
	q.query.ExcludeColumn(columns...)

	return q
}

func (q *BunInsertQuery) ExcludeAll() InsertQuery {
	q.query.ExcludeColumn(columnAll)

	return q
}

func (q *BunInsertQuery) Column(name string, value any) InsertQuery {
	q.query.Value(name, "?", value)

	return q
}

func (q *BunInsertQuery) ColumnExpr(name string, builder func(ExprBuilder) any) InsertQuery {
	expr := builder(q.eb)
	q.query.Value(name, "?", expr)

	return q
}

func (q *BunInsertQuery) Returning(columns ...string) InsertQuery {
	q.query.Returning("?", Names(columns...))

	return q
}

func (q *BunInsertQuery) ReturningAll() InsertQuery {
	q.query.Returning(columnAll)

	return q
}

func (q *BunInsertQuery) ReturningNone() InsertQuery {
	q.query.Returning(sqlNull)

	return q
}

func (q *BunInsertQuery) Apply(fns ...ApplyFunc[InsertQuery]) InsertQuery {
	for _, fn := range fns {
		if fn != nil {
			fn(q)
		}
	}

	return q
}

func (q *BunInsertQuery) ApplyIf(condition bool, fns ...ApplyFunc[InsertQuery]) InsertQuery {
	if condition {
		return q.Apply(fns...)
	}

	return q
}

// beforeInsert applies auto column handlers before executing the insert operation.
// It processes InsertHandler to automatically set values like IDs, timestamps, and user tracking.
func (q *BunInsertQuery) beforeInsert() {
	model := q.query.GetModel()
	if model != nil {
		if tm, ok := model.(bun.TableModel); ok {
			table := tm.Table()
			modelValue := model.Value()
			mv := reflect.Indirect(reflect.ValueOf(modelValue))

			processAutoColumns(autoColumns, q.query, false, table, modelValue, mv)
		}
	}
}

func (q *BunInsertQuery) Exec(ctx context.Context, dest ...any) (sql.Result, error) {
	q.beforeInsert()

	res, err := q.query.Exec(ctx, dest...)
	if err != nil {
		if dbhelpers.IsDuplicateKeyError(err) {
			logger.Warnf("Record already exists: %v", err)

			return nil, result.ErrRecordAlreadyExists
		}

		return nil, err
	}

	return res, nil
}

func (q *BunInsertQuery) Scan(ctx context.Context, dest ...any) error {
	q.beforeInsert()

	if err := q.query.Scan(ctx, dest...); err != nil {
		if dbhelpers.IsDuplicateKeyError(err) {
			logger.Warnf("Record already exists: %v", err)

			return result.ErrRecordAlreadyExists
		}

		return err
	}

	return nil
}

func (q *BunInsertQuery) Unwrap() *bun.InsertQuery {
	return q.query
}
