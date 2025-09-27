package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/orm"
)

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
