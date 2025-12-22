package api

import (
	"encoding/json"
	"enterprise-core/backend/internal/simulation"
	"net/http"
	"time"
)

type SimulationHandler struct {
	engine *simulation.Engine
}

func NewSimulationHandler(engine *simulation.Engine) *SimulationHandler {
	return &SimulationHandler{engine: engine}
}

func (h *SimulationHandler) StartSimulation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	from, _ := time.Parse(time.RFC3339, req.From)
	to, _ := time.Parse(time.RFC3339, req.To)
	sessionID := "sim-" + time.Now().Format("20060102-150405")

	// Run in background for real implementation, but for PoC we wait
	result, err := h.engine.RunReplay(sessionID, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *SimulationHandler) GetResult(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing simulation ID", http.StatusBadRequest)
		return
	}

	result, ok := h.engine.GetResult(id)
	if !ok {
		http.Error(w, "Simulation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
