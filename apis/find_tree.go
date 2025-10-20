package apis

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/dbhelpers"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type findTreeApi[TModel, TSearch any] struct {
	FindApi[TModel, TSearch, []TModel, FindTreeApi[TModel, TSearch]]

	idColumn       string
	parentIdColumn string
	treeBuilder    func(flatModels []TModel) []TModel
}

func (a *findTreeApi[TModel, TSearch]) Provide() api.Spec {
	return a.Build(a.findTree)
}

func (a *findTreeApi[TModel, TSearch]) IdColumn(name string) FindTreeApi[TModel, TSearch] {
	a.idColumn = name

	return a
}

func (a *findTreeApi[TModel, TSearch]) ParentIdColumn(name string) FindTreeApi[TModel, TSearch] {
	a.parentIdColumn = name

	return a
}

func (a *findTreeApi[TModel, TSearch]) findTree(db orm.Db) func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error {
	// Initialize FindApi with database schema information
	// This pre-computes expensive operations like default sort configuration
	if err := a.Init(db); err != nil {
		// If initialization fails, return a handler that always returns the error
		return func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error {
			return err
		}
	}

	return func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error {
		var flatModels []TModel

		query := db.NewSelect().
			WithRecursive("tmp_tree", func(query orm.SelectQuery) {
				// Base query
				a.ApplyDataPermission(query.Model((*TModel)(nil)).SelectModelColumns(), ctx)
				a.ApplyConditions(query, search, ctx)
				a.ApplyRelations(query, search, ctx)
				a.ApplyAuditUserRelations(query)
				a.ApplyQuery(query, search, ctx)

				// Recursive part: find all ancestor nodes
				query.UnionAll(func(query orm.SelectQuery) {
					a.ApplyRelations(query.Model((*TModel)(nil)).SelectModelColumns(), search, ctx)
					a.ApplyAuditUserRelations(query)
					a.ApplyQuery(query, search, ctx)
					query.JoinTable(
						"tmp_tree",
						func(cb orm.ConditionBuilder) {
							cb.EqualsColumn(a.idColumn, dbhelpers.ColumnWithAlias(a.parentIdColumn, "tt"))
						},
						"tt",
					)
				})
			}).
			Distinct().
			Table("tmp_tree")

		a.ApplySort(query, search, ctx)

		// Apply default sort if configured
		a.ApplyDefaultSort(query)

		if err := query.Limit(maxQueryLimit).Scan(ctx.Context(), &flatModels); err != nil {
			return err
		}

		// Transform models if there are any results
		if len(flatModels) > 0 {
			for i := range flatModels {
				if err := transformer.Struct(ctx.Context(), &flatModels[i]); err != nil {
					return err
				}
			}
		} else {
			flatModels = make([]TModel, 0)
		}

		// Build tree structure using the configured TreeBuilder
		models := a.treeBuilder(flatModels)

		return result.Ok(a.Process(models, search, ctx)).Response(ctx)
	}
}
