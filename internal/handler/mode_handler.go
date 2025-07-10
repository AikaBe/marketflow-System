package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"marketflow/internal/app/mode"
)

type MessageResponse struct {
	Message string `json:"message"`
}

func (h *Handler) SwitchToTestMode(w http.ResponseWriter, r *http.Request) {
	slog.Info("SwitchToTestMode called")

	err := h.ModeManager.SetMode(r.Context(), mode.ModeTest)
	if err != nil {
		slog.Warn("SwitchToTestMode failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	slog.Info("Mode switched to TEST")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MessageResponse{Message: "Switched to Test Mode"})
}

func (h *Handler) SwitchToLiveMode(w http.ResponseWriter, r *http.Request) {
	slog.Info("SwitchToLiveMode called")

	err := h.ModeManager.SetMode(r.Context(), mode.ModeLive)
	if err != nil {
		slog.Warn("SwitchToLiveMode failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	slog.Info("Mode switched to LIVE")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MessageResponse{Message: "Switched to Live Mode"})
}
