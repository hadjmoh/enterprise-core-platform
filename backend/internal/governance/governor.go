package governance

import (
	"enterprise-core/backend/pkg/logger"
	"sync"
	"time"
)

type Subsystem string

const (
	IngestSubsystem Subsystem = "ingest"
	SearchSubsystem Subsystem = "search"
	AlertSubsystem  Subsystem = "alerting"
)

// Governor manages emergency system states
type Governor struct {
	killSwitches map[Subsystem]bool
	lockdown     bool
	logger       *logger.Logger
	mu           sync.RWMutex
}

func NewGovernor(l *logger.Logger) *Governor {
	return &Governor{
		killSwitches: make(map[Subsystem]bool),
		logger:       l,
	}
}

// ToggleKillSwitch enables or disables a subsystem
func (g *Governor) ToggleKillSwitch(sub Subsystem, active bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.killSwitches[sub] = active
	g.logger.Warn("Subsystem kill-switch toggled", "subsystem", sub, "active", active)
}

// IsDisabled checks if a subsystem is currently halted
func (g *Governor) IsDisabled(sub Subsystem) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.lockdown {
		return true
	}
	return g.killSwitches[sub]
}

// EnterEmergencyLockdown halts all system operations
func (g *Governor) EnterEmergencyLockdown() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.lockdown = true
	g.logger.Warn("EMERGENCY LOCKDOWN INITIATED", "timestamp", time.Now())
}

// ReleaseLockdown restores normal operation settings
func (g *Governor) ReleaseLockdown() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.lockdown = false
	g.logger.Info("Emergency lockdown released")
}

// GetStatus returns the current state of all controls
func (g *Governor) GetStatus() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	status := map[string]interface{}{
		"lockdown": g.lockdown,
		"switches": make(map[string]bool),
	}
	for k, v := range g.killSwitches {
		status["switches"].(map[string]bool)[string(k)] = v
	}
	return status
}
