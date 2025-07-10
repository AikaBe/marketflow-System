package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"marketflow/internal/adapters/postgres"
	"marketflow/internal/adapters/redis"
	"marketflow/internal/app/aggregator"
	"marketflow/internal/app/api"
	"marketflow/internal/app/mode"
	"marketflow/internal/domain"
	"marketflow/internal/handler"

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
	defer redisAdapter.Close()

	updates := make(chan domain.PriceUpdate, 1000)
	modeManager := mode.NewModeManager(updates)
	modeManager.SetMode(ctx, mode.ModeLive)

	service := aggregator.NewServiceCom(redisAdapter, pgAdapter)
	service.StartRedisWorkerPool(ctx, updates, 5)
	go service.StartAggregator(ctx)

	apiService := api.NewService(apiAdapter)
	apiHandler := handler.NewHandler(apiService, modeManager)

	mux := http.NewServeMux()
	mux.HandleFunc("/prices/latest/", apiHandler.Handle)
	mux.HandleFunc("/prices/highest/", apiHandler.Highest)
	mux.HandleFunc("/prices/lowest/", apiHandler.Lowest)
	mux.HandleFunc("/prices/average/", apiHandler.Average)
	mux.HandleFunc("/mode/test", apiHandler.SwitchToTestMode)
	mux.HandleFunc("/mode/live", apiHandler.SwitchToLiveMode)

	healthHandler := &handler.HealthHandler{
		DB:    apiAdapter,
		Redis: redisAdapter,
	}
	mux.Handle("/health", healthHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		slog.Info("Starting HTTP server", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "err", err)
			cancel()
		}
	}()

	slog.Info("Service is running")

	<-ctx.Done()
	slog.Info("Shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful shutdown failed", "err", err)
	} else {
		slog.Info("HTTP server shutdown complete")
	}

	slog.Info("Application shutdown complete")
}
