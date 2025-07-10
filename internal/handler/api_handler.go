package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"marketflow/internal/app/api"
	"marketflow/internal/app/mode"
)

type Handler struct {
	Service     *api.APIService
	ModeManager *mode.Manager
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewHandler(service *api.APIService, mm *mode.Manager) *Handler {
	return &Handler{Service: service, ModeManager: mm}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) == 5 {
		h.HandleLatestByExchange(w, r)
	} else {
		h.HandleLatestPrice(w, r)
	}
}

func (h *Handler) Highest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) == 5 {
		if r.URL.Query().Has("period") {
			h.HandleHighestByPeriodByExchange(w, r)
		} else {
			h.HandleHighestByExchange(w, r)
		}
	} else {
		if r.URL.Query().Has("period") {
			h.HandleHighestByPeriod(w, r)
		} else {
			h.HandleHighestPrice(w, r)
		}
	}
}

func (h *Handler) Lowest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) == 5 {
		if r.URL.Query().Has("period") {
			h.HandleLowestByPeriodByExchange(w, r)
		} else {
			h.HandleLowestByExchange(w, r)
		}
	} else {
		if r.URL.Query().Has("period") {
			h.HandleLowestByPeriod(w, r)
		} else {
			h.HandleLowestPrice(w, r)
		}
	}
}

func (h *Handler) Average(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) == 5 {
		if r.URL.Query().Has("period") {
			h.HandleAvgByPeriodByExchange(w, r)
		} else {
			h.HandleAvgByExchange(w, r)
		}
	} else {
		h.HandleAvgPrice(w, r)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	slog.Warn("Returning error", "status", status, "message", message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
