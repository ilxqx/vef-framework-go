package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

const (
	// maxQueryLimit is the maximum number of records that can be returned in a single query
	// to prevent excessive memory usage and performance issues
	maxQueryLimit = 10000
)

// FindAllAPI provides unlimited query functionality with filtering and post-processing.
// It returns all matching records up to the maximum limit without pagination.
//
// Type parameters:
//   - TModel: The database model type
//   - TSearch: The search criteria type
type FindAllAPI[TModel, TSearch any] struct {
	*findAPI[TModel, TSearch, PostFindProcessor[[]TModel, []any], FindAllAPI[TModel, TSearch]]
}

// FindAll creates a handler that executes the query and returns all matching records.
// It applies the configured search criteria, filters, and transformations.
//
// Parameters:
//   - db: The database connection for schema introspection
//
// Returns a handler function that processes find-all requests.
func (a *FindAllAPI[TModel, TSearch]) FindAll(db orm.Db) func(ctx fiber.Ctx, db orm.Db, search TSearch) error {
	// Pre-compute schema information
	schema := db.Schema((*TModel)(nil))

	// Pre-compute whether default ordering should be applied
	hasCreatedAt := schema.HasField(orm.ColumnCreatedAt)
	shouldApplyDefaultSort := a.sortApplier == nil && hasCreatedAt

	return func(ctx fiber.Ctx, db orm.Db, search TSearch) error {
		var models []TModel
		query := a.buildQuery(ctx, db, &models, search)

		// Add default ordering by created_at if pre-computed condition is true
		if shouldApplyDefaultSort {
			query.OrderByDesc(orm.ColumnCreatedAt)
		}

		// Execute query with safety limit
		if err := query.Limit(maxQueryLimit).Scan(ctx); err != nil {
			return err
		}

		// Transform models and apply post-processing if results exist
		if len(models) > 0 {
			// Apply transformation to each model
			for _, model := range models {
				if err := apisParams.Transformer.Struct(ctx, &model); err != nil {
					return err
				}
			}

			// Apply post-processing if configured
			if a.processor != nil {
				processed := a.processor(models, ctx)
				return result.Ok(processed).Response(ctx)
			}
		} else {
			// Ensure empty slice instead of nil for consistent JSON response
			models = make([]TModel, 0)
		}

		return result.Ok(models).Response(ctx)
	}
}

// NewFindAllAPI creates a new FindAllAPI instance.
// Use method chaining to configure filters, relations, and post-processing.
//
// Example:
//
//	api := NewFindAllAPI[User, UserSearch]().
//	  WithFilterApplier(myFilter).
//	  WithRelations("profile", "roles").
//	  WithPostFind(myProcessor)
func NewFindAllAPI[TModel, TSearch any]() *FindAllAPI[TModel, TSearch] {
	api := new(FindAllAPI[TModel, TSearch])
	api.findAPI = newFindAPI[TModel, TSearch, PostFindProcessor[[]TModel, []any]](api)

	return api
}
