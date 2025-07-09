package handler

import (
	"context"
	"encoding/json"
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
	dbStatus := "ok"
	if err := h.DB.Ping(); err != nil {
		dbStatus = "disconnected"
	}

	redisStatus := "ok"
	if err := h.Redis.Ping(context.Background()); err != nil {
		redisStatus = "unavailable"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"db":        dbStatus,
		"redis":     redisStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
