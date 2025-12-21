package api

import (
	"encoding/json"
	"net/http"
)

func (rt *Router) handleGetPolicies(w http.ResponseWriter, r *http.Request) {
	// Retrieve active policies from the registry
	// For now, we return all policies (in a real app, this might be filtered by user role)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rt.policies)
}
