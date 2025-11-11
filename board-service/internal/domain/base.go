package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel contains common fields for all domain models
type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	IsDeleted bool      `gorm:"default:false;index" json:"is_deleted"`
}

// BeforeCreate is a GORM hook that generates UUID before creating a record
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// ==================== Entity Interface Implementation ====================
// 이 메서드들은 generic repository 패턴을 위해 필요합니다

// GetID returns the entity ID
func (b *BaseModel) GetID() uuid.UUID {
	return b.ID
}

// SetID sets the entity ID
func (b *BaseModel) SetID(id uuid.UUID) {
	b.ID = id
}

// GetIsDeleted returns the soft delete flag
func (b *BaseModel) GetIsDeleted() bool {
	return b.IsDeleted
}

// SetIsDeleted sets the soft delete flag
func (b *BaseModel) SetIsDeleted(deleted bool) {
	b.IsDeleted = deleted
}
