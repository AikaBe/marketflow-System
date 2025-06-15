package main

import (
	"context"
	"fmt"
	"log"
	"marketflow/internal/adapters/postgres"
	"marketflow/internal/adapters/redis"
	"marketflow/internal/adapters/websocket/tar_files"
	"marketflow/internal/app"
	"marketflow/internal/domain"
	"marketflow/internal/handlers"
	"net/http"
	"os"
	"os/signal"

	_ "github.com/lib/pq"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	connStr := "host=postgres port=5432 user=market password=secret dbname=marketdb sslmode=disable"
	pgAdapter, err := postgres.NewPostgresAdapter(connStr)
	if err != nil {
		log.Fatalf("Postgres error: %v", err)
	}
	defer pgAdapter.Close()

	redisAdapter := redis.NewRedisAdapter("redis:6379", "", 0)

	updates := make(chan domain.PriceUpdate, 1000)
	tar_files.StartReaders(updates)

	app.StartRedisWorkerPool(ctx, redisAdapter, updates, 5)
	go app.StartAggregator(ctx, redisAdapter, pgAdapter)

	// HTTP API
	handler := &handlers.AggregatedHandler{PG: pgAdapter}
	http.Handle("/aggregated", handler)
	go func() {
		log.Println("Starting HTTP server on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	fmt.Println("Service is running...")
	<-ctx.Done()
	fmt.Println("Shutting down...")
}
