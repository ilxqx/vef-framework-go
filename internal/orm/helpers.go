package orm

import (
	"reflect"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/dbhelpers"
	"github.com/ilxqx/vef-framework-go/sort"
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

	logger.Panicf("model must be a struct pointer, got %T", model)

	return nil
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

// applyRelationSpec applies a RelationSpec to a SelectQuery by creating the appropriate JOIN.
// It automatically determines foreign and referenced columns based on table schema and conventions.
func applyRelationSpec(spec RelationSpec, query SelectQuery) {
	var (
		table            = query.Db().TableOf(spec.Model)
		pk               string
		alias            = spec.Alias
		joinType         = spec.JoinType
		foreignColumn    = spec.ForeignColumn
		referencedColumn = spec.ReferencedColumn
	)

	if len(table.PKs) != 1 {
		logger.Panicf("applyRelationSpec: model %q requires exactly one primary key, got %d primary key(s)", table.TypeName, len(table.PKs))
	}

	pk = table.PKs[0].Name

	if alias == constants.Empty {
		alias = table.Alias
	}

	// Default to LEFT JOIN if not specified
	if joinType == JoinDefault {
		joinType = JoinLeft
	}

	if foreignColumn == constants.Empty {
		foreignColumn = table.ModelName + constants.Underscore + pk
	}

	if referencedColumn == constants.Empty {
		referencedColumn = pk
	}

	// Select specified columns from the joined table
	if len(spec.SelectedColumns) > 0 {
		for _, ci := range spec.SelectedColumns {
			column := dbhelpers.ColumnWithAlias(ci.Name, alias)

			columnAlias := ci.Alias
			if ci.AutoAlias {
				columnAlias = table.ModelName + constants.Underscore + ci.Name
			}

			if columnAlias != constants.Empty {
				query.SelectAs(column, columnAlias)
			} else {
				query.Select(column)
			}
		}
	}

	// Build the JOIN condition
	joinCondition := func(cb ConditionBuilder) {
		cb.EqualsColumn(dbhelpers.ColumnWithAlias(referencedColumn, alias), foreignColumn)

		// Apply additional custom ON conditions if provided
		if spec.On != nil {
			spec.On(cb)
		}
	}

	// Apply the appropriate JOIN type
	switch joinType {
	case JoinInner:
		query.Join(spec.Model, joinCondition, alias)
	case JoinLeft:
		query.LeftJoin(spec.Model, joinCondition, alias)
	case JoinRight:
		query.RightJoin(spec.Model, joinCondition, alias)
	}
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
