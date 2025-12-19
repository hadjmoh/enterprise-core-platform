package api

import (
	"enterprise-core/backend/internal/middleware"
	"enterprise-core/backend/internal/pipeline"
	"enterprise-core/backend/internal/query"
	"enterprise-core/backend/pkg/logger"
	"encoding/json"
	"net/http"
	"time"
)

type Router struct {
	mux        *http.ServeMux
	pipeline   *pipeline.IngestionPipeline
	dispatcher *query.Dispatcher
	logger     *logger.Logger
}

func NewRouter(pipe *pipeline.IngestionPipeline, disp *query.Dispatcher, l *logger.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	router := &Router{
		mux:        mux,
		pipeline:   pipe,
		dispatcher: disp,
		logger:     l,
	}

	// Health Check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"status": "up",
			"time":   time.Now().Format(time.RFC3339),
			"service": "enterprise-backend",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// HEC Endpoint
	mux.HandleFunc("POST /services/collector/event", middleware.HECAuthMiddleware(router.handleHEC))
	
	// WebSocket Endpoint
	mux.HandleFunc("/services/collector/ws", router.handleWebSocket)

    // SOAP Endpoint
    mux.HandleFunc("POST /services/collector/soap", router.handleSOAP)

	// Search Endpoint (Streaming SSE)
	mux.HandleFunc("GET /services/search", router.handleSearch)

	return mux
}
