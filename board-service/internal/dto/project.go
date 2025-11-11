package dto

import "time"

// Request DTOs

type CreateProjectRequest struct {
	WorkspaceID string `json:"workspaceId" binding:"required,uuid"`
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

type SearchProjectsRequest struct {
	WorkspaceID string `form:"workspaceId" binding:"required,uuid"`
	Query       string `form:"query" binding:"required,min=1"`
	Page        int    `form:"page" binding:"omitempty,min=1"`
	Limit       int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

type CreateProjectJoinRequestRequest struct {
	ProjectID string `json:"projectId" binding:"required,uuid"`
}

type UpdateProjectJoinRequestRequest struct {
	Status string `json:"status" binding:"required,oneof=APPROVED REJECTED"`
}

type UpdateProjectMemberRoleRequest struct {
	RoleName string `json:"roleName" binding:"required,oneof=OWNER ADMIN MEMBER"`
}

// Response DTOs

type ProjectResponse struct {
	ID          string    `json:"projectId"`
	WorkspaceID string    `json:"workspaceId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"ownerId"`
	OwnerName   string    `json:"ownerName"`
	OwnerEmail  string    `json:"ownerEmail"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ProjectMemberResponse struct {
	ID        string    `json:"memberId"`
	ProjectID string    `json:"projectId"`
	UserID    string    `json:"userId"`
	UserName  string    `json:"userName"`
	UserEmail string    `json:"userEmail"`
	RoleName  string    `json:"roleName"`
	JoinedAt  time.Time `json:"joinedAt"`
}

type ProjectJoinRequestResponse struct {
	ID          string    `json:"requestId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	UserName    string    `json:"userName"`
	UserEmail   string    `json:"userEmail"`
	Status      string    `json:"status"`
	RequestedAt time.Time `json:"requestedAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type PaginatedProjectsResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
}
