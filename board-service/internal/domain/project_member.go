package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProjectMember struct {
	BaseModel
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_project_user" json:"project_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_project_user" json:"user_id"`
	RoleID    uuid.UUID `gorm:"type:uuid;not null" json:"role_id"`
	Role      *Role     `gorm:"foreignKey:RoleID;references:ID" json:"role,omitempty"`
	JoinedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"joined_at"`
}

func (ProjectMember) TableName() string {
	return "project_members"
}
