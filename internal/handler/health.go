package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type HealthHandler struct {
	DB    DBChecker
	Redis RedisChecker
}

type DBChecker interface {
	Ping() error
}

type RedisChecker interface {
	Ping(ctx context.Context) error
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Info("Health check started", "method", r.Method, "url", r.URL.Path)

	dbStatus := "ok"
	if err := h.DB.Ping(); err != nil {
		dbStatus = "disconnected"
		slog.Warn("Database ping failed", "error", err)
	} else {
		slog.Info("Database ping successful")
	}

	redisStatus := "ok"
	if err := h.Redis.Ping(context.Background()); err != nil {
		redisStatus = "unavailable"
		slog.Warn("Redis ping failed", "error", err)
	} else {
		slog.Info("Redis ping successful")
	}

	response := map[string]interface{}{
		"status":    "ok",
		"db":        dbStatus,
		"redis":     redisStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to write health check response", "error", err)
	}

	slog.Info("Health check completed", "db", dbStatus, "redis", redisStatus)
}
