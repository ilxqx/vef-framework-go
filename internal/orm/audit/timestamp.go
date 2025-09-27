package audit

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/datetime"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/feature"
	"github.com/uptrace/bun/schema"
)

// CreatedAtHandler implements InsertHandler for automatically setting created_at timestamps.
// It sets the created_at field to the current time when a new record is inserted.
type CreatedAtHandler struct{}

// OnInsert sets the created_at field to the current timestamp if it's not already set.
func (*CreatedAtHandler) OnInsert(query *bun.InsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	if value.IsZero() {
		value.Set(reflect.ValueOf(datetime.Now()))
	}
}

// Name returns the column name for the created_at field.
func (*CreatedAtHandler) Name() string {
	return constants.ColumnCreatedAt
}

// UpdatedAtHandler implements UpdateHandler for automatically managing updated_at timestamps.
// It sets the updated_at field to the current time during both insert and update operations.
type UpdatedAtHandler struct{}

// OnUpdate automatically sets the updated_at field to the current timestamp during update operations.
// If hasSet is true, it adds a SET clause to the query; otherwise, it updates the model value directly.
func (ua *UpdatedAtHandler) OnUpdate(query *bun.UpdateQuery, hasSet bool, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	if hasSet {
		name := ua.Name()
		if query.DB().HasFeature(feature.UpdateMultiTable) {
			query.Set("?TableAlias.? = ?", bun.Ident(name), datetime.Now())
		} else {
			query.Set("? = ?", bun.Ident(name), datetime.Now())
		}

		query.Returning("?", bun.Name(name))
	} else {
		value.Set(reflect.ValueOf(datetime.Now()))
	}
}

// OnInsert sets the updated_at field to the current timestamp during insert operations.
func (*UpdatedAtHandler) OnInsert(query *bun.InsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	if value.IsZero() {
		value.Set(reflect.ValueOf(datetime.Now()))
	}
}

// Name returns the column name for the updated_at field.
func (*UpdatedAtHandler) Name() string {
	return constants.ColumnUpdatedAt
}
