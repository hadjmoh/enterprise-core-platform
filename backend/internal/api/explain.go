package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"enterprise-core/backend/internal/analytics"
)

// handleExplainAnomaly generates an explanation for a specific anomaly
func (rt *Router) handleExplainAnomaly(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// path: /api/explain/anomaly/:entityId/:anomalyType/:value
	// Simplifying for PoC: /api/explain/anomaly?entity=x&type=y&value=z
	entityID := r.URL.Query().Get("entity")
	anomalyType := r.URL.Query().Get("type")
	// quick parse float
	var value float64
	fmt.Sscanf(r.URL.Query().Get("value"), "%f", &value)

	if entityID == "" || anomalyType == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	explanation := rt.explainerEngine.ExplainAnomaly(entityID, anomalyType, value)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(explanation)
}

// handleFeedback submits analyst feedback
func (rt *Router) handleFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var feedback analytics.Feedback
	if err := json.NewDecoder(r.Body).Decode(&feedback); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := rt.feedbackStore.SubmitFeedback(feedback); err != nil {
		rt.logger.Error("Failed to submit feedback", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
