package governance

import (
	"errors"
	"fmt"
	"sync"
)

// Policy represents a declarative security rule
type Policy struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Action      string   `json:"action"`      // e.g. "ingest", "search", "config_change"
	Effect      string   `json:"effect"`      // "allow", "deny"
	Conditions  []Condition `json:"conditions"`
}

// Condition defines a single rule predicate
type Condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // "eq", "neq", "contains", "in"
	Value    interface{} `json:"value"`
}

// PolicyEngine evaluates system actions against defined policies
type PolicyEngine struct {
	policies map[string]*Policy
	mu       sync.RWMutex
}

func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		policies: make(map[string]*Policy),
	}
}

// AddPolicy registers a new policy
func (e *PolicyEngine) AddPolicy(p *Policy) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.policies[p.ID] = p
}

// Evaluate checks if an action is allowed based on input metadata
func (e *PolicyEngine) Evaluate(action string, input map[string]interface{}) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	allowed := true // Default to allow (adjust to deny-all for higher security)
	
	for _, p := range e.policies {
		if p.Action != action {
			continue
		}

		match := true
		for _, cond := range p.Conditions {
			val, ok := input[cond.Field]
			if !ok {
				match = false
				break
			}

			if !e.checkCondition(val, cond) {
				match = false
				break
			}
		}

		if match {
			if p.Effect == "deny" {
				return false, fmt.Errorf("action %s denied by policy %s", action, p.ID)
			}
			if p.Effect == "allow" {
				allowed = true
			}
		}
	}

	if !allowed {
		return false, errors.New("action not explicitly allowed by any policy")
	}

	return true, nil
}

func (e *PolicyEngine) checkCondition(val interface{}, cond Condition) bool {
	switch cond.Operator {
	case "eq":
		return fmt.Sprintf("%v", val) == fmt.Sprintf("%v", cond.Value)
	case "neq":
		return fmt.Sprintf("%v", val) != fmt.Sprintf("%v", cond.Value)
	// Add more operators as needed
	}
	return false
}
