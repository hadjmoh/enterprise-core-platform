package api

import (
	"encoding/json"
	"net/http"

	"enterprise-core/backend/internal/query/pilot"
)

// SuggestRequest is the input for query suggestions
type SuggestRequest struct {
	Query string `json:"query"`
}

// SuggestResponse is the output for query suggestions
type SuggestResponse struct {
	Suggestion *pilot.Suggestion `json:"suggestion"`
	Error      string            `json:"error,omitempty"`
}

// handlePilotSuggest handles the pilot suggestion endpoint
func (rt *Router) handlePilotSuggest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SuggestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract user role (mock for now, similar to cost handler)
	userRole := rt.extractRole(r)

	suggestion, err := rt.pilotEngine.Suggest(req.Query, userRole)
	if err != nil {
		rt.logger.Error("Pilot suggestion failed", err)
		// Return 400 if it's just translation failure, 500 otherwise
		// Simplification: just return error in JSON
		json.NewEncoder(w).Encode(SuggestResponse{Error: err.Error()})
		return
	}

	resp := SuggestResponse{
		Suggestion: suggestion,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
