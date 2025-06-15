package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Adapter struct {
	client *redis.Client
}

func NewRedisAdapter(addr, password string, db int) *Adapter {
	return &Adapter{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

func (a *Adapter) AppendToList(ctx context.Context, key, value string) error {
	return a.client.LPush(ctx, key, value).Err()
}

func (a *Adapter) GetLastN(ctx context.Context, key string, n int64) ([]string, error) {
	return a.client.LRange(ctx, key, 0, n-1).Result()
}
