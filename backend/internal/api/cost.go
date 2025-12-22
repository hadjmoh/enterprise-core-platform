package api

import (
	"encoding/json"
	"net/http"
	"time"

	"enterprise-core/backend/internal/query/cost"
)

// EstimateRequest is the input for cost estimation
type EstimateRequest struct {
	Query     string `json:"query"`
	TimeStart string `json:"time_start"`
	TimeEnd   string `json:"time_end"`
}

// EstimateResponse is the output for cost estimation
type EstimateResponse struct {
	Estimate *cost.QueryCostEstimate `json:"estimate"`
	Allowed  bool                    `json:"allowed"`
	Reason   string                  `json:"reason"`
}

// handleEstimateCost handles the query cost estimation endpoint
func (rt *Router) handleEstimateCost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EstimateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Default time range if not provided
	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	if req.TimeStart != "" {
		if t, err := time.Parse(time.RFC3339, req.TimeStart); err == nil {
			start = t
		}
	}
	if req.TimeEnd != "" {
		if t, err := time.Parse(time.RFC3339, req.TimeEnd); err == nil {
			end = t
		}
	}

	// Extract user role from token (already validated by middleware presumably)
	// For prototype, we'll extract it again or rely on context
	// In a real middleware, this would be in the context
	userRole := rt.extractRole(r)

	// Estimate and check
	estimate, allowed, reason, err := rt.costEngine.EstimateAndCheck(req.Query, start, end, userRole)
	if err != nil {
		rt.logger.Error("Failed to estimate query cost", err)
		http.Error(w, "Failed to estimate cost", http.StatusInternalServerError)
		return
	}

	resp := EstimateResponse{
		Estimate: estimate,
		Allowed:  allowed,
		Reason:   reason,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (rt *Router) extractRole(r *http.Request) string {
	// Simple extraction for now, matching search.go logic
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "guest"
	}
	// TODO: Use the auth service to validate and extract role
	// For now, assume guest if not easily parseable or mock logic
	return "analyst" // Default to analyst for authenticated users in dev
}
