package constants

const (
	PlaceholderKeyOperator = "Operator" // PlaceholderKeyOperator is the placeholder for the operator in the db

	OperatorSystem    = "system"    // OperatorSystem is the operator for the system
	OperatorCronJob   = "cron_job"  // OperatorCronJob is the operator for the cron job
	OperatorAnonymous = "anonymous" // OperatorAnonymous is the operator for the anonymous

	ExprOperator     = "?Operator"     // ExprOperator is the operator placeholder
	ExprTableColumns = "?TableColumns" // ExprTableColumns is the table columns placeholder
	ExprColumns      = "?Columns"      // ExprColumns is the columns' placeholder
	ExprTablePKs     = "?TablePKs"     // ExprTablePKs is the table primary keys' placeholder
	ExprPKs          = "?PKs"          // ExprPKs is the primary keys placeholder
	ExprTableName    = "?TableName"    // ExprTableName is the table name placeholder
	ExprTableAlias   = "?TableAlias"   // ExprTableAlias is the table alias placeholder

	ColumnId            = "id"              // ColumnId is the column name for the id
	ColumnCreatedAt     = "created_at"      // ColumnCreatedAt is the column name for the created at
	ColumnUpdatedAt     = "updated_at"      // ColumnUpdatedAt is the column name for the updated at
	ColumnCreatedBy     = "created_by"      // ColumnCreatedBy is the column name for the created by
	ColumnUpdatedBy     = "updated_by"      // ColumnUpdatedBy is the column name for the updated by
	ColumnCreatedByName = "created_by_name" // ColumnCreatedByName is the column name for the created by name
	ColumnUpdatedByName = "updated_by_name" // ColumnUpdatedByName is the column name for the updated by name

	FieldId            = "Id"            // FieldId is the field name for the id
	FieldCreatedAt     = "CreatedAt"     // FieldCreatedAt is the field name for the created at
	FieldUpdatedAt     = "UpdatedAt"     // FieldUpdatedAt is the field name for the updated at
	FieldCreatedBy     = "CreatedBy"     // FieldCreatedBy is the field name for the created by
	FieldUpdatedBy     = "UpdatedBy"     // FieldUpdatedBy is the field name for the updated by
	FieldCreatedByName = "CreatedByName" // FieldCreatedByName is the field name for the created by name
	FieldUpdatedByName = "UpdatedByName" // FieldUpdatedByName is the field name for the updated by name
)
