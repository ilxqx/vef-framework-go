package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/dbhelpers"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/trans"
)

type findTreeAPI[TModel, TSearch any] struct {
	FindAPI[TModel, TSearch, []TModel, FindTreeAPI[TModel, TSearch]]

	idField       string
	parentIdField string
	treeBuilder   func(flatModels []TModel) []TModel
}

func (a *findTreeAPI[TModel, TSearch]) Provide() api.Spec {
	return a.FindAPI.Build(a.findTree)
}

// Build should not be called directly on concrete API types.
// Use Provide() to generate api.Spec with the correct handler instead.
func (a *findTreeAPI[TModel, TSearch]) Build(handler any) api.Spec {
	panic("apis: do not call FindAPI.Build on findTreeAPI; call Provide() instead")
}

func (a *findTreeAPI[TModel, TSearch]) IdField(name string) FindTreeAPI[TModel, TSearch] {
	a.idField = name
	return a
}

func (a *findTreeAPI[TModel, TSearch]) ParentIdField(name string) FindTreeAPI[TModel, TSearch] {
	a.parentIdField = name
	return a
}

func (a *findTreeAPI[TModel, TSearch]) findTree(db orm.Db) func(ctx fiber.Ctx, db orm.Db, transformer trans.Transformer, search TSearch) error {
	// Pre-compute schema information
	schema := db.Schema((*TModel)(nil))

	// Pre-compute whether default created_at ordering should be applied
	hasCreatedAt := schema.HasField(constants.ColumnCreatedAt)
	shouldApplyDefaultSort := !a.HasSortApplier() && hasCreatedAt

	return func(ctx fiber.Ctx, db orm.Db, transformer trans.Transformer, search TSearch) error {
		var flatModels []TModel

		query := db.NewSelect().
			WithRecursive("tmp_tree", func(query orm.SelectQuery) {
				// Base query: fetch root nodes and apply filters
				a.ConfigureQuery(query, (*TModel)(nil), search, ctx)

				if shouldApplyDefaultSort {
					query.OrderByDesc(constants.ColumnCreatedAt)
				}

				query = query.Limit(maxQueryLimit)

				// Recursive part: fetch child nodes by joining with parent results
				query.Union(func(query orm.SelectQuery) {
					query.Model((*TModel)(nil))

					if shouldApplyDefaultSort {
						// Apply default sorting for recursive part
						query.OrderByDesc(constants.ColumnCreatedAt)
					} else if a.HasSortApplier() {
						a.ApplySort(query, search, ctx)
					}

					query.JoinTable("tmp_tree", func(cb orm.ConditionBuilder) {
						cb.EqualsColumn(a.idField, dbhelpers.ColumnWithAlias(a.parentIdField, "tt"))
					}, "tt")
				})
			}).
			Table("tmp_tree")

		if err := query.Scan(ctx, &flatModels); err != nil {
			return err
		}

		// Transform models if there are any results
		if len(flatModels) > 0 {
			for _, model := range flatModels {
				if err := transformer.Struct(ctx, &model); err != nil {
					return err
				}
			}
		} else {
			flatModels = make([]TModel, 0)
		}

		// Build tree structure using the configured TreeBuilder
		models := a.treeBuilder(flatModels)
		return result.Ok(a.Process(models, search, ctx)).Response(ctx)
	}
}
