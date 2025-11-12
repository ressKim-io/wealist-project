package domain

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	BaseModel
	WorkspaceID uuid.UUID `gorm:"type:uuid;not null;index" json:"workspace_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	OwnerID     uuid.UUID `gorm:"type:uuid;not null;index" json:"owner_id"`
	IsPublic    bool      `gorm:"default:false" json:"is_public"`
}

func (Project) TableName() string {
	return "projects"
}

// ==================== Rich Domain Model - Business Methods ====================

// IsOwnedBy returns true if the project is owned by the given user
func (p *Project) IsOwnedBy(userID uuid.UUID) bool {
	return p.OwnerID == userID
}

// IsPublicProject returns true if the project is public
func (p *Project) IsPublicProject() bool {
	return p.IsPublic
}

// MakePublic makes the project public
func (p *Project) MakePublic() {
	p.IsPublic = true
	p.UpdatedAt = time.Now()
}

// MakePrivate makes the project private
func (p *Project) MakePrivate() {
	p.IsPublic = false
	p.UpdatedAt = time.Now()
}

// UpdateName updates the project name with validation
func (p *Project) UpdateName(name string) error {
	if name == "" {
		return NewValidationError("name", "프로젝트 이름은 필수입니다")
	}
	if len(name) > 255 {
		return NewValidationError("name", "프로젝트 이름은 255자를 초과할 수 없습니다")
	}
	p.Name = name
	p.UpdatedAt = time.Now()
	return nil
}

// UpdateDescription updates the project description
func (p *Project) UpdateDescription(description string) {
	p.Description = description
	p.UpdatedAt = time.Now()
}

// TransferOwnership transfers the project ownership to another user
func (p *Project) TransferOwnership(newOwnerID uuid.UUID) {
	p.OwnerID = newOwnerID
	p.UpdatedAt = time.Now()
}

// BelongsToWorkspace returns true if the project belongs to the given workspace
func (p *Project) BelongsToWorkspace(workspaceID uuid.UUID) bool {
	return p.WorkspaceID == workspaceID
}

// MarkAsDeleted marks the project as deleted (soft delete)
func (p *Project) MarkAsDeleted() {
	p.IsDeleted = true
	p.UpdatedAt = time.Now()
}
