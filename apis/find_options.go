package apis

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type findOptionsApi[TModel, TSearch any] struct {
	FindApi[TModel, TSearch, []Option, FindOptionsApi[TModel, TSearch]]

	columnMapping *OptionColumnMapping
}

func (a *findOptionsApi[TModel, TSearch]) Provide() api.Spec {
	return a.Build(a.findOptions)
}

func (a *findOptionsApi[TModel, TSearch]) ColumnMapping(mapping *OptionColumnMapping) FindOptionsApi[TModel, TSearch] {
	a.columnMapping = mapping

	return a
}

func (a *findOptionsApi[TModel, TSearch]) findOptions(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, params OptionParams, search TSearch) error, error) {
	if err := a.Init(db); err != nil {
		return nil, err
	}

	table := db.TableOf((*TModel)(nil))

	return func(ctx fiber.Ctx, db orm.Db, params OptionParams, search TSearch) error {
		var options []Option

		selectQuery := a.BuildQuery(db, (*TModel)(nil), search, ctx)

		// Merge column mapping with defaults and validate
		mergeOptionColumnMapping(&params.OptionColumnMapping, a.columnMapping)

		if err := validateOptionColumns(table, &params.OptionColumnMapping); err != nil {
			return err
		}

		// Select only required columns
		if params.ValueColumn == valueColumn {
			selectQuery.Select(params.ValueColumn)
		} else {
			selectQuery.SelectAs(params.ValueColumn, valueColumn)
		}

		if params.LabelColumn == labelColumn {
			selectQuery.Select(params.LabelColumn)
		} else {
			selectQuery.SelectAs(params.LabelColumn, labelColumn)
		}

		if params.DescriptionColumn != constants.Empty {
			if params.DescriptionColumn == descriptionColumn {
				selectQuery.Select(params.DescriptionColumn)
			} else {
				selectQuery.SelectAs(params.DescriptionColumn, descriptionColumn)
			}
		}

		// Apply sorting
		if params.SortColumn != constants.Empty {
			selectQuery.OrderBy(params.SortColumn)
		} else {
			a.ApplyDefaultSort(selectQuery)
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
	}, nil
}
