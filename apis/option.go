package apis

import (
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/null"
)

// Option represents a selectable item with display text and underlying value.
// Commonly used for dropdown lists, radio buttons, and select components.
type Option struct {
	// Display text shown to users
	Label string `json:"label"`
	// Underlying value used in form submissions
	Value string `json:"value"`
	// Additional descriptive text (optional)
	Description string `json:"description,omitempty"`
	// Custom metadata for extended functionality (optional)
	Meta map[string]any `json:"meta,omitempty"`
}

// OptionColumnMapping defines the mapping between database columns and option fields.
type OptionColumnMapping struct {
	// Column name for label (default: "name")
	LabelColumn string `json:"labelColumn"`
	// Column name for value (default: "id")
	ValueColumn string `json:"valueColumn"`
	// Column name for description
	DescriptionColumn string `json:"descriptionColumn"`
	// Column name for sorting
	SortColumn string `json:"sortColumn"`
}

// OptionParams is the API request parameter for querying options.
type OptionParams struct {
	api.In
	OptionColumnMapping
}

// TreeOption represents a hierarchical selectable item that can contain child options.
// Used for tree-structured selections like category menus or organizational hierarchies.
type TreeOption struct {
	Option

	// Unique identifier for the tree node
	Id string `json:"id"`
	// Parent node identifier (null for root nodes)
	ParentId null.String `json:"parentId"`
	// Nested child options forming the tree structure
	Children []TreeOption `json:"children,omitempty"`
}

// TreeOptionColumnMapping defines the mapping between database columns and tree option fields.
type TreeOptionColumnMapping struct {
	OptionColumnMapping

	// Column name for ID (default: "id")
	IdColumn string `json:"idColumn"`
	// Column name for parent ID (default: "parentId")
	ParentIdColumn string `json:"parentIdColumn"`
}

// TreeOptionParams is the API request parameter for querying tree options.
type TreeOptionParams struct {
	api.In
	TreeOptionColumnMapping
}
