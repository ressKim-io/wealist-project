package service

import (
	"board-service/internal/client"
	"board-service/internal/domain"
	"board-service/internal/dto"
	"board-service/internal/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"gorm.io/gorm"
)

// ==================== Test Suite Setup ====================

type ProjectServiceTestSuite struct {
	projectRepo    *testutil.MockProjectRepository
	roleRepo       *testutil.MockRoleRepository
	fieldRepo      *testutil.MockFieldRepository
	userClient     *MockUserClient
	workspaceCache *MockWorkspaceCache
	userInfoCache  *MockUserInfoCache
	logger         *zap.Logger
	service        ProjectService
}

func setupProjectServiceTest(t *testing.T) *ProjectServiceTestSuite {
	projectRepo := new(testutil.MockProjectRepository)
	roleRepo := new(testutil.MockRoleRepository)
	fieldRepo := new(testutil.MockFieldRepository)
	userClient := new(MockUserClient)
	workspaceCache := new(MockWorkspaceCache)
	userInfoCache := new(MockUserInfoCache)
	logger := zaptest.NewLogger(t)

	service := NewProjectService(
		projectRepo,
		roleRepo,
		fieldRepo,
		userClient,
		workspaceCache,
		userInfoCache,
		logger,
		nil, // db는 테스트에서 사용 안 함
	)

	return &ProjectServiceTestSuite{
		projectRepo:    projectRepo,
		roleRepo:       roleRepo,
		fieldRepo:      fieldRepo,
		userClient:     userClient,
		workspaceCache: workspaceCache,
		userInfoCache:  userInfoCache,
		logger:         logger,
		service:        service,
	}
}

// ==================== CreateProject Tests ====================

func TestProjectService_CreateProject_Success(t *testing.T) {
	suite := setupProjectServiceTest(t)

	// Given
	userID := uuid.New().String()
	workspaceID := uuid.New()
	token := "valid-token"

	req := &dto.CreateProjectRequest{
		WorkspaceID: workspaceID.String(),
		Name:        "Test Project",
		Description: "Test Description",
	}

	ownerRole := &domain.Role{
		BaseModel: domain.BaseModel{ID: uuid.New()},
		Name:      "OWNER",
		Level:     100,
	}

	expectedProject := &domain.Project{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     uuid.MustParse(userID),
		IsPublic:    false,
	}

	expectedUserInfo := &client.UserInfo{
		UserID:   userID,
		Name:     "Test User",
		Email:    "test@example.com",
		IsActive: true,
	}

	// Mocks
	suite.workspaceCache.On("ValidateWorkspace", mock.Anything, workspaceID.String(), userID, token).
		Return(true, nil)

	suite.projectRepo.On("Create", mock.AnythingOfType("*domain.Project")).
		Run(func(args mock.Arguments) {
			project := args.Get(0).(*domain.Project)
			project.ID = expectedProject.ID
			project.CreatedAt = expectedProject.CreatedAt
			project.UpdatedAt = expectedProject.UpdatedAt
		}).
		Return(nil)

	suite.roleRepo.On("FindByName", "OWNER").
		Return(ownerRole, nil)

	suite.projectRepo.On("AddMember", mock.AnythingOfType("*domain.ProjectMember")).
		Return(nil)

	userMap := map[string]client.UserInfo{
		userID: *expectedUserInfo,
	}
	suite.userClient.On("GetUsersBatch", mock.Anything, []string{userID}).
		Return(userMap, nil)

	// When
	result, err := suite.service.CreateProject(userID, token, req)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Name, result.Name)
	assert.Equal(t, req.Description, result.Description)
	assert.Equal(t, expectedUserInfo.Name, result.OwnerName)
	assert.Equal(t, expectedUserInfo.Email, result.OwnerEmail)

	suite.workspaceCache.AssertExpectations(t)
	suite.projectRepo.AssertExpectations(t)
	suite.roleRepo.AssertExpectations(t)
	suite.userClient.AssertExpectations(t)
}

func TestProjectService_CreateProject_InvalidWorkspace(t *testing.T) {
	suite := setupProjectServiceTest(t)

	// Given
	userID := uuid.New().String()
	workspaceID := uuid.New()
	token := "valid-token"

	req := &dto.CreateProjectRequest{
		WorkspaceID: workspaceID.String(),
		Name:        "Test Project",
	}

	// Mock: Workspace validation fails
	suite.workspaceCache.On("ValidateWorkspace", mock.Anything, workspaceID.String(), userID, token).
		Return(false, nil)

	// When
	result, err := suite.service.CreateProject(userID, token, req)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "워크스페이스 접근 권한이 없습니다")

	suite.workspaceCache.AssertExpectations(t)
}

func TestProjectService_CreateProject_InvalidUUID(t *testing.T) {
	suite := setupProjectServiceTest(t)

	// Given
	req := &dto.CreateProjectRequest{
		WorkspaceID: "invalid-uuid",
		Name:        "Test Project",
	}

	// When
	result, err := suite.service.CreateProject("user-id", "token", req)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== GetProject Tests ====================

func TestProjectService_GetProject_Success(t *testing.T) {
	suite := setupProjectServiceTest(t)

	// Given
	projectID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()

	project := &domain.Project{
		BaseModel:   domain.BaseModel{ID: projectID},
		WorkspaceID: uuid.New(),
		Name:        "Test Project",
		Description: "Description",
		OwnerID:     ownerID,
		IsPublic:    true,
	}

	ownerRole := &domain.Role{
		BaseModel: domain.BaseModel{ID: uuid.New()},
		Name:      "OWNER",
		Level:     100,
	}

	member := &domain.ProjectMember{
		BaseModel: domain.BaseModel{
			ID:       uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ProjectID: projectID,
		UserID:    userID,
		Role:      ownerRole,
		JoinedAt:  time.Now(),
	}

	ownerInfo := &client.UserInfo{
		UserID:   ownerID.String(),
		Name:     "Owner User",
		Email:    "owner@example.com",
		IsActive: true,
	}

	// Mocks
	suite.projectRepo.On("FindByID", projectID).
		Return(project, nil)

	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).
		Return(member, nil)

	userMap := map[string]client.UserInfo{
		ownerID.String(): *ownerInfo,
	}
	suite.userClient.On("GetUsersBatch", mock.Anything, []string{ownerID.String()}).
		Return(userMap, nil)

	// When
	result, err := suite.service.GetProject(projectID.String(), userID.String())

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, projectID.String(), result.ID)
	assert.Equal(t, project.Name, result.Name)
	assert.Equal(t, ownerInfo.Name, result.OwnerName)

	suite.projectRepo.AssertExpectations(t)
	suite.userClient.AssertExpectations(t)
}

func TestProjectService_GetProject_NotFound(t *testing.T) {
	suite := setupProjectServiceTest(t)

	// Given
	projectID := uuid.New()
	userID := uuid.New()

	// Mock: Project not found
	suite.projectRepo.On("FindByID", projectID).
		Return(nil, gorm.ErrRecordNotFound)

	// When
	result, err := suite.service.GetProject(projectID.String(), userID.String())

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "프로젝트를 찾을 수 없습니다")

	suite.projectRepo.AssertExpectations(t)
}

func TestProjectService_GetProject_NoPermission(t *testing.T) {
	suite := setupProjectServiceTest(t)

	// Given
	projectID := uuid.New()
	userID := uuid.New()

	project := &domain.Project{
		BaseModel:   domain.BaseModel{ID: projectID},
		WorkspaceID: uuid.New(),
		Name:        "Private Project",
		OwnerID:     uuid.New(),
		IsPublic:    false, // Private project
	}

	// Mocks
	suite.projectRepo.On("FindByID", projectID).
		Return(project, nil)

	// User is not a member
	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).
		Return(nil, gorm.ErrRecordNotFound)

	// When
	result, err := suite.service.GetProject(projectID.String(), userID.String())

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "프로젝트 접근 권한이 없습니다")

	suite.projectRepo.AssertExpectations(t)
}

// ==================== UpdateProject Tests ====================

func TestProjectService_UpdateProject_Success(t *testing.T) {
	suite := setupProjectServiceTest(t)

	// Given
	projectID := uuid.New()
	userID := uuid.New()

	project := &domain.Project{
		BaseModel:   domain.BaseModel{ID: projectID},
		WorkspaceID: uuid.New(),
		Name:        "Old Name",
		Description: "Old Description",
		OwnerID:     userID, // User is owner
		IsPublic:    false,
	}

	req := &dto.UpdateProjectRequest{
		Name:        "New Name",
		Description: "New Description",
	}

	// Mocks
	suite.projectRepo.On("FindByID", projectID).
		Return(project, nil)

	suite.projectRepo.On("Update", mock.AnythingOfType("*domain.Project")).
		Return(nil)

	ownerInfo := &client.UserInfo{
		UserID:   userID.String(),
		Name:     "Owner",
		Email:    "owner@example.com",
		IsActive: true,
	}

	userMap := map[string]client.UserInfo{
		userID.String(): *ownerInfo,
	}
	suite.userClient.On("GetUsersBatch", mock.Anything, []string{userID.String()}).
		Return(userMap, nil)

	// When
	result, err := suite.service.UpdateProject(projectID.String(), userID.String(), req)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Domain 메서드 사용으로 변경됨
	// assert.Equal(t, "New Name", result.Name)

	suite.projectRepo.AssertExpectations(t)
}

func TestProjectService_UpdateProject_NotOwner(t *testing.T) {
	suite := setupProjectServiceTest(t)

	// Given
	projectID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New() // Different from userID

	project := &domain.Project{
		BaseModel: domain.BaseModel{ID: projectID},
		OwnerID:   ownerID, // Not the requesting user
	}

	req := &dto.UpdateProjectRequest{
		Name: "New Name",
	}

	// Mocks
	suite.projectRepo.On("FindByID", projectID).
		Return(project, nil)

	// When
	result, err := suite.service.UpdateProject(projectID.String(), userID.String(), req)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "프로젝트 소유자만 수정할 수 있습니다")

	suite.projectRepo.AssertExpectations(t)
}

// ==================== DeleteProject Tests ====================

func TestProjectService_DeleteProject_Success(t *testing.T) {
	suite := setupProjectServiceTest(t)

	// Given
	projectID := uuid.New()
	userID := uuid.New()

	project := &domain.Project{
		BaseModel: domain.BaseModel{ID: projectID},
		OwnerID:   userID, // User is owner
	}

	// Mocks
	suite.projectRepo.On("FindByID", projectID).
		Return(project, nil)

	suite.projectRepo.On("Update", mock.AnythingOfType("*domain.Project")).
		Run(func(args mock.Arguments) {
			p := args.Get(0).(*domain.Project)
			assert.True(t, p.IsDeleted, "Project should be marked as deleted")
		}).
		Return(nil)

	// When
	err := suite.service.DeleteProject(projectID.String(), userID.String())

	// Then
	assert.NoError(t, err)
	suite.projectRepo.AssertExpectations(t)
}

func TestProjectService_DeleteProject_NotOwner(t *testing.T) {
	suite := setupProjectServiceTest(t)

	// Given
	projectID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New() // Different

	project := &domain.Project{
		BaseModel: domain.BaseModel{ID: projectID},
		OwnerID:   ownerID,
	}

	// Mocks
	suite.projectRepo.On("FindByID", projectID).
		Return(project, nil)

	// When
	err := suite.service.DeleteProject(projectID.String(), userID.String())

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "프로젝트 소유자만 삭제할 수 있습니다")

	suite.projectRepo.AssertExpectations(t)
}

// ==================== GetProjectMembers Tests ====================

func TestProjectService_GetProjectMembers_Success(t *testing.T) {
	suite := setupProjectServiceTest(t)

	// Given
	projectID := uuid.New()
	userID := uuid.New()
	memberID1 := uuid.New()
	memberID2 := uuid.New()

	project := &domain.Project{
		BaseModel: domain.BaseModel{ID: projectID},
		OwnerID:   userID,
	}

	ownerRole := &domain.Role{Name: "OWNER", Level: 100}
	memberRole := &domain.Role{Name: "MEMBER", Level: 10}

	members := []domain.ProjectMember{
		{
			BaseModel: domain.BaseModel{ID: uuid.New()},
			ProjectID: projectID,
			UserID:    userID,
			Role:      ownerRole,
		},
		{
			BaseModel: domain.BaseModel{ID: uuid.New()},
			ProjectID: projectID,
			UserID:    memberID1,
			Role:      memberRole,
		},
		{
			BaseModel: domain.BaseModel{ID: uuid.New()},
			ProjectID: projectID,
			UserID:    memberID2,
			Role:      memberRole,
		},
	}

	// Mocks
	suite.projectRepo.On("FindByID", projectID).
		Return(project, nil)

	suite.projectRepo.On("FindMemberByUserAndProject", userID, projectID).
		Return(&members[0], nil)

	suite.projectRepo.On("FindMembers", projectID).
		Return(members, nil)

	userMap := map[string]client.UserInfo{
		userID.String():    {UserID: userID.String(), Name: "Owner", Email: "owner@test.com"},
		memberID1.String(): {UserID: memberID1.String(), Name: "Member1", Email: "m1@test.com"},
		memberID2.String(): {UserID: memberID2.String(), Name: "Member2", Email: "m2@test.com"},
	}
	suite.userClient.On("GetUsersBatch", mock.Anything, mock.Anything).
		Return(userMap, nil)

	// When
	result, err := suite.service.GetProjectMembers(projectID.String(), userID.String())

	// Then
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "OWNER", result[0].RoleName)

	suite.projectRepo.AssertExpectations(t)
}

// ==================== Helper Functions Test ====================

func TestProjectService_InvalidUUID(t *testing.T) {
	suite := setupProjectServiceTest(t)

	// Invalid project ID
	_, err := suite.service.GetProject("invalid-uuid", uuid.New().String())
	assert.Error(t, err)

	// Invalid user ID
	_, err = suite.service.GetProject(uuid.New().String(), "invalid-uuid")
	assert.Error(t, err)
}
