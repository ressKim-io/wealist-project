package dto

import "time"

// ==================== Request DTOs ====================

type CreateBoardRequest struct {
	ProjectID    string   `json:"project_id" binding:"required,uuid"`
	Title        string   `json:"title" binding:"required,min=1,max=200"`
	Content      string   `json:"content" binding:"max=5000"`
	StageID      string   `json:"stage_id" binding:"required,uuid"`
	ImportanceID *string  `json:"importance_id" binding:"omitempty,uuid"`
	RoleIDs      []string `json:"role_ids" binding:"required,min=1,dive,uuid"`
	AssigneeID   *string  `json:"assignee_id" binding:"omitempty,uuid"`
	DueDate      *string  `json:"dueDate" binding:"omitempty"` // ISO 8601 format
}

type UpdateBoardRequest struct {
	Title        string   `json:"title" binding:"omitempty,min=1,max=200"`
	Content      string   `json:"content" binding:"omitempty,max=5000"`
	StageID      string   `json:"stage_id" binding:"omitempty,uuid"`
	ImportanceID *string  `json:"importance_id" binding:"omitempty,uuid"`
	RoleIDs      []string `json:"role_ids" binding:"omitempty,dive,uuid"`
	AssigneeID   *string  `json:"assignee_id" binding:"omitempty,uuid"`
	DueDate      *string  `json:"dueDate" binding:"omitempty"`
}

type GetBoardsRequest struct {
	ProjectID    string `form:"project_id" binding:"required,uuid"`
	StageID      string `form:"stage_id"`       // Filter: by stage
	RoleID       string `form:"role_id"`        // Filter: by role
	ImportanceID string `form:"importance_id"`  // Filter: by importance
	AssigneeID   string `form:"assignee_id"`    // Filter: by assignee
	AuthorID     string `form:"author_id"`      // Filter: by author
	Page         int    `form:"page" binding:"omitempty,min=1"`
	Limit        int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

// ==================== Response DTOs ====================

type BoardResponse struct {
	ID         string                     `json:"board_id"`
	ProjectID  string                     `json:"project_id"`
	Title      string                     `json:"title"`
	Content    string                     `json:"content"`
	Stage      CustomStageResponse        `json:"stage"`
	Importance *CustomImportanceResponse  `json:"importance"`
	Roles      []CustomRoleResponse       `json:"roles"`
	Assignee   *UserInfo                  `json:"assignee"`
	Author     UserInfo                   `json:"author"`
	DueDate    *time.Time                 `json:"dueDate"`
	CreatedAt  time.Time                  `json:"createdAt"`
	UpdatedAt  time.Time                  `json:"updatedAt"`
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
