package handler

import (
	"log"
	"marketflow/internal/app/api"
	"marketflow/internal/app/mode"
	"net/http"
	"strings"
)

type Handler struct {
	Service     *api.APIService
	ModeManager *mode.Manager
}

func NewHandler(service *api.APIService, mm *mode.Manager) *Handler {
	return &Handler{Service: service, ModeManager: mm}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	log.Println("Len", len(parts))

	log.Println("Parts:", parts)

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