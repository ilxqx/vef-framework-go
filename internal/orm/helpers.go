package orm

import (
	"reflect"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	collections "github.com/ilxqx/go-collections"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/dbhelpers"
	"github.com/ilxqx/vef-framework-go/sortx"
)

// getTableSchema extracts the table schema from a struct pointer model.
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
func getTableSchemaFromQuery(query bun.Query) *schema.Table {
	if model := query.GetModel(); model != nil {
		if tm, ok := model.(bun.TableModel); ok {
			return tm.Table()
		}
	}

	return nil
}

// buildColumnExpr builds a column expression with optional table alias.
func buildColumnExpr(column string, alias ...string) schema.QueryWithArgs {
	if len(alias) == 0 {
		return bun.SafeQuery("?TableAlias.?", bun.Name(column))
	}

	return bun.SafeQuery("?.?", bun.Name(alias[0]), bun.Name(column))
}

// applyRelationSpec applies a RelationSpec to a SelectQuery by creating the appropriate JOIN.
func applyRelationSpec(spec *RelationSpec, query SelectQuery) {
	var (
		table            = query.DB().TableOf(spec.Model)
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

	joinCondition := func(cb ConditionBuilder) {
		cb.EqualsColumn(dbhelpers.ColumnWithAlias(referencedColumn, alias), foreignColumn)

		if spec.On != nil {
			spec.On(cb)
		}
	}

	switch joinType {
	case JoinInner:
		query.Join(spec.Model, joinCondition, alias)
	case JoinLeft:
		query.LeftJoin(spec.Model, joinCondition, alias)
	case JoinRight:
		query.RightJoin(spec.Model, joinCondition, alias)
	case JoinFull:
		query.FullJoin(spec.Model, joinCondition, alias)
	case JoinCross:
		logger.Panic("applyRelationSpec: CROSS JOIN is not supported in RelationSpec, use query.CrossJoin() directly")
	default:
		logger.Panicf("applyRelationSpec: unsupported join type %v", joinType)
	}
}

// ApplySort applies the sort orders to the query.
func ApplySort(query SelectQuery, orders []sortx.OrderSpec) {
	for _, order := range orders {
		if !order.IsValid() {
			continue
		}

		query.OrderByExpr(func(eb ExprBuilder) any {
			return eb.Order(func(ob OrderBuilder) {
				ob.Column(order.Column)

				switch order.Direction {
				case sortx.OrderAsc:
					ob.Asc()
				case sortx.OrderDesc:
					ob.Desc()
				}

				switch order.NullsOrder {
				case sortx.NullsFirst:
					ob.NullsFirst()
				case sortx.NullsLast:
					ob.NullsLast()
				}
			})
		})
	}
}

func buildReturningExpr(returningColumns collections.Set[string], eb ExprBuilder) schema.QueryAppender {
	columns := make([]any, 0, returningColumns.Size())

	for column := range returningColumns.Seq() {
		switch column {
		case sqlNull:
			columns = append(columns, bun.Safe(sqlNull))
		case columnAll:
			columns = append(columns, bun.Safe(columnAll))
		default:
			columns = append(columns, bun.Ident(column))
		}
	}

	return eb.Exprs(columns...)
}
