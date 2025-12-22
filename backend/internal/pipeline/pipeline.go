package pipeline

import (
	"enterprise-core/backend/internal/alerting"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/correlation"
	"enterprise-core/backend/internal/detection"
	"enterprise-core/backend/internal/enrichment"
	"enterprise-core/backend/internal/storage"
	"enterprise-core/backend/internal/tenant"
	"enterprise-core/backend/internal/security"
	"enterprise-core/backend/internal/security/risk"
	"enterprise-core/backend/internal/analytics"
	"enterprise-core/backend/internal/compliance"
	"enterprise-core/backend/internal/governance"
	"enterprise-core/backend/pkg/logger"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrOverloaded = errors.New("system overloaded: backpressure applied")
)

// IngestionPipeline orchestrates all enterprise ingestion stages
type IngestionPipeline struct {
	enrichment     *enrichment.Pipeline
	correlation    *correlation.Engine
	detection      *detection.Engine
	alerting       *alerting.Router
	buffer         *buffer.RingBuffer
	storage        storage.StorageEngine
	backpressure   *buffer.BackpressureController
	tenantManager *tenant.Manager
	ueba           *security.UEBAEngine
	risk           *risk.RiskEngine
	soar           *security.Orchestrator
	featureExtractor *analytics.FeatureExtractor
	vault          *compliance.PrivacyVault
	merkleTree     *compliance.MerkleTree
	lineageTracker *security.LineageTracker
	trustGraph     *security.TrustGraph
	governor       *governance.Governor
	logger         *logger.Logger
}

func NewIngestionPipeline(
	enrichment *enrichment.Pipeline,
	correlation *correlation.Engine,
	detection *detection.Engine,
	alerting *alerting.Router,
	buffer *buffer.RingBuffer,
	store storage.StorageEngine,
	backpressure *buffer.BackpressureController,
	tenantManager *tenant.Manager,
	ueba *security.UEBAEngine,
	risk *risk.RiskEngine,
	soar *security.Orchestrator,
	featureExtractor *analytics.FeatureExtractor,
	vault *compliance.PrivacyVault,
	merkleTree *compliance.MerkleTree,
	lineageTracker *security.LineageTracker,
	trustGraph *security.TrustGraph,
	gov *governance.Governor,
	logger *logger.Logger,
) *IngestionPipeline {
	return &IngestionPipeline{
		enrichment:     enrichment,
		correlation:    correlation,
		detection:      detection,
		alerting:       alerting,
		buffer:         buffer,
		storage:        store,
		backpressure:   backpressure,
		tenantManager: tenantManager,
		ueba:           ueba,
		risk:           risk,
		soar:           soar,
		featureExtractor: featureExtractor,
		vault: vault,
		merkleTree: merkleTree,
		lineageTracker: lineageTracker,
		trustGraph: trustGraph,
		governor: gov,
		logger:         logger,
	}
}

func (p *IngestionPipeline) Process(event buffer.Event) error {
	// 0. Update Backpressure load
	p.backpressure.UpdateLoad(p.buffer.Size())
	
	// 0.05 Check Governance Kill-switch
	if p.governor != nil && p.governor.IsDisabled(governance.IngestSubsystem) {
		return errors.New("ingestion halted by security control plane")
	}

	// 0.1 Check Backpressure
	if !p.backpressure.ShouldAccept() {
		return ErrOverloaded
	}

	// 0.2 Extract Tenant and apply isolation logic if needed
	tenantID := p.tenantManager.ExtractTenantID(event.Data)
	event.Data["tenant_id"] = tenantID

	// 0.2.1 Record Trust Flow
	if p.trustGraph != nil {
		p.trustGraph.UpdateNode(event.Source, "system", security.TrustVerified, nil)
		if user, ok := event.Data["user"].(string); ok {
			p.trustGraph.UpdateNode(user, "user", security.TrustReputable, nil)
			p.trustGraph.AddEdge(user, event.Source, "produced", security.TrustReputable, nil)
		}
	}

	// 0.3 Data Governance: Tokenization (Session 7.6)
	if p.vault != nil {
		if email, ok := event.Data["email"].(string); ok {
			event.Data["email"] = p.vault.Tokenize(email)
		}
		if ssn, ok := event.Data["ssn"].(string); ok {
			event.Data["ssn"] = p.vault.Tokenize(ssn)
		}
	}

	// 0.4 Initialize Lineage
	if p.lineageTracker != nil {
		event.Lineage = append(event.Lineage, p.lineageTracker.CreateStep("ingestion", "receival", event.Data))
	}

	// 1. Enrichment
	if err := p.enrichment.Handle(event); err != nil {
		p.logger.Error("Enrichment failed", err)
	} else if p.lineageTracker != nil {
		event.Lineage = append(event.Lineage, p.lineageTracker.CreateStep("enrichment", "enforce", event.Data))
	}

	// 2. Correlation
	correlations := p.correlation.ProcessEvent(event)
	for _, corr := range correlations {
		p.logger.Info("Correlation detected", "rule", corr.RuleName, "severity", corr.Severity)
		// Route correlation to alerting
		p.alerting.SendAlert(&alerting.Alert{
			ID:          corr.RuleID,
			Title:       "Correlation: " + corr.RuleName,
			Description: fmt.Sprintf("Correlation rule triggered at %s", time.Now().Format(time.RFC3339)),
			Severity:    corr.Severity,
			Source:      event.Source,
			Timestamp:   time.Now(),
		})

		// SOAR: Auto-stage Block IP for Critical correlations
		if corr.Severity == "critical" && p.soar != nil {
			if srcIP, ok := event.Data["src_ip"].(string); ok {
				p.soar.StageAction("block_ip", srcIP, "critical", map[string]interface{}{"ip": srcIP})
			}
		}
	}

	if p.lineageTracker != nil {
		event.Lineage = append(event.Lineage, p.lineageTracker.CreateStep("correlation", "match", event.Data))
	}

	// 3. Detection
	detections := p.detection.ProcessEvent(event)
	for _, det := range detections {
		p.logger.Info("Detection triggered", "rule", det.RuleName, "severity", det.Severity)
		// Route detection to alerting
		p.alerting.SendAlert(&alerting.Alert{
			ID:          det.RuleID,
			Title:       "Detection: " + det.RuleName,
			Description: fmt.Sprintf("Detection rule triggered for event from %s", event.Source),
			Severity:    det.Severity,
			Source:      event.Source,
			Timestamp:   time.Now(),
		})
	}

	// 3.1 UEBA Anomaly Detection
	if p.ueba != nil {
		anomalies := p.ueba.Process(event)
		for _, anomaly := range anomalies {
			p.logger.Info("UEBA Anomaly detected", "type", anomaly.Type, "entity", anomaly.EntityID)
			// Route to alerting
			p.alerting.SendAlert(&alerting.Alert{
				ID:          "UEBA-" + anomaly.Type,
				Title:       "UEBA: " + anomaly.Type,
				Description: anomaly.Description,
				Severity:    anomaly.Severity,
				Source:      event.Source,
				Timestamp:   time.Now(),
			})

			// Increment Risk
			if p.risk != nil {
				score := 20 // Default UEBA anomaly weight
				if anomaly.Severity == "high" {
					score = 40
				}
				p.risk.IncrementRisk(anomaly.EntityID, score)
			}

			// SOAR: Auto-stage for High severity UEBA anomalies (e.g. Impossible Travel)
			if anomaly.Severity == "high" && p.soar != nil {
				action := "block_ip"
				params := map[string]interface{}{"ip": anomaly.EntityID} // EntityID might be IP
				if !strings.Contains(anomaly.EntityID, ".") {
					action = "disable_user"
					params = map[string]interface{}{"user": anomaly.EntityID}
				}
				p.soar.StageAction(action, anomaly.EntityID, "high", params)
			}
		}
	}
	
	// 3.2 Update Behavioral Features (Session 7.4)
	if p.featureExtractor != nil {
		p.featureExtractor.ProcessEvent(event)
	}

	// 4. Persistence to Storage Engine
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := p.storage.Write(ctx, event); err != nil {
		p.logger.Error("Storage persistence failed", err)
	}

	// 4.1 Audit Proof Generation (Session 7.6)
	if p.merkleTree != nil {
		// In a real system we would hash the entire event content
		// Here we just use the event ID or Source + Time
		data := fmt.Sprintf("%v-%v", event.Source, event.Timestamp)
		p.merkleTree.AddLeaf([]byte(data))
	}

	// 5. Push to Ring Buffer for real-time alerting/workers
	return p.buffer.Push(event)
}

// ProcessSimulation executes the pipeline logic in 'shadow' mode for testing/replay
func (p *IngestionPipeline) ProcessSimulation(event buffer.Event) (map[string]interface{}, error) {
	outcome := map[string]interface{}{
		"event_id": event.ID,
		"triggered_alert": false,
		"detections": []string{},
		"correlations": []string{},
	}

	// 1. Enrichment (Read-only)
	p.enrichment.Handle(event)

	// 2. Correlation (Analysis only)
	correlations := p.correlation.ProcessEvent(event)
	for _, corr := range correlations {
		outcome["triggered_alert"] = true
		outcome["correlations"] = append(outcome["correlations"].([]string), corr.RuleName)
	}

	// 3. Detection (Analysis only)
	detections := p.detection.ProcessEvent(event)
	for _, det := range detections {
		outcome["triggered_alert"] = true
		outcome["detections"] = append(outcome["detections"].([]string), det.RuleName)
	}

	// 3.1 UEBA (Analysis only)
	if p.ueba != nil {
		anomalies := p.ueba.Process(event)
		if len(anomalies) > 0 {
			outcome["triggered_alert"] = true
			outcome["ueba_anomalies"] = len(anomalies)
		}
	}

	return outcome, nil
}
