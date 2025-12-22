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
	"enterprise-core/backend/internal/cluster"
	"enterprise-core/backend/internal/simulation"
	"enterprise-core/backend/internal/governance"
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
	cluster    *ClusterHandler
	trustGraph *security.TrustGraph
	simulation *SimulationHandler
	governance *GovernanceHandler
	auth       auth.Service
	logger     *logger.Logger
}

func NewRouter(pipe *pipeline.IngestionPipeline, disp *query.Dispatcher, authSvc auth.Service, riskEngine *risk.RiskEngine, uebaEngine *security.UEBAEngine, soarOrch *security.Orchestrator, policies *compliance.PolicyRegistry, huntMgr *security.HuntingManager, mitreMgr *security.MitreManager, costEngine *cost.PolicyEngine, pilotEngine *pilot.PilotEngine, featureStore *analytics.FeatureStore, anomalyDetector *analytics.AnomalyDetector, explainerEngine *analytics.ExplainerEngine, feedbackStore *analytics.FeedbackStore, merkleTree *compliance.MerkleTree, scaler *cluster.AutoScaler, mon *cluster.ResourceMonitor, trustGraph *security.TrustGraph, simEngine *simulation.Engine, gov *governance.Governor, drift *governance.DriftDetector, pe *governance.PolicyEngine, l *logger.Logger) *http.ServeMux {
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
		cluster:    NewClusterHandler(scaler, mon),
		trustGraph: trustGraph,
		simulation: NewSimulationHandler(simEngine),
		governance: NewGovernanceHandler(gov, drift, pe),
		auth:       authSvc,
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

	// Compliance Endpoints
	mux.HandleFunc("GET /api/v1/compliance/proof", router.handleGetProof)
	mux.HandleFunc("POST /api/v1/compliance/verify", router.handleVerifyProof)

	// Cluster Endpoints
	mux.HandleFunc("GET /api/v1/cluster/status", router.cluster.GetStatus)
	mux.HandleFunc("POST /api/v1/cluster/scale", router.cluster.ManualScale)
	mux.HandleFunc("POST /api/v1/cluster/policy", router.cluster.UpdatePolicy)
	mux.HandleFunc("POST /api/v1/cluster/killswitch", router.cluster.ToggleKillSwitch)
	
	// Lineage & Trust Endpoints (Session 7.8)
	mux.HandleFunc("GET /api/v1/lineage/event", router.handleGetEventLineage)
	mux.HandleFunc("GET /api/v1/trust/graph", router.handleGetTrustGraph)

	// Simulation Endpoints (Session 7.9)
	mux.HandleFunc("POST /api/v1/simulation/start", router.simulation.StartSimulation)
	mux.HandleFunc("GET /api/v1/simulation/result", router.simulation.GetResult)

	// Governance Endpoints (Session 7.10)
	mux.HandleFunc("GET /api/v1/governance/status", router.governance.GetStatus)
	mux.HandleFunc("POST /api/v1/governance/emergency/lockdown", router.governance.Lockdown)
	mux.HandleFunc("POST /api/v1/governance/emergency/release", router.governance.ReleaseLockdown)
	mux.HandleFunc("POST /api/v1/governance/emergency/killswitch", router.governance.ToggleKillSwitch)
	mux.HandleFunc("GET /api/v1/governance/drift", router.governance.GetDrift)

	return mux
}
