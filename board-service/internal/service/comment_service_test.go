package service

import (
	"board-service/internal/cache"
	"board-service/internal/client"
	"board-service/internal/domain"
	"board-service/internal/dto"
	"board-service/internal/testutil"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ==================== Mock UserClient ====================

type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) GetUser(ctx context.Context, userID string) (*client.UserInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.UserInfo), args.Error(1)
}

func (m *MockUserClient) GetUsersBatch(ctx context.Context, userIDs []string) ([]client.UserInfo, error) {
	args := m.Called(ctx, userIDs)
	return args.Get(0).([]client.UserInfo), args.Error(1)
}

func (m *MockUserClient) SearchUsers(ctx context.Context, query string) ([]client.UserInfo, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]client.UserInfo), args.Error(1)
}

func (m *MockUserClient) GetSimpleUser(userID string) (*client.SimpleUser, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.SimpleUser), args.Error(1)
}

func (m *MockUserClient) GetSimpleUsers(userIDs []string) ([]client.SimpleUser, error) {
	args := m.Called(userIDs)
	return args.Get(0).([]client.SimpleUser), args.Error(1)
}

func (m *MockUserClient) CheckWorkspaceExists(ctx context.Context, workspaceID string, token string) (bool, error) {
	args := m.Called(ctx, workspaceID, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserClient) ValidateWorkspaceMembership(ctx context.Context, workspaceID string, userID string, token string) (bool, error) {
	args := m.Called(ctx, workspaceID, userID, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserClient) GetWorkspace(ctx context.Context, workspaceID string, token string) (*client.WorkspaceInfo, error) {
	args := m.Called(ctx, workspaceID, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.WorkspaceInfo), args.Error(1)
}

// ==================== Mock UserInfoCache ====================

type MockUserInfoCache struct {
	mock.Mock
}

func (m *MockUserInfoCache) GetUserInfo(ctx context.Context, userID string) (bool, *cache.UserInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(1) == nil {
		return args.Bool(0), nil, args.Error(2)
	}
	return args.Bool(0), args.Get(1).(*cache.UserInfo), args.Error(2)
}

func (m *MockUserInfoCache) SetUserInfo(ctx context.Context, userInfo *cache.UserInfo) error {
	args := m.Called(ctx, userInfo)
	return args.Error(0)
}

func (m *MockUserInfoCache) GetSimpleUser(ctx context.Context, userID string) (bool, *cache.SimpleUser, error) {
	args := m.Called(ctx, userID)
	if args.Get(1) == nil {
		return args.Bool(0), nil, args.Error(2)
	}
	return args.Bool(0), args.Get(1).(*cache.SimpleUser), args.Error(2)
}

func (m *MockUserInfoCache) SetSimpleUser(ctx context.Context, simpleUser *cache.SimpleUser) error {
	args := m.Called(ctx, simpleUser)
	return args.Error(0)
}

func (m *MockUserInfoCache) GetSimpleUsersBatch(ctx context.Context, userIDs []string) (map[string]*cache.SimpleUser, error) {
	args := m.Called(ctx, userIDs)
	return args.Get(0).(map[string]*cache.SimpleUser), args.Error(1)
}

func (m *MockUserInfoCache) SetSimpleUsersBatch(ctx context.Context, simpleUsers []cache.SimpleUser) error {
	args := m.Called(ctx, simpleUsers)
	return args.Error(0)
}

func (m *MockUserInfoCache) DeleteUserInfo(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserInfoCache) DeleteSimpleUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// ==================== Test Suite Setup ====================

type CommentServiceTestSuite struct {
	commentRepo   *testutil.MockCommentRepository
	boardRepo     *testutil.MockBoardRepository
	projectRepo   *testutil.MockProjectRepository
	userClient    *MockUserClient
	userInfoCache *MockUserInfoCache
	logger        *zap.Logger
	service       CommentService
}

func setupCommentServiceTest(t *testing.T) *CommentServiceTestSuite {
	commentRepo := new(testutil.MockCommentRepository)
	boardRepo := new(testutil.MockBoardRepository)
	projectRepo := new(testutil.MockProjectRepository)
	userClient := new(MockUserClient)
	userInfoCache := new(MockUserInfoCache)
	logger := zap.NewNop()

	service := NewCommentService(
		commentRepo,
		boardRepo,
		projectRepo,
		userClient,
		userInfoCache,
		logger,
		nil, // db not used in unit tests
	)

	return &CommentServiceTestSuite{
		commentRepo:   commentRepo,
		boardRepo:     boardRepo,
		projectRepo:   projectRepo,
		userClient:    userClient,
		userInfoCache: userInfoCache,
		logger:        logger,
		service:       service,
	}
}

// ==================== CreateComment Tests ====================

func TestCommentService_CreateComment_Success(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: Valid comment request
	ctx := context.Background()
	userID := uuid.New()
	boardID := uuid.New()
	projectID := uuid.New()
	content := "This is a test comment"

	req := dto.CreateCommentRequest{
		BoardID: boardID,
		Content: content,
	}

	board := &domain.Board{
		ID:        boardID,
		ProjectID: projectID,
		Title:     "Test Board",
	}

	member := &domain.ProjectMember{
		UserID:    userID,
		ProjectID: projectID,
	}

	simpleUser := &cache.SimpleUser{
		ID:        userID.String(),
		Name:      "Test User",
		AvatarURL: "http://example.com/avatar.jpg",
	}

	// Mock setup
	suite.boardRepo.On("FindByID", boardID).Return(board, nil)
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).Return(member, nil)
	suite.commentRepo.On("Create", mock.AnythingOfType("*domain.Comment")).Return(nil)
	suite.userInfoCache.On("GetSimpleUser", ctx, userID.String()).Return(false, (*cache.SimpleUser)(nil), nil)
	suite.userClient.On("GetSimpleUser", userID.String()).Return(&client.SimpleUser{
		ID:        userID.String(),
		Name:      "Test User",
		AvatarURL: "http://example.com/avatar.jpg",
	}, nil)
	suite.userInfoCache.On("SetSimpleUser", ctx, mock.AnythingOfType("*cache.SimpleUser")).Return(nil)

	// When: Create comment
	result, err := suite.service.CreateComment(ctx, req, userID)

	// Then: Verify success
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, content, result.Content)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "Test User", result.UserName)

	suite.boardRepo.AssertExpectations(t)
	suite.projectRepo.AssertExpectations(t)
	suite.commentRepo.AssertExpectations(t)
}

func TestCommentService_CreateComment_BoardNotFound(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: Non-existent board
	ctx := context.Background()
	userID := uuid.New()
	boardID := uuid.New()

	req := dto.CreateCommentRequest{
		BoardID: boardID,
		Content: "Test comment",
	}

	suite.boardRepo.On("FindByID", boardID).Return((*domain.Board)(nil), gorm.ErrRecordNotFound)

	// When: Create comment
	result, err := suite.service.CreateComment(ctx, req, userID)

	// Then: Verify error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")

	suite.boardRepo.AssertExpectations(t)
}

func TestCommentService_CreateComment_UserNotProjectMember(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: User is not a project member
	ctx := context.Background()
	userID := uuid.New()
	boardID := uuid.New()
	projectID := uuid.New()

	req := dto.CreateCommentRequest{
		BoardID: boardID,
		Content: "Test comment",
	}

	board := &domain.Board{
		ID:        boardID,
		ProjectID: projectID,
		Title:     "Test Board",
	}

	suite.boardRepo.On("FindByID", boardID).Return(board, nil)
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).Return((*domain.ProjectMember)(nil), gorm.ErrRecordNotFound)

	// When: Create comment
	result, err := suite.service.CreateComment(ctx, req, userID)

	// Then: Verify forbidden error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not a member")

	suite.boardRepo.AssertExpectations(t)
	suite.projectRepo.AssertExpectations(t)
}

func TestCommentService_CreateComment_RepositoryError(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: Repository error
	ctx := context.Background()
	userID := uuid.New()
	boardID := uuid.New()
	projectID := uuid.New()

	req := dto.CreateCommentRequest{
		BoardID: boardID,
		Content: "Test comment",
	}

	board := &domain.Board{
		ID:        boardID,
		ProjectID: projectID,
		Title:     "Test Board",
	}

	member := &domain.ProjectMember{
		UserID:    userID,
		ProjectID: projectID,
	}

	suite.boardRepo.On("FindByID", boardID).Return(board, nil)
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).Return(member, nil)
	suite.commentRepo.On("Create", mock.AnythingOfType("*domain.Comment")).Return(errors.New("database error"))

	// When: Create comment
	result, err := suite.service.CreateComment(ctx, req, userID)

	// Then: Verify error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create comment")

	suite.boardRepo.AssertExpectations(t)
	suite.projectRepo.AssertExpectations(t)
	suite.commentRepo.AssertExpectations(t)
}

// ==================== GetCommentsByBoardID Tests ====================

func TestCommentService_GetCommentsByBoardID_Success(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: Board with comments
	ctx := context.Background()
	userID := uuid.New()
	boardID := uuid.New()
	projectID := uuid.New()
	commentUserID := uuid.New()

	board := &domain.Board{
		ID:        boardID,
		ProjectID: projectID,
		Title:     "Test Board",
	}

	member := &domain.ProjectMember{
		UserID:    userID,
		ProjectID: projectID,
	}

	comments := []domain.Comment{
		{
			ID:        uuid.New(),
			BoardID:   boardID,
			UserID:    commentUserID,
			Content:   "First comment",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			BoardID:   boardID,
			UserID:    commentUserID,
			Content:   "Second comment",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Mock setup
	suite.boardRepo.On("FindByID", boardID).Return(board, nil)
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).Return(member, nil)
	suite.commentRepo.On("FindByBoardID", boardID).Return(comments, nil)
	suite.userInfoCache.On("GetSimpleUsersBatch", ctx, []string{commentUserID.String(), commentUserID.String()}).Return(make(map[string]*cache.SimpleUser), nil)
	suite.userClient.On("GetSimpleUsers", []string{commentUserID.String(), commentUserID.String()}).Return([]client.SimpleUser{
		{
			ID:        commentUserID.String(),
			Name:      "Comment Author",
			AvatarURL: "http://example.com/avatar.jpg",
		},
	}, nil)
	suite.userInfoCache.On("SetSimpleUsersBatch", ctx, mock.AnythingOfType("[]cache.SimpleUser")).Return(nil)

	// When: Get comments
	result, err := suite.service.GetCommentsByBoardID(ctx, boardID, userID)

	// Then: Verify success
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "First comment", result[0].Content)
	assert.Equal(t, "Second comment", result[1].Content)

	suite.boardRepo.AssertExpectations(t)
	suite.projectRepo.AssertExpectations(t)
	suite.commentRepo.AssertExpectations(t)
}

func TestCommentService_GetCommentsByBoardID_EmptyList(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: Board with no comments
	ctx := context.Background()
	userID := uuid.New()
	boardID := uuid.New()
	projectID := uuid.New()

	board := &domain.Board{
		ID:        boardID,
		ProjectID: projectID,
		Title:     "Test Board",
	}

	member := &domain.ProjectMember{
		UserID:    userID,
		ProjectID: projectID,
	}

	suite.boardRepo.On("FindByID", boardID).Return(board, nil)
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).Return(member, nil)
	suite.commentRepo.On("FindByBoardID", boardID).Return([]domain.Comment{}, nil)
	suite.userInfoCache.On("GetSimpleUsersBatch", ctx, []string{}).Return(make(map[string]*cache.SimpleUser), nil)

	// When: Get comments
	result, err := suite.service.GetCommentsByBoardID(ctx, boardID, userID)

	// Then: Verify empty list
	assert.NoError(t, err)
	assert.Len(t, result, 0)

	suite.boardRepo.AssertExpectations(t)
	suite.projectRepo.AssertExpectations(t)
	suite.commentRepo.AssertExpectations(t)
}

func TestCommentService_GetCommentsByBoardID_BoardNotFound(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: Non-existent board
	ctx := context.Background()
	userID := uuid.New()
	boardID := uuid.New()

	suite.boardRepo.On("FindByID", boardID).Return((*domain.Board)(nil), gorm.ErrRecordNotFound)

	// When: Get comments
	result, err := suite.service.GetCommentsByBoardID(ctx, boardID, userID)

	// Then: Verify error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")

	suite.boardRepo.AssertExpectations(t)
}

func TestCommentService_GetCommentsByBoardID_UserNotProjectMember(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: User is not a project member
	ctx := context.Background()
	userID := uuid.New()
	boardID := uuid.New()
	projectID := uuid.New()

	board := &domain.Board{
		ID:        boardID,
		ProjectID: projectID,
		Title:     "Test Board",
	}

	suite.boardRepo.On("FindByID", boardID).Return(board, nil)
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).Return((*domain.ProjectMember)(nil), gorm.ErrRecordNotFound)

	// When: Get comments
	result, err := suite.service.GetCommentsByBoardID(ctx, boardID, userID)

	// Then: Verify forbidden error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not a member")

	suite.boardRepo.AssertExpectations(t)
	suite.projectRepo.AssertExpectations(t)
}

// ==================== UpdateComment Tests ====================

func TestCommentService_UpdateComment_Success(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: Valid comment update
	ctx := context.Background()
	userID := uuid.New()
	commentID := uuid.New()
	boardID := uuid.New()
	newContent := "Updated comment content"

	req := dto.UpdateCommentRequest{
		Content: newContent,
	}

	comment := &domain.Comment{
		ID:        commentID,
		BoardID:   boardID,
		UserID:    userID,
		Content:   "Old content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	simpleUser := &cache.SimpleUser{
		ID:        userID.String(),
		Name:      "Test User",
		AvatarURL: "http://example.com/avatar.jpg",
	}

	// Mock setup
	suite.commentRepo.On("FindByID", commentID).Return(comment, nil)
	suite.commentRepo.On("Update", mock.AnythingOfType("*domain.Comment")).Return(nil)
	suite.userInfoCache.On("GetSimpleUser", ctx, userID.String()).Return(true, simpleUser, nil)

	// When: Update comment
	result, err := suite.service.UpdateComment(ctx, commentID, req, userID)

	// Then: Verify success
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newContent, result.Content)
	assert.Equal(t, userID, result.UserID)

	suite.commentRepo.AssertExpectations(t)
}

func TestCommentService_UpdateComment_CommentNotFound(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: Non-existent comment
	ctx := context.Background()
	userID := uuid.New()
	commentID := uuid.New()

	req := dto.UpdateCommentRequest{
		Content: "Updated content",
	}

	suite.commentRepo.On("FindByID", commentID).Return((*domain.Comment)(nil), gorm.ErrRecordNotFound)

	// When: Update comment
	result, err := suite.service.UpdateComment(ctx, commentID, req, userID)

	// Then: Verify error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")

	suite.commentRepo.AssertExpectations(t)
}

func TestCommentService_UpdateComment_NotCommentAuthor(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: User is not the comment author
	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()
	commentID := uuid.New()
	boardID := uuid.New()

	req := dto.UpdateCommentRequest{
		Content: "Updated content",
	}

	comment := &domain.Comment{
		ID:        commentID,
		BoardID:   boardID,
		UserID:    otherUserID, // Different user
		Content:   "Old content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	suite.commentRepo.On("FindByID", commentID).Return(comment, nil)

	// When: Update comment
	result, err := suite.service.UpdateComment(ctx, commentID, req, userID)

	// Then: Verify forbidden error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "permission")

	suite.commentRepo.AssertExpectations(t)
}

func TestCommentService_UpdateComment_EmptyContent(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: Empty content (domain validation should fail)
	ctx := context.Background()
	userID := uuid.New()
	commentID := uuid.New()
	boardID := uuid.New()

	req := dto.UpdateCommentRequest{
		Content: "", // Empty content
	}

	comment := &domain.Comment{
		ID:        commentID,
		BoardID:   boardID,
		UserID:    userID,
		Content:   "Old content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	suite.commentRepo.On("FindByID", commentID).Return(comment, nil)

	// When: Update comment with empty content
	result, err := suite.service.UpdateComment(ctx, commentID, req, userID)

	// Then: Verify domain validation error
	assert.Error(t, err)
	assert.Nil(t, result)

	suite.commentRepo.AssertExpectations(t)
}

// ==================== DeleteComment Tests ====================

func TestCommentService_DeleteComment_Success(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: Valid comment deletion
	ctx := context.Background()
	userID := uuid.New()
	commentID := uuid.New()
	boardID := uuid.New()

	comment := &domain.Comment{
		ID:        commentID,
		BoardID:   boardID,
		UserID:    userID,
		Content:   "Test content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock setup
	suite.commentRepo.On("FindByID", commentID).Return(comment, nil)
	suite.commentRepo.On("Delete", commentID).Return(nil)

	// When: Delete comment
	err := suite.service.DeleteComment(ctx, commentID, userID)

	// Then: Verify success
	assert.NoError(t, err)

	suite.commentRepo.AssertExpectations(t)
}

func TestCommentService_DeleteComment_CommentNotFound(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: Non-existent comment
	ctx := context.Background()
	userID := uuid.New()
	commentID := uuid.New()

	suite.commentRepo.On("FindByID", commentID).Return((*domain.Comment)(nil), gorm.ErrRecordNotFound)

	// When: Delete comment
	err := suite.service.DeleteComment(ctx, commentID, userID)

	// Then: Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	suite.commentRepo.AssertExpectations(t)
}

func TestCommentService_DeleteComment_NotCommentAuthor(t *testing.T) {
	suite := setupCommentServiceTest(t)

	// Given: User is not the comment author
	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()
	commentID := uuid.New()
	boardID := uuid.New()

	comment := &domain.Comment{
		ID:        commentID,
		BoardID:   boardID,
		UserID:    otherUserID, // Different user
		Content:   "Test content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	suite.commentRepo.On("FindByID", commentID).Return(comment, nil)

	// When: Delete comment
	err := suite.service.DeleteComment(ctx, commentID, userID)

	// Then: Verify forbidden error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission")

	suite.commentRepo.AssertExpectations(t)
}
