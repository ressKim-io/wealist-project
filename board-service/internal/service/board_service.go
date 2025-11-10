package service

import (
	"board-service/internal/apperrors"
	"board-service/internal/cache"
	"board-service/internal/client"
	"board-service/internal/domain"
	"board-service/internal/dto"
	"board-service/internal/repository"
	"board-service/internal/util"
	"context"
	"encoding/json"
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
	MoveBoard(userID, boardID string, req *dto.MoveBoardRequest) (*dto.MoveBoardResponse, error)
}

type boardService struct {
	repo          repository.BoardRepository
	projectRepo   repository.ProjectRepository
	roleRepo      repository.RoleRepository
	fieldRepo     repository.FieldRepository // For custom fields system
	userClient    client.UserClient
	userInfoCache cache.UserInfoCache
	logger        *zap.Logger
	db            *gorm.DB
}

func NewBoardService(
	repo repository.BoardRepository,
	projectRepo repository.ProjectRepository,
	roleRepo repository.RoleRepository,
	fieldRepo repository.FieldRepository,
	userClient client.UserClient,
	userInfoCache cache.UserInfoCache,
	logger *zap.Logger,
	db *gorm.DB,
) BoardService {
	return &boardService{
		repo:          repo,
		projectRepo:   projectRepo,
		roleRepo:      roleRepo,
		fieldRepo:     fieldRepo,
		userClient:    userClient,
		userInfoCache: userInfoCache,
		logger:        logger,
		db:            db,
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

	// 2. Validate Assignee (optional)
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

	// 3. Parse DueDate (optional)
	var dueDate *time.Time
	if req.DueDate != nil {
		parsed, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 마감일 형식입니다 (ISO 8601 required)", 400)
		}
		dueDate = &parsed
	}

	// 4. Create Board
	board := &domain.Board{
		ProjectID:         projectUUID,
		Title:             req.Title,
		Description:       req.Content,
		AssigneeID:        assigneeUUID,
		CreatedBy:         userUUID,
		DueDate:           dueDate,
		CustomFieldsCache: "{}",  // Initialize empty, use FieldValueService to set values
	}

	err = s.repo.Create(board)
	if err != nil {
		s.logger.Error("Failed to create board", zap.Error(err))
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "칸반 생성 실패", 500)
	}

	// Note: Custom field values (stage, role, importance) should be set via FieldValueService
	// after board creation using /field-values API

	// 5. Build response
	return s.buildBoardResponse(board)
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

	// 3. Build response
	// Note: Custom field values are now in custom_fields_cache (JSONB)
	// Frontend should fetch field definitions and parse custom_fields_cache
	return s.buildBoardResponse(board)
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
	// Note: Custom field filtering (stage, role, importance, etc.) is now done
	// via ViewService using JSONB queries on custom_fields_cache column
	filters := repository.BoardFilters{}
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

	// 5. Collect user IDs for batch queries
	userIDs := make([]string, 0, len(boards)*2)
	for _, board := range boards {
		userIDs = append(userIDs, board.CreatedBy.String())
		if board.AssigneeID != nil {
			userIDs = append(userIDs, board.AssigneeID.String())
		}
	}

	// 6. Batch fetch users
	userMap := s.getUserInfoBatch(ctx, userIDs)

	// 7. Build responses
	// Note: Custom field values are now in custom_fields_cache (JSONB)
	responses := make([]dto.BoardResponse, 0, len(boards))
	for _, board := range boards {
		response, err := s.buildBoardResponseOptimized(&board, userMap)
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

	// Note: Stage, Importance, and Role updates should now be done via FieldValueService
	// using /field-values API endpoints

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

	// 4. Save board
	if err := s.repo.Update(board); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 수정 실패", 500)
	}

	// 5. Return updated board
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

func (s *boardService) buildBoardResponse(board *domain.Board) (*dto.BoardResponse, error) {
	// Collect user IDs for batch query
	userIDs := []string{board.CreatedBy.String()}
	if board.AssigneeID != nil {
		userIDs = append(userIDs, board.AssigneeID.String())
	}

	// Fetch users with caching
	ctx := context.Background()
	userMap := s.getUserInfoBatch(ctx, userIDs)

	// Parse custom_fields_cache
	var customFields map[string]interface{}
	if board.CustomFieldsCache != "" && board.CustomFieldsCache != "{}" {
		if err := json.Unmarshal([]byte(board.CustomFieldsCache), &customFields); err != nil {
			s.logger.Warn("Failed to parse custom_fields_cache", zap.Error(err), zap.String("board_id", board.ID.String()))
			customFields = make(map[string]interface{})
		}
	} else {
		customFields = make(map[string]interface{})
	}

	// Build response
	response := &dto.BoardResponse{
		ID:           board.ID.String(),
		ProjectID:    board.ProjectID.String(),
		Title:        board.Title,
		Content:      board.Description,
		CustomFields: customFields,  // Include JSONB custom fields
		DueDate:      board.DueDate,
		CreatedAt:    board.CreatedAt,
		UpdatedAt:    board.UpdatedAt,
	}

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
	userMap map[string]client.UserInfo,
) (*dto.BoardResponse, error) {
	// Parse custom_fields_cache
	var customFields map[string]interface{}
	if board.CustomFieldsCache != "" && board.CustomFieldsCache != "{}" {
		if err := json.Unmarshal([]byte(board.CustomFieldsCache), &customFields); err != nil {
			s.logger.Warn("Failed to parse custom_fields_cache", zap.Error(err), zap.String("board_id", board.ID.String()))
			customFields = make(map[string]interface{})
		}
	} else {
		customFields = make(map[string]interface{})
	}

	// Build response
	response := &dto.BoardResponse{
		ID:           board.ID.String(),
		ProjectID:    board.ProjectID.String(),
		Title:        board.Title,
		Content:      board.Description,
		CustomFields: customFields,
		DueDate:      board.DueDate,
		CreatedAt:    board.CreatedAt,
		UpdatedAt:    board.UpdatedAt,
	}

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

// ==================== Move Board (Integrated API) ====================

// MoveBoard moves a board to a different column/group in a view
// This API combines field value change + position update in a single transaction
// Uses fractional indexing for O(1) operations - only 1 row updated!
func (s *boardService) MoveBoard(userID, boardID string, req *dto.MoveBoardRequest) (*dto.MoveBoardResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	boardUUID, err := uuid.Parse(boardID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 보드 ID", 400)
	}

	viewUUID, err := uuid.Parse(req.ViewID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 뷰 ID", 400)
	}

	fieldUUID, err := uuid.Parse(req.GroupByFieldID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 필드 ID", 400)
	}

	newValueUUID, err := uuid.Parse(req.NewFieldValue)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 필드 값 ID", 400)
	}

	// 1. Fetch board
	board, err := s.repo.FindByID(boardUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "보드를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 조회 실패", 500)
	}

	// 2. Check project membership
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, board.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// 3. Fetch field to validate
	field, err := s.fieldRepo.FindFieldByID(fieldUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "필드를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 조회 실패", 500)
	}

	// 4. Validate field belongs to board's project
	if field.ProjectID != board.ProjectID {
		return nil, apperrors.New(apperrors.ErrCodeBadRequest, "필드가 보드의 프로젝트에 속하지 않습니다", 400)
	}

	// 5. Validate field type (only single_select and multi_select supported for grouping)
	if field.FieldType != domain.FieldTypeSingleSelect && field.FieldType != domain.FieldTypeMultiSelect {
		return nil, apperrors.New(apperrors.ErrCodeBadRequest, "Single-select 또는 Multi-select 필드만 그룹핑에 사용할 수 있습니다", 400)
	}

	// 6. Validate option exists
	option, err := s.fieldRepo.FindOptionByID(newValueUUID)
	if err != nil || option.FieldID != fieldUUID {
		return nil, apperrors.New(apperrors.ErrCodeBadRequest, "유효하지 않은 옵션입니다", 400)
	}

	// 7. Generate new position using fractional indexing
	var beforePos, afterPos string
	if req.BeforePosition != nil {
		beforePos = *req.BeforePosition
	}
	if req.AfterPosition != nil {
		afterPos = *req.AfterPosition
	}

	// Import util package for fractional indexing
	newPosition := util.GeneratePositionBetween(beforePos, afterPos)

	// 8. Execute in transaction
	var finalPosition string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 8-1. Update field value (change column)
		// Delete old value first
		if err := s.fieldRepo.BatchDeleteFieldValues(boardUUID, fieldUUID); err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "기존 필드 값 삭제 실패", 500)
		}

		// Set new value
		newFieldValue := &domain.BoardFieldValue{
			BoardID:       boardUUID,
			FieldID:       fieldUUID,
			ValueOptionID: &newValueUUID,
			DisplayOrder:  0,
		}
		if err := s.fieldRepo.SetFieldValue(newFieldValue); err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 값 설정 실패", 500)
		}

		// 8-2. Update board position (fractional indexing - only 1 row!)
		boardOrder := domain.UserBoardOrder{
			ViewID:   viewUUID,
			UserID:   userUUID,
			BoardID:  boardUUID,
			Position: newPosition,
		}
		if err := s.fieldRepo.SetBoardOrder(&boardOrder); err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 순서 업데이트 실패", 500)
		}

		finalPosition = newPosition

		// 8-3. Update JSONB cache
		if _, err := s.fieldRepo.UpdateBoardFieldCache(boardUUID); err != nil {
			s.logger.Warn("Failed to update board cache", zap.Error(err))
		}

		return nil
	})

	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			return nil, appErr
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 이동 실패", 500)
	}

	return &dto.MoveBoardResponse{
		BoardID:       boardID,
		NewFieldValue: req.NewFieldValue,
		NewPosition:   finalPosition,
		Message:       "보드가 성공적으로 이동되었습니다 (O(1) 연산)",
	}, nil
}
