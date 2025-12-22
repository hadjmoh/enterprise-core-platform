package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"enterprise-core/backend/internal/analytics"
)

// handleGetFeatures retrieves features for an entity
func (rt *Router) handleGetFeatures(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// path: /api/analytics/features/:entityId
	// Simple path extraction for PoC
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	entityID := parts[4]

	features := rt.featureStore.GetAllFeatures(entityID)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(features)
}

// handleGetAnomalies runs detection and returns anomalies (Simulation)
func (rt *Router) handleGetAnomalies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For PoC, we run detection on key metrics
	// In production, this would return pre-calculated anomalies from the DB
	
	anomalies := []analytics.AnomalyScore{}

	// Check login counts
	loginAnomalies := rt.anomalyDetector.DetectPopulationAnomalies(rt.featureStore, "login_count_1h")
	anomalies = append(anomalies, loginAnomalies...)

	// Check bytes out
	bytesAnomalies := rt.anomalyDetector.DetectPopulationAnomalies(rt.featureStore, "bytes_out_1h")
	anomalies = append(anomalies, bytesAnomalies...)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(anomalies)
}
