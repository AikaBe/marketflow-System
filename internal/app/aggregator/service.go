package aggregator

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"marketflow/internal/app"
	"marketflow/internal/domain"
)

type ServiceCom struct {
	pgSave    app.SavePGRepo
	redisRepo app.RedisRepo
}

func NewServiceCom(redisAdapter app.RedisRepo, pgAdapter app.SavePGRepo) *ServiceCom {
	return &ServiceCom{redisRepo: redisAdapter, pgSave: pgAdapter}
}

func (ls *ServiceCom) StartRedisWorkerPool(ctx context.Context, input <-chan domain.PriceUpdate, workers int) {
	for i := 0; i < workers; i++ {
		go func(id int) {
			slog.Info("Redis worker started", "worker_id", id)
			for update := range input {
				key := "price:" + update.Symbol + ":" + update.Exchange
				timestamp := time.Now().Unix()
				value := strconv.FormatFloat(update.Price, 'f', -1, 64)

				if err := ls.redisRepo.ZAdd(ctx, key, timestamp, value); err != nil {
					slog.Error("Failed to ZAdd to Redis", "worker_id", id, "key", key, "err", err)
					continue
				}
				slog.Debug("ZAdd success", "worker_id", id, "key", key, "price", update.Price)

				if err := ls.redisRepo.ZRemRangeByScore(ctx, key, 0, timestamp-61); err != nil {
					slog.Warn("Failed to clean up old entries in Redis", "worker_id", id, "key", key, "err", err)
				}
			}
			slog.Warn("Redis worker exiting", "worker_id", id)
		}(i)
	}
}

func (ls *ServiceCom) StartAggregator(ctx context.Context) {
	slog.Info("Aggregator started")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	pairs := domain.TradingPairs
	exchanges := domain.ExchangeNames

	for {
		select {
		case <-ctx.Done():
			slog.Warn("Aggregator context cancelled, stopping...")
			return
		case <-ticker.C:
			now := time.Now()
			currentTimestamp := now.Unix()

			for _, pair := range pairs {
				for _, ex := range exchanges {
					key := "price:" + pair + ":" + ex
					values, err := ls.redisRepo.ZRangeByScore(ctx, key, currentTimestamp-60, currentTimestamp)
					if err != nil {
						slog.Error("Failed to get prices from Redis", "key", key, "err", err)
						continue
					}
					if len(values) == 0 {
						slog.Debug("No prices found in Redis", "key", key)
						continue
					}

					var prices []float64
					for _, v := range values {
						price, err := strconv.ParseFloat(v, 64)
						if err != nil {
							slog.Warn("Failed to parse price from Redis", "value", v, "err", err)
							continue
						}
						prices = append(prices, price)
					}

					if len(prices) == 0 {
						slog.Debug("No valid prices parsed for aggregation", "key", key)
						continue
					}

					min, max, avg := calcStats(prices)
					err = ls.pgSave.SaveAggregatedPrice(ctx, pair, ex, now, avg, min, max)
					if err != nil {
						slog.Error("Failed to save aggregated price to DB", "pair", pair, "exchange", ex, "err", err)
					} else {
						slog.Info("Aggregated price saved",
							"pair", pair,
							"exchange", ex,
							"avg", avg,
							"min", min,
							"max", max,
						)
					}
				}
			}
		}
	}
}

func calcStats(prices []float64) (min, max, avg float64) {
	if len(prices) == 0 {
		return 0, 0, 0
	}
	min, max = prices[0], prices[0]
	sum := 0.0
	for _, p := range prices {
		sum += p
		if p < min {
			min = p
		}
		if p > max {
			max = p
		}
	}
	avg = sum / float64(len(prices))
	return
}
