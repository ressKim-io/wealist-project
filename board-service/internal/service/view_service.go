package service

import (
	"board-service/internal/apperrors"
	"board-service/internal/cache"
	"board-service/internal/domain"
	"board-service/internal/dto"
	"board-service/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ViewService interface {
	// View CRUD
	CreateView(userID string, req *dto.CreateViewRequest) (*dto.ViewResponse, error)
	GetViewsByProject(userID, projectID string) ([]dto.ViewResponse, error)
	GetView(userID, viewID string) (*dto.ViewResponse, error)
	UpdateView(userID, viewID string, req *dto.UpdateViewRequest) (*dto.ViewResponse, error)
	DeleteView(userID, viewID string) error

	// Apply view (filter + sort + group)
	ApplyView(userID, viewID string, page, limit int) (interface{}, error)
	ApplyViewWithFilters(userID, projectID, viewID string, filters map[string]interface{}, sortBy, sortDir string, groupByFieldID *string, page, limit int) (interface{}, error)

	// Board order management
	UpdateBoardOrder(userID string, req *dto.UpdateBoardOrderRequest) error
}

type viewService struct {
	repo        repository.FieldRepository
	boardRepo   repository.BoardRepository
	projectRepo repository.ProjectRepository
	cache       cache.FieldCache
	logger      *zap.Logger
	db          *gorm.DB
}

func NewViewService(
	repo repository.FieldRepository,
	boardRepo repository.BoardRepository,
	projectRepo repository.ProjectRepository,
	cache cache.FieldCache,
	logger *zap.Logger,
	db *gorm.DB,
) ViewService {
	return &viewService{
		repo:        repo,
		boardRepo:   boardRepo,
		projectRepo: projectRepo,
		cache:       cache,
		logger:      logger,
		db:          db,
	}
}

// ==================== View CRUD ====================

func (s *viewService) CreateView(userID string, req *dto.CreateViewRequest) (*dto.ViewResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	// Check project membership
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, projectUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// Serialize filters
	filtersJSON, err := json.Marshal(req.Filters)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "필터가 유효하지 않습니다", 400)
	}

	// Validate group by field if specified
	var groupByFieldID *uuid.UUID
	if req.GroupByFieldID != "" {
		fieldUUID, err := uuid.Parse(req.GroupByFieldID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 그룹핑 필드 ID", 400)
		}

		// Verify field exists and is single-select or multi-select
		field, err := s.repo.FindFieldByID(fieldUUID)
		if err != nil {
			return nil, apperrors.New(apperrors.ErrCodeBadRequest, "그룹핑 필드를 찾을 수 없습니다", 400)
		}
		if field.FieldType != domain.FieldTypeSingleSelect && field.FieldType != domain.FieldTypeMultiSelect {
			return nil, apperrors.New(apperrors.ErrCodeBadRequest, "Single-select 또는 Multi-select 필드만 그룹핑에 사용할 수 있습니다", 400)
		}
		groupByFieldID = &fieldUUID
	}

	// Create view
	view := &domain.SavedView{
		ProjectID:      projectUUID,
		CreatedBy:      userUUID,
		Name:           req.Name,
		Description:    req.Description,
		IsDefault:      req.IsDefault,
		IsShared:       req.IsShared,
		Filters:        string(filtersJSON),
		SortDirection:  req.SortDirection,
		GroupByFieldID: groupByFieldID,
	}

	if req.SortBy != "" {
		view.SortBy = &req.SortBy
	}

	if err := s.repo.CreateView(view); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "뷰 생성 실패", 500)
	}

	return s.buildViewResponse(view), nil
}

func (s *viewService) GetViewsByProject(userID, projectID string) ([]dto.ViewResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	// Check project membership
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, projectUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// Fetch views
	views, err := s.repo.FindViewsByProject(projectUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "뷰 조회 실패", 500)
	}

	// Filter: only show shared views or user's own views
	filteredViews := make([]domain.SavedView, 0)
	for _, view := range views {
		if view.IsShared || view.CreatedBy == userUUID {
			filteredViews = append(filteredViews, view)
		}
	}

	responses := make([]dto.ViewResponse, 0, len(filteredViews))
	for i := range filteredViews {
		responses = append(responses, *s.buildViewResponse(&filteredViews[i]))
	}

	return responses, nil
}

func (s *viewService) GetView(userID, viewID string) (*dto.ViewResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	viewUUID, err := uuid.Parse(viewID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 뷰 ID", 400)
	}

	// Fetch view
	view, err := s.repo.FindViewByID(viewUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "뷰를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "뷰 조회 실패", 500)
	}

	// Check access: shared or owner
	if !view.IsShared && view.CreatedBy != userUUID {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "뷰 접근 권한이 없습니다", 403)
	}

	// Check project membership
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, view.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	return s.buildViewResponse(view), nil
}

func (s *viewService) UpdateView(userID, viewID string, req *dto.UpdateViewRequest) (*dto.ViewResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	viewUUID, err := uuid.Parse(viewID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 뷰 ID", 400)
	}

	// Fetch view
	view, err := s.repo.FindViewByID(viewUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "뷰를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "뷰 조회 실패", 500)
	}

	// Only creator can modify
	if view.CreatedBy != userUUID {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "뷰 수정 권한이 없습니다 (작성자만 가능)", 403)
	}

	// Update fields
	if req.Name != "" {
		view.Name = req.Name
	}
	if req.Description != "" {
		view.Description = req.Description
	}
	if req.IsDefault != nil {
		view.IsDefault = *req.IsDefault
	}
	if req.IsShared != nil {
		view.IsShared = *req.IsShared
	}
	if req.Filters != nil {
		filtersJSON, err := json.Marshal(req.Filters)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "필터가 유효하지 않습니다", 400)
		}
		view.Filters = string(filtersJSON)
	}
	if req.SortBy != nil {
		view.SortBy = req.SortBy
	}
	if req.SortDirection != "" {
		view.SortDirection = req.SortDirection
	}
	if req.GroupByFieldID != nil {
		if *req.GroupByFieldID == "" {
			view.GroupByFieldID = nil
		} else {
			fieldUUID, err := uuid.Parse(*req.GroupByFieldID)
			if err != nil {
				return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 그룹핑 필드 ID", 400)
			}
			view.GroupByFieldID = &fieldUUID
		}
	}

	if err := s.repo.UpdateView(view); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "뷰 수정 실패", 500)
	}

	return s.buildViewResponse(view), nil
}

func (s *viewService) DeleteView(userID, viewID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	viewUUID, err := uuid.Parse(viewID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 뷰 ID", 400)
	}

	// Fetch view
	view, err := s.repo.FindViewByID(viewUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "뷰를 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "뷰 조회 실패", 500)
	}

	// Only creator can delete
	if view.CreatedBy != userUUID {
		return apperrors.New(apperrors.ErrCodeForbidden, "뷰 삭제 권한이 없습니다 (작성자만 가능)", 403)
	}

	if err := s.repo.DeleteView(viewUUID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "뷰 삭제 실패", 500)
	}

	// Invalidate view results cache
	ctx := context.Background()
	if err := s.cache.InvalidateViewResults(ctx, viewID); err != nil {
		s.logger.Warn("Failed to invalidate view results cache", zap.Error(err))
	}

	return nil
}

// ==================== Apply View (Filter + Sort + Group) ====================

func (s *viewService) ApplyView(userID, viewID string, page, limit int) (interface{}, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	viewUUID, err := uuid.Parse(viewID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 뷰 ID", 400)
	}

	// Fetch view
	view, err := s.repo.FindViewByID(viewUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "뷰를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "뷰 조회 실패", 500)
	}

	// Check access
	if !view.IsShared && view.CreatedBy != userUUID {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "뷰 접근 권한이 없습니다", 403)
	}

	// Check project membership
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, view.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// Parse filters
	var filters map[string]interface{}
	if view.Filters != "" && view.Filters != "{}" {
		if err := json.Unmarshal([]byte(view.Filters), &filters); err != nil {
			s.logger.Warn("Failed to parse view filters", zap.Error(err))
			filters = make(map[string]interface{})
		}
	}

	var sortBy string
	if view.SortBy != nil {
		sortBy = *view.SortBy
	}

	var groupByFieldIDStr *string
	if view.GroupByFieldID != nil {
		str := view.GroupByFieldID.String()
		groupByFieldIDStr = &str
	}

	return s.ApplyViewWithFilters(userID, view.ProjectID.String(), viewUUID.String(), filters, sortBy, view.SortDirection, groupByFieldIDStr, page, limit)
}

func (s *viewService) ApplyViewWithFilters(userID, projectID, viewID string, filters map[string]interface{}, sortBy, sortDir string, groupByFieldID *string, page, limit int) (interface{}, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	viewUUID, err := uuid.Parse(viewID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 뷰 ID", 400)
	}

	// Check project membership
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, projectUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// Build query with filters
	query := s.db.Model(&domain.Board{}).Where("project_id = ? AND is_deleted = ?", projectUUID, false)

	// Apply filters
	for fieldIDStr, filterConfig := range filters {
		// Parse filter condition
		filterMap, ok := filterConfig.(map[string]interface{})
		if !ok {
			continue
		}

		operator, _ := filterMap["operator"].(string)
		value := filterMap["value"]

		// Special handling for built-in fields
		if fieldIDStr == "title" {
			query = s.applyBuiltInFilter(query, "title", operator, value)
			continue
		}

		// Custom field filtering via custom_fields_cache
		fieldUUID, err := uuid.Parse(fieldIDStr)
		if err != nil {
			continue
		}

		query = s.applyCustomFieldFilter(query, fieldUUID, operator, value)
	}

	// Apply sorting
	if sortBy != "" {
		if sortDir == "" {
			sortDir = "asc"
		}
		query = query.Order(fmt.Sprintf("%s %s", sortBy, strings.ToUpper(sortDir)))
	} else {
		query = query.Order("created_at DESC")
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 카운트 실패", 500)
	}

	// Pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Fetch boards
	var boards []domain.Board
	if err := query.Offset(offset).Limit(limit).Find(&boards).Error; err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 조회 실패", 500)
	}

	// If grouping requested, apply grouping
	if groupByFieldID != nil && *groupByFieldID != "" {
		return s.applyGrouping(boards, *groupByFieldID, total)
	}

	// Fetch board positions for this view and user
	boardIDs := make([]uuid.UUID, len(boards))
	for i, board := range boards {
		boardIDs[i] = board.ID
	}

	var userBoardOrders []domain.UserBoardOrder
	if len(boardIDs) > 0 {
		s.db.Where("view_id = ? AND user_id = ? AND board_id IN ?", viewUUID, userUUID, boardIDs).
			Find(&userBoardOrders)
	}

	// Create position map for quick lookup
	positionMap := make(map[uuid.UUID]string)
	for _, order := range userBoardOrders {
		positionMap[order.BoardID] = order.Position
	}

	// Return paginated results
	boardResponses := make([]dto.BoardResponse, 0, len(boards))
	for _, board := range boards {
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

		// Get position from map
		position := positionMap[board.ID]

		// Simplified board response (can be enhanced with full details)
		boardResponses = append(boardResponses, dto.BoardResponse{
			ID:           board.ID.String(),
			ProjectID:    board.ProjectID.String(),
			Title:        board.Title,
			Content:      board.Description,
			CustomFields: customFields,
			Position:     position, // Include position from user_board_order
			CreatedAt:    board.CreatedAt,
			UpdatedAt:    board.UpdatedAt,
		})
	}

	return map[string]interface{}{
		"boards": boardResponses,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}, nil
}

// ==================== Board Order Management ====================

func (s *viewService) UpdateBoardOrder(userID string, req *dto.UpdateBoardOrderRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	viewUUID, err := uuid.Parse(req.ViewID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 뷰 ID", 400)
	}

	// Fetch view to verify access
	view, err := s.repo.FindViewByID(viewUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "뷰를 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "뷰 조회 실패", 500)
	}

	// Check project membership
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, view.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// Build orders
	orders := make([]domain.UserBoardOrder, 0, len(req.BoardOrders))
	for _, item := range req.BoardOrders {
		boardUUID, err := uuid.Parse(item.BoardID)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 보드 ID", 400)
		}

		orders = append(orders, domain.UserBoardOrder{
			ViewID:   viewUUID,
			UserID:   userUUID,
			BoardID:  boardUUID,
			Position: item.Position,
		})
	}

	// Batch update
	if err := s.repo.BatchUpdateBoardOrders(orders); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 순서 업데이트 실패", 500)
	}

	return nil
}

// ==================== Helper Methods ====================

func (s *viewService) buildViewResponse(view *domain.SavedView) *dto.ViewResponse {
	var filters map[string]interface{}
	if view.Filters != "" && view.Filters != "{}" {
		if err := json.Unmarshal([]byte(view.Filters), &filters); err != nil {
			s.logger.Warn("Failed to parse view filters", zap.Error(err))
			filters = make(map[string]interface{})
		}
	} else {
		filters = make(map[string]interface{})
	}

	var sortBy string
	if view.SortBy != nil {
		sortBy = *view.SortBy
	}

	var groupByFieldID string
	if view.GroupByFieldID != nil {
		groupByFieldID = view.GroupByFieldID.String()
	}

	return &dto.ViewResponse{
		ViewID:         view.ID.String(),
		ProjectID:      view.ProjectID.String(),
		CreatedBy:      view.CreatedBy.String(),
		Name:           view.Name,
		Description:    view.Description,
		IsDefault:      view.IsDefault,
		IsShared:       view.IsShared,
		Filters:        filters,
		SortBy:         sortBy,
		SortDirection:  view.SortDirection,
		GroupByFieldID: groupByFieldID,
		CreatedAt:      view.CreatedAt,
		UpdatedAt:      view.UpdatedAt,
	}
}

func (s *viewService) applyBuiltInFilter(query *gorm.DB, field, operator string, value interface{}) *gorm.DB {
	switch operator {
	case "contains":
		if strVal, ok := value.(string); ok {
			return query.Where(field+" LIKE ?", "%"+strVal+"%")
		}
	case "eq":
		return query.Where(field+" = ?", value)
	case "ne":
		return query.Where(field+" != ?", value)
	}
	return query
}

func (s *viewService) applyCustomFieldFilter(query *gorm.DB, fieldID uuid.UUID, operator string, value interface{}) *gorm.DB {
	// Use JSONB operators on custom_fields_cache
	fieldKey := fieldID.String()

	switch operator {
	case "contains":
		if strVal, ok := value.(string); ok {
			return query.Where("custom_fields_cache->? LIKE ?", fieldKey, "%"+strVal+"%")
		}
	case "in":
		if arr, ok := value.([]interface{}); ok {
			return query.Where("custom_fields_cache->? ?| ARRAY[?]", fieldKey, arr)
		}
	case "eq":
		return query.Where("custom_fields_cache->>? = ?", fieldKey, value)
	}

	return query
}

func (s *viewService) applyGrouping(boards []domain.Board, groupByFieldID string, total int64) (interface{}, error) {
	fieldUUID, err := uuid.Parse(groupByFieldID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 그룹핑 필드 ID", 400)
	}

	// Fetch field
	field, err := s.repo.FindFieldByID(fieldUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "그룹핑 필드 조회 실패", 500)
	}

	// Fetch options for the field
	options, err := s.repo.FindOptionsByField(fieldUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "옵션 조회 실패", 500)
	}

	// Group boards by option
	groups := make(map[string][]dto.BoardResponse)
	for _, board := range boards {
		// Parse custom_fields_cache
		var cache map[string]interface{}
		if board.CustomFieldsCache != "" && board.CustomFieldsCache != "{}" {
			if err := json.Unmarshal([]byte(board.CustomFieldsCache), &cache); err == nil {
				// Get field value
				if fieldVal, exists := cache[groupByFieldID]; exists {
					// Handle array values (multi-select)
					if arr, ok := fieldVal.([]interface{}); ok {
						for _, optionID := range arr {
							optionIDStr := fmt.Sprintf("%v", optionID)
							groups[optionIDStr] = append(groups[optionIDStr], dto.BoardResponse{
								ID:           board.ID.String(),
								ProjectID:    board.ProjectID.String(),
								Title:        board.Title,
								Content:      board.Description,
								CustomFields: cache,
								CreatedAt:    board.CreatedAt,
								UpdatedAt:    board.UpdatedAt,
							})
						}
					} else {
						// Single value
						optionIDStr := fmt.Sprintf("%v", fieldVal)
						groups[optionIDStr] = append(groups[optionIDStr], dto.BoardResponse{
							ID:           board.ID.String(),
							ProjectID:    board.ProjectID.String(),
							Title:        board.Title,
							Content:      board.Description,
							CustomFields: cache,
							CreatedAt:    board.CreatedAt,
							UpdatedAt:    board.UpdatedAt,
						})
					}
				}
			}
		}
	}

	// Build response
	fieldResponse := dto.FieldResponse{
		FieldID:   field.ID.String(),
		ProjectID: field.ProjectID.String(),
		Name:      field.Name,
		FieldType: string(field.FieldType),
	}

	groupResponses := make([]dto.BoardGroup, 0)
	for _, option := range options {
		optionIDStr := option.ID.String()
		boardsInGroup := groups[optionIDStr]
		if boardsInGroup == nil {
			boardsInGroup = make([]dto.BoardResponse, 0)
		}

		groupResponses = append(groupResponses, dto.BoardGroup{
			GroupValue: map[string]interface{}{
				"option_id": option.ID.String(),
				"label":     option.Label,
				"color":     option.Color,
			},
			Boards: boardsInGroup,
			Count:  len(boardsInGroup),
		})
	}

	return dto.GroupedBoardsResponse{
		GroupByField: fieldResponse,
		Groups:       groupResponses,
		Total:        total,
	}, nil
}
