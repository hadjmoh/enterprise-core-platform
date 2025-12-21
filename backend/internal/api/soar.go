package api

import (
	"encoding/json"
	"enterprise-core/backend/internal/security"
	"net/http"
	"strings"
)

type SOARHandler struct {
	orchestrator *security.Orchestrator
}

func NewSOARHandler(orch *security.Orchestrator) *SOARHandler {
	return &SOARHandler{orchestrator: orch}
}

func (h *SOARHandler) GetPendingActions(w http.ResponseWriter, r *http.Request) {
	pending := h.orchestrator.GetPending()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pending)
}

func (h *SOARHandler) ApproveAction(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		http.Error(w, "Action ID required", http.StatusBadRequest)
		return
	}
	actionID := parts[4]

	err := h.orchestrator.ApproveAction(r.Context(), actionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "executed", "id": actionID})
}

func (h *SOARHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	history := h.orchestrator.GetHistory()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
