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
	FindApi[TModel, TSearch, []TreeDataOption, FindTreeOptionsApi[TModel, TSearch]]

	defaultColumnMapping *DataOptionColumnMapping
	idColumn             string
	parentIdColumn       string
}

func (a *findTreeOptionsApi[TModel, TSearch]) Provide() api.Spec {
	return a.Build(a.findTreeOptions)
}

// WithDefaultColumnMapping sets the default column mapping for data option fields.
// This mapping provides fallback values for label, value, description, and sort columns.
func (a *findTreeOptionsApi[TModel, TSearch]) WithDefaultColumnMapping(mapping *DataOptionColumnMapping) FindTreeOptionsApi[TModel, TSearch] {
	a.defaultColumnMapping = mapping

	return a
}

// WithIdColumn sets the column name used as the node ID in tree structures.
// This column is used to identify individual nodes and establish parent-child relationships.
func (a *findTreeOptionsApi[TModel, TSearch]) WithIdColumn(name string) FindTreeOptionsApi[TModel, TSearch] {
	a.idColumn = name

	return a
}

// WithParentIdColumn sets the column name used to reference parent nodes in tree structures.
// This column establishes the hierarchical relationship between parent and child nodes.
func (a *findTreeOptionsApi[TModel, TSearch]) WithParentIdColumn(name string) FindTreeOptionsApi[TModel, TSearch] {
	a.parentIdColumn = name

	return a
}

// WithSelect adds a column to the SELECT clause.
// Applies to all query parts by default (QueryAll) unless specific parts are provided.
func (a *findTreeOptionsApi[TModel, TSearch]) WithSelect(column string, parts ...QueryPart) FindTreeOptionsApi[TModel, TSearch] {
	a.FindApi.WithSelect(column, lo.Ternary(len(parts) > 0, parts, []QueryPart{QueryBase, QueryRecursive})...)

	return a
}

// WithSelectAs adds a column with an alias to the SELECT clause.
// Applies to all query parts by default (QueryAll) unless specific parts are provided.
func (a *findTreeOptionsApi[TModel, TSearch]) WithSelectAs(column, alias string, parts ...QueryPart) FindTreeOptionsApi[TModel, TSearch] {
	a.FindApi.WithSelectAs(column, alias, lo.Ternary(len(parts) > 0, parts, []QueryPart{QueryBase, QueryRecursive})...)

	return a
}

// WithCondition adds a WHERE condition using ConditionBuilder.
// Applies to root query only by default (QueryRoot) unless specific parts are provided.
func (a *findTreeOptionsApi[TModel, TSearch]) WithCondition(fn func(cb orm.ConditionBuilder), parts ...QueryPart) FindTreeOptionsApi[TModel, TSearch] {
	a.FindApi.WithCondition(fn, lo.Ternary(len(parts) > 0, parts, []QueryPart{QueryBase})...)

	return a
}

// WithRelation adds a relation join to the query.
// Applies to all query parts by default (QueryAll) unless specific parts are provided.
func (a *findTreeOptionsApi[TModel, TSearch]) WithRelation(relation *orm.RelationSpec, parts ...QueryPart) FindTreeOptionsApi[TModel, TSearch] {
	a.FindApi.WithRelation(relation, lo.Ternary(len(parts) > 0, parts, []QueryPart{QueryBase, QueryRecursive})...)

	return a
}

// WithQueryApplier adds a custom query applier function.
// Applies to root query only by default (QueryRoot) unless specific parts are provided.
func (a *findTreeOptionsApi[TModel, TSearch]) WithQueryApplier(applier func(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) error, parts ...QueryPart) FindTreeOptionsApi[TModel, TSearch] {
	a.FindApi.WithQueryApplier(applier, lo.Ternary(len(parts) > 0, parts, []QueryPart{QueryBase})...)

	return a
}

func (a *findTreeOptionsApi[TModel, TSearch]) findTreeOptions(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, config DataOptionConfig, sortable Sortable, search TSearch) error, error) {
	if err := a.Setup(db, &FindApiConfig{
		QueryParts: &QueryPartsConfig{
			Condition:         []QueryPart{QueryBase},
			Sort:              []QueryPart{QueryRoot},
			AuditUserRelation: []QueryPart{QueryBase, QueryRecursive},
		},
	}); err != nil {
		return nil, err
	}

	// HACK: Direct access to baseFindApi's defaultSort field via type assertion.
	// This is needed because the recursive CTE query must SELECT all columns that will be used
	// in the outer query's ORDER BY clause, otherwise the database will report "column not found" errors.
	// The proper solution would be to add a method to the FindApi interface to expose defaultSort,
	// but this would require interface changes for a relatively rare use case.
	// TODO: Consider a more appropriate approach if this pattern becomes more common.
	defaultSort := a.FindApi.(*baseFindApi[TModel, TSearch, []TreeDataOption, FindTreeOptionsApi[TModel, TSearch]]).defaultSort

	table := db.TableOf((*TModel)(nil))
	treeAdapter := treebuilder.Adapter[TreeDataOption]{
		GetId:       func(t TreeDataOption) string { return t.Id },
		GetParentId: func(t TreeDataOption) string { return t.ParentId.ValueOrZero() },
		SetChildren: func(t *TreeDataOption, children []TreeDataOption) { t.Children = children },
	}

	if !table.HasField(a.idColumn) {
		return nil, fmt.Errorf("%w: column %q does not exist in model %T (tree node id)", ErrColumnNotFound, a.idColumn, (*TModel)(nil))
	}

	if !table.HasField(a.parentIdColumn) {
		return nil, fmt.Errorf("%w: column %q does not exist in model %T (parent reference)", ErrColumnNotFound, a.parentIdColumn, (*TModel)(nil))
	}

	return func(ctx fiber.Ctx, db orm.Db, config DataOptionConfig, sortable Sortable, search TSearch) error {
		var flatOptions []TreeDataOption

		// Merge column mapping from params with defaults
		mergeOptionColumnMapping(&config.DataOptionColumnMapping, a.defaultColumnMapping)

		if err := validateOptionColumns(table, &config.DataOptionColumnMapping); err != nil {
			return err
		}

		// Helper function to apply column selections with proper aliasing
		applyColumnSelections := func(query orm.SelectQuery) {
			if config.ValueColumn == valueColumn {
				query.Select(config.ValueColumn)
			} else {
				query.SelectAs(config.ValueColumn, valueColumn)
			}

			if config.LabelColumn == labelColumn {
				query.Select(config.LabelColumn)
			} else {
				query.SelectAs(config.LabelColumn, labelColumn)
			}

			if config.DescriptionColumn != constants.Empty {
				if config.DescriptionColumn == descriptionColumn {
					query.Select(config.DescriptionColumn)
				} else {
					query.SelectAs(config.DescriptionColumn, descriptionColumn)
				}
			}

			if len(sortable.Sort) > 0 {
				for i := range sortable.Sort {
					query.Select((&sortable.Sort[i]).Column)
				}
			} else if len(defaultSort) > 0 {
				for _, spec := range defaultSort {
					query.Select(spec.Column)
				}
			}

			if a.idColumn == idColumn {
				query.Select(a.idColumn)
			} else {
				query.SelectAs(a.idColumn, idColumn)
			}

			if a.parentIdColumn == parentIdColumn {
				query.Select(a.parentIdColumn)
			} else {
				query.SelectAs(a.parentIdColumn, parentIdColumn)
			}
		}

		query := db.NewSelect().
			WithRecursive("tmp_tree", func(cteQuery orm.SelectQuery) {
				// Base query - the starting point of the tree traversal
				baseQuery := cteQuery.Model((*TModel)(nil))

				applyColumnSelections(baseQuery)

				if err := a.ConfigureQuery(baseQuery, search, ctx, QueryBase); err != nil {
					SetQueryError(ctx, err)

					return
				}

				// Recursive part: find all ancestor nodes
				cteQuery.UnionAll(func(recursiveQuery orm.SelectQuery) {
					recursiveQuery.Model((*TModel)(nil))

					applyColumnSelections(recursiveQuery)

					if err := a.ConfigureQuery(recursiveQuery, search, ctx, QueryRecursive); err != nil {
						SetQueryError(ctx, err)

						return
					}

					// Join with CTE to traverse the tree
					recursiveQuery.JoinTable(
						"tmp_tree",
						func(cb orm.ConditionBuilder) {
							cb.EqualsColumn(a.idColumn, dbhelpers.ColumnWithAlias(a.parentIdColumn, "tt"))
						},
						"tt",
					)
				})
			}).
			Table("tmp_tree").
			Distinct()

		// Check for errors during query building
		if queryErr := QueryError(ctx); queryErr != nil {
			return queryErr
		}

		// Apply QueryRoot and QueryAll options to the outer query
		if err := a.ConfigureQuery(query, search, ctx, QueryRoot); err != nil {
			return err
		}

		// Execute recursive CTE query
		if err := query.Limit(maxOptionsLimit).
			Scan(ctx.Context(), &flatOptions); err != nil {
			return err
		}

		treeOptions := treebuilder.Build(flatOptions, treeAdapter)

		return result.Ok(a.Process(treeOptions, search, ctx)).Response(ctx)
	}, nil
}
