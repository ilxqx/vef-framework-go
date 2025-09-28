package apis

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/orm"
)

// APIBuilder defines the interface for building API endpoint.
// It provides a fluent API for configuring all aspects of an API endpoint.
type APIBuilder[T any] interface {
	// Action sets the action name for the API endpoint.
	Action(action string) T
	// EnableAudit enables audit logging for this endpoint.
	EnableAudit() T
	// Timeout sets the request timeout duration.
	Timeout(timeout time.Duration) T
	// Public sets this endpoint is publicly accessible.
	Public() T
	// PermissionToken sets the permission token required for access.
	PermissionToken(token string) T
	// RateLimit sets the rate limit configuration for this endpoint.
	RateLimit(max int, expiration time.Duration) T
	// Build builds the API endpoint specification.
	Build(handler any) api.Spec
}

type CreateAPI[TModel, TParams any] interface {
	api.Provider
	APIBuilder[CreateAPI[TModel, TParams]]

	// PreCreate sets the pre-create processor for the CreateAPI.
	// This processor is called before the model is saved to the database.
	PreCreate(processor PreCreateProcessor[TModel, TParams]) CreateAPI[TModel, TParams]
	// PostCreate sets the post-create processor for the CreateAPI.
	// This processor is called after the model is successfully saved within the same transaction.
	PostCreate(processor PostCreateProcessor[TModel, TParams]) CreateAPI[TModel, TParams]
}

type UpdateAPI[TModel, TParams any] interface {
	api.Provider
	APIBuilder[UpdateAPI[TModel, TParams]]

	// PreUpdate sets the pre-update processor for the UpdateAPI.
	// This processor is called before the model is updated in the database.
	PreUpdate(processor PreUpdateProcessor[TModel, TParams]) UpdateAPI[TModel, TParams]
	// PostUpdate sets the post-update processor for the UpdateAPI.
	// This processor is called after the model is successfully updated within the same transaction.
	PostUpdate(processor PostUpdateProcessor[TModel, TParams]) UpdateAPI[TModel, TParams]
}

type DeleteAPI[TModel any] interface {
	api.Provider
	APIBuilder[DeleteAPI[TModel]]

	// PreDelete sets the pre-delete processor for the DeleteAPI.
	// This processor is called before the model is deleted from the database.
	PreDelete(processor PreDeleteProcessor[TModel]) DeleteAPI[TModel]
	// PostDelete sets the post-delete processor for the DeleteAPI.
	// This processor is called after the model is successfully deleted within the same transaction.
	PostDelete(processor PostDeleteProcessor[TModel]) DeleteAPI[TModel]
}

type FindAPI[TModel, TSearch, TProcessorIn, TAPI any] interface {
	APIBuilder[TAPI]

	// QueryApplier sets a custom query applier function for additional query modifications.
	QueryApplier(applier QueryApplier[TSearch]) TAPI
	// FilterApplier sets a custom filter applier function for additional filtering logic.
	FilterApplier(applier FilterApplier[TSearch]) TAPI
	// SortApplier sets a custom sort applier function for additional order modifications.
	SortApplier(applier SortApplier[TSearch]) TAPI
	// Relations adds model relations to be included in the query.
	Relations(relations ...orm.ModelRelation) TAPI
	// Processor sets a post-processing function to transform query results.
	Processor(processor Processor[TProcessorIn, TSearch]) TAPI

	BuildQuery(db orm.Db, model any, search TSearch, ctx fiber.Ctx) orm.SelectQuery
	ConfigureQuery(query orm.SelectQuery, model any, search TSearch, ctx fiber.Ctx)
	Process(input TProcessorIn, search TSearch, ctx fiber.Ctx) any
	ApplySearch(query orm.SelectQuery, search TSearch, ctx fiber.Ctx)
	ApplyFilter(query orm.SelectQuery, search TSearch, ctx fiber.Ctx)
	ApplyConditions(query orm.SelectQuery, search TSearch, ctx fiber.Ctx)
	ApplyQuery(query orm.SelectQuery, search TSearch, ctx fiber.Ctx)
	ApplySort(query orm.SelectQuery, search TSearch, ctx fiber.Ctx)
	ApplyRelations(query orm.SelectQuery, search TSearch, ctx fiber.Ctx)
	HasSortApplier() bool
}

type FindOneAPI[TModel, TSearch any] interface {
	api.Provider
	FindAPI[TModel, TSearch, TModel, FindOneAPI[TModel, TSearch]]
}

type FindAllAPI[TModel, TSearch any] interface {
	api.Provider
	FindAPI[TModel, TSearch, []TModel, FindAllAPI[TModel, TSearch]]
}

type FindPageAPI[TModel, TSearch any] interface {
	api.Provider
	FindAPI[TModel, TSearch, []TModel, FindPageAPI[TModel, TSearch]]
}

type FindTreeAPI[TModel, TSearch any] interface {
	api.Provider
	FindAPI[TModel, TSearch, []TModel, FindTreeAPI[TModel, TSearch]]

	// IdField sets the field name used as the node Id in tree structures.
	// This field is used to identify individual nodes and establish parent-child relationships.
	IdField(name string) FindTreeAPI[TModel, TSearch]
	// ParentIdField sets the field name used to reference parent nodes in tree structures.
	// This field establishes the hierarchical relationship between parent and child nodes.
	ParentIdField(name string) FindTreeAPI[TModel, TSearch]
}

type FindOptionsAPI[TModel, TSearch any] interface {
	api.Provider
	FindAPI[TModel, TSearch, []Option, FindOptionsAPI[TModel, TSearch]]

	// DefaultConfig sets the default configuration for options queries.
	// This configuration provides fallback values for field mapping when not explicitly specified in queries.
	DefaultConfig(config *OptionsConfig) FindOptionsAPI[TModel, TSearch]
}

type FindTreeOptionsAPI[TModel, TSearch any] interface {
	api.Provider
	FindAPI[TModel, TSearch, []TreeOption, FindTreeOptionsAPI[TModel, TSearch]]

	// DefaultConfig sets the default configuration for tree options queries.
	// This configuration provides fallback values for field mapping when not explicitly specified in queries.
	DefaultConfig(config *TreeOptionsConfig) FindTreeOptionsAPI[TModel, TSearch]
}
