package security

import (
	"enterprise-core/backend/internal/correlation"
	"enterprise-core/backend/internal/detection"
	"sync"
)

// MitreTactic represents a high-level MITRE tactic
type MitreTactic struct {
	Name        string   `json:"name"`
	Techniques  []string `json:"techniques"`
	Count       int      `json:"active_rule_count"`
	CoveragePct float64  `json:"coverage_percentage"`
}

// MitreTechnique represents a specific MITRE Technique
type MitreTechnique struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Rules []string `json:"associated_rules"`
}

// MitreCoverage provides a summary of the framework coverage
type MitreCoverage struct {
	Tactics    map[string]*MitreTactic    `json:"tactics"`
	Techniques map[string]*MitreTechnique `json:"techniques"`
}

// MitreManager calculates and serves ATT&CK coverage insights
type MitreManager struct {
	detEngine  *detection.Engine
	corrEngine *correlation.Engine
	mu         sync.RWMutex
}

func NewMitreManager(det *detection.Engine, corr *correlation.Engine) *MitreManager {
	return &MitreManager{
		detEngine:  det,
		corrEngine: corr,
	}
}

func (m *MitreManager) GetCoverage() *MitreCoverage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	coverage := &MitreCoverage{
		Tactics:    make(map[string]*MitreTactic),
		Techniques: make(map[string]*MitreTechnique),
	}

	// 1. Process Correlation Rules
	for _, rule := range correlation.DefaultRules {
		m.processMappings(rule.Name, rule.MitreTactics, rule.MitreTechniques, coverage)
	}

	// 2. Process Detection Rules
	if m.detEngine != nil {
		for _, rule := range m.detEngine.GetRules() {
			m.processMappings(rule.Name, rule.MitreTactics, rule.MitreTechniques, coverage)
		}
	}
	
	m.finalizeCalculations(coverage)
	return coverage
}

func (m *MitreManager) processMappings(ruleName string, tactics, techniques []string, coverage *MitreCoverage) {
	for _, tactic := range tactics {
		if _, ok := coverage.Tactics[tactic]; !ok {
			coverage.Tactics[tactic] = &MitreTactic{Name: tactic, Techniques: []string{}}
		}
		coverage.Tactics[tactic].Count++
	}

	for _, tech := range techniques {
		if _, ok := coverage.Techniques[tech]; !ok {
			coverage.Techniques[tech] = &MitreTechnique{ID: tech, Rules: []string{}}
		}
		coverage.Techniques[tech].Rules = append(coverage.Techniques[tech].Rules, ruleName)
	}
}

func (m *MitreManager) finalizeCalculations(coverage *MitreCoverage) {
	// Mock total counts for percentage calculation
	totalTechniquesPerTactic := 20.0 

	for _, tactic := range coverage.Tactics {
		// Count unique techniques covered for this tactic
		uniqueTechs := make(map[string]bool)
		for techID, tech := range coverage.Techniques {
			// In a real schema, we'd map Tech IDs to Tactics. 
			// For this logic, we'll just simulate progress.
			if len(tech.Rules) > 0 {
				uniqueTechs[techID] = true
			}
		}
		
		tactic.CoveragePct = (float64(tactic.Count) / totalTechniquesPerTactic) * 100
		if tactic.CoveragePct > 100 {
			tactic.CoveragePct = 100
		}
	}
}
