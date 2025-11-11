package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Comment represents a comment on a Board card.
type Comment struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Content   string         `gorm:"type:text;not null"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	BoardID   uuid.UUID      `gorm:"type:uuid;not null;index"`
	Board     Board          `gorm:"foreignKey:BoardID"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for the Comment model.
func (Comment) TableName() string {
	return "comments"
}

// BeforeCreate is a GORM hook that generates a UUID before creating a record.
func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// ==================== Rich Domain Model - Business Methods ====================

// IsWrittenBy returns true if the comment was written by the given user
func (c *Comment) IsWrittenBy(userID uuid.UUID) bool {
	return c.UserID == userID
}

// BelongsToBoard returns true if the comment belongs to the given board
func (c *Comment) BelongsToBoard(boardID uuid.UUID) bool {
	return c.BoardID == boardID
}

// UpdateContent updates the comment content with validation
func (c *Comment) UpdateContent(content string) error {
	if content == "" {
		return &ValidationError{Field: "content", Message: "댓글 내용은 필수입니다"}
	}
	if len(content) > 10000 {
		return &ValidationError{Field: "content", Message: "댓글 내용은 10000자를 초과할 수 없습니다"}
	}
	c.Content = content
	c.UpdatedAt = time.Now()
	return nil
}

// IsEmpty returns true if the comment content is empty
func (c *Comment) IsEmpty() bool {
	return c.Content == ""
}

// GetAge returns the duration since the comment was created
func (c *Comment) GetAge() time.Duration {
	return time.Since(c.CreatedAt)
}

// IsRecent returns true if the comment was created within the last 24 hours
func (c *Comment) IsRecent() bool {
	return c.GetAge() < 24*time.Hour
}

// WasEdited returns true if the comment was edited after creation
func (c *Comment) WasEdited() bool {
	return c.UpdatedAt.After(c.CreatedAt.Add(time.Second))
}
