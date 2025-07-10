package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) HandleLowestPrice(w http.ResponseWriter, r *http.Request) {
	symbol := strings.TrimPrefix(r.URL.Path, "/prices/lowest/")
	slog.Info("HandleLowestPrice called", "symbol", symbol)

	data, err := h.Service.GetLowestBySymbol(symbol)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	slog.Info("GetLowestBySymbol success", "symbol", symbol)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) HandleLowestByExchange(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/lowest/")
	slog.Info("HandleLowestByExchange called", "path", path)

	data, err := h.Service.GetLowestByExchange(path)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	slog.Info("GetLowestByExchange success", "path", path)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) HandleLowestByPeriod(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/lowest/")
	slog.Info("HandleLowestByPeriod called", "path", path)

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

	result, err := h.Service.GetLowestByPeriod(symbol, duration)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	slog.Info("GetLowestByPeriod success", "symbol", symbol, "duration", duration)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (h *Handler) HandleLowestByPeriodByExchange(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/lowest/")
	slog.Info("HandleLowestByPeriodByExchange called", "path", path)

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

	result, err := h.Service.QueryLowestSinceByExchange(exchange, symbol, duration)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	slog.Info("QueryLowestSinceByExchange success", "exchange", exchange, "symbol", symbol, "duration", duration)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
