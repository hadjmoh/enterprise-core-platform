package pilot

import (
	"fmt"
	"time"

	"enterprise-core/backend/internal/query/cost"
)

// PilotEngine orchestrates the translation and safety checking of natural language queries
type PilotEngine struct {
	translator *Translator
	costEngine *cost.PolicyEngine
}

// NewPilotEngine creates a new pilot engine
func NewPilotEngine(costEngine *cost.PolicyEngine) *PilotEngine {
	return &PilotEngine{
		translator: NewTranslator(),
		costEngine: costEngine,
	}
}

// Suggest translates a natural language query into a safe SPL suggestion
func (p *PilotEngine) Suggest(nlQuery string, userRole string) (*Suggestion, error) {
	// 1. Translate
	spl, confidence := p.translator.Translate(nlQuery)
	if spl == "" {
		return nil, fmt.Errorf("could not translate query")
	}

	// 2. Estimate Cost and Check Safety
	// Use a default time range if not specified in the NL (Patterns handle relative time)
	// For estimation purposes, we assume "last 24h" unless the query specifies otherwise.
	// Since our translator produces SPL, we can just feed that SPL into the estimator.
	// Limitation: If the translator produces relative time 'earliest=-1h', the estimator needs to handle it or we pass a dummy range.
	// The current estimator parses the SPL, but ignores explicit earliest/latest in the SPL string in favor of the time range args.
	// We should probably update the estimator to prefer SPL args if present, but for now let's pass a standard 24h window for relative queries.
	
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	// 3. Safety Check
	estimate, allowed, reason, err := p.costEngine.EstimateAndCheck(spl, start, end, userRole)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate cost: %w", err)
	}

	suggestion := &Suggestion{
		OriginalQuery:   nlQuery,
		SuggestedSPL:    spl,
		ConfidenceScore: confidence,
		Explanation:     fmt.Sprintf("Translated based on pattern match (Confidence: %.2f)", confidence),
		CostEstimate:    estimate,
		Allowed:         allowed,
		BlockingReason:  reason,
	}

	// If blocked, we might want to redact the SPL or just warn.
	// For "Safe Mode", we return the result but the frontend will refuse to "Apply" it if Allowed is false.
	// But if it's CRITICAL risk (destructive), maybe we don't even return the SPL?
	if estimate.RiskLevel == cost.RiskCritical {
		suggestion.Explanation += " [CRITICAL RISK DETECTED]"
	}

	return suggestion, nil
}
