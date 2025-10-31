package apis

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

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

		// Execute paginated query and get total count
		if total, err = query.ScanAndCount(ctx.Context()); err != nil {
			return err
		}

		if total > 0 {
			// Apply transformation to each model
			for i := range models {
				if err := transformer.Struct(ctx.Context(), &models[i]); err != nil {
					return err
				}
			}

			// Apply post-processing if configured
			processedModels := a.Process(models, search, ctx)
			if models, ok := processedModels.([]TModel); ok {
				return result.Ok(page.New(pageable, total, models)).Response(ctx)
			}

			// Check if processor returned a slice
			modelsValue := reflect.Indirect(reflect.ValueOf(processedModels))
			if modelsValue.Kind() != reflect.Slice {
				return result.Errf("processor must return a slice, got %T", processedModels)
			}

			// Convert slice to []any for page.New
			items := make([]any, modelsValue.Len())
			for i := range modelsValue.Len() {
				items[i] = modelsValue.Index(i).Interface()
			}

			return result.Ok(page.New(pageable, total, items)).Response(ctx)
		}

		// Ensure empty slice instead of nil for consistent JSON response
		return result.Ok(page.New(pageable, total, []any{})).Response(ctx)
	}, nil
}
