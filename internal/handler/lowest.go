package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) HandleLowestPrice(w http.ResponseWriter, r *http.Request) {
	symbol := strings.TrimPrefix(r.URL.Path, "/prices/lowest/")

	data, err := h.Service.GetLowestBySymbol(symbol)
	if err != nil {
		http.Error(w, "Error fetching data for symbol", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) HandleLowestByExchange(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/lowest/")
	log.Println("Path:", path)

	data, err := h.Service.GetLowestByExchange(path)
	if err != nil {
		http.Error(w, "Error fetching data for symbol", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) HandleLowestByPeriod(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/lowest/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		http.Error(w, "Missing symbol", http.StatusBadRequest)
		return
	}
	symbol := parts[0]

	periodStr := r.URL.Query().Get("period")
	if periodStr == "" {
		http.Error(w, "Missing 'period' query parameter", http.StatusBadRequest)
		return
	}

	duration, err := time.ParseDuration(periodStr)
	if err != nil {
		http.Error(w, "Invalid duration format", http.StatusBadRequest)
		return
	}

	result, err := h.Service.GetLowestByPeriod(symbol, duration)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) HandleLowestByPeriodByExchange(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/lowest/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		http.Error(w, "Missing symbol", http.StatusBadRequest)
		return
	}
	exchange := parts[0]
	symbol := parts[1]

	periodStr := r.URL.Query().Get("period")
	if periodStr == "" {
		http.Error(w, "Missing 'period' query parameter", http.StatusBadRequest)
		return
	}

	duration, err := time.ParseDuration(periodStr)
	if err != nil {
		http.Error(w, "Invalid duration format", http.StatusBadRequest)
		return
	}

	result, err := h.Service.QueryLowestSinceByExchange(exchange, symbol, duration)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
