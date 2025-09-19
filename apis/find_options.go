package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/uptrace/bun/schema"
)

const (
	// maxOptionsLimit is the maximum number of options that can be returned in a single query
	maxOptionsLimit = 10000
	// defaultLabelField is the default field name for option labels
	defaultLabelField = "name"
	// defaultValueField is the default field name for option values
	defaultValueField = orm.ColumnId
	labelField        = "label"
	valueField        = "value"
	descriptionField  = "description"
)

// Option represents a simple option with label and value.
type Option struct {
	Label       string         `json:"label"`                 // Display text for the option
	Value       string         `json:"value"`                 // Actual value of the option
	Description string         `json:"description,omitempty"` // Optional description
	Meta        map[string]any `json:"meta,omitempty"`        // Optional meta data
}

// OptionsConfig defines the mapping between database fields and option fields.
type OptionsConfig struct {
	api.In
	LabelField       string `json:"labelField"`       // Field name for label (default: "name")
	ValueField       string `json:"valueField"`       // Field name for value (default: "id")
	DescriptionField string `json:"descriptionField"` // Field name for description
	SortField        string `json:"sortField"`        // Field name for sorting
}

// applyDefaults applies default values to options configuration.
func (c *OptionsConfig) applyDefaults(defaultConfig *OptionsConfig) {
	applyOptionsDefaults(c, defaultConfig)
}

// validateFields validates that the specified fields exist in the model.
func (c *OptionsConfig) validateFields(schema *schema.Table) error {
	return validateOptionsFields(schema, c)
}

// FindOptionsAPI provides option query functionality with customizable field mapping.
type FindOptionsAPI[TModel, TSearch any] struct {
	*findAPI[TModel, TSearch, PostFindProcessor[[]Option, []Option], FindOptionsAPI[TModel, TSearch]]
	defaultConfig *OptionsConfig
}

// WithDefaultConfig sets the default configuration for options queries.
// This configuration provides fallback values for field mapping when not explicitly specified in queries.
// Returns the API instance for method chaining.
func (a *FindOptionsAPI[TModel, TSearch]) WithDefaultConfig(config *OptionsConfig) *FindOptionsAPI[TModel, TSearch] {
	a.defaultConfig = config
	return a
}

// FindOptions creates a handler that executes the query and returns options with customizable configuration.
//
// Parameters:
//   - db: The database connection for schema introspection
//
// Returns a handler function that processes find-options requests.
func (a *FindOptionsAPI[TModel, TSearch]) FindOptions(db orm.Db) func(ctx fiber.Ctx, db orm.Db, config OptionsConfig, search TSearch) error {
	// Pre-compute schema information
	schema := db.Schema((*TModel)(nil))

	// Pre-compute whether default created_at ordering should be applied
	hasCreatedAt := schema.HasField(orm.ColumnCreatedAt)

	return func(ctx fiber.Ctx, db orm.Db, config OptionsConfig, search TSearch) error {
		var options []Option
		query := a.buildQuery(ctx, db, (*TModel)(nil), search)

		// Apply defaults and validate configuration
		config.applyDefaults(a.defaultConfig)
		if err := config.validateFields(schema); err != nil {
			return err
		}

		// Select only required fields
		if config.ValueField == valueField {
			query.Select(config.ValueField)
		} else {
			query.SelectAs(config.ValueField, valueField)
		}

		if config.LabelField == labelField {
			query.Select(config.LabelField)
		} else {
			query.SelectAs(config.LabelField, labelField)
		}

		if config.DescriptionField != constants.Empty {
			if config.DescriptionField == descriptionField {
				query.Select(config.DescriptionField)
			} else {
				query.SelectAs(config.DescriptionField, descriptionField)
			}
		}

		// Apply sorting
		if config.SortField != constants.Empty {
			query.OrderBy(config.SortField)
		} else if a.sortApplier == nil && hasCreatedAt {
			query.OrderBy(orm.ColumnCreatedAt)
		}

		// Execute query with limit
		if err := query.Limit(maxOptionsLimit).Scan(ctx, &options); err != nil {
			return err
		}

		if a.processor != nil {
			options = a.processor(options, ctx)
		}

		return result.Ok(options).Response(ctx)
	}
}

// NewFindOptionsAPI creates a new FindOptionsAPI with the specified options.
func NewFindOptionsAPI[TModel, TSearch any]() *FindOptionsAPI[TModel, TSearch] {
	api := &FindOptionsAPI[TModel, TSearch]{
		defaultConfig: &OptionsConfig{
			LabelField: defaultLabelField,
			ValueField: defaultValueField,
		},
	}
	api.findAPI = newFindAPI[TModel, TSearch, PostFindProcessor[[]Option, []Option]](api)

	return api
}
