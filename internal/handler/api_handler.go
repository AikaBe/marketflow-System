package handler

import (
	"log"
	"marketflow/internal/app/app_impl"
	"net/http"
	"strings"
)

type Handler struct {
	Service *app_impl.APIService
}

func NewHandler(service *app_impl.APIService) *Handler {
	return &Handler{Service: service}
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
