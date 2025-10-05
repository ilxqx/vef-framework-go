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

type findTreeOptionsAPI[TModel, TSearch any] struct {
	FindAPI[TModel, TSearch, []TreeOption, FindTreeOptionsAPI[TModel, TSearch]]

	fieldMapping *TreeOptionFieldMapping
}

func (a *findTreeOptionsAPI[TModel, TSearch]) Provide() api.Spec {
	return a.FindAPI.Build(a.findTreeOptions)
}

// Build should not be called directly on concrete API types.
// Use Provide() to generate api.Spec with the correct handler instead.
func (a *findTreeOptionsAPI[TModel, TSearch]) Build(handler any) api.Spec {
	panic("apis: do not call FindAPI.Build on findTreeOptionsAPI; call Provide() instead")
}

func (a *findTreeOptionsAPI[TModel, TSearch]) FieldMapping(mapping *TreeOptionFieldMapping) FindTreeOptionsAPI[TModel, TSearch] {
	a.fieldMapping = mapping

	return a
}

func (a *findTreeOptionsAPI[TModel, TSearch]) findTreeOptions(db orm.Db) func(ctx fiber.Ctx, db orm.Db, params TreeOptionParams, search TSearch) error {
	// Pre-compute schema information
	schema := db.TableOf((*TModel)(nil))

	// Pre-compute whether default created_at ordering should be applied
	hasCreatedAt := schema.HasField(constants.ColumnCreatedAt)
	shouldApplyDefaultSort := !a.HasSortApplier() && hasCreatedAt

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

		mergeTreeOptionFieldMapping(&params.TreeOptionFieldMapping, a.fieldMapping)

		if err := validateTreeOptionFields(schema, &params.TreeOptionFieldMapping); err != nil {
			return err
		}

		// Helper function to apply field selections with proper aliasing
		applyFieldSelections := func(selectQuery orm.SelectQuery) {
			if params.ValueField == valueField {
				selectQuery.Select(params.ValueField)
			} else {
				selectQuery.SelectAs(params.ValueField, valueField)
			}

			if params.LabelField == labelField {
				selectQuery.Select(params.LabelField)
			} else {
				selectQuery.SelectAs(params.LabelField, labelField)
			}

			if params.IdField == idField {
				selectQuery.Select(params.IdField)
			} else {
				selectQuery.SelectAs(params.IdField, idField)
			}

			if params.ParentIdField == parentIdField {
				selectQuery.Select(params.ParentIdField)
			} else {
				selectQuery.SelectAs(params.ParentIdField, parentIdField)
			}

			if params.DescriptionField != constants.Empty {
				if params.DescriptionField == descriptionField {
					selectQuery.Select(params.DescriptionField)
				} else {
					selectQuery.SelectAs(params.DescriptionField, descriptionField)
				}
			}

			if params.SortField != constants.Empty {
				selectQuery.Select(params.SortField)
			} else if shouldApplyDefaultSort {
				selectQuery.Select(constants.ColumnCreatedAt)
			}
		}

		cteQuery := db.NewSelect().
			WithRecursive("tmp_tree", func(selectQuery orm.SelectQuery) {
				// Base query
				a.ApplyConditions(selectQuery.Model((*TModel)(nil)), search, ctx)
				a.ApplyRelations(selectQuery, search, ctx)
				a.ApplyQuery(selectQuery, search, ctx)
				applyFieldSelections(selectQuery)

				// Recursive part: find all ancestor nodes
				selectQuery.UnionAll(func(selectQuery orm.SelectQuery) {
					a.ApplyRelations(selectQuery, search, ctx)
					a.ApplyQuery(selectQuery, search, ctx)
					applyFieldSelections(selectQuery.Model((*TModel)(nil)))
					selectQuery.JoinTable(
						"tmp_tree",
						func(cb orm.ConditionBuilder) {
							cb.EqualsColumn(params.IdField, dbhelpers.ColumnWithAlias(params.ParentIdField, "tt"))
						},
						"tt",
					)
				})
			}).
			Table("tmp_tree").
			Distinct()

		// Apply sorting
		if params.SortField != constants.Empty {
			cteQuery.OrderBy(params.SortField)
		} else {
			a.ApplySort(cteQuery, search, ctx)

			if shouldApplyDefaultSort {
				cteQuery.OrderBy(constants.ColumnCreatedAt)
			}
		}

		// Execute recursive CTE query
		if err := cteQuery.Limit(maxOptionsLimit).Scan(ctx.Context(), &flatOptions); err != nil {
			return err
		}

		// Build hierarchical tree structure from flat results using tree adapter
		treeOptions := treebuilder.Build(flatOptions, treeAdapter)

		return result.Ok(a.Process(treeOptions, search, ctx)).Response(ctx)
	}
}
