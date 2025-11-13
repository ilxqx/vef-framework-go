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

		mergeOptionColumnMapping(&config.DataOptionColumnMapping, a.defaultColumnMapping)

		if err := validateOptionColumns(table, &config.DataOptionColumnMapping); err != nil {
			return err
		}

		metaColumns := parseMetaColumns(config.MetaColumns)
		if err := validateMetaColumns(table, metaColumns); err != nil {
			return err
		}

		if config.ValueColumn == ValueColumn {
			query.Select(config.ValueColumn)
		} else {
			query.SelectAs(config.ValueColumn, ValueColumn)
		}

		if config.LabelColumn == LabelColumn {
			query.Select(config.LabelColumn)
		} else {
			query.SelectAs(config.LabelColumn, LabelColumn)
		}

		if config.DescriptionColumn != constants.Empty {
			if config.DescriptionColumn == DescriptionColumn {
				query.Select(config.DescriptionColumn)
			} else {
				query.SelectAs(config.DescriptionColumn, DescriptionColumn)
			}
		}

		query.ApplyIf(len(metaColumns) > 0, func(sq orm.SelectQuery) {
			sq.SelectExpr(
				func(eb orm.ExprBuilder) any {
					return buildMetaJsonExpr(eb, metaColumns)
				},
				"meta",
			)
		})

		if err := a.ConfigureQuery(query, search, ctx, QueryRoot); err != nil {
			return err
		}

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
