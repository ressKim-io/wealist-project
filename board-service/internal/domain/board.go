package domain

import (
	"time"

	"github.com/google/uuid"
)

type Board struct {
	BaseModel
	ProjectID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	Title              string     `gorm:"type:varchar(255);not null" json:"title"`
	Description        string     `gorm:"type:text" json:"description"`

	// Legacy fields (deprecated - use custom_fields_cache instead)
	CustomStageID      *uuid.UUID `gorm:"type:uuid;index" json:"custom_stage_id,omitempty"`
	CustomImportanceID *uuid.UUID `gorm:"type:uuid;index" json:"custom_importance_id,omitempty"`

	AssigneeID         *uuid.UUID `gorm:"type:uuid;index" json:"assignee_id"`
	CreatedBy          uuid.UUID  `gorm:"type:uuid;not null;index" json:"created_by"`
	DueDate            *time.Time `gorm:"index" json:"due_date"`

	// Custom fields cache (JSONB for fast filtering with GIN index)
	CustomFieldsCache  string     `gorm:"type:jsonb;default:'{}'" json:"custom_fields_cache"`
}

func (Board) TableName() string {
	return "boards"
}
