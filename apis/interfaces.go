package apis

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/csv"
	"github.com/ilxqx/vef-framework-go/excel"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/sortx"
)

// Builder defines the interface for building Api endpoints.
type Builder[T any] interface {
	ResourceKind(kind api.Kind) T
	Action(action string) T
	EnableAudit() T
	Timeout(timeout time.Duration) T
	Public() T
	PermToken(token string) T
	RateLimit(maxRequests int, period time.Duration) T
	Build(handler any) api.OperationSpec
}

// Create provides a fluent interface for building create endpoints.
// Supports pre/post processing hooks and transaction-based model creation.
type Create[TModel, TParams any] interface {
	api.OperationsProvider
	Builder[Create[TModel, TParams]]

	// This processor is called before the model is saved to the database.
	WithPreCreate(processor PreCreateProcessor[TModel, TParams]) Create[TModel, TParams]
	// This processor is called after the model is successfully saved within the same transaction.
	WithPostCreate(processor PostCreateProcessor[TModel, TParams]) Create[TModel, TParams]
}

// Update provides a fluent interface for building update endpoints.
// Loads existing model, merges changes, and supports pre/post processing hooks.
type Update[TModel, TParams any] interface {
	api.OperationsProvider
	Builder[Update[TModel, TParams]]

	// This processor is called before the model is updated in the database.
	WithPreUpdate(processor PreUpdateProcessor[TModel, TParams]) Update[TModel, TParams]
	// This processor is called after the model is successfully updated within the same transaction.
	WithPostUpdate(processor PostUpdateProcessor[TModel, TParams]) Update[TModel, TParams]
	DisableDataPerm() Update[TModel, TParams]
}

// Delete provides a fluent interface for building delete endpoints.
// Validates primary key, loads model, and supports pre/post processing hooks.
type Delete[TModel any] interface {
	api.OperationsProvider
	Builder[Delete[TModel]]

	// This processor is called before the model is deleted from the database.
	WithPreDelete(processor PreDeleteProcessor[TModel]) Delete[TModel]
	// This processor is called after the model is successfully deleted within the same transaction.
	WithPostDelete(processor PostDeleteProcessor[TModel]) Delete[TModel]
	DisableDataPerm() Delete[TModel]
}

// CreateMany provides a fluent interface for building batch create endpoints.
// Creates multiple models atomically in a single transaction with pre/post hooks.
type CreateMany[TModel, TParams any] interface {
	api.OperationsProvider
	Builder[CreateMany[TModel, TParams]]

	// This processor is called before the models are saved to the database.
	WithPreCreateMany(processor PreCreateManyProcessor[TModel, TParams]) CreateMany[TModel, TParams]
	// This processor is called after the models are successfully saved within the same transaction.
	WithPostCreateMany(processor PostCreateManyProcessor[TModel, TParams]) CreateMany[TModel, TParams]
}

// UpdateMany provides a fluent interface for building batch update endpoints.
// Updates multiple models atomically with validation, merge, and pre/post hooks.
type UpdateMany[TModel, TParams any] interface {
	api.OperationsProvider
	Builder[UpdateMany[TModel, TParams]]

	// This processor is called before the models are updated in the database.
	WithPreUpdateMany(processor PreUpdateManyProcessor[TModel, TParams]) UpdateMany[TModel, TParams]
	// This processor is called after the models are successfully updated within the same transaction.
	WithPostUpdateMany(processor PostUpdateManyProcessor[TModel, TParams]) UpdateMany[TModel, TParams]
	DisableDataPerm() UpdateMany[TModel, TParams]
}

// DeleteMany provides a fluent interface for building batch delete endpoints.
// Deletes multiple models atomically with validation and pre/post hooks.
type DeleteMany[TModel any] interface {
	api.OperationsProvider
	Builder[DeleteMany[TModel]]

	// This processor is called before the models are deleted from the database.
	WithPreDeleteMany(processor PreDeleteManyProcessor[TModel]) DeleteMany[TModel]
	// This processor is called after the models are successfully deleted within the same transaction.
	WithPostDeleteMany(processor PostDeleteManyProcessor[TModel]) DeleteMany[TModel]
	DisableDataPerm() DeleteMany[TModel]
}

// Find provides a fluent interface for building find endpoints.
// All configuration is done through FindApiOptions passed to NewFindXxxApi constructors.
type Find[TModel, TSearch, TProcessorIn, TApi any] interface {
	Builder[TApi]

	// Setup initializes the FindApi with framework-level options and validates configuration.
	// Must be called before query execution. Config specifies which QueryParts framework options apply to.
	Setup(db orm.DB, config *FindApiConfig, opts ...*FindApiOption) error
	// ConfigureQuery applies FindApiOptions for the specified QueryPart to the query.
	// Returns error if any option applier fails.
	ConfigureQuery(query orm.SelectQuery, search TSearch, meta api.Meta, ctx fiber.Ctx, part QueryPart) error
	// Process applies post-query processing to transform results.
	// Returns input unchanged if no Processor is configured.
	Process(input TProcessorIn, search TSearch, ctx fiber.Ctx) any

	WithProcessor(processor Processor[TProcessorIn, TSearch]) TApi
	// WithOptions appends custom FindApiOptions to the query configuration.
	WithOptions(opts ...*FindApiOption) TApi
	// WithSelect adds a column to the SELECT clause for specified query parts.
	WithSelect(column string, parts ...QueryPart) TApi
	// WithSelectAs adds a column with an alias to the SELECT clause for specified query parts.
	WithSelectAs(column, alias string, parts ...QueryPart) TApi
	WithDefaultSort(sort ...*sortx.OrderSpec) TApi
	// WithCondition adds a WHERE condition using ConditionBuilder for specified query parts.
	WithCondition(fn func(cb orm.ConditionBuilder), parts ...QueryPart) TApi
	// DisableDataPerm disables automatic data permission filtering for this endpoint.
	// IMPORTANT: Must be called before the API is registered (before Setup() is invoked).
	DisableDataPerm() TApi
	// WithRelation adds a relation join to the query for specified query parts.
	WithRelation(relation *orm.RelationSpec, parts ...QueryPart) TApi
	// WithAuditUserNames joins audit user model to populate creator/updater name fields.
	WithAuditUserNames(userModel any, nameColumn ...string) TApi
	// WithQueryApplier adds a custom query modification function for specified query parts.
	WithQueryApplier(applier func(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) error, parts ...QueryPart) TApi
}

// FindOne provides a fluent interface for building find one endpoints.
// Returns a single record matching the search criteria.
type FindOne[TModel, TSearch any] interface {
	api.OperationsProvider
	Find[TModel, TSearch, TModel, FindOne[TModel, TSearch]]
}

// FindAll provides a fluent interface for building find all endpoints.
// Returns all records matching the search criteria (with a safety limit).
type FindAll[TModel, TSearch any] interface {
	api.OperationsProvider
	Find[TModel, TSearch, []TModel, FindAll[TModel, TSearch]]
}

// FindPage provides a fluent interface for building find page endpoints.
// Returns paginated results with total count.
type FindPage[TModel, TSearch any] interface {
	api.OperationsProvider
	Find[TModel, TSearch, []TModel, FindPage[TModel, TSearch]]

	WithDefaultPageSize(size int) FindPage[TModel, TSearch]
}

// FindTree provides a fluent interface for building find tree endpoints.
// Returns hierarchical data using recursive CTEs.
type FindTree[TModel, TSearch any] interface {
	api.OperationsProvider
	Find[TModel, TSearch, []TModel, FindTree[TModel, TSearch]]

	// This column is used to identify individual nodes and establish parent-child relationships.
	WithIDColumn(name string) FindTree[TModel, TSearch]
	// This column establishes the hierarchical relationship between parent and child nodes.
	WithParentIDColumn(name string) FindTree[TModel, TSearch]
}

// FindOptions provides a fluent interface for building find options endpoints.
// Returns a simplified list of options (value, label, description) for dropdowns and selects.
type FindOptions[TModel, TSearch any] interface {
	api.OperationsProvider
	Find[TModel, TSearch, []DataOption, FindOptions[TModel, TSearch]]

	// This mapping provides fallback values for column mapping when not explicitly specified in queries.
	WithDefaultColumnMapping(mapping *DataOptionColumnMapping) FindOptions[TModel, TSearch]
}

// FindTreeOptions provides a fluent interface for building find tree options endpoints.
// Returns hierarchical options using recursive CTEs for tree-structured dropdowns.
type FindTreeOptions[TModel, TSearch any] interface {
	api.OperationsProvider
	Find[TModel, TSearch, []TreeDataOption, FindTreeOptions[TModel, TSearch]]

	// This mapping provides fallback values for label, value, description, and sort columns.
	WithDefaultColumnMapping(mapping *DataOptionColumnMapping) FindTreeOptions[TModel, TSearch]
	WithIDColumn(name string) FindTreeOptions[TModel, TSearch]
	WithParentIDColumn(name string) FindTreeOptions[TModel, TSearch]
}

// Export provides a fluent interface for building export endpoints.
// Queries data based on search conditions and exports to Excel or Csv file.
type Export[TModel, TSearch any] interface {
	api.OperationsProvider
	Find[TModel, TSearch, []TModel, Export[TModel, TSearch]]

	WithDefaultFormat(format TabularFormat) Export[TModel, TSearch]
	WithExcelOptions(opts ...excel.ExportOption) Export[TModel, TSearch]
	WithCsvOptions(opts ...csv.ExportOption) Export[TModel, TSearch]
	WithPreExport(processor PreExportProcessor[TModel, TSearch]) Export[TModel, TSearch]
	WithFilenameBuilder(builder FilenameBuilder[TSearch]) Export[TModel, TSearch]
}

// Import provides a fluent interface for building import endpoints.
// Parses uploaded Excel or Csv file and creates records in database.
type Import[TModel any] interface {
	api.OperationsProvider
	Builder[Import[TModel]]

	WithDefaultFormat(format TabularFormat) Import[TModel]
	WithExcelOptions(opts ...excel.ImportOption) Import[TModel]
	WithCsvOptions(opts ...csv.ImportOption) Import[TModel]
	WithPreImport(processor PreImportProcessor[TModel]) Import[TModel]
	WithPostImport(processor PostImportProcessor[TModel]) Import[TModel]
}
