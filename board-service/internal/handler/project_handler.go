package handler

import (
	"board-service/internal/apperrors"
	"board-service/internal/dto"
	"board-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	service service.ProjectService
}

func NewProjectHandler(service service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

// CreateProject godoc
// @Summary      Create project
// @Description  Create a new project in a workspace (workspace member only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateProjectRequest true "Project details"
// @Success      201 {object} dto.SuccessResponse{data=dto.ProjectResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/projects [post]
// @Security     BearerAuth
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	token := c.GetString("token")
	if token == "" {
		dto.Error(c, apperrors.ErrMissingToken)
		return
	}

	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	project, err := h.service.CreateProject(userID, token, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.SuccessWithStatus(c, http.StatusCreated, project)
}

// GetProject godoc
// @Summary      Get project
// @Description  Get project details (project member only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Success      200 {object} dto.SuccessResponse{data=dto.ProjectResponse}
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/projects/{project_id} [get]
// @Security     BearerAuth
func (h *ProjectHandler) GetProject(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("project_id")

	project, err := h.service.GetProject(projectID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, project)
}

// GetProjects godoc
// @Summary      Get projects
// @Description  Get all projects in a workspace (workspace member only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        workspace_id query string true "Workspace ID"
// @Success      200 {object} dto.SuccessResponse{data=[]dto.ProjectResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/projects [get]
// @Security     BearerAuth
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	userID := c.GetString("user_id")
	token := c.GetString("token")
	workspaceID := c.Query("workspace_id")

	if workspaceID == "" {
		dto.Error(c, apperrors.New(apperrors.ErrCodeBadRequest, "workspace_id가 필요합니다", 400))
		return
	}

	if token == "" {
		dto.Error(c, apperrors.ErrMissingToken)
		return
	}

	projects, err := h.service.GetProjectsByWorkspaceID(workspaceID, userID, token)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, map[string]interface{}{"projects": projects})
}

// UpdateProject godoc
// @Summary      Update project
// @Description  Update project details (OWNER only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Param        request body dto.UpdateProjectRequest true "Updated project details"
// @Success      200 {object} dto.SuccessResponse{data=dto.ProjectResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/projects/{project_id} [put]
// @Security     BearerAuth
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("project_id")

	var req dto.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	project, err := h.service.UpdateProject(projectID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, project)
}

// DeleteProject godoc
// @Summary      Delete project
// @Description  Soft delete a project (OWNER only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Success      200 {object} dto.SuccessResponse{data=object{message=string}}
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/projects/{project_id} [delete]
// @Security     BearerAuth
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("project_id")

	if err := h.service.DeleteProject(projectID, userID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, map[string]string{"message": "프로젝트가 삭제되었습니다"})
}

// SearchProjects godoc
// @Summary      Search projects
// @Description  Search projects in a workspace by name or description
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        workspace_id query string true "Workspace ID"
// @Param        query query string true "Search query"
// @Param        page query int false "Page number (default: 1)"
// @Param        limit query int false "Page size (default: 10, max: 100)"
// @Success      200 {object} dto.SuccessResponse{data=dto.PaginatedProjectsResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/projects/search [get]
// @Security     BearerAuth
func (h *ProjectHandler) SearchProjects(c *gin.Context) {
	userID := c.GetString("user_id")
	token := c.GetString("token")

	if token == "" {
		dto.Error(c, apperrors.ErrMissingToken)
		return
	}

	var req dto.SearchProjectsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	result, err := h.service.SearchProjects(userID, token, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, result)
}

// CreateJoinRequest godoc
// @Summary      Create join request
// @Description  Request to join a project (workspace member only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateProjectJoinRequestRequest true "Join request details"
// @Success      201 {object} dto.SuccessResponse{data=dto.ProjectJoinRequestResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Router       /api/projects/join-requests [post]
// @Security     BearerAuth
func (h *ProjectHandler) CreateJoinRequest(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	token := c.GetString("token")
	if token == "" {
		dto.Error(c, apperrors.ErrMissingToken)
		return
	}

	var req dto.CreateProjectJoinRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	joinReq, err := h.service.CreateJoinRequest(userID, token, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.SuccessWithStatus(c, http.StatusCreated, joinReq)
}

// GetJoinRequests godoc
// @Summary      Get join requests
// @Description  Get join requests for a project (OWNER/ADMIN only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Param        status query string false "Filter by status (PENDING/APPROVED/REJECTED)"
// @Success      200 {object} dto.SuccessResponse{data=[]dto.ProjectJoinRequestResponse}
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/projects/{project_id}/join-requests [get]
// @Security     BearerAuth
func (h *ProjectHandler) GetJoinRequests(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("project_id")
	status := c.Query("status")

	requests, err := h.service.GetJoinRequests(projectID, userID, status)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, requests)
}

// UpdateJoinRequest godoc
// @Summary      Update join request
// @Description  Approve or reject a join request (OWNER/ADMIN only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        join_request_id path string true "Join Request ID"
// @Param        request body dto.UpdateProjectJoinRequestRequest true "Status update"
// @Success      200 {object} dto.SuccessResponse{data=dto.ProjectJoinRequestResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/projects/join-requests/{join_request_id} [put]
// @Security     BearerAuth
func (h *ProjectHandler) UpdateJoinRequest(c *gin.Context) {
	userID := c.GetString("user_id")
	requestID := c.Param("join_request_id")

	var req dto.UpdateProjectJoinRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	joinReq, err := h.service.UpdateJoinRequest(requestID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, joinReq)
}

// GetProjectMembers godoc
// @Summary      Get project members
// @Description  Get all members of a project (member only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Success      200 {object} dto.SuccessResponse{data=[]dto.ProjectMemberResponse}
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/projects/{project_id}/members [get]
// @Security     BearerAuth
func (h *ProjectHandler) GetProjectMembers(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("project_id")

	members, err := h.service.GetProjectMembers(projectID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, members)
}

// UpdateMemberRole godoc
// @Summary      Update member role
// @Description  Update a member's role in project (OWNER only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Param        member_id path string true "Member ID"
// @Param        request body dto.UpdateProjectMemberRoleRequest true "New role"
// @Success      200 {object} dto.SuccessResponse{data=dto.ProjectMemberResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/projects/{project_id}/members/{member_id}/role [put]
// @Security     BearerAuth
func (h *ProjectHandler) UpdateMemberRole(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("project_id")
	memberID := c.Param("member_id")

	var req dto.UpdateProjectMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	member, err := h.service.UpdateMemberRole(projectID, memberID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, member)
}

// RemoveMember godoc
// @Summary      Remove member
// @Description  Remove a member from project (OWNER/ADMIN only, cannot remove OWNER or self)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Param        member_id path string true "Member ID"
// @Success      200 {object} dto.SuccessResponse{data=object{message=string}}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/projects/{project_id}/members/{member_id} [delete]
// @Security     BearerAuth
func (h *ProjectHandler) RemoveMember(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("project_id")
	memberID := c.Param("member_id")

	if err := h.service.RemoveMember(projectID, memberID, userID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, map[string]string{"message": "멤버가 삭제되었습니다"})
}
