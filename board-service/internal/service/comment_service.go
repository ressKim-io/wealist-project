package service

import (
	"board-service/internal/apperrors"
	"board-service/internal/cache"
	"board-service/internal/client"
	"board-service/internal/domain"
	"board-service/internal/dto"
	"board-service/internal/repository"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CommentService defines the interface for comment business logic.
type CommentService interface {
	CreateComment(ctx context.Context, req dto.CreateCommentRequest, userID uuid.UUID) (*dto.CommentResponse, error)
	GetCommentsByBoardID(ctx context.Context, boardID uuid.UUID, userID uuid.UUID) ([]dto.CommentResponse, error)
	UpdateComment(ctx context.Context, commentID uuid.UUID, req dto.UpdateCommentRequest, userID uuid.UUID) (*dto.CommentResponse, error)
	DeleteComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) error
}

type commentService struct {
	commentRepo   repository.CommentRepository
	boardRepo     repository.BoardRepository
	projectRepo   repository.ProjectRepository
	userClient    client.UserClient
	userInfoCache cache.UserInfoCache
	logger        *zap.Logger
	db            *gorm.DB
}

// NewCommentService creates a new instance of CommentService.
func NewCommentService(cr repository.CommentRepository, kr repository.BoardRepository, pr repository.ProjectRepository, uc client.UserClient, uic cache.UserInfoCache, l *zap.Logger, db *gorm.DB) CommentService {
	return &commentService{
		commentRepo:   cr,
		boardRepo:     kr,
		projectRepo:   pr,
		userClient:    uc,
		userInfoCache: uic,
		logger:        l,
		db:            db,
	}
}

// CreateComment creates a new comment on a board.
func (s *commentService) CreateComment(ctx context.Context, req dto.CreateCommentRequest, userID uuid.UUID) (*dto.CommentResponse, error) {
	board, err := s.boardRepo.FindByID(req.BoardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, fmt.Sprintf("board with id %s not found", req.BoardID), 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to find board", 500)
	}

	_, err = s.projectRepo.FindMemberByUserAndProject(userID, board.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "user is not a member of the project", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to check project membership", 500)
	}

	comment := &domain.Comment{
		BoardID: req.BoardID,
		UserID:   userID,
		Content:  req.Content,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to create comment", 500)
	}

	user := s.getSimpleUserWithCache(ctx, userID.String())

	return &dto.CommentResponse{
		ID:         comment.ID,
		UserID:     comment.UserID,
		UserName:   user.Name,
		UserAvatar: user.AvatarURL,
		Content:    comment.Content,
		CreatedAt:  comment.CreatedAt,
		UpdatedAt:  comment.UpdatedAt,
	}, nil
}

// GetCommentsByBoardID retrieves all comments for a given board.
func (s *commentService) GetCommentsByBoardID(ctx context.Context, boardID uuid.UUID, userID uuid.UUID) ([]dto.CommentResponse, error) {
	board, err := s.boardRepo.FindByID(boardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, fmt.Sprintf("board with id %s not found", boardID), 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to find board", 500)
	}

	_, err = s.projectRepo.FindMemberByUserAndProject(userID, board.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "user is not a member of the project", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to check project membership", 500)
	}

	comments, err := s.commentRepo.FindByBoardID(boardID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to get comments", 500)
	}

	// Batch fetch user info with caching
	userIDs := make([]string, 0, len(comments))
	for _, c := range comments {
		userIDs = append(userIDs, c.UserID.String())
	}

	userMap := s.getSimpleUsersBatch(ctx, userIDs)

	responses := make([]dto.CommentResponse, len(comments))
	for i, c := range comments {
		user, ok := userMap[c.UserID.String()]
		if !ok {
			user = cache.SimpleUser{Name: "Unknown User", AvatarURL: ""}
		}
		responses[i] = dto.CommentResponse{
			ID:         c.ID,
			UserID:     c.UserID,
			UserName:   user.Name,
			UserAvatar: user.AvatarURL,
			Content:    c.Content,
			CreatedAt:  c.CreatedAt,
			UpdatedAt:  c.UpdatedAt,
		}
	}

	return responses, nil
}

// UpdateComment updates an existing comment.
func (s *commentService) UpdateComment(ctx context.Context, commentID uuid.UUID, req dto.UpdateCommentRequest, userID uuid.UUID) (*dto.CommentResponse, error) {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, fmt.Sprintf("comment with id %s not found", commentID), 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to find comment", 500)
	}

	if comment.UserID != userID {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "user does not have permission to update this comment", 403)
	}

	// Domain 메서드 사용: 검증 로직이 Domain에 포함됨
	if err := comment.UpdateContent(req.Content); err != nil {
		// Domain 에러를 Infrastructure 에러로 변환
		return nil, apperrors.FromDomainError(err)
	}

	if err := s.commentRepo.Update(comment); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to update comment", 500)
	}

	user := s.getSimpleUserWithCache(ctx, userID.String())

	return &dto.CommentResponse{
		ID:         comment.ID,
		UserID:     comment.UserID,
		UserName:   user.Name,
		UserAvatar: user.AvatarURL,
		Content:    comment.Content,
		CreatedAt:  comment.CreatedAt,
		UpdatedAt:  comment.UpdatedAt,
	}, nil
}

// DeleteComment deletes a comment.
func (s *commentService) DeleteComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) error {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, fmt.Sprintf("comment with id %s not found", commentID), 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to find comment", 500)
	}

	if comment.UserID != userID {
		// Also allow project owner/admin to delete?
		// For now, only the author can delete.
		return apperrors.New(apperrors.ErrCodeForbidden, "user does not have permission to delete this comment", 403)
	}

	return s.commentRepo.Delete(comment.ID)
}

// getSimpleUserWithCache retrieves simple user info with caching
func (s *commentService) getSimpleUserWithCache(ctx context.Context, userID string) cache.SimpleUser {
	// Try cache first
	cacheExists, cachedUser, err := s.userInfoCache.GetSimpleUser(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get simple user from cache", zap.Error(err))
	}

	if cacheExists && cachedUser != nil {
		return *cachedUser
	}

	// Cache miss - fetch from User Service
	user, err := s.userClient.GetSimpleUser(userID)
	if err != nil {
		s.logger.Error("Failed to get user info for comment", zap.Error(err), zap.String("userID", userID))
		return cache.SimpleUser{Name: "Unknown User", AvatarURL: ""}
	}

	// Cache the result
	cacheUser := &cache.SimpleUser{
		ID:        user.ID,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
	}
	if cacheErr := s.userInfoCache.SetSimpleUser(ctx, cacheUser); cacheErr != nil {
		s.logger.Warn("Failed to cache simple user", zap.Error(cacheErr))
	}

	return *cacheUser
}

// getSimpleUsersBatch retrieves simple user info for multiple users with caching
func (s *commentService) getSimpleUsersBatch(ctx context.Context, userIDs []string) map[string]cache.SimpleUser {
	if len(userIDs) == 0 {
		return make(map[string]cache.SimpleUser)
	}

	// Try to get from cache first
	cachedUsers, err := s.userInfoCache.GetSimpleUsersBatch(ctx, userIDs)
	if err != nil {
		s.logger.Warn("Failed to get users from cache", zap.Error(err))
		cachedUsers = make(map[string]*cache.SimpleUser)
	}

	// Find missing user IDs (not in cache)
	missingUserIDs := []string{}
	for _, userID := range userIDs {
		if _, exists := cachedUsers[userID]; !exists {
			missingUserIDs = append(missingUserIDs, userID)
		}
	}

	// Fetch missing users from User Service
	userMap := make(map[string]cache.SimpleUser)

	if len(missingUserIDs) > 0 {
		users, err := s.userClient.GetSimpleUsers(missingUserIDs)
		if err != nil {
			s.logger.Warn("Failed to fetch users from User Service", zap.Error(err))
		} else {
			// Cache the fetched users
			simpleUsers := make([]cache.SimpleUser, 0, len(users))
			for _, user := range users {
				cacheUser := cache.SimpleUser{
					ID:        user.ID,
					Name:      user.Name,
					AvatarURL: user.AvatarURL,
				}
				userMap[user.ID] = cacheUser
				simpleUsers = append(simpleUsers, cacheUser)
			}
			if cacheErr := s.userInfoCache.SetSimpleUsersBatch(ctx, simpleUsers); cacheErr != nil {
				s.logger.Warn("Failed to cache users", zap.Error(cacheErr))
			}
		}
	}

	// Add cached users to result
	for userID, cachedUser := range cachedUsers {
		if _, exists := userMap[userID]; !exists {
			userMap[userID] = *cachedUser
		}
	}

	return userMap
}
