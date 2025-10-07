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
	CreatedAt datetime.DateTime `json:"createdAt"     bun:",notnull,type:timestamp,default:CURRENT_TIMESTAMP"`
	// CreatedBy is the created by of the model
	CreatedBy string `json:"createdBy"     bun:",notnull"                                          mold:"translate=user?"`
	// CreatedByName is the created by name of the model
	CreatedByName string `json:"createdByName" bun:",scanonly"`
	// UpdatedAt is the updated at time of the model
	UpdatedAt datetime.DateTime `json:"updatedAt"     bun:",notnull,type:timestamp,default:CURRENT_TIMESTAMP"`
	// UpdatedBy is the updated by of the model
	UpdatedBy string `json:"updatedBy"     bun:",notnull"                                          mold:"translate=user?"`
	// UpdatedByName is the updated by name of the model
	UpdatedByName string `json:"updatedByName" bun:",scanonly"`
}

// ModelRelation is the relation between two models.
type ModelRelation struct {
	// Model is the model that is being related to
	Model any
	// ForeignColumn is the column of the model that is being related to
	ForeignColumn string
	// ReferencedColumn is the column of the model that is being referenced
	ReferencedColumn string
	// SelectedColumns is the columns that are being selected
	SelectedColumns []string
	// On is the condition that is being applied to the relation
	On ApplyFunc[ConditionBuilder]
}
