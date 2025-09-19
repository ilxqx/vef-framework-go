package orm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// getTableSchema get table schema from struct pointer
func getTableSchema(model any, db *bun.DB) *orm.Table {
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

// getTableSchemaFromQuery get table schema from query
func getTableSchemaFromQuery(query bun.Query) *orm.Table {
	if model := query.GetModel(); model != nil {
		if tm, ok := model.(bun.TableModel); ok {
			return tm.Table()
		}
	}

	return nil
}

// parseColumnExpr parses the column expression for the given column.
func parseColumnExpr(column string) schema.QueryWithArgs {
	dotIndex := strings.IndexByte(column, constants.ByteDot)
	if dotIndex > -1 {
		alias, name := column[:dotIndex], column[dotIndex+1:]
		if strings.IndexByte(alias, constants.ByteQuestionMark) > -1 {
			return bun.SafeQuery(alias+".?", bun.Name(name))
		} else {
			return bun.SafeQuery("?.?", bun.Name(alias), bun.Name(name))
		}
	} else {
		return bun.SafeQuery("?TableAlias.?", bun.Name(column))
	}
}

// buildColumnExpr builds a column expression for the given column and alias.
func buildColumnExpr(column string, alias ...string) schema.QueryWithArgs {
	if len(alias) == 0 {
		return bun.SafeQuery("?TableAlias.?", bun.Name(column))
	}

	return bun.SafeQuery("?.?", bun.Name(alias[0]), bun.Name(column))
}

// applyModelRelation applies a model relation to the query.
func applyModelRelation(relation orm.ModelRelation, query orm.Query) {
	var (
		table            = getTableSchema(relation.Model, query.(*bunQuery).query.DB()) // table is the schema of the related model
		expr             strings.Builder                                                // expr builds the join expression
		foreignColumn    = relation.ForeignColumn                                       // foreignColumn is the foreign key column name
		referencedColumn = relation.ReferencedColumn                                    // referencedColumn is the referenced column name
		pk               = orm.ColumnId                                                 // pk is the primary key column name
	)

	_, _ = expr.WriteString(orm.ExprTableAlias)
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
		func(cb orm.ConditionBuilder) {
			if relation.ForeignColumn == constants.Empty {
				cb.Expr(expr.String(), bun.Name(foreignColumn), bun.Name(table.Alias), bun.Name(referencedColumn))
			}
			if relation.On != nil {
				relation.On(cb)
			}
		},
	)
}
