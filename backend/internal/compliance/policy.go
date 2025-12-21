package compliance

import (
	"enterprise-core/backend/pkg/logger"
	"sync"
)

// MaskingType defines the strategy for sanitizing data
type MaskingType string

const (
	Redact    MaskingType = "redact"    // Full replacement with [MASKED]
	Hash      MaskingType = "hash"      // Deterministic SHA-256 hash
	Partial   MaskingType = "partial"   // Leave first/last N characters
	Cleartext MaskingType = "clear"     // No masking
)

// PrivacyPolicy defines which fields are masked for a specific role
type PrivacyPolicy struct {
	Role       string                 `json:"role"`
	FieldRules map[string]MaskingType `json:"field_rules"`
}

// PolicyRegistry manages the active privacy policies for the platform
type PolicyRegistry struct {
	Policies map[string]*PrivacyPolicy
	mu       sync.RWMutex
	logger   *logger.Logger
}

func NewPolicyRegistry(logg *logger.Logger) *PolicyRegistry {
	r := &PolicyRegistry{
		Policies: make(map[string]*PrivacyPolicy),
		logger:   logg,
	}
	r.loadDefaultPolicies()
	return r
}

func (r *PolicyRegistry) GetPolicy(role string) (*PrivacyPolicy, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.Policies[role]
	return p, ok
}

func (r *PolicyRegistry) loadDefaultPolicies() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. Admin Policy: Full Transparency
	r.Policies["admin"] = &PrivacyPolicy{
		Role: "admin",
		FieldRules: map[string]MaskingType{
			"*": Cleartext,
		},
	}

	// 2. SOC Analyst Policy: Balanced Privacy
	r.Policies["analyst"] = &PrivacyPolicy{
		Role: "analyst",
		FieldRules: map[string]MaskingType{
			"email":       Hash,    // Consistently track users without seeing addresses
			"src_ip":      Partial, // 1.2.3.xxx
			"dst_ip":      Partial,
			"credit_card": Redact,  // Full redaction for high-risk PII
			"ssn":         Redact,
			"password":    Redact,
			"api_key":     Redact,
			"phone":       Hash,
		},
	}

	// 3. Auditor Policy: Minimum Viable Context
	r.Policies["auditor"] = &PrivacyPolicy{
		Role: "auditor",
		FieldRules: map[string]MaskingType{
			"user":   Cleartext, // Auditors need to know "Who"
			"action": Cleartext, // and "What"
			"*":      Redact,    // everything else is sanitized
		},
	}
}
