package service

import (
	"board-service/internal/cache"
	"board-service/internal/client"
	"context"

	"github.com/stretchr/testify/mock"
)

// ==================== Mock UserClient ====================

type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) GetUser(ctx context.Context, userID string) (*client.UserInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.UserInfo), args.Error(1)
}

func (m *MockUserClient) GetUsersBatch(ctx context.Context, userIDs []string) ([]client.UserInfo, error) {
	args := m.Called(ctx, userIDs)
	return args.Get(0).([]client.UserInfo), args.Error(1)
}

func (m *MockUserClient) SearchUsers(ctx context.Context, query string) ([]client.UserInfo, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]client.UserInfo), args.Error(1)
}

func (m *MockUserClient) GetSimpleUser(userID string) (*client.SimpleUser, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.SimpleUser), args.Error(1)
}

func (m *MockUserClient) GetSimpleUsers(userIDs []string) ([]client.SimpleUser, error) {
	args := m.Called(userIDs)
	return args.Get(0).([]client.SimpleUser), args.Error(1)
}

func (m *MockUserClient) CheckWorkspaceExists(ctx context.Context, workspaceID string, token string) (bool, error) {
	args := m.Called(ctx, workspaceID, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserClient) ValidateWorkspaceMembership(ctx context.Context, workspaceID string, userID string, token string) (bool, error) {
	args := m.Called(ctx, workspaceID, userID, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserClient) GetWorkspace(ctx context.Context, workspaceID string, token string) (*client.WorkspaceInfo, error) {
	args := m.Called(ctx, workspaceID, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.WorkspaceInfo), args.Error(1)
}

// ==================== Mock UserInfoCache ====================

type MockUserInfoCache struct {
	mock.Mock
}

func (m *MockUserInfoCache) GetUserInfo(ctx context.Context, userID string) (bool, *cache.UserInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(1) == nil {
		return args.Bool(0), nil, args.Error(2)
	}
	return args.Bool(0), args.Get(1).(*cache.UserInfo), args.Error(2)
}

func (m *MockUserInfoCache) SetUserInfo(ctx context.Context, userInfo *cache.UserInfo) error {
	args := m.Called(ctx, userInfo)
	return args.Error(0)
}

func (m *MockUserInfoCache) GetSimpleUser(ctx context.Context, userID string) (bool, *cache.SimpleUser, error) {
	args := m.Called(ctx, userID)
	if args.Get(1) == nil {
		return args.Bool(0), nil, args.Error(2)
	}
	return args.Bool(0), args.Get(1).(*cache.SimpleUser), args.Error(2)
}

func (m *MockUserInfoCache) SetSimpleUser(ctx context.Context, simpleUser *cache.SimpleUser) error {
	args := m.Called(ctx, simpleUser)
	return args.Error(0)
}

func (m *MockUserInfoCache) GetSimpleUsersBatch(ctx context.Context, userIDs []string) (map[string]*cache.SimpleUser, error) {
	args := m.Called(ctx, userIDs)
	return args.Get(0).(map[string]*cache.SimpleUser), args.Error(1)
}

func (m *MockUserInfoCache) SetSimpleUsersBatch(ctx context.Context, simpleUsers []cache.SimpleUser) error {
	args := m.Called(ctx, simpleUsers)
	return args.Error(0)
}

func (m *MockUserInfoCache) DeleteUserInfo(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserInfoCache) InvalidateUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// ==================== Mock WorkspaceCache ====================

type MockWorkspaceCache struct {
	mock.Mock
}

func (m *MockWorkspaceCache) GetMembership(ctx context.Context, workspaceID, userID string) (cacheExists bool, isMember bool, err error) {
	args := m.Called(ctx, workspaceID, userID)
	return args.Bool(0), args.Bool(1), args.Error(2)
}

func (m *MockWorkspaceCache) SetMembership(ctx context.Context, workspaceID, userID string, isMember bool) error {
	args := m.Called(ctx, workspaceID, userID, isMember)
	return args.Error(0)
}

func (m *MockWorkspaceCache) InvalidateMembership(ctx context.Context, workspaceID, userID string) error {
	args := m.Called(ctx, workspaceID, userID)
	return args.Error(0)
}

func (m *MockWorkspaceCache) ValidateWorkspace(ctx context.Context, workspaceID string, userID string, token string) (bool, error) {
	args := m.Called(ctx, workspaceID, userID, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockWorkspaceCache) InvalidateWorkspace(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}
