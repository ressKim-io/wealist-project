package domain

import (
	"time"

	"github.com/google/uuid"
)

// BaseModel contains common fields for all domain entities
type BaseModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time  `gorm:"not null" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"` // Soft delete
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
