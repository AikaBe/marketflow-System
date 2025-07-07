package handlers

import (
	"encoding/json"
	"marketflow/internal/app"
	"net/http"
	"strings"
)

type AggregatedHandler struct {
	Service *app.Latest
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
