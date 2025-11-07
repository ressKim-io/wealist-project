package client

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// mockUserClient is a mock implementation of UserClient for development/testing
// It bypasses User Service validation and returns hardcoded responses
type mockUserClient struct {
	logger *zap.Logger
}

// NewMockUserClient creates a mock User Service client for development
func NewMockUserClient(logger *zap.Logger) UserClient {
	logger.Warn("Using MOCK User Service client - workspace validation disabled!")
	return &mockUserClient{
		logger: logger,
	}
}

// GetUser returns a mock user
func (c *mockUserClient) GetUser(ctx context.Context, userID string) (*UserInfo, error) {
	c.logger.Debug("Mock GetUser called", zap.String("user_id", userID))
	return &UserInfo{
		ID:        userID,
		Email:     "mock@example.com",
		Name:      "Mock User",
		AvatarURL: "https://via.placeholder.com/150",
	}, nil
}

// GetUsersBatch returns mock users
func (c *mockUserClient) GetUsersBatch(ctx context.Context, userIDs []string) ([]UserInfo, error) {
	c.logger.Debug("Mock GetUsersBatch called", zap.Int("count", len(userIDs)))
	users := make([]UserInfo, len(userIDs))
	for i, id := range userIDs {
		users[i] = UserInfo{
			ID:        id,
			Email:     fmt.Sprintf("mock%d@example.com", i),
			Name:      fmt.Sprintf("Mock User %d", i),
			AvatarURL: "https://via.placeholder.com/150",
		}
	}
	return users, nil
}

// SearchUsers returns mock search results
func (c *mockUserClient) SearchUsers(ctx context.Context, query string) ([]UserInfo, error) {
	c.logger.Debug("Mock SearchUsers called", zap.String("query", query))
	return []UserInfo{
		{
			ID:        "00000000-0000-0000-0000-000000000001",
			Email:     "user1@example.com",
			Name:      "Test User 1",
			AvatarURL: "https://via.placeholder.com/150",
		},
		{
			ID:        "00000000-0000-0000-0000-000000000002",
			Email:     "user2@example.com",
			Name:      "Test User 2",
			AvatarURL: "https://via.placeholder.com/150",
		},
	}, nil
}

// GetSimpleUser returns a mock simple user
func (c *mockUserClient) GetSimpleUser(userID string) (*SimpleUser, error) {
	c.logger.Debug("Mock GetSimpleUser called", zap.String("user_id", userID))
	return &SimpleUser{
		ID:        userID,
		Name:      "Mock User",
		AvatarURL: "https://via.placeholder.com/150",
	}, nil
}

// GetSimpleUsers returns mock simple users
func (c *mockUserClient) GetSimpleUsers(userIDs []string) ([]SimpleUser, error) {
	c.logger.Debug("Mock GetSimpleUsers called", zap.Int("count", len(userIDs)))
	users := make([]SimpleUser, len(userIDs))
	for i, id := range userIDs {
		users[i] = SimpleUser{
			ID:        id,
			Name:      fmt.Sprintf("Mock User %d", i),
			AvatarURL: "https://via.placeholder.com/150",
		}
	}
	return users, nil
}

// CheckWorkspaceExists always returns true in mock mode
func (c *mockUserClient) CheckWorkspaceExists(ctx context.Context, workspaceID string, token string) (bool, error) {
	c.logger.Debug("Mock CheckWorkspaceExists - always returns true", zap.String("workspace_id", workspaceID))
	return true, nil
}

// ValidateWorkspaceMembership always returns true in mock mode
func (c *mockUserClient) ValidateWorkspaceMembership(ctx context.Context, workspaceID string, userID string, token string) (bool, error) {
	c.logger.Debug("Mock ValidateWorkspaceMembership - always returns true",
		zap.String("workspace_id", workspaceID),
		zap.String("user_id", userID))
	return true, nil
}

// GetWorkspace returns a mock workspace
func (c *mockUserClient) GetWorkspace(ctx context.Context, workspaceID string, token string) (*WorkspaceInfo, error) {
	c.logger.Debug("Mock GetWorkspace called", zap.String("workspace_id", workspaceID))
	return &WorkspaceInfo{
		ID:          workspaceID,
		Name:        "Mock Workspace",
		Description: "Mock workspace for development",
		OwnerID:     "00000000-0000-0000-0000-000000000000",
		IsPublic:    true,
	}, nil
}
