package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserBoardOrder represents user-specific manual board ordering within views
// Uses fractional indexing for O(1) insert/move operations
type UserBoardOrder struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ViewID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_view_user_board" json:"view_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_view_user_board" json:"user_id"`
	BoardID   uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_view_user_board" json:"board_id"`
	Position  string    `gorm:"type:varchar(255);not null" json:"position"` // Fractional index (e.g., "a0", "a0V", "a1")
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserBoardOrder) TableName() string {
	return "user_board_order"
}
