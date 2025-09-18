package apis

import (
	"fmt"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/samber/lo"
	"github.com/uptrace/bun/schema"
)

// validateConfigFields validates that the specified fields exist in the model schema.
// This is a helper function to ensure field configurations are valid before query execution.
func validateConfigFields(schema *schema.Table, fields ...struct {
	name  string
	field string
}) error {
	for _, f := range fields {
		if f.field != constants.Empty {
			if field, _ := schema.Field(f.field); field == nil {
				return fmt.Errorf("field '%s' specified in %s does not exist in model", f.field, f.name)
			}
		}
	}
	return nil
}

// validateOptionsFields validates fields for OptionsConfig.
// Ensures that all specified field names exist in the database model schema.
func validateOptionsFields(schema *schema.Table, config *OptionsConfig) error {
	fields := []struct {
		name  string
		field string
	}{
		{"labelField", config.LabelField},
		{"valueField", config.ValueField},
	}

	if config.DescriptionField != constants.Empty {
		fields = append(fields, struct {
			name  string
			field string
		}{"descriptionField", config.DescriptionField})
	}

	if config.SortField != constants.Empty {
		fields = append(fields, struct {
			name  string
			field string
		}{"sortField", config.SortField})
	}

	return validateConfigFields(schema, fields...)
}

// validateTreeOptionsFields validates fields for TreeOptionsConfig.
// Ensures that all specified field names exist in the database model schema.
func validateTreeOptionsFields(schema *schema.Table, config *TreeOptionsConfig) error {
	fields := []struct {
		name  string
		field string
	}{
		{"labelField", config.LabelField},
		{"valueField", config.ValueField},
		{"idField", config.IdField},
		{"parentIdField", config.ParentIdField},
	}

	if config.DescriptionField != constants.Empty {
		fields = append(fields, struct {
			name  string
			field string
		}{"descriptionField", config.DescriptionField})
	}

	if config.SortField != constants.Empty {
		fields = append(fields, struct {
			name  string
			field string
		}{"sortField", config.SortField})
	}

	return validateConfigFields(schema, fields...)
}

// applyOptionsDefaults applies default values to OptionsConfig.
// Uses fallback values for empty fields based on the provided default configuration or system defaults.
func applyOptionsDefaults(config *OptionsConfig, defaultConfig *OptionsConfig) {
	if config.LabelField == constants.Empty {
		config.LabelField = lo.CoalesceOrEmpty(defaultConfig.LabelField, defaultLabelField)
	}
	if config.ValueField == constants.Empty {
		config.ValueField = lo.CoalesceOrEmpty(defaultConfig.ValueField, defaultValueField)
	}
	if config.DescriptionField == constants.Empty {
		config.DescriptionField = defaultConfig.DescriptionField
	}
	if config.SortField == constants.Empty {
		config.SortField = defaultConfig.SortField
	}
}

// applyTreeOptionsDefaults applies default values to TreeOptionsConfig.
// Uses fallback values for empty fields based on the provided default configuration or system defaults.
func applyTreeOptionsDefaults(config *TreeOptionsConfig, defaultConfig *TreeOptionsConfig) {
	if config.LabelField == constants.Empty {
		config.LabelField = lo.CoalesceOrEmpty(defaultConfig.LabelField, defaultLabelField)
	}
	if config.ValueField == constants.Empty {
		config.ValueField = lo.CoalesceOrEmpty(defaultConfig.ValueField, defaultValueField)
	}
	if config.DescriptionField == constants.Empty {
		config.DescriptionField = defaultConfig.DescriptionField
	}
	if config.SortField == constants.Empty {
		config.SortField = defaultConfig.SortField
	}
	if config.IdField == constants.Empty {
		config.IdField = lo.CoalesceOrEmpty(defaultConfig.IdField, idField)
	}
	if config.ParentIdField == constants.Empty {
		config.ParentIdField = lo.CoalesceOrEmpty(defaultConfig.ParentIdField, parentIdField)
	}
}
