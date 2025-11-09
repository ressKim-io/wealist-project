package domain

import "github.com/google/uuid"

// FieldOption represents an option for single_select or multi_select field types
type FieldOption struct {
	BaseModel
	FieldID      uuid.UUID `gorm:"type:uuid;not null;index" json:"field_id"`
	Label        string    `gorm:"type:varchar(255);not null" json:"label"`
	Color        string    `gorm:"type:varchar(7)" json:"color"` // HEX color: #RRGGBB
	Description  string    `gorm:"type:text" json:"description"`
	DisplayOrder int       `gorm:"not null;default:0" json:"display_order"`
}

func (FieldOption) TableName() string {
	return "field_options"
}
