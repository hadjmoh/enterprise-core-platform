package compliance

import (
	"enterprise-core/backend/pkg/logger"
	"regexp"
	"strings"
)

// Masker redacts sensitive information for compliance
type Masker struct {
	patterns map[string]*regexp.Regexp
	logger   *logger.Logger
}

func NewMasker(logger *logger.Logger) *Masker {
	return &Masker{
		patterns: map[string]*regexp.Regexp{
			"email":    regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
			"cc":       regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`),
			"ssn":      regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			"ipv4":     regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),
			"api_key":  regexp.MustCompile(`(?i)(api|secret|key|token|auth)[-_\s]*[:=][-_\s]*([a-zA-Z0-9]{16,})`),
		},
		logger: logger,
	}
}

func (m *Masker) Mask(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		if str, ok := v.(string); ok {
			result[k] = m.maskString(str)
		} else {
			result[k] = v
		}
	}
	return result
}

func (m *Masker) maskString(s string) string {
	res := s
	for type_, pattern := range m.patterns {
		if type_ == "api_key" {
			// Special handling for key-value labels
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
