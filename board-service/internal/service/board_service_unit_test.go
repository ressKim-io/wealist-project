package service_test

import (
	"board-service/internal/apperrors"
	"board-service/internal/cache"
	"board-service/internal/client"
	"board-service/internal/domain"
	"board-service/internal/dto"
	"board-service/internal/service"
	"board-service/internal/testutil"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ==================== Mocks ====================

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

func (m *MockUserClient) CheckWorkspaceExists(ctx context.Context, workspaceID, token string) (bool, error) {
	args := m.Called(ctx, workspaceID, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserClient) ValidateWorkspaceMembership(ctx context.Context, workspaceID, userID, token string) (bool, error) {
	args := m.Called(ctx, workspaceID, userID, token)
	return args.Bool(0), args.Error(1)
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

func (m *MockUserClient) SearchUsers(ctx context.Context, query string) ([]client.UserInfo, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]client.UserInfo), args.Error(1)
}

func (m *MockUserClient) GetWorkspace(ctx context.Context, workspaceID string, token string) (*client.WorkspaceInfo, error) {
	args := m.Called(ctx, workspaceID, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.WorkspaceInfo), args.Error(1)
}

type MockUserInfoCache struct{
	mock.Mock
}

func (m *MockUserInfoCache) GetUserInfo(ctx context.Context, userID string) (bool, *cache.UserInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(1) == nil {
		return args.Bool(0), nil, args.Error(2)
	}
	return args.Bool(0), args.Get(1).(*cache.UserInfo), args.Error(2)
}

func (m *MockUserInfoCache) SetUserInfo(ctx context.Context, user *cache.UserInfo) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserInfoCache) GetSimpleUsersBatch(ctx context.Context, userIDs []string) (map[string]*cache.SimpleUser, error) {
	args := m.Called(ctx, userIDs)
	return args.Get(0).(map[string]*cache.SimpleUser), args.Error(1)
}

func (m *MockUserInfoCache) SetSimpleUsersBatch(ctx context.Context, users []cache.SimpleUser) error {
	args := m.Called(ctx, users)
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

func (m *MockUserInfoCache) DeleteUserInfo(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserInfoCache) InvalidateUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockDB struct {
	mock.Mock
}

func (m *MockDB) Transaction(fc func(tx *gorm.DB) error) error {
	args := m.Called(fc)
	return args.Error(0)
}

// ==================== Test Suite ====================

type BoardServiceTestSuite struct {
	boardRepo     *testutil.MockBoardRepository
	projectRepo   *testutil.MockProjectRepository
	roleRepo      *testutil.MockRoleRepository
	fieldRepo     *testutil.MockFieldRepository
	commentRepo   *testutil.MockCommentRepository
	userClient    *MockUserClient
	userInfoCache *MockUserInfoCache
	logger        *zap.Logger
	service       service.BoardService
}

func setupBoardServiceTest(t *testing.T) *BoardServiceTestSuite {
	suite := &BoardServiceTestSuite{
		boardRepo:     new(testutil.MockBoardRepository),
		projectRepo:   new(testutil.MockProjectRepository),
		roleRepo:      new(testutil.MockRoleRepository),
		fieldRepo:     new(testutil.MockFieldRepository),
		commentRepo:   new(testutil.MockCommentRepository),
		userClient:    new(MockUserClient),
		userInfoCache: new(MockUserInfoCache),
		logger:        zap.NewNop(),
	}

	// Create service with mocks
	suite.service = service.NewBoardService(
		suite.boardRepo,
		suite.projectRepo,
		suite.roleRepo,
		suite.fieldRepo,
		suite.commentRepo,
		suite.userClient,
		suite.userInfoCache,
		suite.logger,
		nil, // db - will be mocked when needed
	)

	return suite
}

// ==================== CreateBoard Tests ====================

func TestCreateBoard_Success(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.boardRepo.AssertExpectations(t)
	defer suite.projectRepo.AssertExpectations(t)

	// Arrange
	userID := uuid.New()
	projectID := uuid.New()
	req := &dto.CreateBoardRequest{
		ProjectID: projectID.String(),
		Title:     "New Board",
		Content:   "Board content",
	}

	member := testutil.NewTestProjectMember(projectID, userID, uuid.New())

	// Mock: Check user is project member
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).
		Return(member, nil)

	// Mock: Create board
	suite.boardRepo.On("Create", mock.MatchedBy(func(b *domain.Board) bool {
		return b.Title == "New Board" && b.ProjectID == projectID
	})).Return(nil)

	// Mock: Cache operations
	suite.userInfoCache.On("GetSimpleUsersBatch", mock.Anything, mock.Anything).
		Return(make(map[string]*cache.SimpleUser), nil)
	suite.userClient.On("GetUsersBatch", mock.Anything, mock.Anything).
		Return([]client.UserInfo{}, nil)
	suite.userInfoCache.On("SetSimpleUsersBatch", mock.Anything, mock.Anything).
		Return(nil)

	// Act
	result, err := suite.service.CreateBoard(userID.String(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "New Board", result.Title)
	assert.Equal(t, projectID.String(), result.ProjectID)
}

func TestCreateBoard_InvalidUserID(t *testing.T) {
	suite := setupBoardServiceTest(t)

	req := &dto.CreateBoardRequest{
		ProjectID: uuid.New().String(),
		Title:     "New Board",
	}

	result, err := suite.service.CreateBoard("invalid-uuid", req)

	assert.Error(t, err)
	assert.Nil(t, result)

	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrCodeBadRequest, appErr.Code)
}

func TestCreateBoard_InvalidProjectID(t *testing.T) {
	suite := setupBoardServiceTest(t)

	req := &dto.CreateBoardRequest{
		ProjectID: "invalid-uuid",
		Title:     "New Board",
	}

	result, err := suite.service.CreateBoard(uuid.New().String(), req)

	assert.Error(t, err)
	assert.Nil(t, result)

	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrCodeBadRequest, appErr.Code)
}

func TestCreateBoard_UserNotMember(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.projectRepo.AssertExpectations(t)

	userID := uuid.New()
	projectID := uuid.New()
	req := &dto.CreateBoardRequest{
		ProjectID: projectID.String(),
		Title:     "New Board",
	}

	// Mock: User is not project member
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).
		Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.CreateBoard(userID.String(), req)

	assert.Error(t, err)
	assert.Nil(t, result)

	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrCodeForbidden, appErr.Code)
}

func TestCreateBoard_WithAssignee_Success(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.boardRepo.AssertExpectations(t)
	defer suite.projectRepo.AssertExpectations(t)

	userID := uuid.New()
	projectID := uuid.New()
	assigneeID := uuid.New()
	assigneeIDStr := assigneeID.String()

	req := &dto.CreateBoardRequest{
		ProjectID:  projectID.String(),
		Title:      "New Board",
		AssigneeID: &assigneeIDStr,
	}

	member := testutil.NewTestProjectMember(projectID, userID, uuid.New())
	assigneeMember := testutil.NewTestProjectMember(projectID, assigneeID, uuid.New())

	// Mock: Check creator is member
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).
		Return(member, nil)

	// Mock: Check assignee is member
	suite.projectRepo.On("FindMemberByUserAndProject", assigneeID, projectID).
		Return(assigneeMember, nil)

	// Mock: Create board
	suite.boardRepo.On("Create", mock.MatchedBy(func(b *domain.Board) bool {
		return b.AssigneeID != nil && *b.AssigneeID == assigneeID
	})).Return(nil)

	// Mock: Cache
	suite.userInfoCache.On("GetSimpleUsersBatch", mock.Anything, mock.Anything).
		Return(make(map[string]*cache.SimpleUser), nil)
	suite.userClient.On("GetUsersBatch", mock.Anything, mock.Anything).
		Return([]client.UserInfo{}, nil)
	suite.userInfoCache.On("SetSimpleUsersBatch", mock.Anything, mock.Anything).
		Return(nil)

	result, err := suite.service.CreateBoard(userID.String(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCreateBoard_AssigneeNotMember(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.projectRepo.AssertExpectations(t)

	userID := uuid.New()
	projectID := uuid.New()
	assigneeID := uuid.New()
	assigneeIDStr := assigneeID.String()

	req := &dto.CreateBoardRequest{
		ProjectID:  projectID.String(),
		Title:      "New Board",
		AssigneeID: &assigneeIDStr,
	}

	member := testutil.NewTestProjectMember(projectID, userID, uuid.New())

	// Mock: Creator is member
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).
		Return(member, nil)

	// Mock: Assignee is NOT member
	suite.projectRepo.On("FindMemberByUserAndProject", assigneeID, projectID).
		Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.CreateBoard(userID.String(), req)

	assert.Error(t, err)
	assert.Nil(t, result)

	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrCodeNotFound, appErr.Code)
}

// ==================== GetBoard Tests ====================

func TestGetBoard_Success(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.boardRepo.AssertExpectations(t)
	defer suite.projectRepo.AssertExpectations(t)

	userID := uuid.New()
	projectID := uuid.New()
	board := testutil.NewTestBoard(projectID, userID)
	member := testutil.NewTestProjectMember(projectID, userID, uuid.New())

	// Mock: Find board
	suite.boardRepo.On("FindByID", board.ID).
		Return(board, nil)

	// Mock: Check user is member
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).
		Return(member, nil)

	// Mock: Cache
	suite.userInfoCache.On("GetSimpleUsersBatch", mock.Anything, mock.Anything).
		Return(make(map[string]*cache.SimpleUser), nil)
	suite.userClient.On("GetUsersBatch", mock.Anything, mock.Anything).
		Return([]client.UserInfo{{UserID: userID.String(), Name: "Test User"}}, nil)
	suite.userInfoCache.On("SetSimpleUsersBatch", mock.Anything, mock.Anything).
		Return(nil)

	result, err := suite.service.GetBoard(board.ID.String(), userID.String())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, board.ID.String(), result.ID)
}

func TestGetBoard_NotFound(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.boardRepo.AssertExpectations(t)

	boardID := uuid.New()
	userID := uuid.New()

	suite.boardRepo.On("FindByID", boardID).
		Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.GetBoard(boardID.String(), userID.String())

	assert.Error(t, err)
	assert.Nil(t, result)

	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrCodeNotFound, appErr.Code)
}

func TestGetBoard_UserNotMember(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.boardRepo.AssertExpectations(t)
	defer suite.projectRepo.AssertExpectations(t)

	userID := uuid.New()
	otherUserID := uuid.New()
	projectID := uuid.New()
	board := testutil.NewTestBoard(projectID, userID)

	suite.boardRepo.On("FindByID", board.ID).
		Return(board, nil)

	suite.projectRepo.On("FindMemberByUserAndProject", otherUserID, projectID).
		Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.GetBoard(board.ID.String(), otherUserID.String())

	assert.Error(t, err)
	assert.Nil(t, result)

	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrCodeForbidden, appErr.Code)
}

// ==================== GetBoards Tests ====================

func TestGetBoards_Success(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.boardRepo.AssertExpectations(t)
	defer suite.projectRepo.AssertExpectations(t)

	userID := uuid.New()
	projectID := uuid.New()
	member := testutil.NewTestProjectMember(projectID, userID, uuid.New())

	boards := []domain.Board{
		*testutil.NewTestBoard(projectID, userID),
		*testutil.NewTestBoard(projectID, userID),
	}

	req := &dto.GetBoardsRequest{
		ProjectID: projectID.String(),
		Page:      1,
		Limit:     20,
	}

	// Mock: Check member
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).
		Return(member, nil)

	// Mock: Find boards
	suite.boardRepo.On("FindByProject", projectID, mock.Anything, 1, 20).
		Return(boards, int64(2), nil)

	// Mock: Cache
	suite.userInfoCache.On("GetSimpleUsersBatch", mock.Anything, mock.Anything).
		Return(make(map[string]*cache.SimpleUser), nil)
	suite.userClient.On("GetUsersBatch", mock.Anything, mock.Anything).
		Return([]client.UserInfo{{UserID: userID.String(), Name: "Test User"}}, nil)
	suite.userInfoCache.On("SetSimpleUsersBatch", mock.Anything, mock.Anything).
		Return(nil)

	result, err := suite.service.GetBoards(userID.String(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result.Boards))
	assert.Equal(t, int64(2), result.Total)
}

func TestGetBoards_DefaultPagination(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.boardRepo.AssertExpectations(t)
	defer suite.projectRepo.AssertExpectations(t)

	userID := uuid.New()
	projectID := uuid.New()
	member := testutil.NewTestProjectMember(projectID, userID, uuid.New())

	req := &dto.GetBoardsRequest{
		ProjectID: projectID.String(),
		Page:      0, // Invalid, should default to 1
		Limit:     0, // Invalid, should default to 20
	}

	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).
		Return(member, nil)

	// Should use default page=1, limit=20
	suite.boardRepo.On("FindByProject", projectID, mock.Anything, 1, 20).
		Return([]domain.Board{}, int64(0), nil)

	result, err := suite.service.GetBoards(userID.String(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Page)   // Default
	assert.Equal(t, 20, result.Limit) // Default
}

// ==================== UpdateBoard Tests ====================

func TestUpdateBoard_Success_AsAuthor(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.boardRepo.AssertExpectations(t)
	defer suite.projectRepo.AssertExpectations(t)
	defer suite.roleRepo.AssertExpectations(t)

	userID := uuid.New()
	projectID := uuid.New()
	board := testutil.NewTestBoard(projectID, userID)
	member := testutil.NewTestProjectMember(projectID, userID, uuid.New())
	memberRole := testutil.NewMemberRole()
	member.Role = memberRole

	req := &dto.UpdateBoardRequest{
		Title:   "Updated Title",
		Content: "Updated Content",
	}

	suite.boardRepo.On("FindByID", board.ID).Return(board, nil)
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).
		Return(member, nil)
	suite.roleRepo.On("FindByID", member.RoleID).Return(memberRole, nil)

	// Author can update their own board (even as MEMBER role)
	suite.boardRepo.On("Update", mock.MatchedBy(func(b *domain.Board) bool {
		return b.Title == "Updated Title" && b.Description == "Updated Content"
	})).Return(nil)

	// Mock for GetBoard call at the end
	suite.boardRepo.On("FindByID", board.ID).Return(board, nil)
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).
		Return(member, nil)
	suite.userInfoCache.On("GetSimpleUsersBatch", mock.Anything, mock.Anything).
		Return(make(map[string]*cache.SimpleUser), nil)
	suite.userClient.On("GetUsersBatch", mock.Anything, mock.Anything).
		Return([]client.UserInfo{{UserID: userID.String(), Name: "Test User"}}, nil)
	suite.userInfoCache.On("SetSimpleUsersBatch", mock.Anything, mock.Anything).
		Return(nil)

	result, err := suite.service.UpdateBoard(board.ID.String(), userID.String(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestUpdateBoard_Forbidden_NotAuthorAndNotAdmin(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.boardRepo.AssertExpectations(t)
	defer suite.projectRepo.AssertExpectations(t)
	defer suite.roleRepo.AssertExpectations(t)

	authorID := uuid.New()
	otherUserID := uuid.New()
	projectID := uuid.New()
	board := testutil.NewTestBoard(projectID, authorID)

	member := testutil.NewTestProjectMember(projectID, otherUserID, uuid.New())
	memberRole := testutil.NewMemberRole()
	member.Role = memberRole

	req := &dto.UpdateBoardRequest{
		Title: "Hacked",
	}

	suite.boardRepo.On("FindByID", board.ID).Return(board, nil)
	suite.projectRepo.On("FindMemberByUserAndProject", otherUserID, projectID).
		Return(member, nil)
	suite.roleRepo.On("FindByID", member.RoleID).Return(memberRole, nil)

	result, err := suite.service.UpdateBoard(board.ID.String(), otherUserID.String(), req)

	assert.Error(t, err)
	assert.Nil(t, result)

	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrCodeForbidden, appErr.Code)
}

func TestUpdateBoard_Success_AsAdmin(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.boardRepo.AssertExpectations(t)
	defer suite.projectRepo.AssertExpectations(t)
	defer suite.roleRepo.AssertExpectations(t)

	authorID := uuid.New()
	adminID := uuid.New()
	projectID := uuid.New()
	board := testutil.NewTestBoard(projectID, authorID)

	adminMember := testutil.NewTestProjectMember(projectID, adminID, uuid.New())
	adminRole := testutil.NewAdminRole()
	adminMember.Role = adminRole

	req := &dto.UpdateBoardRequest{
		Title: "Admin Updated",
	}

	suite.boardRepo.On("FindByID", board.ID).Return(board, nil)
	suite.projectRepo.On("FindMemberByUserAndProject", adminID, projectID).
		Return(adminMember, nil)
	suite.roleRepo.On("FindByID", adminMember.RoleID).Return(adminRole, nil)
	suite.boardRepo.On("Update", mock.Anything).Return(nil)

	// Mock GetBoard
	suite.boardRepo.On("FindByID", board.ID).Return(board, nil)
	suite.projectRepo.On("FindMemberByUserAndProject", adminID, projectID).
		Return(adminMember, nil)
	suite.userInfoCache.On("GetSimpleUsersBatch", mock.Anything, mock.Anything).
		Return(make(map[string]*cache.SimpleUser), nil)
	suite.userClient.On("GetUsersBatch", mock.Anything, mock.Anything).
		Return([]client.UserInfo{}, nil)
	suite.userInfoCache.On("SetSimpleUsersBatch", mock.Anything, mock.Anything).
		Return(nil)

	result, err := suite.service.UpdateBoard(board.ID.String(), adminID.String(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// ==================== DeleteBoard Tests ====================

func TestDeleteBoard_Success_AsAuthor(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.boardRepo.AssertExpectations(t)
	defer suite.projectRepo.AssertExpectations(t)
	defer suite.roleRepo.AssertExpectations(t)

	userID := uuid.New()
	projectID := uuid.New()
	board := testutil.NewTestBoard(projectID, userID)
	member := testutil.NewTestProjectMember(projectID, userID, uuid.New())
	memberRole := testutil.NewMemberRole()

	suite.boardRepo.On("FindByID", board.ID).Return(board, nil)
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).
		Return(member, nil)
	suite.roleRepo.On("FindByID", member.RoleID).Return(memberRole, nil)
	suite.boardRepo.On("Delete", board.ID).Return(nil)

	err := suite.service.DeleteBoard(board.ID.String(), userID.String())

	assert.NoError(t, err)
}

func TestDeleteBoard_Forbidden_NotAuthor(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.boardRepo.AssertExpectations(t)
	defer suite.projectRepo.AssertExpectations(t)
	defer suite.roleRepo.AssertExpectations(t)

	authorID := uuid.New()
	otherUserID := uuid.New()
	projectID := uuid.New()
	board := testutil.NewTestBoard(projectID, authorID)

	member := testutil.NewTestProjectMember(projectID, otherUserID, uuid.New())
	memberRole := testutil.NewMemberRole()

	suite.boardRepo.On("FindByID", board.ID).Return(board, nil)
	suite.projectRepo.On("FindMemberByUserAndProject", otherUserID, projectID).
		Return(member, nil)
	suite.roleRepo.On("FindByID", member.RoleID).Return(memberRole, nil)

	err := suite.service.DeleteBoard(board.ID.String(), otherUserID.String())

	assert.Error(t, err)

	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrCodeForbidden, appErr.Code)
}

// ==================== Edge Cases and Error Handling ====================

func TestBoardService_RepositoryError(t *testing.T) {
	suite := setupBoardServiceTest(t)
	defer suite.boardRepo.AssertExpectations(t)

	boardID := uuid.New()
	userID := uuid.New()

	// Simulate database error
	suite.boardRepo.On("FindByID", boardID).
		Return(nil, errors.New("database connection failed"))

	result, err := suite.service.GetBoard(boardID.String(), userID.String())

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestBoardService_InvalidUUIDs(t *testing.T) {
	suite := setupBoardServiceTest(t)

	tests := []struct {
		name    string
		boardID string
		userID  string
	}{
		{"Invalid boardID", "not-a-uuid", uuid.New().String()},
		{"Invalid userID", uuid.New().String(), "not-a-uuid"},
		{"Both invalid", "invalid", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := suite.service.GetBoard(tt.boardID, tt.userID)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}
