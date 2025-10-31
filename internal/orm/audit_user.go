package orm

import (
	"reflect"

	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

// CreatedByHandler implements InsertHandler for automatically setting created_by user information.
type CreatedByHandler struct{}

func (cb *CreatedByHandler) OnInsert(query *BunInsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	if value.IsZero() {
		name := cb.Name()
		query.ColumnExpr(name, func(eb ExprBuilder) any {
			return eb.Expr(constants.ExprOperator)
		})
	}
}

func (*CreatedByHandler) Name() string {
	return constants.ColumnCreatedBy
}

// UpdatedByHandler implements UpdateHandler for automatically managing updated_by user information.
type UpdatedByHandler struct{}

func (ub *UpdatedByHandler) OnUpdate(query *BunUpdateQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	name := ub.Name()

	if query.hasSet {
		query.SetExpr(name, func(eb ExprBuilder) any {
			return eb.Expr(constants.ExprOperator)
		})
	} else {
		query.ColumnExpr(name, func(eb ExprBuilder) any {
			return eb.Expr(constants.ExprOperator)
		})
	}
}

func (ub *UpdatedByHandler) OnInsert(query *BunInsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value) {
	if value.IsZero() {
		name := ub.Name()
		query.ColumnExpr(name, func(eb ExprBuilder) any {
			return eb.Expr(constants.ExprOperator)
		})
	}
}

// Name returns the column name for the updated_by field.
func (*UpdatedByHandler) Name() string {
	return constants.ColumnUpdatedBy
}
