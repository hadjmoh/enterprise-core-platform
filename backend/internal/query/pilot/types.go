package pilot

import (
	"enterprise-core/backend/internal/query/cost"
)

// Suggestion represents a suggested SPL query and its metadata
type Suggestion struct {
	OriginalQuery      string                  `json:"original_query"`
	SuggestedSPL       string                  `json:"suggested_spl"`
	ConfidenceScore    float64                 `json:"confidence_score"`
	Explanation        string                  `json:"explanation"`
	CostEstimate       *cost.QueryCostEstimate `json:"cost_estimate"`
	Allowed            bool                    `json:"allowed"`
	BlockingReason     string                  `json:"blocking_reason"`
}

// Pattern represents a regex pattern to match natural language
type Pattern struct {
	Regex       string
	Template    string
	Description string
	BaseConfidence float64
}
