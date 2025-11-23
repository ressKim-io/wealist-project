package auth

import (
	"board-service/internal/apperrors"
	"board-service/internal/domain"
	"board-service/internal/repository"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectAuthorizer handles project-level authorization
type ProjectAuthorizer interface {
	// RequireMember checks if user is a member of the project
	RequireMember(userID, projectID uuid.UUID) (*domain.ProjectMember, error)

	// RequireAdmin checks if user has ADMIN or OWNER role in the project
	RequireAdmin(userID, projectID uuid.UUID) (*domain.ProjectMember, error)

	// RequireOwner checks if user has OWNER role in the project
	RequireOwner(userID, projectID uuid.UUID) (*domain.ProjectMember, error)

	// RequireMinLevel checks if user has minimum role level in the project
	RequireMinLevel(userID, projectID uuid.UUID, minLevel int) (*domain.ProjectMember, error)

	// GetRole retrieves user's role in the project (nil if not a member)
	GetRole(userID, projectID uuid.UUID) (*domain.Role, error)

	// CanEdit checks if user can edit a resource (is author OR has ADMIN+ role)
	CanEdit(userID, projectID, authorID uuid.UUID) (bool, error)

	// CanDelete checks if user can delete a resource (is author OR has ADMIN+ role)
	CanDelete(userID, projectID, authorID uuid.UUID) (bool, error)
}

type projectAuthorizer struct {
	projectRepo repository.ProjectRepository
	roleRepo    repository.RoleRepository
}

// NewProjectAuthorizer creates a new project authorizer
func NewProjectAuthorizer(
	projectRepo repository.ProjectRepository,
	roleRepo repository.RoleRepository,
) ProjectAuthorizer {
	return &projectAuthorizer{
		projectRepo: projectRepo,
		roleRepo:    roleRepo,
	}
}

// RequireMember checks if user is a member of the project
func (a *projectAuthorizer) RequireMember(userID, projectID uuid.UUID) (*domain.ProjectMember, error) {
	member, err := a.projectRepo.FindMemberByUserAndProject(userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				apperrors.ErrCodeForbidden,
				"프로젝트 멤버가 아닙니다",
				403,
			)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// Load role if not already loaded
	if member.Role == nil {
		role, err := a.roleRepo.FindByID(member.RoleID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
		}
		member.Role = role
	}

	return member, nil
}

// RequireAdmin checks if user has ADMIN or OWNER role (level >= 50)
func (a *projectAuthorizer) RequireAdmin(userID, projectID uuid.UUID) (*domain.ProjectMember, error) {
	member, err := a.RequireMember(userID, projectID)
	if err != nil {
		return nil, err
	}

	if member.Role.Level < 50 {
		return nil, apperrors.New(
			apperrors.ErrCodeForbidden,
			"ADMIN 이상의 권한이 필요합니다",
			403,
		)
	}

	return member, nil
}

// RequireOwner checks if user has OWNER role (level >= 100)
func (a *projectAuthorizer) RequireOwner(userID, projectID uuid.UUID) (*domain.ProjectMember, error) {
	member, err := a.RequireMember(userID, projectID)
	if err != nil {
		return nil, err
	}

	if member.Role.Name != "OWNER" {
		return nil, apperrors.New(
			apperrors.ErrCodeForbidden,
			"OWNER 권한이 필요합니다",
			403,
		)
	}

	return member, nil
}

// RequireMinLevel checks if user has minimum role level
func (a *projectAuthorizer) RequireMinLevel(userID, projectID uuid.UUID, minLevel int) (*domain.ProjectMember, error) {
	member, err := a.RequireMember(userID, projectID)
	if err != nil {
		return nil, err
	}

	if member.Role.Level < minLevel {
		return nil, apperrors.New(
			apperrors.ErrCodeForbidden,
			"권한이 부족합니다",
			403,
		)
	}

	return member, nil
}

// GetRole retrieves user's role in the project
func (a *projectAuthorizer) GetRole(userID, projectID uuid.UUID) (*domain.Role, error) {
	member, err := a.projectRepo.FindMemberByUserAndProject(userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not a member
		}
		return nil, err
	}

	role, err := a.roleRepo.FindByID(member.RoleID)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// CanEdit checks if user can edit a resource
// User can edit if they are the author OR have ADMIN+ role
func (a *projectAuthorizer) CanEdit(userID, projectID, authorID uuid.UUID) (bool, error) {
	// Author can always edit their own content
	if userID == authorID {
		return true, nil
	}

	// Check if user has ADMIN+ role
	member, err := a.RequireMember(userID, projectID)
	if err != nil {
		return false, err
	}

	// ADMIN+ (level >= 50) can edit any content
	return member.Role.Level >= 50, nil
}

// CanDelete checks if user can delete a resource
// User can delete if they are the author OR have ADMIN+ role
func (a *projectAuthorizer) CanDelete(userID, projectID, authorID uuid.UUID) (bool, error) {
	// Author can always delete their own content
	if userID == authorID {
		return true, nil
	}

	// Check if user has ADMIN+ role
	member, err := a.RequireMember(userID, projectID)
	if err != nil {
		return false, err
	}

	// ADMIN+ (level >= 50) can delete any content
	return member.Role.Level >= 50, nil
}

// ==================== Role Level Constants ====================

const (
	// RoleLevelMember is the level for regular members
	RoleLevelMember = 10

	// RoleLevelAdmin is the level for administrators
	RoleLevelAdmin = 50

	// RoleLevelOwner is the level for project owners
	RoleLevelOwner = 100
)

// ==================== Helper Functions ====================

// IsAdmin checks if a role is ADMIN or higher
func IsAdmin(role *domain.Role) bool {
	return role != nil && role.Level >= RoleLevelAdmin
}

// IsOwner checks if a role is OWNER
func IsOwner(role *domain.Role) bool {
	return role != nil && role.Name == "OWNER"
}

// IsMember checks if a role is at least MEMBER
func IsMember(role *domain.Role) bool {
	return role != nil && role.Level >= RoleLevelMember
}
