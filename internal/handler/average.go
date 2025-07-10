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
		writeJSONError(w, http.StatusBadRequest, "Symbol is required")
		return
	}

	data, err := h.Service.GetAvgBySymbol(symbol)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if data == nil {
		writeJSONError(w, http.StatusNotFound, "No data found")
		return
	}

	slog.Info("Responded with average price", "symbol", symbol, "data", data)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) HandleAvgByExchange(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/average/")
	slog.Info("HandleAvgByExchange called", "path", path)

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		writeJSONError(w, http.StatusBadRequest, "Invalid path. Format: /prices/average/{exchange}/{symbol}")
		return
	}

	data, err := h.Service.GetAvgByExchange(path)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if data == nil {
		writeJSONError(w, http.StatusNotFound, "No data found")
		return
	}

	slog.Info("Responded with average price by exchange", "data", data)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) HandleAvgByPeriodByExchange(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/average/")
	slog.Info("HandleAvgByPeriodByExchange called", "path", path)

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		writeJSONError(w, http.StatusBadRequest, "Invalid path. Format: /prices/average/{exchange}/{symbol}")
		return
	}

	exchange := parts[0]
	symbol := parts[1]

	periodStr := r.URL.Query().Get("period")
	if periodStr == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing 'period' query parameter")
		return
	}

	duration, err := time.ParseDuration(periodStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid period format: "+err.Error())
		return
	}

	result, err := h.Service.QueryAvgSinceByExchange(exchange, symbol, duration)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if result == nil {
		writeJSONError(w, http.StatusNotFound, "No data found")
		return
	}

	slog.Info("Responded with average price by period", "exchange", exchange, "symbol", symbol, "period", duration)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
