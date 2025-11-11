package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateCommentRequest defines the structure for creating a new comment.
type CreateCommentRequest struct {
	BoardID uuid.UUID `json:"boardId" binding:"required"`
	Content  string    `json:"content" binding:"required"`
}

// UpdateCommentRequest defines the structure for updating a comment.
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

// CommentResponse defines the structure for a comment response.
type CommentResponse struct {
	ID          uuid.UUID `json:"commentId"`
	UserID      uuid.UUID `json:"userId"`
	UserName    string    `json:"userName"`
	UserAvatar  string    `json:"userAvatar"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
