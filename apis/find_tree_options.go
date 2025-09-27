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

	defaultConfig *TreeOptionsConfig
}

func (a *findTreeOptionsAPI[TModel, TSearch]) Provide() api.Spec {
	return a.FindAPI.Build(a.findTreeOptions)
}

// Build should not be called directly on concrete API types.
// Use Provide() to generate api.Spec with the correct handler instead.
func (a *findTreeOptionsAPI[TModel, TSearch]) Build(handler any) api.Spec {
	panic("apis: do not call FindAPI.Build on findTreeOptionsAPI; call Provide() instead")
}

func (a *findTreeOptionsAPI[TModel, TSearch]) DefaultConfig(config *TreeOptionsConfig) FindTreeOptionsAPI[TModel, TSearch] {
	a.defaultConfig = config
	return a
}

func (a *findTreeOptionsAPI[TModel, TSearch]) findTreeOptions(db orm.Db) func(ctx fiber.Ctx, db orm.Db, config TreeOptionsConfig, search TSearch) error {
	// Pre-compute schema information
	schema := db.Schema((*TModel)(nil))

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

	return func(ctx fiber.Ctx, db orm.Db, config TreeOptionsConfig, search TSearch) error {
		var flatOptions []TreeOption

		config.applyDefaults(a.defaultConfig)
		if err := config.validateFields(schema); err != nil {
			return err
		}

		// Helper function to apply field selections with proper aliasing
		applyFieldSelections := func(query orm.SelectQuery) orm.SelectQuery {
			if config.ValueField == valueField {
				query.Select(config.ValueField)
			} else {
				query.SelectAs(config.ValueField, valueField)
			}

			if config.LabelField == labelField {
				query.Select(config.LabelField)
			} else {
				query.SelectAs(config.LabelField, labelField)
			}

			if config.IdField == idField {
				query.Select(config.IdField)
			} else {
				query.SelectAs(config.IdField, idField)
			}

			if config.ParentIdField == parentIdField {
				query.Select(config.ParentIdField)
			} else {
				query.SelectAs(config.ParentIdField, parentIdField)
			}

			if config.DescriptionField != constants.Empty {
				if config.DescriptionField == descriptionField {
					query.Select(config.DescriptionField)
				} else {
					query.SelectAs(config.DescriptionField, descriptionField)
				}
			}

			return query
		}

		query := db.NewSelect().
			WithRecursive("tmp_tree", func(query orm.SelectQuery) {
				// Base query: fetch root nodes and apply filters
				a.ConfigureQuery(query, (*TModel)(nil), search, ctx)
				query = applyFieldSelections(query)

				// Apply sorting
				if config.SortField != constants.Empty {
					query.OrderBy(config.SortField)
				} else if shouldApplyDefaultSort {
					query.OrderBy(constants.ColumnCreatedAt)
				}

				query = query.Limit(maxOptionsLimit)

				// Recursive part: fetch child nodes by joining with parent results
				query.Union(func(query orm.SelectQuery) {
					query = applyFieldSelections(query.Model((*TModel)(nil)))

					// Apply sorting
					if config.SortField != constants.Empty {
						query.OrderBy(config.SortField)
					} else if shouldApplyDefaultSort {
						query.OrderBy(constants.ColumnCreatedAt)
					} else if a.HasSortApplier() {
						a.ApplySort(query, search, ctx)
					}

					query.JoinTable("tmp_tree", func(cb orm.ConditionBuilder) {
						cb.EqualsColumn(config.IdField, dbhelpers.ColumnWithAlias(config.ParentIdField, "tt"))
					}, "tt")
				})
			}).
			Table("tmp_tree")

		// Execute recursive CTE query
		if err := query.Scan(ctx, &flatOptions); err != nil {
			return err
		}

		// Build hierarchical tree structure from flat results using tree adapter
		treeOptions := treebuilder.Build(flatOptions, treeAdapter)
		return result.Ok(a.Process(treeOptions, search, ctx)).Response(ctx)
	}
}
