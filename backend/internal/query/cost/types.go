package cost


// RiskLevel represents the risk classification of a query
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// QueryCostEstimate contains the estimated resource usage and risk
type QueryCostEstimate struct {
	EstimatedRows       int64     `json:"estimated_rows"`
	EstimatedScanGB     float64   `json:"estimated_scan_gb"`
	EstimatedCPUSeconds float64   `json:"estimated_cpu_seconds"`
	RiskLevel           RiskLevel `json:"risk_level"`
	RiskFactors         []string  `json:"risk_factors"`
	Warnings            []string  `json:"warnings"`
	RequiredPermissions []string  `json:"required_permissions"`
}

// CostPolicy defines the limits for a specific risk level
type CostPolicy struct {
	MaxScanGB     float64
	MaxRows       int64
	MaxCPUSeconds float64
	RequiredRole  string
	RequiresAdmin bool
}
