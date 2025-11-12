package repository_test

import (
	"board-service/internal/domain"
	"board-service/internal/repository"
	"board-service/internal/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// ==================== Test Suite Setup ====================

type CommentRepositoryTestSuite struct {
	db   *testutil.TestDB
	repo repository.CommentRepository
}

func setupCommentRepoTest(t *testing.T) *CommentRepositoryTestSuite {
	db := testutil.SetupTestDB()
	repo := repository.NewCommentRepository(db.DB)

	return &CommentRepositoryTestSuite{
		db:   db,
		repo: repo,
	}
}

func (suite *CommentRepositoryTestSuite) teardown() {
	suite.db.Teardown()
}

func (suite *CommentRepositoryTestSuite) clean() {
	suite.db.Clean()
}

// ==================== Create Tests ====================

func TestCommentRepository_Create_Success(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	// Given: Valid comment
	boardID := uuid.New()
	userID := uuid.New()
	comment := &domain.Comment{
		BoardID: boardID,
		UserID:  userID,
		Content: "This is a test comment",
	}

	// When: Create comment
	err := suite.repo.Create(comment)

	// Then: Verify success
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, comment.ID)
	assert.NotZero(t, comment.CreatedAt)
	assert.NotZero(t, comment.UpdatedAt)
}

func TestCommentRepository_Create_InitializesTimestamps(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	comment := &domain.Comment{
		BoardID: uuid.New(),
		UserID:  uuid.New(),
		Content: "Test comment",
	}

	before := time.Now()
	err := suite.repo.Create(comment)
	after := time.Now()

	assert.NoError(t, err)
	assert.True(t, comment.CreatedAt.After(before) || comment.CreatedAt.Equal(before))
	assert.True(t, comment.CreatedAt.Before(after) || comment.CreatedAt.Equal(after))
	assert.Equal(t, comment.CreatedAt, comment.UpdatedAt)
}

func TestCommentRepository_Create_MultipleComments(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	boardID := uuid.New()
	userID := uuid.New()

	// Create 3 comments
	for i := 0; i < 3; i++ {
		comment := &domain.Comment{
			BoardID: boardID,
			UserID:  userID,
			Content: "Comment " + string(rune('A'+i)),
		}
		err := suite.repo.Create(comment)
		assert.NoError(t, err)
	}

	// Verify all created
	comments, err := suite.repo.FindByBoardID(boardID)
	assert.NoError(t, err)
	assert.Len(t, comments, 3)
}

// ==================== FindByID Tests ====================

func TestCommentRepository_FindByID_Success(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	// Given: Created comment
	comment := &domain.Comment{
		BoardID: uuid.New(),
		UserID:  uuid.New(),
		Content: "Test comment",
	}
	suite.repo.Create(comment)

	// When: Find by ID
	found, err := suite.repo.FindByID(comment.ID)

	// Then: Verify success
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, comment.ID, found.ID)
	assert.Equal(t, comment.Content, found.Content)
	assert.Equal(t, comment.BoardID, found.BoardID)
	assert.Equal(t, comment.UserID, found.UserID)
}

func TestCommentRepository_FindByID_NotFound(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	// Given: Non-existent comment ID
	nonExistentID := uuid.New()

	// When: Find by ID
	found, err := suite.repo.FindByID(nonExistentID)

	// Then: Verify not found error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.Nil(t, found)
}

func TestCommentRepository_FindByID_IgnoresDeletedComments(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	// Given: Created and deleted comment
	comment := &domain.Comment{
		BoardID: uuid.New(),
		UserID:  uuid.New(),
		Content: "Test comment",
	}
	suite.repo.Create(comment)
	suite.repo.Delete(comment.ID)

	// When: Find by ID
	found, err := suite.repo.FindByID(comment.ID)

	// Then: Verify not found
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.Nil(t, found)
}

// ==================== FindByBoardID Tests ====================

func TestCommentRepository_FindByBoardID_Success(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	// Given: Board with multiple comments
	boardID := uuid.New()
	userID := uuid.New()

	comment1 := &domain.Comment{BoardID: boardID, UserID: userID, Content: "First comment"}
	comment2 := &domain.Comment{BoardID: boardID, UserID: userID, Content: "Second comment"}
	comment3 := &domain.Comment{BoardID: uuid.New(), UserID: userID, Content: "Other board comment"}

	suite.repo.Create(comment1)
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	suite.repo.Create(comment2)
	suite.repo.Create(comment3)

	// When: Find by board ID
	comments, err := suite.repo.FindByBoardID(boardID)

	// Then: Verify results
	assert.NoError(t, err)
	assert.Len(t, comments, 2)
	assert.Equal(t, "First comment", comments[0].Content)
	assert.Equal(t, "Second comment", comments[1].Content)
}

func TestCommentRepository_FindByBoardID_OrderedByCreatedAt(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	boardID := uuid.New()
	userID := uuid.New()

	// Create comments with explicit time gaps
	comment1 := &domain.Comment{BoardID: boardID, UserID: userID, Content: "First"}
	suite.repo.Create(comment1)
	time.Sleep(10 * time.Millisecond)

	comment2 := &domain.Comment{BoardID: boardID, UserID: userID, Content: "Second"}
	suite.repo.Create(comment2)
	time.Sleep(10 * time.Millisecond)

	comment3 := &domain.Comment{BoardID: boardID, UserID: userID, Content: "Third"}
	suite.repo.Create(comment3)

	// Retrieve comments
	comments, err := suite.repo.FindByBoardID(boardID)

	assert.NoError(t, err)
	assert.Len(t, comments, 3)
	// Verify ascending order by created_at
	assert.True(t, comments[0].CreatedAt.Before(comments[1].CreatedAt) || comments[0].CreatedAt.Equal(comments[1].CreatedAt))
	assert.True(t, comments[1].CreatedAt.Before(comments[2].CreatedAt) || comments[1].CreatedAt.Equal(comments[2].CreatedAt))
}

func TestCommentRepository_FindByBoardID_EmptyResult(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	// Given: Board with no comments
	boardID := uuid.New()

	// When: Find by board ID
	comments, err := suite.repo.FindByBoardID(boardID)

	// Then: Verify empty list
	assert.NoError(t, err)
	assert.Len(t, comments, 0)
	assert.NotNil(t, comments) // Should be empty slice, not nil
}

func TestCommentRepository_FindByBoardID_IgnoresDeletedComments(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	boardID := uuid.New()
	userID := uuid.New()

	comment1 := &domain.Comment{BoardID: boardID, UserID: userID, Content: "Active comment"}
	comment2 := &domain.Comment{BoardID: boardID, UserID: userID, Content: "Deleted comment"}

	suite.repo.Create(comment1)
	suite.repo.Create(comment2)

	// Delete comment2
	suite.repo.Delete(comment2.ID)

	// Find comments
	comments, err := suite.repo.FindByBoardID(boardID)

	assert.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.Equal(t, "Active comment", comments[0].Content)
}

// ==================== Update Tests ====================

func TestCommentRepository_Update_Success(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	// Given: Created comment
	comment := &domain.Comment{
		BoardID: uuid.New(),
		UserID:  uuid.New(),
		Content: "Original content",
	}
	suite.repo.Create(comment)

	// When: Update comment
	comment.Content = "Updated content"
	err := suite.repo.Update(comment)

	// Then: Verify update
	assert.NoError(t, err)

	found, _ := suite.repo.FindByID(comment.ID)
	assert.Equal(t, "Updated content", found.Content)
}

func TestCommentRepository_Update_UpdatesTimestamp(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	comment := &domain.Comment{
		BoardID: uuid.New(),
		UserID:  uuid.New(),
		Content: "Original content",
	}
	suite.repo.Create(comment)

	originalUpdatedAt := comment.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	// Update
	comment.Content = "New content"
	suite.repo.Update(comment)

	// Verify UpdatedAt changed
	found, _ := suite.repo.FindByID(comment.ID)
	assert.True(t, found.UpdatedAt.After(originalUpdatedAt))
	assert.Equal(t, comment.CreatedAt, found.CreatedAt) // CreatedAt should not change
}

func TestCommentRepository_Update_NonExistentComment(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	// Given: Non-existent comment
	comment := &domain.Comment{
		ID:      uuid.New(),
		BoardID: uuid.New(),
		UserID:  uuid.New(),
		Content: "Test content",
	}

	// When: Update non-existent comment
	err := suite.repo.Update(comment)

	// Then: Should not error (GORM behavior - updates 0 rows)
	assert.NoError(t, err)
}

// ==================== Delete Tests ====================

func TestCommentRepository_Delete_Success(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	// Given: Created comment
	comment := &domain.Comment{
		BoardID: uuid.New(),
		UserID:  uuid.New(),
		Content: "Test comment",
	}
	suite.repo.Create(comment)

	// When: Delete comment
	err := suite.repo.Delete(comment.ID)

	// Then: Verify soft delete
	assert.NoError(t, err)

	// Should not be findable
	found, err := suite.repo.FindByID(comment.ID)
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestCommentRepository_Delete_SoftDelete(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	comment := &domain.Comment{
		BoardID: uuid.New(),
		UserID:  uuid.New(),
		Content: "Test comment",
	}
	suite.repo.Create(comment)

	// Delete
	suite.repo.Delete(comment.ID)

	// Verify it's soft deleted (still in DB with is_deleted=true)
	var deletedComment domain.Comment
	err := suite.db.DB.Unscoped().Where("id = ?", comment.ID).First(&deletedComment).Error
	assert.NoError(t, err)
	assert.True(t, deletedComment.IsDeleted)
}

func TestCommentRepository_Delete_NonExistentComment(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	// Given: Non-existent comment ID
	nonExistentID := uuid.New()

	// When: Delete non-existent comment
	err := suite.repo.Delete(nonExistentID)

	// Then: Should not error (soft delete updates 0 rows)
	assert.NoError(t, err)
}

func TestCommentRepository_Delete_IdempotentDelete(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	comment := &domain.Comment{
		BoardID: uuid.New(),
		UserID:  uuid.New(),
		Content: "Test comment",
	}
	suite.repo.Create(comment)

	// Delete twice
	err1 := suite.repo.Delete(comment.ID)
	err2 := suite.repo.Delete(comment.ID)

	assert.NoError(t, err1)
	assert.NoError(t, err2) // Second delete should not error
}

// ==================== Edge Cases ====================

func TestCommentRepository_ConcurrentCreates(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	boardID := uuid.New()
	userID := uuid.New()

	// Create comments concurrently
	done := make(chan bool)
	count := 10

	for i := 0; i < count; i++ {
		go func(index int) {
			comment := &domain.Comment{
				BoardID: boardID,
				UserID:  userID,
				Content: "Concurrent comment",
			}
			suite.repo.Create(comment)
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < count; i++ {
		<-done
	}

	// Verify all created
	comments, _ := suite.repo.FindByBoardID(boardID)
	assert.Equal(t, count, len(comments))
}

func TestCommentRepository_LongContent(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	// Create comment with very long content
	longContent := string(make([]byte, 5000)) // 5KB content
	for i := range longContent {
		longContent = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " + longContent
	}

	comment := &domain.Comment{
		BoardID: uuid.New(),
		UserID:  uuid.New(),
		Content: longContent,
	}

	err := suite.repo.Create(comment)
	assert.NoError(t, err)

	// Verify content preserved
	found, _ := suite.repo.FindByID(comment.ID)
	assert.Equal(t, longContent, found.Content)
}

func TestCommentRepository_DifferentUsers(t *testing.T) {
	suite := setupCommentRepoTest(t)
	defer suite.teardown()

	boardID := uuid.New()
	user1 := uuid.New()
	user2 := uuid.New()
	user3 := uuid.New()

	// Create comments from different users
	comment1 := &domain.Comment{BoardID: boardID, UserID: user1, Content: "User 1 comment"}
	comment2 := &domain.Comment{BoardID: boardID, UserID: user2, Content: "User 2 comment"}
	comment3 := &domain.Comment{BoardID: boardID, UserID: user3, Content: "User 3 comment"}

	suite.repo.Create(comment1)
	suite.repo.Create(comment2)
	suite.repo.Create(comment3)

	// Find all
	comments, err := suite.repo.FindByBoardID(boardID)
	assert.NoError(t, err)
	assert.Len(t, comments, 3)

	// Verify different users
	userSet := make(map[uuid.UUID]bool)
	for _, c := range comments {
		userSet[c.UserID] = true
	}
	assert.Len(t, userSet, 3)
}
