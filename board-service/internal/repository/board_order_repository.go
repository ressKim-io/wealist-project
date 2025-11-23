package repository

import (
	"board-service/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BoardOrderRepository는 UserBoardOrder 엔티티만 관리합니다
// Fractional indexing을 사용한 O(1) 순서 변경을 지원합니다
// UPSERT와 hard delete를 사용하므로 base repository를 사용하지 않습니다
type BoardOrderRepository interface {
	Set(order *domain.UserBoardOrder) error
	FindByView(viewID, userID uuid.UUID) ([]domain.UserBoardOrder, error)
	BatchUpdate(orders []domain.UserBoardOrder) error
	Delete(viewID, userID, boardID uuid.UUID) error
}

type boardOrderRepository struct {
	db *gorm.DB
}

// NewBoardOrderRepository는 새로운 BoardOrderRepository를 생성합니다
func NewBoardOrderRepository(db *gorm.DB) BoardOrderRepository {
	return &boardOrderRepository{
		db: db,
	}
}

// Set은 UPSERT를 사용하여 보드 순서를 설정합니다 (PostgreSQL ON CONFLICT DO UPDATE)
func (r *boardOrderRepository) Set(order *domain.UserBoardOrder) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "view_id"}, {Name: "user_id"}, {Name: "board_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"position", "updated_at"}),
	}).Create(order).Error
}

func (r *boardOrderRepository) FindByView(viewID, userID uuid.UUID) ([]domain.UserBoardOrder, error) {
	var orders []domain.UserBoardOrder
	if err := r.db.Where("view_id = ? AND user_id = ?", viewID, userID).
		Order("position ASC"). // Fractional indexing: 사전순 정렬
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *boardOrderRepository) BatchUpdate(orders []domain.UserBoardOrder) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, order := range orders {
			if err := tx.Save(&order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *boardOrderRepository) Delete(viewID, userID, boardID uuid.UUID) error {
	return r.db.Where("view_id = ? AND user_id = ? AND board_id = ?", viewID, userID, boardID).
		Delete(&domain.UserBoardOrder{}).Error
}
