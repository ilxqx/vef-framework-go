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

// validateOptionColumns validates columns for OptionColumnMapping.
// Ensures that all specified column names exist in the database model schema.
func validateOptionColumns(schema *schema.Table, mapping *OptionColumnMapping) error {
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

	if mapping.SortColumn != constants.Empty {
		columns = append(columns, struct {
			name   string
			column string
		}{"sortColumn", mapping.SortColumn})
	}

	return validateColumnsExist(schema, columns...)
}

// validateTreeOptionColumns validates columns for TreeOptionColumnMapping.
// Ensures that all specified column names exist in the database model schema.
func validateTreeOptionColumns(schema *schema.Table, mapping *TreeOptionColumnMapping) error {
	columns := []struct {
		name   string
		column string
	}{
		{"labelColumn", mapping.LabelColumn},
		{"valueColumn", mapping.ValueColumn},
		{"idColumn", mapping.IdColumn},
		{"parentIdColumn", mapping.ParentIdColumn},
	}

	if mapping.DescriptionColumn != constants.Empty {
		columns = append(columns, struct {
			name   string
			column string
		}{"descriptionColumn", mapping.DescriptionColumn})
	}

	if mapping.SortColumn != constants.Empty {
		columns = append(columns, struct {
			name   string
			column string
		}{"sortColumn", mapping.SortColumn})
	}

	return validateColumnsExist(schema, columns...)
}

// mergeOptionColumnMapping merges the provided mapping with default mapping.
// Uses fallback values for empty columns based on the provided default mapping or system defaults.
func mergeOptionColumnMapping(mapping, defaultMapping *OptionColumnMapping) {
	if mapping.LabelColumn == constants.Empty {
		mapping.LabelColumn = lo.CoalesceOrEmpty(defaultMapping.LabelColumn, defaultLabelColumn)
	}

	if mapping.ValueColumn == constants.Empty {
		mapping.ValueColumn = lo.CoalesceOrEmpty(defaultMapping.ValueColumn, defaultValueColumn)
	}

	if mapping.DescriptionColumn == constants.Empty {
		mapping.DescriptionColumn = defaultMapping.DescriptionColumn
	}

	if mapping.SortColumn == constants.Empty {
		mapping.SortColumn = defaultMapping.SortColumn
	}
}

// mergeTreeOptionColumnMapping merges the provided mapping with default mapping.
// Uses fallback values for empty columns based on the provided default mapping or system defaults.
func mergeTreeOptionColumnMapping(mapping, defaultMapping *TreeOptionColumnMapping) {
	if mapping.LabelColumn == constants.Empty {
		mapping.LabelColumn = lo.CoalesceOrEmpty(defaultMapping.LabelColumn, defaultLabelColumn)
	}

	if mapping.ValueColumn == constants.Empty {
		mapping.ValueColumn = lo.CoalesceOrEmpty(defaultMapping.ValueColumn, defaultValueColumn)
	}

	if mapping.DescriptionColumn == constants.Empty {
		mapping.DescriptionColumn = defaultMapping.DescriptionColumn
	}

	if mapping.SortColumn == constants.Empty {
		mapping.SortColumn = defaultMapping.SortColumn
	}

	if mapping.IdColumn == constants.Empty {
		mapping.IdColumn = lo.CoalesceOrEmpty(defaultMapping.IdColumn, idColumn)
	}

	if mapping.ParentIdColumn == constants.Empty {
		mapping.ParentIdColumn = lo.CoalesceOrEmpty(defaultMapping.ParentIdColumn, parentIdColumn)
	}
}

// applyDataPermission is a helper function that applies data permission filtering to a SelectQuery.
// This function is designed to be reused across different Api types (Update, Delete, etc.).
// The caller is responsible for checking if data permission should be applied.
// Returns an error if data permission application fails.
func applyDataPermission(query orm.SelectQuery, ctx fiber.Ctx) error {
	applier := contextx.DataPermApplier(ctx)
	if applier == nil {
		return nil
	}

	if err := applier.Apply(query); err != nil {
		return fmt.Errorf("failed to apply data permission: %w", err)
	}

	return nil
}
