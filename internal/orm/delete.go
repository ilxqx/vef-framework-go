package orm

import (
	"context"
	"database/sql"

	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/uptrace/bun"
)

// NewDelete creates a new Delete instance.
func NewDelete(db bun.IDB) orm.Delete {
	return &bunDelete{
		query: db.NewDelete(),
	}
}

type bunDelete struct {
	query *bun.DeleteQuery
}

func (*bunDelete) subQuery(subQuery *bun.SelectQuery) orm.Query {
	return &bunQuery{
		query:      subQuery,
		isSubQuery: true,
	}
}

func (d *bunDelete) buildSubQuery(builder func(query orm.Query)) *bun.SelectQuery {
	subQuery := d.query.NewSelect()
	builder(d.subQuery(subQuery))

	return subQuery
}

func (d *bunDelete) With(name string, builder func(orm.Query)) orm.Delete {
	d.query.With(name, d.buildSubQuery(builder))
	return d
}

func (d *bunDelete) WithValues(name string, model any, withOrder ...bool) orm.Delete {
	values := d.query.NewValues(model)
	if len(withOrder) > 0 && withOrder[0] {
		values.WithOrder()
	}

	d.query.With(name, values)
	return d
}

func (d *bunDelete) WithRecursive(name string, builder func(orm.Query)) orm.Delete {
	d.query.WithRecursive(name, d.buildSubQuery(builder))
	return d
}

func (d *bunDelete) Model(model any) orm.Delete {
	d.query.Model(model)
	return d
}

func (d *bunDelete) Table(name string) orm.Delete {
	d.query.Table(name)
	return d
}

func (d *bunDelete) TableAs(name string, alias string) orm.Delete {
	d.query.TableExpr("? AS ?", bun.Name(name), bun.Name(alias))
	return d
}

func (d *bunDelete) TableExpr(expr string, args ...any) orm.Delete {
	d.query.TableExpr(expr, args...)
	return d
}

func (d *bunDelete) TableExprAs(expr string, alias string, args ...any) orm.Delete {
	d.query.TableExpr("? AS ?", bun.SafeQuery(expr, args...), bun.Name(alias))
	return d
}

func (d *bunDelete) TableSubQuery(builder func(orm.Query)) orm.Delete {
	d.query.TableExpr("(?)", d.buildSubQuery(builder))
	return d
}

func (d *bunDelete) TableSubQueryAs(builder func(orm.Query), alias string) orm.Delete {
	d.query.TableExpr("(?) AS ?", d.buildSubQuery(builder), bun.Name(alias))
	return d
}

func (d *bunDelete) Where(builder func(orm.ConditionBuilder)) orm.Delete {
	cb := newQueryConditionBuilder(getTableSchemaFromQuery(d.query), d.query.QueryBuilder(), d.buildSubQuery)
	builder(cb)
	return d
}

func (d *bunDelete) WherePK(columns ...string) orm.Delete {
	d.query.WherePK(columns...)
	return d
}

func (d *bunDelete) WhereDeleted() orm.Delete {
	d.query.WhereDeleted()
	return d
}

func (d *bunDelete) WhereAllWithDeleted() orm.Delete {
	d.query.WhereAllWithDeleted()
	return d
}

func (d *bunDelete) OrderBy(columns ...string) orm.Delete {
	d.query.Order(columns...)
	return d
}

func (d *bunDelete) OrderByNullsFirst(columns ...string) orm.Delete {
	for _, column := range columns {
		d.query.OrderExpr("? NULLS FIRST", bun.Ident(column))
	}

	return d
}

func (d *bunDelete) OrderByNullsLast(columns ...string) orm.Delete {
	for _, column := range columns {
		d.query.OrderExpr("? NULLS LAST", bun.Ident(column))
	}

	return d
}

func (d *bunDelete) OrderByDesc(columns ...string) orm.Delete {
	for _, column := range columns {
		d.query.OrderExpr("? DESC", bun.Ident(column))
	}

	return d
}

func (d *bunDelete) OrderByDescNullsFirst(columns ...string) orm.Delete {
	for _, column := range columns {
		d.query.OrderExpr("? DESC NULLS FIRST", bun.Ident(column))
	}

	return d
}

func (d *bunDelete) OrderByDescNullsLast(columns ...string) orm.Delete {
	for _, column := range columns {
		d.query.OrderExpr("? DESC NULLS LAST", bun.Ident(column))
	}

	return d
}

func (d *bunDelete) OrderByExpr(expr string, args ...any) orm.Delete {
	d.query.OrderExpr(expr, args...)
	return d
}

func (d *bunDelete) ForceDelete() orm.Delete {
	d.query.ForceDelete()
	return d
}

func (d *bunDelete) Limit(limit int) orm.Delete {
	d.query.Limit(limit)
	return d
}

func (d *bunDelete) Returning(columns ...string) orm.Delete {
	d.query.Returning("?", Names(columns...))
	return d
}

func (d *bunDelete) ReturningAll() orm.Delete {
	d.query.Returning(orm.ColumnAll)
	return d
}

func (d *bunDelete) ReturningNull() orm.Delete {
	d.query.Returning(orm.Null)
	return d
}

func (d *bunDelete) Apply(fns ...orm.ApplyFunc[orm.Delete]) orm.Delete {
	var del orm.Delete = d
	for _, fn := range fns {
		if fn != nil {
			r := fn(del)
			if r != nil {
				del = r
			}
		}
	}

	return del
}

func (d *bunDelete) Exec(ctx context.Context, dest ...any) (sql.Result, error) {
	return d.query.Exec(ctx, dest...)
}

func (d *bunDelete) Scan(ctx context.Context, dest ...any) error {
	return d.query.Scan(ctx, dest...)
}
