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

type BoardService interface {
	CreateBoard(userID string, req *dto.CreateBoardRequest) (*dto.BoardResponse, error)
	GetBoard(boardID, userID string) (*dto.BoardResponse, error)
	GetBoards(userID string, req *dto.GetBoardsRequest) (*dto.PaginatedBoardsResponse, error)
	UpdateBoard(boardID, userID string, req *dto.UpdateBoardRequest) (*dto.BoardResponse, error)
	DeleteBoard(boardID, userID string) error
}

type boardService struct {
	repo            repository.BoardRepository
	projectRepo     repository.ProjectRepository
	customFieldRepo repository.CustomFieldRepository
	roleRepo        repository.RoleRepository
	userClient      client.UserClient
	userInfoCache   cache.UserInfoCache
	logger          *zap.Logger
	db              *gorm.DB
}

func NewBoardService(
	repo repository.BoardRepository,
	projectRepo repository.ProjectRepository,
	customFieldRepo repository.CustomFieldRepository,
	roleRepo repository.RoleRepository,
	userClient client.UserClient,
	userInfoCache cache.UserInfoCache,
	logger *zap.Logger,
	db *gorm.DB,
) BoardService {
	return &boardService{
		repo:            repo,
		projectRepo:     projectRepo,
		customFieldRepo: customFieldRepo,
		roleRepo:        roleRepo,
		userClient:      userClient,
		userInfoCache:   userInfoCache,
		logger:          logger,
		db:              db,
	}
}

// ==================== Create Board ====================

func (s *boardService) CreateBoard(userID string, req *dto.CreateBoardRequest) (*dto.BoardResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	// 1. Check if user is project member
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, projectUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// 2. Validate Stage (required)
	stageUUID, err := uuid.Parse(req.StageID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 진행단계 ID", 400)
	}

	stage, err := s.customFieldRepo.FindCustomStageByID(stageUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "진행단계를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "진행단계 조회 실패", 500)
	}

	if stage.ProjectID != projectUUID {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "다른 프로젝트의 진행단계입니다", 403)
	}

	// 3. Validate Importance (optional)
	var importance *domain.CustomImportance
	var importanceUUID *uuid.UUID
	if req.ImportanceID != nil {
		parsedImportanceUUID, err := uuid.Parse(*req.ImportanceID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 중요도 ID", 400)
		}
		importanceUUID = &parsedImportanceUUID

		importance, err = s.customFieldRepo.FindCustomImportanceByID(parsedImportanceUUID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.New(apperrors.ErrCodeNotFound, "중요도를 찾을 수 없습니다", 404)
			}
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "중요도 조회 실패", 500)
		}

		if importance.ProjectID != projectUUID {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "다른 프로젝트의 중요도입니다", 403)
		}
	}

	// 4. Validate Roles (required, at least 1)
	roleUUIDs := make([]uuid.UUID, 0, len(req.RoleIDs))
	roles := make([]*domain.CustomRole, 0, len(req.RoleIDs))
	for _, roleID := range req.RoleIDs {
		roleUUID, err := uuid.Parse(roleID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 역할 ID", 400)
		}

		role, err := s.customFieldRepo.FindCustomRoleByID(roleUUID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.New(apperrors.ErrCodeNotFound, "역할을 찾을 수 없습니다", 404)
			}
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 조회 실패", 500)
		}

		if role.ProjectID != projectUUID {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "다른 프로젝트의 역할입니다", 403)
		}

		roleUUIDs = append(roleUUIDs, roleUUID)
		roles = append(roles, role)
	}

	// 5. Validate Assignee (optional)
	var assigneeUUID *uuid.UUID
	if req.AssigneeID != nil {
		parsedAssigneeUUID, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 담당자 ID", 400)
		}
		assigneeUUID = &parsedAssigneeUUID

		_, err = s.projectRepo.FindMemberByUserAndProject(parsedAssigneeUUID, projectUUID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.New(apperrors.ErrCodeNotFound, "담당자가 프로젝트 멤버가 아닙니다", 404)
			}
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "담당자 확인 실패", 500)
		}
	}

	// 6. Parse DueDate (optional)
	var dueDate *time.Time
	if req.DueDate != nil {
		parsed, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 마감일 형식입니다 (ISO 8601 required)", 400)
		}
		dueDate = &parsed
	}

	// 7. Create Board in transaction
	var board *domain.Board
	err = s.db.Transaction(func(tx *gorm.DB) error {
		board = &domain.Board{
			ProjectID:          projectUUID,
			Title:              req.Title,
			Description:        req.Content,
			CustomStageID:      stageUUID,
			CustomImportanceID: importanceUUID,
			AssigneeID:         assigneeUUID,
			CreatedBy:          userUUID,
			DueDate:            dueDate,
		}

		if err := s.repo.Create(board); err != nil {
			s.logger.Error("Failed to create board", zap.Error(err))
			return err
		}

		// Create board_roles (many-to-many)
		if err := s.repo.CreateBoardRoles(board.ID, roleUUIDs); err != nil {
			s.logger.Error("Failed to create board roles", zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 생성 실패", 500)
	}

	// 8. Build response with user info
	return s.buildBoardResponse(board, stage, importance, roles)
}

// ==================== Get Single Board ====================

func (s *boardService) GetBoard(boardID, userID string) (*dto.BoardResponse, error) {
	boardUUID, err := uuid.Parse(boardID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 보드 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// 1. Find board
	board, err := s.repo.FindByID(boardUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "보드을 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 조회 실패", 500)
	}

	// 2. Check if user is project member
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, board.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// 3. Fetch related data
	stage, err := s.customFieldRepo.FindCustomStageByID(board.CustomStageID)
	if err != nil {
		s.logger.Warn("Failed to fetch stage", zap.Error(err), zap.String("stage_id", board.CustomStageID.String()))
	}

	var importance *domain.CustomImportance
	if board.CustomImportanceID != nil {
		importance, err = s.customFieldRepo.FindCustomImportanceByID(*board.CustomImportanceID)
		if err != nil {
			s.logger.Warn("Failed to fetch importance", zap.Error(err), zap.String("importance_id", board.CustomImportanceID.String()))
		}
	}

	boardRoles, err := s.repo.FindRolesByBoard(board.ID)
	if err != nil {
		s.logger.Warn("Failed to fetch board roles", zap.Error(err))
	}

	roles := make([]*domain.CustomRole, 0, len(boardRoles))
	for _, kr := range boardRoles {
		role, err := s.customFieldRepo.FindCustomRoleByID(kr.CustomRoleID)
		if err == nil && role != nil {
			roles = append(roles, role)
		}
	}

	// 4. Build response
	return s.buildBoardResponse(board, stage, importance, roles)
}

// ==================== Get Boards (List with Filters) ====================

func (s *boardService) GetBoards(userID string, req *dto.GetBoardsRequest) (*dto.PaginatedBoardsResponse, error) {
	projectUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	ctx := context.Background()

	// 1. Check if user is project member
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, projectUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// 2. Build filters
	filters := repository.BoardFilters{}
	if req.StageID != "" {
		stageUUID, err := uuid.Parse(req.StageID)
		if err == nil {
			filters.StageID = stageUUID
		}
	}
	if req.RoleID != "" {
		roleUUID, err := uuid.Parse(req.RoleID)
		if err == nil {
			filters.RoleID = roleUUID
		}
	}
	if req.ImportanceID != "" {
		importanceUUID, err := uuid.Parse(req.ImportanceID)
		if err == nil {
			filters.ImportanceID = importanceUUID
		}
	}
	if req.AssigneeID != "" {
		assigneeUUID, err := uuid.Parse(req.AssigneeID)
		if err == nil {
			filters.AssigneeID = assigneeUUID
		}
	}
	if req.AuthorID != "" {
		authorUUID, err := uuid.Parse(req.AuthorID)
		if err == nil {
			filters.AuthorID = authorUUID
		}
	}

	// 3. Default pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}

	// 4. Fetch boards
	boards, total, err := s.repo.FindByProject(projectUUID, filters, page, limit)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 조회 실패", 500)
	}

	if len(boards) == 0 {
		return &dto.PaginatedBoardsResponse{
			Boards: []dto.BoardResponse{},
			Total:  total,
			Page:   page,
			Limit:  limit,
		}, nil
	}

	// 5. Collect all IDs for batch queries
	stageIDs := make([]uuid.UUID, 0, len(boards))
	importanceIDs := make([]uuid.UUID, 0)
	userIDs := make([]string, 0, len(boards)*2)

	for _, board := range boards {
		stageIDs = append(stageIDs, board.CustomStageID)
		if board.CustomImportanceID != nil {
			importanceIDs = append(importanceIDs, *board.CustomImportanceID)
		}
		userIDs = append(userIDs, board.CreatedBy.String())
		if board.AssigneeID != nil {
			userIDs = append(userIDs, board.AssigneeID.String())
		}
	}

	// 6. Batch fetch custom fields
	stagesSlice, _ := s.customFieldRepo.FindCustomStagesByIDs(stageIDs)
	stagesMap := make(map[uuid.UUID]*domain.CustomStage)
	for i := range stagesSlice {
		stagesMap[stagesSlice[i].ID] = &stagesSlice[i]
	}

	importancesSlice, _ := s.customFieldRepo.FindCustomImportancesByIDs(importanceIDs)
	importancesMap := make(map[uuid.UUID]*domain.CustomImportance)
	for i := range importancesSlice {
		importancesMap[importancesSlice[i].ID] = &importancesSlice[i]
	}

	// 7. Batch fetch users
	userMap := s.getUserInfoBatch(ctx, userIDs)

	// 8. Batch fetch board roles for all boards
	boardRolesMap := make(map[uuid.UUID][]*domain.CustomRole)
	allRoleIDs := make([]uuid.UUID, 0)
	boardToRoleIDs := make(map[uuid.UUID][]uuid.UUID)

	// Collect all board IDs
	boardIDs := make([]uuid.UUID, 0, len(boards))
	for _, board := range boards {
		boardIDs = append(boardIDs, board.ID)
	}

	// Batch fetch board roles (1 query instead of N)
	boardRolesData, _ := s.repo.FindRolesByBoards(boardIDs)

	// Process board roles
	for boardID, boardRoles := range boardRolesData {
		if len(boardRoles) > 0 {
			roleIDs := make([]uuid.UUID, 0, len(boardRoles))
			for _, kr := range boardRoles {
				roleIDs = append(roleIDs, kr.CustomRoleID)
				allRoleIDs = append(allRoleIDs, kr.CustomRoleID)
			}
			boardToRoleIDs[boardID] = roleIDs
		}
	}

	// Batch fetch all roles at once
	if len(allRoleIDs) > 0 {
		rolesSlice, _ := s.customFieldRepo.FindCustomRolesByIDs(allRoleIDs)
		rolesMapByID := make(map[uuid.UUID]*domain.CustomRole)
		for i := range rolesSlice {
			rolesMapByID[rolesSlice[i].ID] = &rolesSlice[i]
		}

		// Map roles to boards
		for boardID, roleIDs := range boardToRoleIDs {
			roles := make([]*domain.CustomRole, 0, len(roleIDs))
			for _, roleID := range roleIDs {
				if role, ok := rolesMapByID[roleID]; ok {
					roles = append(roles, role)
				}
			}
			boardRolesMap[boardID] = roles
		}
	}

	// 9. Build responses
	responses := make([]dto.BoardResponse, 0, len(boards))
	for _, board := range boards {
		stage := stagesMap[board.CustomStageID]
		var importance *domain.CustomImportance
		if board.CustomImportanceID != nil {
			importance = importancesMap[*board.CustomImportanceID]
		}
		roles := boardRolesMap[board.ID]

		response, err := s.buildBoardResponseOptimized(&board, stage, importance, roles, userMap)
		if err == nil && response != nil {
			responses = append(responses, *response)
		}
	}

	return &dto.PaginatedBoardsResponse{
		Boards: responses,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}, nil
}

// ==================== Update Board ====================

func (s *boardService) UpdateBoard(boardID, userID string, req *dto.UpdateBoardRequest) (*dto.BoardResponse, error) {
	boardUUID, err := uuid.Parse(boardID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 보드 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// 1. Find board
	board, err := s.repo.FindByID(boardUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "보드을 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 조회 실패", 500)
	}

	// 2. Check permission (author or ADMIN+)
	member, err := s.projectRepo.FindMemberByUserAndProject(userUUID, board.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// Get member role
	role, err := s.roleRepo.FindByID(member.RoleID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
	}

	// Check if user is author or has ADMIN+ permission
	if board.CreatedBy != userUUID && role.Name == "MEMBER" {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "수정 권한이 없습니다", 403)
	}

	// 3. Update fields
	if req.Title != "" {
		board.Title = req.Title
	}
	if req.Content != "" {
		board.Description = req.Content
	}

	if req.StageID != "" {
		stageUUID, err := uuid.Parse(req.StageID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 진행단계 ID", 400)
		}

		stage, err := s.customFieldRepo.FindCustomStageByID(stageUUID)
		if err != nil || stage.ProjectID != board.ProjectID {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "진행단계를 찾을 수 없습니다", 404)
		}
		board.CustomStageID = stageUUID
	}

	if req.ImportanceID != nil {
		importanceUUID, err := uuid.Parse(*req.ImportanceID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 중요도 ID", 400)
		}

		importance, err := s.customFieldRepo.FindCustomImportanceByID(importanceUUID)
		if err != nil || importance.ProjectID != board.ProjectID {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "중요도를 찾을 수 없습니다", 404)
		}
		board.CustomImportanceID = &importanceUUID
	}

	if req.AssigneeID != nil {
		assigneeUUID, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 담당자 ID", 400)
		}

		_, err = s.projectRepo.FindMemberByUserAndProject(assigneeUUID, board.ProjectID)
		if err != nil {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "담당자가 프로젝트 멤버가 아닙니다", 404)
		}
		board.AssigneeID = &assigneeUUID
	}

	if req.DueDate != nil {
		parsed, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 마감일 형식입니다 (ISO 8601 required)", 400)
		}
		board.DueDate = &parsed
	}

	// 4. Update roles if provided
	if len(req.RoleIDs) > 0 {
		// Validate all roles first
		roleUUIDs := make([]uuid.UUID, 0, len(req.RoleIDs))
		for _, roleID := range req.RoleIDs {
			roleUUID, err := uuid.Parse(roleID)
			if err != nil {
				return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 역할 ID", 400)
			}

			role, err := s.customFieldRepo.FindCustomRoleByID(roleUUID)
			if err != nil || role.ProjectID != board.ProjectID {
				return nil, apperrors.New(apperrors.ErrCodeNotFound, "역할을 찾을 수 없습니다", 404)
			}

			roleUUIDs = append(roleUUIDs, roleUUID)
		}

		// Delete existing roles and create new ones in transaction
		err = s.db.Transaction(func(tx *gorm.DB) error {
			if err := s.repo.DeleteBoardRolesByBoard(board.ID); err != nil {
				return err
			}
			if err := s.repo.CreateBoardRoles(board.ID, roleUUIDs); err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 업데이트 실패", 500)
		}
	}

	// 5. Save board
	if err := s.repo.Update(board); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 수정 실패", 500)
	}

	// 6. Return updated board
	return s.GetBoard(board.ID.String(), userID)
}

// ==================== Delete Board (Soft) ====================

func (s *boardService) DeleteBoard(boardID, userID string) error {
	boardUUID, err := uuid.Parse(boardID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 보드 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// 1. Find board
	board, err := s.repo.FindByID(boardUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "보드을 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 조회 실패", 500)
	}

	// 2. Check permission (author or ADMIN+)
	member, err := s.projectRepo.FindMemberByUserAndProject(userUUID, board.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// Get member role
	role, err := s.roleRepo.FindByID(member.RoleID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
	}

	// Check if user is author or has ADMIN+ permission
	if board.CreatedBy != userUUID && role.Name == "MEMBER" {
		return apperrors.New(apperrors.ErrCodeForbidden, "삭제 권한이 없습니다", 403)
	}

	// 3. Soft delete
	if err := s.repo.Delete(board.ID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 삭제 실패", 500)
	}

	return nil
}

// ==================== Helper: Build Board Response ====================

func (s *boardService) buildBoardResponse(
	board *domain.Board,
	stage *domain.CustomStage,
	importance *domain.CustomImportance,
	roles []*domain.CustomRole,
) (*dto.BoardResponse, error) {
	// Collect user IDs for batch query
	userIDs := []string{board.CreatedBy.String()}
	if board.AssigneeID != nil {
		userIDs = append(userIDs, board.AssigneeID.String())
	}

	// Fetch users with caching
	ctx := context.Background()
	userMap := s.getUserInfoBatch(ctx, userIDs)

	// Build response
	response := &dto.BoardResponse{
		ID:        board.ID.String(),
		ProjectID: board.ProjectID.String(),
		Title:     board.Title,
		Content:   board.Description,
		DueDate:   board.DueDate,
		CreatedAt: board.CreatedAt,
		UpdatedAt: board.UpdatedAt,
	}

	// Stage
	if stage != nil {
		response.Stage = dto.CustomStageResponse{
			ID:              stage.ID.String(),
			ProjectID:       stage.ProjectID.String(),
			Name:            stage.Name,
			Color:           stage.Color,
			IsSystemDefault: stage.IsSystemDefault,
			DisplayOrder:    stage.DisplayOrder,
			CreatedAt:       stage.CreatedAt,
			UpdatedAt:       stage.UpdatedAt,
		}
	}

	// Importance
	if importance != nil {
		response.Importance = &dto.CustomImportanceResponse{
			ID:              importance.ID.String(),
			ProjectID:       importance.ProjectID.String(),
			Name:            importance.Name,
			Color:           importance.Color,
			IsSystemDefault: importance.IsSystemDefault,
			DisplayOrder:    importance.DisplayOrder,
			CreatedAt:       importance.CreatedAt,
			UpdatedAt:       importance.UpdatedAt,
		}
	}

	// Roles
	roleResponses := make([]dto.CustomRoleResponse, 0, len(roles))
	for _, role := range roles {
		roleResponses = append(roleResponses, dto.CustomRoleResponse{
			ID:              role.ID.String(),
			ProjectID:       role.ProjectID.String(),
			Name:            role.Name,
			Color:           role.Color,
			IsSystemDefault: role.IsSystemDefault,
			DisplayOrder:    role.DisplayOrder,
			CreatedAt:       role.CreatedAt,
			UpdatedAt:       role.UpdatedAt,
		})
	}
	response.Roles = roleResponses

	// Author
	if author, ok := userMap[board.CreatedBy.String()]; ok {
		response.Author = dto.UserInfo{
			UserID:   author.UserID,
			Name:     author.Name,
			Email:    author.Email,
			IsActive: author.IsActive,
		}
	} else {
		// Fallback if user not found
		response.Author = dto.UserInfo{
			UserID:   board.CreatedBy.String(),
			Name:     "Unknown User",
			Email:    "",
			IsActive: false,
		}
	}

	// Assignee
	if board.AssigneeID != nil {
		if assignee, ok := userMap[board.AssigneeID.String()]; ok {
			response.Assignee = &dto.UserInfo{
				UserID:   assignee.UserID,
				Name:     assignee.Name,
				Email:    assignee.Email,
				IsActive: assignee.IsActive,
			}
		} else {
			// Fallback if user not found
			response.Assignee = &dto.UserInfo{
				UserID:   board.AssigneeID.String(),
				Name:     "Unknown User",
				Email:    "",
				IsActive: false,
			}
		}
	}

	return response, nil
}

// buildBoardResponseOptimized builds a board response using pre-fetched data (batch optimized)
func (s *boardService) buildBoardResponseOptimized(
	board *domain.Board,
	stage *domain.CustomStage,
	importance *domain.CustomImportance,
	roles []*domain.CustomRole,
	userMap map[string]client.UserInfo,
) (*dto.BoardResponse, error) {
	// Build response
	response := &dto.BoardResponse{
		ID:        board.ID.String(),
		ProjectID: board.ProjectID.String(),
		Title:     board.Title,
		Content:   board.Description,
		DueDate:   board.DueDate,
		CreatedAt: board.CreatedAt,
		UpdatedAt: board.UpdatedAt,
	}

	// Stage
	if stage != nil {
		response.Stage = dto.CustomStageResponse{
			ID:              stage.ID.String(),
			ProjectID:       stage.ProjectID.String(),
			Name:            stage.Name,
			Color:           stage.Color,
			IsSystemDefault: stage.IsSystemDefault,
			DisplayOrder:    stage.DisplayOrder,
			CreatedAt:       stage.CreatedAt,
			UpdatedAt:       stage.UpdatedAt,
		}
	}

	// Importance
	if importance != nil {
		response.Importance = &dto.CustomImportanceResponse{
			ID:              importance.ID.String(),
			ProjectID:       importance.ProjectID.String(),
			Name:            importance.Name,
			Color:           importance.Color,
			IsSystemDefault: importance.IsSystemDefault,
			DisplayOrder:    importance.DisplayOrder,
			CreatedAt:       importance.CreatedAt,
			UpdatedAt:       importance.UpdatedAt,
		}
	}

	// Roles
	roleResponses := make([]dto.CustomRoleResponse, 0, len(roles))
	for _, role := range roles {
		roleResponses = append(roleResponses, dto.CustomRoleResponse{
			ID:              role.ID.String(),
			ProjectID:       role.ProjectID.String(),
			Name:            role.Name,
			Color:           role.Color,
			IsSystemDefault: role.IsSystemDefault,
			DisplayOrder:    role.DisplayOrder,
			CreatedAt:       role.CreatedAt,
			UpdatedAt:       role.UpdatedAt,
		})
	}
	response.Roles = roleResponses

	// Author (from userMap)
	if author, ok := userMap[board.CreatedBy.String()]; ok {
		response.Author = dto.UserInfo{
			UserID:   author.UserID,
			Name:     author.Name,
			Email:    author.Email,
			IsActive: author.IsActive,
		}
	} else {
		// Fallback if user not found
		response.Author = dto.UserInfo{
			UserID:   board.CreatedBy.String(),
			Name:     "Unknown User",
			Email:    "",
			IsActive: false,
		}
	}

	// Assignee (from userMap)
	if board.AssigneeID != nil {
		if assignee, ok := userMap[board.AssigneeID.String()]; ok {
			response.Assignee = &dto.UserInfo{
				UserID:   assignee.UserID,
				Name:     assignee.Name,
				Email:    assignee.Email,
				IsActive: assignee.IsActive,
			}
		} else {
			// Fallback if user not found
			response.Assignee = &dto.UserInfo{
				UserID:   board.AssigneeID.String(),
				Name:     "Unknown User",
				Email:    "",
				IsActive: false,
			}
		}
	}

	return response, nil
}

// getUserInfoBatch fetches user info for multiple users with caching
func (s *boardService) getUserInfoBatch(ctx context.Context, userIDs []string) map[string]client.UserInfo {
	if len(userIDs) == 0 {
		return make(map[string]client.UserInfo)
	}

	// Try to get from cache first
	cachedUsers, err := s.userInfoCache.GetSimpleUsersBatch(ctx, userIDs)
	if err != nil {
		s.logger.Warn("Failed to get users from cache", zap.Error(err))
		cachedUsers = make(map[string]*cache.SimpleUser)
	}

	// Find missing user IDs (not in cache)
	missingUserIDs := []string{}
	for _, userID := range userIDs {
		if _, exists := cachedUsers[userID]; !exists {
			missingUserIDs = append(missingUserIDs, userID)
		}
	}

	// Fetch missing users from User Service
	userMap := make(map[string]client.UserInfo)

	if len(missingUserIDs) > 0 {
		users, err := s.userClient.GetUsersBatch(ctx, missingUserIDs)
		if err != nil {
			s.logger.Warn("Failed to fetch users from User Service", zap.Error(err))
		} else {
			// Cache the fetched users
			simpleUsers := make([]cache.SimpleUser, 0, len(users))
			for _, user := range users {
				userMap[user.UserID] = user
				// Note: UserInfo and SimpleUser have different fields
				// For now, we'll just cache what we got
				simpleUsers = append(simpleUsers, cache.SimpleUser{
					ID:        user.UserID,
					Name:      user.Name,
					AvatarURL: "", // UserInfo doesn't have avatar URL
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
			userMap[userID] = client.UserInfo{
				UserID:   cachedUser.ID,
				Name:     cachedUser.Name,
				Email:    "", // SimpleUser doesn't have email
				IsActive: true,
			}
		}
	}

	return userMap
}
