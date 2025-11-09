package domain

import "github.com/google/uuid"

// FieldType represents the data type of a custom field
type FieldType string

const (
	FieldTypeText        FieldType = "text"
	FieldTypeNumber      FieldType = "number"
	FieldTypeSingleSelect FieldType = "single_select"
	FieldTypeMultiSelect FieldType = "multi_select"
	FieldTypeDate        FieldType = "date"
	FieldTypeDateTime    FieldType = "datetime"
	FieldTypeSingleUser  FieldType = "single_user"
	FieldTypeMultiUser   FieldType = "multi_user"
	FieldTypeCheckbox    FieldType = "checkbox"
	FieldTypeURL         FieldType = "url"
)

// ProjectField represents a custom field definition for a project (Jira-style)
type ProjectField struct {
	BaseModel
	ProjectID       uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	Name            string    `gorm:"type:varchar(255);not null" json:"name"`
	FieldType       FieldType `gorm:"type:varchar(50);not null;index" json:"field_type"`
	Description     string    `gorm:"type:text" json:"description"`
	DisplayOrder    int       `gorm:"not null;default:0" json:"display_order"`
	IsRequired      bool      `gorm:"not null;default:false" json:"is_required"`
	IsSystemDefault bool      `gorm:"not null;default:false;index" json:"is_system_default"`
	Config          string    `gorm:"type:text;not null;default:'{}'" json:"config"` // JSON stored as string
	CanEditRoles    *string   `gorm:"type:text" json:"can_edit_roles"`              // Comma-separated roles
}

func (ProjectField) TableName() string {
	return "project_fields"
}

// FieldConfig represents type-specific configuration
// This is parsed from/to the Config JSON string
type FieldConfig struct {
	// Text
	MaxLength *int  `json:"max_length,omitempty"`
	IsLong    *bool `json:"is_long,omitempty"`

	// Number
	Min            *float64 `json:"min,omitempty"`
	Max            *float64 `json:"max,omitempty"`
	DecimalPlaces  *int     `json:"decimal_places,omitempty"`

	// Multi-select
	MaxSelections *int `json:"max_selections,omitempty"`

	// DateTime
	IncludeTime *bool `json:"include_time,omitempty"`

	// Multi-user
	MaxUsers             *int  `json:"max_users,omitempty"`
	ProjectMembersOnly   *bool `json:"project_members_only,omitempty"`

	// Checkbox
	DefaultValue *bool `json:"default_value,omitempty"`

	// URL
	EnablePreview *bool `json:"enable_preview,omitempty"`
}
