package redis

import (
	"context"
	"strconv"

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

func (r *Adapter) ZAdd(ctx context.Context, key string, score int64, value string) error {
	return r.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(score),
		Member: value,
	}).Err()
}

func (r *Adapter) ZRangeByScore(ctx context.Context, key string, min, max int64) ([]string, error) {
	return r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: strconv.FormatInt(min, 10),
		Max: strconv.FormatInt(max, 10),
	}).Result()
}

func (r *Adapter) ZRemRangeByScore(ctx context.Context, key string, min, max int64) error {
	return r.client.ZRemRangeByScore(ctx, key,
		strconv.FormatInt(min, 10),
		strconv.FormatInt(max, 10),
	).Err()
}
