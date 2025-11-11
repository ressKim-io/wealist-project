package repository

import (
	"board-service/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectFieldRepository는 ProjectField 엔티티만 관리합니다
type ProjectFieldRepository interface {
	Create(field *domain.ProjectField) error
	FindByID(id uuid.UUID) (*domain.ProjectField, error)
	FindByProject(projectID uuid.UUID) ([]domain.ProjectField, error)
	FindByIDs(ids []uuid.UUID) ([]domain.ProjectField, error)
	Update(field *domain.ProjectField) error
	Delete(id uuid.UUID) error
	UpdateOrder(fieldID uuid.UUID, newOrder int) error
	BatchUpdateOrders(orders map[uuid.UUID]int) error
}

type projectFieldRepository struct {
	db *gorm.DB
}

// NewProjectFieldRepository는 새로운 ProjectFieldRepository를 생성합니다
func NewProjectFieldRepository(db *gorm.DB) ProjectFieldRepository {
	return &projectFieldRepository{db: db}
}

func (r *projectFieldRepository) Create(field *domain.ProjectField) error {
	return r.db.Create(field).Error
}

func (r *projectFieldRepository) FindByID(id uuid.UUID) (*domain.ProjectField, error) {
	var field domain.ProjectField
	if err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&field).Error; err != nil {
		return nil, err
	}
	return &field, nil
}

func (r *projectFieldRepository) FindByProject(projectID uuid.UUID) ([]domain.ProjectField, error) {
	var fields []domain.ProjectField
	if err := r.db.Where("project_id = ? AND is_deleted = ?", projectID, false).
		Order("display_order ASC, created_at ASC").
		Find(&fields).Error; err != nil {
		return nil, err
	}
	return fields, nil
}

func (r *projectFieldRepository) FindByIDs(ids []uuid.UUID) ([]domain.ProjectField, error) {
	var fields []domain.ProjectField
	if err := r.db.Where("id IN ? AND is_deleted = ?", ids, false).
		Order("display_order ASC").
		Find(&fields).Error; err != nil {
		return nil, err
	}
	return fields, nil
}

func (r *projectFieldRepository) Update(field *domain.ProjectField) error {
	return r.db.Save(field).Error
}

func (r *projectFieldRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&domain.ProjectField{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}

func (r *projectFieldRepository) UpdateOrder(fieldID uuid.UUID, newOrder int) error {
	return r.db.Model(&domain.ProjectField{}).
		Where("id = ?", fieldID).
		Update("display_order", newOrder).Error
}

func (r *projectFieldRepository) BatchUpdateOrders(orders map[uuid.UUID]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for fieldID, order := range orders {
			if err := tx.Model(&domain.ProjectField{}).
				Where("id = ?", fieldID).
				Update("display_order", order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
