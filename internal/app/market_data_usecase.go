package app

import (
	"context"
	"fmt"
	"marketflow/internal/adapters/postgres"
	"marketflow/internal/adapters/redis"
	"marketflow/internal/domain"
	"strconv"
	"strings"
	"time"
)

func StartRedisWorkerPool(ctx context.Context, redisAdapter *redis.Adapter, input <-chan domain.PriceUpdate, workers int) {
	for i := 0; i < workers; i++ {
		go func(id int) {
			for update := range input {
				key := fmt.Sprintf("price:%s:%s", update.Symbol, update.Exchange)
				value := fmt.Sprintf("%f:%d", update.Price, update.Timestamp)
				redisAdapter.AppendToList(ctx, key, value)
			}
		}(i)
	}
}

func StartAggregator(ctx context.Context, redisAdapter *redis.Adapter, pgAdapter *postgres.Adapter) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	pairs := []string{"BTCUSDT", "ETHUSDT", "DOGEUSDT", "TONUSDT", "SOLUSDT"}
	exchanges := []string{"Exchange1", "Exchange2", "Exchange3"}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, pair := range pairs {
				for _, ex := range exchanges {
					key := fmt.Sprintf("price:%s:%s", pair, ex)
					values, err := redisAdapter.GetLastN(ctx, key, 60)
					if err != nil || len(values) == 0 {
						continue
					}
					var prices []float64
					for _, v := range values {
						parts := strings.Split(v, ":")
						if len(parts) != 2 {
							continue
						}
						price, err := strconv.ParseFloat(parts[0], 64)
						if err == nil {
							prices = append(prices, price)
						}
					}
					if len(prices) == 0 {
						continue
					}
					min, max, avg := calcStats(prices)
					pgAdapter.SaveAggregatedPrice(ctx, pair, ex, time.Now(), avg, min, max)
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
