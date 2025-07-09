package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func (h *Handler) HandleLatestPrice(w http.ResponseWriter, r *http.Request) {
	symbol := strings.TrimPrefix(r.URL.Path, "/prices/latest/")

	data, err := h.Service.GetAggregatedPriceForSymbol(symbol)
	if err != nil {
		http.Error(w, "Error fetching data for symbol", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) HandleLatestByExchange(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/prices/latest/")
	log.Println("Path:", path)

	data, err := h.Service.GetAggregatedPriceForExchange(path)
	if err != nil {
		http.Error(w, "Error fetching data for symbol", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
