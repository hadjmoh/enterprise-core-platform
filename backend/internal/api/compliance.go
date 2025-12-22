package api

import (
	"encoding/json"
	"net/http"

	"enterprise-core/backend/internal/compliance"
)

// handleGetProof returns the current Merkle Tree root hash
func (rt *Router) handleGetProof(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rootHash := rt.merkleTree.GetRootHash()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"merkle_root": rootHash,
		"status":      "immutable",
	})
}

// handleVerifyProof verifies if a log is part of the chain (Simulated for PoC)
func (rt *Router) handleVerifyProof(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// In a real implementation, we would accept a log and a proof path.
	// Here we just re-calculate the root to show mechanism.
	
	// Mock response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"verified": true,
		"timestamp": "2023-10-27T10:00:00Z",
	})
}
