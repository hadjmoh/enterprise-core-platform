package correlation

import (
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/pkg/logger"
	"sync"
	"time"
)

// Engine performs event correlation and pattern detection
type Engine struct {
	rules      []*Rule
	windows    map[string]*TimeWindow
	logger     *logger.Logger
	mu         sync.RWMutex
	windowSize time.Duration
}

type TimeWindow struct {
	Events    []buffer.Event
	StartTime time.Time
	EndTime   time.Time
}

func NewEngine(windowSize time.Duration, logger *logger.Logger) *Engine {
	return &Engine{
		rules:      make([]*Rule, 0),
		windows:    make(map[string]*TimeWindow),
		logger:     logger,
		windowSize: windowSize,
	}
}

func (e *Engine) AddRule(rule *Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.rules = append(e.rules, rule)
	e.logger.Info("Added correlation rule", "name", rule.Name)
}

func (e *Engine) ProcessEvent(event buffer.Event) []Match {
	e.mu.Lock()
	defer e.mu.Unlock()

	matches := make([]Match, 0)

	// Add event to time windows
	e.addToWindows(event)

	// Check all rules
	for _, rule := range e.rules {
		if rule.Evaluate(e.windows, event) {
			match := Match{
				RuleID:    rule.ID,
				RuleName:  rule.Name,
				Severity:  rule.Severity,
				Timestamp: time.Now(),
				Events:    e.getRelatedEvents(event, rule),
			}
			matches = append(matches, match)
			e.logger.Info("Correlation match", "rule", rule.Name, "severity", rule.Severity)
		}
	}

	// Clean old windows
	e.cleanOldWindows()

	return matches
}

func (e *Engine) addToWindows(event buffer.Event) {
	// Create windows by source, user, IP, etc.
	keys := e.generateWindowKeys(event)
	
	for _, key := range keys {
		if window, ok := e.windows[key]; ok {
			window.Events = append(window.Events, event)
			window.EndTime = time.Now()
		} else {
			e.windows[key] = &TimeWindow{
				Events:    []buffer.Event{event},
				StartTime: time.Now(),
				EndTime:   time.Now(),
			}
		}
	}
}

func (e *Engine) generateWindowKeys(event buffer.Event) []string {
	keys := make([]string, 0)
	
	// Create windows by different dimensions
	if srcIP, ok := event.Data["src_ip"].(string); ok {
		keys = append(keys, "src_ip:"+srcIP)
	}
	if user, ok := event.Data["user"].(string); ok {
		keys = append(keys, "user:"+user)
	}
	if hostname, ok := event.Data["hostname"].(string); ok {
		keys = append(keys, "hostname:"+hostname)
	}
	
	// Global window for cross-source correlation
	keys = append(keys, "global")
	
	return keys
}

func (e *Engine) getRelatedEvents(event buffer.Event, rule *Rule) []buffer.Event {
	// Return events from the relevant window
	for _, window := range e.windows {
		for _, e := range window.Events {
			if e.Timestamp == event.Timestamp {
				return window.Events
			}
		}
	}
	return []buffer.Event{event}
}

func (e *Engine) cleanOldWindows() {
	now := time.Now()
	for key, window := range e.windows {
		if now.Sub(window.EndTime) > e.windowSize {
			delete(e.windows, key)
		}
	}
}

type Match struct {
	RuleID    string
	RuleName  string
	Severity  string
	Timestamp time.Time
	Events    []buffer.Event
}
