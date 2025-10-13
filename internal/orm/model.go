package orm

import "github.com/ilxqx/vef-framework-go/datetime"

// ModelPK is the primary key of the model.
type ModelPK struct {
	Id string `json:"id" bun:",pk"`
}

// Model is the base model for all models.
type Model struct {
	ModelPK `bun:"extend"`

	// CreatedAt is the created at time of the model
	CreatedAt datetime.DateTime `json:"createdAt" bun:",notnull,type:timestamp,default:CURRENT_TIMESTAMP"`
	// CreatedBy is the created by of the model
	CreatedBy string `json:"createdBy" bun:",notnull" mold:"translate=user?"`
	// CreatedByName is the created by name of the model
	CreatedByName string `json:"createdByName" bun:",scanonly"`
	// UpdatedAt is the updated at time of the model
	UpdatedAt datetime.DateTime `json:"updatedAt" bun:",notnull,type:timestamp,default:CURRENT_TIMESTAMP"`
	// UpdatedBy is the updated by of the model
	UpdatedBy string `json:"updatedBy" bun:",notnull" mold:"translate=user?"`
	// UpdatedByName is the updated by name of the model
	UpdatedByName string `json:"updatedByName" bun:",scanonly"`
}

// RelationSpec specifies how to join a related model using automatic column resolution.
// It provides a declarative way to define JOIN operations between models with minimal configuration.
// The spec automatically resolves foreign keys and primary keys based on model metadata and naming conventions.
type RelationSpec struct {
	// Model is the related model to join (e.g., (*User)(nil))
	Model any
	// Alias is the table alias for the joined model.
	// If empty, defaults to the model's default alias from table metadata.
	Alias string
	// JoinType specifies the type of JOIN operation (INNER, LEFT, RIGHT).
	// If not specified (JoinDefault), defaults to LEFT JOIN.
	JoinType JoinType
	// ForeignColumn is the column in the main table that references the joined table.
	// If empty, automatically resolves to "{model_name}_{primary_key}".
	// Example: For a User model with pk "id", defaults to "user_id".
	ForeignColumn string
	// ReferencedColumn is the column in the joined table being referenced.
	// If empty, defaults to the primary key of the joined model.
	ReferencedColumn string
	// SelectedColumns specifies which columns to select from the joined table.
	// Use ColumnInfo to configure column aliases and auto-prefixing to avoid name conflicts.
	SelectedColumns []ColumnInfo
	// On is an optional function to add custom conditions to the JOIN clause.
	// The basic equality condition (foreign_key = referenced_key) is applied automatically.
	// Use this for additional filters like soft delete checks or status conditions.
	// Example: func(cb ConditionBuilder) { cb.Equals("status", "active") }
	On ApplyFunc[ConditionBuilder]
}

// ColumnInfo represents the configuration for selecting a column from a related model.
type ColumnInfo struct {
	// Name is the column name in the database
	Name string
	// Alias is the custom alias for the column in the result set.
	// If empty and AutoAlias is false, the column uses its original name.
	Alias string
	// AutoAlias automatically generates an alias by prefixing the column name with the model name.
	// For example, if model is "User" and column is "name", the alias becomes "user_name".
	// This helps avoid column name conflicts when joining multiple tables.
	AutoAlias bool
}
