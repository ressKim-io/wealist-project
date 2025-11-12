package repository

import (
	"board-service/internal/domain"
	"board-service/internal/repository/base"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FieldOptionRepository는 FieldOption 엔티티만 관리합니다
type FieldOptionRepository interface {
	// 공통 CRUD 메서드 (base repository에서 제공)
	Create(option *domain.FieldOption) error
	FindByID(id uuid.UUID) (*domain.FieldOption, error)
	Update(option *domain.FieldOption) error
	Delete(id uuid.UUID) error

	// FieldOption 전용 메서드
	FindByField(fieldID uuid.UUID) ([]domain.FieldOption, error)
	FindByIDs(ids []uuid.UUID) ([]domain.FieldOption, error)
	UpdateOrder(optionID uuid.UUID, newOrder int) error
	BatchUpdateOrders(orders map[uuid.UUID]int) error
}

type fieldOptionRepository struct {
	base.BaseRepository[*domain.FieldOption]
	db *gorm.DB
}

// NewFieldOptionRepository는 새로운 FieldOptionRepository를 생성합니다
func NewFieldOptionRepository(db *gorm.DB) FieldOptionRepository {
	return &fieldOptionRepository{
		BaseRepository: base.NewBaseRepository[*domain.FieldOption](db),
		db:             db,
	}
}

// ==================== 공통 CRUD는 base repository에 위임 ====================
// Create, FindByID, Update, Delete는 BaseRepository의 구현을 사용합니다

// ==================== FieldOption 전용 메서드 ====================

func (r *fieldOptionRepository) FindByField(fieldID uuid.UUID) ([]domain.FieldOption, error) {
	var options []domain.FieldOption
	if err := r.db.Where("field_id = ? AND is_deleted = ?", fieldID, false).
		Order("display_order ASC, created_at ASC").
		Find(&options).Error; err != nil {
		return nil, err
	}
	return options, nil
}

func (r *fieldOptionRepository) FindByIDs(ids []uuid.UUID) ([]domain.FieldOption, error) {
	var options []domain.FieldOption
	if err := r.db.Where("id IN ? AND is_deleted = ?", ids, false).
		Order("display_order ASC").
		Find(&options).Error; err != nil {
		return nil, err
	}
	return options, nil
}

func (r *fieldOptionRepository) UpdateOrder(optionID uuid.UUID, newOrder int) error {
	return r.db.Model(&domain.FieldOption{}).
		Where("id = ?", optionID).
		Update("display_order", newOrder).Error
}

func (r *fieldOptionRepository) BatchUpdateOrders(orders map[uuid.UUID]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for optionID, order := range orders {
			if err := tx.Model(&domain.FieldOption{}).
				Where("id = ?", optionID).
				Update("display_order", order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
