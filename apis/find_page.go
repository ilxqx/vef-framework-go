package apis

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/go-streams"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/page"
	"github.com/ilxqx/vef-framework-go/result"
)

type findPageApi[TModel, TSearch any] struct {
	FindApi[TModel, TSearch, []TModel, FindPageApi[TModel, TSearch]]

	defaultPageSize int
}

func (a *findPageApi[TModel, TSearch]) Provide() api.Spec {
	return a.Build(a.findPage)
}

// This value is used when the request's page size is zero or invalid.
func (a *findPageApi[TModel, TSearch]) WithDefaultPageSize(size int) FindPageApi[TModel, TSearch] {
	a.defaultPageSize = size

	return a
}

func (a *findPageApi[TModel, TSearch]) findPage(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, pageable page.Pageable, search TSearch) error, error) {
	if err := a.Setup(db, &FindApiConfig{
		QueryParts: &QueryPartsConfig{
			Condition:         []QueryPart{QueryRoot},
			Sort:              []QueryPart{QueryRoot},
			AuditUserRelation: []QueryPart{QueryRoot},
		},
	}); err != nil {
		return nil, err
	}

	return func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, pageable page.Pageable, search TSearch) (err error) {
		pageable.Normalize(a.defaultPageSize)

		var (
			models []TModel
			query  = db.NewSelect().Model(&models).SelectModelColumns().Paginate(pageable)
			total  int64
		)

		if err = a.ConfigureQuery(query, search, ctx, QueryRoot); err != nil {
			return err
		}

		if total, err = query.ScanAndCount(ctx.Context()); err != nil {
			return err
		}

		if total > 0 {
			if err := streams.Range(0, len(models)).ForEachErr(func(i int) error {
				return transformer.Struct(ctx.Context(), &models[i])
			}); err != nil {
				return err
			}

			processedModels := a.Process(models, search, ctx)
			if models, ok := processedModels.([]TModel); ok {
				return result.Ok(page.New(pageable, total, models)).Response(ctx)
			}

			modelsValue := reflect.Indirect(reflect.ValueOf(processedModels))
			if modelsValue.Kind() != reflect.Slice {
				return result.Errf("processor must return a slice, got %T", processedModels)
			}

			items := streams.MapTo(
				streams.Range(0, modelsValue.Len()),
				func(i int) any { return modelsValue.Index(i).Interface() },
			).Collect()

			return result.Ok(page.New(pageable, total, items)).Response(ctx)
		}

		// Ensure empty slice instead of nil for consistent JSON response
		return result.Ok(page.New(pageable, total, []any{})).Response(ctx)
	}, nil
}
