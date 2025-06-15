package handlers

import (
	"encoding/json"
	"marketflow/internal/adapters/postgres"
	"net/http"
)

type AggregatedHandler struct {
	PG *postgres.Adapter
}

type AggregatedResponse struct {
	Pair      string  `json:"pair"`
	Exchange  string  `json:"exchange"`
	Timestamp string  `json:"timestamp"`
	Avg       float64 `json:"avg"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
}

func (h *AggregatedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := h.PG.GetLastAggregatedPrices()
	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(data)
}
