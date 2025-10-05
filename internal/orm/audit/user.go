package audit

import (
	"reflect"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/feature"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

// CreatedByHandler implements InsertHandler for automatically setting created_by user information.
// It sets the created_by field to the current user identifier when a new record is inserted.
type CreatedByHandler struct{}

// OnInsert sets the created_by field to the current user if it's not already set.
func (cb *CreatedByHandler) OnInsert(query *bun.InsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	if value.IsZero() {
		query.Value(cb.Name(), constants.ExprOperator)
	}
}

// Name returns the column name for the created_by field.
func (*CreatedByHandler) Name() string {
	return constants.ColumnCreatedBy
}

// UpdatedByHandler implements UpdateHandler for automatically managing updated_by user information.
// It sets the updated_by field to the current user identifier during both insert and update operations.
type UpdatedByHandler struct{}

// OnUpdate automatically sets the updated_by field to the current user during update operations.
// If hasSet is true, it adds a SET clause to the query; otherwise, it adds a value to the query.
func (ub *UpdatedByHandler) OnUpdate(query *bun.UpdateQuery, hasSet bool, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	name := ub.Name()

	if hasSet {
		if query.DB().HasFeature(feature.UpdateMultiTable) {
			query.Set("?TableAlias.? = "+constants.ExprOperator, bun.Ident(name))
		} else {
			query.Set("? = "+constants.ExprOperator, bun.Ident(name))
		}
	} else {
		query.Value(name, constants.ExprOperator)
	}

	query.Returning("?", bun.Name(name))
}

// OnInsert sets the updated_by field to the current user during insert operations.
func (ub *UpdatedByHandler) OnInsert(query *bun.InsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	if value.IsZero() {
		query.Value(ub.Name(), constants.ExprOperator)
	}
}

// Name returns the column name for the updated_by field.
func (*UpdatedByHandler) Name() string {
	return constants.ColumnUpdatedBy
}
