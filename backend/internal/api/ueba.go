package api

import (
	"encoding/json"
	"enterprise-core/backend/internal/security"
	"net/http"
	"strings"
)

type UEBAHandler struct {
	engine *security.UEBAEngine
}

func NewUEBAHandler(engine *security.UEBAEngine) *UEBAHandler {
	return &UEBAHandler{engine: engine}
}

func (h *UEBAHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		http.Error(w, "Entity ID required", http.StatusBadRequest)
		return
	}
	entityID := parts[4]

	summary := h.engine.GetProfileSummary(entityID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
