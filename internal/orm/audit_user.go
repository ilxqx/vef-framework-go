package orm

import (
	"reflect"

	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

// operatorExprBuilder returns the current operator expression for audit columns.
func operatorExprBuilder(eb ExprBuilder) any {
	return eb.Expr(constants.ExprOperator)
}

// CreatedByHandler implements InsertHandler for automatically setting created_by user information.
type CreatedByHandler struct{}

func (cb *CreatedByHandler) OnInsert(query *BunInsertQuery, _ *schema.Table, _ *schema.Field, _ any, value reflect.Value) {
	if value.IsZero() {
		query.ColumnExpr(cb.Name(), operatorExprBuilder)
	}
}

func (*CreatedByHandler) Name() string {
	return constants.ColumnCreatedBy
}

// UpdatedByHandler implements UpdateHandler for automatically managing updated_by user information.
type UpdatedByHandler struct{}

func (ub *UpdatedByHandler) OnUpdate(query *BunUpdateQuery, _ *schema.Table, _ *schema.Field, _ any, _ reflect.Value) {
	name := ub.Name()

	if query.hasSet {
		query.SetExpr(name, operatorExprBuilder)
	} else {
		query.ColumnExpr(name, operatorExprBuilder)
	}
}

func (ub *UpdatedByHandler) OnInsert(query *BunInsertQuery, _ *schema.Table, _ *schema.Field, _ any, value reflect.Value) {
	if value.IsZero() {
		query.ColumnExpr(ub.Name(), operatorExprBuilder)
	}
}

// Name returns the column name for the updated_by field.
func (*UpdatedByHandler) Name() string {
	return constants.ColumnUpdatedBy
}
