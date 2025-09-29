package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/page"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/spf13/cast"
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

func (a *findPageAPI[TModel, TSearch]) findPage(db orm.Db) func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, pageable page.Pageable, search TSearch) error {
	// Pre-compute schema information
	schema := db.Schema((*TModel)(nil))

	// Pre-compute whether default created_at ordering should be applied
	hasCreatedAt := schema.HasField(constants.ColumnCreatedAt)
	shouldApplyDefaultSort := !a.HasSortApplier() && hasCreatedAt

	return func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, pageable page.Pageable, search TSearch) error {
		var models []TModel
		query := a.BuildQuery(db, &models, search, ctx)

		// Normalize pagination parameters
		pageable.Normalize()

		if shouldApplyDefaultSort && len(pageable.Sort) == 0 {
			// Add default ordering by created_at
			query.OrderByDesc(constants.ColumnCreatedAt)
		}

		// Execute paginated query and get total count
		total, err := query.Paginate(pageable).ScanAndCount(ctx)
		if err != nil {
			return err
		}

		if total > 0 {
			// Apply transformation to each model
			for _, model := range models {
				if err := transformer.Struct(ctx, &model); err != nil {
					return err
				}
			}

			// Apply post-processing if configured and convert to interface slice for JSON serialization
			processedData := a.Process(models, search, ctx)
			return result.Ok(page.New(pageable, total, cast.ToSlice(processedData))).Response(ctx)
		} else {
			// Ensure empty slice instead of nil for consistent JSON response
			models = make([]TModel, 0)
		}

		return result.Ok(page.New(pageable, total, models)).Response(ctx)
	}
}
