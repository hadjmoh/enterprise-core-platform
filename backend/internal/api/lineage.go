package api

import (
	"encoding/json"
	"net/http"
)

// handleGetEventLineage returns the full chain-of-custody for a specific event
// In a real system, this would fetch from storage by ID. 
// For PoC, we return a mock or search the recent buffer if needed.
func (rt *Router) handleGetEventLineage(w http.ResponseWriter, r *http.Request) {
	eventID := r.URL.Query().Get("id")
	if eventID == "" {
		http.Error(w, "Missing event ID", http.StatusBadRequest)
		return
	}

	// This is a PoC: In real life, we'd query the StorageEngine or an Index for this ID.
	// Here we return a mock lineage to demonstrate the frontend.
	mockLineage := map[string]interface{}{
		"event_id": eventID,
		"steps": []map[string]interface{}{
			{
				"stage":     "ingestion",
				"node_id":    "ingest-node-01",
				"timestamp":  "2023-11-20T10:00:00Z",
				"action":    "receival",
				"signature": "valid-hmac-sha256-abc",
			},
			{
				"stage":     "enrichment",
				"node_id":    "worker-node-04",
				"timestamp":  "2023-11-20T10:00:01Z",
				"action":    "geo_lookup",
				"signature": "valid-hmac-sha256-def",
			},
			{
				"stage":     "correlation",
				"node_id":    "correlator-02",
				"timestamp":  "2023-11-20T10:00:02Z",
				"action":    "match",
				"signature": "valid-hmac-sha256-ghi",
			},
		},
		"verified": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockLineage)
}

// handleGetTrustGraph returns the full trust graph for visualization
func (rt *Router) handleGetTrustGraph(w http.ResponseWriter, r *http.Request) {
	if rt.trustGraph == nil {
		http.Error(w, "Trust graph not initialized", http.StatusServiceUnavailable)
		return
	}

	nodes := rt.trustGraph.GetNodes()
	edges := rt.trustGraph.GetEdges()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"nodes": nodes,
		"edges": edges,
	})
}
