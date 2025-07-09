package generator

import (
	"context"
	"log/slog"
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
	slog.Info("Test generator started", "exchange", exchange)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Warn("Test generator stopped", "exchange", exchange)
			return

		case t := <-ticker.C:
			for _, pair := range testPairs {
				price := rand.Float64()*100 + 1

				update := domain.PriceUpdate{
					Symbol:    pair,
					Price:     price,
					Timestamp: t.Unix(),
					Exchange:  exchange,
				}

				select {
				case out <- update:
					slog.Debug("Generated test price", "exchange", exchange, "pair", pair, "price", price)
				default:
					slog.Warn("Output channel full, price dropped", "exchange", exchange, "pair", pair)
				}
			}
		}
	}
}
