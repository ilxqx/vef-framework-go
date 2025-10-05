package apis

import (
	"github.com/samber/lo"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/result"
)

// validateFieldsExist validates that the specified fields exist in the model schema.
// This is a helper function to ensure field mappings are valid before query execution.
func validateFieldsExist(schema *schema.Table, fields ...struct {
	name  string
	field string
},
) error {
	for _, f := range fields {
		if f.field != constants.Empty {
			if field, _ := schema.Field(f.field); field == nil {
				return result.Err(i18n.T("field_not_exist_in_model", map[string]any{
					"field": f.field,
					"name":  f.name,
					"model": schema.TypeName,
				}))
			}
		}
	}

	return nil
}

// validateOptionFields validates fields for OptionFieldMapping.
// Ensures that all specified field names exist in the database model schema.
func validateOptionFields(schema *schema.Table, mapping *OptionFieldMapping) error {
	fields := []struct {
		name  string
		field string
	}{
		{"labelField", mapping.LabelField},
		{"valueField", mapping.ValueField},
	}

	if mapping.DescriptionField != constants.Empty {
		fields = append(fields, struct {
			name  string
			field string
		}{"descriptionField", mapping.DescriptionField})
	}

	if mapping.SortField != constants.Empty {
		fields = append(fields, struct {
			name  string
			field string
		}{"sortField", mapping.SortField})
	}

	return validateFieldsExist(schema, fields...)
}

// validateTreeOptionFields validates fields for TreeOptionFieldMapping.
// Ensures that all specified field names exist in the database model schema.
func validateTreeOptionFields(schema *schema.Table, mapping *TreeOptionFieldMapping) error {
	fields := []struct {
		name  string
		field string
	}{
		{"labelField", mapping.LabelField},
		{"valueField", mapping.ValueField},
		{"idField", mapping.IdField},
		{"parentIdField", mapping.ParentIdField},
	}

	if mapping.DescriptionField != constants.Empty {
		fields = append(fields, struct {
			name  string
			field string
		}{"descriptionField", mapping.DescriptionField})
	}

	if mapping.SortField != constants.Empty {
		fields = append(fields, struct {
			name  string
			field string
		}{"sortField", mapping.SortField})
	}

	return validateFieldsExist(schema, fields...)
}

// mergeOptionFieldMapping merges the provided mapping with default mapping.
// Uses fallback values for empty fields based on the provided default mapping or system defaults.
func mergeOptionFieldMapping(mapping, defaultMapping *OptionFieldMapping) {
	if mapping.LabelField == constants.Empty {
		mapping.LabelField = lo.CoalesceOrEmpty(defaultMapping.LabelField, defaultLabelField)
	}

	if mapping.ValueField == constants.Empty {
		mapping.ValueField = lo.CoalesceOrEmpty(defaultMapping.ValueField, defaultValueField)
	}

	if mapping.DescriptionField == constants.Empty {
		mapping.DescriptionField = defaultMapping.DescriptionField
	}

	if mapping.SortField == constants.Empty {
		mapping.SortField = defaultMapping.SortField
	}
}

// mergeTreeOptionFieldMapping merges the provided mapping with default mapping.
// Uses fallback values for empty fields based on the provided default mapping or system defaults.
func mergeTreeOptionFieldMapping(mapping, defaultMapping *TreeOptionFieldMapping) {
	if mapping.LabelField == constants.Empty {
		mapping.LabelField = lo.CoalesceOrEmpty(defaultMapping.LabelField, defaultLabelField)
	}

	if mapping.ValueField == constants.Empty {
		mapping.ValueField = lo.CoalesceOrEmpty(defaultMapping.ValueField, defaultValueField)
	}

	if mapping.DescriptionField == constants.Empty {
		mapping.DescriptionField = defaultMapping.DescriptionField
	}

	if mapping.SortField == constants.Empty {
		mapping.SortField = defaultMapping.SortField
	}

	if mapping.IdField == constants.Empty {
		mapping.IdField = lo.CoalesceOrEmpty(defaultMapping.IdField, idField)
	}

	if mapping.ParentIdField == constants.Empty {
		mapping.ParentIdField = lo.CoalesceOrEmpty(defaultMapping.ParentIdField, parentIdField)
	}
}
