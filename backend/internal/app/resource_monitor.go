package app

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ResourceUsage tracks current resource consumption
type ResourceUsage struct {
	CPUPercent    float64
	MemoryMB      int64
	DiskUsedMB    int64
	NetworkRateMB int64
	LastUpdated   time.Time
}

// ResourceMonitor tracks resource usage for all apps
type ResourceMonitor struct {
	mu      sync.RWMutex
	usage   map[string]*ResourceUsage
	limits  map[string]ResourceLimits
	stopCh  chan struct{}
	running bool
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor() *ResourceMonitor {
	return &ResourceMonitor{
		usage:  make(map[string]*ResourceUsage),
		limits: make(map[string]ResourceLimits),
		stopCh: make(chan struct{}),
	}
}

// Start begins monitoring resources
func (rm *ResourceMonitor) Start() {
	rm.mu.Lock()
	if rm.running {
		rm.mu.Unlock()
		return
	}
	rm.running = true
	rm.mu.Unlock()

	go rm.monitorLoop()
}

// Stop halts resource monitoring
func (rm *ResourceMonitor) Stop() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.running {
		return
	}

	rm.running = false
	close(rm.stopCh)
}

// monitorLoop periodically checks resource usage
func (rm *ResourceMonitor) monitorLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rm.updateUsage()
			rm.enforceLimit()
		case <-rm.stopCh:
			return
		}
	}
}

// updateUsage refreshes resource usage stats
func (rm *ResourceMonitor) updateUsage() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Get current memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// For simplicity, we'll track aggregate usage
	// In production, you'd track per-process using OS-specific APIs
	for appName := range rm.usage {
		// Simulate resource tracking (in production, use actual process metrics)
		rm.usage[appName] = &ResourceUsage{
			CPUPercent:    0.1, // Placeholder
			MemoryMB:      int64(m.Alloc / 1024 / 1024),
			DiskUsedMB:    0, // Placeholder
			NetworkRateMB: 0, // Placeholder
			LastUpdated:   time.Now(),
		}
	}
}

// enforceLimit checks if any app exceeds limits and takes action
func (rm *ResourceMonitor) enforceLimit() {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for appName, usage := range rm.usage {
		limits, exists := rm.limits[appName]
		if !exists {
			continue
		}

		// Check memory limit
		if usage.MemoryMB > limits.MemoryMB {
			// In production, kill the process
			fmt.Printf("WARNING: App %s exceeded memory limit (%d MB > %d MB)\n",
				appName, usage.MemoryMB, limits.MemoryMB)
		}

		// Check CPU limit
		if usage.CPUPercent > limits.CPUPercent {
			fmt.Printf("WARNING: App %s exceeded CPU limit (%.2f%% > %.2f%%)\n",
				appName, usage.CPUPercent*100, limits.CPUPercent*100)
		}
	}
}

// StartMonitoring begins tracking an app
func (rm *ResourceMonitor) StartMonitoring(appName string, limits ResourceLimits) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.usage[appName] = &ResourceUsage{
		LastUpdated: time.Now(),
	}
	rm.limits[appName] = limits
}

// StopMonitoring stops tracking an app
func (rm *ResourceMonitor) StopMonitoring(appName string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.usage, appName)
	delete(rm.limits, appName)
}

// GetUsage returns current resource usage for an app
func (rm *ResourceMonitor) GetUsage(appName string) (*ResourceUsage, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	usage, exists := rm.usage[appName]
	if !exists {
		return nil, fmt.Errorf("app not being monitored: %s", appName)
	}

	// Return a copy
	return &ResourceUsage{
		CPUPercent:    usage.CPUPercent,
		MemoryMB:      usage.MemoryMB,
		DiskUsedMB:    usage.DiskUsedMB,
		NetworkRateMB: usage.NetworkRateMB,
		LastUpdated:   usage.LastUpdated,
	}, nil
}

// UpdateLimits modifies the resource limits for an app
func (rm *ResourceMonitor) UpdateLimits(appName string, limits ResourceLimits) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.limits[appName]; !exists {
		return fmt.Errorf("app not being monitored: %s", appName)
	}

	if err := limits.Validate(); err != nil {
		return fmt.Errorf("invalid limits: %w", err)
	}

	rm.limits[appName] = limits
	return nil
}

// GetAllUsage returns usage for all monitored apps
func (rm *ResourceMonitor) GetAllUsage() map[string]*ResourceUsage {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make(map[string]*ResourceUsage, len(rm.usage))
	for name, usage := range rm.usage {
		result[name] = &ResourceUsage{
			CPUPercent:    usage.CPUPercent,
			MemoryMB:      usage.MemoryMB,
			DiskUsedMB:    usage.DiskUsedMB,
			NetworkRateMB: usage.NetworkRateMB,
			LastUpdated:   usage.LastUpdated,
		}
	}
	return result
}
