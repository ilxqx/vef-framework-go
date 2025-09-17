// Package apis provides preset API implementations for common CRUD operations
// with support for method chaining, filtering, and post-processing.
package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/trans"
	"go.uber.org/fx"
)

var (
	// apisParams holds the dependency injection parameters for the APIs package
	apisParams = new(presetApisParams)

	// Module is the fx module for dependency injection
	Module = fx.Module(
		"vef:apis",
		fx.Populate(apisParams),
	)
)

// presetApisParams defines the dependency injection parameters
type presetApisParams struct {
	fx.In
	Transformer trans.Transformer
}

// QueryApplier is a function that applies additional query conditions.
type QueryApplier[TSearch any] func(search TSearch, ctx fiber.Ctx) orm.ApplyFunc[orm.Query]

// SearchApplier is a function that applies filter conditions to the query builder.
type SearchApplier[TSearch any] func(search TSearch) orm.ApplyFunc[orm.ConditionBuilder]

// FilterApplier is a function that applies filter conditions to the query builder.
type FilterApplier[TSearch any] func(search TSearch, ctx fiber.Ctx) orm.ApplyFunc[orm.ConditionBuilder]

// PostFindProcessor is a function that processes the result after query execution.
type PostFindProcessor[T, R any] func(T, fiber.Ctx) R

// PreCreateProcessor is a function that pre-processes the model before creating it.
type PreCreateProcessor[TModel, TParams any] func(model *TModel, params *TParams, ctx fiber.Ctx) error

// PostCreateProcessor is a function that post-processes the model after creating it.
type PostCreateProcessor[TModel, TParams any] func(model *TModel, params *TParams, ctx fiber.Ctx, tx orm.Db) error

// PreUpdateProcessor is a function that pre-processes the model before updating it.
type PreUpdateProcessor[TModel, TParams any] func(oldModel, model *TModel, params *TParams, ctx fiber.Ctx) error

// PostUpdateProcessor is a function that post-processes the model after updating it.
type PostUpdateProcessor[TModel, TParams any] func(oldModel, model *TModel, params *TParams, ctx fiber.Ctx, tx orm.Db) error

// PreDeleteProcessor is a function that pre-processes the model before deleting it.
type PreDeleteProcessor[TModel any] func(model *TModel, ctx fiber.Ctx) error

// PostDeleteProcessor is a function that post-processes the model after deleting it.
type PostDeleteProcessor[TModel any] func(model *TModel, ctx fiber.Ctx, tx orm.Db) error
