package dto

import "time"

// ==================== Request DTOs ====================

type CreateBoardRequest struct {
	ProjectID    string   `json:"project_id" binding:"required,uuid"`
	Title        string   `json:"title" binding:"required,min=1,max=200"`
	Content      string   `json:"content" binding:"max=5000"`

	// Legacy fields (deprecated - use custom fields instead)
	StageID      *string  `json:"stage_id" binding:"omitempty,uuid"`
	ImportanceID *string  `json:"importance_id" binding:"omitempty,uuid"`
	RoleIDs      []string `json:"role_ids" binding:"omitempty,dive,uuid"`

	AssigneeID   *string  `json:"assignee_id" binding:"omitempty,uuid"`
	DueDate      *string  `json:"dueDate" binding:"omitempty"` // ISO 8601 format
}

type UpdateBoardRequest struct {
	Title        string   `json:"title" binding:"omitempty,min=1,max=200"`
	Content      string   `json:"content" binding:"omitempty,max=5000"`

	// Legacy fields (deprecated - use custom fields instead)
	StageID      *string  `json:"stage_id" binding:"omitempty,uuid"`
	ImportanceID *string  `json:"importance_id" binding:"omitempty,uuid"`
	RoleIDs      []string `json:"role_ids" binding:"omitempty,dive,uuid"`

	AssigneeID   *string  `json:"assignee_id" binding:"omitempty,uuid"`
	DueDate      *string  `json:"dueDate" binding:"omitempty"`
}

type GetBoardsRequest struct {
	ProjectID    string `form:"projectId" binding:"required,uuid"`
	StageID      string `form:"stageId"`       // Filter: by stage
	RoleID       string `form:"roleId"`        // Filter: by role
	ImportanceID string `form:"importanceId"`  // Filter: by importance
	AssigneeID   string `form:"assigneeId"`    // Filter: by assignee
	AuthorID     string `form:"authorId"`      // Filter: by author
	Page         int    `form:"page" binding:"omitempty,min=1"`
	Limit        int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

// ==================== Response DTOs ====================

type BoardResponse struct {
	ID            string                     `json:"board_id"`
	ProjectID     string                     `json:"project_id"`
	Title         string                     `json:"title"`
	Content       string                     `json:"content"`
	Assignee      *UserInfo                  `json:"assignee"`
	Author        UserInfo                   `json:"author"`
	DueDate       *time.Time                 `json:"dueDate"`
	CreatedAt     time.Time                  `json:"createdAt"`
	UpdatedAt     time.Time                  `json:"updatedAt"`
	CustomFields  map[string]interface{}     `json:"custom_fields,omitempty"`  // Parsed custom_fields_cache
	Position      string                     `json:"position,omitempty"`       // Board position in view
}

type UserInfo struct {
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"isActive"`
}

type PaginatedBoardsResponse struct {
	Boards []BoardResponse `json:"boards"`
	Total  int64           `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

// MoveBoardRequest represents a request to move a board to a different column/group
// This API combines field value change + position update in a single transaction
// Uses fractional indexing for O(1) operations without affecting other boards
type MoveBoardRequest struct {
	ViewID         string  `json:"view_id" binding:"required,uuid"`
	GroupByFieldID string  `json:"group_by_field_id" binding:"required,uuid"` // Which field is used for grouping
	NewFieldValue  string  `json:"new_field_value" binding:"required,uuid"`   // New option_id (destination column)
	BeforePosition *string `json:"before_position"`                           // Position of board before insertion point (optional)
	AfterPosition  *string `json:"after_position"`                            // Position of board after insertion point (optional)
}

// MoveBoardResponse represents the result of a board move operation
type MoveBoardResponse struct {
	BoardID       string `json:"board_id"`
	NewFieldValue string `json:"new_field_value"`
	NewPosition   string `json:"new_position"` // New fractional index position
	Message       string `json:"message"`
}
