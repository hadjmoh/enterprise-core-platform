package compliance

import (
	"crypto/sha256"
	"encoding/hex"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"regexp"
	"strings"
)

// Masker redacts sensitive information for compliance
type Masker struct {
	patterns map[string]*regexp.Regexp
	policies *PolicyRegistry
	salt     string
	logger   *logger.Logger
}

func NewMasker(policies *PolicyRegistry, logger *logger.Logger) *Masker {
	return &Masker{
		patterns: map[string]*regexp.Regexp{
			"email":    regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
			"cc":       regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`),
			"ssn":      regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			"ipv4":     regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),
			"api_key":  regexp.MustCompile(`(?i)(api|secret|key|token|auth)[-_\s]*[:=][-_\s]*([a-zA-Z0-9]{16,})`),
		},
		policies: policies,
		salt:     "platform-wide-static-salt-for-deterministic-hashing",
		logger:   logger,
	}
}

func (m *Masker) Mask(data map[string]interface{}, role string) map[string]interface{} {
	role = strings.ToLower(role)
	policy, ok := m.policies.GetPolicy(role)
	if !ok {
		// Default to analyst if role unknown for safety
		policy, _ = m.policies.GetPolicy("analyst")
	}

	result := make(map[string]interface{})
	for k, v := range data {
		if str, ok := v.(string); ok {
			result[k] = m.applyPolicy(k, str, policy)
		} else {
			result[k] = v
		}
	}
	return result
}

func (m *Masker) applyPolicy(field, value string, p *PrivacyPolicy) string {
	strategy := p.FieldRules[field]
	if strategy == "" {
		// Check for wildcard rule
		strategy = p.FieldRules["*"]
	}

	switch strategy {
	case Cleartext:
		return value
	case Redact:
		return "[REDACTED]"
	case Hash:
		return m.hash(value)
	case Partial:
		if strings.Contains(value, ".") { // IP
			return m.maskIP(value)
		}
		// Generic partial: keep first 2 and last 2
		if len(value) > 6 {
			return value[:2] + "..." + value[len(value)-2:]
		}
		return "[MASKED]"
	default:
		// Fallback to pattern-based masking for deep sanity check
		return m.maskStringPatterns(value)
	}
}

func (m *Masker) maskStringPatterns(s string) string {
	res := s
	for type_, pattern := range m.patterns {
		if type_ == "api_key" {
			res = pattern.ReplaceAllStringFunc(res, func(match string) string {
				parts := strings.SplitN(match, ":", 2)
				if len(parts) < 2 {
					parts = strings.SplitN(match, "=", 2)
				}
				if len(parts) == 2 {
					return parts[0] + match[len(parts[0]):len(match)-len(parts[1])] + "********"
				}
				return match
			})
		} else {
			res = pattern.ReplaceAllString(res, "[MASKED_"+strings.ToUpper(type_)+"]")
		}
	}
	return res
}

func (m *Masker) hash(s string) string {
	// Deterministic hash with salt for analytics use-cases
	h := sha256.Sum256([]byte(s + m.salt))
	return hex.EncodeToString(h[:8]) // Return first 16 chars of hash
}

func (m *Masker) maskIP(s string) string {
	parts := strings.Split(s, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s.%s.%s.xxx", parts[0], parts[1], parts[2])
	}
	return "[MASKED_IP]"
}
