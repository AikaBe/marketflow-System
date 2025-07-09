package aggregator

import (
	"context"
	"fmt"
	"marketflow/internal/app"
	"marketflow/internal/domain"
	"strconv"
	"time"
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
			for update := range input {
				key := fmt.Sprintf("price:%s:%s", update.Symbol, update.Exchange)
				timestamp := time.Now().Unix()
				value := fmt.Sprintf("%f", update.Price)
				ls.redisRepo.ZAdd(ctx, key, timestamp, value)
				_ = ls.redisRepo.ZRemRangeByScore(ctx, key, 0, timestamp-61)
			}
		}(i)
	}
}

func (ls *ServiceCom) StartAggregator(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	pairs := domain.TradingPairs
	exchanges := domain.ExchangeNames

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, pair := range pairs {
				for _, ex := range exchanges {
					key := fmt.Sprintf("price:%s:%s", pair, ex)
					now := time.Now().Unix()
					values, err := ls.redisRepo.ZRangeByScore(ctx, key, now-60, now)
					if err != nil || len(values) == 0 {
						continue
					}
					var prices []float64
					for _, v := range values {
						price, err := strconv.ParseFloat(v, 64)
						if err == nil {
							prices = append(prices, price)
						}
					}
					if len(prices) == 0 {
						continue
					}
					min, max, avg := calcStats(prices)
					ls.pgSave.SaveAggregatedPrice(ctx, pair, ex, time.Now(), avg, min, max)
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
