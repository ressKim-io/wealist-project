package repository

import (
	"board-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomFieldRepository interface {
	// Custom Roles
	CreateCustomRole(role *domain.CustomRole) error
	FindCustomRoleByID(id uuid.UUID) (*domain.CustomRole, error)
	FindCustomRolesByProject(projectID uuid.UUID) ([]domain.CustomRole, error)
	FindCustomRoleByProjectAndName(projectID uuid.UUID, name string) (*domain.CustomRole, error)
	FindCustomRolesByIDs(ids []uuid.UUID) ([]domain.CustomRole, error) // Batch
	UpdateCustomRole(role *domain.CustomRole) error
	DeleteCustomRole(id uuid.UUID) error
	UpdateCustomRoleOrders(roles []domain.CustomRole) error

	// Custom Stages
	CreateCustomStage(stage *domain.CustomStage) error
	FindCustomStageByID(id uuid.UUID) (*domain.CustomStage, error)
	FindCustomStagesByProject(projectID uuid.UUID) ([]domain.CustomStage, error)
	FindCustomStageByProjectAndName(projectID uuid.UUID, name string) (*domain.CustomStage, error)
	FindCustomStagesByIDs(ids []uuid.UUID) ([]domain.CustomStage, error) // Batch
	UpdateCustomStage(stage *domain.CustomStage) error
	DeleteCustomStage(id uuid.UUID) error
	UpdateCustomStageOrders(stages []domain.CustomStage) error

	// Custom Importance
	CreateCustomImportance(importance *domain.CustomImportance) error
	FindCustomImportanceByID(id uuid.UUID) (*domain.CustomImportance, error)
	FindCustomImportancesByProject(projectID uuid.UUID) ([]domain.CustomImportance, error)
	FindCustomImportanceByProjectAndName(projectID uuid.UUID, name string) (*domain.CustomImportance, error)
	FindCustomImportancesByIDs(ids []uuid.UUID) ([]domain.CustomImportance, error) // Batch
	UpdateCustomImportance(importance *domain.CustomImportance) error
	DeleteCustomImportance(id uuid.UUID) error
	UpdateCustomImportanceOrders(importances []domain.CustomImportance) error
}

type customFieldRepository struct {
	db *gorm.DB
}

func NewCustomFieldRepository(db *gorm.DB) CustomFieldRepository {
	return &customFieldRepository{db: db}
}

// ==================== Custom Roles ====================

func (r *customFieldRepository) CreateCustomRole(role *domain.CustomRole) error {
	return r.db.Create(role).Error
}

func (r *customFieldRepository) FindCustomRoleByID(id uuid.UUID) (*domain.CustomRole, error) {
	var role domain.CustomRole
	if err := r.db.Where("id = ?", id).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *customFieldRepository) FindCustomRolesByProject(projectID uuid.UUID) ([]domain.CustomRole, error) {
	var roles []domain.CustomRole
	if err := r.db.Where("project_id = ?", projectID).
		Order("display_order ASC").
		Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *customFieldRepository) FindCustomRoleByProjectAndName(projectID uuid.UUID, name string) (*domain.CustomRole, error) {
	var role domain.CustomRole
	if err := r.db.Where("project_id = ? AND name = ?", projectID, name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *customFieldRepository) UpdateCustomRole(role *domain.CustomRole) error {
	return r.db.Save(role).Error
}

func (r *customFieldRepository) DeleteCustomRole(id uuid.UUID) error {
	return r.db.Delete(&domain.CustomRole{}, id).Error
}

func (r *customFieldRepository) UpdateCustomRoleOrders(roles []domain.CustomRole) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, role := range roles {
			if err := tx.Model(&domain.CustomRole{}).
				Where("id = ?", role.ID).
				Update("display_order", role.DisplayOrder).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== Custom Stages ====================

func (r *customFieldRepository) CreateCustomStage(stage *domain.CustomStage) error {
	return r.db.Create(stage).Error
}

func (r *customFieldRepository) FindCustomStageByID(id uuid.UUID) (*domain.CustomStage, error) {
	var stage domain.CustomStage
	if err := r.db.Where("id = ?", id).First(&stage).Error; err != nil {
		return nil, err
	}
	return &stage, nil
}

func (r *customFieldRepository) FindCustomStagesByProject(projectID uuid.UUID) ([]domain.CustomStage, error) {
	var stages []domain.CustomStage
	if err := r.db.Where("project_id = ?", projectID).
		Order("display_order ASC").
		Find(&stages).Error; err != nil {
		return nil, err
	}
	return stages, nil
}

func (r *customFieldRepository) FindCustomStageByProjectAndName(projectID uuid.UUID, name string) (*domain.CustomStage, error) {
	var stage domain.CustomStage
	if err := r.db.Where("project_id = ? AND name = ?", projectID, name).First(&stage).Error; err != nil {
		return nil, err
	}
	return &stage, nil
}

func (r *customFieldRepository) UpdateCustomStage(stage *domain.CustomStage) error {
	return r.db.Save(stage).Error
}

func (r *customFieldRepository) DeleteCustomStage(id uuid.UUID) error {
	return r.db.Delete(&domain.CustomStage{}, id).Error
}

func (r *customFieldRepository) UpdateCustomStageOrders(stages []domain.CustomStage) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, stage := range stages {
			if err := tx.Model(&domain.CustomStage{}).
				Where("id = ?", stage.ID).
				Update("display_order", stage.DisplayOrder).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== Custom Importance ====================

func (r *customFieldRepository) CreateCustomImportance(importance *domain.CustomImportance) error {
	return r.db.Create(importance).Error
}

func (r *customFieldRepository) FindCustomImportanceByID(id uuid.UUID) (*domain.CustomImportance, error) {
	var importance domain.CustomImportance
	if err := r.db.Where("id = ?", id).First(&importance).Error; err != nil {
		return nil, err
	}
	return &importance, nil
}

func (r *customFieldRepository) FindCustomImportancesByProject(projectID uuid.UUID) ([]domain.CustomImportance, error) {
	var importances []domain.CustomImportance
	if err := r.db.Where("project_id = ?", projectID).
		Order("display_order ASC").
		Find(&importances).Error; err != nil {
		return nil, err
	}
	return importances, nil
}

func (r *customFieldRepository) FindCustomImportanceByProjectAndName(projectID uuid.UUID, name string) (*domain.CustomImportance, error) {
	var importance domain.CustomImportance
	if err := r.db.Where("project_id = ? AND name = ?", projectID, name).First(&importance).Error; err != nil {
		return nil, err
	}
	return &importance, nil
}

func (r *customFieldRepository) UpdateCustomImportance(importance *domain.CustomImportance) error {
	return r.db.Save(importance).Error
}

func (r *customFieldRepository) DeleteCustomImportance(id uuid.UUID) error {
	return r.db.Delete(&domain.CustomImportance{}, id).Error
}

func (r *customFieldRepository) UpdateCustomImportanceOrders(importances []domain.CustomImportance) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, importance := range importances {
			if err := tx.Model(&domain.CustomImportance{}).
				Where("id = ?", importance.ID).
				Update("display_order", importance.DisplayOrder).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== Batch Methods ====================

// FindCustomRolesByIDs fetches multiple roles by IDs in a single query
func (r *customFieldRepository) FindCustomRolesByIDs(ids []uuid.UUID) ([]domain.CustomRole, error) {
	if len(ids) == 0 {
		return []domain.CustomRole{}, nil
	}

	var roles []domain.CustomRole
	if err := r.db.Where("id IN ?", ids).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// FindCustomStagesByIDs fetches multiple stages by IDs in a single query
func (r *customFieldRepository) FindCustomStagesByIDs(ids []uuid.UUID) ([]domain.CustomStage, error) {
	if len(ids) == 0 {
		return []domain.CustomStage{}, nil
	}

	var stages []domain.CustomStage
	if err := r.db.Where("id IN ?", ids).Find(&stages).Error; err != nil {
		return nil, err
	}
	return stages, nil
}

// FindCustomImportancesByIDs fetches multiple importances by IDs in a single query
func (r *customFieldRepository) FindCustomImportancesByIDs(ids []uuid.UUID) ([]domain.CustomImportance, error) {
	if len(ids) == 0 {
		return []domain.CustomImportance{}, nil
	}

	var importances []domain.CustomImportance
	if err := r.db.Where("id IN ?", ids).Find(&importances).Error; err != nil {
		return nil, err
	}
	return importances, nil
}
