package apis

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type findOptionsApi[TModel, TSearch any] struct {
	FindApi[TModel, TSearch, []DataOption, FindOptionsApi[TModel, TSearch]]

	defaultColumnMapping *DataOptionColumnMapping
}

func (a *findOptionsApi[TModel, TSearch]) Provide() api.Spec {
	return a.Build(a.findOptions)
}

// WithDefaultColumnMapping sets the default column mapping for options queries.
// This mapping provides fallback values for column mapping when not explicitly specified in queries.
func (a *findOptionsApi[TModel, TSearch]) WithDefaultColumnMapping(mapping *DataOptionColumnMapping) FindOptionsApi[TModel, TSearch] {
	a.defaultColumnMapping = mapping

	return a
}

func (a *findOptionsApi[TModel, TSearch]) findOptions(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, config DataOptionConfig, search TSearch) error, error) {
	if err := a.Setup(db, &FindApiConfig{
		QueryParts: &QueryPartsConfig{
			Condition:         []QueryPart{QueryRoot},
			Sort:              []QueryPart{QueryRoot},
			AuditUserRelation: []QueryPart{QueryRoot},
		},
	}); err != nil {
		return nil, err
	}

	table := db.TableOf((*TModel)(nil))

	return func(ctx fiber.Ctx, db orm.Db, config DataOptionConfig, search TSearch) error {
		var (
			options []DataOption
			query   = db.NewSelect().Model((*TModel)(nil))
		)

		// Merge column mapping with defaults and validate
		mergeOptionColumnMapping(&config.DataOptionColumnMapping, a.defaultColumnMapping)

		if err := validateOptionColumns(table, &config.DataOptionColumnMapping); err != nil {
			return err
		}

		// Select only required columns
		if config.ValueColumn == valueColumn {
			query.Select(config.ValueColumn)
		} else {
			query.SelectAs(config.ValueColumn, valueColumn)
		}

		if config.LabelColumn == labelColumn {
			query.Select(config.LabelColumn)
		} else {
			query.SelectAs(config.LabelColumn, labelColumn)
		}

		if config.DescriptionColumn != constants.Empty {
			if config.DescriptionColumn == descriptionColumn {
				query.Select(config.DescriptionColumn)
			} else {
				query.SelectAs(config.DescriptionColumn, descriptionColumn)
			}
		}

		if err := a.ConfigureQuery(query, search, ctx, QueryRoot); err != nil {
			return err
		}

		// Execute query with limit
		if err := query.Limit(maxOptionsLimit).
			Scan(ctx.Context(), &options); err != nil {
			return err
		}

		// Ensure empty slice instead of nil for consistent JSON response
		if options == nil {
			options = []DataOption{}
		}

		return result.Ok(a.Process(options, search, ctx)).Response(ctx)
	}, nil
}
