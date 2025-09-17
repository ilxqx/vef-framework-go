package orm

const (
	PlaceholderKeyOperator = "Operator"      // PlaceholderKeyOperator is the placeholder for the operator in the db
	ExprOperator           = "?Operator"     // ExprOperator is the operator placeholder
	ExprTableColumns       = "?TableColumns" // ExprTableColumns is the table columns placeholder
	ExprColumns            = "?Columns"      // ExprColumns is the columns' placeholder
	ExprTablePKs           = "?TablePKs"     // ExprTablePKs is the table primary keys' placeholder
	ExprPKs                = "?PKs"          // ExprPKs is the primary keys placeholder
	ExprTableName          = "?TableName"    // ExprTableName is the table name placeholder
	ExprTableAlias         = "?TableAlias"   // ExprTableAlias is the table alias placeholder
	Null                   = "NULL"          // Null is the constant for the NULL value
	ColumnAll              = "*"             // ColumnAll is the constant for the all column
	SeparatorAnd           = " AND "         // SeparatorAnd is the separator for the AND condition
	SeparatorOr            = " OR "          // SeparatorOr is the separator for the OR condition
	OrderAsc               = "asc"           // OrderAsc is the constant for the ascending order
	OrderDesc              = "desc"          // OrderDesc is the constant for the descending order

	OperatorSystem    = "system"    // OperatorSystem is the operator for the system
	OperatorCronJob   = "cron_job"  // OperatorCronJob is the operator for the cron job
	OperatorAnonymous = "anonymous" // OperatorAnonymous is the operator for the anonymous

	ColumnId        = "id"         // ColumnId is the column name for the id
	ColumnCreatedAt = "created_at" // ColumnCreatedAt is the column name for the created at
	ColumnUpdatedAt = "updated_at" // ColumnUpdatedAt is the column name for the updated at
	ColumnCreatedBy = "created_by" // ColumnCreatedBy is the column name for the created by
	ColumnUpdatedBy = "updated_by" // ColumnUpdatedBy is the column name for the updated by

	FieldId        = "Id"        // FieldId is the field name for the id
	FieldCreatedAt = "CreatedAt" // FieldCreatedAt is the field name for the created at
	FieldUpdatedAt = "UpdatedAt" // FieldUpdatedAt is the field name for the updated at
	FieldCreatedBy = "CreatedBy" // FieldCreatedBy is the field name for the created by
	FieldUpdatedBy = "UpdatedBy" // FieldUpdatedBy is the field name for the updated by
)
