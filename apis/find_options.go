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

	fieldMapping *OptionFieldMapping
}

func (a *findOptionsAPI[TModel, TSearch]) Provide() api.Spec {
	return a.Build(a.findOptions)
}

func (a *findOptionsAPI[TModel, TSearch]) FieldMapping(mapping *OptionFieldMapping) FindOptionsAPI[TModel, TSearch] {
	a.fieldMapping = mapping

	return a
}

func (a *findOptionsAPI[TModel, TSearch]) findOptions(db orm.Db) func(ctx fiber.Ctx, db orm.Db, params OptionParams, search TSearch) error {
	// Pre-compute schema information
	schema := db.TableOf((*TModel)(nil))

	// Pre-compute whether default created_at ordering should be applied
	hasCreatedAt := schema.HasField(constants.ColumnCreatedAt)
	shouldApplyDefaultSort := !a.HasSortApplier() && hasCreatedAt

	return func(ctx fiber.Ctx, db orm.Db, params OptionParams, search TSearch) error {
		var options []Option

		selectQuery := a.BuildQuery(db, (*TModel)(nil), search, ctx)

		// Merge field mapping with defaults and validate
		mergeOptionFieldMapping(&params.OptionFieldMapping, a.fieldMapping)

		if err := validateOptionFields(schema, &params.OptionFieldMapping); err != nil {
			return err
		}

		// Select only required fields
		if params.ValueField == valueField {
			selectQuery.Select(params.ValueField)
		} else {
			selectQuery.SelectAs(params.ValueField, valueField)
		}

		if params.LabelField == labelField {
			selectQuery.Select(params.LabelField)
		} else {
			selectQuery.SelectAs(params.LabelField, labelField)
		}

		if params.DescriptionField != constants.Empty {
			if params.DescriptionField == descriptionField {
				selectQuery.Select(params.DescriptionField)
			} else {
				selectQuery.SelectAs(params.DescriptionField, descriptionField)
			}
		}

		// Apply sorting
		if params.SortField != constants.Empty {
			selectQuery.OrderBy(params.SortField)
		} else if shouldApplyDefaultSort {
			selectQuery.OrderBy(constants.ColumnCreatedAt)
		}

		// Execute query with limit
		if err := selectQuery.Limit(maxOptionsLimit).Scan(ctx.Context(), &options); err != nil {
			return err
		}

		// Ensure empty slice instead of nil for consistent JSON response
		if options == nil {
			options = []Option{}
		}

		return result.Ok(a.Process(options, search, ctx)).Response(ctx)
	}
}
