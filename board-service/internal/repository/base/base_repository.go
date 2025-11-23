package base

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity는 모든 도메인 엔티티가 구현해야 하는 인터페이스입니다
// BaseModel을 임베딩하는 모든 엔티티는 자동으로 이 인터페이스를 만족합니다
type Entity interface {
	GetID() uuid.UUID
	SetID(uuid.UUID)
	GetIsDeleted() bool
	SetIsDeleted(bool)
}

// BaseRepository는 공통 CRUD 메서드를 제공하는 제네릭 repository입니다
// T는 Entity 인터페이스를 구현하는 타입이어야 합니다
type BaseRepository[T Entity] interface {
	Create(entity T) error
	FindByID(id uuid.UUID) (T, error)
	FindAll() ([]T, error)
	Update(entity T) error
	Delete(id uuid.UUID) error
	HardDelete(id uuid.UUID) error
	Count() (int64, error)
	Exists(id uuid.UUID) (bool, error)
}

// baseRepository는 BaseRepository 인터페이스의 구현체입니다
type baseRepository[T Entity] struct {
	db *gorm.DB
}

// NewBaseRepository는 새로운 BaseRepository를 생성합니다
func NewBaseRepository[T Entity](db *gorm.DB) BaseRepository[T] {
	return &baseRepository[T]{db: db}
}

// Create는 새로운 엔티티를 생성합니다
func (r *baseRepository[T]) Create(entity T) error {
	return r.db.Create(&entity).Error
}

// FindByID는 ID로 엔티티를 조회합니다 (soft deleted 제외)
func (r *baseRepository[T]) FindByID(id uuid.UUID) (T, error) {
	var entity T
	if err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&entity).Error; err != nil {
		var zero T
		return zero, err
	}
	return entity, nil
}

// FindAll은 모든 엔티티를 조회합니다 (soft deleted 제외)
func (r *baseRepository[T]) FindAll() ([]T, error) {
	var entities []T
	if err := r.db.Where("is_deleted = ?", false).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// Update는 엔티티를 업데이트합니다
func (r *baseRepository[T]) Update(entity T) error {
	return r.db.Save(&entity).Error
}

// Delete는 엔티티를 soft delete 처리합니다
func (r *baseRepository[T]) Delete(id uuid.UUID) error {
	var entity T
	return r.db.Model(&entity).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}

// HardDelete는 엔티티를 실제로 삭제합니다
func (r *baseRepository[T]) HardDelete(id uuid.UUID) error {
	var entity T
	return r.db.Where("id = ?", id).Delete(&entity).Error
}

// Count는 엔티티의 총 개수를 반환합니다 (soft deleted 제외)
func (r *baseRepository[T]) Count() (int64, error) {
	var count int64
	var entity T
	if err := r.db.Model(&entity).Where("is_deleted = ?", false).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Exists는 엔티티가 존재하는지 확인합니다 (soft deleted 제외)
func (r *baseRepository[T]) Exists(id uuid.UUID) (bool, error) {
	var count int64
	var entity T
	if err := r.db.Model(&entity).
		Where("id = ? AND is_deleted = ?", id, false).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ==================== Helper Functions for Query Building ====================

// ScopeSoftDeleted는 soft deleted 레코드를 제외하는 scope입니다
func ScopeSoftDeleted() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_deleted = ?", false)
	}
}

// ScopeWithOrder는 정렬을 추가하는 scope입니다
func ScopeWithOrder(orderBy string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(orderBy)
	}
}

// ScopeWithLimit는 제한을 추가하는 scope입니다
func ScopeWithLimit(limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}

// ScopeWithOffset는 오프셋을 추가하는 scope입니다
func ScopeWithOffset(offset int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset)
	}
}
