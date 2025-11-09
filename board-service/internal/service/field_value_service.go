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
	"net/url"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FieldValueService interface {
	// Set field values
	SetFieldValue(userID string, req *dto.SetFieldValueRequest) error
	SetMultiSelectValue(userID string, req *dto.SetMultiSelectValueRequest) error

	// Get field values
	GetBoardFieldValues(userID, boardID string) (*dto.BoardFieldValuesResponse, error)

	// Delete field value
	DeleteFieldValue(userID, boardID, fieldID string) error

	// Update board order in multi-select
	UpdateMultiSelectOrder(userID, boardID, fieldID string, req *dto.SetMultiSelectValueRequest) error
}

type fieldValueService struct {
	repo         repository.FieldRepository
	boardRepo    repository.BoardRepository
	projectRepo  repository.ProjectRepository
	cache        cache.FieldCache
	logger       *zap.Logger
	db           *gorm.DB
}

func NewFieldValueService(
	repo repository.FieldRepository,
	boardRepo repository.BoardRepository,
	projectRepo repository.ProjectRepository,
	cache cache.FieldCache,
	logger *zap.Logger,
	db *gorm.DB,
) FieldValueService {
	return &fieldValueService{
		repo:        repo,
		boardRepo:   boardRepo,
		projectRepo: projectRepo,
		cache:       cache,
		logger:      logger,
		db:          db,
	}
}

// ==================== Set Field Values ====================

func (s *fieldValueService) SetFieldValue(userID string, req *dto.SetFieldValueRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	boardUUID, err := uuid.Parse(req.BoardID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 보드 ID", 400)
	}

	fieldUUID, err := uuid.Parse(req.FieldID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 필드 ID", 400)
	}

	// 1. Fetch board
	board, err := s.boardRepo.FindByID(boardUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "보드를 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 조회 실패", 500)
	}

	// 2. Check project membership
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, board.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// 3. Fetch field
	field, err := s.repo.FindFieldByID(fieldUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "필드를 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 조회 실패", 500)
	}

	// 4. Validate field belongs to board's project
	if field.ProjectID != board.ProjectID {
		return apperrors.New(apperrors.ErrCodeBadRequest, "필드가 보드의 프로젝트에 속하지 않습니다", 400)
	}

	// 5. Validate and set value based on field type
	if err := s.setValueByType(boardUUID, fieldUUID, field.FieldType, field.Config, req.Value, req.Values); err != nil {
		return err
	}

	// 6. Update board's custom_fields_cache
	if err := s.updateBoardCache(boardUUID); err != nil {
		s.logger.Warn("Failed to update board cache", zap.Error(err))
	}

	return nil
}

func (s *fieldValueService) SetMultiSelectValue(userID string, req *dto.SetMultiSelectValueRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	boardUUID, err := uuid.Parse(req.BoardID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 보드 ID", 400)
	}

	fieldUUID, err := uuid.Parse(req.FieldID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 필드 ID", 400)
	}

	// 1. Fetch board
	board, err := s.boardRepo.FindByID(boardUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "보드를 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 조회 실패", 500)
	}

	// 2. Check project membership
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, board.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// 3. Fetch field
	field, err := s.repo.FindFieldByID(fieldUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "필드를 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 조회 실패", 500)
	}

	// 4. Validate field type
	if field.FieldType != domain.FieldTypeMultiSelect && field.FieldType != domain.FieldTypeMultiUser {
		return apperrors.New(apperrors.ErrCodeBadRequest, "Multi-select 또는 Multi-user 필드만 지원합니다", 400)
	}

	// 5. Delete existing values
	if err := s.repo.BatchDeleteFieldValues(boardUUID, fieldUUID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "기존 값 삭제 실패", 500)
	}

	// 6. Set new ordered values
	values := make([]domain.BoardFieldValue, 0, len(req.Values))
	for _, orderedVal := range req.Values {
		valueUUID, err := uuid.Parse(orderedVal.ValueID)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 값 ID", 400)
		}

		value := domain.BoardFieldValue{
			BoardID:      boardUUID,
			FieldID:      fieldUUID,
			DisplayOrder: orderedVal.DisplayOrder,
		}

		if field.FieldType == domain.FieldTypeMultiSelect {
			value.ValueOptionID = &valueUUID
		} else { // multi_user
			value.ValueUserID = &valueUUID
		}

		values = append(values, value)
	}

	if err := s.repo.BatchSetFieldValues(values); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "값 설정 실패", 500)
	}

	// 7. Update board cache
	if err := s.updateBoardCache(boardUUID); err != nil {
		s.logger.Warn("Failed to update board cache", zap.Error(err))
	}

	return nil
}

// ==================== Get Field Values ====================

func (s *fieldValueService) GetBoardFieldValues(userID, boardID string) (*dto.BoardFieldValuesResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	boardUUID, err := uuid.Parse(boardID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 보드 ID", 400)
	}

	// 1. Fetch board
	board, err := s.boardRepo.FindByID(boardUUID)
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

	// 3. Get values from cache if available
	if board.CustomFieldsCache != "" && board.CustomFieldsCache != "{}" {
		var fields map[string]interface{}
		if err := json.Unmarshal([]byte(board.CustomFieldsCache), &fields); err == nil {
			return &dto.BoardFieldValuesResponse{
				BoardID: boardID,
				Fields:  fields,
			}, nil
		}
	}

	// 4. Fetch from database
	values, err := s.repo.FindFieldValuesByBoard(boardUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 값 조회 실패", 500)
	}

	// 5. Build response
	fields := make(map[string]interface{})
	for _, val := range values {
		fieldID := val.FieldID.String()

		// Get actual value based on type
		var actualValue interface{}
		if val.ValueText != nil {
			actualValue = *val.ValueText
		} else if val.ValueNumber != nil {
			actualValue = *val.ValueNumber
		} else if val.ValueDate != nil {
			actualValue = *val.ValueDate
		} else if val.ValueBoolean != nil {
			actualValue = *val.ValueBoolean
		} else if val.ValueOptionID != nil {
			actualValue = val.ValueOptionID.String()
		} else if val.ValueUserID != nil {
			actualValue = val.ValueUserID.String()
		}

		// For multi-select/multi-user, accumulate as array
		if existingVal, exists := fields[fieldID]; exists {
			if arr, ok := existingVal.([]interface{}); ok {
				fields[fieldID] = append(arr, actualValue)
			}
		} else {
			// Check if this is first value for multi-select field
			fields[fieldID] = actualValue
		}
	}

	return &dto.BoardFieldValuesResponse{
		BoardID: boardID,
		Fields:  fields,
	}, nil
}

// ==================== Delete Field Value ====================

func (s *fieldValueService) DeleteFieldValue(userID, boardID, fieldID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	boardUUID, err := uuid.Parse(boardID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 보드 ID", 400)
	}

	fieldUUID, err := uuid.Parse(fieldID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 필드 ID", 400)
	}

	// 1. Fetch board
	board, err := s.boardRepo.FindByID(boardUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "보드를 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 조회 실패", 500)
	}

	// 2. Check project membership
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, board.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// 3. Delete field value
	if err := s.repo.DeleteFieldValue(boardUUID, fieldUUID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 값 삭제 실패", 500)
	}

	// 4. Update board cache
	if err := s.updateBoardCache(boardUUID); err != nil {
		s.logger.Warn("Failed to update board cache", zap.Error(err))
	}

	return nil
}

func (s *fieldValueService) UpdateMultiSelectOrder(userID, boardID, fieldID string, req *dto.SetMultiSelectValueRequest) error {
	return s.SetMultiSelectValue(userID, req)
}

// ==================== Helper Methods ====================

func (s *fieldValueService) setValueByType(boardID, fieldID uuid.UUID, fieldType domain.FieldType, configJSON string, singleValue, multiValue interface{}) error {
	// Parse config
	var config domain.FieldConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		s.logger.Warn("Failed to parse config", zap.Error(err))
	}

	switch fieldType {
	case domain.FieldTypeText:
		return s.setTextValue(boardID, fieldID, singleValue, config)
	case domain.FieldTypeNumber:
		return s.setNumberValue(boardID, fieldID, singleValue, config)
	case domain.FieldTypeSingleSelect:
		return s.setSingleSelectValue(boardID, fieldID, singleValue)
	case domain.FieldTypeMultiSelect:
		return s.setMultiSelectValues(boardID, fieldID, multiValue, config)
	case domain.FieldTypeDate, domain.FieldTypeDateTime:
		return s.setDateValue(boardID, fieldID, singleValue)
	case domain.FieldTypeSingleUser:
		return s.setSingleUserValue(boardID, fieldID, singleValue)
	case domain.FieldTypeMultiUser:
		return s.setMultiUserValues(boardID, fieldID, multiValue, config)
	case domain.FieldTypeCheckbox:
		return s.setCheckboxValue(boardID, fieldID, singleValue)
	case domain.FieldTypeURL:
		return s.setURLValue(boardID, fieldID, singleValue)
	default:
		return apperrors.New(apperrors.ErrCodeBadRequest, "지원하지 않는 필드 타입입니다", 400)
	}
}

func (s *fieldValueService) setTextValue(boardID, fieldID uuid.UUID, value interface{}, config domain.FieldConfig) error {
	strVal, ok := value.(string)
	if !ok {
		return apperrors.New(apperrors.ErrCodeBadRequest, "텍스트 값이 필요합니다", 400)
	}

	// Validate max length
	if config.MaxLength != nil && len(strVal) > *config.MaxLength {
		return apperrors.New(apperrors.ErrCodeBadRequest, fmt.Sprintf("텍스트 길이가 최대값(%d)을 초과했습니다", *config.MaxLength), 400)
	}

	val := &domain.BoardFieldValue{
		BoardID:   boardID,
		FieldID:   fieldID,
		ValueText: &strVal,
	}

	return s.repo.SetFieldValue(val)
}

func (s *fieldValueService) setNumberValue(boardID, fieldID uuid.UUID, value interface{}, config domain.FieldConfig) error {
	var numVal float64
	switch v := value.(type) {
	case float64:
		numVal = v
	case int:
		numVal = float64(v)
	default:
		return apperrors.New(apperrors.ErrCodeBadRequest, "숫자 값이 필요합니다", 400)
	}

	// Validate range
	if config.Min != nil && numVal < *config.Min {
		return apperrors.New(apperrors.ErrCodeBadRequest, fmt.Sprintf("값이 최소값(%.2f)보다 작습니다", *config.Min), 400)
	}
	if config.Max != nil && numVal > *config.Max {
		return apperrors.New(apperrors.ErrCodeBadRequest, fmt.Sprintf("값이 최대값(%.2f)보다 큽니다", *config.Max), 400)
	}

	val := &domain.BoardFieldValue{
		BoardID:     boardID,
		FieldID:     fieldID,
		ValueNumber: &numVal,
	}

	return s.repo.SetFieldValue(val)
}

func (s *fieldValueService) setSingleSelectValue(boardID, fieldID uuid.UUID, value interface{}) error {
	optionIDStr, ok := value.(string)
	if !ok {
		return apperrors.New(apperrors.ErrCodeBadRequest, "옵션 ID가 필요합니다", 400)
	}

	optionID, err := uuid.Parse(optionIDStr)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 옵션 ID", 400)
	}

	// Validate option exists and belongs to field
	option, err := s.repo.FindOptionByID(optionID)
	if err != nil || option.FieldID != fieldID {
		return apperrors.New(apperrors.ErrCodeBadRequest, "유효하지 않은 옵션입니다", 400)
	}

	val := &domain.BoardFieldValue{
		BoardID:       boardID,
		FieldID:       fieldID,
		ValueOptionID: &optionID,
	}

	return s.repo.SetFieldValue(val)
}

func (s *fieldValueService) setMultiSelectValues(boardID, fieldID uuid.UUID, values interface{}, config domain.FieldConfig) error {
	optionIDs, ok := values.([]interface{})
	if !ok {
		return apperrors.New(apperrors.ErrCodeBadRequest, "옵션 ID 배열이 필요합니다", 400)
	}

	// Validate max selections
	if config.MaxSelections != nil && len(optionIDs) > *config.MaxSelections {
		return apperrors.New(apperrors.ErrCodeBadRequest, fmt.Sprintf("선택 개수가 최대값(%d)을 초과했습니다", *config.MaxSelections), 400)
	}

	// Delete existing values
	if err := s.repo.BatchDeleteFieldValues(boardID, fieldID); err != nil {
		return err
	}

	// Set new values
	fieldValues := make([]domain.BoardFieldValue, 0, len(optionIDs))
	for i, optionIDVal := range optionIDs {
		optionIDStr, ok := optionIDVal.(string)
		if !ok {
			return apperrors.New(apperrors.ErrCodeBadRequest, "잘못된 옵션 ID 형식", 400)
		}

		optionID, err := uuid.Parse(optionIDStr)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 옵션 ID", 400)
		}

		fieldValues = append(fieldValues, domain.BoardFieldValue{
			BoardID:       boardID,
			FieldID:       fieldID,
			ValueOptionID: &optionID,
			DisplayOrder:  i,
		})
	}

	return s.repo.BatchSetFieldValues(fieldValues)
}

func (s *fieldValueService) setDateValue(boardID, fieldID uuid.UUID, value interface{}) error {
	dateStr, ok := value.(string)
	if !ok {
		return apperrors.New(apperrors.ErrCodeBadRequest, "날짜 문자열이 필요합니다 (ISO 8601)", 400)
	}

	dateVal, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 날짜 형식입니다 (ISO 8601)", 400)
	}

	val := &domain.BoardFieldValue{
		BoardID:   boardID,
		FieldID:   fieldID,
		ValueDate: &dateVal,
	}

	return s.repo.SetFieldValue(val)
}

func (s *fieldValueService) setSingleUserValue(boardID, fieldID uuid.UUID, value interface{}) error {
	userIDStr, ok := value.(string)
	if !ok {
		return apperrors.New(apperrors.ErrCodeBadRequest, "사용자 ID가 필요합니다", 400)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	val := &domain.BoardFieldValue{
		BoardID:     boardID,
		FieldID:     fieldID,
		ValueUserID: &userID,
	}

	return s.repo.SetFieldValue(val)
}

func (s *fieldValueService) setMultiUserValues(boardID, fieldID uuid.UUID, values interface{}, config domain.FieldConfig) error {
	userIDs, ok := values.([]interface{})
	if !ok {
		return apperrors.New(apperrors.ErrCodeBadRequest, "사용자 ID 배열이 필요합니다", 400)
	}

	// Validate max users
	if config.MaxUsers != nil && len(userIDs) > *config.MaxUsers {
		return apperrors.New(apperrors.ErrCodeBadRequest, fmt.Sprintf("사용자 개수가 최대값(%d)을 초과했습니다", *config.MaxUsers), 400)
	}

	// Delete existing values
	if err := s.repo.BatchDeleteFieldValues(boardID, fieldID); err != nil {
		return err
	}

	// Set new values
	fieldValues := make([]domain.BoardFieldValue, 0, len(userIDs))
	for i, userIDVal := range userIDs {
		userIDStr, ok := userIDVal.(string)
		if !ok {
			return apperrors.New(apperrors.ErrCodeBadRequest, "잘못된 사용자 ID 형식", 400)
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
		}

		fieldValues = append(fieldValues, domain.BoardFieldValue{
			BoardID:      boardID,
			FieldID:      fieldID,
			ValueUserID:  &userID,
			DisplayOrder: i,
		})
	}

	return s.repo.BatchSetFieldValues(fieldValues)
}

func (s *fieldValueService) setCheckboxValue(boardID, fieldID uuid.UUID, value interface{}) error {
	boolVal, ok := value.(bool)
	if !ok {
		return apperrors.New(apperrors.ErrCodeBadRequest, "불린 값이 필요합니다", 400)
	}

	val := &domain.BoardFieldValue{
		BoardID:      boardID,
		FieldID:      fieldID,
		ValueBoolean: &boolVal,
	}

	return s.repo.SetFieldValue(val)
}

func (s *fieldValueService) setURLValue(boardID, fieldID uuid.UUID, value interface{}) error {
	urlStr, ok := value.(string)
	if !ok {
		return apperrors.New(apperrors.ErrCodeBadRequest, "URL 문자열이 필요합니다", 400)
	}

	// Validate URL format
	if _, err := url.ParseRequestURI(urlStr); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "유효하지 않은 URL 형식입니다", 400)
	}

	val := &domain.BoardFieldValue{
		BoardID:   boardID,
		FieldID:   fieldID,
		ValueText: &urlStr,
	}

	return s.repo.SetFieldValue(val)
}

func (s *fieldValueService) updateBoardCache(boardID uuid.UUID) error {
	// Fetch all field values for the board
	values, err := s.repo.FindFieldValuesByBoard(boardID)
	if err != nil {
		return err
	}

	// Build cache map
	cache := make(map[string]interface{})
	multiValueFields := make(map[string][]interface{})

	for _, val := range values {
		fieldIDStr := val.FieldID.String()

		var actualValue interface{}
		if val.ValueText != nil {
			actualValue = *val.ValueText
		} else if val.ValueNumber != nil {
			actualValue = *val.ValueNumber
		} else if val.ValueDate != nil {
			actualValue = val.ValueDate.Format(time.RFC3339)
		} else if val.ValueBoolean != nil {
			actualValue = *val.ValueBoolean
		} else if val.ValueOptionID != nil {
			actualValue = val.ValueOptionID.String()
		} else if val.ValueUserID != nil {
			actualValue = val.ValueUserID.String()
		}

		// Check if this is a multi-value field
		if _, exists := multiValueFields[fieldIDStr]; exists || val.DisplayOrder > 0 {
			multiValueFields[fieldIDStr] = append(multiValueFields[fieldIDStr], actualValue)
		} else {
			// First value - check if there are more
			nextValues, _ := s.repo.FindFieldValuesByBoardAndField(boardID, val.FieldID)
			if len(nextValues) > 1 {
				multiValueFields[fieldIDStr] = []interface{}{actualValue}
			} else {
				cache[fieldIDStr] = actualValue
			}
		}
	}

	// Merge multi-value fields
	for fieldID, values := range multiValueFields {
		cache[fieldID] = values
	}

	// Serialize to JSON
	cacheJSON, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	// Update board's custom_fields_cache
	if err := s.db.Model(&domain.Board{}).
		Where("id = ?", boardID).
		Update("custom_fields_cache", string(cacheJSON)).Error; err != nil {
		return err
	}

	// Invalidate Redis cache
	ctx := context.Background()
	if err := s.cache.InvalidateBoardFieldValues(ctx, boardID.String()); err != nil {
		s.logger.Warn("Failed to invalidate board field values cache", zap.Error(err))
	}

	return nil
}
