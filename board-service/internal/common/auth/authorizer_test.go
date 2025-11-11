package auth_test

import (
	"board-service/internal/common/auth"
	"board-service/internal/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type AuthorizerTestSuite struct {
	db         *testutil.TestDB
	authorizer auth.ProjectAuthorizer
}

func setupAuthorizerTest(t *testing.T) *AuthorizerTestSuite {
	db := testutil.SetupTestDB()

	// Create repositories
	projectRepo := testutil.NewProjectRepository(db.DB)
	roleRepo := testutil.NewRoleRepository(db.DB)

	authorizer := auth.NewProjectAuthorizer(projectRepo, roleRepo)

	return &AuthorizerTestSuite{
		db:         db,
		authorizer: authorizer,
	}
}

func (suite *AuthorizerTestSuite) teardown() {
	suite.db.Teardown()
}

// ==================== RequireMember Tests ====================

func TestRequireMember_Success(t *testing.T) {
	suite := setupAuthorizerTest(t)
	defer suite.teardown()

	// Seed data
	project, ownerRole, userID := suite.db.SeedBasicSetup()

	// Test
	member, err := suite.authorizer.RequireMember(userID, project.ID)

	assert.NoError(t, err)
	assert.NotNil(t, member)
	assert.Equal(t, userID, member.UserID)
	assert.Equal(t, project.ID, member.ProjectID)
	assert.NotNil(t, member.Role)
	assert.Equal(t, ownerRole.ID, member.Role.ID)
}

func TestRequireMember_NotMember(t *testing.T) {
	suite := setupAuthorizerTest(t)
	defer suite.teardown()

	project, _, _ := suite.db.SeedBasicSetup()
	otherUserID := uuid.New()

	member, err := suite.authorizer.RequireMember(otherUserID, project.ID)

	assert.Error(t, err)
	assert.Nil(t, member)
}

// ==================== RequireAdmin Tests ====================

func TestRequireAdmin_Success_AsOwner(t *testing.T) {
	suite := setupAuthorizerTest(t)
	defer suite.teardown()

	project, _, userID := suite.db.SeedBasicSetup()

	member, err := suite.authorizer.RequireAdmin(userID, project.ID)

	assert.NoError(t, err)
	assert.NotNil(t, member)
	assert.True(t, member.Role.Level >= auth.RoleLevelAdmin)
}

func TestRequireAdmin_Forbidden_AsMember(t *testing.T) {
	suite := setupAuthorizerTest(t)
	defer suite.teardown()

	_, _, memberRole := suite.db.SeedRoles()
	project := testutil.NewTestProject()
	suite.db.DB.Create(project)

	memberUserID := uuid.New()
	member := testutil.NewTestProjectMember(project.ID, memberUserID, memberRole.ID)
	member.Role = memberRole
	suite.db.DB.Create(member)

	_, err := suite.authorizer.RequireAdmin(memberUserID, project.ID)

	assert.Error(t, err)
}

// ==================== RequireOwner Tests ====================

func TestRequireOwner_Success(t *testing.T) {
	suite := setupAuthorizerTest(t)
	defer suite.teardown()

	project, _, userID := suite.db.SeedBasicSetup()

	member, err := suite.authorizer.RequireOwner(userID, project.ID)

	assert.NoError(t, err)
	assert.NotNil(t, member)
	assert.Equal(t, "OWNER", member.Role.Name)
}

func TestRequireOwner_Forbidden_AsAdmin(t *testing.T) {
	suite := setupAuthorizerTest(t)
	defer suite.teardown()

	_, adminRole, _ := suite.db.SeedRoles()
	project := testutil.NewTestProject()
	suite.db.DB.Create(project)

	adminUserID := uuid.New()
	member := testutil.NewTestProjectMember(project.ID, adminUserID, adminRole.ID)
	member.Role = adminRole
	suite.db.DB.Create(member)

	_, err := suite.authorizer.RequireOwner(adminUserID, project.ID)

	assert.Error(t, err)
}

// ==================== CanEdit Tests ====================

func TestCanEdit_AsAuthor(t *testing.T) {
	suite := setupAuthorizerTest(t)
	defer suite.teardown()

	project, _, userID := suite.db.SeedBasicSetup()

	canEdit, err := suite.authorizer.CanEdit(userID, project.ID, userID)

	assert.NoError(t, err)
	assert.True(t, canEdit)
}

func TestCanEdit_AsAdmin_NotAuthor(t *testing.T) {
	suite := setupAuthorizerTest(t)
	defer suite.teardown()

	_, adminRole, _ := suite.db.SeedRoles()
	project := testutil.NewTestProject()
	suite.db.DB.Create(project)

	adminUserID := uuid.New()
	authorID := uuid.New()

	member := testutil.NewTestProjectMember(project.ID, adminUserID, adminRole.ID)
	member.Role = adminRole
	suite.db.DB.Create(member)

	canEdit, err := suite.authorizer.CanEdit(adminUserID, project.ID, authorID)

	assert.NoError(t, err)
	assert.True(t, canEdit) // Admin can edit others' content
}

func TestCanEdit_AsMember_NotAuthor(t *testing.T) {
	suite := setupAuthorizerTest(t)
	defer suite.teardown()

	_, _, memberRole := suite.db.SeedRoles()
	project := testutil.NewTestProject()
	suite.db.DB.Create(project)

	memberUserID := uuid.New()
	authorID := uuid.New()

	member := testutil.NewTestProjectMember(project.ID, memberUserID, memberRole.ID)
	member.Role = memberRole
	suite.db.DB.Create(member)

	canEdit, err := suite.authorizer.CanEdit(memberUserID, project.ID, authorID)

	assert.NoError(t, err)
	assert.False(t, canEdit) // Member cannot edit others' content
}

// ==================== Helper Function Tests ====================

func TestIsAdmin(t *testing.T) {
	ownerRole := testutil.NewOwnerRole()
	adminRole := testutil.NewAdminRole()
	memberRole := testutil.NewMemberRole()

	assert.True(t, auth.IsAdmin(ownerRole))
	assert.True(t, auth.IsAdmin(adminRole))
	assert.False(t, auth.IsAdmin(memberRole))
	assert.False(t, auth.IsAdmin(nil))
}

func TestIsOwner(t *testing.T) {
	ownerRole := testutil.NewOwnerRole()
	adminRole := testutil.NewAdminRole()

	assert.True(t, auth.IsOwner(ownerRole))
	assert.False(t, auth.IsOwner(adminRole))
	assert.False(t, auth.IsOwner(nil))
}
