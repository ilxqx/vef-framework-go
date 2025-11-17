package apis

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"

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

// This column is used to identify individual nodes and establish parent-child relationships.
func (a *findTreeApi[TModel, TSearch]) WithIdColumn(name string) FindTreeApi[TModel, TSearch] {
	a.idColumn = name

	return a
}

// This column establishes the hierarchical relationship between parent and child nodes.
func (a *findTreeApi[TModel, TSearch]) WithParentIdColumn(name string) FindTreeApi[TModel, TSearch] {
	a.parentIdColumn = name

	return a
}

// WithSelect adds a column to the SELECT clause.
// Defaults to QueryBase and QueryRecursive for tree queries unless specific parts are provided.
func (a *findTreeApi[TModel, TSearch]) WithSelect(column string, parts ...QueryPart) FindTreeApi[TModel, TSearch] {
	a.FindApi.WithSelect(column, lo.Ternary(len(parts) > 0, parts, []QueryPart{QueryBase, QueryRecursive})...)

	return a
}

// WithSelectAs adds a column with an alias to the SELECT clause.
// Defaults to QueryBase and QueryRecursive for tree queries unless specific parts are provided.
func (a *findTreeApi[TModel, TSearch]) WithSelectAs(column, alias string, parts ...QueryPart) FindTreeApi[TModel, TSearch] {
	a.FindApi.WithSelectAs(column, alias, lo.Ternary(len(parts) > 0, parts, []QueryPart{QueryBase, QueryRecursive})...)

	return a
}

// WithCondition adds a WHERE condition using ConditionBuilder.
// Defaults to QueryBase (filter starting nodes) unless specific parts are provided.
func (a *findTreeApi[TModel, TSearch]) WithCondition(fn func(cb orm.ConditionBuilder), parts ...QueryPart) FindTreeApi[TModel, TSearch] {
	a.FindApi.WithCondition(fn, lo.Ternary(len(parts) > 0, parts, []QueryPart{QueryBase})...)

	return a
}

// WithRelation adds a relation join to the query.
// Defaults to QueryBase and QueryRecursive for tree queries unless specific parts are provided.
func (a *findTreeApi[TModel, TSearch]) WithRelation(relation *orm.RelationSpec, parts ...QueryPart) FindTreeApi[TModel, TSearch] {
	a.FindApi.WithRelation(relation, lo.Ternary(len(parts) > 0, parts, []QueryPart{QueryBase, QueryRecursive})...)

	return a
}

// WithQueryApplier adds a custom query applier function.
// Defaults to QueryBase (apply during base CTE selection) unless specific parts are provided.
func (a *findTreeApi[TModel, TSearch]) WithQueryApplier(applier func(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) error, parts ...QueryPart) FindTreeApi[TModel, TSearch] {
	a.FindApi.WithQueryApplier(applier, lo.Ternary(len(parts) > 0, parts, []QueryPart{QueryBase})...)

	return a
}

func (a *findTreeApi[TModel, TSearch]) findTree(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error, error) {
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
	if !table.HasField(a.idColumn) {
		return nil, fmt.Errorf("%w: column %q does not exist in model %T (tree node id)", ErrColumnNotFound, a.idColumn, (*TModel)(nil))
	}

	if !table.HasField(a.parentIdColumn) {
		return nil, fmt.Errorf("%w: column %q does not exist in model %T (parent reference)", ErrColumnNotFound, a.parentIdColumn, (*TModel)(nil))
	}

	return func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error {
		var (
			flatModels []TModel
			query      = db.NewSelect()
		)

		query.WithRecursive(
			"tmp_tree", func(cteQuery orm.SelectQuery) {
				// Base query - the starting point of the tree traversal
				baseQuery := cteQuery.Model((*TModel)(nil)).SelectModelColumns()

				if err := a.ConfigureQuery(baseQuery, search, ctx, QueryBase); err != nil {
					// Store error for later return
					SetQueryError(ctx, err)

					return
				}

				// Recursive part: find all ancestor/descendant nodes
				cteQuery.UnionAll(func(recursiveQuery orm.SelectQuery) {
					recursiveQuery.Model((*TModel)(nil)).SelectModelColumns()

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
			Distinct().
			Table("tmp_tree")

		if queryErr := QueryError(ctx); queryErr != nil {
			return queryErr
		}

		if err := a.ConfigureQuery(query, search, ctx, QueryRoot); err != nil {
			return err
		}

		if err := query.Limit(maxQueryLimit).
			Scan(ctx.Context(), &flatModels); err != nil {
			return err
		}

		if len(flatModels) > 0 {
			for i := range flatModels {
				if err := transformer.Struct(ctx.Context(), &flatModels[i]); err != nil {
					return err
				}
			}
		} else {
			flatModels = make([]TModel, 0)
		}

		models := a.treeBuilder(flatModels)

		return result.Ok(a.Process(models, search, ctx)).Response(ctx)
	}, nil
}
