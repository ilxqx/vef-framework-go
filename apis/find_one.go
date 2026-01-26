package apis

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type findOneApi[TModel, TSearch any] struct {
	Find[TModel, TSearch, TModel, FindOne[TModel, TSearch]]
}

func (a *findOneApi[TModel, TSearch]) Provide() []api.OperationSpec {
	return []api.OperationSpec{a.Build(a.findOne)}
}

func (a *findOneApi[TModel, TSearch]) findOne(db orm.DB) (func(ctx fiber.Ctx, db orm.DB, transformer mold.Transformer, search TSearch, meta api.Meta) error, error) {
	if err := a.Setup(db, &FindApiConfig{
		QueryParts: &QueryPartsConfig{
			Condition:         []QueryPart{QueryRoot},
			Sort:              []QueryPart{QueryRoot},
			AuditUserRelation: []QueryPart{QueryRoot},
		},
	}); err != nil {
		return nil, err
	}

	return func(ctx fiber.Ctx, db orm.DB, transformer mold.Transformer, search TSearch, meta api.Meta) error {
		var (
			model TModel
			query = db.NewSelect().Model(&model)
		)

		if err := a.ConfigureQuery(query, search, meta, ctx, QueryRoot); err != nil {
			return err
		}

		// Limit to 1 record for efficiency
		if err := query.SelectModelColumns().
			Limit(1).
			Scan(ctx.Context()); err != nil {
			return err
		}

		if err := transformer.Struct(ctx.Context(), &model); err != nil {
			return err
		}

		return result.Ok(a.Process(model, search, ctx)).Response(ctx)
	}, nil
}
