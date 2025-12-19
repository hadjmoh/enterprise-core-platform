package detection

import (
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/pkg/logger"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Engine processes events against detection rules
type Engine struct {
	rules   []*Rule
	logger  *logger.Logger
	mu      sync.RWMutex
	enabled bool
}

type Rule struct {
	ID          string                 `yaml:"id" json:"id"`
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Severity    string                 `yaml:"severity" json:"severity"` // low, medium, high, critical
	Enabled     bool                   `yaml:"enabled" json:"enabled"`
	Conditions  []Condition            `yaml:"conditions" json:"conditions"`
	Metadata    map[string]interface{} `yaml:"metadata" json:"metadata"`
}

type Condition struct {
	Field    string      `yaml:"field" json:"field"`
	Operator string      `yaml:"operator" json:"operator"` // eq, ne, gt, lt, contains, regex, in
	Value    interface{} `yaml:"value" json:"value"`
}

type Detection struct {
	RuleID      string
	RuleName    string
	Severity    string
	Event       buffer.Event
	MatchedData map[string]interface{}
}

func NewEngine(logger *logger.Logger) *Engine {
	return &Engine{
		rules:   make([]*Rule, 0),
		logger:  logger,
		enabled: true,
	}
}

func (e *Engine) LoadRulesFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var rules []*Rule
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.rules = rules
	e.logger.Info("Loaded detection rules", "count", len(rules), "file", path)
	return nil
}

func (e *Engine) AddRule(rule *Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.rules = append(e.rules, rule)
	e.logger.Info("Added detection rule", "id", rule.ID, "name", rule.Name)
}

func (e *Engine) ProcessEvent(event buffer.Event) []*Detection {
	if !e.enabled {
		return nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	detections := make([]*Detection, 0)

	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}

		if e.matchRule(rule, event) {
			detection := &Detection{
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				Severity:    rule.Severity,
				Event:       event,
				MatchedData: event.Data,
			}
			detections = append(detections, detection)
			e.logger.Info("Detection triggered", "rule", rule.Name, "severity", rule.Severity)
		}
	}

	return detections
}

func (e *Engine) matchRule(rule *Rule, event buffer.Event) bool {
	// All conditions must match (AND logic)
	for _, cond := range rule.Conditions {
		if !e.matchCondition(cond, event.Data) {
			return false
		}
	}
	return len(rule.Conditions) > 0
}

func (e *Engine) matchCondition(cond Condition, data map[string]interface{}) bool {
	value, exists := data[cond.Field]
	if !exists {
		return false
	}

	switch cond.Operator {
	case "eq", "equals":
		return value == cond.Value
	case "ne", "not_equals":
		return value != cond.Value
	case "gt", "greater_than":
		return compareNumbers(value, cond.Value, ">")
	case "lt", "less_than":
		return compareNumbers(value, cond.Value, "<")
	case "gte":
		return compareNumbers(value, cond.Value, ">=")
	case "lte":
		return compareNumbers(value, cond.Value, "<=")
	case "contains":
		return containsString(value, cond.Value)
	case "in":
		return inList(value, cond.Value)
	}

	return false
}

func (e *Engine) Enable() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.enabled = true
}

func (e *Engine) Disable() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.enabled = false
}

// Helper functions
func compareNumbers(a, b interface{}, op string) bool {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)
	if !aOk || !bOk {
		return false
	}

	switch op {
	case ">":
		return aFloat > bFloat
	case "<":
		return aFloat < bFloat
	case ">=":
		return aFloat >= bFloat
	case "<=":
		return aFloat <= bFloat
	}
	return false
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	}
	return 0, false
}

func containsString(haystack, needle interface{}) bool {
	h, hOk := haystack.(string)
	n, nOk := needle.(string)
	if !hOk || !nOk {
		return false
	}
	return len(h) > 0 && len(n) > 0 && len(h) >= len(n) && h != n && (h == n || len(h) > len(n))
}

func inList(value, list interface{}) bool {
	arr, ok := list.([]interface{})
	if !ok {
		return false
	}
	for _, item := range arr {
		if item == value {
			return true
		}
	}
	return false
}
