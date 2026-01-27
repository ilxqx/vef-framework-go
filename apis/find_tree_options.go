package apis

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/dbhelpers"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/treebuilder"
)

type findTreeOptionsApi[TModel, TSearch any] struct {
	Find[TModel, TSearch, []TreeDataOption, FindTreeOptions[TModel, TSearch]]

	defaultColumnMapping *DataOptionColumnMapping
	idColumn             string
	parentIDColumn       string
}

func (a *findTreeOptionsApi[TModel, TSearch]) Provide() []api.OperationSpec {
	return []api.OperationSpec{a.Build(a.findTreeOptions)}
}

// This mapping provides fallback values for label, value, description, and sort columns.
func (a *findTreeOptionsApi[TModel, TSearch]) WithDefaultColumnMapping(mapping *DataOptionColumnMapping) FindTreeOptions[TModel, TSearch] {
	a.defaultColumnMapping = mapping

	return a
}

// This column is used to identify individual nodes and establish parent-child relationships.
func (a *findTreeOptionsApi[TModel, TSearch]) WithIDColumn(name string) FindTreeOptions[TModel, TSearch] {
	a.idColumn = name

	return a
}

// This column establishes the hierarchical relationship between parent and child nodes.
func (a *findTreeOptionsApi[TModel, TSearch]) WithParentIDColumn(name string) FindTreeOptions[TModel, TSearch] {
	a.parentIDColumn = name

	return a
}

// WithCondition adds a WHERE condition using ConditionBuilder.
// Defaults to QueryBase for tree options unless specific parts are provided.
func (a *findTreeOptionsApi[TModel, TSearch]) WithCondition(fn func(cb orm.ConditionBuilder), parts ...QueryPart) FindTreeOptions[TModel, TSearch] {
	a.Find.WithCondition(fn, lo.Ternary(len(parts) > 0, parts, []QueryPart{QueryBase})...)

	return a
}

// WithQueryApplier adds a custom query applier function.
// Defaults to QueryBase for tree options unless specific parts are provided.
func (a *findTreeOptionsApi[TModel, TSearch]) WithQueryApplier(applier func(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) error, parts ...QueryPart) FindTreeOptions[TModel, TSearch] {
	a.Find.WithQueryApplier(applier, lo.Ternary(len(parts) > 0, parts, []QueryPart{QueryBase})...)

	return a
}

func (a *findTreeOptionsApi[TModel, TSearch]) findTreeOptions(db orm.DB) (func(ctx fiber.Ctx, db orm.DB, config DataOptionConfig, _ Sortable, search TSearch, meta api.Meta) error, error) {
	if err := a.Setup(db, &FindApiConfig{
		QueryParts: &QueryPartsConfig{
			Condition:         []QueryPart{QueryBase},
			Sort:              []QueryPart{QueryRoot},
			AuditUserRelation: []QueryPart{QueryBase, QueryRecursive},
		},
	}); err != nil {
		return nil, err
	}

	table := db.TableOf((*TModel)(nil))
	treeAdapter := treebuilder.Adapter[TreeDataOption]{
		GetID:       func(t TreeDataOption) string { return t.ID },
		GetParentID: func(t TreeDataOption) string { return t.ParentID.ValueOrZero() },
		SetChildren: func(t *TreeDataOption, children []TreeDataOption) { t.Children = children },
	}

	if !table.HasField(a.idColumn) {
		return nil, fmt.Errorf("%w: column %q does not exist in model %T (tree node id)", ErrColumnNotFound, a.idColumn, (*TModel)(nil))
	}

	if !table.HasField(a.parentIDColumn) {
		return nil, fmt.Errorf("%w: column %q does not exist in model %T (parent reference)", ErrColumnNotFound, a.parentIDColumn, (*TModel)(nil))
	}

	return func(ctx fiber.Ctx, db orm.DB, config DataOptionConfig, _ Sortable, search TSearch, meta api.Meta) error {
		var (
			flatOptions []TreeDataOption
			query       = db.NewSelect().Model((*TModel)(nil))
		)

		mergeOptionColumnMapping(&config.DataOptionColumnMapping, a.defaultColumnMapping)

		if err := validateOptionColumns(table, &config.DataOptionColumnMapping); err != nil {
			return err
		}

		metaColumns := parseMetaColumns(config.MetaColumns)
		if err := validateMetaColumns(table, metaColumns); err != nil {
			return err
		}

		// Helper function to apply column selections with proper aliasing
		applyColumnSelections := func(query orm.SelectQuery) {
			if a.idColumn == IDColumn {
				query.Select(a.idColumn)
			} else {
				query.SelectAs(a.idColumn, IDColumn)
			}

			if a.parentIDColumn == ParentIDColumn {
				query.Select(a.parentIDColumn)
			} else {
				query.SelectAs(a.parentIDColumn, ParentIDColumn)
			}
		}

		query.WithRecursive(
			"_tree", func(cteQuery orm.SelectQuery) {
				applyColumnSelections(cteQuery.Model((*TModel)(nil)))

				if err := a.ConfigureQuery(cteQuery, search, meta, ctx, QueryBase); err != nil {
					SetQueryError(ctx, err)

					return
				}

				// Recursive part: find all ancestor nodes
				cteQuery.UnionAll(func(recursiveQuery orm.SelectQuery) {
					applyColumnSelections(recursiveQuery.Model((*TModel)(nil)))

					if err := a.ConfigureQuery(recursiveQuery, search, meta, ctx, QueryRecursive); err != nil {
						SetQueryError(ctx, err)

						return
					}

					// Join with CTE to traverse the tree
					recursiveQuery.JoinTable(
						"_tree",
						func(cb orm.ConditionBuilder) {
							cb.EqualsColumn(a.idColumn, dbhelpers.ColumnWithAlias(a.parentIDColumn, "_tree"))
						},
					)
				})
			}).
			With("_ids", func(query orm.SelectQuery) {
				query.Table("_tree").
					Select(IDColumn).
					Distinct()
			})

		if queryErr := QueryError(ctx); queryErr != nil {
			return queryErr
		}

		applyColumnSelections(query)

		if config.LabelColumn == LabelColumn {
			query.Select(config.LabelColumn)
		} else {
			query.SelectAs(config.LabelColumn, LabelColumn)
		}

		if config.ValueColumn == ValueColumn {
			query.Select(config.ValueColumn)
		} else {
			query.SelectAs(config.ValueColumn, ValueColumn)
		}

		if config.DescriptionColumn != constants.Empty {
			if config.DescriptionColumn == DescriptionColumn {
				query.Select(config.DescriptionColumn)
			} else {
				query.SelectAs(config.DescriptionColumn, DescriptionColumn)
			}
		}

		query.ApplyIf(len(metaColumns) > 0, func(sq orm.SelectQuery) {
			sq.SelectExpr(
				func(eb orm.ExprBuilder) any {
					return buildMetaJSONExpr(eb, metaColumns)
				},
				"meta",
			)
		})

		query.Where(func(cb orm.ConditionBuilder) {
			cb.InSubQuery(a.idColumn, func(query orm.SelectQuery) {
				query.Table("_ids")
			})
		})

		if err := a.ConfigureQuery(query, search, meta, ctx, QueryRoot); err != nil {
			return err
		}

		if err := query.Limit(maxOptionsLimit).
			Scan(ctx.Context(), &flatOptions); err != nil {
			return err
		}

		treeOptions := treebuilder.Build(flatOptions, treeAdapter)

		return result.Ok(a.Process(treeOptions, search, ctx)).Response(ctx)
	}, nil
}
