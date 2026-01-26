package apis

import "github.com/ilxqx/vef-framework-go/constants"

// TabularFormat represents the format type for import/export operations.
type TabularFormat string

// i18n message keys for APIs.
const (
	ErrMessageProcessorMustReturnSlice = "processor_must_return_slice"
)

// Error codes for APIs.
const (
	ErrCodeProcessorInvalidReturn = 2400
)

const (
	RPCActionCreate          = "create"
	RPCActionUpdate          = "update"
	RPCActionDelete          = "delete"
	RPCActionCreateMany      = "create_many"
	RPCActionUpdateMany      = "update_many"
	RPCActionDeleteMany      = "delete_many"
	RPCActionFindOne         = "find_one"
	RPCActionFindAll         = "find_all"
	RPCActionFindPage        = "find_page"
	RPCActionFindOptions     = "find_options"
	RPCActionFindTree        = "find_tree"
	RPCActionFindTreeOptions = "find_tree_options"
	RPCActionImport          = "import"
	RPCActionExport          = "export"

	// REST Action format: "<method> <path>", path supports Fiber route patterns (e.g., /:id).
	RESTActionCreate          = "post /"
	RESTActionUpdate          = "put /:" + IDColumn
	RESTActionDelete          = "delete /:" + IDColumn
	RESTActionCreateMany      = "post /many"
	RESTActionUpdateMany      = "put /many"
	RESTActionDeleteMany      = "delete /many"
	RESTActionFindOne         = "get /:" + IDColumn
	RESTActionFindAll         = "get /"
	RESTActionFindPage        = "get /page"
	RESTActionFindOptions     = "get /options"
	RESTActionFindTree        = "get /tree"
	RESTActionFindTreeOptions = "get /tree/options"
	RESTActionImport          = "post /import"
	RESTActionExport          = "get /export"

	// Tabular format types for import/export.
	FormatExcel TabularFormat = "excel"
	FormatCsv   TabularFormat = "csv"

	maxQueryLimit              = 10000
	maxOptionsLimit            = 10000
	defaultAuditUserNameColumn = "name"
	defaultLabelColumn         = "name"
	defaultValueColumn         = constants.ColumnID

	IDColumn          = constants.ColumnID
	ParentIDColumn    = "parent_id"
	LabelColumn       = "label"
	ValueColumn       = "value"
	DescriptionColumn = "description"
)
