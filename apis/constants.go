package apis

import "github.com/ilxqx/vef-framework-go/constants"

// TabularFormat represents the format type for import/export operations.
type TabularFormat string

const (
	// Standard CRUD action names.
	ActionCreate          = "create"
	ActionUpdate          = "update"
	ActionDelete          = "delete"
	ActionCreateMany      = "create_many"
	ActionUpdateMany      = "update_many"
	ActionDeleteMany      = "delete_many"
	ActionFindOne         = "find_one"
	ActionFindAll         = "find_all"
	ActionFindPage        = "find_page"
	ActionFindOptions     = "find_options"
	ActionFindTree        = "find_tree"
	ActionFindTreeOptions = "find_tree_options"
	ActionImport          = "import"
	ActionExport          = "export"

	// Tabular format types for import/export.
	FormatExcel TabularFormat = "excel"
	FormatCSV   TabularFormat = "csv"

	// Internal configuration constants.

	// MaxQueryLimit is the maximum number of records that can be returned in a single query
	// to prevent excessive memory usage and performance issues.
	maxQueryLimit = 10000
	// MaxOptionsLimit is the maximum number of options that can be returned in a single query.
	maxOptionsLimit = 10000
	// DefaultLabelColumn is the default column name for option labels.
	defaultLabelColumn = "name"
	// DefaultValueColumn is the default column name for option values.
	defaultValueColumn = constants.ColumnId
	idColumn           = constants.ColumnId
	parentIdColumn     = "parent_id"
	labelColumn        = "label"
	valueColumn        = "value"
	descriptionColumn  = "description"
)
