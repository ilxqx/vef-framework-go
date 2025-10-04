package apis

import "github.com/ilxqx/vef-framework-go/constants"

const (
	// Standard CRUD action names
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

	// Internal configuration constants

	// maxQueryLimit is the maximum number of records that can be returned in a single query
	// to prevent excessive memory usage and performance issues
	maxQueryLimit = 10000
	// maxOptionsLimit is the maximum number of options that can be returned in a single query
	maxOptionsLimit = 10000
	// defaultLabelField is the default field name for option labels
	defaultLabelField = "name"
	// defaultValueField is the default field name for option values
	defaultValueField = constants.ColumnId
	idField           = constants.ColumnId
	parentIdField     = "parent_id"
	labelField        = "label"
	valueField        = "value"
	descriptionField  = "description"
)
