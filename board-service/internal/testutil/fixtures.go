package testutil

import (
	"board-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

// ==================== Project Fixtures ====================

func NewTestProject() *domain.Project {
	return &domain.Project{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		WorkspaceID: uuid.New(),
		Name:        "Test Project",
		Description: "Test Project Description",
		OwnerID:     uuid.New(),
	}
}

func NewTestProjectWithID(id, ownerID uuid.UUID) *domain.Project {
	return &domain.Project{
		BaseModel: domain.BaseModel{
			ID:        id,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		WorkspaceID: uuid.New(),
		Name:        "Test Project",
		Description: "Test Project Description",
		OwnerID:     ownerID,
	}
}

// ==================== Role Fixtures ====================

func NewOwnerRole() *domain.Role {
	return &domain.Role{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "OWNER",
		Description: "Owner",
		Level:       100,
	}
}

func NewAdminRole() *domain.Role {
	return &domain.Role{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "ADMIN",
		Description: "Admin",
		Level:       50,
	}
}

func NewMemberRole() *domain.Role {
	return &domain.Role{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "MEMBER",
		Description: "Member",
		Level:       10,
	}
}

// ==================== Project Member Fixtures ====================

func NewTestProjectMember(projectID, userID, roleID uuid.UUID) *domain.ProjectMember {
	return &domain.ProjectMember{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ProjectID: projectID,
		UserID:    userID,
		RoleID:    roleID,
		JoinedAt:  time.Now(),
	}
}

func NewTestProjectOwner(projectID, userID uuid.UUID, role *domain.Role) *domain.ProjectMember {
	member := NewTestProjectMember(projectID, userID, role.ID)
	member.Role = role
	return member
}

// ==================== Board Fixtures ====================

func NewTestBoard(projectID, userID uuid.UUID) *domain.Board {
	return &domain.Board{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ProjectID:         projectID,
		Title:             "Test Board",
		Description:       "Test Description",
		CreatedBy:         userID,
		CustomFieldsCache: "{}",
	}
}

func NewTestBoardWithFields(projectID, userID uuid.UUID, customFields string) *domain.Board {
	board := NewTestBoard(projectID, userID)
	board.CustomFieldsCache = customFields
	return board
}

func NewTestBoardWithAssignee(projectID, userID, assigneeID uuid.UUID) *domain.Board {
	board := NewTestBoard(projectID, userID)
	board.AssigneeID = &assigneeID
	return board
}

// ==================== Field Fixtures ====================

func NewTestField(projectID uuid.UUID, fieldType domain.FieldType) *domain.ProjectField {
	return &domain.ProjectField{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ProjectID:    projectID,
		Name:         "Test Field",
		FieldType:    fieldType,
		Description:  "Test field description",
		DisplayOrder: 0,
		IsRequired:   false,
		Config:       "{}",
	}
}

func NewTestSingleSelectField(projectID uuid.UUID, name string) *domain.ProjectField {
	return &domain.ProjectField{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ProjectID:       projectID,
		Name:            name,
		FieldType:       domain.FieldTypeSingleSelect,
		Description:     "Test single select field",
		DisplayOrder:    0,
		IsRequired:      false,
		IsSystemDefault: false,
		Config:          "{}",
	}
}

func NewTestSystemDefaultField(projectID uuid.UUID, name string, fieldType domain.FieldType) *domain.ProjectField {
	field := NewTestField(projectID, fieldType)
	field.Name = name
	field.IsSystemDefault = true
	field.IsRequired = true
	return field
}

// ==================== Field Option Fixtures ====================

func NewTestFieldOption(fieldID uuid.UUID, label, color string, order int) *domain.FieldOption {
	return &domain.FieldOption{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		FieldID:      fieldID,
		Label:        label,
		Color:        color,
		DisplayOrder: order,
	}
}

func NewTestStageOptions(fieldID uuid.UUID) []domain.FieldOption {
	return []domain.FieldOption{
		*NewTestFieldOption(fieldID, "대기", "#F59E0B", 0),
		*NewTestFieldOption(fieldID, "진행중", "#3B82F6", 1),
		*NewTestFieldOption(fieldID, "완료", "#10B981", 2),
	}
}

// ==================== Field Value Fixtures ====================

func NewTestFieldValue(boardID, fieldID, optionID uuid.UUID) *domain.BoardFieldValue {
	return &domain.BoardFieldValue{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		BoardID:       boardID,
		FieldID:       fieldID,
		ValueOptionID: &optionID,
		DisplayOrder:  0,
	}
}

// ==================== View Fixtures ====================

func NewTestView(projectID, userID uuid.UUID, isDefault bool) *domain.SavedView {
	return &domain.SavedView{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ProjectID:   projectID,
		CreatedBy:   userID,
		Name:        "Test View",
		Description: "Test view description",
		IsDefault:   isDefault,
		Filters:     "{}",
	}
}

// ==================== Join Request Fixtures ====================

func NewTestJoinRequest(projectID, userID uuid.UUID, status domain.ProjectJoinRequestStatus) *domain.ProjectJoinRequest {
	return &domain.ProjectJoinRequest{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ProjectID:   projectID,
		UserID:      userID,
		Status:      status,
		RequestedAt: time.Now(),
	}
}

// ==================== Helper Functions ====================

// NewUUIDs generates n UUIDs for testing
func NewUUIDs(n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := 0; i < n; i++ {
		ids[i] = uuid.New()
	}
	return ids
}

// NewTestTime returns a fixed time for testing
func NewTestTime() time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
}

// NewTestTimePtr returns a pointer to a fixed time
func NewTestTimePtr() *time.Time {
	t := NewTestTime()
	return &t
}
