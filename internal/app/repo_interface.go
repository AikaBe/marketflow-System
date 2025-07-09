package app

import (
	"context"
	"marketflow/internal/domain"
	"time"
)

type APIRepo interface {
	GetPriceForSymbol(symbol string) (*domain.AggregatedResponse, error)
	GetPriceForExchange(exchange, symbol string) (*domain.AggregatedResponse, error)

	GetHighestBySymbol(symbol string) (*domain.AggregatedResponse, error)
	GetHighestByExchange(exchange, symbol string) (*domain.AggregatedResponse, error)

	QueryHighestPriceSince(symbol string, since time.Time) (*domain.AggregatedResponse, error)
	QueryHighestSinceByExchange(exchange, symbol string, since time.Time) (*domain.AggregatedResponse, error)

	GetLowestBySymbol(symbol string) (*domain.AggregatedResponse, error)
	GetLowestByExchange(exchange, symbol string) (*domain.AggregatedResponse, error)

	QueryLowestPriceSince(symbol string, since time.Time) (*domain.AggregatedResponse, error)
	QueryLowestSinceByExchange(exchange, symbol string, since time.Time) (*domain.AggregatedResponse, error)

	GetAvgBySymbol(symbol string) (*domain.AggregatedResponse, error)
	GetAvgByExchange(exchange, symbol string) (*domain.AggregatedResponse, error)

	QueryAvgSinceByExchange(exchange, symbol string, since time.Time) (*domain.AggregatedResponse, error)

	Ping() error
}

type RedisRepo interface {
	ZAdd(ctx context.Context, key string, score int64, value string) error
	ZRangeByScore(ctx context.Context, key string, min, max int64) ([]string, error)
	ZRemRangeByScore(ctx context.Context, key string, min, max int64) error
	Ping(ctx context.Context) error
}

type SavePGRepo interface {
	SaveAggregatedPrice(ctx context.Context, pair, exchange string, ts time.Time, avg, min, max float64) error
}
