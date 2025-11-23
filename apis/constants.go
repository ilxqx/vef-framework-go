package apis

import "github.com/ilxqx/vef-framework-go/constants"

// TabularFormat represents the format type for import/export operations.
type TabularFormat string

const (
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
	FormatCsv   TabularFormat = "csv"

	maxQueryLimit              = 10000
	maxOptionsLimit            = 10000
	defaultAuditUserNameColumn = "name"
	defaultLabelColumn         = "name"
	defaultValueColumn         = constants.ColumnId

	IdColumn          = constants.ColumnId
	ParentIdColumn    = "parent_id"
	LabelColumn       = "label"
	ValueColumn       = "value"
	DescriptionColumn = "description"
)
