package dto

// ==================== Request DTOs ====================

// UpdateOrderRequest is used for drag-and-drop reordering
type UpdateOrderRequest struct {
	ItemIds []string `json:"itemIds" binding:"required,min=1,dive,uuid"` // 다른 핸들러와 일치하도록 수정
}

// ==================== Response DTOs ====================

// ColumnOrderResponse represents a single column's order information
type ColumnOrderResponse struct {
	ID            string `json:"column_id"`
	CustomRoleID  string `json:"custom_role_id,omitempty"`
	CustomStageID string `json:"custom_stage_id,omitempty"`
	DisplayOrder  int    `json:"displayOrder"`
	Name          string `json:"name"` // Retrieved from cache/DB
}

// BoardOrderResponse represents a single board's order information
type BoardOrderResponse struct {
	BoardID     string `json:"board_id"`
	Title        string `json:"title"`
	DisplayOrder int    `json:"displayOrder"`
}

// RoleBasedBoardView represents the complete board view organized by roles
type RoleBasedBoardView struct {
	ProjectID string           `json:"project_id"`
	ViewType  string           `json:"viewType"` // "role"
	Columns   []RoleColumnView `json:"columns"`
}

// RoleColumnView represents a single role column with its boards
type RoleColumnView struct {
	CustomRoleID string                `json:"custom_role_id"`
	RoleName     string                `json:"roleName"`
	RoleColor    string                `json:"roleColor,omitempty"`
	DisplayOrder int                   `json:"displayOrder"`
	Boards      []BoardOrderResponse `json:"boards"`
}

// StageBasedBoardView represents the complete board view organized by stages
type StageBasedBoardView struct {
	ProjectID string            `json:"project_id"`
	ViewType  string            `json:"viewType"` // "stage"
	Columns   []StageColumnView `json:"columns"`
}

// StageColumnView represents a single stage column with its boards
type StageColumnView struct {
	CustomStageID string                `json:"custom_stage_id"`
	StageName     string                `json:"stageName"`
	StageColor    string                `json:"stageColor,omitempty"`
	DisplayOrder  int                   `json:"displayOrder"`
	Boards       []BoardOrderResponse `json:"boards"`
}

// ==================== Internal DTOs ====================

// RoleColumnOrderData is used internally for caching
type RoleColumnOrderData struct {
	CustomRoleID string `json:"custom_role_id"`
	DisplayOrder int    `json:"displayOrder"`
	RoleName     string `json:"roleName"`
	RoleColor    string `json:"roleColor"`
}

// StageColumnOrderData is used internally for caching
type StageColumnOrderData struct {
	CustomStageID string `json:"custom_stage_id"`
	DisplayOrder  int    `json:"displayOrder"`
	StageName     string `json:"stageName"`
	StageColor    string `json:"stageColor"`
}

// BoardOrderData is used internally for caching
type BoardOrderData struct {
	BoardID     string `json:"board_id"`
	DisplayOrder int    `json:"displayOrder"`
	Title        string `json:"title"`
}
