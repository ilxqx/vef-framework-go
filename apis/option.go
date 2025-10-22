package apis

import (
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/null"
)

// DataOption represents a selectable item with display text and underlying value.
// Commonly used for dropdown lists, radio buttons, and select components.
type DataOption struct {
	// Display text shown to users
	Label string `json:"label"`
	// Underlying value used in form submissions
	Value string `json:"value"`
	// Additional descriptive text (optional)
	Description string `json:"description,omitempty"`
	// Custom metadata for extended functionality (optional)
	Meta map[string]any `json:"meta,omitempty"`
}

// DataOptionColumnMapping defines the mapping between database columns and option fields.
type DataOptionColumnMapping struct {
	// Column name for label (default: "name")
	LabelColumn string `json:"labelColumn"`
	// Column name for value (default: "id")
	ValueColumn string `json:"valueColumn"`
	// Column name for description
	DescriptionColumn string `json:"descriptionColumn"`
}

// DataOptionConfig is the Api request meta for querying options.
type DataOptionConfig struct {
	api.M

	DataOptionColumnMapping
}

// TreeDataOption represents a hierarchical selectable item that can contain child options.
// Used for tree-structured selections like category menus or organizational hierarchies.
type TreeDataOption struct {
	DataOption

	// Unique identifier for the tree node
	Id string `json:"-"`
	// Parent node identifier (null for root nodes)
	ParentId null.String `json:"-"`
	// Nested child options forming the tree structure
	Children []TreeDataOption `json:"children,omitempty"`
}
