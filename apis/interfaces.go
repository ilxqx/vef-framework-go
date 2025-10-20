package apis

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/csv"
	"github.com/ilxqx/vef-framework-go/excel"
	"github.com/ilxqx/vef-framework-go/orm"
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
	PreCreate(processor PreCreateProcessor[TModel, TParams]) CreateApi[TModel, TParams]
	// PostCreate sets the post-create processor for the CreateApi.
	// This processor is called after the model is successfully saved within the same transaction.
	PostCreate(processor PostCreateProcessor[TModel, TParams]) CreateApi[TModel, TParams]
}

// UpdateApi provides a fluent interface for building update endpoints.
// Loads existing model, merges changes, and supports pre/post processing hooks.
type UpdateApi[TModel, TParams any] interface {
	api.Provider
	ApiBuilder[UpdateApi[TModel, TParams]]

	// PreUpdate sets the pre-update processor for the UpdateApi.
	// This processor is called before the model is updated in the database.
	PreUpdate(processor PreUpdateProcessor[TModel, TParams]) UpdateApi[TModel, TParams]
	// PostUpdate sets the post-update processor for the UpdateApi.
	// This processor is called after the model is successfully updated within the same transaction.
	PostUpdate(processor PostUpdateProcessor[TModel, TParams]) UpdateApi[TModel, TParams]
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
	PreDelete(processor PreDeleteProcessor[TModel]) DeleteApi[TModel]
	// PostDelete sets the post-delete processor for the DeleteApi.
	// This processor is called after the model is successfully deleted within the same transaction.
	PostDelete(processor PostDeleteProcessor[TModel]) DeleteApi[TModel]
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
	PreCreateMany(processor PreCreateManyProcessor[TModel, TParams]) CreateManyApi[TModel, TParams]
	// PostCreateMany sets the post-create processor for batch creation.
	// This processor is called after the models are successfully saved within the same transaction.
	PostCreateMany(processor PostCreateManyProcessor[TModel, TParams]) CreateManyApi[TModel, TParams]
}

// UpdateManyApi provides a fluent interface for building batch update endpoints.
// Updates multiple models atomically with validation, merge, and pre/post hooks.
type UpdateManyApi[TModel, TParams any] interface {
	api.Provider
	ApiBuilder[UpdateManyApi[TModel, TParams]]

	// PreUpdateMany sets the pre-update processor for batch update.
	// This processor is called before the models are updated in the database.
	PreUpdateMany(processor PreUpdateManyProcessor[TModel, TParams]) UpdateManyApi[TModel, TParams]
	// PostUpdateMany sets the post-update processor for batch update.
	// This processor is called after the models are successfully updated within the same transaction.
	PostUpdateMany(processor PostUpdateManyProcessor[TModel, TParams]) UpdateManyApi[TModel, TParams]
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
	PreDeleteMany(processor PreDeleteManyProcessor[TModel]) DeleteManyApi[TModel]
	// PostDeleteMany sets the post-delete processor for batch deletion.
	// This processor is called after the models are successfully deleted within the same transaction.
	PostDeleteMany(processor PostDeleteManyProcessor[TModel]) DeleteManyApi[TModel]
	// DisableDataPerm disables data permission filtering for this endpoint.
	// By default, data permission filtering is enabled when loading existing models for batch deletion.
	DisableDataPerm() DeleteManyApi[TModel]
}

// FindApi provides a fluent interface for building find endpoints.
// Supports custom query modifications, search conditions, filtering, sorting, and post-processing.
type FindApi[TModel, TSearch, TProcessorIn, TApi any] interface {
	ApiBuilder[TApi]

	// QueryApplier sets a custom query applier function for additional query modifications.
	QueryApplier(applier QueryApplier[TSearch]) TApi
	// FilterApplier sets a custom filter applier function for additional filtering logic.
	FilterApplier(applier FilterApplier[TSearch]) TApi
	// SortApplier sets a custom sort applier function for additional order modifications.
	SortApplier(applier SortApplier[TSearch]) TApi
	// Relations adds RelationSpec configurations to be included in the query.
	Relations(relations ...orm.RelationSpec) TApi
	// Processor sets a post-processing function to transform query results.
	Processor(processor Processor[TProcessorIn, TSearch]) TApi
	// DisableDataPerm disables data permission filtering for this endpoint.
	// By default, data permission filtering is enabled for all Find operations.
	DisableDataPerm() TApi
	// WithAuditUserNames enables querying audit user names (created_by_name, updated_by_name).
	// The userModel parameter specifies the user table model (e.g., (*User)(nil)).
	// The optional nameColumn parameter specifies the name column in user table (default: "name").
	// This method performs LEFT JOINs with the user table to populate audit user names.
	// Note: Only single primary key is supported; composite primary keys will result in an error.
	WithAuditUserNames(userModel any, nameColumn ...string) TApi

	// Init initializes the FindApi with database schema information.
	// This method should be called once in factory functions to pre-compute
	// and cache expensive operations (e.g., schema analysis, default sort configuration).
	// It's safe to call multiple times - subsequent calls are no-ops.
	// Returns error if initialization fails (e.g., schema validation errors).
	Init(db orm.Db) error
	// BuildQuery creates a new SelectQuery with the given model and applies all query configurations.
	// This is the main entry point for constructing a complete query with search conditions, filters,
	// relations, sorting, and data permissions. It delegates to ConfigureQuery for the actual setup.
	BuildQuery(db orm.Db, model any, search TSearch, ctx fiber.Ctx) orm.SelectQuery
	// ConfigureQuery applies all query configuration steps to an existing SelectQuery in the correct order.
	// This method orchestrates the query building pipeline:
	//   1. Sets the model (required for data permission)
	//   2. Applies data permission filtering (if enabled)
	//   3. Applies search conditions and custom filters (via ApplyConditions)
	//   4. Applies relation joins (via ApplyRelations)
	//   5. Applies audit user relations for created_by_name/updated_by_name (if configured)
	//   6. Applies custom query modifications (via ApplyQuery)
	//   7. Applies sorting logic (via ApplySort)
	// Note: Default sorting (ORDER BY created_at DESC) must be applied separately via ApplyDefaultSort.
	ConfigureQuery(query orm.SelectQuery, model any, search TSearch, ctx fiber.Ctx)
	// Process applies post-query processing to transform or enrich the query results.
	// This method is called after data is fetched from the database but before returning to the client.
	// If no Processor is configured via Processor(), it returns the input unchanged.
	// Common use cases: data masking, computed fields, nested structure transformation, aggregation.
	Process(input TProcessorIn, search TSearch, ctx fiber.Ctx) any
	// ApplySearch applies search conditions from the SearchApplier to the query.
	// This method wraps the search applier in a WHERE clause using the ConditionBuilder.
	// SearchApplier typically handles struct-tagged search conditions (e.g., `search:"eq"`, `search:"contains"`).
	// Note: This method is rarely called directly; use ApplyConditions instead which combines search and filter logic.
	ApplySearch(query orm.SelectQuery, search TSearch, ctx fiber.Ctx)
	// ApplyFilter applies custom filter conditions from the FilterApplier to the query.
	// This method wraps the filter applier in a WHERE clause using the ConditionBuilder.
	// FilterApplier is typically used for dynamic business logic filters (e.g., user-specific filtering).
	// Note: This method is rarely called directly; use ApplyConditions instead which combines search and filter logic.
	ApplyFilter(query orm.SelectQuery, search TSearch, ctx fiber.Ctx)
	// ApplyConditions applies both search conditions and custom filters to the query in a single WHERE clause.
	// This is the primary method for applying filtering logic, combining:
	//   - Search conditions from struct tags (always applied)
	//   - Custom filter logic from FilterApplier (applied if configured)
	// Both are wrapped in a single ConditionBuilder to ensure proper SQL generation (AND/OR logic).
	ApplyConditions(query orm.SelectQuery, search TSearch, ctx fiber.Ctx)
	// ApplyQuery applies custom query modifications via the QueryApplier if configured.
	// This provides a flexible extension point for advanced query customization that doesn't fit
	// into the standard search/filter/sort pipeline (e.g., custom subqueries, DISTINCT, GROUP BY, HAVING).
	// If no QueryApplier is set via QueryApplier(), this is a no-op.
	ApplyQuery(query orm.SelectQuery, search TSearch, ctx fiber.Ctx)
	// ApplySort applies sorting logic to the query via the SortApplier if configured.
	// If a custom SortApplier is set via SortApplier(), it will be called to modify the query's ORDER BY clause.
	// If no SortApplier is configured, this is a no-op (rely on ApplyDefaultSort or pagination sorting).
	ApplySort(query orm.SelectQuery, search TSearch, ctx fiber.Ctx)
	// ApplyRelations applies relation joins to the query based on the Relations configuration.
	// This method calls JoinRelations with all RelationSpecs configured via Relations().
	// Each RelationSpec defines a join (INNER/LEFT/RIGHT) with another table, including selected columns and aliases.
	// If no relations are configured, this is a no-op.
	ApplyRelations(query orm.SelectQuery, search TSearch, ctx fiber.Ctx)
	// ApplyAuditUserRelations applies RelationSpec configurations to query audit user names.
	// This method creates LEFT JOIN relations with the user model to populate:
	//   - created_by_name (creator's name)
	//   - updated_by_name (updater's name)
	// It is called automatically during query building if WithAuditUserNames was configured.
	// The user model must have a single primary key; composite primary keys are not supported.
	// If no audit user model is configured via WithAuditUserNames(), this is a no-op.
	ApplyAuditUserRelations(query orm.SelectQuery)
	// ApplyDataPermission applies data permission filtering to the query if enabled.
	// This method retrieves the DataPermissionApplier from context and applies row-level security rules.
	// Data permission filtering is enabled by default and can be disabled via DisableDataPerm().
	// The applier is typically injected by middleware and defines which records the current user can access.
	// If data permission is disabled or no applier is available in context, this is a no-op.
	// Errors during permission application are logged but do not fail the request.
	ApplyDataPermission(query orm.SelectQuery, ctx fiber.Ctx)
	// ShouldApplyDefaultSort returns whether default sorting should be applied to the query.
	// Default sorting (ORDER BY created_at DESC) is enabled when:
	//   1. No custom SortApplier is configured
	//   2. The model has a created_at field
	// This is pre-computed during Init() for efficiency and used by ApplyDefaultSort.
	ShouldApplyDefaultSort() bool
	// ApplyDefaultSort applies default sorting to the query if applicable.
	// This method checks if default sorting should be applied (based on SortApplier and model schema)
	// and adds ORDER BY created_at DESC if applicable.
	ApplyDefaultSort(query orm.SelectQuery)
}

// FindOneApi provides a fluent interface for building find one endpoints.
// Supports custom query modifications, search conditions, filtering, sorting, and post-processing.
type FindOneApi[TModel, TSearch any] interface {
	api.Provider
	FindApi[TModel, TSearch, TModel, FindOneApi[TModel, TSearch]]
}

// FindAllApi provides a fluent interface for building find all endpoints.
// Supports custom query modifications, search conditions, filtering, sorting, and post-processing.
type FindAllApi[TModel, TSearch any] interface {
	api.Provider
	FindApi[TModel, TSearch, []TModel, FindAllApi[TModel, TSearch]]
}

// FindPageApi provides a fluent interface for building find page endpoints.
// Supports custom query modifications, search conditions, filtering, sorting, and post-processing.
type FindPageApi[TModel, TSearch any] interface {
	api.Provider
	FindApi[TModel, TSearch, []TModel, FindPageApi[TModel, TSearch]]
}

// FindTreeApi provides a fluent interface for building find tree endpoints.
// Supports custom query modifications, search conditions, filtering, sorting, and post-processing.
type FindTreeApi[TModel, TSearch any] interface {
	api.Provider
	FindApi[TModel, TSearch, []TModel, FindTreeApi[TModel, TSearch]]

	// IdColumn sets the column name used as the node Id in tree structures.
	// This column is used to identify individual nodes and establish parent-child relationships.
	IdColumn(name string) FindTreeApi[TModel, TSearch]
	// ParentIdColumn sets the column name used to reference parent nodes in tree structures.
	// This column establishes the hierarchical relationship between parent and child nodes.
	ParentIdColumn(name string) FindTreeApi[TModel, TSearch]
}

// FindOptionsApi provides a fluent interface for building find options endpoints.
// Supports custom query modifications, search conditions, filtering, and post-processing.
// Note: sorting is controlled by OptionColumnMapping.SortColumn, not by SortApplier.
type FindOptionsApi[TModel, TSearch any] interface {
	api.Provider
	FindApi[TModel, TSearch, []Option, FindOptionsApi[TModel, TSearch]]

	// ColumnMapping sets the default column mapping for options queries.
	// This mapping provides fallback values for column mapping when not explicitly specified in queries.
	ColumnMapping(mapping *OptionColumnMapping) FindOptionsApi[TModel, TSearch]
}

// FindTreeOptionsApi provides a fluent interface for building find tree options endpoints.
// Supports custom query modifications, search conditions, filtering, and post-processing.
// Note: sorting is primarily controlled by TreeOptionColumnMapping.SortColumn; SortApplier is used only when SortColumn is empty.
type FindTreeOptionsApi[TModel, TSearch any] interface {
	api.Provider
	FindApi[TModel, TSearch, []TreeOption, FindTreeOptionsApi[TModel, TSearch]]

	// ColumnMapping sets the default column mapping for tree options queries.
	// This mapping provides fallback values for column mapping when not explicitly specified in queries.
	ColumnMapping(mapping *TreeOptionColumnMapping) FindTreeOptionsApi[TModel, TSearch]
}

// ExportApi provides a fluent interface for building export endpoints.
// Queries data based on search conditions and exports to Excel or CSV file.
type ExportApi[TModel, TSearch any] interface {
	api.Provider
	FindApi[TModel, TSearch, []TModel, ExportApi[TModel, TSearch]]

	// Format sets the export format (Excel or CSV). Default is Excel.
	Format(format TabularFormat) ExportApi[TModel, TSearch]
	// ExcelOptions sets Excel exporter configuration options.
	ExcelOptions(opts ...excel.ExportOption) ExportApi[TModel, TSearch]
	// CSVOptions sets CSV exporter configuration options.
	CSVOptions(opts ...csv.ExportOption) ExportApi[TModel, TSearch]
	// PreExport sets a processor to modify data before exporting.
	PreExport(processor PreExportProcessor[TModel, TSearch]) ExportApi[TModel, TSearch]
	// FilenameBuilder sets a function to generate the export filename dynamically.
	FilenameBuilder(builder FilenameBuilder[TSearch]) ExportApi[TModel, TSearch]
}

// ImportApi provides a fluent interface for building import endpoints.
// Parses uploaded Excel or CSV file and creates records in database.
type ImportApi[TModel, TSearch any] interface {
	api.Provider
	ApiBuilder[ImportApi[TModel, TSearch]]

	// Format sets the import format (Excel or CSV). Default is Excel.
	Format(format TabularFormat) ImportApi[TModel, TSearch]
	// ExcelOptions sets Excel importer configuration options.
	ExcelOptions(opts ...excel.ImportOption) ImportApi[TModel, TSearch]
	// CSVOptions sets CSV importer configuration options.
	CSVOptions(opts ...csv.ImportOption) ImportApi[TModel, TSearch]
	// PreImport sets a processor to validate or modify data before saving.
	PreImport(processor PreImportProcessor[TModel, TSearch]) ImportApi[TModel, TSearch]
	// PostImport sets a processor to perform additional actions after import.
	PostImport(processor PostImportProcessor[TModel, TSearch]) ImportApi[TModel, TSearch]
}
