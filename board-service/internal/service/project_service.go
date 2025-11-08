package service

import (
	"board-service/internal/apperrors"
	"board-service/internal/cache"
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
	workspaceCache     cache.WorkspaceCache
	userInfoCache      cache.UserInfoCache
	logger             *zap.Logger
	db                 *gorm.DB
}

func NewProjectService(
	repo repository.ProjectRepository,
	roleRepo repository.RoleRepository,
	userOrderRepo repository.UserOrderRepository,
	customFieldService CustomFieldService,
	userClient client.UserClient,
	workspaceCache cache.WorkspaceCache,
	userInfoCache cache.UserInfoCache,
	logger *zap.Logger,
	db *gorm.DB,
) ProjectService {
	return &projectService{
		repo:               repo,
		roleRepo:           roleRepo,
		userOrderRepo:      userOrderRepo,
		customFieldService: customFieldService,
		userClient:         userClient,
		workspaceCache:     workspaceCache,
		userInfoCache:      userInfoCache,
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

	// Check workspace membership with caching
	ctx := context.Background()
	if err := s.validateWorkspaceMembership(ctx, req.WorkspaceID, userID, token); err != nil {
		return nil, err
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

	// Check workspace membership with caching
	ctx := context.Background()
	if err := s.validateWorkspaceMembership(ctx, workspaceID, userID, token); err != nil {
		return nil, err
	}

	projects, err := s.repo.FindByWorkspaceID(workspaceUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 조회 실패", 500)
	}

	// Batch fetch owner info
	ownerIDs := make([]string, 0, len(projects))
	for _, proj := range projects {
		ownerIDs = append(ownerIDs, proj.OwnerID.String())
	}
	userMap := s.getUserInfoBatch(ctx, ownerIDs)

	// Convert to responses
	responses := make([]dto.ProjectResponse, 0, len(projects))
	for _, proj := range projects {
		response := &dto.ProjectResponse{
			ID:          proj.ID.String(),
			WorkspaceID: proj.WorkspaceID.String(),
			Name:        proj.Name,
			Description: proj.Description,
			OwnerID:     proj.OwnerID.String(),
			CreatedAt:   proj.CreatedAt,
			UpdatedAt:   proj.UpdatedAt,
		}

		// Add owner info from batch result
		if userInfo, ok := userMap[proj.OwnerID.String()]; ok {
			response.OwnerName = userInfo.Name
			response.OwnerEmail = userInfo.Email
		}

		responses = append(responses, *response)
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

	// Check workspace membership with caching
	ctx := context.Background()
	if err := s.validateWorkspaceMembership(ctx, req.WorkspaceID, userID, token); err != nil {
		return nil, err
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

	// Batch fetch owner info
	ownerIDs := make([]string, 0, len(projects))
	for _, proj := range projects {
		ownerIDs = append(ownerIDs, proj.OwnerID.String())
	}
	userMap := s.getUserInfoBatch(ctx, ownerIDs)

	// Convert to response DTOs
	responses := make([]dto.ProjectResponse, 0, len(projects))
	for _, proj := range projects {
		response := &dto.ProjectResponse{
			ID:          proj.ID.String(),
			WorkspaceID: proj.WorkspaceID.String(),
			Name:        proj.Name,
			Description: proj.Description,
			OwnerID:     proj.OwnerID.String(),
			CreatedAt:   proj.CreatedAt,
			UpdatedAt:   proj.UpdatedAt,
		}

		// Add owner info from batch result
		if userInfo, ok := userMap[proj.OwnerID.String()]; ok {
			response.OwnerName = userInfo.Name
			response.OwnerEmail = userInfo.Email
		}

		responses = append(responses, *response)
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

	// Check workspace membership with caching
	ctx := context.Background()
	if err := s.validateWorkspaceMembership(ctx, project.WorkspaceID.String(), userID, token); err != nil {
		return nil, err
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

	// Batch fetch user info
	ctx := context.Background()
	userIDs := make([]string, 0, len(requests))
	for _, req := range requests {
		userIDs = append(userIDs, req.UserID.String())
	}
	userMap := s.getUserInfoBatch(ctx, userIDs)

	// Convert to responses
	responses := make([]dto.ProjectJoinRequestResponse, 0, len(requests))
	for _, req := range requests {
		response := &dto.ProjectJoinRequestResponse{
			ID:          req.ID.String(),
			ProjectID:   req.ProjectID.String(),
			UserID:      req.UserID.String(),
			Status:      string(req.Status),
			RequestedAt: req.RequestedAt,
			UpdatedAt:   req.UpdatedAt,
		}

		// Add user info from batch result
		if userInfo, ok := userMap[req.UserID.String()]; ok {
			response.UserName = userInfo.Name
			response.UserEmail = userInfo.Email
		}

		responses = append(responses, *response)
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

	// Batch fetch user info
	ctx := context.Background()
	userIDs := make([]string, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID.String())
	}
	userMap := s.getUserInfoBatch(ctx, userIDs)

	// Convert to responses
	responses := make([]dto.ProjectMemberResponse, 0, len(members))
	for _, member := range members {
		// Get role name
		role, err := s.roleRepo.FindByID(member.RoleID)
		if err != nil {
			s.logger.Warn("Failed to get role", zap.Error(err))
			continue
		}

		response := &dto.ProjectMemberResponse{
			ID:        member.ID.String(),
			ProjectID: member.ProjectID.String(),
			UserID:    member.UserID.String(),
			RoleName:  role.Name,
			JoinedAt:  member.JoinedAt,
		}

		// Add user info from batch result
		if userInfo, ok := userMap[member.UserID.String()]; ok {
			response.UserName = userInfo.Name
			response.UserEmail = userInfo.Email
		}

		responses = append(responses, *response)
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

// validateWorkspaceMembership checks workspace membership with caching
func (s *projectService) validateWorkspaceMembership(ctx context.Context, workspaceID, userID, token string) error {
	// Try cache first
	cacheExists, isMember, err := s.workspaceCache.GetMembership(ctx, workspaceID, userID)
	if err != nil {
		s.logger.Warn("Failed to get workspace membership from cache", zap.Error(err))
		// Continue to User Service call on cache error
	}

	if cacheExists {
		// Cache hit
		if !isMember {
			return apperrors.New(apperrors.ErrCodeWorkspaceAccessDenied, "워크스페이스 멤버가 아닙니다", 403)
		}
		return nil
	}

	// Cache miss - validate via User Service
	// First check if workspace exists
	workspaceExists, err := s.userClient.CheckWorkspaceExists(ctx, workspaceID, token)
	if err != nil {
		s.logger.Error("Failed to check workspace existence", zap.Error(err), zap.String("workspace_id", workspaceID))
		return apperrors.Wrap(err, apperrors.ErrCodeWorkspaceValidationFailed, "워크스페이스 확인 실패", 500)
	}
	if !workspaceExists {
		return apperrors.New(apperrors.ErrCodeWorkspaceNotFound, "워크스페이스를 찾을 수 없습니다", 404)
	}

	// Check membership
	isMember, err = s.userClient.ValidateWorkspaceMembership(ctx, workspaceID, userID, token)
	if err != nil {
		s.logger.Error("Failed to validate workspace membership", zap.Error(err), zap.String("workspace_id", workspaceID), zap.String("user_id", userID))
		return apperrors.Wrap(err, apperrors.ErrCodeWorkspaceValidationFailed, "워크스페이스 멤버십 확인 실패", 500)
	}

	// Cache the result
	if cacheErr := s.workspaceCache.SetMembership(ctx, workspaceID, userID, isMember); cacheErr != nil {
		s.logger.Warn("Failed to cache workspace membership", zap.Error(cacheErr))
		// Don't fail the request on cache write error
	}

	if !isMember {
		return apperrors.New(apperrors.ErrCodeWorkspaceAccessDenied, "워크스페이스 멤버가 아닙니다", 403)
	}

	return nil
}

// getUserInfoWithCache retrieves user info with caching
func (s *projectService) getUserInfoWithCache(ctx context.Context, userID string) (*dto.UserInfo, error) {
	// Try cache first
	cacheExists, cachedUser, err := s.userInfoCache.GetUserInfo(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get user info from cache", zap.Error(err))
	}

	if cacheExists && cachedUser != nil {
		// Cache hit - convert to dto.UserInfo
		return &dto.UserInfo{
			UserID: cachedUser.UserID,
			Name:   cachedUser.Name,
			Email:  cachedUser.Email,
		}, nil
	}

	// Cache miss - fetch from User Service
	userInfo, err := s.userClient.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	cacheUser := &cache.UserInfo{
		UserID:   userInfo.UserID,
		Name:     userInfo.Name,
		Email:    userInfo.Email,
		IsActive: userInfo.IsActive,
	}
	if cacheErr := s.userInfoCache.SetUserInfo(ctx, cacheUser); cacheErr != nil {
		s.logger.Warn("Failed to cache user info", zap.Error(cacheErr))
	}

	return &dto.UserInfo{
		UserID: userInfo.UserID,
		Name:   userInfo.Name,
		Email:  userInfo.Email,
	}, nil
}

// getUserInfoBatch retrieves multiple user infos with caching
func (s *projectService) getUserInfoBatch(ctx context.Context, userIDs []string) map[string]*dto.UserInfo {
	if len(userIDs) == 0 {
		return make(map[string]*dto.UserInfo)
	}

	// Try cache first
	cachedUsers, err := s.userInfoCache.GetSimpleUsersBatch(ctx, userIDs)
	if err != nil {
		s.logger.Warn("Failed to get users from cache", zap.Error(err))
		cachedUsers = make(map[string]*cache.SimpleUser)
	}

	// Find missing user IDs
	missingUserIDs := []string{}
	for _, userID := range userIDs {
		if _, exists := cachedUsers[userID]; !exists {
			missingUserIDs = append(missingUserIDs, userID)
		}
	}

	// Build result map
	userMap := make(map[string]*dto.UserInfo)

	// Fetch missing users from User Service
	if len(missingUserIDs) > 0 {
		users, err := s.userClient.GetUsersBatch(ctx, missingUserIDs)
		if err != nil {
			s.logger.Warn("Failed to fetch users from User Service", zap.Error(err))
		} else {
			// Cache the fetched users
			simpleUsers := make([]cache.SimpleUser, 0, len(users))
			for _, user := range users {
				userMap[user.UserID] = &dto.UserInfo{
					UserID: user.UserID,
					Name:   user.Name,
					Email:  user.Email,
				}
				simpleUsers = append(simpleUsers, cache.SimpleUser{
					ID:        user.UserID,
					Name:      user.Name,
					AvatarURL: "",
				})
			}
			if cacheErr := s.userInfoCache.SetSimpleUsersBatch(ctx, simpleUsers); cacheErr != nil {
				s.logger.Warn("Failed to cache users", zap.Error(cacheErr))
			}
		}
	}

	// Add cached users to result
	for userID, cachedUser := range cachedUsers {
		if _, exists := userMap[userID]; !exists {
			userMap[userID] = &dto.UserInfo{
				UserID: cachedUser.ID,
				Name:   cachedUser.Name,
				Email:  "", // SimpleUser doesn't have email
			}
		}
	}

	return userMap
}

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

	// Fetch user info with caching
	ctx := context.Background()
	userInfo, err := s.getUserInfoWithCache(ctx, project.OwnerID.String())
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

	// Fetch user info with caching
	ctx := context.Background()
	userInfo, err := s.getUserInfoWithCache(ctx, member.UserID.String())
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

	// Fetch user info with caching
	ctx := context.Background()
	userInfo, err := s.getUserInfoWithCache(ctx, req.UserID.String())
	if err != nil {
		s.logger.Warn("Failed to fetch user info", zap.Error(err), zap.String("user_id", req.UserID.String()))
	} else {
		response.UserName = userInfo.Name
		response.UserEmail = userInfo.Email
	}

	return response, nil
}
