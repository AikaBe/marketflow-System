package main

import (
	"context"
	"fmt"
	"log"
	"marketflow/internal/adapters/postgres"
	"marketflow/internal/adapters/redis"
	"marketflow/internal/app/aggregator"
	"marketflow/internal/app/api"
	"marketflow/internal/app/mode"
	"marketflow/internal/domain"
	"marketflow/internal/handler"
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

	aggregatedAdapter, err := postgres.NewAggregatedAdapter(connStr)
	if err != nil {
		log.Fatalf("LatestAdapter error: %v", err)
	}
	defer aggregatedAdapter.Close()

	redisAdapter := redis.NewRedisAdapter("redis:6379", "", 0)

	updates := make(chan domain.PriceUpdate, 1000)
	modeManager := mode.NewModeManager(updates)
	modeManager.SetMode(ctx, mode.ModeLive)

	service := aggregator.NewServiceCom(redisAdapter, pgAdapter)

	service.StartRedisWorkerPool(ctx, updates, 5)
	go service.StartAggregator(ctx)

	apiService := api.NewService(aggregatedAdapter)

	handler := handler.NewHandler(apiService, modeManager)

	http.HandleFunc("/prices/latest/", handler.Handle)
	http.HandleFunc("/prices/highest/", handler.Highest)
	http.HandleFunc("/mode/test", handler.SwitchToTestMode)
	http.HandleFunc("/mode/live", handler.SwitchToLiveMode)

	go func() {
		log.Println("Starting HTTP server on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	fmt.Println("Service is running...")
	<-ctx.Done()
	fmt.Println("Shutting down...")
}
