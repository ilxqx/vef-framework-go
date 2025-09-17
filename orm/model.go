package orm

import "github.com/ilxqx/vef-framework-go/mo"

// ModelPK is the primary key of the model.
type ModelPK struct {
	Id string `json:"id" bun:",pk"` // Id is the id of the model
}

// Model is the base model for all models.
type Model struct {
	ModelPK `bun:"extend"`

	CreatedAt     mo.DateTime `json:"createdAt"`                     // CreatedAt is the created at time of the model
	CreatedBy     string      `json:"createdBy"`                     // CreatedBy is the created by of the model
	CreatedByName string      `json:"createdByName" bun:",scanonly"` // CreatedByName is the created by name of the model
	UpdatedAt     mo.DateTime `json:"updatedAt"`                     // UpdatedAt is the updated at time of the model
	UpdatedBy     string      `json:"updatedBy"`                     // UpdatedBy is the updated by of the model
	UpdatedByName string      `json:"updatedByName" bun:",scanonly"` // UpdatedByName is the updated by name of the model
}

// ModelRelation is the relation between two models.
type ModelRelation struct {
	Model            any                         // Model is the model that is being related to
	ForeignColumn    string                      // ForeignColumn is the column of the model that is being related to
	ReferencedColumn string                      // ReferencedColumn is the column of the model that is being referenced
	SelectedColumns  []string                    // SelectedColumns is the columns that are being selected
	On               ApplyFunc[ConditionBuilder] // On is the condition that is being applied to the relation
}
