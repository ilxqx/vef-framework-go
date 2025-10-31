package apis

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// validateColumnsExist validates that the specified columns exist in the model schema.
// This is a helper function to ensure column mappings are valid before query execution.
func validateColumnsExist(schema *schema.Table, columns ...struct {
	name   string
	column string
},
) error {
	for _, c := range columns {
		if c.column != constants.Empty {
			if !schema.HasField(c.column) {
				return result.Err(i18n.T("field_not_exist_in_model", map[string]any{
					"field": c.column,
					"name":  c.name,
					"model": schema.TypeName,
				}))
			}
		}
	}

	return nil
}

// validateOptionColumns validates columns for DataOptionColumnMapping.
// Ensures that all specified column names exist in the database model schema.
func validateOptionColumns(schema *schema.Table, mapping *DataOptionColumnMapping) error {
	columns := []struct {
		name   string
		column string
	}{
		{"labelColumn", mapping.LabelColumn},
		{"valueColumn", mapping.ValueColumn},
	}

	if mapping.DescriptionColumn != constants.Empty {
		columns = append(columns, struct {
			name   string
			column string
		}{"descriptionColumn", mapping.DescriptionColumn})
	}

	return validateColumnsExist(schema, columns...)
}

// mergeOptionColumnMapping merges the provided mapping with default mapping.
// Uses fallback values for empty columns based on the provided default mapping or system defaults.
func mergeOptionColumnMapping(mapping, defaultMapping *DataOptionColumnMapping) {
	if mapping.LabelColumn == constants.Empty {
		mapping.LabelColumn = lo.CoalesceOrEmpty(defaultMapping.LabelColumn, defaultLabelColumn)
	}

	if mapping.ValueColumn == constants.Empty {
		mapping.ValueColumn = lo.CoalesceOrEmpty(defaultMapping.ValueColumn, defaultValueColumn)
	}

	if mapping.DescriptionColumn == constants.Empty {
		mapping.DescriptionColumn = defaultMapping.DescriptionColumn
	}

	// Merge MetaColumns if not specified
	if len(mapping.MetaColumns) == 0 && defaultMapping != nil {
		mapping.MetaColumns = defaultMapping.MetaColumns
	}
}

// ApplyDataPermission is a helper function that applies data permission filtering to a SelectQuery.
// Returns an error if data permission application fails.
func ApplyDataPermission(query orm.SelectQuery, ctx fiber.Ctx) error {
	if applier := contextx.DataPermApplier(ctx); applier != nil {
		if err := applier.Apply(query); err != nil {
			return fmt.Errorf("failed to apply data permission: %w", err)
		}
	}

	return nil
}

// GetAuditUserNameRelations returns RelationSpecs for creator and updater joins.
func GetAuditUserNameRelations(userModel any, nameColumn ...string) []*orm.RelationSpec {
	nc := defaultAuditUserNameColumn
	if len(nameColumn) > 0 {
		nc = nameColumn[0]
	}

	// Create RelationSpecs for creator and updater
	relations := []*orm.RelationSpec{
		{
			Model:         userModel,
			Alias:         "creator",
			JoinType:      orm.LeftJoin,
			ForeignColumn: "created_by",
			SelectedColumns: []orm.ColumnInfo{
				{
					Name:  nc,
					Alias: constants.ColumnCreatedByName,
				},
			},
		},
		{
			Model:         userModel,
			Alias:         "updater",
			JoinType:      orm.LeftJoin,
			ForeignColumn: "updated_by",
			SelectedColumns: []orm.ColumnInfo{
				{
					Name:  nc,
					Alias: constants.ColumnUpdatedByName,
				},
			},
		},
	}

	return relations
}

// columnAliasPattern matches "column AS alias" format (case-insensitive AS, flexible spaces).
var columnAliasPattern = regexp.MustCompile(`^\s*(.+?)\s+(?i:as)\s+(.+?)\s*$`)

// parseMetaColumn parses a single meta column specification into (column, alias).
// Supports formats:
//   - "column" -> ("column", "column")
//   - "column AS alias" -> ("column", "alias")
//   - "column as alias" -> ("column", "alias")
func parseMetaColumn(spec string) (column, alias string) {
	if matches := columnAliasPattern.FindStringSubmatch(spec); len(matches) == 3 {
		column = strings.TrimSpace(matches[1])
		alias = strings.TrimSpace(matches[2])

		return column, alias
	}

	// No alias specified, use column name as alias
	trimmed := strings.TrimSpace(spec)

	return trimmed, trimmed
}

// parseMetaColumns parses meta column specifications into structured info.
// This function should be called once to avoid redundant parsing.
// Returns nil if specs is empty.
func parseMetaColumns(specs []string) []orm.ColumnInfo {
	if len(specs) == 0 {
		return nil
	}

	result := make([]orm.ColumnInfo, len(specs))
	for i, spec := range specs {
		columnName, aliasName := parseMetaColumn(spec)
		result[i] = orm.ColumnInfo{
			Name:  columnName,
			Alias: aliasName,
		}
	}

	return result
}

// validateMetaColumns validates that all meta columns exist in the table schema.
func validateMetaColumns(schema *schema.Table, metaColumns []orm.ColumnInfo) error {
	for _, col := range metaColumns {
		if !schema.HasField(col.Name) {
			return result.Err(i18n.T("field_not_exist_in_model", map[string]any{
				"field": col.Name,
				"name":  "metaColumns",
				"model": schema.TypeName,
			}))
		}
	}

	return nil
}

// buildMetaJsonExpr constructs a JSON_OBJECT expression for meta columns.
func buildMetaJsonExpr(eb orm.ExprBuilder, metaColumns []orm.ColumnInfo) schema.QueryAppender {
	// Build JSON_OBJECT arguments: key1, value1, key2, value2, ...
	jsonArgs := make([]any, 0, len(metaColumns)*2)
	for _, col := range metaColumns {
		// Add key-value pair: alias as key, column expression as value
		jsonArgs = append(jsonArgs, col.Alias, eb.Column(col.Name))
	}

	return eb.JsonObject(jsonArgs...)
}
