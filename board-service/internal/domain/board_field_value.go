package domain

import (
	"time"

	"github.com/google/uuid"
)

// BoardFieldValue represents a field value for a specific board (EAV pattern)
// Only one of value_* columns should be non-null based on the field type
type BoardFieldValue struct {
	BaseModel
	BoardID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"board_id"`
	FieldID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"field_id"`

	// Value columns (only one should be NOT NULL)
	ValueText     *string    `gorm:"type:text" json:"value_text,omitempty"`
	ValueNumber   *float64   `gorm:"type:numeric(15,4)" json:"value_number,omitempty"`
	ValueDate     *time.Time `gorm:"type:timestamp" json:"value_date,omitempty"`
	ValueBoolean  *bool      `gorm:"type:boolean" json:"value_boolean,omitempty"`
	ValueOptionID *uuid.UUID `gorm:"type:uuid;index" json:"value_option_id,omitempty"`
	ValueUserID   *uuid.UUID `gorm:"type:uuid;index" json:"value_user_id,omitempty"`

	// Display order (for multi-select and multi-user)
	DisplayOrder  int        `gorm:"default:0" json:"display_order"`
}

func (BoardFieldValue) TableName() string {
	return "board_field_values"
}
