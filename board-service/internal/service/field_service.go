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
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FieldService interface {
	// Field CRUD
	CreateField(userID string, req *dto.CreateFieldRequest) (*dto.FieldResponse, error)
	GetFieldsByProject(userID, projectID string) ([]dto.FieldResponse, error)
	GetField(userID, fieldID string) (*dto.FieldResponse, error)
	UpdateField(userID, fieldID string, req *dto.UpdateFieldRequest) (*dto.FieldResponse, error)
	DeleteField(userID, fieldID string) error
	UpdateFieldOrder(userID, projectID string, req *dto.UpdateFieldOrderRequest) error

	// Option CRUD
	CreateOption(userID string, req *dto.CreateOptionRequest) (*dto.OptionResponse, error)
	GetOptionsByField(userID, fieldID string) ([]dto.OptionResponse, error)
	GetOption(userID, optionID string) (*dto.OptionResponse, error)
	UpdateOption(userID, optionID string, req *dto.UpdateOptionRequest) (*dto.OptionResponse, error)
	DeleteOption(userID, optionID string) error
	UpdateOptionOrder(userID, fieldID string, req *dto.UpdateOptionOrderRequest) error
}

type fieldService struct {
	repo        repository.FieldRepository
	projectRepo repository.ProjectRepository
	cache       cache.FieldCache
	logger      *zap.Logger
	db          *gorm.DB
}

func NewFieldService(
	repo repository.FieldRepository,
	projectRepo repository.ProjectRepository,
	cache cache.FieldCache,
	logger *zap.Logger,
	db *gorm.DB,
) FieldService {
	return &fieldService{
		repo:        repo,
		projectRepo: projectRepo,
		cache:       cache,
		logger:      logger,
		db:          db,
	}
}

// ==================== Field CRUD ====================

func (s *fieldService) CreateField(userID string, req *dto.CreateFieldRequest) (*dto.FieldResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	// 1. Check project membership and permissions
	member, err := s.projectRepo.FindMemberByUserAndProject(userUUID, projectUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// Only ADMIN and OWNER can create fields (Level >= 50)
	if member.Role == nil || member.Role.Level < 50 {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "필드 생성 권한이 없습니다 (ADMIN 이상)", 403)
	}

	// 2. Validate field type
	if !isValidFieldType(req.FieldType) {
		return nil, apperrors.New(apperrors.ErrCodeBadRequest, "지원하지 않는 필드 타입입니다", 400)
	}

	// 3. Validate and serialize config
	configJSON, err := s.validateAndSerializeConfig(req.FieldType, req.Config)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "필드 설정이 유효하지 않습니다", 400)
	}

	// 4. Get next display order
	existingFields, err := s.repo.FindFieldsByProject(projectUUID)
	if err != nil {
		s.logger.Warn("Failed to fetch existing fields", zap.Error(err))
	}
	nextOrder := len(existingFields)

	// 5. Create field
	field := &domain.ProjectField{
		ProjectID:    projectUUID,
		Name:         req.Name,
		FieldType:    domain.FieldType(req.FieldType),
		Description:  req.Description,
		DisplayOrder: nextOrder,
		IsRequired:   req.IsRequired,
		Config:       configJSON,
	}

	if err := s.repo.CreateField(field); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 생성 실패", 500)
	}

	// Invalidate cache
	ctx := context.Background()
	if err := s.cache.InvalidateProjectFields(ctx, req.ProjectID); err != nil {
		s.logger.Warn("Failed to invalidate project fields cache", zap.Error(err))
	}

	return s.buildFieldResponse(field), nil
}

func (s *fieldService) GetFieldsByProject(userID, projectID string) ([]dto.FieldResponse, error) {
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

	// Try to get from cache first
	ctx := context.Background()
	if cachedData, err := s.cache.GetProjectFields(ctx, projectID); err == nil {
		var responses []dto.FieldResponse
		if json.Unmarshal(cachedData, &responses) == nil {
			return responses, nil
		}
	}

	// Fetch fields from DB
	fields, err := s.repo.FindFieldsByProject(projectUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 조회 실패", 500)
	}

	responses := make([]dto.FieldResponse, 0, len(fields))
	for i := range fields {
		responses = append(responses, *s.buildFieldResponse(&fields[i]))
	}

	// Cache the result (TTL: 5 minutes)
	if data, err := json.Marshal(responses); err == nil {
		if err := s.cache.SetProjectFields(ctx, projectID, data, 5*time.Minute); err != nil {
			s.logger.Warn("Failed to cache project fields", zap.Error(err))
		}
	}

	return responses, nil
}

func (s *fieldService) GetField(userID, fieldID string) (*dto.FieldResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	fieldUUID, err := uuid.Parse(fieldID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 필드 ID", 400)
	}

	// Fetch field
	field, err := s.repo.FindFieldByID(fieldUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "필드를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 조회 실패", 500)
	}

	// Check project membership
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, field.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	return s.buildFieldResponse(field), nil
}

func (s *fieldService) UpdateField(userID, fieldID string, req *dto.UpdateFieldRequest) (*dto.FieldResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	fieldUUID, err := uuid.Parse(fieldID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 필드 ID", 400)
	}

	// Fetch field
	field, err := s.repo.FindFieldByID(fieldUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "필드를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 조회 실패", 500)
	}

	// Check permissions
	member, err := s.projectRepo.FindMemberByUserAndProject(userUUID, field.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	if member.Role == nil || member.Role.Level < 50 {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "필드 수정 권한이 없습니다 (ADMIN 이상)", 403)
	}

	// Update fields
	if req.Name != "" {
		field.Name = req.Name
	}
	if req.Description != "" {
		field.Description = req.Description
	}
	if req.IsRequired != nil {
		field.IsRequired = *req.IsRequired
	}
	if req.Config != nil {
		configJSON, err := s.validateAndSerializeConfig(string(field.FieldType), req.Config)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "필드 설정이 유효하지 않습니다", 400)
		}
		field.Config = configJSON
	}
	if req.DisplayOrder != nil {
		field.DisplayOrder = *req.DisplayOrder
	}

	// Save
	if err := s.repo.UpdateField(field); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 수정 실패", 500)
	}

	// Invalidate cache
	ctx := context.Background()
	if err := s.cache.InvalidateProjectFields(ctx, field.ProjectID.String()); err != nil {
		s.logger.Warn("Failed to invalidate project fields cache", zap.Error(err))
	}

	return s.buildFieldResponse(field), nil
}

func (s *fieldService) DeleteField(userID, fieldID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	fieldUUID, err := uuid.Parse(fieldID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 필드 ID", 400)
	}

	// Fetch field
	field, err := s.repo.FindFieldByID(fieldUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "필드를 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 조회 실패", 500)
	}

	// Check permissions
	member, err := s.projectRepo.FindMemberByUserAndProject(userUUID, field.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	if member.Role == nil || member.Role.Level < 50 {
		return apperrors.New(apperrors.ErrCodeForbidden, "필드 삭제 권한이 없습니다 (ADMIN 이상)", 403)
	}

	// Cannot delete system default fields
	if field.IsSystemDefault {
		return apperrors.New(apperrors.ErrCodeBadRequest, "시스템 기본 필드는 삭제할 수 없습니다", 400)
	}

	// Soft delete
	if err := s.repo.DeleteField(fieldUUID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 삭제 실패", 500)
	}

	// Invalidate cache
	ctx := context.Background()
	if err := s.cache.InvalidateProjectFields(ctx, field.ProjectID.String()); err != nil {
		s.logger.Warn("Failed to invalidate project fields cache", zap.Error(err))
	}

	return nil
}

func (s *fieldService) UpdateFieldOrder(userID, projectID string, req *dto.UpdateFieldOrderRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	// Check permissions
	member, err := s.projectRepo.FindMemberByUserAndProject(userUUID, projectUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	if member.Role == nil || member.Role.Level < 50 {
		return apperrors.New(apperrors.ErrCodeForbidden, "필드 순서 변경 권한이 없습니다 (ADMIN 이상)", 403)
	}

	// Build orders map
	orders := make(map[uuid.UUID]int)
	for _, item := range req.FieldOrders {
		fieldUUID, err := uuid.Parse(item.FieldID)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 필드 ID", 400)
		}
		orders[fieldUUID] = item.DisplayOrder
	}

	// Batch update
	if err := s.repo.BatchUpdateFieldOrders(orders); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 순서 업데이트 실패", 500)
	}

	// Invalidate cache
	ctx := context.Background()
	if err := s.cache.InvalidateProjectFields(ctx, projectID); err != nil {
		s.logger.Warn("Failed to invalidate project fields cache", zap.Error(err))
	}

	return nil
}

// ==================== Option CRUD ====================

func (s *fieldService) CreateOption(userID string, req *dto.CreateOptionRequest) (*dto.OptionResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	fieldUUID, err := uuid.Parse(req.FieldID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 필드 ID", 400)
	}

	// Fetch field
	field, err := s.repo.FindFieldByID(fieldUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "필드를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 조회 실패", 500)
	}

	// Validate field type (only select types can have options)
	if field.FieldType != domain.FieldTypeSingleSelect && field.FieldType != domain.FieldTypeMultiSelect {
		return nil, apperrors.New(apperrors.ErrCodeBadRequest, "Select 타입 필드만 옵션을 가질 수 있습니다", 400)
	}

	// Check permissions
	member, err := s.projectRepo.FindMemberByUserAndProject(userUUID, field.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	if member.Role == nil || member.Role.Level < 50 {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "옵션 생성 권한이 없습니다 (ADMIN 이상)", 403)
	}

	// Get next display order
	existingOptions, err := s.repo.FindOptionsByField(fieldUUID)
	if err != nil {
		s.logger.Warn("Failed to fetch existing options", zap.Error(err))
	}
	nextOrder := len(existingOptions)

	// Create option
	option := &domain.FieldOption{
		FieldID:      fieldUUID,
		Label:        req.Label,
		Color:        req.Color,
		Description:  req.Description,
		DisplayOrder: nextOrder,
	}

	if err := s.repo.CreateOption(option); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "옵션 생성 실패", 500)
	}

	// Invalidate cache
	ctx := context.Background()
	if err := s.cache.InvalidateFieldOptions(ctx, req.FieldID); err != nil {
		s.logger.Warn("Failed to invalidate field options cache", zap.Error(err))
	}

	return s.buildOptionResponse(option), nil
}

func (s *fieldService) GetOptionsByField(userID, fieldID string) ([]dto.OptionResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	fieldUUID, err := uuid.Parse(fieldID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 필드 ID", 400)
	}

	// Fetch field
	field, err := s.repo.FindFieldByID(fieldUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "필드를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 조회 실패", 500)
	}

	// Check project membership
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, field.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// Try to get from cache first
	ctx := context.Background()
	if cachedData, err := s.cache.GetFieldOptions(ctx, fieldID); err == nil {
		var responses []dto.OptionResponse
		if json.Unmarshal(cachedData, &responses) == nil {
			return responses, nil
		}
	}

	// Fetch options from DB
	options, err := s.repo.FindOptionsByField(fieldUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "옵션 조회 실패", 500)
	}

	responses := make([]dto.OptionResponse, 0, len(options))
	for i := range options {
		responses = append(responses, *s.buildOptionResponse(&options[i]))
	}

	// Cache the result (TTL: 5 minutes)
	if data, err := json.Marshal(responses); err == nil {
		if err := s.cache.SetFieldOptions(ctx, fieldID, data, 5*time.Minute); err != nil {
			s.logger.Warn("Failed to cache field options", zap.Error(err))
		}
	}

	return responses, nil
}

func (s *fieldService) GetOption(userID, optionID string) (*dto.OptionResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	optionUUID, err := uuid.Parse(optionID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 옵션 ID", 400)
	}

	// Fetch option
	option, err := s.repo.FindOptionByID(optionUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "옵션을 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "옵션 조회 실패", 500)
	}

	// Fetch field to check project membership
	field, err := s.repo.FindFieldByID(option.FieldID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 조회 실패", 500)
	}

	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, field.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	return s.buildOptionResponse(option), nil
}

func (s *fieldService) UpdateOption(userID, optionID string, req *dto.UpdateOptionRequest) (*dto.OptionResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	optionUUID, err := uuid.Parse(optionID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 옵션 ID", 400)
	}

	// Fetch option
	option, err := s.repo.FindOptionByID(optionUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "옵션을 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "옵션 조회 실패", 500)
	}

	// Check permissions
	field, err := s.repo.FindFieldByID(option.FieldID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 조회 실패", 500)
	}

	member, err := s.projectRepo.FindMemberByUserAndProject(userUUID, field.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	if member.Role == nil || member.Role.Level < 50 {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "옵션 수정 권한이 없습니다 (ADMIN 이상)", 403)
	}

	// Update fields
	if req.Label != "" {
		option.Label = req.Label
	}
	if req.Color != "" {
		option.Color = req.Color
	}
	if req.Description != "" {
		option.Description = req.Description
	}

	// Save
	if err := s.repo.UpdateOption(option); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "옵션 수정 실패", 500)
	}

	// Invalidate cache
	ctx := context.Background()
	if err := s.cache.InvalidateFieldOptions(ctx, option.FieldID.String()); err != nil {
		s.logger.Warn("Failed to invalidate field options cache", zap.Error(err))
	}

	return s.buildOptionResponse(option), nil
}

func (s *fieldService) DeleteOption(userID, optionID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	optionUUID, err := uuid.Parse(optionID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 옵션 ID", 400)
	}

	// Fetch option
	option, err := s.repo.FindOptionByID(optionUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "옵션을 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "옵션 조회 실패", 500)
	}

	// Check permissions
	field, err := s.repo.FindFieldByID(option.FieldID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 조회 실패", 500)
	}

	member, err := s.projectRepo.FindMemberByUserAndProject(userUUID, field.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	if member.Role == nil || member.Role.Level < 50 {
		return apperrors.New(apperrors.ErrCodeForbidden, "옵션 삭제 권한이 없습니다 (ADMIN 이상)", 403)
	}

	// Soft delete
	if err := s.repo.DeleteOption(optionUUID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "옵션 삭제 실패", 500)
	}

	// Invalidate cache
	ctx := context.Background()
	if err := s.cache.InvalidateFieldOptions(ctx, option.FieldID.String()); err != nil {
		s.logger.Warn("Failed to invalidate field options cache", zap.Error(err))
	}

	return nil
}

func (s *fieldService) UpdateOptionOrder(userID, fieldID string, req *dto.UpdateOptionOrderRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	fieldUUID, err := uuid.Parse(fieldID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 필드 ID", 400)
	}

	// Fetch field
	field, err := s.repo.FindFieldByID(fieldUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "필드를 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 조회 실패", 500)
	}

	// Check permissions
	member, err := s.projectRepo.FindMemberByUserAndProject(userUUID, field.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	if member.Role == nil || member.Role.Level < 50 {
		return apperrors.New(apperrors.ErrCodeForbidden, "옵션 순서 변경 권한이 없습니다 (ADMIN 이상)", 403)
	}

	// Build orders map
	orders := make(map[uuid.UUID]int)
	for _, item := range req.OptionOrders {
		optionUUID, err := uuid.Parse(item.OptionID)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 옵션 ID", 400)
		}
		orders[optionUUID] = item.DisplayOrder
	}

	// Batch update
	if err := s.repo.BatchUpdateOptionOrders(orders); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "옵션 순서 업데이트 실패", 500)
	}

	// Invalidate cache
	ctx := context.Background()
	if err := s.cache.InvalidateFieldOptions(ctx, fieldID); err != nil {
		s.logger.Warn("Failed to invalidate field options cache", zap.Error(err))
	}

	return nil
}

// ==================== Helper Methods ====================

func (s *fieldService) buildFieldResponse(field *domain.ProjectField) *dto.FieldResponse {
	// Parse config JSON
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(field.Config), &config); err != nil {
		s.logger.Warn("Failed to parse field config", zap.Error(err))
		config = make(map[string]interface{})
	}

	// Parse can_edit_roles
	var canEditRoles []string
	if field.CanEditRoles != nil && *field.CanEditRoles != "" {
		canEditRoles = strings.Split(*field.CanEditRoles, ",")
	}

	return &dto.FieldResponse{
		FieldID:         field.ID.String(),
		ProjectID:       field.ProjectID.String(),
		Name:            field.Name,
		FieldType:       string(field.FieldType),
		Description:     field.Description,
		DisplayOrder:    field.DisplayOrder,
		IsRequired:      field.IsRequired,
		IsSystemDefault: field.IsSystemDefault,
		Config:          config,
		CanEditRoles:    canEditRoles,
		CreatedAt:       field.CreatedAt,
		UpdatedAt:       field.UpdatedAt,
	}
}

func (s *fieldService) buildOptionResponse(option *domain.FieldOption) *dto.OptionResponse {
	return &dto.OptionResponse{
		OptionID:     option.ID.String(),
		FieldID:      option.FieldID.String(),
		Label:        option.Label,
		Color:        option.Color,
		Description:  option.Description,
		DisplayOrder: option.DisplayOrder,
		CreatedAt:    option.CreatedAt,
		UpdatedAt:    option.UpdatedAt,
	}
}

func (s *fieldService) validateAndSerializeConfig(fieldType string, config map[string]interface{}) (string, error) {
	// Validate config based on field type
	switch fieldType {
	case "text":
		// Validate max_length if present
		if maxLen, ok := config["max_length"]; ok {
			if val, ok := maxLen.(float64); ok && val <= 0 {
				return "", fmt.Errorf("max_length must be positive")
			}
		}
	case "number":
		// Validate min/max if present
		if min, ok := config["min"]; ok {
			if max, ok := config["max"]; ok {
				if min.(float64) > max.(float64) {
					return "", fmt.Errorf("min cannot be greater than max")
				}
			}
		}
	case "multi_select":
		// Validate max_selections if present
		if maxSel, ok := config["max_selections"]; ok {
			if val, ok := maxSel.(float64); ok && val <= 0 {
				return "", fmt.Errorf("max_selections must be positive")
			}
		}
	case "multi_user":
		// Validate max_users if present
		if maxUsers, ok := config["max_users"]; ok {
			if val, ok := maxUsers.(float64); ok && val <= 0 {
				return "", fmt.Errorf("max_users must be positive")
			}
		}
	}

	// Serialize to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	return string(configJSON), nil
}

func isValidFieldType(fieldType string) bool {
	validTypes := []string{
		"text", "number",
		"single_select", "multi_select",
		"date", "datetime",
		"single_user", "multi_user",
		"checkbox", "url",
	}

	for _, t := range validTypes {
		if fieldType == t {
			return true
		}
	}
	return false
}
