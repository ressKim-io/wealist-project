package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserBoardOrder represents user-specific manual board ordering within views
type UserBoardOrder struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ViewID       uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_view_user_board" json:"view_id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_view_user_board" json:"user_id"`
	BoardID      uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_view_user_board" json:"board_id"`
	DisplayOrder int       `gorm:"not null" json:"display_order"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserBoardOrder) TableName() string {
	return "user_board_order"
}
