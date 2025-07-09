package handler

import (
	"marketflow/internal/app/mode"
	"net/http"
)

func (h *Handler) SwitchToTestMode(w http.ResponseWriter, r *http.Request) {
	h.ModeManager.SetMode(r.Context(), mode.ModeTest)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Switched to Test Mode"))
}

func (h *Handler) SwitchToLiveMode(w http.ResponseWriter, r *http.Request) {
	h.ModeManager.SetMode(r.Context(), mode.ModeLive)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Switched to Live Mode"))
}
