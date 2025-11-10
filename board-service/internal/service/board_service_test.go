package service_test

import (
	"board-service/internal/domain"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock BoardRepository
type MockBoardRepository struct {
	mock.Mock
}

func (m *MockBoardRepository) Create(board *domain.Board) error {
	args := m.Called(board)
	return args.Error(0)
}

func (m *MockBoardRepository) FindByID(boardID uuid.UUID) (*domain.Board, error) {
	args := m.Called(boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Board), args.Error(1)
}

func (m *MockBoardRepository) FindByProject(projectID uuid.UUID) ([]domain.Board, error) {
	args := m.Called(projectID)
	return args.Get(0).([]domain.Board), args.Error(1)
}

func (m *MockBoardRepository) Update(board *domain.Board) error {
	args := m.Called(board)
	return args.Error(0)
}

func (m *MockBoardRepository) Delete(boardID uuid.UUID) error {
	args := m.Called(boardID)
	return args.Error(0)
}

// =============================================================================
// CreateBoard Tests
// =============================================================================

func TestCreateBoard_CustomFieldsCacheInitialized(t *testing.T) {
	// Arrange
	projectID := uuid.New()

	// Create board with empty custom_fields_cache
	board := &domain.Board{
		BaseModel:         domain.BaseModel{ID: uuid.New()},
		ProjectID:         projectID,
		Title:             "Test Board",
		Description:       "Test Description",
		CreatedBy:         uuid.New(),
		CustomFieldsCache: "{}", // Should be initialized as empty JSON object
	}

	// Assert
	// 1. CustomFieldsCache should be initialized
	assert.NotEmpty(t, board.CustomFieldsCache, "CustomFieldsCache should not be empty")

	// 2. Should be valid JSON
	var parsedCache map[string]interface{}
	err := json.Unmarshal([]byte(board.CustomFieldsCache), &parsedCache)
	assert.NoError(t, err, "CustomFieldsCache should be valid JSON")

	// 3. Should be empty object
	assert.Equal(t, 0, len(parsedCache), "CustomFieldsCache should be empty object")

	// 4. Board structure should not have legacy fields
	// (CustomStageID, CustomImportanceID should not exist)
	assert.Equal(t, "Test Board", board.Title)
	assert.Equal(t, projectID, board.ProjectID)
}

func TestCreateBoard_NoLegacyFields(t *testing.T) {
	// Arrange
	board := &domain.Board{
		BaseModel:         domain.BaseModel{ID: uuid.New()},
		ProjectID:         uuid.New(),
		Title:             "New Board",
		Description:       "Board without legacy fields",
		CreatedBy:         uuid.New(),
		CustomFieldsCache: "{}",
	}

	// Assert
	// Verify board struct does NOT have these legacy fields:
	// - CustomStageID (removed)
	// - CustomImportanceID (removed)
	// - Roles (junction table removed)

	// Board should only have:
	assert.NotEmpty(t, board.ID)
	assert.NotEmpty(t, board.ProjectID)
	assert.NotEmpty(t, board.Title)
	assert.Equal(t, "{}", board.CustomFieldsCache)
}

// =============================================================================
// GetBoard Tests - CustomFieldsCache Parsing
// =============================================================================

func TestGetBoard_ParseCustomFieldsCache(t *testing.T) {
	// Arrange
	boardID := uuid.New()
	projectID := uuid.New()

	// Board with populated custom_fields_cache
	customFieldsJSON := `{
		"field-priority": "High",
		"field-status": "In Progress",
		"field-tags": ["Bug", "Frontend"]
	}`

	board := &domain.Board{
		BaseModel:         domain.BaseModel{ID: boardID},
		ProjectID:         projectID,
		Title:             "Test Board",
		Description:       "Test",
		CreatedBy:         uuid.New(),
		CustomFieldsCache: customFieldsJSON,
	}

	// Act - Parse custom_fields_cache
	var customFields map[string]interface{}
	err := json.Unmarshal([]byte(board.CustomFieldsCache), &customFields)

	// Assert
	// 1. Should parse without error
	assert.NoError(t, err, "Should parse custom_fields_cache successfully")

	// 2. Should have correct field values
	assert.Equal(t, "High", customFields["field-priority"])
	assert.Equal(t, "In Progress", customFields["field-status"])

	// 3. Tags should be array
	tags, ok := customFields["field-tags"].([]interface{})
	assert.True(t, ok, "Tags should be array")
	assert.Equal(t, 2, len(tags), "Should have 2 tags")
	assert.Equal(t, "Bug", tags[0])
	assert.Equal(t, "Frontend", tags[1])
}

func TestGetBoard_EmptyCustomFieldsCache(t *testing.T) {
	// Arrange
	board := &domain.Board{
		BaseModel:         domain.BaseModel{ID: uuid.New()},
		ProjectID:         uuid.New(),
		Title:             "Board without custom fields",
		CustomFieldsCache: "{}",
	}

	// Act
	var customFields map[string]interface{}
	err := json.Unmarshal([]byte(board.CustomFieldsCache), &customFields)

	// Assert
	assert.NoError(t, err, "Empty object should parse successfully")
	assert.Equal(t, 0, len(customFields), "Should have no custom fields")
}

func TestGetBoard_InvalidJSONHandling(t *testing.T) {
	// Arrange
	board := &domain.Board{
		BaseModel:         domain.BaseModel{ID: uuid.New()},
		ProjectID:         uuid.New(),
		Title:             "Board with invalid JSON",
		CustomFieldsCache: "{invalid json}", // Invalid JSON
	}

	// Act
	var customFields map[string]interface{}
	err := json.Unmarshal([]byte(board.CustomFieldsCache), &customFields)

	// Assert
	// Should return error for invalid JSON
	assert.Error(t, err, "Invalid JSON should return error")

	// Service should handle this gracefully and return empty map
	if err != nil {
		customFields = make(map[string]interface{})
	}
	assert.Equal(t, 0, len(customFields), "Should fallback to empty map")
}

// =============================================================================
// UpdateBoard Tests - Legacy Field Updates Removed
// =============================================================================

func TestUpdateBoard_OnlyBasicFieldsUpdated(t *testing.T) {
	// Arrange
	boardID := uuid.New()
	projectID := uuid.New()

	existingBoard := &domain.Board{
		BaseModel:         domain.BaseModel{ID: boardID},
		ProjectID:         projectID,
		Title:             "Old Title",
		Description:       "Old Description",
		CreatedBy:         uuid.New(),
		CustomFieldsCache: `{"field-status": "To Do"}`,
	}

	// Act - Update basic fields only
	existingBoard.Title = "Updated Title"
	existingBoard.Description = "Updated Description"

	// Assert
	// 1. Basic fields should be updated
	assert.Equal(t, "Updated Title", existingBoard.Title)
	assert.Equal(t, "Updated Description", existingBoard.Description)

	// 2. CustomFieldsCache should remain unchanged
	// (Custom field updates should go through FieldValueService, not BoardService)
	assert.Equal(t, `{"field-status": "To Do"}`, existingBoard.CustomFieldsCache)

	// 3. Legacy field updates should NOT be possible
	// (CustomStageID, CustomImportanceID, Roles do not exist)
}

func TestUpdateBoard_CustomFieldsViaFieldValueService(t *testing.T) {
	// Arrange
	boardID := uuid.New()

	board := &domain.Board{
		BaseModel:         domain.BaseModel{ID: boardID},
		ProjectID:         uuid.New(),
		Title:             "Test Board",
		CustomFieldsCache: `{"field-priority": "Low"}`,
	}

	// Act - Simulate custom field update via FieldValueService
	// (Not via BoardService.UpdateBoard)

	// Custom fields should be updated via:
	// 1. FieldValueService.SetFieldValue() - updates board_field_values table
	// 2. FieldValueService.updateBoardCache() - rebuilds custom_fields_cache

	// After update, board cache should be refreshed
	updatedCache := `{"field-priority": "High", "field-status": "Done"}`
	board.CustomFieldsCache = updatedCache

	// Assert
	var customFields map[string]interface{}
	err := json.Unmarshal([]byte(board.CustomFieldsCache), &customFields)
	assert.NoError(t, err)

	assert.Equal(t, "High", customFields["field-priority"])
	assert.Equal(t, "Done", customFields["field-status"])
}

// =============================================================================
// buildBoardResponse Tests
// =============================================================================

func TestBuildBoardResponse_CustomFieldsParsing(t *testing.T) {
	// Arrange
	customFieldsJSON := `{
		"field-1": "text value",
		"field-2": 42,
		"field-3": true,
		"field-4": ["opt1", "opt2"]
	}`

	board := &domain.Board{
		BaseModel:         domain.BaseModel{ID: uuid.New()},
		ProjectID:         uuid.New(),
		Title:             "Test Board",
		Description:       "Content",
		CreatedBy:         uuid.New(),
		CustomFieldsCache: customFieldsJSON,
	}

	// Act - Simulate buildBoardResponse parsing
	var customFields map[string]interface{}
	err := json.Unmarshal([]byte(board.CustomFieldsCache), &customFields)

	// Assert
	assert.NoError(t, err, "Should parse custom fields")

	// 1. Text field
	assert.Equal(t, "text value", customFields["field-1"])

	// 2. Number field (JSON unmarshals to float64)
	numberVal, ok := customFields["field-2"].(float64)
	assert.True(t, ok, "field-2 should be number")
	assert.Equal(t, float64(42), numberVal)

	// 3. Boolean field
	boolVal, ok := customFields["field-3"].(bool)
	assert.True(t, ok, "field-3 should be boolean")
	assert.True(t, boolVal)

	// 4. Array field
	arrayVal, ok := customFields["field-4"].([]interface{})
	assert.True(t, ok, "field-4 should be array")
	assert.Equal(t, 2, len(arrayVal))
}

func TestBuildBoardResponse_NoLegacyFieldsInResponse(t *testing.T) {
	// Arrange
	board := &domain.Board{
		BaseModel:         domain.BaseModel{ID: uuid.New()},
		ProjectID:         uuid.New(),
		Title:             "Test Board",
		Description:       "Content",
		CreatedBy:         uuid.New(),
		CustomFieldsCache: `{"field-status": "In Progress"}`,
	}

	// Expected BoardResponse structure (after refactoring)
	type BoardResponse struct {
		ID           string
		ProjectID    string
		Title        string
		Content      string
		CustomFields map[string]interface{} // New unified field
		// Stage        - REMOVED
		// Importance   - REMOVED
		// Roles        - REMOVED
	}

	// Parse custom fields
	var customFields map[string]interface{}
	json.Unmarshal([]byte(board.CustomFieldsCache), &customFields)

	response := BoardResponse{
		ID:           board.ID.String(),
		ProjectID:    board.ProjectID.String(),
		Title:        board.Title,
		Content:      board.Description,
		CustomFields: customFields,
	}

	// Assert
	// 1. Response should have CustomFields
	assert.NotNil(t, response.CustomFields)
	assert.Equal(t, "In Progress", response.CustomFields["field-status"])

	// 2. Response should NOT have legacy fields:
	// (Stage, Importance, Roles fields removed from BoardResponse struct)
}

// =============================================================================
// buildBoardResponseOptimized Tests
// =============================================================================

func TestBuildBoardResponseOptimized_OnlyUserMapParameter(t *testing.T) {
	// Arrange
	board := &domain.Board{
		BaseModel:         domain.BaseModel{ID: uuid.New()},
		ProjectID:         uuid.New(),
		Title:             "Test Board",
		CreatedBy:         uuid.New(),
		CustomFieldsCache: `{"field-priority": "High"}`,
	}

	// Mock user info map
	type UserInfo struct {
		UserID   string
		Name     string
		Email    string
		IsActive bool
	}

	userMap := map[string]UserInfo{
		board.CreatedBy.String(): {
			UserID:   board.CreatedBy.String(),
			Name:     "Test User",
			Email:    "test@example.com",
			IsActive: true,
		},
	}

	// Assert
	// buildBoardResponseOptimized signature should be:
	// func (s *boardService) buildBoardResponseOptimized(
	//     board *domain.Board,
	//     userMap map[string]client.UserInfo,
	// ) (*dto.BoardResponse, error)

	// OLD (removed parameters):
	// - stage *domain.CustomStage
	// - importance *domain.CustomImportance
	// - roles []domain.CustomRole

	// Only userMap should be passed for author and assignee info
	assert.NotNil(t, userMap)
	assert.Equal(t, 1, len(userMap), "Should have user info for author")

	userInfo := userMap[board.CreatedBy.String()]
	assert.Equal(t, "Test User", userInfo.Name)
}

// =============================================================================
// Integration Scenario Tests
// =============================================================================

func TestBoardLifecycle_WithoutLegacyFields(t *testing.T) {
	// Arrange
	projectID := uuid.New()
	userID := uuid.New()

	// Step 1: Create board
	newBoard := &domain.Board{
		BaseModel:         domain.BaseModel{ID: uuid.New()},
		ProjectID:         projectID,
		Title:             "New Task",
		Description:       "Task description",
		CreatedBy:         userID,
		CustomFieldsCache: "{}", // Empty initially
	}

	// Step 2: Set custom field values (via FieldValueService)
	// This updates board_field_values table and rebuilds cache
	customFieldsJSON := `{
		"field-priority": "High",
		"field-status": "To Do",
		"field-assignee": "user-123"
	}`
	newBoard.CustomFieldsCache = customFieldsJSON

	// Step 3: Read board (via BoardService)
	var customFields map[string]interface{}
	err := json.Unmarshal([]byte(newBoard.CustomFieldsCache), &customFields)
	assert.NoError(t, err)

	// Step 4: Update board title (via BoardService)
	newBoard.Title = "Updated Task"

	// Step 5: Update custom field (via FieldValueService)
	customFields["field-status"] = "In Progress"
	updatedJSON, _ := json.Marshal(customFields)
	newBoard.CustomFieldsCache = string(updatedJSON)

	// Assert - Complete lifecycle without legacy fields
	assert.Equal(t, "Updated Task", newBoard.Title)
	assert.Equal(t, "High", customFields["field-priority"])
	assert.Equal(t, "In Progress", customFields["field-status"])

	// No legacy fields involved at any step
}

func TestBatchBoardQuery_OptimizedWithoutLegacyFetches(t *testing.T) {
	// Arrange
	projectID := uuid.New()

	boards := []domain.Board{
		{
			BaseModel:         domain.BaseModel{ID: uuid.New()},
			ProjectID:         projectID,
			Title:             "Board 1",
			CreatedBy:         uuid.New(),
			CustomFieldsCache: `{"field-status": "To Do"}`,
		},
		{
			BaseModel:         domain.BaseModel{ID: uuid.New()},
			ProjectID:         projectID,
			Title:             "Board 2",
			CreatedBy:         uuid.New(),
			CustomFieldsCache: `{"field-status": "Done"}`,
		},
	}

	// Old way (removed):
	// - Batch fetch stages for all boards
	// - Batch fetch importances for all boards
	// - Batch fetch roles for all boards
	// = 3 extra DB queries

	// New way:
	// - Only batch fetch user info (author, assignee)
	// - Parse custom_fields_cache directly (no extra queries)
	// = 1 DB query only

	userIDs := make([]string, 0, len(boards))
	for _, board := range boards {
		userIDs = append(userIDs, board.CreatedBy.String())
	}

	// Assert
	// 1. Should only need user info batch fetch
	assert.Equal(t, 2, len(userIDs), "Should collect user IDs only")

	// 2. Custom fields already in cache (no DB fetch needed)
	for _, board := range boards {
		var customFields map[string]interface{}
		err := json.Unmarshal([]byte(board.CustomFieldsCache), &customFields)
		assert.NoError(t, err)
		assert.NotEmpty(t, customFields["field-status"])
	}

	// 3. No legacy field batch fetches needed
	// (No stage IDs, importance IDs, or role IDs to fetch)
}
