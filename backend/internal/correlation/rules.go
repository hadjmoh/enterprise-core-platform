package correlation

import (
	"enterprise-core/backend/internal/buffer"
	"strings"
)

// Rule defines a correlation pattern to detect
type Rule struct {
	ID          string
	Name        string
	Description string
	Severity    string // low, medium, high, critical
	Conditions  []Condition
	Threshold   int           // Minimum events to trigger
	TimeWindow  string        // e.g., "5m", "1h"
}

type Condition struct {
	Field    string
	Operator string // equals, contains, gt, lt, regex
	Value    interface{}
}

func (r *Rule) Evaluate(windows map[string]*TimeWindow, event buffer.Event) bool {
	// Find relevant window
	var relevantEvents []buffer.Event
	
	for _, window := range windows {
		for _, e := range window.Events {
			if r.matchesConditions(e) {
				relevantEvents = append(relevantEvents, e)
			}
		}
	}

	// Check if threshold is met
	return len(relevantEvents) >= r.Threshold
}

func (r *Rule) matchesConditions(event buffer.Event) bool {
	for _, cond := range r.Conditions {
		if !cond.Evaluate(event.Data) {
			return false
		}
	}
	return true
}

func (c *Condition) Evaluate(data map[string]interface{}) bool {
	value, ok := data[c.Field]
	if !ok {
		return false
	}

	switch c.Operator {
	case "equals":
		return value == c.Value
	case "contains":
		if str, ok := value.(string); ok {
			if substr, ok := c.Value.(string); ok {
				return strings.Contains(str, substr)
			}
		}
	case "gt":
		if num, ok := value.(float64); ok {
			if threshold, ok := c.Value.(float64); ok {
				return num > threshold
			}
		}
	case "lt":
		if num, ok := value.(float64); ok {
			if threshold, ok := c.Value.(float64); ok {
				return num < threshold
			}
		}
	}

	return false
}

// Example pre-defined rules
var DefaultRules = []*Rule{
	{
		ID:          "CORR-001",
		Name:        "Brute Force Detection",
		Description: "Multiple failed login attempts from same IP",
		Severity:    "high",
		Conditions: []Condition{
			{Field: "event_type", Operator: "equals", Value: "authentication"},
			{Field: "status", Operator: "equals", Value: "failed"},
		},
		Threshold: 5,
		TimeWindow: "5m",
	},
	{
		ID:          "CORR-002",
		Name:        "Port Scan Detection",
		Description: "Multiple connections to different ports from same IP",
		Severity:    "medium",
		Conditions: []Condition{
			{Field: "event_type", Operator: "equals", Value: "network"},
		},
		Threshold: 10,
		TimeWindow: "1m",
	},
	{
		ID:          "CORR-003",
		Name:        "Data Exfiltration",
		Description: "Large outbound data transfer",
		Severity:    "critical",
		Conditions: []Condition{
			{Field: "direction", Operator: "equals", Value: "outbound"},
			{Field: "bytes", Operator: "gt", Value: 1000000.0}, // 1MB
		},
		Threshold: 1,
		TimeWindow: "1m",
	},
}
