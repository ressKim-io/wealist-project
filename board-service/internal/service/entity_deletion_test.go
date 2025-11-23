package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"project-board-api/internal/domain"
)

// MockS3ClientForDeletion is a mock implementation of S3Client for deletion tests
type MockS3ClientForDeletion struct {
	DeleteFileFunc func(ctx context.Context, key string) error
	DeletedKeys    []string
}

func (m *MockS3ClientForDeletion) DeleteFile(ctx context.Context, key string) error {
	m.DeletedKeys = append(m.DeletedKeys, key)
	if m.DeleteFileFunc != nil {
		return m.DeleteFileFunc(ctx, key)
	}
	return nil
}

// TestBoardDeletion_WithAttachments tests that board deletion also deletes associated attachments
func TestBoardDeletion_WithAttachments(t *testing.T) {
	// Given
	boardID := uuid.New()
	projectID := uuid.New()
	
	board := &domain.Board{
		BaseModel: domain.BaseModel{ID: boardID},
		ProjectID: projectID,
		Title:     "Test Board",
		Content:   "Test Content",
	}
	
	attachments := []*domain.Attachment{
		{
			BaseModel:   domain.BaseModel{ID: uuid.New()},
			EntityType:  domain.EntityTypeBoard,
			EntityID:    &boardID,
			Status:      domain.AttachmentStatusConfirmed,
			FileName:    "file1.jpg",
			FileURL:     "https://bucket.s3.region.amazonaws.com/board/boards/workspace/2024/01/file1.jpg",
			FileSize:    1024,
			ContentType: "image/jpeg",
			UploadedBy:  uuid.New(),
		},
		{
			BaseModel:   domain.BaseModel{ID: uuid.New()},
			EntityType:  domain.EntityTypeBoard,
			EntityID:    &boardID,
			Status:      domain.AttachmentStatusConfirmed,
			FileName:    "file2.pdf",
			FileURL:     "https://bucket.s3.region.amazonaws.com/board/boards/workspace/2024/01/file2.pdf",
			FileSize:    2048,
			ContentType: "application/pdf",
			UploadedBy:  uuid.New(),
		},
	}
	
	mockBoardRepo := &MockBoardRepository{
		FindByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
			if id == boardID {
				return board, nil
			}
			return nil, nil
		},
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	
	deletedAttachmentIDs := []uuid.UUID{}
	mockAttachmentRepo := &MockAttachmentRepository{
		FindByEntityIDFunc: func(ctx context.Context, entityType domain.EntityType, entityID uuid.UUID) ([]*domain.Attachment, error) {
			if entityType == domain.EntityTypeBoard && entityID == boardID {
				return attachments, nil
			}
			return nil, nil
		},
		DeleteBatchFunc: func(ctx context.Context, ids []uuid.UUID) error {
			deletedAttachmentIDs = ids
			return nil
		},
	}
	
	mockS3Client := &MockS3ClientForDeletion{}
	logger, _ := zap.NewDevelopment()
	
	service := NewBoardService(
		mockBoardRepo,
		&MockProjectRepository{},
		&MockFieldOptionRepository{},
		&MockParticipantRepository{},
		mockAttachmentRepo,
		mockS3Client,
		&MockFieldOptionConverter{},
		nil,
		logger,
	)
	
	// When
	err := service.DeleteBoard(context.Background(), boardID)
	
	// Then
	assert.NoError(t, err)
	
	// Verify S3 files were deleted
	assert.Len(t, mockS3Client.DeletedKeys, 2)
	assert.Contains(t, mockS3Client.DeletedKeys, "board/boards/workspace/2024/01/file1.jpg")
	assert.Contains(t, mockS3Client.DeletedKeys, "board/boards/workspace/2024/01/file2.pdf")
	
	// Verify attachments were deleted from database
	assert.Len(t, deletedAttachmentIDs, 2)
}

// TestBoardDeletion_WithoutAttachments tests that board deletion works when there are no attachments
func TestBoardDeletion_WithoutAttachments(t *testing.T) {
	// Given
	boardID := uuid.New()
	projectID := uuid.New()
	
	board := &domain.Board{
		BaseModel: domain.BaseModel{ID: boardID},
		ProjectID: projectID,
		Title:     "Test Board",
		Content:   "Test Content",
	}
	
	mockBoardRepo := &MockBoardRepository{
		FindByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
			if id == boardID {
				return board, nil
			}
			return nil, nil
		},
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	
	mockAttachmentRepo := &MockAttachmentRepository{
		FindByEntityIDFunc: func(ctx context.Context, entityType domain.EntityType, entityID uuid.UUID) ([]*domain.Attachment, error) {
			return []*domain.Attachment{}, nil
		},
	}
	
	mockS3Client := &MockS3ClientForDeletion{}
	logger, _ := zap.NewDevelopment()
	
	service := NewBoardService(
		mockBoardRepo,
		&MockProjectRepository{},
		&MockFieldOptionRepository{},
		&MockParticipantRepository{},
		mockAttachmentRepo,
		mockS3Client,
		&MockFieldOptionConverter{},
		nil,
		logger,
	)
	
	// When
	err := service.DeleteBoard(context.Background(), boardID)
	
	// Then
	assert.NoError(t, err)
	
	// Verify S3 client was not called
	assert.Len(t, mockS3Client.DeletedKeys, 0)
}

// TestCommentDeletion_WithAttachments tests that comment deletion also deletes associated attachments
func TestCommentDeletion_WithAttachments(t *testing.T) {
	// Given
	commentID := uuid.New()
	boardID := uuid.New()
	
	comment := &domain.Comment{
		BaseModel: domain.BaseModel{ID: commentID},
		BoardID:   boardID,
		UserID:    uuid.New(),
		Content:   "Test Comment",
	}
	
	attachments := []*domain.Attachment{
		{
			BaseModel:   domain.BaseModel{ID: uuid.New()},
			EntityType:  domain.EntityTypeComment,
			EntityID:    &commentID,
			Status:      domain.AttachmentStatusConfirmed,
			FileName:    "comment_file.jpg",
			FileURL:     "https://bucket.s3.region.amazonaws.com/board/comments/workspace/2024/01/comment_file.jpg",
			FileSize:    512,
			ContentType: "image/jpeg",
			UploadedBy:  uuid.New(),
		},
	}
	
	mockCommentRepo := &MockCommentRepository{
		FindByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
			if id == commentID {
				return comment, nil
			}
			return nil, nil
		},
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	
	deletedAttachmentIDs := []uuid.UUID{}
	mockAttachmentRepo := &MockAttachmentRepository{
		FindByEntityIDFunc: func(ctx context.Context, entityType domain.EntityType, entityID uuid.UUID) ([]*domain.Attachment, error) {
			if entityType == domain.EntityTypeComment && entityID == commentID {
				return attachments, nil
			}
			return nil, nil
		},
		DeleteBatchFunc: func(ctx context.Context, ids []uuid.UUID) error {
			deletedAttachmentIDs = ids
			return nil
		},
	}
	
	mockS3Client := &MockS3ClientForDeletion{}
	logger, _ := zap.NewDevelopment()
	
	service := NewCommentService(
		mockCommentRepo,
		&MockBoardRepository{},
		mockAttachmentRepo,
		mockS3Client,
		logger,
	)
	
	// When
	err := service.DeleteComment(context.Background(), commentID)
	
	// Then
	assert.NoError(t, err)
	
	// Verify S3 file was deleted
	assert.Len(t, mockS3Client.DeletedKeys, 1)
	assert.Contains(t, mockS3Client.DeletedKeys, "board/comments/workspace/2024/01/comment_file.jpg")
	
	// Verify attachments were deleted from database
	assert.Len(t, deletedAttachmentIDs, 1)
}

// TestProjectDeletion_WithAttachments tests that project deletion also deletes associated attachments
func TestProjectDeletion_WithAttachments(t *testing.T) {
	// Given
	projectID := uuid.New()
	userID := uuid.New()
	workspaceID := uuid.New()
	
	project := &domain.Project{
		BaseModel:   domain.BaseModel{ID: projectID},
		WorkspaceID: workspaceID,
		OwnerID:     userID,
		Name:        "Test Project",
		Description: "Test Description",
	}
	
	member := &domain.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		RoleName:  domain.ProjectRoleOwner,
	}
	
	attachments := []*domain.Attachment{
		{
			BaseModel:   domain.BaseModel{ID: uuid.New()},
			EntityType:  domain.EntityTypeProject,
			EntityID:    &projectID,
			Status:      domain.AttachmentStatusConfirmed,
			FileName:    "project_doc.pdf",
			FileURL:     "https://bucket.s3.region.amazonaws.com/board/projects/workspace/2024/01/project_doc.pdf",
			FileSize:    4096,
			ContentType: "application/pdf",
			UploadedBy:  userID,
		},
	}
	
	mockProjectRepo := &MockProjectRepository{
		FindByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
			if id == projectID {
				return project, nil
			}
			return nil, nil
		},
		FindMemberByProjectAndUserFunc: func(ctx context.Context, projID, uid uuid.UUID) (*domain.ProjectMember, error) {
			if projID == projectID && uid == userID {
				return member, nil
			}
			return nil, nil
		},
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	
	deletedAttachmentIDs := []uuid.UUID{}
	mockAttachmentRepo := &MockAttachmentRepository{
		FindByEntityIDFunc: func(ctx context.Context, entityType domain.EntityType, entityID uuid.UUID) ([]*domain.Attachment, error) {
			if entityType == domain.EntityTypeProject && entityID == projectID {
				return attachments, nil
			}
			return nil, nil
		},
		DeleteBatchFunc: func(ctx context.Context, ids []uuid.UUID) error {
			deletedAttachmentIDs = ids
			return nil
		},
	}
	
	mockS3Client := &MockS3ClientForDeletion{}
	logger, _ := zap.NewDevelopment()
	
	service := NewProjectService(
		mockProjectRepo,
		&MockFieldOptionRepository{},
		mockAttachmentRepo,
		mockS3Client,
		&MockUserClient{},
		nil,
		logger,
	)
	
	// When
	err := service.DeleteProject(context.Background(), projectID, userID)
	
	// Then
	assert.NoError(t, err)
	
	// Verify S3 file was deleted
	assert.Len(t, mockS3Client.DeletedKeys, 1)
	assert.Contains(t, mockS3Client.DeletedKeys, "board/projects/workspace/2024/01/project_doc.pdf")
	
	// Verify attachments were deleted from database
	assert.Len(t, deletedAttachmentIDs, 1)
}



// TestExtractS3KeyFromURL tests the S3 key extraction helper function
func TestExtractS3KeyFromURL(t *testing.T) {
	tests := []struct {
		name     string
		fileURL  string
		expected string
	}{
		{
			name:     "AWS S3 URL",
			fileURL:  "https://bucket.s3.region.amazonaws.com/board/boards/workspace/2024/01/file.jpg",
			expected: "board/boards/workspace/2024/01/file.jpg",
		},
		{
			name:     "MinIO URL",
			fileURL:  "http://localhost:9000/bucket/board/comments/workspace/2024/01/file.pdf",
			expected: "board/comments/workspace/2024/01/file.pdf",
		},
		{
			name:     "Invalid URL",
			fileURL:  "invalid-url",
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractS3KeyFromURL(tt.fileURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}
