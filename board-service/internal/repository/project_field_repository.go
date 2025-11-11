package repository

import (
	"board-service/internal/domain"
	"board-service/internal/repository/base"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectFieldRepository는 ProjectField 엔티티만 관리합니다
type ProjectFieldRepository interface {
	// 공통 CRUD 메서드 (base repository에서 제공)
	Create(field *domain.ProjectField) error
	FindByID(id uuid.UUID) (*domain.ProjectField, error)
	Update(field *domain.ProjectField) error
	Delete(id uuid.UUID) error

	// ProjectField 전용 메서드
	FindByProject(projectID uuid.UUID) ([]domain.ProjectField, error)
	FindByIDs(ids []uuid.UUID) ([]domain.ProjectField, error)
	UpdateOrder(fieldID uuid.UUID, newOrder int) error
	BatchUpdateOrders(orders map[uuid.UUID]int) error
}

type projectFieldRepository struct {
	base.BaseRepository[*domain.ProjectField]
	db *gorm.DB
}

// NewProjectFieldRepository는 새로운 ProjectFieldRepository를 생성합니다
func NewProjectFieldRepository(db *gorm.DB) ProjectFieldRepository {
	return &projectFieldRepository{
		BaseRepository: base.NewBaseRepository[*domain.ProjectField](db),
		db:             db,
	}
}

// ==================== 공통 CRUD는 base repository에 위임 ====================
// Create, FindByID, Update, Delete는 BaseRepository의 구현을 사용합니다

// ==================== ProjectField 전용 메서드 ====================

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
