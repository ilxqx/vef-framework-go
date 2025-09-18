package orm

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/mo"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/feature"
	"github.com/uptrace/bun/schema"
)

var (
	autoColumns = []orm.AutoColumn{
		new(idGenerator),
		new(autoCreatedAt),
		new(autoUpdatedAt),
		new(autoCreatedBy),
		new(autoUpdatedBy),
	}
)

// autoCreatedAt is a struct that implements the CreateAutoColumn interface for auto-generating a created_at column.
type autoCreatedAt struct {
}

func (*autoCreatedAt) OnCreate(query *bun.InsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	if value.IsZero() {
		value.Set(reflect.ValueOf(mo.DateTimeNow()))
	}
}

func (*autoCreatedAt) Name() string {
	return orm.ColumnCreatedAt
}

// autoUpdatedAt is a struct that implements the UpdateAutoColumn interface for auto-generating an updated_at column.
type autoUpdatedAt struct {
}

func (ua *autoUpdatedAt) OnUpdate(query *bun.UpdateQuery, hasSet bool, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	if hasSet {
		name := ua.Name()
		if query.DB().HasFeature(feature.UpdateMultiTable) {
			query.Set("?TableAlias.? = ?", bun.Ident(name), mo.DateTimeNow())
		} else {
			query.Set("? = ?", bun.Ident(name), mo.DateTimeNow())
		}

		query.Returning("?", bun.Name(name))
	} else {
		value.Set(reflect.ValueOf(mo.DateTimeNow()))
	}
}

func (*autoUpdatedAt) OnCreate(query *bun.InsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	if value.IsZero() {
		value.Set(reflect.ValueOf(mo.DateTimeNow()))
	}
}

func (*autoUpdatedAt) Name() string {
	return orm.ColumnUpdatedAt
}

// autoCreatedBy is a struct that implements the CreateAutoColumn interface for auto-generating a created_by column.
type autoCreatedBy struct {
}

func (cb *autoCreatedBy) OnCreate(query *bun.InsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	if value.IsZero() {
		query.Value(cb.Name(), orm.ExprOperator)
	}
}

func (*autoCreatedBy) Name() string {
	return orm.ColumnCreatedBy
}

// autoUpdatedBy is a struct that implements the UpdateAutoColumn interface for auto-generating an updated_by column.
type autoUpdatedBy struct {
}

func (ub *autoUpdatedBy) OnUpdate(query *bun.UpdateQuery, hasSet bool, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	name := ub.Name()
	if hasSet {
		if query.DB().HasFeature(feature.UpdateMultiTable) {
			query.Set("?TableAlias.? = "+orm.ExprOperator, bun.Ident(name))
		} else {
			query.Set("? = "+orm.ExprOperator, bun.Ident(name))
		}
	} else {
		query.Value(name, orm.ExprOperator)
	}

	query.Returning("?", bun.Name(name))
}

func (ub *autoUpdatedBy) OnCreate(query *bun.InsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	if value.IsZero() {
		query.Value(ub.Name(), orm.ExprOperator)
	}
}

func (*autoUpdatedBy) Name() string {
	return orm.ColumnUpdatedBy
}
