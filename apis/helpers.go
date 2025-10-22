package apis

import (
	"fmt"

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
			if field, _ := schema.Field(c.column); field == nil {
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
