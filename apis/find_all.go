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
	FindApi[TModel, TSearch, []TModel, FindAllApi[TModel, TSearch]]
}

func (a *findAllApi[TModel, TSearch]) Provide() api.Spec {
	return a.Build(a.findAll)
}

func (a *findAllApi[TModel, TSearch]) findAll(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error, error) {
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
			models []TModel
			query  = db.NewSelect().Model(&models)
		)

		if err := a.ConfigureQuery(query, search, ctx, QueryRoot); err != nil {
			return err
		}

		if err := query.SelectModelColumns().
			Limit(maxQueryLimit).
			Scan(ctx.Context()); err != nil {
			return err
		}

		if len(models) > 0 {
			if err := streams.Range(0, len(models)).ForEachErr(func(i int) error {
				return transformer.Struct(ctx.Context(), &models[i])
			}); err != nil {
				return err
			}

			return result.Ok(a.Process(models, search, ctx)).Response(ctx)
		} else {
			// Ensure empty slice instead of nil for consistent JSON response
			models = make([]TModel, 0)
		}

		return result.Ok(models).Response(ctx)
	}, nil
}
