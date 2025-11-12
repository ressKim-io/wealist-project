package testutil

import (
	"board-service/internal/domain"
	"board-service/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ==================== Mock BoardRepository ====================

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

func (m *MockBoardRepository) FindByProject(projectID uuid.UUID, filters repository.BoardFilters, page, limit int) ([]domain.Board, int64, error) {
	args := m.Called(projectID, filters, page, limit)
	return args.Get(0).([]domain.Board), args.Get(1).(int64), args.Error(2)
}

func (m *MockBoardRepository) Update(board *domain.Board) error {
	args := m.Called(board)
	return args.Error(0)
}

func (m *MockBoardRepository) Delete(boardID uuid.UUID) error {
	args := m.Called(boardID)
	return args.Error(0)
}

// ==================== Mock ProjectRepository ====================

type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(project *domain.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockProjectRepository) FindByID(projectID uuid.UUID) (*domain.Project, error) {
	args := m.Called(projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *MockProjectRepository) FindByWorkspaceID(workspaceID uuid.UUID) ([]domain.Project, error) {
	args := m.Called(workspaceID)
	return args.Get(0).([]domain.Project), args.Error(1)
}

func (m *MockProjectRepository) Update(project *domain.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockProjectRepository) Delete(projectID uuid.UUID) error {
	args := m.Called(projectID)
	return args.Error(0)
}

func (m *MockProjectRepository) Search(workspaceID uuid.UUID, query string, page, limit int) ([]domain.Project, int64, error) {
	args := m.Called(workspaceID, query, page, limit)
	return args.Get(0).([]domain.Project), args.Get(1).(int64), args.Error(2)
}

// Project Member methods
func (m *MockProjectRepository) CreateMember(member *domain.ProjectMember) error {
	args := m.Called(member)
	return args.Error(0)
}

func (m *MockProjectRepository) FindMemberByID(memberID uuid.UUID) (*domain.ProjectMember, error) {
	args := m.Called(memberID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProjectMember), args.Error(1)
}

func (m *MockProjectRepository) FindMemberByUserAndProject(userID, projectID uuid.UUID) (*domain.ProjectMember, error) {
	args := m.Called(userID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProjectMember), args.Error(1)
}

func (m *MockProjectRepository) FindMembersByProject(projectID uuid.UUID) ([]domain.ProjectMember, error) {
	args := m.Called(projectID)
	return args.Get(0).([]domain.ProjectMember), args.Error(1)
}

func (m *MockProjectRepository) UpdateMember(member *domain.ProjectMember) error {
	args := m.Called(member)
	return args.Error(0)
}

func (m *MockProjectRepository) DeleteMember(memberID uuid.UUID) error {
	args := m.Called(memberID)
	return args.Error(0)
}

// Join Request methods
func (m *MockProjectRepository) CreateJoinRequest(req *domain.ProjectJoinRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockProjectRepository) FindJoinRequestByID(reqID uuid.UUID) (*domain.ProjectJoinRequest, error) {
	args := m.Called(reqID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProjectJoinRequest), args.Error(1)
}

func (m *MockProjectRepository) FindJoinRequestByUserAndProject(userID, projectID uuid.UUID) (*domain.ProjectJoinRequest, error) {
	args := m.Called(userID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProjectJoinRequest), args.Error(1)
}

func (m *MockProjectRepository) FindJoinRequestsByProject(projectID uuid.UUID, status string) ([]domain.ProjectJoinRequest, error) {
	args := m.Called(projectID, status)
	return args.Get(0).([]domain.ProjectJoinRequest), args.Error(1)
}

func (m *MockProjectRepository) UpdateJoinRequest(req *domain.ProjectJoinRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

// ==================== Mock RoleRepository ====================

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) FindByID(roleID uuid.UUID) (*domain.Role, error) {
	args := m.Called(roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *MockRoleRepository) FindByName(name string) (*domain.Role, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *MockRoleRepository) FindAll() ([]domain.Role, error) {
	args := m.Called()
	return args.Get(0).([]domain.Role), args.Error(1)
}

// ==================== Mock FieldRepository ====================

type MockFieldRepository struct {
	mock.Mock
}

// Field methods
func (m *MockFieldRepository) CreateField(field *domain.ProjectField) error {
	args := m.Called(field)
	return args.Error(0)
}

func (m *MockFieldRepository) FindFieldByID(id uuid.UUID) (*domain.ProjectField, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProjectField), args.Error(1)
}

func (m *MockFieldRepository) FindFieldsByProject(projectID uuid.UUID) ([]domain.ProjectField, error) {
	args := m.Called(projectID)
	return args.Get(0).([]domain.ProjectField), args.Error(1)
}

func (m *MockFieldRepository) FindFieldsByIDs(ids []uuid.UUID) ([]domain.ProjectField, error) {
	args := m.Called(ids)
	return args.Get(0).([]domain.ProjectField), args.Error(1)
}

func (m *MockFieldRepository) UpdateField(field *domain.ProjectField) error {
	args := m.Called(field)
	return args.Error(0)
}

func (m *MockFieldRepository) DeleteField(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFieldRepository) UpdateFieldOrder(fieldID uuid.UUID, newOrder int) error {
	args := m.Called(fieldID, newOrder)
	return args.Error(0)
}

func (m *MockFieldRepository) BatchUpdateFieldOrders(orders map[uuid.UUID]int) error {
	args := m.Called(orders)
	return args.Error(0)
}

// Option methods
func (m *MockFieldRepository) CreateOption(option *domain.FieldOption) error {
	args := m.Called(option)
	return args.Error(0)
}

func (m *MockFieldRepository) FindOptionByID(id uuid.UUID) (*domain.FieldOption, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FieldOption), args.Error(1)
}

func (m *MockFieldRepository) FindOptionsByField(fieldID uuid.UUID) ([]domain.FieldOption, error) {
	args := m.Called(fieldID)
	return args.Get(0).([]domain.FieldOption), args.Error(1)
}

func (m *MockFieldRepository) FindOptionsByIDs(ids []uuid.UUID) ([]domain.FieldOption, error) {
	args := m.Called(ids)
	return args.Get(0).([]domain.FieldOption), args.Error(1)
}

func (m *MockFieldRepository) UpdateOption(option *domain.FieldOption) error {
	args := m.Called(option)
	return args.Error(0)
}

func (m *MockFieldRepository) DeleteOption(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFieldRepository) UpdateOptionOrder(optionID uuid.UUID, newOrder int) error {
	args := m.Called(optionID, newOrder)
	return args.Error(0)
}

func (m *MockFieldRepository) BatchUpdateOptionOrders(orders map[uuid.UUID]int) error {
	args := m.Called(orders)
	return args.Error(0)
}

// Field Value methods
func (m *MockFieldRepository) SetFieldValue(value *domain.BoardFieldValue) error {
	args := m.Called(value)
	return args.Error(0)
}

func (m *MockFieldRepository) FindFieldValuesByBoard(boardID uuid.UUID) ([]domain.BoardFieldValue, error) {
	args := m.Called(boardID)
	return args.Get(0).([]domain.BoardFieldValue), args.Error(1)
}

func (m *MockFieldRepository) FindFieldValuesByBoardAndField(boardID, fieldID uuid.UUID) ([]domain.BoardFieldValue, error) {
	args := m.Called(boardID, fieldID)
	return args.Get(0).([]domain.BoardFieldValue), args.Error(1)
}

func (m *MockFieldRepository) FindFieldValuesByBoards(boardIDs []uuid.UUID) (map[uuid.UUID][]domain.BoardFieldValue, error) {
	args := m.Called(boardIDs)
	return args.Get(0).(map[uuid.UUID][]domain.BoardFieldValue), args.Error(1)
}

func (m *MockFieldRepository) DeleteFieldValue(boardID, fieldID uuid.UUID) error {
	args := m.Called(boardID, fieldID)
	return args.Error(0)
}

func (m *MockFieldRepository) DeleteFieldValueByID(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFieldRepository) BatchSetFieldValues(values []domain.BoardFieldValue) error {
	args := m.Called(values)
	return args.Error(0)
}

func (m *MockFieldRepository) BatchDeleteFieldValues(boardID, fieldID uuid.UUID) error {
	args := m.Called(boardID, fieldID)
	return args.Error(0)
}

func (m *MockFieldRepository) UpdateBoardFieldCache(boardID uuid.UUID) (string, error) {
	args := m.Called(boardID)
	return args.String(0), args.Error(1)
}

// View methods
func (m *MockFieldRepository) CreateView(view *domain.SavedView) error {
	args := m.Called(view)
	return args.Error(0)
}

func (m *MockFieldRepository) FindViewByID(id uuid.UUID) (*domain.SavedView, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SavedView), args.Error(1)
}

func (m *MockFieldRepository) FindViewsByProject(projectID uuid.UUID) ([]domain.SavedView, error) {
	args := m.Called(projectID)
	return args.Get(0).([]domain.SavedView), args.Error(1)
}

func (m *MockFieldRepository) FindDefaultView(projectID uuid.UUID) (*domain.SavedView, error) {
	args := m.Called(projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SavedView), args.Error(1)
}

func (m *MockFieldRepository) UpdateView(view *domain.SavedView) error {
	args := m.Called(view)
	return args.Error(0)
}

func (m *MockFieldRepository) DeleteView(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// Board Order methods
func (m *MockFieldRepository) SetBoardOrder(order *domain.UserBoardOrder) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockFieldRepository) FindBoardOrdersByView(viewID, userID uuid.UUID) ([]domain.UserBoardOrder, error) {
	args := m.Called(viewID, userID)
	return args.Get(0).([]domain.UserBoardOrder), args.Error(1)
}

func (m *MockFieldRepository) BatchUpdateBoardOrders(orders []domain.UserBoardOrder) error {
	args := m.Called(orders)
	return args.Error(0)
}

func (m *MockFieldRepository) DeleteBoardOrder(viewID, userID, boardID uuid.UUID) error {
	args := m.Called(viewID, userID, boardID)
	return args.Error(0)
}

// ==================== Mock CommentRepository ====================

type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) Create(comment *domain.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockCommentRepository) FindByID(id uuid.UUID) (*domain.Comment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Comment), args.Error(1)
}

func (m *MockCommentRepository) FindByBoardID(boardID uuid.UUID) ([]domain.Comment, error) {
	args := m.Called(boardID)
	return args.Get(0).([]domain.Comment), args.Error(1)
}

func (m *MockCommentRepository) Update(comment *domain.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockCommentRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// ==================== Helper Functions ====================

// ExpectNotFoundError configures mock to return gorm.ErrRecordNotFound
func ExpectNotFoundError() error {
	return gorm.ErrRecordNotFound
}
