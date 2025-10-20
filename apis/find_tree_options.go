package apis

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/dbhelpers"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/treebuilder"
)

type findTreeOptionsApi[TModel, TSearch any] struct {
	FindApi[TModel, TSearch, []TreeOption, FindTreeOptionsApi[TModel, TSearch]]

	columnMapping *TreeOptionColumnMapping
}

func (a *findTreeOptionsApi[TModel, TSearch]) Provide() api.Spec {
	return a.Build(a.findTreeOptions)
}

func (a *findTreeOptionsApi[TModel, TSearch]) ColumnMapping(mapping *TreeOptionColumnMapping) FindTreeOptionsApi[TModel, TSearch] {
	a.columnMapping = mapping

	return a
}

func (a *findTreeOptionsApi[TModel, TSearch]) findTreeOptions(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, params TreeOptionParams, search TSearch) error, error) {
	if err := a.Init(db); err != nil {
		return nil, err
	}

	// Pre-compute schema information for field validation
	schema := db.TableOf((*TModel)(nil))

	// Pre-compute tree adapter for building hierarchical structure
	treeAdapter := treebuilder.Adapter[TreeOption]{
		GetId: func(t TreeOption) string {
			return t.Id
		},
		GetParentId: func(t TreeOption) string {
			return t.ParentId.ValueOr(constants.Empty)
		},
		SetChildren: func(t *TreeOption, children []TreeOption) {
			t.Children = children
		},
	}

	return func(ctx fiber.Ctx, db orm.Db, params TreeOptionParams, search TSearch) error {
		var flatOptions []TreeOption

		mergeTreeOptionColumnMapping(&params.TreeOptionColumnMapping, a.columnMapping)

		if err := validateTreeOptionColumns(schema, &params.TreeOptionColumnMapping); err != nil {
			return err
		}

		// Helper function to apply column selections with proper aliasing
		applyColumnSelections := func(selectQuery orm.SelectQuery) {
			if params.ValueColumn == valueColumn {
				selectQuery.Select(params.ValueColumn)
			} else {
				selectQuery.SelectAs(params.ValueColumn, valueColumn)
			}

			if params.LabelColumn == labelColumn {
				selectQuery.Select(params.LabelColumn)
			} else {
				selectQuery.SelectAs(params.LabelColumn, labelColumn)
			}

			if params.IdColumn == idColumn {
				selectQuery.Select(params.IdColumn)
			} else {
				selectQuery.SelectAs(params.IdColumn, idColumn)
			}

			if params.ParentIdColumn == parentIdColumn {
				selectQuery.Select(params.ParentIdColumn)
			} else {
				selectQuery.SelectAs(params.ParentIdColumn, parentIdColumn)
			}

			if params.DescriptionColumn != constants.Empty {
				if params.DescriptionColumn == descriptionColumn {
					selectQuery.Select(params.DescriptionColumn)
				} else {
					selectQuery.SelectAs(params.DescriptionColumn, descriptionColumn)
				}
			}

			if params.SortColumn != constants.Empty {
				selectQuery.Select(params.SortColumn)
			} else if a.ShouldApplyDefaultSort() {
				selectQuery.Select(constants.ColumnCreatedAt)
			}
		}

		cteQuery := db.NewSelect().
			WithRecursive("tmp_tree", func(selectQuery orm.SelectQuery) {
				// Base query
				a.ApplyDataPermission(selectQuery.Model((*TModel)(nil)), ctx)
				a.ApplyConditions(selectQuery, search, ctx)
				a.ApplyRelations(selectQuery, search, ctx)
				a.ApplyQuery(selectQuery, search, ctx)
				applyColumnSelections(selectQuery)

				// Recursive part: find all ancestor nodes
				selectQuery.UnionAll(func(selectQuery orm.SelectQuery) {
					a.ApplyRelations(selectQuery.Model((*TModel)(nil)), search, ctx)
					a.ApplyQuery(selectQuery, search, ctx)
					applyColumnSelections(selectQuery)
					selectQuery.JoinTable(
						"tmp_tree",
						func(cb orm.ConditionBuilder) {
							cb.EqualsColumn(params.IdColumn, dbhelpers.ColumnWithAlias(params.ParentIdColumn, "tt"))
						},
						"tt",
					)
				})
			}).
			Table("tmp_tree").
			Distinct()

		// Apply sorting
		if params.SortColumn != constants.Empty {
			cteQuery.OrderBy(params.SortColumn)
		} else {
			a.ApplySort(cteQuery, search, ctx)
			a.ApplyDefaultSort(cteQuery)
		}

		// Execute recursive CTE query
		if err := cteQuery.Limit(maxOptionsLimit).Scan(ctx.Context(), &flatOptions); err != nil {
			return err
		}

		// Build hierarchical tree structure from flat results using tree adapter
		treeOptions := treebuilder.Build(flatOptions, treeAdapter)

		return result.Ok(a.Process(treeOptions, search, ctx)).Response(ctx)
	}, nil
}
