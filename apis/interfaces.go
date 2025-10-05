package apis

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/csv"
	"github.com/ilxqx/vef-framework-go/excel"
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

// CreateAPI provides a fluent interface for building create endpoints.
// Supports pre/post processing hooks and transaction-based model creation.
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

// UpdateAPI provides a fluent interface for building update endpoints.
// Loads existing model, merges changes, and supports pre/post processing hooks.
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

// DeleteAPI provides a fluent interface for building delete endpoints.
// Validates primary key, loads model, and supports pre/post processing hooks.
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

// CreateManyAPI provides a fluent interface for building batch create endpoints.
// Creates multiple models atomically in a single transaction with pre/post hooks.
type CreateManyAPI[TModel, TParams any] interface {
	api.Provider
	APIBuilder[CreateManyAPI[TModel, TParams]]

	// PreCreateMany sets the pre-create processor for batch creation.
	// This processor is called before the models are saved to the database.
	PreCreateMany(processor PreCreateManyProcessor[TModel, TParams]) CreateManyAPI[TModel, TParams]
	// PostCreateMany sets the post-create processor for batch creation.
	// This processor is called after the models are successfully saved within the same transaction.
	PostCreateMany(processor PostCreateManyProcessor[TModel, TParams]) CreateManyAPI[TModel, TParams]
}

// UpdateManyAPI provides a fluent interface for building batch update endpoints.
// Updates multiple models atomically with validation, merge, and pre/post hooks.
type UpdateManyAPI[TModel, TParams any] interface {
	api.Provider
	APIBuilder[UpdateManyAPI[TModel, TParams]]

	// PreUpdateMany sets the pre-update processor for batch update.
	// This processor is called before the models are updated in the database.
	PreUpdateMany(processor PreUpdateManyProcessor[TModel, TParams]) UpdateManyAPI[TModel, TParams]
	// PostUpdateMany sets the post-update processor for batch update.
	// This processor is called after the models are successfully updated within the same transaction.
	PostUpdateMany(processor PostUpdateManyProcessor[TModel, TParams]) UpdateManyAPI[TModel, TParams]
}

// DeleteManyAPI provides a fluent interface for building batch delete endpoints.
// Deletes multiple models atomically with validation and pre/post hooks.
type DeleteManyAPI[TModel any] interface {
	api.Provider
	APIBuilder[DeleteManyAPI[TModel]]

	// PreDeleteMany sets the pre-delete processor for batch deletion.
	// This processor is called before the models are deleted from the database.
	PreDeleteMany(processor PreDeleteManyProcessor[TModel]) DeleteManyAPI[TModel]
	// PostDeleteMany sets the post-delete processor for batch deletion.
	// This processor is called after the models are successfully deleted within the same transaction.
	PostDeleteMany(processor PostDeleteManyProcessor[TModel]) DeleteManyAPI[TModel]
}

// FindAPI provides a fluent interface for building find endpoints.
// Supports custom query modifications, search conditions, filtering, sorting, and post-processing.
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

// FindOneAPI provides a fluent interface for building find one endpoints.
// Supports custom query modifications, search conditions, filtering, sorting, and post-processing.
type FindOneAPI[TModel, TSearch any] interface {
	api.Provider
	FindAPI[TModel, TSearch, TModel, FindOneAPI[TModel, TSearch]]
}

// FindAllAPI provides a fluent interface for building find all endpoints.
// Supports custom query modifications, search conditions, filtering, sorting, and post-processing.
type FindAllAPI[TModel, TSearch any] interface {
	api.Provider
	FindAPI[TModel, TSearch, []TModel, FindAllAPI[TModel, TSearch]]
}

// FindPageAPI provides a fluent interface for building find page endpoints.
// Supports custom query modifications, search conditions, filtering, sorting, and post-processing.
type FindPageAPI[TModel, TSearch any] interface {
	api.Provider
	FindAPI[TModel, TSearch, []TModel, FindPageAPI[TModel, TSearch]]
}

// FindTreeAPI provides a fluent interface for building find tree endpoints.
// Supports custom query modifications, search conditions, filtering, sorting, and post-processing.
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

// FindOptionsAPI provides a fluent interface for building find options endpoints.
// Supports custom query modifications, search conditions, filtering, and post-processing.
// Note: sorting is controlled by OptionFieldMapping.SortField, not by SortApplier.
type FindOptionsAPI[TModel, TSearch any] interface {
	api.Provider
	FindAPI[TModel, TSearch, []Option, FindOptionsAPI[TModel, TSearch]]

	// FieldMapping sets the default field mapping for options queries.
	// This mapping provides fallback values for field mapping when not explicitly specified in queries.
	FieldMapping(mapping *OptionFieldMapping) FindOptionsAPI[TModel, TSearch]
}

// FindTreeOptionsAPI provides a fluent interface for building find tree options endpoints.
// Supports custom query modifications, search conditions, filtering, and post-processing.
// Note: sorting is primarily controlled by TreeOptionFieldMapping.SortField; SortApplier is used only when SortField is empty.
type FindTreeOptionsAPI[TModel, TSearch any] interface {
	api.Provider
	FindAPI[TModel, TSearch, []TreeOption, FindTreeOptionsAPI[TModel, TSearch]]

	// FieldMapping sets the default field mapping for tree options queries.
	// This mapping provides fallback values for field mapping when not explicitly specified in queries.
	FieldMapping(mapping *TreeOptionFieldMapping) FindTreeOptionsAPI[TModel, TSearch]
}

// ExportAPI provides a fluent interface for building export endpoints.
// Queries data based on search conditions and exports to Excel or CSV file.
type ExportAPI[TModel, TSearch any] interface {
	api.Provider
	FindAPI[TModel, TSearch, []TModel, ExportAPI[TModel, TSearch]]

	// Format sets the export format (Excel or CSV). Default is Excel.
	Format(format TabularFormat) ExportAPI[TModel, TSearch]
	// ExcelOptions sets Excel exporter configuration options.
	ExcelOptions(opts ...excel.ExportOption) ExportAPI[TModel, TSearch]
	// CSVOptions sets CSV exporter configuration options.
	CSVOptions(opts ...csv.ExportOption) ExportAPI[TModel, TSearch]
	// PreExport sets a processor to modify data before exporting.
	PreExport(processor PreExportProcessor[TModel, TSearch]) ExportAPI[TModel, TSearch]
	// FilenameBuilder sets a function to generate the export filename dynamically.
	FilenameBuilder(builder FilenameBuilder[TSearch]) ExportAPI[TModel, TSearch]
}

// ImportAPI provides a fluent interface for building import endpoints.
// Parses uploaded Excel or CSV file and creates records in database.
type ImportAPI[TModel, TSearch any] interface {
	api.Provider
	APIBuilder[ImportAPI[TModel, TSearch]]

	// Format sets the import format (Excel or CSV). Default is Excel.
	Format(format TabularFormat) ImportAPI[TModel, TSearch]
	// ExcelOptions sets Excel importer configuration options.
	ExcelOptions(opts ...excel.ImportOption) ImportAPI[TModel, TSearch]
	// CSVOptions sets CSV importer configuration options.
	CSVOptions(opts ...csv.ImportOption) ImportAPI[TModel, TSearch]
	// PreImport sets a processor to validate or modify data before saving.
	PreImport(processor PreImportProcessor[TModel, TSearch]) ImportAPI[TModel, TSearch]
	// PostImport sets a processor to perform additional actions after import.
	PostImport(processor PostImportProcessor[TModel, TSearch]) ImportAPI[TModel, TSearch]
}
