package dto

import "time"

// ==================== Project Field DTOs ====================

// CreateFieldRequest represents a request to create a custom field
type CreateFieldRequest struct {
	ProjectID   string                 `json:"projectId" binding:"required,uuid"`
	Name        string                 `json:"name" binding:"required,min=1,max=255"`
	FieldType   string                 `json:"fieldType" binding:"required,oneof=text number single_select multi_select date datetime single_user multi_user checkbox url"`
	Description string                 `json:"description" binding:"omitempty,max=1000"`
	IsRequired  bool                   `json:"isRequired"`
	Config      map[string]interface{} `json:"config"` // Type-specific configuration
}

// UpdateFieldRequest represents a request to update a custom field
type UpdateFieldRequest struct {
	Name        string                 `json:"name" binding:"omitempty,min=1,max=255"`
	Description string                 `json:"description" binding:"omitempty,max=1000"`
	IsRequired  *bool                  `json:"isRequired"`
	Config      map[string]interface{} `json:"config"`
	DisplayOrder *int                  `json:"displayOrder"`
}

// UpdateFieldOrderRequest represents a request to update field display order
type UpdateFieldOrderRequest struct {
	FieldOrders []FieldOrder `json:"fieldOrders" binding:"required,min=1,dive"`
}

type FieldOrder struct {
	FieldID      string `json:"fieldId" binding:"required,uuid"`
	DisplayOrder int    `json:"displayOrder" binding:"min=0"`
}

// FieldResponse represents a custom field
type FieldResponse struct {
	FieldID         string                 `json:"fieldId"`
	ProjectID       string                 `json:"projectId"`
	Name            string                 `json:"name"`
	FieldType       string                 `json:"fieldType"`
	Description     string                 `json:"description"`
	DisplayOrder    int                    `json:"displayOrder"`
	IsRequired      bool                   `json:"isRequired"`
	IsSystemDefault bool                   `json:"isSystemDefault"`
	Config          map[string]interface{} `json:"config"`
	CanEditRoles    []string               `json:"canEditRoles"`
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}

// ==================== Field Option DTOs ====================

// CreateOptionRequest represents a request to create a field option
type CreateOptionRequest struct {
	FieldID     string `json:"fieldId" binding:"required,uuid"`
	Label       string `json:"label" binding:"required,min=1,max=255"`
	Color       string `json:"color" binding:"omitempty,len=7"` // #RRGGBB
	Description string `json:"description" binding:"omitempty,max=500"`
}

// UpdateOptionRequest represents a request to update a field option
type UpdateOptionRequest struct {
	Label       string `json:"label" binding:"omitempty,min=1,max=255"`
	Color       string `json:"color" binding:"omitempty,len=7"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

// UpdateOptionOrderRequest represents a request to update option display order
type UpdateOptionOrderRequest struct {
	OptionOrders []OptionOrder `json:"optionOrders" binding:"required,min=1,dive"`
}

type OptionOrder struct {
	OptionID     string `json:"optionId" binding:"required,uuid"`
	DisplayOrder int    `json:"displayOrder" binding:"min=0"`
}

// OptionResponse represents a field option
type OptionResponse struct {
	OptionID     string    `json:"optionId"`
	FieldID      string    `json:"fieldId"`
	Label        string    `json:"label"`
	Color        string    `json:"color"`
	Description  string    `json:"description"`
	DisplayOrder int       `json:"displayOrder"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ==================== Board Field Value DTOs ====================

// SetFieldValueRequest represents a request to set a field value for a board
type SetFieldValueRequest struct {
	BoardID   string      `json:"boardId" binding:"required,uuid"`
	FieldID   string      `json:"fieldId" binding:"required,uuid"`
	Value     interface{} `json:"value"`  // Type depends on field type
	Values    interface{} `json:"values"` // For multi_select, multi_user (array of IDs or ordered map)
}

// SetMultiSelectValueRequest represents ordered multi-select values
type SetMultiSelectValueRequest struct {
	BoardID  string                  `json:"boardId" binding:"required,uuid"`
	FieldID  string                  `json:"fieldId" binding:"required,uuid"`
	Values   []OrderedValue          `json:"values" binding:"required,dive"`
}

type OrderedValue struct {
	ValueID      string `json:"valueId" binding:"required,uuid"` // option_id or user_id
	DisplayOrder int    `json:"displayOrder" binding:"min=0"`
}

// FieldValueResponse represents a board field value
type FieldValueResponse struct {
	ValueID       string      `json:"valueId"`
	BoardID       string      `json:"boardId"`
	FieldID       string      `json:"fieldId"`
	Value         interface{} `json:"value"`  // Actual value (text, number, date, boolean, option object, user object)
	DisplayOrder  int         `json:"displayOrder,omitempty"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
}

// BoardFieldValuesResponse represents all field values for a board
type BoardFieldValuesResponse struct {
	BoardID string                 `json:"boardId"`
	Fields  map[string]interface{} `json:"fields"` // map[field_id]value
}

// ==================== Saved View DTOs ====================

// CreateViewRequest represents a request to create a saved view
type CreateViewRequest struct {
	ProjectID      string                 `json:"projectId" binding:"required,uuid"`
	Name           string                 `json:"name" binding:"required,min=1,max=255"`
	Description    string                 `json:"description" binding:"omitempty,max=1000"`
	IsDefault      bool                   `json:"isDefault"`
	IsShared       bool                   `json:"isShared"`
	Filters        map[string]interface{} `json:"filters"`
	SortBy         string                 `json:"sortBy" binding:"omitempty"`
	SortDirection  string                 `json:"sortDirection" binding:"omitempty,oneof=asc desc"`
	GroupByFieldID string                 `json:"groupByFieldId" binding:"omitempty,uuid"`
}

// UpdateViewRequest represents a request to update a saved view
type UpdateViewRequest struct {
	Name           string                 `json:"name" binding:"omitempty,min=1,max=255"`
	Description    string                 `json:"description" binding:"omitempty,max=1000"`
	IsDefault      *bool                  `json:"isDefault"`
	IsShared       *bool                  `json:"isShared"`
	Filters        map[string]interface{} `json:"filters"`
	SortBy         *string                `json:"sortBy"`
	SortDirection  string                 `json:"sortDirection" binding:"omitempty,oneof=asc desc"`
	GroupByFieldID *string                `json:"groupByFieldId" binding:"omitempty,uuid"`
}

// ViewResponse represents a saved view
type ViewResponse struct {
	ViewID         string                 `json:"viewId"`
	ProjectID      string                 `json:"projectId"`
	CreatedBy      string                 `json:"createdBy"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	IsDefault      bool                   `json:"isDefault"`
	IsShared       bool                   `json:"isShared"`
	Filters        map[string]interface{} `json:"filters"`
	SortBy         string                 `json:"sortBy"`
	SortDirection  string                 `json:"sortDirection"`
	GroupByFieldID string                 `json:"groupByFieldId"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
}

// ApplyViewRequest represents a request to apply a view and get filtered boards
type ApplyViewRequest struct {
	ViewID string `form:"viewId" binding:"required,uuid"`
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

// GroupedBoardsResponse represents boards grouped by a field
type GroupedBoardsResponse struct {
	GroupByField FieldResponse              `json:"groupByField"`
	Groups       []BoardGroup               `json:"groups"`
	Total        int64                      `json:"total"`
}

type BoardGroup struct {
	GroupValue interface{}     `json:"groupValue"` // Option or User object
	Boards     []BoardResponse `json:"boards"`
	Count      int             `json:"count"`
}

// ==================== User Board Order DTOs ====================

// UpdateBoardOrderRequest represents a request to update board order in a view
// Note: Prefer using MoveBoardRequest API for single board moves (O(1) fractional indexing)
type UpdateBoardOrderRequest struct {
	ViewID      string       `json:"viewId" binding:"required,uuid"`
	BoardOrders []BoardOrder `json:"boardOrders" binding:"required,dive"`
}

type BoardOrder struct {
	BoardID  string `json:"boardId" binding:"required,uuid"`
	Position string `json:"position" binding:"required"` // Fractional index position
}
