package apis

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/search"
	"github.com/ilxqx/vef-framework-go/sort"
)

// QueryPart defines which part(s) of a query an option applies to.
// This is particularly important for tree APIs that use recursive CTEs,
// where different parts of the query require different configurations.
type QueryPart int

const (
	// QueryRoot applies the option to the root/outer query only.
	// This is the default for sorting and limiting operations.
	// For tree APIs, this is the final SELECT from the CTE.
	// For normal APIs, this is the main query.
	QueryRoot QueryPart = iota

	// QueryBase applies the option to the base/main query in recursive CTEs.
	// This is the initial SELECT inside the WITH clause before UNION ALL.
	// Use this for conditions that should only filter the starting nodes.
	QueryBase

	// QueryRecursive applies the option to the recursive query in CTEs.
	// This is the SELECT after UNION ALL that joins with the CTE itself.
	// Use this for configurations needed in the recursive traversal.
	QueryRecursive

	// QueryAll applies the option to all query parts.
	// Use this for configurations that should be consistent across all parts,
	// such as column selections or joins.
	QueryAll
)

// FindApiOption defines a configuration option for Find APIs.
// It can target specific query parts for fine-grained control,
// which is essential for recursive CTE queries used in tree operations.
//
// The Applier function receives the query, search parameters, and fiber context,
// allowing for dynamic configuration based on runtime data.
type FindApiOption struct {
	Parts   []QueryPart
	Applier func(query orm.SelectQuery, search any, ctx fiber.Ctx) error
}

// resolveQueryParts resolves the target query parts for an option.
// If no parts are specified, it returns QueryRoot as the default.
func resolveQueryParts(parts ...QueryPart) []QueryPart {
	if len(parts) == 0 {
		return []QueryPart{QueryRoot}
	}

	return parts
}

// withSelect adds a single column to the SELECT clause.
// Applies to root query by default (QueryRoot).
func withSelect(column string, parts ...QueryPart) *FindApiOption {
	return &FindApiOption{
		Parts: resolveQueryParts(parts...),
		Applier: func(query orm.SelectQuery, search any, ctx fiber.Ctx) error {
			query.Select(column)

			return nil
		},
	}
}

// withSelectAs adds a column with an alias to the SELECT clause.
// Applies to root query by default (QueryRoot).
func withSelectAs(column, alias string, parts ...QueryPart) *FindApiOption {
	return &FindApiOption{
		Parts: resolveQueryParts(parts...),
		Applier: func(query orm.SelectQuery, search any, ctx fiber.Ctx) error {
			query.SelectAs(column, alias)

			return nil
		},
	}
}

// withSort adds ordering based on sort.OrderSpec specifications.
// Applies to root query only by default (QueryRoot).
// Supports ascending/descending order and NULLS FIRST/LAST positioning.
func withSort(specs []*sort.OrderSpec, parts ...QueryPart) *FindApiOption {
	return &FindApiOption{
		Parts: resolveQueryParts(parts...),
		Applier: func(query orm.SelectQuery, search any, ctx fiber.Ctx) error {
			var sortable Sortable
			if err := contextx.ApiRequest(ctx).Meta.Decode(&sortable); err != nil {
				return err
			}

			applyOrderSpec := func(spec *sort.OrderSpec) {
				if !spec.IsValid() {
					return
				}

				query.OrderByExpr(func(eb orm.ExprBuilder) any {
					return eb.Order(func(ob orm.OrderBuilder) {
						ob.Column(spec.Column)

						switch spec.Direction {
						case sort.OrderAsc:
							ob.Asc()
						case sort.OrderDesc:
							ob.Desc()
						}

						switch spec.NullsOrder {
						case sort.NullsFirst:
							ob.NullsFirst()
						case sort.NullsLast:
							ob.NullsLast()
						}
					})
				})
			}

			if len(sortable.Sort) > 0 {
				for i := range sortable.Sort {
					applyOrderSpec(&sortable.Sort[i])
				}
			} else {
				for _, spec := range specs {
					applyOrderSpec(spec)
				}
			}

			return nil
		},
	}
}

// withCondition adds a WHERE condition using ConditionBuilder.
// Applies to root query only by default (QueryRoot).
// This is useful for adding simple filtering conditions.
func withCondition(fn func(cb orm.ConditionBuilder), parts ...QueryPart) *FindApiOption {
	return &FindApiOption{
		Parts: resolveQueryParts(parts...),
		Applier: func(query orm.SelectQuery, search any, ctx fiber.Ctx) error {
			query.Where(fn)

			return nil
		},
	}
}

// withSearchApplier creates a FindApiOption that automatically applies search conditions
// based on struct field tags (search:"eq", search:"contains", search:"gte", etc.).
func withSearchApplier[TSearch any](parts ...QueryPart) *FindApiOption {
	applier := search.Applier[TSearch]()

	return &FindApiOption{
		Parts: resolveQueryParts(parts...),
		Applier: func(query orm.SelectQuery, search any, ctx fiber.Ctx) error {
			s, ok := search.(TSearch)
			if !ok {
				var expectedType TSearch

				return fmt.Errorf(
					"%w: expected search type %T, but got %T. "+
						"Make sure the TSearch type parameter in WithSearch matches the search type used in NewFindXxxApi/NewExportApi",
					ErrSearchTypeMismatch, expectedType, search,
				)
			}

			query.Where(func(cb orm.ConditionBuilder) {
				cb.Apply(applier(s))
			})

			return nil
		},
	}
}

// withQueryApplier adds custom query modifications with type-safe access to search parameters.
// This is more powerful than WithCondition as it allows full query manipulation.
// Applies to root query only by default (QueryRoot).
func withQueryApplier[TSearch any](applier func(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) error, parts ...QueryPart) *FindApiOption {
	return &FindApiOption{
		Parts: resolveQueryParts(parts...),
		Applier: func(query orm.SelectQuery, search any, ctx fiber.Ctx) error {
			s, ok := search.(TSearch)
			if !ok {
				var expectedType TSearch

				return fmt.Errorf(
					"%w: expected search type %T, but got %T. "+
						"Make sure the TSearch type parameter in WithQueryApplier matches the search type used in NewFindXxxApi/NewExportApi",
					ErrSearchTypeMismatch, expectedType, search,
				)
			}

			return applier(query, s, ctx)
		},
	}
}

// withDataPerm creates a FindApiOption that applies data permission filtering from the request context.
// The data permission applier is retrieved using contextx.DataPermApplier(ctx).
func withDataPerm(parts ...QueryPart) *FindApiOption {
	return &FindApiOption{
		Parts: resolveQueryParts(parts...),
		Applier: func(query orm.SelectQuery, search any, ctx fiber.Ctx) error {
			return ApplyDataPermission(query, ctx)
		},
	}
}

// withRelation adds a single relation join to the query.
// Applies to root query by default (QueryRoot).
func withRelation(relation *orm.RelationSpec, parts ...QueryPart) *FindApiOption {
	return &FindApiOption{
		Parts: resolveQueryParts(parts...),
		Applier: func(query orm.SelectQuery, search any, ctx fiber.Ctx) error {
			query.JoinRelations(relation)

			return nil
		},
	}
}

// withAuditUserNames adds joins to fetch audit user names (created_by_name, updated_by_name).
// Allows specifying a custom column name for the user's display name.
func withAuditUserNames(userModel any, nameColumn string, parts ...QueryPart) *FindApiOption {
	// Create RelationSpecs for creator and updater
	relations := GetAuditUserNameRelations(userModel, nameColumn)

	return &FindApiOption{
		Parts: resolveQueryParts(parts...),
		Applier: func(query orm.SelectQuery, search any, ctx fiber.Ctx) error {
			query.JoinRelations(relations...)

			return nil
		},
	}
}
