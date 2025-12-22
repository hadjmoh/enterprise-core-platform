package simulation

import (
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/storage"
	"enterprise-core/backend/pkg/logger"
	"sync"
	"time"
)

// SimulationResult captures the outcome of a simulation run
type SimulationResult struct {
	SessionID   string                 `json:"session_id"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	EventsCount int                    `json:"events_count"`
	AlertsCount int                    `json:"alerts_count"`
	Outcomes    []map[string]interface{} `json:"outcomes"`
}

// PipelineProxy is an interface for parts of the pipeline that can be simulated
type PipelineProxy interface {
	ProcessSimulation(event buffer.Event) (map[string]interface{}, error)
}

// Engine orchestrates simulation and replay tasks
type Engine struct {
	storage storage.StorageEngine
	pipeline PipelineProxy
	logger  *logger.Logger
	results map[string]*SimulationResult
	mu      sync.RWMutex
}

func NewEngine(store storage.StorageEngine, pipe PipelineProxy, l *logger.Logger) *Engine {
	return &Engine{
		storage: store,
		pipeline: pipe,
		logger:  l,
		results: make(map[string]*SimulationResult),
	}
}

// RunReplay replays events from storage through the simulation pipeline
func (e *Engine) RunReplay(sessionID string, from, to time.Time) (*SimulationResult, error) {
	e.logger.Info("Starting simulation replay", "session", sessionID, "from", from, "to", to)
	
	result := &SimulationResult{
		SessionID: sessionID,
		StartTime: time.Now(),
		Outcomes:  make([]map[string]interface{}, 0),
	}

	// This is a PoC: In real implementation, we would use storage.Query to get events
	// For now, we simulate finding some events
	mockEvents := []buffer.Event{
		{ID: "sim-1", Source: "firewall", Data: map[string]interface{}{"src_ip": "10.0.0.1", "action": "allow"}},
		{ID: "sim-2", Source: "firewall", Data: map[string]interface{}{"src_ip": "10.0.0.5", "action": "deny"}},
	}

	for _, evt := range mockEvents {
		outcome, err := e.pipeline.ProcessSimulation(evt)
		if err != nil {
			e.logger.Error("Simulation processing error", err)
			continue
		}
		result.EventsCount++
		if outcome["triggered_alert"] == true {
			result.AlertsCount++
		}
		result.Outcomes = append(result.Outcomes, outcome)
	}

	result.EndTime = time.Now()
	
	e.mu.Lock()
	e.results[sessionID] = result
	e.mu.Unlock()

	return result, nil
}

func (e *Engine) GetResult(sessionID string) (*SimulationResult, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	res, ok := e.results[sessionID]
	return res, ok
}
