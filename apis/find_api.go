package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/search"
)

// findAPI is the base struct for all find operations, providing common functionality
// for search, filter, query application, relations, and post-processing.
//
// Type parameters:
//   - TModel: The database model type
//   - TSearch: The search criteria type
//   - TPostFindProcessor: The post-processing function type
//   - TFindAPI: The concrete API type that embeds this struct
type findAPI[TModel, TSearch, TPostFindProcessor, TFindAPI any] struct {
	api           *TFindAPI              // Reference to the concrete API instance
	searchApplier SearchApplier[TSearch] // Function to apply search conditions
	filterApplier FilterApplier[TSearch] // Function to apply filter conditions
	queryApplier  QueryApplier[TSearch]  // Function to apply additional query modifications
	relations     []orm.ModelRelation    // Model relations to include in queries
	processor     TPostFindProcessor     // Post-processing function for results
}

// WithFilterApplier sets a custom filter applier function for additional filtering logic.
// Returns the concrete API instance for method chaining.
func (a *findAPI[TModel, TSearch, TPostFindProcessor, TFindAPI]) WithFilterApplier(applier FilterApplier[TSearch]) *TFindAPI {
	a.filterApplier = applier
	return a.api
}

// WithQueryApplier sets a custom query applier function for additional query modifications.
// Returns the concrete API instance for method chaining.
func (a *findAPI[TModel, TSearch, TPostFindProcessor, TFindAPI]) WithQueryApplier(applier QueryApplier[TSearch]) *TFindAPI {
	a.queryApplier = applier
	return a.api
}

// WithRelations adds model relations to be included in the query.
// Returns the concrete API instance for method chaining.
func (a *findAPI[TModel, TSearch, TPostFindProcessor, TFindAPI]) WithRelations(relations ...orm.ModelRelation) *TFindAPI {
	a.relations = append(a.relations, relations...)
	return a.api
}

// WithPostFind sets a post-processing function to transform query results.
// Returns the concrete API instance for method chaining.
func (a *findAPI[TModel, TSearch, TPostFindProcessor, TFindAPI]) WithPostFind(processor TPostFindProcessor) *TFindAPI {
	a.processor = processor
	return a.api
}

// buildQuery creates a new query with the configured model and search criteria.
func (a *findAPI[TModel, TSearch, TPostFindProcessor, TFindAPI]) buildQuery(ctx fiber.Ctx, db orm.Db, model any, search TSearch) orm.Query {
	return a.configQuery(ctx, db.NewQuery(), model, search)
}

// configQuery applies all configured search, filter, query, and relation settings to the given query.
func (a *findAPI[TModel, TSearch, TPostFindProcessor, TFindAPI]) configQuery(ctx fiber.Ctx, query orm.Query, model any, search TSearch) orm.Query {
	query = query.
		Model(model).
		Where(func(cb orm.ConditionBuilder) {
			// Apply basic search conditions
			cb.Apply(a.searchApplier(search))
			// Apply additional filter conditions if configured
			if a.filterApplier != nil {
				cb.Apply(a.filterApplier(search, ctx))
			}
		})

	// Include model relations if specified
	if len(a.relations) > 0 {
		query.ModelRelation(a.relations...)
	}

	// Apply additional query modifications if configured
	if a.queryApplier != nil {
		query.Apply(a.queryApplier(search, ctx))
	}

	return query
}

// newFindAPI creates a new findAPI instance with default search applier.
func newFindAPI[TModel, TSearch, TPostFindProcessor, TFindAPI any](api *TFindAPI) *findAPI[TModel, TSearch, TPostFindProcessor, TFindAPI] {
	return &findAPI[TModel, TSearch, TPostFindProcessor, TFindAPI]{
		api:           api,
		searchApplier: search.Applier[TSearch](),
	}
}
