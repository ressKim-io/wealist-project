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

	// Board Roles (many-to-many)
	CreateBoardRoles(boardID uuid.UUID, roleIDs []uuid.UUID) error
	DeleteBoardRolesByBoard(boardID uuid.UUID) error
	FindRolesByBoard(boardID uuid.UUID) ([]domain.BoardRole, error)
	FindRolesByBoards(boardIDs []uuid.UUID) (map[uuid.UUID][]domain.BoardRole, error)
	FindBoardsByRole(roleID uuid.UUID) ([]domain.Board, error)

	// Phase 4 TODO implementation: Update boards when custom fields are deleted
	UpdateBoardsRoleToDefault(oldRoleID, defaultRoleID uuid.UUID) error
	UpdateBoardsStageToDefault(oldStageID, defaultStageID uuid.UUID) error
	UpdateBoardsImportanceToDefault(oldImportanceID, defaultImportanceID uuid.UUID) error
}

type BoardFilters struct {
	StageID      uuid.UUID
	RoleID       uuid.UUID
	ImportanceID uuid.UUID
	AssigneeID   uuid.UUID
	AuthorID     uuid.UUID
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

	// Apply filters
	if filters.StageID != uuid.Nil {
		query = query.Where("custom_stage_id = ?", filters.StageID)
	}
	if filters.ImportanceID != uuid.Nil {
		query = query.Where("custom_importance_id = ?", filters.ImportanceID)
	}
	if filters.AssigneeID != uuid.Nil {
		query = query.Where("assignee_id = ?", filters.AssigneeID)
	}
	if filters.AuthorID != uuid.Nil {
		query = query.Where("created_by = ?", filters.AuthorID)
	}

	// Role filter (JOIN)
	if filters.RoleID != uuid.Nil {
		query = query.Joins("INNER JOIN board_roles ON boards.id = board_roles.board_id").
			Where("board_roles.custom_role_id = ?", filters.RoleID)
	}

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

// ==================== Board Roles (Many-to-Many) ====================

func (r *boardRepository) CreateBoardRoles(boardID uuid.UUID, roleIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, roleID := range roleIDs {
			boardRole := &domain.BoardRole{
				BoardID:      boardID,
				CustomRoleID: roleID,
			}
			if err := tx.Create(boardRole).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *boardRepository) DeleteBoardRolesByBoard(boardID uuid.UUID) error {
	return r.db.Where("board_id = ?", boardID).Delete(&domain.BoardRole{}).Error
}

func (r *boardRepository) FindRolesByBoard(boardID uuid.UUID) ([]domain.BoardRole, error) {
	var roles []domain.BoardRole
	if err := r.db.Where("board_id = ?", boardID).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// FindRolesByBoards fetches board roles for multiple boards in a single query
func (r *boardRepository) FindRolesByBoards(boardIDs []uuid.UUID) (map[uuid.UUID][]domain.BoardRole, error) {
	if len(boardIDs) == 0 {
		return make(map[uuid.UUID][]domain.BoardRole), nil
	}

	var boardRoles []domain.BoardRole
	if err := r.db.Where("board_id IN ?", boardIDs).Find(&boardRoles).Error; err != nil {
		return nil, err
	}

	// Group by board_id
	result := make(map[uuid.UUID][]domain.BoardRole)
	for _, br := range boardRoles {
		result[br.BoardID] = append(result[br.BoardID], br)
	}

	return result, nil
}

func (r *boardRepository) FindBoardsByRole(roleID uuid.UUID) ([]domain.Board, error) {
	var boards []domain.Board
	if err := r.db.Joins("INNER JOIN board_roles ON boards.id = board_roles.board_id").
		Where("board_roles.custom_role_id = ? AND boards.is_deleted = ?", roleID, false).
		Find(&boards).Error; err != nil {
		return nil, err
	}
	return boards, nil
}

// ==================== Phase 4 TODO Implementation ====================

func (r *boardRepository) UpdateBoardsRoleToDefault(oldRoleID, defaultRoleID uuid.UUID) error {
	// Update all board_roles entries using oldRoleID to defaultRoleID
	return r.db.Model(&domain.BoardRole{}).
		Where("custom_role_id = ?", oldRoleID).
		Update("custom_role_id", defaultRoleID).Error
}

func (r *boardRepository) UpdateBoardsStageToDefault(oldStageID, defaultStageID uuid.UUID) error {
	// Update all boards using oldStageID to defaultStageID
	return r.db.Model(&domain.Board{}).
		Where("custom_stage_id = ? AND is_deleted = ?", oldStageID, false).
		Update("custom_stage_id", defaultStageID).Error
}

func (r *boardRepository) UpdateBoardsImportanceToDefault(oldImportanceID, defaultImportanceID uuid.UUID) error {
	// Update all boards using oldImportanceID to defaultImportanceID
	return r.db.Model(&domain.Board{}).
		Where("custom_importance_id = ? AND is_deleted = ?", oldImportanceID, false).
		Update("custom_importance_id", defaultImportanceID).Error
}
