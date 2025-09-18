package orm

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/uptrace/bun"
)

// NewCreate creates a new Create instance.
func NewCreate(db bun.IDB) orm.Create {
	return &bunCreate{
		query: db.NewInsert(),
	}
}

type bunCreate struct {
	query *bun.InsertQuery // query is the underlying bun insert query
}

func (c *bunCreate) subQuery(subQuery *bun.SelectQuery) orm.Query {
	return &bunQuery{
		query:      subQuery,
		isSubQuery: true,
	}
}

func (c *bunCreate) buildSubQuery(builder func(query orm.Query)) *bun.SelectQuery {
	subQuery := c.query.NewSelect()
	builder(c.subQuery(subQuery))

	return subQuery
}

func (c *bunCreate) With(name string, builder func(orm.Query)) orm.Create {
	c.query.With(name, c.buildSubQuery(builder))
	return c
}

func (c *bunCreate) WithValues(name string, model any, withOrder ...bool) orm.Create {
	values := c.query.NewValues(model)
	if len(withOrder) > 0 && withOrder[0] {
		values.WithOrder()
	}

	c.query.With(name, values)
	return c
}

func (c *bunCreate) WithRecursive(name string, builder func(orm.Query)) orm.Create {
	c.query.WithRecursive(name, c.buildSubQuery(builder))
	return c
}

func (c *bunCreate) Model(model any) orm.Create {
	c.query.Model(model)
	return c
}

func (c *bunCreate) Table(name string) orm.Create {
	c.query.Table(name)
	return c
}

func (c *bunCreate) TableAs(name string, alias string) orm.Create {
	c.query.TableExpr("? AS ?", bun.Name(name), bun.Name(alias))
	return c
}

func (c *bunCreate) TableExpr(expr string, args ...any) orm.Create {
	c.query.TableExpr(expr, args...)
	return c
}

func (c *bunCreate) TableExprAs(expr string, alias string, args ...any) orm.Create {
	c.query.TableExpr("? AS ?", bun.SafeQuery(expr, args...), bun.Name(alias))
	return c
}

func (c *bunCreate) TableSubQuery(builder func(orm.Query)) orm.Create {
	c.query.TableExpr("(?) AS ?", c.buildSubQuery(builder))
	return c
}

func (c *bunCreate) TableSubQueryAs(builder func(orm.Query), alias string) orm.Create {
	c.query.TableExpr("(?) AS ?", c.buildSubQuery(builder), bun.Name(alias))
	return c
}

func (c *bunCreate) SelectAll() orm.Create {
	c.query.Column(orm.ColumnAll)
	return c
}

func (c *bunCreate) Select(columns ...string) orm.Create {
	c.query.Column(columns...)
	return c
}

func (c *bunCreate) Exclude(columns ...string) orm.Create {
	c.query.ExcludeColumn(columns...)
	return c
}

func (c *bunCreate) ExcludeAll() orm.Create {
	c.query.ExcludeColumn(orm.ColumnAll)
	return c
}

func (c *bunCreate) Column(name string, value any) orm.Create {
	c.query.Value(name, "?", value)
	return c
}

func (c *bunCreate) ColumnExpr(name, expr string, args ...any) orm.Create {
	c.query.Value(name, expr, args...)
	return c
}

func (c *bunCreate) OnConflict(columns ...string) orm.Create {
	c.query.On("CONFLICT (?) DO UPDATE", Names(columns...))
	return c
}

func (c *bunCreate) OnConflictConstraint(constraint string) orm.Create {
	c.query.On("CONFLICT ON CONSTRAINT ? DO UPDATE", bun.Name(constraint))
	return c
}

func (c *bunCreate) OnConflictDoNothing() orm.Create {
	c.query.Ignore()
	return c
}

func (c *bunCreate) OnConflictDoUpdate() orm.Create {
	c.query.On("DUPLICATE KEY UPDATE")
	return c
}

func (c *bunCreate) Set(name string, value ...any) orm.Create {
	if len(value) == 0 {
		c.query.Set("? = EXCLUDED.?", bun.Name(name), bun.Name(name))
	} else {
		c.query.Set("? = ?", bun.Name(name), value[0])
	}
	return c
}

func (c *bunCreate) SetExpr(name, expr string, args ...any) orm.Create {
	c.query.Set("? = ?", bun.Name(name), bun.SafeQuery(expr, args...))
	return c
}

func (c *bunCreate) Where(builder func(orm.ConditionBuilder)) orm.Create {
	cb := newCommonConditionBuilder(getTableSchemaFromQuery(c.query), c.buildSubQuery)
	builder(cb)
	c.query.Where("?", cb)
	return c
}

func (c *bunCreate) Returning(columns ...string) orm.Create {
	c.query.Returning("?", Names(columns...))
	return c
}

func (c *bunCreate) ReturningAll() orm.Create {
	c.query.Returning(orm.ColumnAll)
	return c
}

func (c *bunCreate) ReturningNull() orm.Create {
	c.query.Returning(orm.Null)
	return c
}

func (c *bunCreate) Apply(fns ...orm.ApplyFunc[orm.Create]) orm.Create {
	var crt orm.Create = c
	for _, fn := range fns {
		if fn != nil {
			r := fn(crt)
			if r != nil {
				crt = r
			}
		}
	}

	return crt
}

func (c *bunCreate) beforeCreate() {
	model := c.query.GetModel()
	if model != nil {
		if tm, ok := model.(bun.TableModel); ok {
			table := tm.Table()
			modelValue := model.Value()
			mv := reflect.Indirect(reflect.ValueOf(modelValue))

			for _, autoColumn := range autoColumns {
				if ac, ok := autoColumn.(orm.CreateAutoColumn); ok {
					if field, ok := table.FieldMap[ac.Name()]; ok {
						value := field.Value(mv)
						ac.OnCreate(c.query, table, field, modelValue, value)
					}
				}
			}
		}
	}
}

func (c *bunCreate) Exec(ctx context.Context, dest ...any) (sql.Result, error) {
	c.beforeCreate()
	return c.query.Exec(ctx, dest...)
}

func (c *bunCreate) Scan(ctx context.Context, dest ...any) error {
	c.beforeCreate()
	return c.query.Scan(ctx, dest...)
}
