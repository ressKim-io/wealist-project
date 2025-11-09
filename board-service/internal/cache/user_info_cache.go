package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// UserInfo represents detailed user information (matches client.UserInfo)
type UserInfo struct {
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"isActive"`
}

// SimpleUser represents basic user information (matches client.SimpleUser)
type SimpleUser struct {
	ID        string `json:"user_id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

// UserInfoCache handles caching of user information
type UserInfoCache interface {
	// GetUserInfo returns cached detailed user info
	// Returns (cacheExists, userInfo, error)
	GetUserInfo(ctx context.Context, userID string) (cacheExists bool, userInfo *UserInfo, err error)

	// SetUserInfo stores detailed user info in cache
	SetUserInfo(ctx context.Context, userInfo *UserInfo) error

	// GetSimpleUser returns cached simple user info
	// Returns (cacheExists, simpleUser, error)
	GetSimpleUser(ctx context.Context, userID string) (cacheExists bool, simpleUser *SimpleUser, err error)

	// SetSimpleUser stores simple user info in cache
	SetSimpleUser(ctx context.Context, simpleUser *SimpleUser) error

	// GetSimpleUsersBatch returns multiple cached simple users
	// Returns map[userID]SimpleUser with only cached entries (missing IDs not included)
	GetSimpleUsersBatch(ctx context.Context, userIDs []string) (map[string]*SimpleUser, error)

	// SetSimpleUsersBatch stores multiple simple users in cache
	SetSimpleUsersBatch(ctx context.Context, simpleUsers []SimpleUser) error

	// InvalidateUser removes all cached data for a specific user
	InvalidateUser(ctx context.Context, userID string) error
}

type userInfoCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewUserInfoCache creates a new user info cache instance
// TTL is set to 10 minutes to balance between performance and data freshness
func NewUserInfoCache(client *redis.Client) UserInfoCache {
	return &userInfoCache{
		client: client,
		ttl:    10 * time.Minute, // 10분 TTL
	}
}

func (c *userInfoCache) GetUserInfo(ctx context.Context, userID string) (bool, *UserInfo, error) {
	key := c.userInfoKey(userID)
	val, err := c.client.Get(ctx, key).Result()

	if err == redis.Nil {
		// Cache miss
		return false, nil, nil
	}
	if err != nil {
		return false, nil, fmt.Errorf("failed to get user info from cache: %w", err)
	}

	var userInfo UserInfo
	if err := json.Unmarshal([]byte(val), &userInfo); err != nil {
		return false, nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	return true, &userInfo, nil
}

func (c *userInfoCache) SetUserInfo(ctx context.Context, userInfo *UserInfo) error {
	key := c.userInfoKey(userInfo.UserID)

	data, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal user info: %w", err)
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set user info in cache: %w", err)
	}

	return nil
}

func (c *userInfoCache) GetSimpleUser(ctx context.Context, userID string) (bool, *SimpleUser, error) {
	key := c.simpleUserKey(userID)
	val, err := c.client.Get(ctx, key).Result()

	if err == redis.Nil {
		// Cache miss
		return false, nil, nil
	}
	if err != nil {
		return false, nil, fmt.Errorf("failed to get simple user from cache: %w", err)
	}

	var simpleUser SimpleUser
	if err := json.Unmarshal([]byte(val), &simpleUser); err != nil {
		return false, nil, fmt.Errorf("failed to unmarshal simple user: %w", err)
	}

	return true, &simpleUser, nil
}

func (c *userInfoCache) SetSimpleUser(ctx context.Context, simpleUser *SimpleUser) error {
	key := c.simpleUserKey(simpleUser.ID)

	data, err := json.Marshal(simpleUser)
	if err != nil {
		return fmt.Errorf("failed to marshal simple user: %w", err)
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set simple user in cache: %w", err)
	}

	return nil
}

func (c *userInfoCache) GetSimpleUsersBatch(ctx context.Context, userIDs []string) (map[string]*SimpleUser, error) {
	if len(userIDs) == 0 {
		return make(map[string]*SimpleUser), nil
	}

	keys := make([]string, len(userIDs))
	for i, userID := range userIDs {
		keys[i] = c.simpleUserKey(userID)
	}

	// Use MGET to fetch multiple keys at once
	values, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get simple users batch from cache: %w", err)
	}

	result := make(map[string]*SimpleUser)
	for i, val := range values {
		if val == nil {
			// Cache miss for this user
			continue
		}

		valStr, ok := val.(string)
		if !ok {
			continue
		}

		var simpleUser SimpleUser
		if err := json.Unmarshal([]byte(valStr), &simpleUser); err != nil {
			// Log but continue with other users
			continue
		}

		result[userIDs[i]] = &simpleUser
	}

	return result, nil
}

func (c *userInfoCache) SetSimpleUsersBatch(ctx context.Context, simpleUsers []SimpleUser) error {
	if len(simpleUsers) == 0 {
		return nil
	}

	// Use pipeline for efficient batch write
	pipe := c.client.Pipeline()

	for _, user := range simpleUsers {
		key := c.simpleUserKey(user.ID)
		data, err := json.Marshal(user)
		if err != nil {
			// Skip this user but continue with others
			continue
		}
		pipe.Set(ctx, key, data, c.ttl)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to set simple users batch in cache: %w", err)
	}

	return nil
}

func (c *userInfoCache) InvalidateUser(ctx context.Context, userID string) error {
	keys := []string{
		c.userInfoKey(userID),
		c.simpleUserKey(userID),
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to invalidate user cache: %w", err)
	}

	return nil
}

func (c *userInfoCache) userInfoKey(userID string) string {
	return fmt.Sprintf("user_info:%s", userID)
}

func (c *userInfoCache) simpleUserKey(userID string) string {
	return fmt.Sprintf("simple_user:%s", userID)
}
