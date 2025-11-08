package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// WorkspaceCache handles caching of workspace membership validation results
type WorkspaceCache interface {
	// GetMembership returns (cacheExists, isMember, error)
	// If cacheExists is false, the caller should validate via User Service and call SetMembership
	GetMembership(ctx context.Context, workspaceID, userID string) (cacheExists bool, isMember bool, err error)

	// SetMembership stores the membership validation result in cache
	SetMembership(ctx context.Context, workspaceID, userID string, isMember bool) error

	// InvalidateMembership removes the cached membership for a specific user-workspace pair
	InvalidateMembership(ctx context.Context, workspaceID, userID string) error

	// InvalidateWorkspace removes all cached memberships for a workspace (e.g., when workspace is deleted)
	InvalidateWorkspace(ctx context.Context, workspaceID string) error
}

type workspaceCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewWorkspaceCache creates a new workspace cache instance
// TTL is set to 5 minutes to balance between performance and data freshness
func NewWorkspaceCache(client *redis.Client) WorkspaceCache {
	return &workspaceCache{
		client: client,
		ttl:    5 * time.Minute, // 5분 TTL
	}
}

func (c *workspaceCache) GetMembership(ctx context.Context, workspaceID, userID string) (bool, bool, error) {
	key := c.membershipKey(workspaceID, userID)
	val, err := c.client.Get(ctx, key).Result()

	if err == redis.Nil {
		// Cache miss - not an error, just means we need to fetch from User Service
		return false, false, nil
	}
	if err != nil {
		// Actual error (network, Redis down, etc.)
		return false, false, fmt.Errorf("failed to get workspace membership from cache: %w", err)
	}

	// Cache hit
	isMember := val == "true"
	return true, isMember, nil
}

func (c *workspaceCache) SetMembership(ctx context.Context, workspaceID, userID string, isMember bool) error {
	key := c.membershipKey(workspaceID, userID)
	val := "false"
	if isMember {
		val = "true"
	}

	err := c.client.Set(ctx, key, val, c.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set workspace membership in cache: %w", err)
	}

	return nil
}

func (c *workspaceCache) InvalidateMembership(ctx context.Context, workspaceID, userID string) error {
	key := c.membershipKey(workspaceID, userID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to invalidate workspace membership: %w", err)
	}
	return nil
}

func (c *workspaceCache) InvalidateWorkspace(ctx context.Context, workspaceID string) error {
	// Find all keys matching the pattern
	pattern := fmt.Sprintf("workspace_member:%s:*", workspaceID)

	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	keys := []string{}

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan workspace membership keys: %w", err)
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete workspace membership keys: %w", err)
		}
	}

	return nil
}

func (c *workspaceCache) membershipKey(workspaceID, userID string) string {
	return fmt.Sprintf("workspace_member:%s:%s", workspaceID, userID)
}
