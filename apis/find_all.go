package apis

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type findAllAPI[TModel, TSearch any] struct {
	FindAPI[TModel, TSearch, []TModel, FindAllAPI[TModel, TSearch]]
}

func (a *findAllAPI[TModel, TSearch]) Provide() api.Spec {
	return a.FindAPI.Build(a.findAll)
}

// Build should not be called directly on concrete API types.
// Use Provide() to generate api.Spec with the correct handler instead.
func (a *findAllAPI[TModel, TSearch]) Build(handler any) api.Spec {
	panic("apis: do not call FindAPI.Build on findAllAPI; call Provide() instead")
}

func (a *findAllAPI[TModel, TSearch]) findAll(db orm.Db) func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error {
	// Pre-compute schema information
	schema := db.TableOf((*TModel)(nil))

	// Pre-compute whether default ordering should be applied
	hasCreatedAt := schema.HasField(constants.ColumnCreatedAt)
	shouldApplyDefaultSort := !a.HasSortApplier() && hasCreatedAt

	return func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error {
		var models []TModel

		query := a.BuildQuery(db, &models, search, ctx)

		if shouldApplyDefaultSort {
			// Add default ordering by created_at
			query.OrderByDesc(constants.ColumnCreatedAt)
		}

		// Execute query with safety limit
		if err := query.Limit(maxQueryLimit).Scan(ctx.Context()); err != nil {
			return err
		}

		if len(models) > 0 {
			// Apply transformation to each model
			for _, model := range models {
				if err := transformer.Struct(ctx.Context(), &model); err != nil {
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
	}
}
