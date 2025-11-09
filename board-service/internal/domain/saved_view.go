package domain

import "github.com/google/uuid"

// SavedView represents a user-defined view with filters, sorting, and grouping
type SavedView struct {
	BaseModel
	ProjectID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	CreatedBy      uuid.UUID  `gorm:"type:uuid;not null;index" json:"created_by"`
	Name           string     `gorm:"type:varchar(255);not null" json:"name"`
	Description    string     `gorm:"type:text" json:"description"`
	IsDefault      bool       `gorm:"default:false;index" json:"is_default"`
	IsShared       bool       `gorm:"default:false" json:"is_shared"`
	Filters        string     `gorm:"type:text;default:'{}'" json:"filters"`       // JSON stored as string
	SortBy         *string    `gorm:"type:varchar(255)" json:"sort_by"`
	SortDirection  string     `gorm:"type:varchar(4);default:'asc'" json:"sort_direction"` // 'asc' or 'desc'
	GroupByFieldID *uuid.UUID `gorm:"type:uuid" json:"group_by_field_id"`
}

func (SavedView) TableName() string {
	return "saved_views"
}

// ViewFilters represents filter configuration
// This is parsed from/to the Filters JSON string
type ViewFilters map[string]FilterCondition

type FilterCondition struct {
	Operator string      `json:"operator"` // 'eq', 'ne', 'in', 'not_in', 'contains', 'gt', 'gte', 'lt', 'lte', 'is_null', 'is_not_null'
	Value    interface{} `json:"value"`
}
