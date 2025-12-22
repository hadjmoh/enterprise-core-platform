package cost

import (
	"fmt"
	"strings"
	"time"

	"enterprise-core/backend/internal/query"
)

// Estimator calculates the cost of a query
type Estimator struct {
	defaultScanRateGB float64 // Default GB scanned per hour
}

// NewEstimator creates a new cost estimator
func NewEstimator() *Estimator {
	return &Estimator{
		defaultScanRateGB: 10.0, // Assume 10GB/hour scan rate for unstructured logs
	}
}

// Estimate calculates the estimated cost and risk of a query
func (e *Estimator) Estimate(rawSPL string, timeRangeStart, timeRangeEnd time.Time) (*QueryCostEstimate, error) {
	// 1. Parse the query
	lexer := query.NewLexer(rawSPL)
	parser := query.NewParser(lexer)
	pipeline, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// 2. Initialize estimate
	estimate := &QueryCostEstimate{
		RiskLevel:   RiskLow,
		RiskFactors: []string{},
		Warnings:    []string{},
	}

	// 3. Analyze time range
	duration := timeRangeEnd.Sub(timeRangeStart)
	hours := duration.Hours()

	if hours > 24*30 { // > 30 days
		estimate.RiskLevel = RiskHigh
		estimate.RiskFactors = append(estimate.RiskFactors, "Time range exceeds 30 days")
		estimate.Warnings = append(estimate.Warnings, "Query covers a large time range. Consider reducing the range or using specific filters.")
	} else if hours <= 0 { // Unbounded (implied by 0 or negative, though usually handled by API defaults)
		// Assuming 0 means "all time" or invalid
		estimate.RiskLevel = RiskHigh
		estimate.RiskFactors = append(estimate.RiskFactors, "Unbounded time range")
	}

	// 4. Calculate estimated scan volume
	// Heuristic: GB = hours * defaultScanRate
	// In a real system, we'd query index stats for the specific sourcetype/index
	estimate.EstimatedScanGB = hours * e.defaultScanRateGB
	estimate.EstimatedRows = int64(hours * 1_000_000) // Heuristic: 1M events/hour

	if estimate.EstimatedScanGB > 1000 {
		estimate.RiskLevel = RiskCritical
		estimate.RiskFactors = append(estimate.RiskFactors, "Estimated scan volume > 1TB")
	} else if estimate.EstimatedScanGB > 100 {
		if estimate.RiskLevel != RiskCritical {
			estimate.RiskLevel = RiskHigh
		}
		estimate.RiskFactors = append(estimate.RiskFactors, "Estimated scan volume > 100GB")
	}

	// 5. Analyze commands
	for _, cmd := range pipeline.Commands {
		e.analyzeCommand(cmd, estimate)
	}

	// 6. Calculate CPU Impact (Heuristic)
	// Base CPU = 0.1s per GB scanned
	// Complex commands add multipliers
	cpuMultiplier := 1.0
	for _, cmd := range pipeline.Commands {
		switch cmd.Name {
		case "join", "transaction":
			cpuMultiplier *= 5.0
		case "sort", "stats", "timechart":
			cpuMultiplier *= 1.5
		case "regex", "grok":
			cpuMultiplier *= 2.0
		}
	}
	estimate.EstimatedCPUSeconds = estimate.EstimatedScanGB * 0.1 * cpuMultiplier

	if estimate.EstimatedCPUSeconds > 300 { // > 5 minutes CPU
		if estimate.RiskLevel != RiskCritical {
			estimate.RiskLevel = RiskHigh
		}
		estimate.RiskFactors = append(estimate.RiskFactors, "High CPU complexity")
	}

	return estimate, nil
}

func (e *Estimator) analyzeCommand(cmd query.CommandNode, estimate *QueryCostEstimate) {
	switch cmd.Name {
	case "delete":
		estimate.RiskLevel = RiskCritical
		estimate.RiskFactors = append(estimate.RiskFactors, "Destructive operation (delete)")
		estimate.RequiredPermissions = append(estimate.RequiredPermissions, "admin:delete")
		
	case "join":
		if estimate.RiskLevel != RiskCritical {
			estimate.RiskLevel = RiskHigh
		}
		estimate.RiskFactors = append(estimate.RiskFactors, "Cartesian join operation")
		estimate.Warnings = append(estimate.Warnings, "Joins can be very expensive. Ensure you are joining on high-selectivity keys.")

	case "transaction":
		if estimate.RiskLevel != RiskCritical {
			estimate.RiskLevel = RiskHigh
		}
		estimate.RiskFactors = append(estimate.RiskFactors, "Transaction operation (high memory)")

	case "search":
		// Check for wildcard prefix in arguments
		// This is a naive check; `search` args are key=value or raw strings
		for _, arg := range cmd.Args {
			if strings.HasPrefix(arg.Value, "*") && len(arg.Value) > 1 {
				estimate.RiskFactors = append(estimate.RiskFactors, "Prefix wildcard search")
				estimate.Warnings = append(estimate.Warnings, "Leading wildcards prevent index usage.")
			}
		}

	case "stats", "timechart":
		// Check if grouping by high cardinality fields could be risky
		// For now, just mark as memory intensive if many fields
		// TODO: Deep analysis of args
	}
}
