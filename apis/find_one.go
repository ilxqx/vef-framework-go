package apis

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type findOneApi[TModel, TSearch any] struct {
	FindApi[TModel, TSearch, TModel, FindOneApi[TModel, TSearch]]
}

func (a *findOneApi[TModel, TSearch]) Provide() api.Spec {
	return a.Build(a.findOne)
}

func (a *findOneApi[TModel, TSearch]) findOne(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error, error) {
	if err := a.Setup(db, &FindApiConfig{
		QueryParts: &QueryPartsConfig{
			Condition:         []QueryPart{QueryRoot},
			Sort:              []QueryPart{QueryRoot},
			AuditUserRelation: []QueryPart{QueryRoot},
		},
	}); err != nil {
		return nil, err
	}

	return func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error {
		var (
			model TModel
			query = db.NewSelect().Model(&model)
		)

		if err := a.ConfigureQuery(query, search, ctx, QueryRoot); err != nil {
			return err
		}

		// Limit to 1 record for efficiency
		if err := query.SelectModelColumns().
			Limit(1).
			Scan(ctx.Context()); err != nil {
			return err
		}

		// Apply transformation to the model
		if err := transformer.Struct(ctx.Context(), &model); err != nil {
			return err
		}

		return result.Ok(a.Process(model, search, ctx)).Response(ctx)
	}, nil
}
