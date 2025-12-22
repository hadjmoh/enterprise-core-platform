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
	"enterprise-core/backend/pkg/logger"
	"context"
	"errors"
	"fmt"
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
		logger:         logger,
	}
}

func (p *IngestionPipeline) Process(event buffer.Event) error {
	// 0. Update Backpressure load
	p.backpressure.UpdateLoad(p.buffer.Size())

	// 0.1 Check Backpressure
	if !p.backpressure.ShouldAccept() {
		return ErrOverloaded
	}

	// 0.2 Extract Tenant and apply isolation logic if needed
	tenantID := p.tenantManager.ExtractTenantID(event.Data)
	event.Data["tenant_id"] = tenantID

	// 0.3 Data Governance: Tokenization (Session 7.6)
	if p.vault != nil {
		if email, ok := event.Data["email"].(string); ok {
			event.Data["email"] = p.vault.Tokenize(email)
		}
		if ssn, ok := event.Data["ssn"].(string); ok {
			event.Data["ssn"] = p.vault.Tokenize(ssn)
		}
	}

	// 1. Enrichment
	if err := p.enrichment.Handle(event); err != nil {
		p.logger.Error("Enrichment failed", err)
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
