package handlers

import (
	"encoding/json"
	"marketflow/internal/app/impl"
	"net/http"
	"strings"
)

type AggregatedHandler struct {
	Service *impl.LatestService
}

func (h *AggregatedHandler) Handle(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) > 3 {
		h.HandleLatestByExchange(w, r)
	} else {
		h.HandleLatestPrice(w, r)
	}
}

func (h *AggregatedHandler) HandleLatestPrice(w http.ResponseWriter, r *http.Request) {
	symbol := strings.TrimPrefix(r.URL.Path, "/prices/latest/")

	data, err := h.Service.GetAggregatedPriceForSymbol(symbol)
	if err != nil {
		http.Error(w, "Error fetching data for symbol", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *AggregatedHandler) HandleLatestByExchange(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	data, err := h.Service.GetAggregatedPriceForExchange(path)
	if err != nil {
		http.Error(w, "Error fetching data for symbol", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
