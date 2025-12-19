package api

import (
	"encoding/json"
	"enterprise-core/backend/internal/buffer"
	"net/http"
	"time"
)

// handleHEC processes incoming HTTP events
func (rt *Router) handleHEC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		rt.logger.Error("Failed to decode HEC event", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(HECResponse{Text: "Invalid Data Format", Code: 400})
		return
	}

	// Set default time if missing
	if event.Time == 0 {
		event.Time = time.Now().Unix()
	}

	// Log reception and process through pipeline
	rt.logger.Info("Received HEC Event", 
		"host", event.Host, 
		"sourcetype", event.SourceType,
		"payload_size", r.ContentLength,
	)

	// Process through pipeline
	if err := rt.pipeline.Process(buffer.Event{
		Timestamp: time.Unix(event.Time, 0).Format(time.RFC3339),
		Source:    event.Source,
		Data:      event.Fields, // Map HEC fields to internal event data
	}); err != nil {
		rt.logger.Error("Pipeline processing failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(HECResponse{Text: "Internal Processing Error", Code: 500})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HECResponse{Text: "Success", Code: 0})
}
