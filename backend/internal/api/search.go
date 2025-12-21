package api

import (
	"context"
	"encoding/json"
	"enterprise-core/backend/internal/auth"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SearchMetadata represents the metadata sent at the start of a streaming search
type SearchMetadata struct {
	Type          string `json:"_type"`
	Span          int64  `json:"_span,omitempty"`
	GeneratedAt   string `json:"_generated_at"`
	PipelineID    int    `json:"_pipeline_id"`
	IsRealtime    bool   `json:"_is_realtime"`
}

// SearchEvent wraps the data with a running count for a batch
type SearchEvent struct {
	Count int           `json:"_total_count"`
	Batch []interface{} `json:"batch"`
}

// SearchError defines a structured error for the UI
type SearchError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Fatal   bool   `json:"fatal"`
}

// handleSearch handles streaming SPL queries via SSE
func (rt *Router) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// Extract Role from Token
	authHeader := r.Header.Get("Authorization")
	token := ""
	if strings.HasPrefix(authHeader, "Splunk ") {
		token = strings.TrimPrefix(authHeader, "Splunk ")
	}

	role := "guest"
	if token != "" {
		claims, err := rt.auth.ValidateToken(token)
		if err == nil && claims.Role != "" {
			role = claims.Role
		}
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Context for query execution
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()
	
	startTime := time.Now()

	// Start the streaming execution
	results, meta, errs, err := rt.dispatcher.ExecuteStream(ctx, query, role, nil)
	if err != nil {
		rt.sendSSE(w, "error", SearchError{Code: "AUTH_OR_COMPILATION_ERROR", Message: err.Error(), Fatal: true})
		return
	}

	// 1. Send Metadata event
	rt.sendSSE(w, "metadata", SearchMetadata{
		Type:        meta.Type,
		Span:        int64(meta.Span.Seconds()),
		GeneratedAt: time.Now().Format(time.RFC3339),
		PipelineID:  meta.Version,
		IsRealtime:  meta.Type == "search" && strings.Contains(query, "EARLIEST=rt"), // Basic heuristic
	})

	totalCount := 0
	pingTicker := time.NewTicker(15 * time.Second)
	defer pingTicker.Stop()

	batchTicker := time.NewTicker(200 * time.Millisecond)
	defer batchTicker.Stop()

	var batch []interface{}
	sendBatch := func() {
		if len(batch) == 0 {
			return
		}
		rt.sendSSE(w, "event", SearchEvent{
			Count: totalCount,
			Batch: batch,
		})
		batch = nil
	}

	// 2. Stream results
	for {
		select {
		case <-ctx.Done():
			sendBatch()
			if ctx.Err() == context.DeadlineExceeded {
				rt.sendSSE(w, "error", SearchError{Code: "TIMEOUT", Message: "Query timed out", Fatal: true})
			}
			return
		case <-pingTicker.C:
			// Send keep-alive ping
			rt.sendSSE(w, "ping", time.Now().Unix())
		case <-batchTicker.C:
			sendBatch()
		case err, ok := <-errs:
			if ok && err != nil {
				rt.sendSSE(w, "error", SearchError{Code: "PIPELINE_ERROR", Message: err.Error(), Fatal: false})
			}
			if !ok {
				errs = nil
			}
		case ev, ok := <-results:
			if !ok {
				sendBatch()
				// Search complete
				rt.sendSSE(w, "done", map[string]interface{}{
					"status":      "complete",
					"total_count": totalCount,
					"time_taken":  fmt.Sprintf("%.3fs", time.Since(startTime).Seconds()),
				})
				return
			}
			totalCount++
			batch = append(batch, ev)
			if len(batch) >= 100 {
				sendBatch()
			}
		}
	}
}

// sendSSE is a helper to format and send SSE events
func (rt *Router) sendSSE(w http.ResponseWriter, event string, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		rt.logger.Error("Failed to marshal SSE data", err)
		return
	}

	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(payload))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
