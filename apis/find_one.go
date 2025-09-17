package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// FindOneAPI provides single record query functionality with filtering and post-processing.
// It returns the first matching record based on the search criteria.
//
// Type parameters:
//   - TModel: The database model type
//   - TSearch: The search criteria type
type FindOneAPI[TModel, TSearch any] struct {
	*findAPI[TModel, TSearch, PostFindProcessor[TModel, any], FindOneAPI[TModel, TSearch]]
}

// FindOne executes the query and returns a single record.
// It applies the configured search criteria, filters, and transformations.
//
// Parameters:
//   - ctx: The Fiber context
//   - db: The database connection
//   - search: The search criteria
//
// Returns an error if the query fails, otherwise returns the first matching record via HTTP response.
// If no record is found, returns the zero value of the model type.
func (a *FindOneAPI[TModel, TSearch]) FindOne(ctx fiber.Ctx, db orm.Db, search TSearch) error {
	var (
		model TModel
		query = a.buildQuery(ctx, db, &model, search)
	)

	// Limit to 1 record for efficiency
	if err := query.Limit(1).Scan(ctx); err != nil {
		return err
	}

	// Apply transformation to the model
	if err := apisParams.Transformer.Struct(ctx, &model); err != nil {
		return err
	}

	// Apply post-processing if configured
	if a.processor != nil {
		return result.Ok(a.processor(model, ctx)).Response(ctx)
	}

	return result.Ok(model).Response(ctx)
}

// NewFindOneAPI creates a new FindOneAPI instance.
// Use method chaining to configure filters, relations, and post-processing.
//
// Example:
//
//	api := NewFindOneAPI[User, UserSearch]().
//	  WithFilterApplier(myFilter).
//	  WithRelations("profile").
//	  WithPostFind(myProcessor)
func NewFindOneAPI[TModel, TSearch any]() *FindOneAPI[TModel, TSearch] {
	api := new(FindOneAPI[TModel, TSearch])
	api.findAPI = newFindAPI[TModel, TSearch, PostFindProcessor[TModel, any]](api)

	return api
}
