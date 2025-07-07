package app

import (
	"context"
	"marketflow/internal/adapters/postgres"
	"marketflow/internal/adapters/redis"
	"marketflow/internal/domain"
)

type General interface {
	StartRedisWorkerPool(ctx context.Context, redisAdapter *redis.Adapter, input <-chan domain.PriceUpdate, workers int)
	StartAggregator(ctx context.Context, redisAdapter *redis.Adapter, pgAdapter *postgres.Adapter)
}

type Latest interface {
	GetAggregatedPriceForSymbol(symbol string) (*domain.AggregatedResponse, error)
	GetAggregatedPriceForExchange(parts string) (*domain.AggregatedResponse, error)
}
