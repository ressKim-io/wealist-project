package domain

import (
	"time"

	"github.com/google/uuid"
)

type Board struct {
	BaseModel
	ProjectID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	Title              string     `gorm:"type:varchar(255);not null" json:"title"`
	Description        string     `gorm:"type:text" json:"description"`
	AssigneeID         *uuid.UUID `gorm:"type:uuid;index" json:"assignee_id"`
	CreatedBy          uuid.UUID  `gorm:"type:uuid;not null;index" json:"created_by"`
	DueDate            *time.Time `gorm:"index" json:"due_date"`

	// Custom fields cache (JSONB for fast filtering with GIN index)
	// All custom fields (stages, roles, importance, etc.) are stored here
	CustomFieldsCache  string     `gorm:"type:jsonb;default:'{}'" json:"custom_fields_cache"`
}

func (Board) TableName() string {
	return "boards"
}

// ==================== Rich Domain Model - Business Methods ====================

// IsOverdue returns true if the board has a due date and it's in the past
func (b *Board) IsOverdue() bool {
	if b.DueDate == nil {
		return false
	}
	return b.DueDate.Before(time.Now())
}

// IsAssigned returns true if the board has an assignee
func (b *Board) IsAssigned() bool {
	return b.AssigneeID != nil
}

// Assign assigns the board to a user
func (b *Board) Assign(userID uuid.UUID) {
	b.AssigneeID = &userID
	b.UpdatedAt = time.Now()
}

// Unassign removes the assignee from the board
func (b *Board) Unassign() {
	b.AssigneeID = nil
	b.UpdatedAt = time.Now()
}

// UpdateTitle updates the board title with validation
func (b *Board) UpdateTitle(title string) error {
	if title == "" {
		return NewValidationError("title", "제목은 필수입니다")
	}
	if len(title) > 255 {
		return NewValidationError("title", "제목은 255자를 초과할 수 없습니다")
	}
	b.Title = title
	b.UpdatedAt = time.Now()
	return nil
}

// UpdateDescription updates the board description
func (b *Board) UpdateDescription(description string) {
	b.Description = description
	b.UpdatedAt = time.Now()
}

// SetDueDate sets the due date for the board
func (b *Board) SetDueDate(dueDate time.Time) {
	b.DueDate = &dueDate
	b.UpdatedAt = time.Now()
}

// ClearDueDate removes the due date from the board
func (b *Board) ClearDueDate() {
	b.DueDate = nil
	b.UpdatedAt = time.Now()
}

// IsCreatedBy returns true if the board was created by the given user
func (b *Board) IsCreatedBy(userID uuid.UUID) bool {
	return b.CreatedBy == userID
}

// MarkAsDeleted marks the board as deleted (soft delete)
func (b *Board) MarkAsDeleted() {
	b.IsDeleted = true
	b.UpdatedAt = time.Now()
}
