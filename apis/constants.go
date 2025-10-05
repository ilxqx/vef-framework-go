package apis

import "github.com/ilxqx/vef-framework-go/constants"

// TabularFormat represents the format type for import/export operations.
type TabularFormat string

const (
	// Standard CRUD action names.
	ActionCreate          = "create"
	ActionUpdate          = "update"
	ActionDelete          = "delete"
	ActionCreateMany      = "createMany"
	ActionUpdateMany      = "updateMany"
	ActionDeleteMany      = "deleteMany"
	ActionFindOne         = "findOne"
	ActionFindAll         = "findAll"
	ActionFindPage        = "findPage"
	ActionFindOptions     = "findOptions"
	ActionFindTree        = "findTree"
	ActionFindTreeOptions = "findTreeOptions"
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
	// DefaultLabelField is the default field name for option labels.
	defaultLabelField = "name"
	// DefaultValueField is the default field name for option values.
	defaultValueField = constants.ColumnId
	idField           = constants.ColumnId
	parentIdField     = "parent_id"
	labelField        = "label"
	valueField        = "value"
	descriptionField  = "description"
)
