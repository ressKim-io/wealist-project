package service

import (
	"board-service/internal/apperrors"
	"board-service/internal/client"
	"board-service/internal/domain"
	"board-service/internal/dto"
	"board-service/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ProjectService interface {
	CreateProject(userID string, token string, req *dto.CreateProjectRequest) (*dto.ProjectResponse, error)
	GetProject(projectID, userID string) (*dto.ProjectResponse, error)
	GetProjectsByWorkspaceID(workspaceID, userID string, token string) ([]dto.ProjectResponse, error)
	UpdateProject(projectID, userID string, req *dto.UpdateProjectRequest) (*dto.ProjectResponse, error)
	DeleteProject(projectID, userID string) error
	SearchProjects(userID string, token string, req *dto.SearchProjectsRequest) (*dto.PaginatedProjectsResponse, error)

	// Join Request
	CreateJoinRequest(userID string, token string, req *dto.CreateProjectJoinRequestRequest) (*dto.ProjectJoinRequestResponse, error)
	GetJoinRequests(projectID, userID string, status string) ([]dto.ProjectJoinRequestResponse, error)
	UpdateJoinRequest(requestID, userID string, req *dto.UpdateProjectJoinRequestRequest) (*dto.ProjectJoinRequestResponse, error)

	// Member
	GetProjectMembers(projectID, userID string) ([]dto.ProjectMemberResponse, error)
	UpdateMemberRole(projectID, memberID, requestUserID string, req *dto.UpdateProjectMemberRoleRequest) (*dto.ProjectMemberResponse, error)
	RemoveMember(projectID, memberID, requestUserID string) error
}

type projectService struct {
	repo               repository.ProjectRepository
	roleRepo           repository.RoleRepository
	userOrderRepo      repository.UserOrderRepository
	customFieldService CustomFieldService
	userClient         client.UserClient
	logger             *zap.Logger
	db                 *gorm.DB
}

func NewProjectService(
	repo repository.ProjectRepository,
	roleRepo repository.RoleRepository,
	userOrderRepo repository.UserOrderRepository,
	customFieldService CustomFieldService,
	userClient client.UserClient,
	logger *zap.Logger,
	db *gorm.DB,
) ProjectService {
	return &projectService{
		repo:               repo,
		roleRepo:           roleRepo,
		userOrderRepo:      userOrderRepo,
		customFieldService: customFieldService,
		userClient:         userClient,
		logger:             logger,
		db:                 db,
	}
}

// CreateProject creates a new project
func (s *projectService) CreateProject(userID string, token string, req *dto.CreateProjectRequest) (*dto.ProjectResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	workspaceUUID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	// Check if workspace exists via User Service
	ctx := context.Background()
	workspaceExists, err := s.userClient.CheckWorkspaceExists(ctx, req.WorkspaceID, token)
	if err != nil {
		s.logger.Error("Failed to check workspace existence", zap.Error(err), zap.String("workspace_id", req.WorkspaceID))
		return nil, apperrors.Wrap(err, apperrors.ErrCodeWorkspaceValidationFailed, "워크스페이스 확인 실패", 500)
	}
	if !workspaceExists {
		return nil, apperrors.New(apperrors.ErrCodeWorkspaceNotFound, "워크스페이스를 찾을 수 없습니다", 404)
	}

	// Check if user is workspace member via User Service
	isMember, err := s.userClient.ValidateWorkspaceMembership(ctx, req.WorkspaceID, userID, token)
	if err != nil {
		s.logger.Error("Failed to validate workspace membership", zap.Error(err), zap.String("workspace_id", req.WorkspaceID), zap.String("user_id", userID))
		return nil, apperrors.Wrap(err, apperrors.ErrCodeWorkspaceValidationFailed, "워크스페이스 멤버십 확인 실패", 500)
	}
	if !isMember {
		return nil, apperrors.New(apperrors.ErrCodeWorkspaceAccessDenied, "워크스페이스 멤버가 아닙니다", 403)
	}

	// Get OWNER role
	ownerRole, err := s.roleRepo.FindByName("OWNER")
	if err != nil {
		s.logger.Error("Failed to find OWNER role", zap.Error(err))
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
	}

	// Create project and member in transaction
	var project *domain.Project
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create project
		project = &domain.Project{
			WorkspaceID: workspaceUUID,
			Name:        req.Name,
			Description: req.Description,
			OwnerID:     userUUID,
		}
		if err := s.repo.Create(project); err != nil {
			return err
		}

		// Create owner member
		member := &domain.ProjectMember{
			ProjectID: project.ID,
			UserID:    userUUID,
			RoleID:    ownerRole.ID,
			JoinedAt:  time.Now(),
		}
		if err := s.repo.CreateMember(member); err != nil {
			return err
		}

		// Phase 4 - Create default custom fields
		// - Custom Roles: "없음" (system default)
		// - Custom Stages: "없음", "대기", "진행중", "완료" (system defaults)
		// - Custom Importance: "없음", "낮음", "보통", "높음", "긴급" (system defaults)
		if err := s.customFieldService.CreateDefaultCustomFields(project.ID); err != nil {
			s.logger.Error("Failed to create default custom fields", zap.Error(err), zap.String("project_id", project.ID.String()))
			return err
		}

		// Phase 6 - Initialize user-specific order settings for project owner
		if err := s.userOrderRepo.InitializeUserOrders(context.Background(), userUUID, project.ID); err != nil {
			s.logger.Error("Failed to initialize user orders", zap.Error(err), zap.String("project_id", project.ID.String()), zap.String("user_id", userUUID.String()))
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to create project", zap.Error(err))
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 생성 실패", 500)
	}

	// Fetch user info
	return s.toProjectResponse(project)
}

// GetProject retrieves a project by ID
func (s *projectService) GetProject(projectID, userID string) (*dto.ProjectResponse, error) {
	projUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is project member
	_, err = s.repo.FindMemberByUserAndProject(userUUID, projUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	project, err := s.repo.FindByID(projUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "프로젝트를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 조회 실패", 500)
	}

	return s.toProjectResponse(project)
}

// GetProjectsByWorkspaceID retrieves all projects in a workspace
func (s *projectService) GetProjectsByWorkspaceID(workspaceID, userID string, token string) ([]dto.ProjectResponse, error) {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	// Check if user is workspace member via User Service
	ctx := context.Background()
	isMember, err := s.userClient.ValidateWorkspaceMembership(ctx, workspaceID, userID, token)
	if err != nil {
		s.logger.Error("Failed to validate workspace membership", zap.Error(err), zap.String("workspace_id", workspaceID), zap.String("user_id", userID))
		return nil, apperrors.Wrap(err, apperrors.ErrCodeWorkspaceValidationFailed, "워크스페이스 멤버십 확인 실패", 500)
	}
	if !isMember {
		return nil, apperrors.New(apperrors.ErrCodeWorkspaceAccessDenied, "워크스페이스 멤버가 아닙니다", 403)
	}

	projects, err := s.repo.FindByWorkspaceID(workspaceUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 조회 실패", 500)
	}

	var responses []dto.ProjectResponse
	for _, proj := range projects {
		resp, err := s.toProjectResponse(&proj)
		if err != nil {
			s.logger.Warn("Failed to convert project to response", zap.Error(err))
			continue
		}
		responses = append(responses, *resp)
	}

	return responses, nil
}

// UpdateProject updates project information
func (s *projectService) UpdateProject(projectID, userID string, req *dto.UpdateProjectRequest) (*dto.ProjectResponse, error) {
	projUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is project OWNER
	if err := s.checkProjectOwnerPermission(userUUID, projUUID); err != nil {
		return nil, err
	}

	project, err := s.repo.FindByID(projUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "프로젝트를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 조회 실패", 500)
	}

	// Update fields
	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}

	if err := s.repo.Update(project); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 수정 실패", 500)
	}

	return s.toProjectResponse(project)
}

// DeleteProject soft deletes a project
func (s *projectService) DeleteProject(projectID, userID string) error {
	projUUID, err := uuid.Parse(projectID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is project OWNER
	if err := s.checkProjectOwnerPermission(userUUID, projUUID); err != nil {
		return err
	}

	if err := s.repo.Delete(projUUID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 삭제 실패", 500)
	}

	return nil
}

// SearchProjects searches projects in a workspace
func (s *projectService) SearchProjects(userID string, token string, req *dto.SearchProjectsRequest) (*dto.PaginatedProjectsResponse, error) {
	workspaceUUID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	// Check if user is workspace member via User Service
	ctx := context.Background()
	isMember, err := s.userClient.ValidateWorkspaceMembership(ctx, req.WorkspaceID, userID, token)
	if err != nil {
		s.logger.Error("Failed to validate workspace membership", zap.Error(err), zap.String("workspace_id", req.WorkspaceID), zap.String("user_id", userID))
		return nil, apperrors.Wrap(err, apperrors.ErrCodeWorkspaceValidationFailed, "워크스페이스 멤버십 확인 실패", 500)
	}
	if !isMember {
		return nil, apperrors.New(apperrors.ErrCodeWorkspaceAccessDenied, "워크스페이스 멤버가 아닙니다", 403)
	}

	// Default values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	projects, total, err := s.repo.Search(workspaceUUID, req.Query, req.Page, req.Limit)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 검색 실패", 500)
	}

	// Convert to response DTOs
	var responses []dto.ProjectResponse
	for _, proj := range projects {
		resp, err := s.toProjectResponse(&proj)
		if err != nil {
			s.logger.Warn("Failed to convert project to response", zap.Error(err))
			continue
		}
		responses = append(responses, *resp)
	}

	return &dto.PaginatedProjectsResponse{
		Projects: responses,
		Total:    total,
		Page:     req.Page,
		Limit:    req.Limit,
	}, nil
}

// CreateJoinRequest creates a join request
func (s *projectService) CreateJoinRequest(userID string, token string, req *dto.CreateProjectJoinRequestRequest) (*dto.ProjectJoinRequestResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	// Check if project exists
	project, err := s.repo.FindByID(projUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "프로젝트를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 조회 실패", 500)
	}

	// Check if user is workspace member via User Service
	ctx := context.Background()
	isMember, err := s.userClient.ValidateWorkspaceMembership(ctx, project.WorkspaceID.String(), userID, token)
	if err != nil {
		s.logger.Error("Failed to validate workspace membership", zap.Error(err), zap.String("workspace_id", project.WorkspaceID.String()), zap.String("user_id", userID))
		return nil, apperrors.Wrap(err, apperrors.ErrCodeWorkspaceValidationFailed, "워크스페이스 멤버십 확인 실패", 500)
	}
	if !isMember {
		return nil, apperrors.New(apperrors.ErrCodeWorkspaceAccessDenied, "워크스페이스 멤버가 아닙니다", 403)
	}

	// Check if already a member
	_, err = s.repo.FindMemberByUserAndProject(userUUID, projUUID)
	if err == nil {
		return nil, apperrors.New(apperrors.ErrCodeConflict, "이미 프로젝트 멤버입니다", 409)
	}

	// Check if request already exists
	existingReq, err := s.repo.FindJoinRequestByUserAndProject(userUUID, projUUID)
	if err == nil && existingReq.Status == domain.ProjectJoinRequestPending {
		return nil, apperrors.New(apperrors.ErrCodeConflict, "이미 참여 신청이 있습니다", 409)
	}

	// Create join request
	joinReq := &domain.ProjectJoinRequest{
		ProjectID:   projUUID,
		UserID:      userUUID,
		Status:      domain.ProjectJoinRequestPending,
		RequestedAt: time.Now(),
	}

	if err := s.repo.CreateJoinRequest(joinReq); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "참여 신청 실패", 500)
	}

	return s.toJoinRequestResponse(joinReq)
}

// GetJoinRequests retrieves join requests for a project
func (s *projectService) GetJoinRequests(projectID, userID string, status string) ([]dto.ProjectJoinRequestResponse, error) {
	projUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is project OWNER or ADMIN
	if err := s.checkProjectAdminPermission(userUUID, projUUID); err != nil {
		return nil, err
	}

	requests, err := s.repo.FindJoinRequestsByProject(projUUID, status)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "참여 신청 조회 실패", 500)
	}

	var responses []dto.ProjectJoinRequestResponse
	for _, req := range requests {
		resp, err := s.toJoinRequestResponse(&req)
		if err != nil {
			s.logger.Warn("Failed to convert join request to response", zap.Error(err))
			continue
		}
		responses = append(responses, *resp)
	}

	return responses, nil
}

// UpdateJoinRequest approves or rejects a join request
func (s *projectService) UpdateJoinRequest(requestID, userID string, req *dto.UpdateProjectJoinRequestRequest) (*dto.ProjectJoinRequestResponse, error) {
	reqUUID, err := uuid.Parse(requestID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 요청 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	joinReq, err := s.repo.FindJoinRequestByID(reqUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "참여 신청을 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "참여 신청 조회 실패", 500)
	}

	// Check if user is project OWNER or ADMIN
	if err := s.checkProjectAdminPermission(userUUID, joinReq.ProjectID); err != nil {
		return nil, err
	}

	// Update status
	joinReq.Status = domain.ProjectJoinRequestStatus(req.Status)

	// If approved, create member
	if req.Status == "APPROVED" {
		memberRole, err := s.roleRepo.FindByName("MEMBER")
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
		}

		member := &domain.ProjectMember{
			ProjectID: joinReq.ProjectID,
			UserID:    joinReq.UserID,
			RoleID:    memberRole.ID,
			JoinedAt:  time.Now(),
		}

		if err := s.repo.CreateMember(member); err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 생성 실패", 500)
		}

		// Phase 6 - Initialize user-specific order settings for new member
		if err := s.userOrderRepo.InitializeUserOrders(context.Background(), joinReq.UserID, joinReq.ProjectID); err != nil {
			s.logger.Warn("Failed to initialize user orders for new member",
				zap.Error(err),
				zap.String("project_id", joinReq.ProjectID.String()),
				zap.String("user_id", joinReq.UserID.String()))
			// Don't fail the join request if order initialization fails
		}
	}

	if err := s.repo.UpdateJoinRequest(joinReq); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "참여 신청 처리 실패", 500)
	}

	return s.toJoinRequestResponse(joinReq)
}

// GetProjectMembers retrieves all members of a project
func (s *projectService) GetProjectMembers(projectID, userID string) ([]dto.ProjectMemberResponse, error) {
	projUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is project member
	_, err = s.repo.FindMemberByUserAndProject(userUUID, projUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	members, err := s.repo.FindMembersByProject(projUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 조회 실패", 500)
	}

	var responses []dto.ProjectMemberResponse
	for _, member := range members {
		resp, err := s.toMemberResponse(&member)
		if err != nil {
			s.logger.Warn("Failed to convert member to response", zap.Error(err))
			continue
		}
		responses = append(responses, *resp)
	}

	return responses, nil
}

// UpdateMemberRole updates a member's role
func (s *projectService) UpdateMemberRole(projectID, memberID, requestUserID string, req *dto.UpdateProjectMemberRoleRequest) (*dto.ProjectMemberResponse, error) {
	projUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	memberUUID, err := uuid.Parse(memberID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 멤버 ID", 400)
	}

	reqUserUUID, err := uuid.Parse(requestUserID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if requestUser is project OWNER
	if err := s.checkProjectOwnerPermission(reqUserUUID, projUUID); err != nil {
		return nil, err
	}

	member, err := s.repo.FindMemberByID(memberUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "멤버를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 조회 실패", 500)
	}

	// Cannot change own role
	if member.UserID == reqUserUUID {
		return nil, apperrors.New(apperrors.ErrCodeBadRequest, "자신의 권한은 변경할 수 없습니다", 400)
	}

	// Get new role
	newRole, err := s.roleRepo.FindByName(req.RoleName)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
	}

	member.RoleID = newRole.ID
	if err := s.repo.UpdateMember(member); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 권한 수정 실패", 500)
	}

	return s.toMemberResponse(member)
}

// RemoveMember removes a member from project
func (s *projectService) RemoveMember(projectID, memberID, requestUserID string) error {
	projUUID, err := uuid.Parse(projectID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	memberUUID, err := uuid.Parse(memberID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 멤버 ID", 400)
	}

	reqUserUUID, err := uuid.Parse(requestUserID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if requestUser is project OWNER or ADMIN
	if err := s.checkProjectAdminPermission(reqUserUUID, projUUID); err != nil {
		return err
	}

	member, err := s.repo.FindMemberByID(memberUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "멤버를 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 조회 실패", 500)
	}

	// Cannot remove self
	if member.UserID == reqUserUUID {
		return apperrors.New(apperrors.ErrCodeBadRequest, "자신을 삭제할 수 없습니다", 400)
	}

	// Cannot remove OWNER
	role, err := s.roleRepo.FindByID(member.RoleID)
	if err == nil && role.Name == "OWNER" {
		return apperrors.New(apperrors.ErrCodeBadRequest, "OWNER는 삭제할 수 없습니다", 400)
	}

	if err := s.repo.DeleteMember(memberUUID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 삭제 실패", 500)
	}

	return nil
}

// Helper methods

func (s *projectService) checkProjectOwnerPermission(userID, projectID uuid.UUID) error {
	member, err := s.repo.FindMemberByUserAndProject(userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	role, err := s.roleRepo.FindByID(member.RoleID)
	if err != nil || role.Name != "OWNER" {
		return apperrors.New(apperrors.ErrCodeForbidden, "OWNER 권한이 필요합니다", 403)
	}

	return nil
}

func (s *projectService) checkProjectAdminPermission(userID, projectID uuid.UUID) error {
	member, err := s.repo.FindMemberByUserAndProject(userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	role, err := s.roleRepo.FindByID(member.RoleID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
	}

	if role.Name != "OWNER" && role.Name != "ADMIN" {
		return apperrors.New(apperrors.ErrCodeForbidden, "OWNER 또는 ADMIN 권한이 필요합니다", 403)
	}

	return nil
}

func (s *projectService) toProjectResponse(project *domain.Project) (*dto.ProjectResponse, error) {
	response := &dto.ProjectResponse{
		ID:          project.ID.String(),
		WorkspaceID: project.WorkspaceID.String(),
		Name:        project.Name,
		Description: project.Description,
		OwnerID:     project.OwnerID.String(),
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}

	// Fetch user info from User Service
	ctx := context.Background()
	userInfo, err := s.userClient.GetUser(ctx, project.OwnerID.String())
	if err != nil {
		s.logger.Warn("Failed to fetch user info", zap.Error(err), zap.String("user_id", project.OwnerID.String()))
		// Continue without user info
	} else {
		response.OwnerName = userInfo.Name
		response.OwnerEmail = userInfo.Email
	}

	return response, nil
}

func (s *projectService) toMemberResponse(member *domain.ProjectMember) (*dto.ProjectMemberResponse, error) {
	// Get role name
	role, err := s.roleRepo.FindByID(member.RoleID)
	if err != nil {
		return nil, err
	}

	response := &dto.ProjectMemberResponse{
		ID:        member.ID.String(),
		ProjectID: member.ProjectID.String(),
		UserID:    member.UserID.String(),
		RoleName:  role.Name,
		JoinedAt:  member.JoinedAt,
	}

	// Fetch user info from User Service
	ctx := context.Background()
	userInfo, err := s.userClient.GetUser(ctx, member.UserID.String())
	if err != nil {
		s.logger.Warn("Failed to fetch user info", zap.Error(err), zap.String("user_id", member.UserID.String()))
	} else {
		response.UserName = userInfo.Name
		response.UserEmail = userInfo.Email
	}

	return response, nil
}

func (s *projectService) toJoinRequestResponse(req *domain.ProjectJoinRequest) (*dto.ProjectJoinRequestResponse, error) {
	response := &dto.ProjectJoinRequestResponse{
		ID:          req.ID.String(),
		ProjectID:   req.ProjectID.String(),
		UserID:      req.UserID.String(),
		Status:      string(req.Status),
		RequestedAt: req.RequestedAt,
		UpdatedAt:   req.UpdatedAt,
	}

	// Fetch user info from User Service
	ctx := context.Background()
	userInfo, err := s.userClient.GetUser(ctx, req.UserID.String())
	if err != nil {
		s.logger.Warn("Failed to fetch user info", zap.Error(err), zap.String("user_id", req.UserID.String()))
	} else {
		response.UserName = userInfo.Name
		response.UserEmail = userInfo.Email
	}

	return response, nil
}
