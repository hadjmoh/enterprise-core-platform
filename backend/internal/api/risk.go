package api

import (
	"encoding/json"
	"enterprise-core/backend/internal/security/risk"
	"net/http"
	"strings"
)

type RiskHandler struct {
	engine *risk.RiskEngine
}

func NewRiskHandler(engine *risk.RiskEngine) *RiskHandler {
	return &RiskHandler{engine: engine}
}

func (h *RiskHandler) GetTopEntities(w http.ResponseWriter, r *http.Request) {
	entities := h.engine.GetTopEntities(10)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}

func (h *RiskHandler) GetEntityRisk(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		http.Error(w, "Entity ID required", http.StatusBadRequest)
		return
	}
	entityID := parts[4]

	score, ok := h.engine.GetEntityScore(entityID)
	if !ok {
		http.Error(w, "Entity not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(score)
}
