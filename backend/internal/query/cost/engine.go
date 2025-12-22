package cost

import (
	"fmt"
	"time"

)

// PolicyEngine enforces cost policies
type PolicyEngine struct {
	policies map[RiskLevel]CostPolicy
	estimator *Estimator
}

// NewPolicyEngine creates a new policy engine with default policies
func NewPolicyEngine() *PolicyEngine {
	// Define default policies
	policies := map[RiskLevel]CostPolicy{
		RiskLow: {
			MaxScanGB:     10.0,
			MaxRows:       1_000_000,
			MaxCPUSeconds: 10.0,
			RequiredRole:  "guest", // Everyone
		},
		RiskMedium: {
			MaxScanGB:     50.0,
			MaxRows:       5_000_000,
			MaxCPUSeconds: 60.0,
			RequiredRole:  "analyst", // Regular users
		},
		RiskHigh: {
			MaxScanGB:     500.0,
			MaxRows:       50_000_000,
			MaxCPUSeconds: 300.0,
			RequiredRole:  "power_user", // Senior analysts
			RequiresAdmin: false,
		},
		RiskCritical: {
			MaxScanGB:     5000.0,
			MaxRows:       500_000_000,
			MaxCPUSeconds: 3600.0,
			RequiredRole:  "admin", // Admins only
			RequiresAdmin: true,
		},
	}

	return &PolicyEngine{
		policies:  policies,
		estimator: NewEstimator(),
	}
}

// EstimateAndCheck calculates cost and checks if the user is allowed to run it
func (e *PolicyEngine) EstimateAndCheck(spl string, timeRangeStart, timeRangeEnd time.Time, userRole string) (*QueryCostEstimate, bool, string, error) {
	// 1. Estimate Cost
	estimate, err := e.estimator.Estimate(spl, timeRangeStart, timeRangeEnd)
	if err != nil {
		return nil, false, "", fmt.Errorf("estimation failed: %w", err)
	}

	// 2. Get Policy for Risk Level
	policy, ok := e.policies[estimate.RiskLevel]
	if !ok {
		// Fallback to most restrictive if risk level unknown
		policy = e.policies[RiskCritical]
	}

	// 3. Check Permissions
	allowed := false
	reason := ""

	// Mapping roles to hierarchy (simplistic for now)
	// guest < analyst < power_user < admin
	rolePower := getRolePower(userRole)
	requiredPower := getRolePower(policy.RequiredRole)

	if rolePower >= requiredPower {
		allowed = true
	} else {
		allowed = false
		reason = fmt.Sprintf("Query risk is %s (requires %s role, you have %s)", estimate.RiskLevel, policy.RequiredRole, userRole)
	}

	// 4. Special Admin Requirement
	if policy.RequiresAdmin && userRole != "admin" {
		allowed = false
		reason = fmt.Sprintf("Query risk is %s and requires explicitly Admin role", estimate.RiskLevel)
	}

	// 5. Add warnings to reason if allowed
	if allowed && len(estimate.Warnings) > 0 {
		reason = "Allowed with warnings: " + fmt.Sprint(estimate.Warnings)
	}

	return estimate, allowed, reason, nil
}

func getRolePower(role string) int {
	switch role {
	case "admin":
		return 100
	case "power_user":
		return 50
	case "analyst":
		return 10
	case "guest":
		return 0
	default:
		return 0
	}
}
