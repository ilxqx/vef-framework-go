package orm

import (
	"reflect"

	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/datetime"
)

// CreatedAtHandler implements InsertHandler for automatically setting created_at timestamps.
type CreatedAtHandler struct{}

func (*CreatedAtHandler) OnInsert(_ *BunInsertQuery, _ *schema.Table, _ *schema.Field, _ any, value reflect.Value) {
	if value.IsZero() {
		value.Set(reflect.ValueOf(datetime.Now()))
	}
}

func (*CreatedAtHandler) Name() string {
	return constants.ColumnCreatedAt
}

// UpdatedAtHandler implements UpdateHandler for automatically managing updated_at timestamps.
type UpdatedAtHandler struct{}

func (ua *UpdatedAtHandler) OnUpdate(query *BunUpdateQuery, _ *schema.Table, _ *schema.Field, _ any, value reflect.Value) {
	if query.hasSet {
		query.Set(ua.Name(), datetime.Now())
	} else {
		value.Set(reflect.ValueOf(datetime.Now()))
	}
}

func (*UpdatedAtHandler) OnInsert(_ *BunInsertQuery, _ *schema.Table, _ *schema.Field, _ any, value reflect.Value) {
	if value.IsZero() {
		value.Set(reflect.ValueOf(datetime.Now()))
	}
}

// Name returns the column name for the updated_at field.
func (*UpdatedAtHandler) Name() string {
	return constants.ColumnUpdatedAt
}
