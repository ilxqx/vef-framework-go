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

// OptionFieldMapping defines the mapping between database fields and option fields.
type OptionFieldMapping struct {
	// Field name for label (default: "name")
	LabelField string `json:"labelField"`
	// Field name for value (default: "id")
	ValueField string `json:"valueField"`
	// Field name for description
	DescriptionField string `json:"descriptionField"`
	// Field name for sorting
	SortField string `json:"sortField"`
}

// OptionParams is the API request parameter for querying options.
type OptionParams struct {
	api.In
	OptionFieldMapping
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

// TreeOptionFieldMapping defines the mapping between database fields and tree option fields.
type TreeOptionFieldMapping struct {
	OptionFieldMapping

	// Field name for ID (default: "id")
	IdField string `json:"idField"`
	// Field name for parent ID (default: "parentId")
	ParentIdField string `json:"parentIdField"`
}

// TreeOptionParams is the API request parameter for querying tree options.
type TreeOptionParams struct {
	api.In
	TreeOptionFieldMapping
}
