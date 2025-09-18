package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// FindTreeAPI provides hierarchical tree query functionality with filtering and post-processing.
// The tree structure is built using a custom TreeBuilder function provided in the configuration.
type FindTreeAPI[TModel, TSearch any] struct {
	*findAPI[TModel, TSearch, PostFindProcessor[[]TModel, []any], FindTreeAPI[TModel, TSearch]]
	idField       string
	parentIdField string
	treeBuilder   func(flatModels []TModel) []TModel
}

func (a *FindTreeAPI[TModel, TSearch]) WithIdField(name string) *FindTreeAPI[TModel, TSearch] {
	a.idField = name
	return a
}

func (a *FindTreeAPI[TModel, TSearch]) WithParentIdField(name string) *FindTreeAPI[TModel, TSearch] {
	a.parentIdField = name
	return a
}

// FindTree executes the query and returns hierarchical tree structure.
// Uses recursive CTE to efficiently fetch all nodes, then delegates tree building to the configured TreeBuilder function.
func (a *FindTreeAPI[TModel, TSearch]) FindTree(ctx fiber.Ctx, db orm.Db, search TSearch) error {
	var (
		flatModels []TModel
		schema     = db.Schema((*TModel)(nil))
	)

	query := db.NewQuery().
		WithRecursive("tmp_tree", func(query orm.Query) {
			// Base query: fetch root nodes and apply filters
			query = a.configQuery(ctx, query, (*TModel)(nil), search)

			if a.queryApplier == nil {
				if field, _ := schema.Field(orm.ColumnCreatedAt); field != nil {
					query.OrderByDesc(orm.ColumnCreatedAt)
				}
			}

			query = query.Limit(maxQueryLimit)

			// Recursive part: fetch child nodes by joining with parent results
			query.Union(func(query orm.Query) {
				query.Model((*TModel)(nil))

				if a.queryApplier == nil {
					// Apply default sorting for recursive part
					if field, _ := schema.Field(orm.ColumnCreatedAt); field != nil {
						query.OrderByDesc(orm.ColumnCreatedAt)
					}
				} else {
					query.Apply(a.queryApplier(search, ctx))
				}

				query.JoinTableAs("tmp_tree", "t", func(cb orm.ConditionBuilder) {
					cb.EqualsExpr(a.idField, "?.?", orm.Name("t"), orm.Name(a.parentIdField))
				})
			})
		}).
		Table("tmp_tree")

	// Execute recursive CTE query
	if err := query.Scan(ctx, &flatModels); err != nil {
		return err
	}

	// Transform models if there are any results
	if len(flatModels) > 0 {
		for _, model := range flatModels {
			if err := apisParams.Transformer.Struct(ctx, &model); err != nil {
				return err
			}
		}
	} else {
		flatModels = make([]TModel, 0)
	}

	// Build tree structure using the configured TreeBuilder function
	models := a.treeBuilder(flatModels)
	if a.processor != nil {
		processed := a.processor(models, ctx)
		return result.Ok(processed).Response(ctx)
	}

	return result.Ok(models).Response(ctx)
}

// NewFindTreeAPI creates a new FindTreeAPI with the specified options.
// Requires TreeConfig with a custom TreeBuilder function to convert flat models to tree structure.
func NewFindTreeAPI[TModel, TSearch any](treeBuilder func(flatModels []TModel) []TModel) *FindTreeAPI[TModel, TSearch] {
	api := &FindTreeAPI[TModel, TSearch]{
		idField:       idField,
		parentIdField: parentIdField,
		treeBuilder:   treeBuilder,
	}
	api.findAPI = newFindAPI[TModel, TSearch, PostFindProcessor[[]TModel, []any]](api)

	return api
}
