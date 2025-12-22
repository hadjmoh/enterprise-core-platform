package analytics

import (
	"fmt"
)

// Explanation represents a reason why an anomaly or score was generated
type Explanation struct {
	EntityID     string   `json:"entity_id"`
	AnomalyType  string   `json:"anomaly_type"`
	Factors      []Factor `json:"factors"`
	Confidence   float64  `json:"confidence"`
	Description  string   `json:"description"`
}

// Factor represents a single contributing factor to the explanation
type Factor struct {
	Name        string  `json:"name"`
	Weight      float64 `json:"weight"`      // Contribution to the score (0.0 to 1.0)
	Value       float64 `json:"value"`       // Actual value
	Description string  `json:"description"` // Human readable description
}

// ExplainerEngine generates explanations for anomalies
type ExplainerEngine struct {
	// In a real system, this would have access to the model or historical data
}

// NewExplainerEngine creates a new explainer engine
func NewExplainerEngine() *ExplainerEngine {
	return &ExplainerEngine{}
}

// ExplainAnomaly generates a mock explanation for a given anomaly
// Simulates SHAP values by assigning robust weights to known anomaly types
func (e *ExplainerEngine) ExplainAnomaly(entityID string, anomalyType string, value float64) Explanation {
	factors := []Factor{}
	description := ""

	switch anomalyType {
	case "login_count_1h":
		factors = append(factors, Factor{
			Name:        "Login Frequency",
			Weight:      0.85,
			Value:       value,
			Description: fmt.Sprintf("User logged in %d times in the last hour, which is significantly higher than the baseline.", int(value)),
		})
		factors = append(factors, Factor{
			Name:        "Time of Day",
			Weight:      0.15,
			Value:       0, // Placeholder
			Description: "Activity occurred during off-peak hours.",
		})
		description = "Unusual high frequency of login attempts detected."

	case "bytes_out_1h":
		factors = append(factors, Factor{
			Name:        "Data Volume",
			Weight:      0.90,
			Value:       value,
			Description: fmt.Sprintf("Data transfer of %.2f MB exceeds 99th percentile of peer group.", value/1024/1024),
		})
		factors = append(factors, Factor{
			Name:        "Destination IP Reputation",
			Weight:      0.10,
			Value:       0,
			Description: "Destination IP has no prior history with this entity.",
		})
		description = "Potential data exfiltration detected based on volume."
	
	default:
		factors = append(factors, Factor{
			Name:        "Unknown Factor",
			Weight:      1.0,
			Value:       value,
			Description: "Anomaly detected based on statistical deviation.",
		})
		description = "Statistical anomaly detected."
	}

	return Explanation{
		EntityID:    entityID,
		AnomalyType: anomalyType,
		Factors:     factors,
		Confidence:  0.92, // Mock high confidence
		Description: description,
	}
}
