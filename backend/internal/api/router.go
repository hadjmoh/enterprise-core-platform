package api

import (
	"enterprise-core/backend/internal/auth"
	"enterprise-core/backend/internal/middleware"
	"enterprise-core/backend/internal/pipeline"
	"enterprise-core/backend/internal/query"
	"enterprise-core/backend/internal/query/cost"
	"enterprise-core/backend/internal/query/pilot"
	"enterprise-core/backend/internal/analytics"
	"enterprise-core/backend/internal/security"
	"enterprise-core/backend/internal/security/risk"
	"enterprise-core/backend/internal/compliance"
	"enterprise-core/backend/pkg/logger"
	"encoding/json"
	"net/http"
	"time"
)

type Router struct {
	mux        *http.ServeMux
	pipeline   *pipeline.IngestionPipeline
	dispatcher *query.Dispatcher
	risk       *RiskHandler
	ueba       *UEBAHandler
	soar       *SOARHandler
	policies   *compliance.PolicyRegistry
	hunting    *HuntingHandler
	mitre      *MitreHandler
	costEngine *cost.PolicyEngine
	pilotEngine *pilot.PilotEngine
	featureStore *analytics.FeatureStore
	anomalyDetector *analytics.AnomalyDetector
	explainerEngine *analytics.ExplainerEngine
	feedbackStore *analytics.FeedbackStore
	merkleTree *compliance.MerkleTree
	logger     *logger.Logger
}

func NewRouter(pipe *pipeline.IngestionPipeline, disp *query.Dispatcher, authSvc auth.Service, riskEngine *risk.RiskEngine, uebaEngine *security.UEBAEngine, soarOrch *security.Orchestrator, policies *compliance.PolicyRegistry, huntMgr *security.HuntingManager, mitreMgr *security.MitreManager, costEngine *cost.PolicyEngine, pilotEngine *pilot.PilotEngine, featureStore *analytics.FeatureStore, anomalyDetector *analytics.AnomalyDetector, explainerEngine *analytics.ExplainerEngine, feedbackStore *analytics.FeedbackStore, merkleTree *compliance.MerkleTree, l *logger.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	router := &Router{
		mux:        mux,
		pipeline:   pipe,
		dispatcher: disp,
		risk:       NewRiskHandler(riskEngine),
		ueba:       NewUEBAHandler(uebaEngine),
		soar:       NewSOARHandler(soarOrch),
		policies:   policies,
		hunting:    NewHuntingHandler(huntMgr),
		mitre:      NewMitreHandler(mitreMgr),
		costEngine: costEngine,
		pilotEngine: pilotEngine,
		featureStore: featureStore,
		anomalyDetector: anomalyDetector,
		explainerEngine: explainerEngine,
		feedbackStore: feedbackStore,
		merkleTree: merkleTree,
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
	mux.HandleFunc("POST /api/search/estimate", router.handleEstimateCost)

	// Pilot Endpoint
	mux.HandleFunc("POST /api/pilot/suggest", router.handlePilotSuggest)

	// Risk Scoring Endpoints
	mux.HandleFunc("GET /api/v1/risk/top-entities", router.risk.GetTopEntities)
	mux.HandleFunc("GET /api/v1/risk/entity/", router.risk.GetEntityRisk)

	// UEBA Endpoints
	mux.HandleFunc("GET /api/v1/ueba/profile/", router.ueba.GetProfile)

	// SOAR Endpoints
	mux.HandleFunc("GET /api/v1/soar/pending", router.soar.GetPendingActions)
	mux.HandleFunc("POST /api/v1/soar/approve/", router.soar.ApproveAction)
	mux.HandleFunc("GET /api/v1/soar/history", router.soar.GetHistory)

	// Privacy Endpoints
	mux.HandleFunc("GET /api/v1/privacy/policies", router.handleGetPolicies)

	// Hunting Endpoints
	mux.HandleFunc("GET /api/v1/hunting/hunts", router.hunting.GetHunts)
	mux.HandleFunc("POST /api/v1/hunting/hunts", router.hunting.SaveHunt)
	mux.HandleFunc("GET /api/v1/hunting/notebooks", router.hunting.GetNotebooks)
	mux.HandleFunc("POST /api/v1/hunting/notebooks", router.hunting.SaveNotebook)
	mux.HandleFunc("GET /api/v1/hunting/notebooks/", router.hunting.GetNotebook)

	// MITRE Endpoints
	mux.HandleFunc("GET /api/v1/security/mitre/coverage", router.mitre.GetCoverage)

	return mux
}
