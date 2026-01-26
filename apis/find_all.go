package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/go-streams"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type findAllApi[TModel, TSearch any] struct {
	Find[TModel, TSearch, []TModel, FindAll[TModel, TSearch]]
}

func (a *findAllApi[TModel, TSearch]) Provide() []api.OperationSpec {
	return []api.OperationSpec{a.Build(a.findAll)}
}

func (a *findAllApi[TModel, TSearch]) findAll(db orm.DB) (func(ctx fiber.Ctx, db orm.DB, transformer mold.Transformer, search TSearch, meta api.Meta) error, error) {
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
			models []TModel
			query  = db.NewSelect().Model(&models)
		)

		if err := a.ConfigureQuery(query, search, meta, ctx, QueryRoot); err != nil {
			return err
		}

		if err := query.SelectModelColumns().
			Limit(maxQueryLimit).
			Scan(ctx.Context()); err != nil {
			return err
		}

		if err := streams.Range(0, len(models)).ForEachErr(func(i int) error {
			return transformer.Struct(ctx.Context(), &models[i])
		}); err != nil {
			return err
		}

		if models == nil {
			models = []TModel{}
		}

		return result.Ok(a.Process(models, search, ctx)).Response(ctx)
	}, nil
}
