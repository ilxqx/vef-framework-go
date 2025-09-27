package orm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/sort"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// getTableSchema extracts the table schema from a struct pointer model.
// It validates that the model is a pointer to a struct and returns the corresponding schema.
func getTableSchema(model any, db *bun.DB) *schema.Table {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Pointer {
		modelType = modelType.Elem()
		if modelType.Kind() == reflect.Struct {
			return db.Table(modelType)
		}
	}

	panic(
		fmt.Sprintf("model must be a struct pointer, got %T", model),
	)
}

// getTableSchemaFromQuery extracts the table schema from a bun.Query instance.
// It returns the schema if the query has a model that implements bun.TableModel, otherwise nil.
func getTableSchemaFromQuery(query bun.Query) *schema.Table {
	if model := query.GetModel(); model != nil {
		if tm, ok := model.(bun.TableModel); ok {
			return tm.Table()
		}
	}

	return nil
}

// buildColumnExpr builds a column expression with optional table alias.
// If no alias is provided, it uses the default ?TableAlias placeholder.
func buildColumnExpr(column string, alias ...string) schema.QueryWithArgs {
	if len(alias) == 0 {
		return bun.SafeQuery("?TableAlias.?", bun.Name(column))
	}

	return bun.SafeQuery("?.?", bun.Name(alias[0]), bun.Name(column))
}

// applyModelRelation applies a model relation to a SelectQuery by creating the appropriate LEFT JOIN.
// It automatically determines foreign and referenced columns based on table schema and conventions.
func applyModelRelation(relation ModelRelation, query SelectQuery) {
	var (
		table            = getTableSchema(relation.Model, query.(*BunSelectQuery).query.DB())
		expr             strings.Builder
		foreignColumn    = relation.ForeignColumn
		referencedColumn = relation.ReferencedColumn
		pk               = constants.ColumnId
	)

	_, _ = expr.WriteString(constants.ExprTableAlias)
	_, _ = expr.WriteString(".? = ?.?")

	if len(table.PKs) > 0 {
		pk = table.PKs[0].Name
	}
	if foreignColumn == constants.Empty {
		foreignColumn = table.ModelName + constants.Underscore + pk
	}
	if referencedColumn == constants.Empty {
		referencedColumn = pk
	}

	query.LeftJoin(
		relation.Model,
		func(cb ConditionBuilder) {
			if relation.ForeignColumn == constants.Empty {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.Expr(expr.String(), bun.Name(foreignColumn), bun.Name(table.Alias), bun.Name(referencedColumn))
				})
			}
			if relation.On != nil {
				relation.On(cb)
			}
		},
	)
}

// applySort applies the sort orders to the query.
func applySort(orders []sort.OrderSpec, query SelectQuery) {
	for _, order := range orders {
		if !order.IsValid() {
			continue
		}

		query.OrderByExpr(func(eb ExprBuilder) any {
			return eb.Order(func(ob OrderBuilder) {
				ob.Column(order.Column)
				if order.Direction == sort.OrderDesc {
					ob.Desc()
				} else {
					ob.Asc()
				}

				switch order.NullsOrder {
				case sort.NullsFirst:
					ob.NullsFirst()
				case sort.NullsLast:
					ob.NullsLast()
				}
			})
		})
	}
}
