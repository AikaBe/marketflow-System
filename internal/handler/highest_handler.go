package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) HandleHighestPrice(w http.ResponseWriter, r *http.Request) {
	symbol := strings.TrimPrefix(r.URL.Path, "/prices/highest/")
	slog.Info("HandleHighestPrice called", "symbol", symbol)

	data, err := h.Service.GetHighestBySymbol(symbol)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	slog.Info("GetHighestBySymbol success", "symbol", symbol)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) HandleHighestByExchange(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/highest/")
	slog.Info("HandleHighestByExchange called", "path", path)

	data, err := h.Service.GetHighestByExchange(path)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	slog.Info("GetHighestByExchange success", "path", path)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) HandleHighestByPeriod(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/highest/")
	slog.Info("HandleHighestByPeriod called", "path", path)

	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		writeJSONError(w, http.StatusBadRequest, "Missing symbol")
		return
	}
	symbol := parts[0]

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

	result, err := h.Service.GetHighestByPeriod(symbol, duration)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	slog.Info("GetHighestByPeriod success", "symbol", symbol, "period", duration)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) HandleHighestByPeriodByExchange(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/highest/")
	slog.Info("HandleHighestByPeriodByExchange called", "path", path)

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		writeJSONError(w, http.StatusBadRequest, "Missing exchange or symbol")
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

	result, err := h.Service.QueryHighestSinceByExchange(exchange, symbol, duration)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	slog.Info("QueryHighestSinceByExchange success", "exchange", exchange, "symbol", symbol, "period", duration)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
