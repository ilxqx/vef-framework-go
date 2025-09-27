package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type findOptionsAPI[TModel, TSearch any] struct {
	FindAPI[TModel, TSearch, []Option, FindOptionsAPI[TModel, TSearch]]

	defaultConfig *OptionsConfig
}

func (a *findOptionsAPI[TModel, TSearch]) Provide() api.Spec {
	return a.FindAPI.Build(a.findOptions)
}

// Build should not be called directly on concrete API types.
// Use Provide() to generate api.Spec with the correct handler instead.
func (a *findOptionsAPI[TModel, TSearch]) Build(handler any) api.Spec {
	panic("apis: do not call FindAPI.Build on findOptionsAPI; call Provide() instead")
}

func (a *findOptionsAPI[TModel, TSearch]) DefaultConfig(config *OptionsConfig) FindOptionsAPI[TModel, TSearch] {
	a.defaultConfig = config
	return a
}

func (a *findOptionsAPI[TModel, TSearch]) findOptions(db orm.Db) func(ctx fiber.Ctx, db orm.Db, config OptionsConfig, search TSearch) error {
	// Pre-compute schema information
	schema := db.Schema((*TModel)(nil))

	// Pre-compute whether default created_at ordering should be applied
	hasCreatedAt := schema.HasField(constants.ColumnCreatedAt)
	shouldApplyDefaultSort := !a.HasSortApplier() && hasCreatedAt

	return func(ctx fiber.Ctx, db orm.Db, config OptionsConfig, search TSearch) error {
		var options []Option
		query := a.BuildQuery(db, (*TModel)(nil), search, ctx)

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
		} else if shouldApplyDefaultSort {
			query.OrderBy(constants.ColumnCreatedAt)
		}

		// Execute query with limit
		if err := query.Limit(maxOptionsLimit).Scan(ctx, &options); err != nil {
			return err
		}

		return result.Ok(a.Process(options, search, ctx)).Response(ctx)
	}
}
