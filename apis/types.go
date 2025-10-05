package apis

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/orm"
)

// CreateManyParams is a wrapper type for batch create parameters.
// It contains a List field holding the slice of individual item parameters.
type CreateManyParams[TParams any] struct {
	api.In

	List []TParams `json:"list" validate:"required,min=1,dive" label_i18n:"batch_create_list"`
}

// UpdateManyParams is a wrapper type for batch update parameters.
// It contains a List field holding the slice of individual item parameters.
type UpdateManyParams[TParams any] struct {
	api.In

	List []TParams `json:"list" validate:"required,min=1,dive" label_i18n:"batch_update_list"`
}

// DeleteManyParams is a wrapper type for batch delete parameters.
// It contains a PKs field holding the slice of primary key values.
// For single primary key models: PKs can be []any with direct values (e.g., ["id1", "id2"])
// For composite primary key models: PKs should be []map[string]any with each map containing all PK fields.
type DeleteManyParams struct {
	api.In

	PKs []any `json:"pks" validate:"required,min=1" label_i18n:"batch_delete_pks"`
}

// QueryApplier applies additional query modifications like joins, subqueries, etc.
// Used for complex query customization beyond basic filtering and sorting.
type QueryApplier[TSearch any] func(search TSearch, ctx fiber.Ctx) orm.ApplyFunc[orm.SelectQuery]

// SearchApplier applies search conditions based on the search parameters.
// This is the primary mechanism for converting search inputs to database conditions.
type SearchApplier[TSearch any] func(search TSearch) orm.ApplyFunc[orm.ConditionBuilder]

// FilterApplier applies additional filtering logic with access to request context.
// Useful for context-dependent filters like user permissions or tenant isolation.
type FilterApplier[TSearch any] func(search TSearch, ctx fiber.Ctx) orm.ApplyFunc[orm.ConditionBuilder]

// SortApplier applies custom ordering logic to query results.
// Provides access to both search parameters and request context for dynamic sorting.
type SortApplier[TSearch any] func(search TSearch, ctx fiber.Ctx) orm.ApplyFunc[Sorter]

// Processor transforms query results after execution but before JSON serialization.
// Commonly used for data formatting, field selection, or computed properties.
type Processor[TIn, TSearch any] func(input TIn, search TSearch, ctx fiber.Ctx) any

// PreCreateProcessor handles business logic before model creation.
// Common uses: validation, default values, related data setup.
type PreCreateProcessor[TModel, TParams any] func(model *TModel, params *TParams, ctx fiber.Ctx, db orm.Db) error

// PostCreateProcessor handles side effects after successful model creation.
// Runs within the same transaction. Uses: audit logging, notifications, cache updates.
type PostCreateProcessor[TModel, TParams any] func(model *TModel, params *TParams, ctx fiber.Ctx, tx orm.Db) error

// PreUpdateProcessor handles business logic before model update.
// Provides both old and new model states for comparison and validation.
type PreUpdateProcessor[TModel, TParams any] func(oldModel, model *TModel, params *TParams, ctx fiber.Ctx, db orm.Db) error

// PostUpdateProcessor handles side effects after successful model update.
// Runs within the same transaction. Uses: audit trails, change notifications.
type PostUpdateProcessor[TModel, TParams any] func(oldModel, model *TModel, params *TParams, ctx fiber.Ctx, tx orm.Db) error

// PreDeleteProcessor handles validation and checks before model deletion.
// Common uses: referential integrity checks, soft delete logic.
type PreDeleteProcessor[TModel any] func(model *TModel, ctx fiber.Ctx, db orm.Db) error

// PostDeleteProcessor handles cleanup tasks after successful deletion.
// Runs within the same transaction. Uses: cascade operations, audit logging.
type PostDeleteProcessor[TModel any] func(model *TModel, ctx fiber.Ctx, tx orm.Db) error

// PreCreateManyProcessor handles business logic before batch model creation.
// Common uses: batch validation, default values, related data setup.
type PreCreateManyProcessor[TModel, TParams any] func(models []TModel, paramsList []TParams, ctx fiber.Ctx, db orm.Db) error

// PostCreateManyProcessor handles side effects after successful batch model creation.
// Runs within the same transaction. Uses: audit logging, notifications, cache updates.
type PostCreateManyProcessor[TModel, TParams any] func(models []TModel, paramsList []TParams, ctx fiber.Ctx, tx orm.Db) error

// PreUpdateManyProcessor handles business logic before batch model update.
// Provides both old and new model states for comparison and validation.
type PreUpdateManyProcessor[TModel, TParams any] func(oldModels, models []TModel, paramsList []TParams, ctx fiber.Ctx, db orm.Db) error

// PostUpdateManyProcessor handles side effects after successful batch model update.
// Runs within the same transaction. Uses: audit trails, change notifications.
type PostUpdateManyProcessor[TModel, TParams any] func(oldModels, models []TModel, paramsList []TParams, ctx fiber.Ctx, tx orm.Db) error

// PreDeleteManyProcessor handles validation and checks before batch model deletion.
// Common uses: referential integrity checks, soft delete logic.
type PreDeleteManyProcessor[TModel any] func(models []TModel, ctx fiber.Ctx, db orm.Db) error

// PostDeleteManyProcessor handles cleanup tasks after successful batch deletion.
// Runs within the same transaction. Uses: cascade operations, audit logging.
type PostDeleteManyProcessor[TModel any] func(models []TModel, ctx fiber.Ctx, tx orm.Db) error

// PreExportProcessor handles data modification before exporting to Excel.
// Common uses: data formatting, field filtering, additional data loading.
type PreExportProcessor[TModel, TSearch any] func(models []TModel, search TSearch, ctx fiber.Ctx, db orm.Db) error

// FilenameBuilder generates the filename for exported Excel files based on search parameters.
// Allows dynamic filename generation with timestamp, filters, etc.
type FilenameBuilder[TSearch any] func(search TSearch, ctx fiber.Ctx) string

// PreImportProcessor handles validation and transformation before saving imported data.
// Common uses: data validation, default values, duplicate checking.
type PreImportProcessor[TModel, TSearch any] func(models []TModel, search TSearch, ctx fiber.Ctx, db orm.Db) error

// PostImportProcessor handles side effects after successful import.
// Runs within the same transaction. Uses: audit logging, notifications, cache updates.
type PostImportProcessor[TModel, TSearch any] func(models []TModel, search TSearch, ctx fiber.Ctx, tx orm.Db) error
