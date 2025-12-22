package pilot

import (
	"fmt"
	"regexp"
	"strings"
)

// Translator converts natural language to SPL using patterns
type Translator struct {
	patterns []Pattern
}

// NewTranslator creates a new translator with default patterns
func NewTranslator() *Translator {
	t := &Translator{
		patterns: []Pattern{},
	}
	t.loadDefaultPatterns()
	return t
}

// Translate converts natural language to SPL
func (t *Translator) Translate(nlQuery string) (string, float64) {
	nlQuery = strings.TrimSpace(nlQuery)
	bestScore := 0.0
	bestSPL := ""

	for _, p := range t.patterns {
		re := regexp.MustCompile("(?i)" + p.Regex)
		match := re.FindStringSubmatch(nlQuery)
		
		if match != nil {
			// Found a match
			// Replace placeholders in template
			spl := p.Template
			
			// If patterns have capture groups, replace $1, $2, etc.
			// Note: Go regex groups are 1-indexed in FindStringSubmatch
			for i := 1; i < len(match); i++ {
				placeholder := fmt.Sprintf("$%d", i)
				spl = strings.ReplaceAll(spl, placeholder, match[i])
			}

			// Simple confidence scoring based on pattern specificity
			score := p.BaseConfidence
			if score > bestScore {
				bestScore = score
				bestSPL = spl
			}
		}
	}

	if bestScore == 0 {
		return "", 0.0
	}

	return bestSPL, bestScore
}

func (t *Translator) loadDefaultPatterns() {
	t.patterns = []Pattern{
		{
			Regex:          `^show me errors$`,
			Template:       `search index=* "error" | stats count by sourcetype`,
			Description:    "Find generic errors",
			BaseConfidence: 0.9,
		},
		{
			Regex:          `^show errors from last (\d+)([hm])$`,
			Template:       `search index=* "error" earliest=-$1$2 | stats count by sourcetype`,
			Description:    "Find errors with relative time",
			BaseConfidence: 0.95,
		},
		{
			Regex:          `^find failed logins$`,
			Template:       `search index=auth "failed" | stats count by user`,
			Description:    "Find authentication failures",
			BaseConfidence: 0.9,
		},
		{
			Regex:          `^count (.*) by (.*)$`,
			Template:       `search index=* | stats count($1) by $2`,
			Description:    "Generic aggregation",
			BaseConfidence: 0.8,
		},
		{
			Regex:          `^search for (.*)$`,
			Template:       `search index=* "$1"`,
			Description:    "Basic keyword search",
			BaseConfidence: 0.85,
		},
		{
			Regex:          `^who logged in from ip (.*)$`,
			Template:       `search index=auth src_ip="$1" action="login"`,
			Description:    "Find user by IP",
			BaseConfidence: 0.9,
		},
		{
			Regex:          `^show traffic from (.*) to (.*)$`,
			Template:       `search index=network src_ip="$1" dest_ip="$2"`,
			Description:    "Network traffic flow",
			BaseConfidence: 0.9,
		},
	}
}
