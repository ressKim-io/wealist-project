package repository

import (
	"board-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BoardRepository interface {
	// CRUD
	Create(board *domain.Board) error
	FindByID(id uuid.UUID) (*domain.Board, error)
	FindByProject(projectID uuid.UUID, filters BoardFilters, page, limit int) ([]domain.Board, int64, error)
	Update(board *domain.Board) error
	Delete(id uuid.UUID) error
}

type BoardFilters struct {
	AssigneeID uuid.UUID
	AuthorID   uuid.UUID
	// Custom field filtering is now done via JSONB queries in ViewService
	// using custom_fields_cache column with GIN index
}

type boardRepository struct {
	db *gorm.DB
}

func NewBoardRepository(db *gorm.DB) BoardRepository {
	return &boardRepository{db: db}
}

// ==================== CRUD ====================

func (r *boardRepository) Create(board *domain.Board) error {
	return r.db.Create(board).Error
}

func (r *boardRepository) FindByID(id uuid.UUID) (*domain.Board, error) {
	var board domain.Board
	if err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&board).Error; err != nil {
		return nil, err
	}
	return &board, nil
}

func (r *boardRepository) FindByProject(projectID uuid.UUID, filters BoardFilters, page, limit int) ([]domain.Board, int64, error) {
	var boards []domain.Board
	var total int64

	query := r.db.Model(&domain.Board{}).Where("project_id = ? AND is_deleted = ?", projectID, false)

	// Apply basic filters (Assignee, Author)
	if filters.AssigneeID != uuid.Nil {
		query = query.Where("assignee_id = ?", filters.AssigneeID)
	}
	if filters.AuthorID != uuid.Nil {
		query = query.Where("created_by = ?", filters.AuthorID)
	}

	// Note: Custom field filtering (stage, role, importance, etc.) is now done
	// via ViewService using JSONB queries on custom_fields_cache column
	// Example: WHERE custom_fields_cache->>'field-id' = 'value'

	// Total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&boards).Error; err != nil {
		return nil, 0, err
	}

	return boards, total, nil
}

func (r *boardRepository) Update(board *domain.Board) error {
	return r.db.Save(board).Error
}

func (r *boardRepository) Delete(id uuid.UUID) error {
	// Soft delete
	return r.db.Model(&domain.Board{}).Where("id = ?", id).Update("is_deleted", true).Error
}
