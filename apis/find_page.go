package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/mo"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// FindPageAPI provides paginated query functionality with filtering and post-processing.
// It returns results in pages with metadata about total count and pagination state.
//
// Type parameters:
//   - TModel: The database model type
//   - TSearch: The search criteria type
type FindPageAPI[TModel, TSearch any] struct {
	*findAPI[TModel, TSearch, PostFindProcessor[[]TModel, []any], FindPageAPI[TModel, TSearch]]
}

// FindPage executes the paginated query and returns the results with pagination metadata.
// It applies the configured search criteria, filters, and transformations.
//
// Parameters:
//   - ctx: The Fiber context
//   - db: The database connection
//   - pageable: The pagination parameters (page, size, sort)
//   - search: The search criteria
//
// Returns an error if the query fails, otherwise returns a Page object with data and metadata via HTTP response.
func (a *FindPageAPI[TModel, TSearch]) FindPage(ctx fiber.Ctx, db orm.Db, pageable mo.Pageable, search TSearch) error {
	var (
		models []TModel
		schema = db.Schema((*TModel)(nil))
		query  = a.buildQuery(ctx, db, &models, search)
	)

	// Normalize pagination parameters
	pageable.Normalize()

	// Add default ordering by created_at if no custom query applier is set and no sort specified
	if a.queryApplier == nil && pageable.Sort == constants.Empty {
		if schema.HasField(orm.ColumnCreatedAt) {
			query.OrderByDesc(orm.ColumnCreatedAt)
		}
	}

	// Execute paginated query and get total count
	total, err := query.Paginate(pageable).ScanAndCount(ctx)
	if err != nil {
		return err
	}

	// Transform models and apply post-processing if results exist
	if total > 0 {
		// Apply transformation to each model
		for _, model := range models {
			if err := apisParams.Transformer.Struct(ctx, &model); err != nil {
				return err
			}
		}

		// Apply post-processing if configured
		if a.processor != nil {
			processed := a.processor(models, ctx)
			return result.Ok(mo.NewPage(pageable, total, processed)).Response(ctx)
		}
	} else {
		// Ensure empty slice instead of nil for consistent JSON response
		models = make([]TModel, 0)
	}

	return result.Ok(mo.NewPage(pageable, total, models)).Response(ctx)
}

// NewFindPageAPI creates a new FindPageAPI instance.
// Use method chaining to configure filters, relations, and post-processing.
//
// Example:
//
//	api := NewFindPageAPI[User, UserSearch]().
//	  WithFilterApplier(myFilter).
//	  WithRelations("profile", "roles").
//	  WithPostFind(myProcessor)
func NewFindPageAPI[TModel, TSearch any]() *FindPageAPI[TModel, TSearch] {
	api := new(FindPageAPI[TModel, TSearch])
	api.findAPI = newFindAPI[TModel, TSearch, PostFindProcessor[[]TModel, []any]](api)

	return api
}
