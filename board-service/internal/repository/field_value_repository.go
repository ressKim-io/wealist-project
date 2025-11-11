package repository

import (
	"board-service/internal/domain"
	"board-service/internal/repository/base"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FieldValueRepository는 BoardFieldValue 엔티티만 관리합니다
// BoardFieldValue는 특수한 UPSERT 로직이 많아 base repository를 부분적으로만 사용합니다
type FieldValueRepository interface {
	Set(value *domain.BoardFieldValue) error
	FindByBoard(boardID uuid.UUID) ([]domain.BoardFieldValue, error)
	FindByBoardAndField(boardID, fieldID uuid.UUID) ([]domain.BoardFieldValue, error)
	FindByBoards(boardIDs []uuid.UUID) (map[uuid.UUID][]domain.BoardFieldValue, error)
	Delete(boardID, fieldID uuid.UUID) error
	DeleteByID(id uuid.UUID) error
	BatchSet(values []domain.BoardFieldValue) error
	BatchDelete(boardID, fieldID uuid.UUID) error
	UpdateBoardCache(boardID uuid.UUID) (string, error) // JSON 캐시 업데이트
}

type fieldValueRepository struct {
	base.BaseRepository[*domain.BoardFieldValue]
	db *gorm.DB
}

// NewFieldValueRepository는 새로운 FieldValueRepository를 생성합니다
func NewFieldValueRepository(db *gorm.DB) FieldValueRepository {
	return &fieldValueRepository{
		BaseRepository: base.NewBaseRepository[*domain.BoardFieldValue](db),
		db:             db,
	}
}

func (r *fieldValueRepository) Set(value *domain.BoardFieldValue) error {
	return r.db.Save(value).Error
}

func (r *fieldValueRepository) FindByBoard(boardID uuid.UUID) ([]domain.BoardFieldValue, error) {
	var values []domain.BoardFieldValue
	if err := r.db.Where("board_id = ? AND is_deleted = ?", boardID, false).
		Order("field_id, display_order ASC").
		Find(&values).Error; err != nil {
		return nil, err
	}
	return values, nil
}

func (r *fieldValueRepository) FindByBoardAndField(boardID, fieldID uuid.UUID) ([]domain.BoardFieldValue, error) {
	var values []domain.BoardFieldValue
	if err := r.db.Where("board_id = ? AND field_id = ? AND is_deleted = ?", boardID, fieldID, false).
		Order("display_order ASC").
		Find(&values).Error; err != nil {
		return nil, err
	}
	return values, nil
}

func (r *fieldValueRepository) FindByBoards(boardIDs []uuid.UUID) (map[uuid.UUID][]domain.BoardFieldValue, error) {
	var values []domain.BoardFieldValue
	if err := r.db.Where("board_id IN ? AND is_deleted = ?", boardIDs, false).
		Order("board_id, field_id, display_order ASC").
		Find(&values).Error; err != nil {
		return nil, err
	}

	// 보드 ID별로 그룹화
	result := make(map[uuid.UUID][]domain.BoardFieldValue)
	for _, value := range values {
		result[value.BoardID] = append(result[value.BoardID], value)
	}
	return result, nil
}

func (r *fieldValueRepository) Delete(boardID, fieldID uuid.UUID) error {
	return r.db.Model(&domain.BoardFieldValue{}).
		Where("board_id = ? AND field_id = ?", boardID, fieldID).
		Update("is_deleted", true).Error
}

func (r *fieldValueRepository) DeleteByID(id uuid.UUID) error {
	return r.db.Model(&domain.BoardFieldValue{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}

func (r *fieldValueRepository) BatchSet(values []domain.BoardFieldValue) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, value := range values {
			if err := tx.Save(&value).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *fieldValueRepository) BatchDelete(boardID, fieldID uuid.UUID) error {
	return r.db.Model(&domain.BoardFieldValue{}).
		Where("board_id = ? AND field_id = ?", boardID, fieldID).
		Update("is_deleted", true).Error
}

// UpdateBoardCache는 보드의 custom_fields_cache를 업데이트합니다
func (r *fieldValueRepository) UpdateBoardCache(boardID uuid.UUID) (string, error) {
	// 서비스 레이어에서 구현될 예정 (JSON 마샬링 필요)
	return "{}", nil
}
