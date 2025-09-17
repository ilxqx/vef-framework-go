package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/null"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/utils"
	"github.com/uptrace/bun/schema"
)

const (
	idField       = orm.ColumnId
	parentIdField = "parentId"
)

// TreeOptionsConfig defines the mapping between database fields and tree option fields.
type TreeOptionsConfig struct {
	api.Params
	OptionsConfig
	IdField       string `json:"idField"`       // Field name for ID (default: "id")
	ParentIdField string `json:"parentIdField"` // Field name for parent ID (default: "parentId")
}

// applyDefaults applies default values to tree options configuration.
func (c *TreeOptionsConfig) applyDefaults(defaultConfig *TreeOptionsConfig) {
	applyTreeOptionsDefaults(c, defaultConfig)
}

// validateFields validates that the specified fields exist in the model.
func (c *TreeOptionsConfig) validateFields(schema *schema.Table) error {
	return validateTreeOptionsFields(schema, c)
}

// TreeOption represents a hierarchical option with children.
type TreeOption struct {
	Option
	Id       string       `json:"id"`                 // Unique identifier
	ParentId null.String  `json:"parentId,omitzero"`  // Parent identifier
	Children []TreeOption `json:"children,omitempty"` // Child options
}

// FindTreeOptionsAPI provides hierarchical option query functionality.
type FindTreeOptionsAPI[TModel, TSearch any] struct {
	*findAPI[TModel, TSearch, PostFindProcessor[[]TreeOption, []TreeOption], FindTreeOptionsAPI[TModel, TSearch]]
	defaultConfig *TreeOptionsConfig
}

func (a *FindTreeOptionsAPI[TModel, TSearch]) WithDefaultConfig(config *TreeOptionsConfig) *FindTreeOptionsAPI[TModel, TSearch] {
	a.defaultConfig = config
	return a
}

// FindTreeOptions executes the query and returns hierarchical options with customizable configuration.
// Uses recursive CTE to efficiently fetch all nodes in the tree structure.
func (a *FindTreeOptionsAPI[TModel, TSearch]) FindTreeOptions(ctx fiber.Ctx, db orm.Db, config TreeOptionsConfig, search TSearch) error {
	var (
		flatOptions []TreeOption
		schema      = db.Schema((*TModel)(nil))
	)

	config.applyDefaults(a.defaultConfig)
	if err := config.validateFields(schema); err != nil {
		return err
	}

	// Helper function to apply field selections with proper aliasing
	applyFieldSelections := func(query orm.Query) orm.Query {
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

	query := db.NewQuery().
		Model((*TModel)(nil)).
		WithRecursive("tmp_tree", func(query orm.Query) {
			// Base query: fetch root nodes and apply filters
			query = a.configQuery(ctx, query, (*TModel)(nil), search)
			query = applyFieldSelections(query)

			// Apply sorting
			if config.SortField != constants.Empty {
				query.OrderBy(config.SortField)
			}

			if a.queryApplier == nil {
				if field, _ := schema.Field(orm.ColumnCreatedAt); field != nil {
					query.OrderBy(orm.ColumnCreatedAt)
				}
			}

			query = query.Limit(maxOptionsLimit)

			// Recursive part: fetch child nodes by joining with parent results
			query.Union(func(query orm.Query) {
				query = applyFieldSelections(query)

				// Apply sorting
				if config.SortField != constants.Empty {
					query.OrderBy(config.SortField)
				}
				if field, _ := schema.Field(orm.ColumnCreatedAt); field != nil {
					query.OrderBy(orm.ColumnCreatedAt)
				}

				query.JoinTableAs("tmp_tree", "t", func(cb orm.ConditionBuilder) {
					cb.EqualsExpr(config.IdField, "t.?", orm.Name(config.ParentIdField))
				})
			})
		}).
		Table("tmp_tree")

	// Execute recursive CTE query
	if err := query.Limit(maxOptionsLimit).Scan(ctx, &flatOptions); err != nil {
		return err
	}

	// Build hierarchical tree structure from flat results
	treeOptions := utils.BuildTree(flatOptions, utils.TreeAdapter[TreeOption]{
		GetId: func(t TreeOption) string {
			return t.Id
		},
		GetParentId: func(t TreeOption) string {
			return t.ParentId.ValueOr(constants.Empty)
		},
		GetChildren: func(t TreeOption) []TreeOption {
			return t.Children
		},
		SetChildren: func(t *TreeOption, children []TreeOption) {
			t.Children = children
		},
	})

	if a.processor != nil {
		treeOptions = a.processor(treeOptions, ctx)
	}

	return result.Ok(treeOptions).Response(ctx)
}

// NewFindTreeOptionsAPI creates a new FindTreeOptionsAPI with the specified options.
func NewFindTreeOptionsAPI[TModel, TSearch any]() *FindTreeOptionsAPI[TModel, TSearch] {
	api := new(FindTreeOptionsAPI[TModel, TSearch])
	api.findAPI = newFindAPI[TModel, TSearch, PostFindProcessor[[]TreeOption, []TreeOption]](api)

	return api
}
