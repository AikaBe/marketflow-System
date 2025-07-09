package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) HandleAvgPrice(w http.ResponseWriter, r *http.Request) {
	symbol := strings.TrimPrefix(r.URL.Path, "/prices/average/")
	slog.Info("HandleAvgPrice called", "symbol", symbol)

	if symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		slog.Warn("Symbol is missing")
		return
	}

	data, err := h.Service.GetAvgBySymbol(symbol)
	if err != nil {
		slog.Error("Failed to get average price", "symbol", symbol, "error", err)
		http.Error(w, "Error fetching data for symbol", http.StatusInternalServerError)
		return
	}
	if data == nil {
		slog.Warn("No data found for symbol", "symbol", symbol)
		http.Error(w, "No data found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
	slog.Info("Responded with average price", "data", data)
}

func (h *Handler) HandleAvgByExchange(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/average/")
	slog.Info("HandleAvgByExchange called", "path", path)

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid path. Format: /prices/average/{exchange}/{symbol}", http.StatusBadRequest)
		slog.Warn("Invalid path for average by exchange", "path", path)
		return
	}

	data, err := h.Service.GetAvgByExchange(path)
	if err != nil {
		slog.Error("Failed to get average by exchange", "path", path, "error", err)
		http.Error(w, "Error fetching data for exchange and symbol", http.StatusInternalServerError)
		return
	}
	if data == nil {
		slog.Warn("No data found", "path", path)
		http.Error(w, "No data found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
	slog.Info("Responded with average price by exchange", "data", data)
}

func (h *Handler) HandleAvgByPeriodByExchange(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/average/")
	slog.Info("HandleAvgByPeriodByExchange called", "path", path)

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid path. Format: /prices/average/{exchange}/{symbol}", http.StatusBadRequest)
		slog.Warn("Invalid path format", "path", path)
		return
	}

	exchange := parts[0]
	symbol := parts[1]

	periodStr := r.URL.Query().Get("period")
	if periodStr == "" {
		http.Error(w, "Missing 'period' query parameter", http.StatusBadRequest)
		slog.Warn("Missing 'period' query parameter")
		return
	}

	duration, err := time.ParseDuration(periodStr)
	if err != nil {
		http.Error(w, "Invalid duration format. Example: 1m, 2h", http.StatusBadRequest)
		slog.Warn("Invalid duration format", "period", periodStr, "error", err)
		return
	}

	result, err := h.Service.QueryAvgSinceByExchange(exchange, symbol, duration)
	if err != nil {
		slog.Error("Failed to query avg by period and exchange", "exchange", exchange, "symbol", symbol, "period", duration, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if result == nil {
		slog.Warn("No data found for period", "exchange", exchange, "symbol", symbol, "period", duration)
		http.Error(w, "No data found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
	slog.Info("Responded with average price by period", "result", result)
}
