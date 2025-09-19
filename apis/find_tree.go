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

// WithIdField sets the field name used as the node ID in tree structures.
// This field is used to identify individual nodes and establish parent-child relationships.
// Returns the API instance for method chaining.
func (a *FindTreeAPI[TModel, TSearch]) WithIdField(name string) *FindTreeAPI[TModel, TSearch] {
	a.idField = name
	return a
}

// WithParentIdField sets the field name used to reference parent nodes in tree structures.
// This field establishes the hierarchical relationship between parent and child nodes.
// Returns the API instance for method chaining.
func (a *FindTreeAPI[TModel, TSearch]) WithParentIdField(name string) *FindTreeAPI[TModel, TSearch] {
	a.parentIdField = name
	return a
}

// FindTree creates a handler that executes the query and returns hierarchical tree structure.
// Uses recursive CTE to efficiently fetch all nodes, then delegates tree building to the configured TreeBuilder function.
//
// Parameters:
//   - db: The database connection for schema introspection
//
// Returns a handler function that processes find-tree requests.
func (a *FindTreeAPI[TModel, TSearch]) FindTree(db orm.Db) func(ctx fiber.Ctx, db orm.Db, search TSearch) error {
	// Pre-compute schema information
	schema := db.Schema((*TModel)(nil))

	// Pre-compute whether default created_at ordering should be applied
	hasCreatedAt := schema.HasField(orm.ColumnCreatedAt)

	return func(ctx fiber.Ctx, db orm.Db, search TSearch) error {
		var flatModels []TModel

		query := db.NewQuery().
			WithRecursive("tmp_tree", func(query orm.Query) {
				// Base query: fetch root nodes and apply filters
				query = a.configQuery(ctx, query, (*TModel)(nil), search)

				if a.sortApplier == nil && hasCreatedAt {
					query.OrderByDesc(orm.ColumnCreatedAt)
				}

				query = query.Limit(maxQueryLimit)

				// Recursive part: fetch child nodes by joining with parent results
				query.Union(func(query orm.Query) {
					query.Model((*TModel)(nil))

					if a.sortApplier == nil && hasCreatedAt {
						// Apply default sorting for recursive part
						query.OrderByDesc(orm.ColumnCreatedAt)
					} else if a.sortApplier != nil {
						applySort(ctx, query, a.sortApplier)
					}

					query.JoinTableAs("tmp_tree", "tt", func(cb orm.ConditionBuilder) {
						cb.EqualsExpr(a.idField, "?.?", orm.Name("tt"), orm.Name(a.parentIdField))
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
