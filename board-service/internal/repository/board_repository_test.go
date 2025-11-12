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

type BoardRepositoryTestSuite struct {
	db   *testutil.TestDB
	repo repository.BoardRepository
}

func setupBoardRepoTest(t *testing.T) *BoardRepositoryTestSuite {
	db := testutil.SetupTestDB()
	repo := repository.NewBoardRepository(db.DB)

	return &BoardRepositoryTestSuite{
		db:   db,
		repo: repo,
	}
}

func (suite *BoardRepositoryTestSuite) teardown() {
	suite.db.Teardown()
}

func (suite *BoardRepositoryTestSuite) clean() {
	suite.db.Clean()
}

// ==================== Create Tests ====================

func TestBoardRepository_Create_Success(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	// Arrange
	board := testutil.NewTestBoard(uuid.New(), uuid.New())

	// Act
	err := suite.repo.Create(board)

	// Assert
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, board.ID)
	assert.NotZero(t, board.CreatedAt)
	assert.NotZero(t, board.UpdatedAt)
}

func TestBoardRepository_Create_InitializesCustomFieldsCache(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	board := testutil.NewTestBoard(uuid.New(), uuid.New())

	err := suite.repo.Create(board)
	assert.NoError(t, err)

	// Verify custom_fields_cache is initialized
	assert.NotEmpty(t, board.CustomFieldsCache)
	assert.Equal(t, "{}", board.CustomFieldsCache)
}

// ==================== FindByID Tests ====================

func TestBoardRepository_FindByID_Success(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	// Create board
	board := testutil.NewTestBoard(uuid.New(), uuid.New())
	suite.repo.Create(board)

	// Find by ID
	found, err := suite.repo.FindByID(board.ID)

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, board.ID, found.ID)
	assert.Equal(t, board.Title, found.Title)
	assert.Equal(t, board.ProjectID, found.ProjectID)
}

func TestBoardRepository_FindByID_NotFound(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	nonExistentID := uuid.New()
	found, err := suite.repo.FindByID(nonExistentID)

	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.Nil(t, found)
}

func TestBoardRepository_FindByID_IgnoresDeletedBoards(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	// Create and delete board
	board := testutil.NewTestBoard(uuid.New(), uuid.New())
	suite.repo.Create(board)
	suite.repo.Delete(board.ID)

	// Should not find deleted board
	found, err := suite.repo.FindByID(board.ID)

	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.Nil(t, found)
}

// ==================== FindByProject Tests ====================

func TestBoardRepository_FindByProject_Success(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	projectID := uuid.New()
	userID := uuid.New()

	// Create multiple boards
	board1 := testutil.NewTestBoard(projectID, userID)
	board2 := testutil.NewTestBoard(projectID, userID)
	board3 := testutil.NewTestBoard(uuid.New(), userID) // Different project

	suite.repo.Create(board1)
	suite.repo.Create(board2)
	suite.repo.Create(board3)

	// Find boards by project
	filters := repository.BoardFilters{}
	boards, total, err := suite.repo.FindByProject(projectID, filters, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(boards))
	assert.Equal(t, int64(2), total)
}

func TestBoardRepository_FindByProject_WithFilters_AssigneeID(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	projectID := uuid.New()
	assigneeID := uuid.New()
	otherUserID := uuid.New()

	board1 := testutil.NewTestBoardWithAssignee(projectID, uuid.New(), assigneeID)
	board2 := testutil.NewTestBoardWithAssignee(projectID, uuid.New(), otherUserID)

	suite.repo.Create(board1)
	suite.repo.Create(board2)

	// Filter by assignee
	filters := repository.BoardFilters{AssigneeID: assigneeID}
	boards, total, err := suite.repo.FindByProject(projectID, filters, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(boards))
	assert.Equal(t, int64(1), total)
	assert.Equal(t, assigneeID, *boards[0].AssigneeID)
}

func TestBoardRepository_FindByProject_WithFilters_AuthorID(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	projectID := uuid.New()
	authorID := uuid.New()
	otherAuthorID := uuid.New()

	board1 := testutil.NewTestBoard(projectID, authorID)
	board2 := testutil.NewTestBoard(projectID, otherAuthorID)

	suite.repo.Create(board1)
	suite.repo.Create(board2)

	// Filter by author
	filters := repository.BoardFilters{AuthorID: authorID}
	boards, total, err := suite.repo.FindByProject(projectID, filters, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(boards))
	assert.Equal(t, int64(1), total)
	assert.Equal(t, authorID, boards[0].CreatedBy)
}

func TestBoardRepository_FindByProject_Pagination(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	projectID := uuid.New()
	userID := uuid.New()

	// Create 25 boards
	for i := 0; i < 25; i++ {
		board := testutil.NewTestBoard(projectID, userID)
		suite.repo.Create(board)
	}

	// Page 1 (limit 10)
	filters := repository.BoardFilters{}
	page1Boards, total, err := suite.repo.FindByProject(projectID, filters, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 10, len(page1Boards))
	assert.Equal(t, int64(25), total)

	// Page 2 (limit 10)
	page2Boards, _, err := suite.repo.FindByProject(projectID, filters, 2, 10)

	assert.NoError(t, err)
	assert.Equal(t, 10, len(page2Boards))

	// Page 3 (limit 10, should have 5)
	page3Boards, _, err := suite.repo.FindByProject(projectID, filters, 3, 10)

	assert.NoError(t, err)
	assert.Equal(t, 5, len(page3Boards))
}

func TestBoardRepository_FindByProject_EmptyResult(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	projectID := uuid.New()
	filters := repository.BoardFilters{}
	boards, total, err := suite.repo.FindByProject(projectID, filters, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(boards))
	assert.Equal(t, int64(0), total)
}

// ==================== Update Tests ====================

func TestBoardRepository_Update_Success(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	// Create board
	board := testutil.NewTestBoard(uuid.New(), uuid.New())
	suite.repo.Create(board)

	// Update board
	board.Title = "Updated Title"
	board.Description = "Updated Description"
	err := suite.repo.Update(board)

	assert.NoError(t, err)

	// Verify update
	found, _ := suite.repo.FindByID(board.ID)
	assert.Equal(t, "Updated Title", found.Title)
	assert.Equal(t, "Updated Description", found.Description)
}

func TestBoardRepository_Update_UpdatesTimestamp(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	board := testutil.NewTestBoard(uuid.New(), uuid.New())
	suite.repo.Create(board)

	originalUpdatedAt := board.UpdatedAt

	// Update
	board.Title = "New Title"
	suite.repo.Update(board)

	// Verify UpdatedAt changed
	found, _ := suite.repo.FindByID(board.ID)
	assert.True(t, found.UpdatedAt.After(originalUpdatedAt))
}

// ==================== Delete Tests (Soft Delete) ====================

func TestBoardRepository_Delete_SoftDelete(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	board := testutil.NewTestBoard(uuid.New(), uuid.New())
	suite.repo.Create(board)

	// Delete board
	err := suite.repo.Delete(board.ID)
	assert.NoError(t, err)

	// Should not be findable
	found, err := suite.repo.FindByID(board.ID)
	assert.Error(t, err)
	assert.Nil(t, found)

	// Verify it's soft deleted (still in DB with is_deleted=true)
	var deletedBoard domain.Board
	suite.db.DB.Unscoped().Where("id = ?", board.ID).First(&deletedBoard)
	assert.True(t, deletedBoard.IsDeleted)
}

func TestBoardRepository_Delete_NotFoundError(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	nonExistentID := uuid.New()
	err := suite.repo.Delete(nonExistentID)

	// Should not error (soft delete updates 0 rows)
	assert.NoError(t, err)
}

// ==================== Edge Cases ====================

func TestBoardRepository_ConcurrentCreates(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	projectID := uuid.New()
	userID := uuid.New()

	// Create multiple boards concurrently
	done := make(chan bool)
	count := 10

	for i := 0; i < count; i++ {
		go func() {
			board := testutil.NewTestBoard(projectID, userID)
			suite.repo.Create(board)
			done <- true
		}()
	}

	// Wait for all
	for i := 0; i < count; i++ {
		<-done
	}

	// Verify all created
	filters := repository.BoardFilters{}
	boards, total, _ := suite.repo.FindByProject(projectID, filters, 1, 20)

	assert.Equal(t, count, len(boards))
	assert.Equal(t, int64(count), total)
}

func TestBoardRepository_WithAssignee_Update(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	assignee1 := uuid.New()
	assignee2 := uuid.New()

	board := testutil.NewTestBoardWithAssignee(uuid.New(), uuid.New(), assignee1)
	suite.repo.Create(board)

	// Update assignee
	board.AssigneeID = &assignee2
	suite.repo.Update(board)

	// Verify
	found, _ := suite.repo.FindByID(board.ID)
	assert.NotNil(t, found.AssigneeID)
	assert.Equal(t, assignee2, *found.AssigneeID)
}

func TestBoardRepository_RemoveAssignee(t *testing.T) {
	suite := setupBoardRepoTest(t)
	defer suite.teardown()

	assignee := uuid.New()
	board := testutil.NewTestBoardWithAssignee(uuid.New(), uuid.New(), assignee)
	suite.repo.Create(board)

	// Remove assignee
	board.AssigneeID = nil
	suite.repo.Update(board)

	// Verify
	found, _ := suite.repo.FindByID(board.ID)
	assert.Nil(t, found.AssigneeID)
}
