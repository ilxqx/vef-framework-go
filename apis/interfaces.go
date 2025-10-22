package apis

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/csv"
	"github.com/ilxqx/vef-framework-go/excel"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/sort"
)

// ApiBuilder defines the interface for building Api endpoint.
// It provides a fluent Api for configuring all aspects of an Api endpoint.
type ApiBuilder[T any] interface {
	// Action sets the action name for the Api endpoint.
	Action(action string) T
	// EnableAudit enables audit logging for this endpoint.
	EnableAudit() T
	// Timeout sets the request timeout duration.
	Timeout(timeout time.Duration) T
	// Public sets this endpoint is publicly accessible.
	Public() T
	// PermToken sets the permission token required for access.
	PermToken(token string) T
	// RateLimit sets the rate limit configuration for this endpoint.
	RateLimit(max int, expiration time.Duration) T
	// Build builds the Api endpoint specification.
	Build(handler any) api.Spec
}

// CreateApi provides a fluent interface for building create endpoints.
// Supports pre/post processing hooks and transaction-based model creation.
type CreateApi[TModel, TParams any] interface {
	api.Provider
	ApiBuilder[CreateApi[TModel, TParams]]

	// PreCreate sets the pre-create processor for the CreateApi.
	// This processor is called before the model is saved to the database.
	WithPreCreate(processor PreCreateProcessor[TModel, TParams]) CreateApi[TModel, TParams]
	// PostCreate sets the post-create processor for the CreateApi.
	// This processor is called after the model is successfully saved within the same transaction.
	WithPostCreate(processor PostCreateProcessor[TModel, TParams]) CreateApi[TModel, TParams]
}

// UpdateApi provides a fluent interface for building update endpoints.
// Loads existing model, merges changes, and supports pre/post processing hooks.
type UpdateApi[TModel, TParams any] interface {
	api.Provider
	ApiBuilder[UpdateApi[TModel, TParams]]

	// PreUpdate sets the pre-update processor for the UpdateApi.
	// This processor is called before the model is updated in the database.
	WithPreUpdate(processor PreUpdateProcessor[TModel, TParams]) UpdateApi[TModel, TParams]
	// PostUpdate sets the post-update processor for the UpdateApi.
	// This processor is called after the model is successfully updated within the same transaction.
	WithPostUpdate(processor PostUpdateProcessor[TModel, TParams]) UpdateApi[TModel, TParams]
	// DisableDataPerm disables data permission filtering for this endpoint.
	// By default, data permission filtering is enabled when loading the existing model for update.
	DisableDataPerm() UpdateApi[TModel, TParams]
}

// DeleteApi provides a fluent interface for building delete endpoints.
// Validates primary key, loads model, and supports pre/post processing hooks.
type DeleteApi[TModel any] interface {
	api.Provider
	ApiBuilder[DeleteApi[TModel]]

	// PreDelete sets the pre-delete processor for the DeleteApi.
	// This processor is called before the model is deleted from the database.
	WithPreDelete(processor PreDeleteProcessor[TModel]) DeleteApi[TModel]
	// PostDelete sets the post-delete processor for the DeleteApi.
	// This processor is called after the model is successfully deleted within the same transaction.
	WithPostDelete(processor PostDeleteProcessor[TModel]) DeleteApi[TModel]
	// DisableDataPerm disables data permission filtering for this endpoint.
	// By default, data permission filtering is enabled when loading the existing model for deletion.
	DisableDataPerm() DeleteApi[TModel]
}

// CreateManyApi provides a fluent interface for building batch create endpoints.
// Creates multiple models atomically in a single transaction with pre/post hooks.
type CreateManyApi[TModel, TParams any] interface {
	api.Provider
	ApiBuilder[CreateManyApi[TModel, TParams]]

	// PreCreateMany sets the pre-create processor for batch creation.
	// This processor is called before the models are saved to the database.
	WithPreCreateMany(processor PreCreateManyProcessor[TModel, TParams]) CreateManyApi[TModel, TParams]
	// PostCreateMany sets the post-create processor for batch creation.
	// This processor is called after the models are successfully saved within the same transaction.
	WithPostCreateMany(processor PostCreateManyProcessor[TModel, TParams]) CreateManyApi[TModel, TParams]
}

// UpdateManyApi provides a fluent interface for building batch update endpoints.
// Updates multiple models atomically with validation, merge, and pre/post hooks.
type UpdateManyApi[TModel, TParams any] interface {
	api.Provider
	ApiBuilder[UpdateManyApi[TModel, TParams]]

	// PreUpdateMany sets the pre-update processor for batch update.
	// This processor is called before the models are updated in the database.
	WithPreUpdateMany(processor PreUpdateManyProcessor[TModel, TParams]) UpdateManyApi[TModel, TParams]
	// PostUpdateMany sets the post-update processor for batch update.
	// This processor is called after the models are successfully updated within the same transaction.
	WithPostUpdateMany(processor PostUpdateManyProcessor[TModel, TParams]) UpdateManyApi[TModel, TParams]
	// DisableDataPerm disables data permission filtering for this endpoint.
	// By default, data permission filtering is enabled when loading existing models for batch update.
	DisableDataPerm() UpdateManyApi[TModel, TParams]
}

// DeleteManyApi provides a fluent interface for building batch delete endpoints.
// Deletes multiple models atomically with validation and pre/post hooks.
type DeleteManyApi[TModel any] interface {
	api.Provider
	ApiBuilder[DeleteManyApi[TModel]]

	// PreDeleteMany sets the pre-delete processor for batch deletion.
	// This processor is called before the models are deleted from the database.
	WithPreDeleteMany(processor PreDeleteManyProcessor[TModel]) DeleteManyApi[TModel]
	// PostDeleteMany sets the post-delete processor for batch deletion.
	// This processor is called after the models are successfully deleted within the same transaction.
	WithPostDeleteMany(processor PostDeleteManyProcessor[TModel]) DeleteManyApi[TModel]
	// DisableDataPerm disables data permission filtering for this endpoint.
	// By default, data permission filtering is enabled when loading existing models for batch deletion.
	DisableDataPerm() DeleteManyApi[TModel]
}

// FindApi provides a fluent interface for building find endpoints.
// All configuration is done through FindApiOptions passed to NewFindXxxApi constructors.
type FindApi[TModel, TSearch, TProcessorIn, TApi any] interface {
	ApiBuilder[TApi]

	// Setup initializes the FindApi with framework-level options and validates configuration.
	// Must be called before query execution. Config specifies which QueryParts framework options apply to.
	Setup(db orm.Db, config *FindApiConfig, opts ...*FindApiOption) error
	// ConfigureQuery applies FindApiOptions for the specified QueryPart to the query.
	// Returns error if any option applier fails.
	ConfigureQuery(query orm.SelectQuery, search TSearch, ctx fiber.Ctx, part QueryPart) error
	// Process applies post-query processing to transform results.
	// Returns input unchanged if no Processor is configured.
	Process(input TProcessorIn, search TSearch, ctx fiber.Ctx) any

	// WithProcessor sets a custom processor to transform query results before response serialization.
	WithProcessor(processor Processor[TProcessorIn, TSearch]) TApi
	// WithOptions appends custom FindApiOptions to the query configuration.
	WithOptions(opts ...*FindApiOption) TApi
	// WithSelect adds a column to the SELECT clause for specified query parts.
	WithSelect(column string, parts ...QueryPart) TApi
	// WithSelectAs adds a column with an alias to the SELECT clause for specified query parts.
	WithSelectAs(column, alias string, parts ...QueryPart) TApi
	// WithDefaultSort sets default sorting order. Pass no args to disable, pass specs to customize.
	WithDefaultSort(sort ...*sort.OrderSpec) TApi
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

// FindOneApi provides a fluent interface for building find one endpoints.
// Returns a single record matching the search criteria.
type FindOneApi[TModel, TSearch any] interface {
	api.Provider
	FindApi[TModel, TSearch, TModel, FindOneApi[TModel, TSearch]]
}

// FindAllApi provides a fluent interface for building find all endpoints.
// Returns all records matching the search criteria (with a safety limit).
type FindAllApi[TModel, TSearch any] interface {
	api.Provider
	FindApi[TModel, TSearch, []TModel, FindAllApi[TModel, TSearch]]
}

// FindPageApi provides a fluent interface for building find page endpoints.
// Returns paginated results with total count.
type FindPageApi[TModel, TSearch any] interface {
	api.Provider
	FindApi[TModel, TSearch, []TModel, FindPageApi[TModel, TSearch]]

	// WithDefaultPageSize sets the default page size when not specified in the request.
	WithDefaultPageSize(size int) FindPageApi[TModel, TSearch]
}

// FindTreeApi provides a fluent interface for building find tree endpoints.
// Returns hierarchical data using recursive CTEs.
type FindTreeApi[TModel, TSearch any] interface {
	api.Provider
	FindApi[TModel, TSearch, []TModel, FindTreeApi[TModel, TSearch]]

	// IdColumn sets the column name used as the node Id in tree structures.
	// This column is used to identify individual nodes and establish parent-child relationships.
	WithIdColumn(name string) FindTreeApi[TModel, TSearch]
	// ParentIdColumn sets the column name used to reference parent nodes in tree structures.
	// This column establishes the hierarchical relationship between parent and child nodes.
	WithParentIdColumn(name string) FindTreeApi[TModel, TSearch]
}

// FindOptionsApi provides a fluent interface for building find options endpoints.
// Returns a simplified list of options (value, label, description) for dropdowns and selects.
type FindOptionsApi[TModel, TSearch any] interface {
	api.Provider
	FindApi[TModel, TSearch, []DataOption, FindOptionsApi[TModel, TSearch]]

	// ColumnMapping sets the default column mapping for options queries.
	// This mapping provides fallback values for column mapping when not explicitly specified in queries.
	WithDefaultColumnMapping(mapping *DataOptionColumnMapping) FindOptionsApi[TModel, TSearch]
}

// FindTreeOptionsApi provides a fluent interface for building find tree options endpoints.
// Returns hierarchical options using recursive CTEs for tree-structured dropdowns.
type FindTreeOptionsApi[TModel, TSearch any] interface {
	api.Provider
	FindApi[TModel, TSearch, []TreeDataOption, FindTreeOptionsApi[TModel, TSearch]]

	// WithDefaultColumnMapping sets the default column mapping for data option fields.
	// This mapping provides fallback values for label, value, description, and sort columns.
	WithDefaultColumnMapping(mapping *DataOptionColumnMapping) FindTreeOptionsApi[TModel, TSearch]
	// WithIdColumn sets the column name used as the node Id in tree structures.
	WithIdColumn(name string) FindTreeOptionsApi[TModel, TSearch]
	// WithParentIdColumn sets the column name used to reference parent nodes in tree structures.
	WithParentIdColumn(name string) FindTreeOptionsApi[TModel, TSearch]
}

// ExportApi provides a fluent interface for building export endpoints.
// Queries data based on search conditions and exports to Excel or Csv file.
type ExportApi[TModel, TSearch any] interface {
	api.Provider
	FindApi[TModel, TSearch, []TModel, ExportApi[TModel, TSearch]]

	// WithDefaultFormat sets the default export format (Excel or Csv). Default is Excel.
	WithDefaultFormat(format TabularFormat) ExportApi[TModel, TSearch]
	// ExcelOptions sets Excel exporter configuration options.
	WithExcelOptions(opts ...excel.ExportOption) ExportApi[TModel, TSearch]
	// CsvOptions sets Csv exporter configuration options.
	WithCsvOptions(opts ...csv.ExportOption) ExportApi[TModel, TSearch]
	// PreExport sets a processor to modify data before exporting.
	WithPreExport(processor PreExportProcessor[TModel, TSearch]) ExportApi[TModel, TSearch]
	// FilenameBuilder sets a function to generate the export filename dynamically.
	WithFilenameBuilder(builder FilenameBuilder[TSearch]) ExportApi[TModel, TSearch]
}

// ImportApi provides a fluent interface for building import endpoints.
// Parses uploaded Excel or Csv file and creates records in database.
type ImportApi[TModel any] interface {
	api.Provider
	ApiBuilder[ImportApi[TModel]]

	// WithDefaultFormat sets the default import format (Excel or Csv). Default is Excel.
	WithDefaultFormat(format TabularFormat) ImportApi[TModel]
	// ExcelOptions sets Excel importer configuration options.
	WithExcelOptions(opts ...excel.ImportOption) ImportApi[TModel]
	// CsvOptions sets Csv importer configuration options.
	WithCsvOptions(opts ...csv.ImportOption) ImportApi[TModel]
	// PreImport sets a processor to validate or modify data before saving.
	WithPreImport(processor PreImportProcessor[TModel]) ImportApi[TModel]
	// PostImport sets a processor to perform additional actions after import.
	WithPostImport(processor PostImportProcessor[TModel]) ImportApi[TModel]
}
