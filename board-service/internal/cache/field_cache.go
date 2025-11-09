package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// FieldCache handles caching for custom fields system
type FieldCache interface {
	// Project fields caching
	GetProjectFields(ctx context.Context, projectID string) ([]byte, error)
	SetProjectFields(ctx context.Context, projectID string, fieldsJSON []byte, ttl time.Duration) error
	InvalidateProjectFields(ctx context.Context, projectID string) error

	// Field options caching
	GetFieldOptions(ctx context.Context, fieldID string) ([]byte, error)
	SetFieldOptions(ctx context.Context, fieldID string, optionsJSON []byte, ttl time.Duration) error
	InvalidateFieldOptions(ctx context.Context, fieldID string) error

	// Board field values caching
	GetBoardFieldValues(ctx context.Context, boardID string) (map[string]interface{}, error)
	SetBoardFieldValues(ctx context.Context, boardID string, values map[string]interface{}, ttl time.Duration) error
	InvalidateBoardFieldValues(ctx context.Context, boardID string) error

	// View results caching (with filter hash)
	GetViewResults(ctx context.Context, viewID, filterHash string) ([]byte, error)
	SetViewResults(ctx context.Context, viewID, filterHash string, resultsJSON []byte, ttl time.Duration) error
	InvalidateViewResults(ctx context.Context, viewID string) error
}

type fieldCache struct {
	client *redis.Client
}

func NewFieldCache(client *redis.Client) FieldCache {
	return &fieldCache{client: client}
}

// ==================== Project Fields ====================

func (c *fieldCache) GetProjectFields(ctx context.Context, projectID string) ([]byte, error) {
	key := fmt.Sprintf("project:%s:fields", projectID)
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (c *fieldCache) SetProjectFields(ctx context.Context, projectID string, fieldsJSON []byte, ttl time.Duration) error {
	key := fmt.Sprintf("project:%s:fields", projectID)
	return c.client.Set(ctx, key, fieldsJSON, ttl).Err()
}

func (c *fieldCache) InvalidateProjectFields(ctx context.Context, projectID string) error {
	key := fmt.Sprintf("project:%s:fields", projectID)
	return c.client.Del(ctx, key).Err()
}

// ==================== Field Options ====================

func (c *fieldCache) GetFieldOptions(ctx context.Context, fieldID string) ([]byte, error) {
	key := fmt.Sprintf("field:%s:options", fieldID)
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (c *fieldCache) SetFieldOptions(ctx context.Context, fieldID string, optionsJSON []byte, ttl time.Duration) error {
	key := fmt.Sprintf("field:%s:options", fieldID)
	return c.client.Set(ctx, key, optionsJSON, ttl).Err()
}

func (c *fieldCache) InvalidateFieldOptions(ctx context.Context, fieldID string) error {
	key := fmt.Sprintf("field:%s:options", fieldID)
	return c.client.Del(ctx, key).Err()
}

// ==================== Board Field Values ====================

func (c *fieldCache) GetBoardFieldValues(ctx context.Context, boardID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("board:%s:field_values", boardID)
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var values map[string]interface{}
	if err := json.Unmarshal(val, &values); err != nil {
		return nil, err
	}

	return values, nil
}

func (c *fieldCache) SetBoardFieldValues(ctx context.Context, boardID string, values map[string]interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("board:%s:field_values", boardID)
	valJSON, err := json.Marshal(values)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, valJSON, ttl).Err()
}

func (c *fieldCache) InvalidateBoardFieldValues(ctx context.Context, boardID string) error {
	key := fmt.Sprintf("board:%s:field_values", boardID)
	return c.client.Del(ctx, key).Err()
}

// ==================== View Results ====================

func (c *fieldCache) GetViewResults(ctx context.Context, viewID, filterHash string) ([]byte, error) {
	key := fmt.Sprintf("view:%s:results:%s", viewID, filterHash)
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (c *fieldCache) SetViewResults(ctx context.Context, viewID, filterHash string, resultsJSON []byte, ttl time.Duration) error {
	key := fmt.Sprintf("view:%s:results:%s", viewID, filterHash)
	return c.client.Set(ctx, key, resultsJSON, ttl).Err()
}

func (c *fieldCache) InvalidateViewResults(ctx context.Context, viewID string) error {
	// Delete all view results for this view
	pattern := fmt.Sprintf("view:%s:results:*", viewID)
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}
