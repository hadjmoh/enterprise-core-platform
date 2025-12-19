package ha

import (
	"enterprise-core/backend/pkg/logger"
	"net/http"
	"sync"
	"time"
)

// HealthMonitor tracks system health for HA failover
type HealthMonitor struct {
	status     map[string]HealthStatus
	logger     *logger.Logger
	mu         sync.RWMutex
	checkers   map[string]HealthChecker
}

type HealthStatus struct {
	Status    string    `json:"status"` // up, down, degraded
	LastCheck time.Time `json:"last_check"`
	Message   string    `json:"message,omitempty"`
}

type HealthChecker func() (status string, message string)

func NewHealthMonitor(logger *logger.Logger) *HealthMonitor {
	return &HealthMonitor{
		status:   make(map[string]HealthStatus),
		logger:   logger,
		checkers: make(map[string]HealthChecker),
	}
}

func (h *HealthMonitor) AddChecker(name string, checker HealthChecker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers[name] = checker
}

func (h *HealthMonitor) Run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			h.checkAll()
		}
	}()
}

func (h *HealthMonitor) checkAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for name, checker := range h.checkers {
		status, message := checker()
		h.status[name] = HealthStatus{
			Status:    status,
			LastCheck: time.Now(),
			Message:   message,
		}
		
		if status != "up" {
			h.logger.Warn("Health check failed", "component", name, "status", status, "message", message)
		}
	}
}

func (h *HealthMonitor) GetStatus() map[string]HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	status := make(map[string]HealthStatus)
	for k, v := range h.status {
		status[k] = v
	}
	return status
}

func (h *HealthMonitor) IsOverallHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, s := range h.status {
		if s.Status == "down" {
			return false
		}
	}
	return true
}

func (h *HealthMonitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.IsOverallHealthy() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	// Simplified response: just JSON status
}
