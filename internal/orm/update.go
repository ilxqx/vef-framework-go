package orm

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/feature"
	"github.com/uptrace/bun/schema"

	collections "github.com/ilxqx/go-collections"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/dbhelpers"
	"github.com/ilxqx/vef-framework-go/result"
)

// NewUpdateQuery creates a new UpdateQuery instance with the provided database instance.
// It initializes the query builders and sets up the table schema context for proper query building.
func NewUpdateQuery(db *BunDb) *BunUpdateQuery {
	eb := &QueryExprBuilder{}
	uq := db.db.NewUpdate()
	dialect := db.db.Dialect()
	query := &BunUpdateQuery{
		QueryBuilder: newQueryBuilder(db, dialect, uq, eb),

		db:      db,
		dialect: dialect,
		eb:      eb,
		query:   uq,

		selectedColumns:  collections.NewHashSet[string](),
		returningColumns: collections.NewHashSet[string](),
	}
	eb.qb = query

	return query
}

// BunUpdateQuery is the concrete implementation of UpdateQuery interface.
// It wraps bun.UpdateQuery and provides additional functionality for expression building.
type BunUpdateQuery struct {
	QueryBuilder

	db               *BunDb
	dialect          schema.Dialect
	eb               ExprBuilder
	query            *bun.UpdateQuery
	hasSet           bool
	isBulk           bool
	selectedColumns  collections.Set[string]
	returningColumns collections.Set[string]
}

func (q *BunUpdateQuery) Db() Db {
	return q.db
}

func (q *BunUpdateQuery) With(name string, builder func(SelectQuery)) UpdateQuery {
	q.query.With(name, q.BuildSubQuery(builder))

	return q
}

func (q *BunUpdateQuery) WithValues(name string, model any, withOrder ...bool) UpdateQuery {
	values := q.query.NewValues(model)
	if len(withOrder) > 0 && withOrder[0] {
		values.WithOrder()
	}

	q.query.With(name, values)

	return q
}

func (q *BunUpdateQuery) WithRecursive(name string, builder func(SelectQuery)) UpdateQuery {
	q.query.WithRecursive(name, q.BuildSubQuery(builder))

	return q
}

func (q *BunUpdateQuery) Model(model any) UpdateQuery {
	q.query.Model(model)

	return q
}

func (q *BunUpdateQuery) ModelTable(name string, alias ...string) UpdateQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.ModelTableExpr("? AS ?", bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.ModelTableExpr("? AS ?TableAlias", bun.Name(name))
	}

	return q
}

func (q *BunUpdateQuery) Table(name string, alias ...string) UpdateQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.TableExpr("? AS ?", bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.Table(name)
	}

	return q
}

func (q *BunUpdateQuery) TableFrom(model any, alias ...string) UpdateQuery {
	table := q.db.TableOf(model)

	aliasToUse := table.Alias
	if len(alias) > 0 && alias[0] != constants.Empty {
		aliasToUse = alias[0]
	}

	q.query.TableExpr("? AS ?", bun.Name(table.Name), bun.Name(aliasToUse))

	return q
}

func (q *BunUpdateQuery) TableExpr(builder func(ExprBuilder) any, alias ...string) UpdateQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.TableExpr("(?) AS ?", builder(q.eb), bun.Name(alias[0]))
	} else {
		q.query.TableExpr("(?)", builder(q.eb))
	}

	return q
}

func (q *BunUpdateQuery) TableSubQuery(builder func(SelectQuery), alias ...string) UpdateQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.TableExpr("(?) AS ?", q.BuildSubQuery(builder), bun.Name(alias[0]))
	} else {
		q.query.TableExpr("(?)", q.BuildSubQuery(builder))
	}

	return q
}

func (q *BunUpdateQuery) Where(builder func(ConditionBuilder)) UpdateQuery {
	cb := newQueryConditionBuilder(q.query.QueryBuilder(), q)
	builder(cb)

	return q
}

func (q *BunUpdateQuery) WherePk(columns ...string) UpdateQuery {
	q.query.WherePK(columns...)

	return q
}

func (q *BunUpdateQuery) WhereDeleted() UpdateQuery {
	q.query.WhereDeleted()

	return q
}

func (q *BunUpdateQuery) IncludeDeleted() UpdateQuery {
	q.query.WhereAllWithDeleted()

	return q
}

func (q *BunUpdateQuery) SelectAll() UpdateQuery {
	return q
}

func (q *BunUpdateQuery) Select(columns ...string) UpdateQuery {
	q.query.Column(columns...)
	q.selectedColumns.AddAll(columns...)

	return q
}

func (q *BunUpdateQuery) Exclude(columns ...string) UpdateQuery {
	q.query.ExcludeColumn(columns...)
	q.selectedColumns.RemoveAll(columns...)

	return q
}

func (q *BunUpdateQuery) ExcludeAll() UpdateQuery {
	q.query.ExcludeColumn(columnAll)
	q.selectedColumns.Clear()

	return q
}

func (q *BunUpdateQuery) Column(name string, value any) UpdateQuery {
	q.query.Value(name, "?", value)

	return q
}

func (q *BunUpdateQuery) ColumnExpr(name string, builder func(ExprBuilder) any) UpdateQuery {
	q.query.Value(name, "?", builder(q.eb))

	return q
}

func (q *BunUpdateQuery) Set(name string, value any) UpdateQuery {
	if q.query.DB().HasFeature(feature.UpdateMultiTable) {
		q.query.Set("? = ?", q.eb.Column(name), value)
	} else {
		q.query.Set("? = ?", bun.Name(name), value)
	}

	q.hasSet = true

	return q
}

func (q *BunUpdateQuery) SetExpr(name string, builder func(ExprBuilder) any) UpdateQuery {
	if q.query.DB().HasFeature(feature.UpdateMultiTable) {
		q.query.Set("? = ?", q.eb.Column(name), builder(q.eb))
	} else {
		q.query.Set("? = ?", bun.Name(name), builder(q.eb))
	}

	q.hasSet = true

	return q
}

func (q *BunUpdateQuery) OmitZero() UpdateQuery {
	q.query.OmitZero()

	return q
}

func (q *BunUpdateQuery) OrderBy(columns ...string) UpdateQuery {
	q.query.Order(columns...)

	return q
}

func (q *BunUpdateQuery) OrderByDesc(columns ...string) UpdateQuery {
	for _, column := range columns {
		q.query.OrderExpr("? DESC", q.eb.Column(column))
	}

	return q
}

func (q *BunUpdateQuery) OrderByExpr(builder func(ExprBuilder) any) UpdateQuery {
	q.query.OrderExpr("?", builder(q.eb))

	return q
}

func (q *BunUpdateQuery) Limit(limit int) UpdateQuery {
	q.query.Limit(limit)

	return q
}

func (q *BunUpdateQuery) Returning(columns ...string) UpdateQuery {
	q.returningColumns.AddAll(columns...)

	return q
}

func (q *BunUpdateQuery) ReturningAll() UpdateQuery {
	q.returningColumns.Clear()
	q.returningColumns.Add(columnAll)

	return q
}

func (q *BunUpdateQuery) ReturningNone() UpdateQuery {
	q.returningColumns.Clear()
	q.returningColumns.Add(sqlNull)

	return q
}

func (q *BunUpdateQuery) Bulk() UpdateQuery {
	q.isBulk = true
	q.query.Bulk()

	return q
}

func (q *BunUpdateQuery) Apply(fns ...ApplyFunc[UpdateQuery]) UpdateQuery {
	for _, fn := range fns {
		if fn != nil {
			fn(q)
		}
	}

	return q
}

func (q *BunUpdateQuery) ApplyIf(condition bool, fns ...ApplyFunc[UpdateQuery]) UpdateQuery {
	if condition {
		return q.Apply(fns...)
	}

	return q
}

func (q *BunUpdateQuery) beforeUpdate() {
	if table := q.GetTable(); table != nil {
		q.skipCreateAuditColumns(table)

		modelValue := q.query.GetModel().Value()
		mv := reflect.Indirect(reflect.ValueOf(modelValue))

		processAutoColumns(q, table, modelValue, mv)
	}

	if !q.returningColumns.IsEmpty() {
		q.query.Returning("?", buildReturningExpr(q.returningColumns, q.eb))
	}
}

func (q *BunUpdateQuery) skipCreateAuditColumns(table *schema.Table) {
	if q.hasSet || !q.selectedColumns.IsEmpty() {
		return
	}

	if table.HasField(constants.ColumnCreatedAt) {
		q.Exclude(constants.ColumnCreatedAt)
	}

	if table.HasField(constants.ColumnCreatedBy) {
		q.Exclude(constants.ColumnCreatedBy)
	}
}

func (q *BunUpdateQuery) Exec(ctx context.Context, dest ...any) (res sql.Result, err error) {
	q.beforeUpdate()

	if res, err = q.query.Exec(ctx, dest...); err != nil {
		if dbhelpers.IsDuplicateKeyError(err) {
			logger.Warnf("Record already exists: %v", err)

			return nil, result.ErrRecordAlreadyExists
		}

		if dbhelpers.IsForeignKeyError(err) {
			logger.Warnf("Foreign key violation: %v", err)

			return nil, result.ErrForeignKeyViolation
		}
	}

	return res, err
}

func (q *BunUpdateQuery) Scan(ctx context.Context, dest ...any) (err error) {
	q.beforeUpdate()

	if err = q.query.Scan(ctx, dest...); err != nil {
		if dbhelpers.IsDuplicateKeyError(err) {
			logger.Warnf("Record already exists: %v", err)

			return result.ErrRecordAlreadyExists
		}

		if dbhelpers.IsForeignKeyError(err) {
			logger.Warnf("Foreign key violation: %v", err)

			return result.ErrForeignKeyViolation
		}
	}

	return err
}

func (q *BunUpdateQuery) Unwrap() *bun.UpdateQuery {
	return q.query
}
