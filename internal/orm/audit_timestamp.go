package orm

import (
	"reflect"

	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/datetime"
)

// CreatedAtHandler implements InsertHandler for automatically setting created_at timestamps.
type CreatedAtHandler struct{}

func (*CreatedAtHandler) OnInsert(query *BunInsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	if value.IsZero() {
		value.Set(reflect.ValueOf(datetime.Now()))
	}
}

func (*CreatedAtHandler) Name() string {
	return constants.ColumnCreatedAt
}

// UpdatedAtHandler implements UpdateHandler for automatically managing updated_at timestamps.
type UpdatedAtHandler struct{}

func (ua *UpdatedAtHandler) OnUpdate(query *BunUpdateQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	name := ua.Name()

	if query.hasSet {
		query.Set(name, datetime.Now())
	} else {
		value.Set(reflect.ValueOf(datetime.Now()))
	}
}

func (*UpdatedAtHandler) OnInsert(query *BunInsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	if value.IsZero() {
		value.Set(reflect.ValueOf(datetime.Now()))
	}
}

// Name returns the column name for the updated_at field.
func (*UpdatedAtHandler) Name() string {
	return constants.ColumnUpdatedAt
}
