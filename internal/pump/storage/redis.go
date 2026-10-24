// Package storage reads analytics records queued by iam-authz-server.
package storage

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// AnalyticsKey is the Redis list consumed by iam-pump. iam-authz-server writes
// to the same key.
const AnalyticsKey = "analytics-iam-system-analytics"

// RedisStorage reads and deletes analytics batches atomically.
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage creates a storage backed by the given Redis client.
func NewRedisStorage(client *redis.Client) *RedisStorage {
	return &RedisStorage{client: client}
}

// GetAndDeleteSet reads every element of the list and deletes it in one
// transaction, so that concurrent pump instances cannot process the same
// record twice.
func (s *RedisStorage) GetAndDeleteSet(ctx context.Context) ([]string, error) {
	var lrange *redis.StringSliceCmd

	_, err := s.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		lrange = pipe.LRange(ctx, AnalyticsKey, 0, -1)
		pipe.Del(ctx, AnalyticsKey)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read analytics batch: %w", err)
	}

	return lrange.Val(), nil
}
