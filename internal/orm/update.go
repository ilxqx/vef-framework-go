package orm

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/feature"
)

// NewUpdate creates a new Update instance.
func NewUpdate(db bun.IDB) orm.Update {
	return &bunUpdate{
		query: db.NewUpdate(),
	}
}

type bunUpdate struct {
	query  *bun.UpdateQuery // query is the underlying bun update query
	hasSet bool             // hasSet indicates if any columns have been explicitly set
}

func (u *bunUpdate) subQuery(subQuery *bun.SelectQuery) orm.Query {
	return &bunQuery{
		query:      subQuery,
		isSubQuery: true,
	}
}

func (u *bunUpdate) buildSubQuery(builder func(query orm.Query)) *bun.SelectQuery {
	subQuery := u.query.NewSelect()
	builder(u.subQuery(subQuery))

	return subQuery
}

func (u *bunUpdate) buildCondition(builder func(orm.ConditionBuilder)) orm.ConditionBuilder {
	cb := newCommonConditionBuilder(getTableSchemaFromQuery(u.query), u.buildSubQuery)
	builder(cb)
	return cb
}

func (u *bunUpdate) With(name string, builder func(orm.Query)) orm.Update {
	u.query.With(name, u.buildSubQuery(builder))
	return u
}

func (u *bunUpdate) WithValues(name string, model any, withOrder ...bool) orm.Update {
	values := u.query.NewValues(model)
	if len(withOrder) > 0 && withOrder[0] {
		values.WithOrder()
	}

	u.query.With(name, values)
	return u
}

func (u *bunUpdate) WithRecursive(name string, builder func(orm.Query)) orm.Update {
	u.query.WithRecursive(name, u.buildSubQuery(builder))
	return u
}

func (u *bunUpdate) Model(model any) orm.Update {
	u.query.Model(model)
	return u
}

func (u *bunUpdate) Table(name string) orm.Update {
	u.query.Table(name)
	return u
}

func (u *bunUpdate) TableAs(name string, alias string) orm.Update {
	u.query.TableExpr("? AS ?", bun.Name(name), bun.Name(alias))
	return u
}

func (u *bunUpdate) TableExpr(expr string, args ...any) orm.Update {
	u.query.TableExpr(expr, args...)
	return u
}

func (u *bunUpdate) TableExprAs(expr string, alias string, args ...any) orm.Update {
	u.query.TableExpr("? AS ?", bun.SafeQuery(expr, args), bun.Name(alias))
	return u
}

func (u *bunUpdate) TableSubQuery(builder func(orm.Query)) orm.Update {
	u.query.TableExpr("(?)", u.buildSubQuery(builder))
	return u
}

func (u *bunUpdate) TableSubQueryAs(builder func(orm.Query), alias string) orm.Update {
	u.query.TableExpr("(?) AS ?", u.buildSubQuery(builder), bun.Name(alias))
	return u
}

func (u *bunUpdate) Join(model any, builder func(orm.ConditionBuilder)) orm.Update {
	table := getTableSchema(model, u.query.DB())
	u.query.Join("JOIN ? AS ?", bun.Name(table.Name), bun.Name(table.Alias))
	u.query.JoinOn("?", u.buildCondition(builder))
	return u
}

func (u *bunUpdate) JoinAs(model any, alias string, builder func(orm.ConditionBuilder)) orm.Update {
	table := getTableSchema(model, u.query.DB())
	u.query.Join("JOIN ? AS ?", bun.Name(table.Name), bun.Name(alias))
	u.query.JoinOn("?", u.buildCondition(builder))
	return u
}

func (u *bunUpdate) JoinTable(name string, builder func(orm.ConditionBuilder)) orm.Update {
	u.query.Join("JOIN ?", bun.Name(name))
	u.query.JoinOn("?", u.buildCondition(builder))
	return u
}

func (u *bunUpdate) JoinTableAs(name string, alias string, builder func(orm.ConditionBuilder)) orm.Update {
	u.query.Join("JOIN ? AS ?", bun.Name(name), bun.Name(alias))
	u.query.JoinOn("?", u.buildCondition(builder))
	return u
}

func (u *bunUpdate) JoinSubQuery(builder func(query orm.Query), conditionBuilder func(orm.ConditionBuilder)) orm.Update {
	u.query.Join("JOIN (?)", u.buildSubQuery(builder))
	u.query.JoinOn("?", u.buildCondition(conditionBuilder))
	return u
}

func (u *bunUpdate) JoinSubQueryAs(builder func(query orm.Query), alias string, conditionBuilder func(orm.ConditionBuilder)) orm.Update {
	u.query.Join("JOIN (?) AS ?", u.buildSubQuery(builder), bun.Name(alias))
	u.query.JoinOn("?", u.buildCondition(conditionBuilder))
	return u
}

func (u *bunUpdate) JoinExpr(expr string, builder func(cb orm.ConditionBuilder), args ...any) orm.Update {
	u.query.Join("JOIN ?", bun.SafeQuery(expr, args))
	u.query.JoinOn("?", u.buildCondition(builder))
	return u
}

func (u *bunUpdate) JoinExprAs(expr string, alias string, builder func(cb orm.ConditionBuilder), args ...any) orm.Update {
	u.query.Join("JOIN ? AS ?", bun.SafeQuery(expr, args), bun.Name(alias))
	u.query.JoinOn("?", u.buildCondition(builder))
	return u
}

func (u *bunUpdate) Where(builder func(cb orm.ConditionBuilder)) orm.Update {
	cb := newQueryConditionBuilder(getTableSchemaFromQuery(u.query), u.query.QueryBuilder(), u.buildSubQuery)
	builder(cb)
	return u
}

func (u *bunUpdate) WherePK(columns ...string) orm.Update {
	u.query.WherePK(columns...)
	return u
}

func (u *bunUpdate) WhereDeleted() orm.Update {
	u.query.WhereDeleted()
	return u
}

func (u *bunUpdate) WhereAllWithDeleted() orm.Update {
	u.query.WhereAllWithDeleted()
	return u
}

func (u *bunUpdate) SelectAll() orm.Update {
	u.query.Column(orm.ColumnAll)
	return u
}

func (u *bunUpdate) Select(columns ...string) orm.Update {
	u.query.Column(columns...)
	return u
}

func (u *bunUpdate) Exclude(columns ...string) orm.Update {
	u.query.ExcludeColumn(columns...)
	return u
}

func (u *bunUpdate) ExcludeAll() orm.Update {
	u.query.ExcludeColumn(orm.ColumnAll)
	return u
}

func (u *bunUpdate) Column(name string, value any) orm.Update {
	u.query.Value(name, "?", value)
	u.query.Returning("?", bun.Ident(name))
	return u
}

func (u *bunUpdate) ColumnExpr(name, expr string, args ...any) orm.Update {
	u.query.Value(name, expr, args...)
	u.query.Returning("?", bun.Ident(name))
	return u
}

func (u *bunUpdate) Set(name string, value any) orm.Update {
	if u.query.DB().HasFeature(feature.UpdateMultiTable) {
		u.query.Set("?TableAlias.? = ?", bun.Ident(name), value)
	} else {
		u.query.Set("? = ?", bun.Ident(name), value)
	}
	u.query.Returning("?", bun.Ident(name))
	u.hasSet = true
	return u
}

func (u *bunUpdate) SetExpr(name, expr string, args ...any) orm.Update {
	if u.query.DB().HasFeature(feature.UpdateMultiTable) {
		u.query.Set("?TableAlias.? = ?", bun.Ident(name), bun.SafeQuery(expr, args))
	} else {
		u.query.Set("? = ?", bun.Ident(name), bun.SafeQuery(expr, args))
	}
	u.query.Returning("?", bun.Ident(name))
	u.hasSet = true
	return u
}

func (u *bunUpdate) OmitZero() orm.Update {
	u.query.OmitZero()
	return u
}

func (u *bunUpdate) OrderBy(columns ...string) orm.Update {
	u.query.Order(columns...)
	return u
}

func (u *bunUpdate) OrderByNullsFirst(columns ...string) orm.Update {
	for _, column := range columns {
		u.query.OrderExpr("? NULLS FIRST", bun.Ident(column))
	}

	return u
}

func (u *bunUpdate) OrderByNullsLast(columns ...string) orm.Update {
	for _, column := range columns {
		u.query.OrderExpr("? NULLS LAST", bun.Ident(column))
	}

	return u
}

func (u *bunUpdate) OrderByDesc(columns ...string) orm.Update {
	for _, column := range columns {
		u.query.OrderExpr("? DESC", bun.Ident(column))
	}

	return u
}

func (u *bunUpdate) OrderByDescNullsFirst(columns ...string) orm.Update {
	for _, column := range columns {
		u.query.OrderExpr("? DESC NULLS FIRST", bun.Ident(column))
	}

	return u
}

func (u *bunUpdate) OrderByDescNullsLast(columns ...string) orm.Update {
	for _, column := range columns {
		u.query.OrderExpr("? DESC NULLS LAST", bun.Ident(column))
	}

	return u
}

func (u *bunUpdate) OrderByExpr(expr string, args ...any) orm.Update {
	u.query.OrderExpr(expr, args...)
	return u
}

func (u *bunUpdate) Limit(limit int) orm.Update {
	u.query.Limit(limit)
	return u
}

func (u *bunUpdate) Returning(columns ...string) orm.Update {
	u.query.Returning("?", Names(columns...))
	return u
}

func (u *bunUpdate) ReturningAll() orm.Update {
	u.query.Returning(orm.ColumnAll)
	return u
}

func (u *bunUpdate) ReturningNull() orm.Update {
	u.query.Returning(orm.Null)
	return u
}

func (u *bunUpdate) Bulk() orm.Update {
	u.query.Bulk()
	return u
}

func (u *bunUpdate) Apply(fns ...orm.ApplyFunc[orm.Update]) orm.Update {
	var up orm.Update = u
	for _, fn := range fns {
		if fn != nil {
			r := fn(up)
			if r != nil {
				up = r
			}
		}
	}

	return up
}

func (u *bunUpdate) beforeUpdate() {
	model := u.query.GetModel()
	if model != nil {
		if tm, ok := model.(bun.TableModel); ok {
			table := tm.Table()
			modelValue := model.Value()
			mv := reflect.Indirect(reflect.ValueOf(modelValue))

			for _, autoColumn := range autoColumns {
				if ac, ok := autoColumn.(orm.UpdateAutoColumn); ok {
					if field, ok := table.FieldMap[ac.Name()]; ok {
						value := field.Value(mv)
						ac.OnUpdate(u.query, u.hasSet, table, field, modelValue, value)
					}
				} else {
					if ac, ok := autoColumn.(orm.CreateAutoColumn); ok {
						if field, ok := table.FieldMap[ac.Name()]; ok {
							u.query.ExcludeColumn(field.Name)
						}
					}
				}
			}
		}
	}
}

func (u *bunUpdate) Exec(ctx context.Context, dest ...any) (sql.Result, error) {
	u.beforeUpdate()
	return u.query.Exec(ctx, dest...)
}

func (u *bunUpdate) Scan(ctx context.Context, dest ...any) error {
	u.beforeUpdate()
	return u.query.Scan(ctx, dest...)
}
