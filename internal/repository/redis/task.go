package redis

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type TaskCacheRepository struct {
	client *redis.Client
}

func NewTaskCacheRepository(client *redis.Client) *TaskCacheRepository {
	return &TaskCacheRepository{
		client: client,
	}
}

func (r *TaskCacheRepository) Get(ctx context.Context, key string) ([]byte, error) {
	bytes, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			slog.Info("Redis cashe not found")
			return nil, nil
		}
		return nil, err
	}
	slog.Info("Redis cashe found")
	return bytes, nil
}

func (r *TaskCacheRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return err
	}
	slog.Info("Redis cashe set")
	return nil
}

func (r *TaskCacheRepository) InvalidateCalendar(ctx context.Context) error {
	var cursor uint64
	prefix := "tasks:cal:l:*"

	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, prefix, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}
	slog.Info("Redis cashe invalidated")
	return nil
}
