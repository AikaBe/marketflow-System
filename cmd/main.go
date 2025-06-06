package main

import (
	"context"
	"fmt"
	"log"
	"marketflow/internal/adapters/postgres"
	"marketflow/internal/adapters/redis"
	"marketflow/internal/adapters/websocket/impl"
	"marketflow/internal/app"
	"os"
	"os/signal"

	_ "github.com/lib/pq"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	connStr := "host=postgres port=5432 user=market password=secret dbname=marketdb sslmode=disable"
	pgAdapter, err := postgres.NewPostgresAdapter(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer func() {
		if err := pgAdapter.Close(); err != nil {
			log.Printf("Failed to close DB: %v", err)
		}
	}()
	fmt.Println("Successfully connected to PostgreSQL")

	redisAdapter := redis.NewRedisAdapter("redis:6379", "", 0)
	if err := redisAdapter.Set(ctx, "health", "ok"); err != nil {
		log.Fatalf("Redis error: %v", err)
	}
	fmt.Println("Successfully connected to Redis")

	binance := impl.NewBinanceAdapter()
	bybit := impl.NewBybitAdapter()

	fmt.Println("Binance and Bybit adapters created")

	service := app.NewMarketDataService(
		binance,
		bybit,
	)

	dataChan := service.Start(ctx)

	fmt.Println("Market data service started")

	go func() {
		for data := range dataChan {
			fmt.Printf("[%s] %s: %.2f (%.2f) at %d\n",
				data.Exchange, data.Symbol, data.Price, data.Volume, data.Time)
		}
	}()

	<-ctx.Done()
	fmt.Println("Stopping market data service...")
	service.Stop()
	fmt.Println("Graceful shutdown")
}
