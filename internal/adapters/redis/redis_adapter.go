package redis

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Adapter struct {
	client *redis.Client
}

func NewRedisAdapter(addr, password string, db int) *Adapter {
	slog.Info("Connecting to Redis",
		"addr", addr,
		"db", db,
	)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	slog.Info("Redis client created")
	return &Adapter{client: client}
}

func (r *Adapter) Ping(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		slog.Error("Redis ping failed", "err", err)
	} else {
		slog.Info("Redis ping successful")
	}
	return err
}

func (r *Adapter) ZAdd(ctx context.Context, key string, score int64, value string) error {
	slog.Info("ZAdd operation", "key", key, "score", score, "value", value)

	err := r.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(score),
		Member: value,
	}).Err()
	if err != nil {
		slog.Error("ZAdd failed", "key", key, "score", score, "value", value, "err", err)
	}
	return err
}

func (r *Adapter) ZRangeByScore(ctx context.Context, key string, min, max int64) ([]string, error) {
	slog.Info("ZRangeByScore operation", "key", key, "min", min, "max", max)

	result, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: strconv.FormatInt(min, 10),
		Max: strconv.FormatInt(max, 10),
	}).Result()

	if err != nil {
		slog.Error("ZRangeByScore failed", "key", key, "min", min, "max", max, "err", err)
	} else {
		slog.Info("ZRangeByScore success", "key", key, "count", len(result))
	}
	return result, err
}

func (r *Adapter) ZRemRangeByScore(ctx context.Context, key string, min, max int64) error {
	slog.Info("ZRemRangeByScore operation", "key", key, "min", min, "max", max)

	err := r.client.ZRemRangeByScore(ctx, key,
		strconv.FormatInt(min, 10),
		strconv.FormatInt(max, 10),
	).Err()
	if err != nil {
		slog.Error("ZRemRangeByScore failed", "key", key, "min", min, "max", max, "err", err)
	}
	return err
}
