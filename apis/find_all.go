package apis

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type findAllApi[TModel, TSearch any] struct {
	FindApi[TModel, TSearch, []TModel, FindAllApi[TModel, TSearch]]
}

func (a *findAllApi[TModel, TSearch]) Provide() api.Spec {
	return a.Build(a.findAll)
}

func (a *findAllApi[TModel, TSearch]) findAll(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error, error) {
	if err := a.Init(db); err != nil {
		return nil, err
	}

	return func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error {
		var models []TModel

		query := a.BuildQuery(db, &models, search, ctx).SelectModelColumns()

		// Apply default sort if configured
		a.ApplyDefaultSort(query)

		// Execute query with safety limit
		if err := query.Limit(maxQueryLimit).Scan(ctx.Context()); err != nil {
			return err
		}

		if len(models) > 0 {
			// Apply transformation to each model
			for i := range models {
				if err := transformer.Struct(ctx.Context(), &models[i]); err != nil {
					return err
				}
			}

			// Apply post-processing if configured
			return result.Ok(a.Process(models, search, ctx)).Response(ctx)
		} else {
			// Ensure empty slice instead of nil for consistent JSON response
			models = make([]TModel, 0)
		}

		return result.Ok(models).Response(ctx)
	}, nil
}
