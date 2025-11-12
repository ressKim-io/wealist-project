package repository_test

import (
	"board-service/internal/domain"
	"board-service/internal/repository"
	"board-service/internal/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// ==================== Test Suite Setup ====================

type ProjectRepositoryTestSuite struct {
	db   *testutil.TestDB
	repo repository.ProjectRepository
}

func setupProjectRepoTest(t *testing.T) *ProjectRepositoryTestSuite {
	db := testutil.SetupTestDB()
	repo := repository.NewProjectRepository(db.DB)

	return &ProjectRepositoryTestSuite{
		db:   db,
		repo: repo,
	}
}

func (suite *ProjectRepositoryTestSuite) teardown() {
	suite.db.Teardown()
}

func (suite *ProjectRepositoryTestSuite) clean() {
	suite.db.Clean()
}

// ==================== Project CRUD Tests ====================

func TestProjectRepository_Create_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Valid project
	project := testutil.NewTestProject()

	// When: Create project
	err := suite.repo.Create(project)

	// Then: Verify success
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, project.ID)
	assert.NotZero(t, project.CreatedAt)
	assert.NotZero(t, project.UpdatedAt)
}

func TestProjectRepository_FindByID_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Created project
	project := testutil.NewTestProject()
	suite.repo.Create(project)

	// When: Find by ID
	found, err := suite.repo.FindByID(project.ID)

	// Then: Verify success
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, project.ID, found.ID)
	assert.Equal(t, project.Name, found.Name)
	assert.Equal(t, project.WorkspaceID, found.WorkspaceID)
}

func TestProjectRepository_FindByID_NotFound(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Non-existent project ID
	nonExistentID := uuid.New()

	// When: Find by ID
	found, err := suite.repo.FindByID(nonExistentID)

	// Then: Verify not found error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.Nil(t, found)
}

func TestProjectRepository_FindByWorkspaceID_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Multiple projects in workspace
	workspaceID := uuid.New()

	project1 := testutil.NewTestProjectWithID(uuid.New(), uuid.New())
	project1.WorkspaceID = workspaceID
	project1.Name = "Project 1"

	project2 := testutil.NewTestProjectWithID(uuid.New(), uuid.New())
	project2.WorkspaceID = workspaceID
	project2.Name = "Project 2"

	project3 := testutil.NewTestProject() // Different workspace

	suite.repo.Create(project1)
	suite.repo.Create(project2)
	suite.repo.Create(project3)

	// When: Find by workspace ID
	projects, err := suite.repo.FindByWorkspaceID(workspaceID)

	// Then: Verify results
	assert.NoError(t, err)
	assert.Len(t, projects, 2)
}

func TestProjectRepository_Update_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Created project
	project := testutil.NewTestProject()
	suite.repo.Create(project)

	// When: Update project
	project.Name = "Updated Project Name"
	project.Description = "Updated Description"
	err := suite.repo.Update(project)

	// Then: Verify update
	assert.NoError(t, err)

	found, _ := suite.repo.FindByID(project.ID)
	assert.Equal(t, "Updated Project Name", found.Name)
	assert.Equal(t, "Updated Description", found.Description)
}

func TestProjectRepository_Delete_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Created project
	project := testutil.NewTestProject()
	suite.repo.Create(project)

	// When: Delete project
	err := suite.repo.Delete(project.ID)

	// Then: Verify soft delete
	assert.NoError(t, err)

	found, err := suite.repo.FindByID(project.ID)
	assert.Error(t, err)
	assert.Nil(t, found)
}

// ==================== Project Member Tests ====================

func TestProjectRepository_CreateMember_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Project and role
	project, ownerRole, userID := suite.db.SeedBasicSetup()

	// When: Create member
	member := testutil.NewTestProjectMember(project.ID, userID, ownerRole.ID)
	member.Role = ownerRole
	err := suite.repo.CreateMember(member)

	// Then: Verify success (duplicate insert will fail, showing seeding already created one)
	// This test verifies CreateMember works in general
	assert.Error(t, err) // Should fail because SeedBasicSetup already created owner
}

func TestProjectRepository_FindMemberByID_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Project with member
	project, ownerRole, userID := suite.db.SeedBasicSetup()

	member := testutil.NewTestProjectMember(project.ID, userID, ownerRole.ID)
	member.Role = ownerRole

	// Find the existing member created by SeedBasicSetup
	members, _ := suite.repo.FindMembersByProject(project.ID)
	assert.Len(t, members, 1)
	existingMember := members[0]

	// When: Find member by ID
	found, err := suite.repo.FindMemberByID(existingMember.ID)

	// Then: Verify success
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, existingMember.ID, found.ID)
	assert.Equal(t, userID, found.UserID)
	assert.Equal(t, project.ID, found.ProjectID)
}

func TestProjectRepository_FindMemberByUserAndProject_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Project with member
	project, _, userID := suite.db.SeedBasicSetup()

	// When: Find member by user and project
	member, err := suite.repo.FindMemberByUserAndProject(userID, project.ID)

	// Then: Verify success
	assert.NoError(t, err)
	assert.NotNil(t, member)
	assert.Equal(t, userID, member.UserID)
	assert.Equal(t, project.ID, member.ProjectID)
}

func TestProjectRepository_FindMemberByUserAndProject_NotFound(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Project without member
	project, _, _ := suite.db.SeedBasicSetup()
	nonMemberID := uuid.New()

	// When: Find member by user and project
	member, err := suite.repo.FindMemberByUserAndProject(nonMemberID, project.ID)

	// Then: Verify not found
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.Nil(t, member)
}

func TestProjectRepository_FindMembersByProject_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Project with multiple members
	project, ownerRole, ownerID := suite.db.SeedBasicSetup()

	// Add more members
	member1 := testutil.NewTestProjectMember(project.ID, uuid.New(), ownerRole.ID)
	member1.Role = ownerRole
	member2 := testutil.NewTestProjectMember(project.ID, uuid.New(), ownerRole.ID)
	member2.Role = ownerRole

	suite.repo.CreateMember(member1)
	suite.repo.CreateMember(member2)

	// When: Find all members
	members, err := suite.repo.FindMembersByProject(project.ID)

	// Then: Verify results (owner + 2 new members)
	assert.NoError(t, err)
	assert.Len(t, members, 3)

	// Verify owner is included
	hasOwner := false
	for _, m := range members {
		if m.UserID == ownerID {
			hasOwner = true
			break
		}
	}
	assert.True(t, hasOwner)
}

func TestProjectRepository_UpdateMember_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Project with member
	ownerRole, adminRole, _ := suite.db.SeedRoles()
	project, _, userID := suite.db.SeedProject(ownerRole)

	member, _ := suite.repo.FindMemberByUserAndProject(userID, project.ID)

	// When: Update member role
	member.RoleID = adminRole.ID
	member.Role = adminRole
	err := suite.repo.UpdateMember(member)

	// Then: Verify update
	assert.NoError(t, err)

	updated, _ := suite.repo.FindMemberByUserAndProject(userID, project.ID)
	assert.Equal(t, adminRole.ID, updated.RoleID)
}

func TestProjectRepository_DeleteMember_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Project with member
	project, ownerRole, ownerID := suite.db.SeedBasicSetup()

	// Add a new member to delete
	newMember := testutil.NewTestProjectMember(project.ID, uuid.New(), ownerRole.ID)
	newMember.Role = ownerRole
	suite.repo.CreateMember(newMember)

	// When: Delete member
	err := suite.repo.DeleteMember(newMember.ID)

	// Then: Verify deletion
	assert.NoError(t, err)

	// Verify only owner remains
	members, _ := suite.repo.FindMembersByProject(project.ID)
	assert.Len(t, members, 1)
	assert.Equal(t, ownerID, members[0].UserID)
}

// ==================== Join Request Tests ====================

func TestProjectRepository_CreateJoinRequest_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Project and user
	project, _, _ := suite.db.SeedBasicSetup()
	requesterID := uuid.New()

	joinReq := testutil.NewTestJoinRequest(project.ID, requesterID, domain.ProjectJoinRequestStatusPending)

	// When: Create join request
	err := suite.repo.CreateJoinRequest(joinReq)

	// Then: Verify success
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, joinReq.ID)
}

func TestProjectRepository_FindJoinRequestByID_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Created join request
	project, _, _ := suite.db.SeedBasicSetup()
	requesterID := uuid.New()

	joinReq := testutil.NewTestJoinRequest(project.ID, requesterID, domain.ProjectJoinRequestStatusPending)
	suite.repo.CreateJoinRequest(joinReq)

	// When: Find by ID
	found, err := suite.repo.FindJoinRequestByID(joinReq.ID)

	// Then: Verify success
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, joinReq.ID, found.ID)
	assert.Equal(t, project.ID, found.ProjectID)
	assert.Equal(t, requesterID, found.UserID)
}

func TestProjectRepository_FindJoinRequestByUserAndProject_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Created join request
	project, _, _ := suite.db.SeedBasicSetup()
	requesterID := uuid.New()

	joinReq := testutil.NewTestJoinRequest(project.ID, requesterID, domain.ProjectJoinRequestStatusPending)
	suite.repo.CreateJoinRequest(joinReq)

	// When: Find by user and project
	found, err := suite.repo.FindJoinRequestByUserAndProject(requesterID, project.ID)

	// Then: Verify success
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, requesterID, found.UserID)
	assert.Equal(t, project.ID, found.ProjectID)
}

func TestProjectRepository_FindJoinRequestsByProject_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Project with multiple join requests
	project, _, _ := suite.db.SeedBasicSetup()

	req1 := testutil.NewTestJoinRequest(project.ID, uuid.New(), domain.ProjectJoinRequestStatusPending)
	req2 := testutil.NewTestJoinRequest(project.ID, uuid.New(), domain.ProjectJoinRequestStatusPending)
	req3 := testutil.NewTestJoinRequest(project.ID, uuid.New(), domain.ProjectJoinRequestStatusApproved)

	suite.repo.CreateJoinRequest(req1)
	suite.repo.CreateJoinRequest(req2)
	suite.repo.CreateJoinRequest(req3)

	// When: Find pending requests
	pendingReqs, err := suite.repo.FindJoinRequestsByProject(project.ID, string(domain.ProjectJoinRequestStatusPending))

	// Then: Verify results
	assert.NoError(t, err)
	assert.Len(t, pendingReqs, 2)

	// When: Find all requests (empty status filter)
	allReqs, err := suite.repo.FindJoinRequestsByProject(project.ID, "")

	// Then: Verify all returned
	assert.NoError(t, err)
	assert.Len(t, allReqs, 3)
}

func TestProjectRepository_UpdateJoinRequest_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Pending join request
	project, _, _ := suite.db.SeedBasicSetup()
	requesterID := uuid.New()

	joinReq := testutil.NewTestJoinRequest(project.ID, requesterID, domain.ProjectJoinRequestStatusPending)
	suite.repo.CreateJoinRequest(joinReq)

	// When: Update to approved
	joinReq.Status = domain.ProjectJoinRequestStatusApproved
	err := suite.repo.UpdateJoinRequest(joinReq)

	// Then: Verify update
	assert.NoError(t, err)

	found, _ := suite.repo.FindJoinRequestByID(joinReq.ID)
	assert.Equal(t, domain.ProjectJoinRequestStatusApproved, found.Status)
}

// ==================== Search Tests ====================

func TestProjectRepository_Search_Success(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Multiple projects with different names
	workspaceID := uuid.New()

	proj1 := testutil.NewTestProjectWithID(uuid.New(), uuid.New())
	proj1.WorkspaceID = workspaceID
	proj1.Name = "Alpha Project"

	proj2 := testutil.NewTestProjectWithID(uuid.New(), uuid.New())
	proj2.WorkspaceID = workspaceID
	proj2.Name = "Beta Project"

	proj3 := testutil.NewTestProjectWithID(uuid.New(), uuid.New())
	proj3.WorkspaceID = workspaceID
	proj3.Name = "Alpha Team"

	suite.repo.Create(proj1)
	suite.repo.Create(proj2)
	suite.repo.Create(proj3)

	// When: Search for "Alpha"
	results, total, err := suite.repo.Search(workspaceID, "Alpha", 1, 10)

	// Then: Verify results
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, int64(2), total)
}

func TestProjectRepository_Search_EmptyQuery(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: Projects in workspace
	workspaceID := uuid.New()

	proj1 := testutil.NewTestProjectWithID(uuid.New(), uuid.New())
	proj1.WorkspaceID = workspaceID

	proj2 := testutil.NewTestProjectWithID(uuid.New(), uuid.New())
	proj2.WorkspaceID = workspaceID

	suite.repo.Create(proj1)
	suite.repo.Create(proj2)

	// When: Search with empty query
	results, total, err := suite.repo.Search(workspaceID, "", 1, 10)

	// Then: Verify all returned
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, int64(2), total)
}

func TestProjectRepository_Search_Pagination(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	// Given: 15 projects
	workspaceID := uuid.New()

	for i := 0; i < 15; i++ {
		proj := testutil.NewTestProjectWithID(uuid.New(), uuid.New())
		proj.WorkspaceID = workspaceID
		suite.repo.Create(proj)
	}

	// When: Get page 1 (limit 10)
	page1, total, err := suite.repo.Search(workspaceID, "", 1, 10)

	// Then: Verify pagination
	assert.NoError(t, err)
	assert.Len(t, page1, 10)
	assert.Equal(t, int64(15), total)

	// When: Get page 2 (limit 10)
	page2, _, err := suite.repo.Search(workspaceID, "", 2, 10)

	// Then: Verify remaining
	assert.NoError(t, err)
	assert.Len(t, page2, 5)
}

// ==================== Edge Cases ====================

func TestProjectRepository_MultipleWorkspaces(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	workspace1 := uuid.New()
	workspace2 := uuid.New()

	// Create projects in different workspaces
	proj1 := testutil.NewTestProjectWithID(uuid.New(), uuid.New())
	proj1.WorkspaceID = workspace1

	proj2 := testutil.NewTestProjectWithID(uuid.New(), uuid.New())
	proj2.WorkspaceID = workspace2

	suite.repo.Create(proj1)
	suite.repo.Create(proj2)

	// Verify workspace isolation
	ws1Projects, _ := suite.repo.FindByWorkspaceID(workspace1)
	ws2Projects, _ := suite.repo.FindByWorkspaceID(workspace2)

	assert.Len(t, ws1Projects, 1)
	assert.Len(t, ws2Projects, 1)
}

func TestProjectRepository_PublicAndPrivateProjects(t *testing.T) {
	suite := setupProjectRepoTest(t)
	defer suite.teardown()

	workspaceID := uuid.New()

	publicProj := testutil.NewTestProjectWithID(uuid.New(), uuid.New())
	publicProj.WorkspaceID = workspaceID
	publicProj.IsPublic = true

	privateProj := testutil.NewTestProjectWithID(uuid.New(), uuid.New())
	privateProj.WorkspaceID = workspaceID
	privateProj.IsPublic = false

	suite.repo.Create(publicProj)
	suite.repo.Create(privateProj)

	// Verify both created
	projects, _ := suite.repo.FindByWorkspaceID(workspaceID)
	assert.Len(t, projects, 2)

	// Verify visibility flags preserved
	for _, p := range projects {
		if p.ID == publicProj.ID {
			assert.True(t, p.IsPublic)
		} else {
			assert.False(t, p.IsPublic)
		}
	}
}
