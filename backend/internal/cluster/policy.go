package cluster

import "time"

// ScalingPolicy defines the rules and constraints for auto-scaling
type ScalingPolicy struct {
	MinNodes            int           `json:"min_nodes"`
	MaxNodes            int           `json:"max_nodes"`
	TargetCPU           float64       `json:"target_cpu"`            // Target CPU utilization percentage
	TargetMemory        float64       `json:"target_memory"`         // Target Memory utilization percentage
	ScalingCooldown     time.Duration `json:"scaling_cooldown"`      // Wait time between scaling events
	BlastRadiusLimit    int           `json:"blast_radius_limit"`   // Max nodes to add/remove in one event
	CostLimitPerMonth   float64       `json:"cost_limit_per_month"`  // Max monthly cost before blocking scaling
	AutoScalingEnabled bool          `json:"autoscaling_enabled"`
}

func DefaultPolicy() ScalingPolicy {
	return ScalingPolicy{
		MinNodes:            2,
		MaxNodes:            10,
		TargetCPU:           80.0,
		TargetMemory:        85.0,
		ScalingCooldown:     5 * time.Minute,
		BlastRadiusLimit:    2,
		CostLimitPerMonth:   5000.0,
		AutoScalingEnabled: true,
	}
}
