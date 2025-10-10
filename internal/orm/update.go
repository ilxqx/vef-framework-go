package orm

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/samber/lo"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/feature"
	"github.com/uptrace/bun/schema"

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
		QueryBuilder: newQueryBuilder(dialect, uq, eb),

		db:      db,
		dialect: dialect,
		eb:      eb,
		query:   uq,
	}
	eb.qb = query

	return query
}

// BunUpdateQuery is the concrete implementation of UpdateQuery interface.
// It wraps bun.UpdateQuery and provides additional functionality for expression building.
type BunUpdateQuery struct {
	QueryBuilder

	db      *BunDb
	dialect schema.Dialect
	eb      ExprBuilder
	query   *bun.UpdateQuery
	hasSet  bool
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

func (q *BunUpdateQuery) TableExpr(alias string, builder func(ExprBuilder) any) UpdateQuery {
	q.query.TableExpr("(?) AS ?", builder(q.eb), bun.Name(alias))

	return q
}

func (q *BunUpdateQuery) TableSubQuery(alias string, builder func(SelectQuery)) UpdateQuery {
	q.query.TableExpr("(?) AS ?", q.BuildSubQuery(builder), bun.Name(alias))

	return q
}

func (q *BunUpdateQuery) Join(model any, builder func(ConditionBuilder), alias ...string) UpdateQuery {
	table := getTableSchema(model, q.query.DB())

	aliasToUse := table.Alias
	if len(alias) > 0 && alias[0] != constants.Empty {
		aliasToUse = alias[0]
	}

	q.query.Join(
		"? ? AS ?",
		bun.Safe(JoinInner.String()),
		bun.Name(table.Name),
		bun.Name(aliasToUse),
	)
	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunUpdateQuery) JoinTable(name string, builder func(ConditionBuilder), alias ...string) UpdateQuery {
	if len(alias) > 0 && alias[0] != constants.Empty {
		q.query.Join("? ? AS ?", bun.Safe(JoinInner.String()), bun.Name(name), bun.Name(alias[0]))
	} else {
		q.query.Join("? ?", bun.Safe(JoinInner.String()), bun.Name(name))
	}

	q.query.JoinOn("?", q.BuildCondition(builder))

	return q
}

func (q *BunUpdateQuery) JoinSubQuery(alias string, sqBuilder func(query SelectQuery), cBuilder func(ConditionBuilder)) UpdateQuery {
	q.query.Join("? (?) AS ?", bun.Safe(JoinInner.String()), q.BuildSubQuery(sqBuilder), bun.Name(alias))
	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunUpdateQuery) JoinExpr(alias string, eBuilder func(ExprBuilder) any, cBuilder func(ConditionBuilder)) UpdateQuery {
	q.query.Join("? (?) AS ?", bun.Safe(JoinInner.String()), eBuilder(q.eb), bun.Name(alias))
	q.query.JoinOn("?", q.BuildCondition(cBuilder))

	return q
}

func (q *BunUpdateQuery) Where(builder func(ConditionBuilder)) UpdateQuery {
	cb := newQueryConditionBuilder(q.query.QueryBuilder(), q)
	builder(cb)

	return q
}

func (q *BunUpdateQuery) WherePK(columns ...string) UpdateQuery {
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
	q.query.Column(columnAll)

	return q
}

func (q *BunUpdateQuery) Select(columns ...string) UpdateQuery {
	q.query.Column(columns...)

	return q
}

func (q *BunUpdateQuery) Exclude(columns ...string) UpdateQuery {
	q.query.ExcludeColumn(columns...)

	return q
}

func (q *BunUpdateQuery) ExcludeAll() UpdateQuery {
	q.query.ExcludeColumn(columnAll)

	return q
}

func (q *BunUpdateQuery) Column(name string, value any) UpdateQuery {
	q.query.Value(name, "?", value)
	q.query.Returning("?", bun.Ident(name))

	return q
}

func (q *BunUpdateQuery) ColumnExpr(name string, builder func(ExprBuilder) any) UpdateQuery {
	q.query.Value(name, "?", builder(q.eb))
	q.query.Returning("?", bun.Ident(name))

	return q
}

func (q *BunUpdateQuery) Set(name string, value any) UpdateQuery {
	if q.query.DB().HasFeature(feature.UpdateMultiTable) {
		q.query.Set("?TableAlias.? = ?", bun.Ident(name), value)
	} else {
		q.query.Set("? = ?", bun.Ident(name), value)
	}

	if lo.IsNotNil(q.query.GetModel().Value()) {
		q.query.Returning("?", bun.Ident(name))
	}

	q.hasSet = true

	return q
}

func (q *BunUpdateQuery) SetExpr(name string, builder func(ExprBuilder) any) UpdateQuery {
	if q.query.DB().HasFeature(feature.UpdateMultiTable) {
		q.query.Set("?TableAlias.? = ?", bun.Ident(name), builder(q.eb))
	} else {
		q.query.Set("? = ?", bun.Ident(name), builder(q.eb))
	}

	if lo.IsNotNil(q.query.GetModel().Value()) {
		q.query.Returning("?", bun.Ident(name))
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
		q.query.OrderExpr("? DESC", bun.Ident(column))
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
	q.query.Returning("?", Names(columns...))

	return q
}

func (q *BunUpdateQuery) ReturningAll() UpdateQuery {
	q.query.Returning(columnAll)

	return q
}

func (q *BunUpdateQuery) ReturningNone() UpdateQuery {
	q.query.Returning(sqlNull)

	return q
}

func (q *BunUpdateQuery) Bulk() UpdateQuery {
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

// beforeUpdate applies auto column handlers before executing the update operation.
// It processes UpdateHandler for fields like updated_at and updated_by,
// and excludes InsertHandler-only fields from being updated.
func (q *BunUpdateQuery) beforeUpdate() {
	if table := q.GetTable(); table != nil {
		modelValue := q.query.GetModel().Value()
		mv := reflect.Indirect(reflect.ValueOf(modelValue))

		processAutoColumns(autoColumns, q.query, q.hasSet, table, modelValue, mv)
	}
}

func (q *BunUpdateQuery) Exec(ctx context.Context, dest ...any) (res sql.Result, err error) {
	q.beforeUpdate()

	if res, err = q.query.Exec(ctx, dest...); err != nil && dbhelpers.IsDuplicateKeyError(err) {
		logger.Warnf("Record already exists: %v", err)

		return nil, result.ErrRecordAlreadyExists
	}

	return res, err
}

func (q *BunUpdateQuery) Scan(ctx context.Context, dest ...any) (err error) {
	q.beforeUpdate()

	if err = q.query.Scan(ctx, dest...); err != nil && dbhelpers.IsDuplicateKeyError(err) {
		logger.Warnf("Record already exists: %v", err)

		return result.ErrRecordAlreadyExists
	}

	return err
}

func (q *BunUpdateQuery) Unwrap() *bun.UpdateQuery {
	return q.query
}
