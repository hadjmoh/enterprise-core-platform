package cluster

import (
	"context"
	"enterprise-core/backend/internal/query/cost"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"sync"
	"time"
)

// ScalingEvent records an auto-scaling action
type ScalingEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"` // "scale_up" or "scale_down"
	Nodes     int       `json:"nodes_delta"`
	Reason    string    `json:"reason"`
	CostEst   float64   `json:"cost_estimate"`
}

// AutoScaler orchestrates cluster scaling operations
type AutoScaler struct {
	monitor     *ResourceMonitor
	policy      ScalingPolicy
	costEngine  *cost.PolicyEngine
	logger      *logger.Logger
	lastScaled  time.Time
	history     []ScalingEvent
	killSwitch  bool
	mu          sync.RWMutex
}

func NewAutoScaler(mon *ResourceMonitor, ce *cost.PolicyEngine, l *logger.Logger) *AutoScaler {
	return &AutoScaler{
		monitor:    mon,
		policy:     DefaultPolicy(),
		costEngine: ce,
		logger:     l,
		history:    make([]ScalingEvent, 0),
	}
}

func (as *AutoScaler) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			as.Evaluate(ctx)
		}
	}
}

func (as *AutoScaler) Evaluate(ctx context.Context) {
	as.mu.Lock()
	if !as.policy.AutoScalingEnabled || as.killSwitch {
		as.mu.Unlock()
		return
	}
	as.mu.Unlock()

	if time.Since(as.lastScaled) < as.policy.ScalingCooldown {
		return
	}

	metrics := as.monitor.GetClusterMetrics()
	nodes := as.monitor.GetNodes()
	nodeCount := len(nodes)

	// Scale Up Logic
	if metrics.CPUUsage > as.policy.TargetCPU || metrics.MemUsage > as.policy.TargetMemory {
		if nodeCount < as.policy.MaxNodes {
			delta := 1
			if metrics.CPUUsage > 95 { // Critical load
				delta = as.policy.BlastRadiusLimit
			}
			as.Scale(ctx, "scale_up", delta, fmt.Sprintf("High utilization: CPU=%.2f%%, MEM=%.2f%%", metrics.CPUUsage, metrics.MemUsage))
		}
	}

	// Scale Down Logic (Omitted for brevity in PoC, but would follow similar threshold logic)
}

func (as *AutoScaler) Scale(ctx context.Context, action string, delta int, reason string) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	// 1. Cost Safety Check
	// Mock node cost: $100/mo per node
	nodeCost := 100.0
	extraCost := float64(delta) * nodeCost
	
	// Real implementation would call costEngine.EstimateScaling(...)
	
	if action == "scale_up" {
		as.logger.Info("Scaling up cluster", "delta", delta, "reason", reason, "estimated_cost", extraCost)
		
		// 2. Perform scaling (simulation)
		for i := 0; i < delta; i++ {
			newNodeID := fmt.Sprintf("node-%02d", len(as.monitor.GetNodes())+1)
			as.monitor.AddNode(&ClusterNode{
				ID:       newNodeID,
				Status:   NodeOnline,
				LastSeen: time.Now(),
				Metrics:  Metrics{CPUUsage: 10.0, MemUsage: 15.0}, // Fresh node
			})
		}
	}

	as.lastScaled = time.Now()
	as.history = append(as.history, ScalingEvent{
		Timestamp: as.lastScaled,
		Action:    action,
		Nodes:     delta,
		Reason:    reason,
		CostEst:   extraCost,
	})

	return nil
}

func (as *AutoScaler) ToggleKillSwitch(status bool) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.killSwitch = status
	as.logger.Warn("Auto-scaling kill-switch toggled", "status", status)
}

func (as *AutoScaler) UpdatePolicy(p ScalingPolicy) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.policy = p
}

func (as *AutoScaler) GetStatus() map[string]interface{} {
	as.mu.RLock()
	defer as.mu.RUnlock()

	return map[string]interface{}{
		"kill_switch":   as.killSwitch,
		"last_scaled":   as.lastScaled,
		"history":       as.history,
		"policy":        as.policy,
		"active_nodes":  len(as.monitor.GetNodes()),
	}
}
