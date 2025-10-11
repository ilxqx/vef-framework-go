package apis

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type findOneAPI[TModel, TSearch any] struct {
	FindAPI[TModel, TSearch, TModel, FindOneAPI[TModel, TSearch]]
}

func (a *findOneAPI[TModel, TSearch]) Provide() api.Spec {
	return a.FindAPI.Build(a.findOne)
}

// Build should not be called directly on concrete API types.
// Use Provide() to generate api.Spec with the correct handler instead.
func (a *findOneAPI[TModel, TSearch]) Build(handler any) api.Spec {
	panic("apis: do not call FindAPI.Build on findOneAPI; call Provide() instead")
}

func (a *findOneAPI[TModel, TSearch]) findOne(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error, error) {
	if err := a.Init(db); err != nil {
		return nil, err
	}

	return func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error {
		var (
			model TModel
			query = a.BuildQuery(db, &model, search, ctx).SelectModelColumns()
		)

		// Limit to 1 record for efficiency
		if err := query.Limit(1).Scan(ctx.Context()); err != nil {
			return err
		}

		// Apply transformation to the model
		if err := transformer.Struct(ctx.Context(), &model); err != nil {
			return err
		}

		return result.Ok(a.Process(model, search, ctx)).Response(ctx)
	}, nil
}
