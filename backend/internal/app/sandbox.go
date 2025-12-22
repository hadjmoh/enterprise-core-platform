package app

import (
	"fmt"
	"sync"
	"time"
)

// ResourceLimits defines resource constraints for an app
type ResourceLimits struct {
	CPUPercent    float64 // Percentage of one CPU core (0.5 = 50%)
	MemoryMB      int64   // Memory limit in megabytes
	DiskQuotaMB   int64   // Disk quota in megabytes
	NetworkRateMB int64   // Network rate limit in MB/s
}

// DefaultLimits returns sensible default resource limits
func DefaultLimits() ResourceLimits {
	return ResourceLimits{
		CPUPercent:    0.5,  // 50% of one core
		MemoryMB:      512,  // 512 MB
		DiskQuotaMB:   1024, // 1 GB
		NetworkRateMB: 10,   // 10 MB/s
	}
}

// Sandbox represents an isolated execution environment for an app
type Sandbox struct {
	mu           sync.RWMutex
	appName      string
	limits       ResourceLimits
	capabilities *CapabilitySet
	active       bool
	createdAt    time.Time
}

// NewSandbox creates a new sandbox for an app
func NewSandbox(appName string, limits ResourceLimits, caps []Capability) (*Sandbox, error) {
	capSet := NewCapabilitySet()
	for _, cap := range caps {
		if err := capSet.Grant(cap); err != nil {
			return nil, fmt.Errorf("failed to grant capability %s: %w", cap, err)
		}
	}

	return &Sandbox{
		appName:      appName,
		limits:       limits,
		capabilities: capSet,
		active:       true,
		createdAt:    time.Now(),
	}, nil
}

// CheckPermission verifies if the sandbox has a specific capability
func (s *Sandbox) CheckPermission(cap Capability) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.active {
		return fmt.Errorf("sandbox is not active")
	}

	if !s.capabilities.Has(cap) {
		return fmt.Errorf("permission denied: app does not have capability %s", cap)
	}

	return nil
}

// GetLimits returns the current resource limits
func (s *Sandbox) GetLimits() ResourceLimits {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.limits
}

// UpdateLimits modifies the resource limits
func (s *Sandbox) UpdateLimits(limits ResourceLimits) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limits = limits
}

// GetCapabilities returns all granted capabilities
func (s *Sandbox) GetCapabilities() []Capability {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.capabilities.List()
}

// Deactivate marks the sandbox as inactive
func (s *Sandbox) Deactivate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = false
}

// IsActive returns whether the sandbox is active
func (s *Sandbox) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

// GetAppName returns the app name
func (s *Sandbox) GetAppName() string {
	return s.appName
}

// GetCreatedAt returns when the sandbox was created
func (s *Sandbox) GetCreatedAt() time.Time {
	return s.createdAt
}

// Validate checks if the limits are within acceptable ranges
func (rl *ResourceLimits) Validate() error {
	if rl.CPUPercent <= 0 || rl.CPUPercent > 2.0 {
		return fmt.Errorf("CPU limit must be between 0 and 2.0 (200%%)")
	}
	if rl.MemoryMB <= 0 || rl.MemoryMB > 4096 {
		return fmt.Errorf("memory limit must be between 0 and 4096 MB")
	}
	if rl.DiskQuotaMB <= 0 || rl.DiskQuotaMB > 10240 {
		return fmt.Errorf("disk quota must be between 0 and 10240 MB (10 GB)")
	}
	if rl.NetworkRateMB < 0 || rl.NetworkRateMB > 100 {
		return fmt.Errorf("network rate must be between 0 and 100 MB/s")
	}
	return nil
}
