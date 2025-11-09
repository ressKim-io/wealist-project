package repository

import (
	"board-service/internal/domain"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	// Project CRUD
	Create(project *domain.Project) error
	FindByID(id uuid.UUID) (*domain.Project, error)
	FindByWorkspaceID(workspaceID uuid.UUID) ([]domain.Project, error)
	Update(project *domain.Project) error
	Delete(id uuid.UUID) error
	Search(workspaceID uuid.UUID, query string, page, limit int) ([]domain.Project, int64, error)

	// Join Request
	CreateJoinRequest(req *domain.ProjectJoinRequest) error
	FindJoinRequestByID(id uuid.UUID) (*domain.ProjectJoinRequest, error)
	FindJoinRequestsByProject(projectID uuid.UUID, status string) ([]domain.ProjectJoinRequest, error)
	FindJoinRequestByUserAndProject(userID, projectID uuid.UUID) (*domain.ProjectJoinRequest, error)
	UpdateJoinRequest(req *domain.ProjectJoinRequest) error

	// Member
	CreateMember(member *domain.ProjectMember) error
	FindMemberByID(id uuid.UUID) (*domain.ProjectMember, error)
	FindMembersByProject(projectID uuid.UUID) ([]domain.ProjectMember, error)
	FindMemberByUserAndProject(userID, projectID uuid.UUID) (*domain.ProjectMember, error)
	UpdateMember(member *domain.ProjectMember) error
	DeleteMember(id uuid.UUID) error
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

// Project CRUD

func (r *projectRepository) Create(project *domain.Project) error {
	return r.db.Create(project).Error
}

func (r *projectRepository) FindByID(id uuid.UUID) (*domain.Project, error) {
	var project domain.Project
	if err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) FindByWorkspaceID(workspaceID uuid.UUID) ([]domain.Project, error) {
	var projects []domain.Project
	if err := r.db.Where("workspace_id = ? AND is_deleted = ?", workspaceID, false).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) Update(project *domain.Project) error {
	return r.db.Save(project).Error
}

func (r *projectRepository) Delete(id uuid.UUID) error {
	// Soft delete
	return r.db.Model(&domain.Project{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}

func (r *projectRepository) Search(workspaceID uuid.UUID, query string, page, limit int) ([]domain.Project, int64, error) {
	var projects []domain.Project
	var total int64

	// Default pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Build query
	db := r.db.Model(&domain.Project{}).
		Where("workspace_id = ? AND is_deleted = ?", workspaceID, false)

	if query != "" {
		searchPattern := fmt.Sprintf("%%%s%%", query)
		db = db.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := db.Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// Join Request

func (r *projectRepository) CreateJoinRequest(req *domain.ProjectJoinRequest) error {
	return r.db.Create(req).Error
}

func (r *projectRepository) FindJoinRequestByID(id uuid.UUID) (*domain.ProjectJoinRequest, error) {
	var req domain.ProjectJoinRequest
	if err := r.db.Where("id = ?", id).First(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *projectRepository) FindJoinRequestsByProject(projectID uuid.UUID, status string) ([]domain.ProjectJoinRequest, error) {
	var requests []domain.ProjectJoinRequest

	query := r.db.Where("project_id = ?", projectID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("requested_at DESC").Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *projectRepository) FindJoinRequestByUserAndProject(userID, projectID uuid.UUID) (*domain.ProjectJoinRequest, error) {
	var req domain.ProjectJoinRequest
	if err := r.db.Where("user_id = ? AND project_id = ?", userID, projectID).
		First(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *projectRepository) UpdateJoinRequest(req *domain.ProjectJoinRequest) error {
	return r.db.Save(req).Error
}

// Member

func (r *projectRepository) CreateMember(member *domain.ProjectMember) error {
	return r.db.Create(member).Error
}

func (r *projectRepository) FindMemberByID(id uuid.UUID) (*domain.ProjectMember, error) {
	var member domain.ProjectMember
	if err := r.db.Where("id = ?", id).First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *projectRepository) FindMembersByProject(projectID uuid.UUID) ([]domain.ProjectMember, error) {
	var members []domain.ProjectMember
	if err := r.db.Where("project_id = ?", projectID).
		Order("joined_at ASC").
		Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (r *projectRepository) FindMemberByUserAndProject(userID, projectID uuid.UUID) (*domain.ProjectMember, error) {
	var member domain.ProjectMember
	if err := r.db.Preload("Role").
		Where("user_id = ? AND project_id = ?", userID, projectID).
		First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *projectRepository) UpdateMember(member *domain.ProjectMember) error {
	return r.db.Save(member).Error
}

func (r *projectRepository) DeleteMember(id uuid.UUID) error {
	return r.db.Delete(&domain.ProjectMember{}, "id = ?", id).Error
}
