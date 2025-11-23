package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Board represents a work board entity within a project
type Board struct {
	BaseModel
	ProjectID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_boards_project_id" json:"project_id"`
	AuthorID     uuid.UUID      `gorm:"type:uuid;not null;index:idx_boards_author_id" json:"author_id"`
	AssigneeID   *uuid.UUID     `gorm:"type:uuid;index:idx_boards_assignee_id" json:"assignee_id"`
	Title        string         `gorm:"type:varchar(255);not null" json:"title"`
	Content      string         `gorm:"type:text" json:"content"`
	CustomFields datatypes.JSON `gorm:"type:jsonb" json:"custom_fields"`
	StartDate    *time.Time     `gorm:"type:timestamp;index:idx_boards_start_date" json:"start_date"`
	DueDate      *time.Time     `gorm:"type:timestamp;index:idx_boards_due_date" json:"due_date"`
	Project      Project        `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
	Participants []Participant  `gorm:"foreignKey:BoardID;constraint:OnDelete:CASCADE" json:"participants,omitempty"`
	Comments     []Comment      `gorm:"foreignKey:BoardID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
	// ✅ 수정: Attachments는 다형성 관계이므로 FK 제거, Repository에서 별도 조회
	Attachments []Attachment `gorm:"-" json:"attachments,omitempty"`
}

// TableName specifies the table name for Board
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
