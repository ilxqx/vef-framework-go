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

type findPageAPI[TModel, TSearch any] struct {
	FindAPI[TModel, TSearch, []TModel, FindPageAPI[TModel, TSearch]]
}

func (a *findPageAPI[TModel, TSearch]) Provide() api.Spec {
	return a.FindAPI.Build(a.findPage)
}

// Build should not be called directly on concrete API types.
// Use Provide() to generate api.Spec with the correct handler instead.
func (a *findPageAPI[TModel, TSearch]) Build(handler any) api.Spec {
	panic("apis: do not call FindAPI.Build on findPageAPI; call Provide() instead")
}

func (a *findPageAPI[TModel, TSearch]) findPage(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, pageable page.Pageable, search TSearch) error, error) {
	if err := a.Init(db); err != nil {
		return nil, err
	}

	return func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, pageable page.Pageable, search TSearch) error {
		var models []TModel

		query := a.BuildQuery(db, (*TModel)(nil), search, ctx).SelectModelColumns()

		// Normalize pagination parameters
		pageable.Normalize()

		// Apply default sort only if user hasn't specified custom sorting
		if len(pageable.Sort) == 0 {
			a.ApplyDefaultSort(query)
		}

		// Execute paginated query and get total count
		total, err := query.Paginate(pageable).ScanAndCount(ctx.Context(), &models)
		if err != nil {
			return err
		}

		if total > 0 {
			// Apply transformation to each model
			for _, model := range models {
				if err := transformer.Struct(ctx.Context(), &model); err != nil {
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
