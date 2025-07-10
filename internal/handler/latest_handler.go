package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

func (h *Handler) HandleLatestPrice(w http.ResponseWriter, r *http.Request) {
	symbol := strings.TrimPrefix(r.URL.Path, "/prices/latest/")
	slog.Info("HandleLatestPrice called", "symbol", symbol)

	data, err := h.Service.GetAggregatedPriceForSymbol(symbol)
	if err != nil {
		slog.Warn("HandleLatestPrice failed", "symbol", symbol, "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	slog.Info("HandleLatestPrice success", "symbol", symbol)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) HandleLatestByExchange(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/latest/")
	slog.Info("HandleLatestByExchange called", "path", path)

	data, err := h.Service.GetAggregatedPriceForExchange(path)
	if err != nil {
		slog.Warn("HandleLatestPrice failed", "path", path, "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	slog.Info("HandleLatestPrice success", "path", path)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
