package dto

import "time"

// ==================== Request DTOs ====================

type CreateBoardRequest struct {
	ProjectID    string   `json:"projectId" binding:"required,uuid"`
	Title        string   `json:"title" binding:"required,min=1,max=200"`
	Content      string   `json:"content" binding:"max=5000"`

	// Legacy fields (deprecated - use custom fields instead)
	StageID      *string  `json:"stageId" binding:"omitempty,uuid"`
	ImportanceID *string  `json:"importanceId" binding:"omitempty,uuid"`
	RoleIDs      []string `json:"roleIds" binding:"omitempty,dive,uuid"`

	AssigneeID   *string  `json:"assigneeId" binding:"omitempty,uuid"`
	DueDate      *string  `json:"dueDate" binding:"omitempty"` // ISO 8601 format
}

type UpdateBoardRequest struct {
	Title        string   `json:"title" binding:"omitempty,min=1,max=200"`
	Content      string   `json:"content" binding:"omitempty,max=5000"`

	// Legacy fields (deprecated - use custom fields instead)
	StageID      *string  `json:"stageId" binding:"omitempty,uuid"`
	ImportanceID *string  `json:"importanceId" binding:"omitempty,uuid"`
	RoleIDs      []string `json:"roleIds" binding:"omitempty,dive,uuid"`

	AssigneeID   *string  `json:"assigneeId" binding:"omitempty,uuid"`
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
	ID            string                     `json:"boardId"`
	ProjectID     string                     `json:"projectId"`
	Title         string                     `json:"title"`
	Content       string                     `json:"content"`
	Assignee      *UserInfo                  `json:"assignee"`
	Author        UserInfo                   `json:"author"`
	DueDate       *time.Time                 `json:"dueDate"`
	CreatedAt     time.Time                  `json:"createdAt"`
	UpdatedAt     time.Time                  `json:"updatedAt"`
	CustomFields  map[string]interface{}     `json:"customFields,omitempty"`  // Parsed custom_fields_cache
	Position      string                     `json:"position,omitempty"`       // Board position in view
}

type UserInfo struct {
	UserID   string `json:"userId"`
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
	ViewID         string  `json:"viewId" binding:"required,uuid"`
	GroupByFieldID string  `json:"groupByFieldId" binding:"required,uuid"` // Which field is used for grouping
	NewFieldValue  string  `json:"newFieldValue" binding:"required,uuid"`   // New option_id (destination column)
	BeforePosition *string `json:"beforePosition"`                           // Position of board before insertion point (optional)
	AfterPosition  *string `json:"afterPosition"`                            // Position of board after insertion point (optional)
}

// MoveBoardResponse represents the result of a board move operation
type MoveBoardResponse struct {
	BoardID       string `json:"boardId"`
	NewFieldValue string `json:"newFieldValue"`
	NewPosition   string `json:"newPosition"` // New fractional index position
	Message       string `json:"message"`
}
