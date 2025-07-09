package main

import (
	"context"
	"log/slog"
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
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	connStr := "host=postgres port=5432 user=market password=secret dbname=marketdb sslmode=disable"
	pgAdapter, err := postgres.NewPostgresAdapter(connStr)
	if err != nil {
		slog.Error("Postgres connection error", "err", err)
		os.Exit(1)
	}
	defer pgAdapter.Close()

	apiAdapter, err := postgres.NewApiAdapter(connStr)
	if err != nil {
		slog.Error("API adapter connection error", "err", err)
		os.Exit(1)
	}
	defer apiAdapter.Close()

	redisAdapter := redis.NewRedisAdapter("redis:6379", "", 0)

	updates := make(chan domain.PriceUpdate, 1000)
	modeManager := mode.NewModeManager(updates)
	modeManager.SetMode(ctx, mode.ModeLive)

	service := aggregator.NewServiceCom(redisAdapter, pgAdapter)
	service.StartRedisWorkerPool(ctx, updates, 5)
	go service.StartAggregator(ctx)

	apiService := api.NewService(apiAdapter)
	apiHandler := handler.NewHandler(apiService, modeManager)

	http.HandleFunc("/prices/latest/", apiHandler.Handle)
	http.HandleFunc("/prices/highest/", apiHandler.Highest)
	http.HandleFunc("/prices/lowest/", apiHandler.Lowest)
	http.HandleFunc("/prices/average/", apiHandler.Average)
	http.HandleFunc("/mode/test", apiHandler.SwitchToTestMode)
	http.HandleFunc("/mode/live", apiHandler.SwitchToLiveMode)

	healthHandler := &handler.HealthHandler{
		DB:    apiAdapter,
		Redis: redisAdapter,
	}
	http.Handle("/health", healthHandler)

	go func() {
		slog.Info("Starting HTTP server", "address", ":8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			slog.Error("HTTP server failed", "err", err)
		}
	}()

	slog.Info("Service is running")
	<-ctx.Done()
	slog.Info("Shutting down")
}
