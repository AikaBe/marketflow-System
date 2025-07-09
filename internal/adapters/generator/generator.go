package generator

import (
	"context"
	"log"
	"marketflow/internal/domain"
	"math/rand"
	"time"
)

var testPairs = []string{"BTCUSDT", "ETHUSDT", "DOGEUSDT", "TONUSDT", "SOLUSDT"}

func StartTestGenerators(ctx context.Context, out chan<- domain.PriceUpdate) {
	exchanges := []string{"Exchange1", "Exchange2", "Exchange3"}
	for _, exchange := range exchanges {
		go generateForExchange(ctx, exchange, out)
	}
}

func generateForExchange(ctx context.Context, exchange string, out chan<- domain.PriceUpdate) {
	log.Printf("[%s] [TEST MODE] Generator started", exchange)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] [TEST MODE] Generator stopped", exchange)
			return
		case <-ticker.C:
			for _, pair := range testPairs {
				price := rand.Float64()*100 + 1
				out <- domain.PriceUpdate{
					Symbol:    pair,
					Price:     price,
					Timestamp: time.Now().Unix(),
					Exchange:  exchange,
				}
			}
		}
	}
}
