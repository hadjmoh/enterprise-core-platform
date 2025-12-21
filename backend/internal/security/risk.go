package security

import (
	"context"
	"enterprise-core/backend/pkg/logger"
	"sort"
	"sync"
	"time"
)

// RiskScore represents the calculated risk for an entity
type RiskScore struct {
	EntityID      string    `json:"entity_id"`
	CurrentScore  int       `json:"current_score"`
	LastUpdate    time.Time `json:"last_update"`
	IncidentCount int       `json:"incident_count"`
	SourceCount   int       `json:"source_count"`
}

// RiskEngine manages the calculation and storage of risk scores
type RiskEngine struct {
	scores map[string]*RiskScore
	mu     sync.RWMutex
	logger *logger.Logger
}

func NewRiskEngine(logg *logger.Logger) *RiskEngine {
	return &RiskEngine{
		scores: make(map[string]*RiskScore),
		logger: logg,
	}
}

func (e *RiskEngine) CalculateScore(severity string, confidence float64) int {
	base := 0
	switch severity {
	case "critical":
		base = 50
	case "high":
		base = 30
	case "medium":
		base = 15
	case "low":
		base = 5
	}

	return int(float64(base) * confidence)
}

func (e *RiskEngine) IncrementRisk(entityID string, score int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	s, ok := e.scores[entityID]
	if !ok {
		s = &RiskScore{
			EntityID: entityID,
		}
		e.scores[entityID] = s
	}

	s.CurrentScore += score
	if s.CurrentScore > 100 {
		s.CurrentScore = 100 // Cap at 100
	}
	s.LastUpdate = time.Now()
	s.IncidentCount++
	
	e.logger.Info("Incremented entity risk", "entity", entityID, "added", score, "total", s.CurrentScore)
}

func (e *RiskEngine) GetTopEntities(limit int) []*RiskScore {
	e.mu.RLock()
	defer e.mu.RUnlock()

	results := make([]*RiskScore, 0, len(e.scores))
	for _, s := range e.scores {
		results = append(results, s)
	}
	
	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].CurrentScore > results[j].CurrentScore
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

func (e *RiskEngine) GetEntityScore(entityID string) (*RiskScore, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	s, ok := e.scores[entityID]
	return s, ok
}

// RiskDecayService handles the reduction of risk scores over time
type RiskDecayService struct {
	engine      *RiskEngine
	interval    time.Duration
	decayFactor float64
	logger      *logger.Logger
}

func NewRiskDecayService(engine *RiskEngine, interval time.Duration, factor float64, logg *logger.Logger) *RiskDecayService {
	return &RiskDecayService{
		engine:      engine,
		interval:    interval,
		decayFactor: factor,
		logger:      logg,
	}
}

func (s *RiskDecayService) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.decayScores()
		}
	}
}

func (s *RiskDecayService) decayScores() {
	s.engine.mu.Lock()
	defer s.engine.mu.Unlock()

	for id, score := range s.engine.scores {
		if score.CurrentScore > 0 {
			old := score.CurrentScore
			score.CurrentScore = int(float64(score.CurrentScore) * s.decayFactor)
			if score.CurrentScore < 0 {
				score.CurrentScore = 0
			}
			
			if old != score.CurrentScore {
				s.logger.Debug("Decayed risk score", "entity", id, "from", old, "to", score.CurrentScore)
			}
		}
	}
}
