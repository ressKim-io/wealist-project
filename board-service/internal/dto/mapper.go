package dto

import (
	"board-service/internal/cache"
	"board-service/internal/client"
	"board-service/internal/domain"
	"encoding/json"

	"go.uber.org/zap"
)

// ==================== DTO Mappers ====================
// DTO 변환 로직을 중앙화하여 Service의 중복을 제거합니다.
//
// Best Practices (Go 2024):
// - 명시적 변환 (reflection 기반 automapper 대신)
// - 타입 안전성 보장
// - 테스트 가능한 순수 함수
// - Service 레이어의 복잡도 감소

// ==================== Board Mapper ====================

type BoardMapper struct {
	logger *zap.Logger
}

func NewBoardMapper(logger *zap.Logger) *BoardMapper {
	return &BoardMapper{logger: logger}
}

// ToResponse converts a domain.Board to BoardResponse (single board, fetches users)
// Use ToResponseWithUserMap for batch operations
func (m *BoardMapper) ToResponse(board *domain.Board) *BoardResponse {
	return m.ToResponseWithUserMap(board, make(map[string]client.UserInfo))
}

// ToResponseWithUserMap converts a domain.Board to BoardResponse using pre-fetched user map
func (m *BoardMapper) ToResponseWithUserMap(
	board *domain.Board,
	userMap map[string]client.UserInfo,
) *BoardResponse {
	response := &BoardResponse{
		ID:        board.ID.String(),
		ProjectID: board.ProjectID.String(),
		Title:     board.Title,
		Content:   board.Description,
		DueDate:   board.DueDate,
		CreatedAt: board.CreatedAt,
		UpdatedAt: board.UpdatedAt,
	}

	// Parse CustomFieldsCache (JSONB)
	if board.CustomFieldsCache != "" && board.CustomFieldsCache != "{}" {
		var customFields map[string]interface{}
		if err := json.Unmarshal([]byte(board.CustomFieldsCache), &customFields); err != nil {
			if m.logger != nil {
				m.logger.Warn("Failed to parse custom_fields_cache",
					zap.Error(err),
					zap.String("board_id", board.ID.String()),
				)
			}
			response.CustomFields = make(map[string]interface{})
		} else {
			response.CustomFields = customFields
		}
	} else {
		response.CustomFields = make(map[string]interface{})
	}

	// Set author info (from userMap)
	if author, ok := userMap[board.CreatedBy.String()]; ok {
		response.Author = UserInfo{
			UserID:   author.UserID,
			Name:     author.Name,
			Email:    author.Email,
			IsActive: author.IsActive,
		}
	} else {
		// Fallback if user not found
		response.Author = UserInfo{
			UserID:   board.CreatedBy.String(),
			Name:     "Unknown User",
			Email:    "",
			IsActive: false,
		}
	}

	// Set assignee info (from userMap)
	if board.AssigneeID != nil {
		if assignee, ok := userMap[board.AssigneeID.String()]; ok {
			response.Assignee = &UserInfo{
				UserID:   assignee.UserID,
				Name:     assignee.Name,
				Email:    assignee.Email,
				IsActive: assignee.IsActive,
			}
		} else {
			// Fallback if user not found
			response.Assignee = &UserInfo{
				UserID:   board.AssigneeID.String(),
				Name:     "Unknown User",
				Email:    "",
				IsActive: false,
			}
		}
	}

	return response
}

// ToResponseList converts multiple domain.Board to BoardResponse list
func (m *BoardMapper) ToResponseList(
	boards []domain.Board,
	userMap map[string]client.UserInfo,
) []BoardResponse {
	responses := make([]BoardResponse, len(boards))
	for i, board := range boards {
		responses[i] = *m.ToResponseWithUserMap(&board, userMap)
	}
	return responses
}

// ==================== Project Mapper ====================

type ProjectMapper struct{}

func NewProjectMapper() *ProjectMapper {
	return &ProjectMapper{}
}

// ToResponse converts a domain.Project to ProjectResponse
func (m *ProjectMapper) ToResponse(project *domain.Project) *ProjectResponse {
	return &ProjectResponse{
		ID:          project.ID.String(),
		WorkspaceID: project.WorkspaceID.String(),
		Name:        project.Name,
		Description: project.Description,
		OwnerID:     project.OwnerID.String(),
		IsPublic:    project.IsPublic,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}

// ToResponseWithUser converts a domain.Project to ProjectResponse with owner info
func (m *ProjectMapper) ToResponseWithUser(
	project *domain.Project,
	owner *cache.UserInfo,
) *ProjectResponse {
	response := m.ToResponse(project)
	if owner != nil {
		response.OwnerName = owner.Name
		response.OwnerEmail = owner.Email
	}
	return response
}

// ToResponseList converts multiple domain.Project to ProjectResponse list
func (m *ProjectMapper) ToResponseList(
	projects []domain.Project,
	userMap map[string]cache.UserInfo,
) []ProjectResponse {
	responses := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		if owner, ok := userMap[project.OwnerID.String()]; ok {
			responses[i] = *m.ToResponseWithUser(&project, &owner)
		} else {
			responses[i] = *m.ToResponse(&project)
		}
	}
	return responses
}

// ==================== Comment Mapper ====================

type CommentMapper struct{}

func NewCommentMapper() *CommentMapper {
	return &CommentMapper{}
}

// ToResponse converts a domain.Comment to CommentResponse
func (m *CommentMapper) ToResponse(comment *domain.Comment) *CommentResponse {
	return &CommentResponse{
		ID:        comment.ID,
		UserID:    comment.UserID,
		UserName:  "Unknown User",
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}
}

// ToResponseWithUser converts a domain.Comment to CommentResponse with user info
func (m *CommentMapper) ToResponseWithUser(
	comment *domain.Comment,
	user *cache.SimpleUser,
) *CommentResponse {
	response := m.ToResponse(comment)
	if user != nil {
		response.UserName = user.Name
		response.UserAvatar = user.AvatarURL
	}
	return response
}

// ToResponseList converts multiple domain.Comment to CommentResponse list
func (m *CommentMapper) ToResponseList(
	comments []domain.Comment,
	userMap map[string]cache.SimpleUser,
) []CommentResponse {
	responses := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		if user, ok := userMap[comment.UserID.String()]; ok {
			responses[i] = *m.ToResponseWithUser(&comment, &user)
		} else {
			responses[i] = *m.ToResponse(&comment)
		}
	}
	return responses
}

// ==================== Project Member Mapper ====================

type ProjectMemberMapper struct{}

func NewProjectMemberMapper() *ProjectMemberMapper {
	return &ProjectMemberMapper{}
}

// ToResponse converts a domain.ProjectMember to ProjectMemberResponse
func (m *ProjectMemberMapper) ToResponse(member *domain.ProjectMember) *ProjectMemberResponse {
	return &ProjectMemberResponse{
		UserID:    member.UserID.String(),
		ProjectID: member.ProjectID.String(),
		RoleName:  member.Role.Name,
		JoinedAt:  member.JoinedAt,
	}
}

// ToResponseWithUser converts a domain.ProjectMember to ProjectMemberResponse with user info
func (m *ProjectMemberMapper) ToResponseWithUser(
	member *domain.ProjectMember,
	user *cache.UserInfo,
) *ProjectMemberResponse {
	response := m.ToResponse(member)
	if user != nil {
		response.UserName = user.Name
		response.UserEmail = user.Email
	}
	return response
}

// ToResponseList converts multiple domain.ProjectMember to ProjectMemberResponse list
func (m *ProjectMemberMapper) ToResponseList(
	members []domain.ProjectMember,
	userMap map[string]cache.UserInfo,
) []ProjectMemberResponse {
	responses := make([]ProjectMemberResponse, len(members))
	for i, member := range members {
		if user, ok := userMap[member.UserID.String()]; ok {
			responses[i] = *m.ToResponseWithUser(&member, &user)
		} else {
			responses[i] = *m.ToResponse(&member)
		}
	}
	return responses
}
