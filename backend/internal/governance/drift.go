package governance

import (
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"sync"
	"time"
)

// DriftReport captures differences between runtime and golden config
type DriftReport struct {
	Timestamp     time.Time `json:"timestamp"`
	DriftDetected bool      `json:"drift_detected"`
	Discrepancies []string  `json:"discrepancies"`
}

// DriftDetector monitors runtime configuration integrity
type DriftDetector struct {
	goldenConfig map[string]interface{}
	runtime      map[string]interface{}
	logger       *logger.Logger
	mu           sync.RWMutex
}

func NewDriftDetector(golden map[string]interface{}, l *logger.Logger) *DriftDetector {
	return &DriftDetector{
		goldenConfig: golden,
		runtime:      make(map[string]interface{}),
		logger:       l,
	}
}

// UpdateRuntime records the current active configuration
func (d *DriftDetector) UpdateRuntime(config map[string]interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.runtime = config
}

// Check compares runtime config against golden state
func (d *DriftDetector) Check() *DriftReport {
	d.mu.RLock()
	defer d.mu.RUnlock()

	report := &DriftReport{
		Timestamp:     time.Now(),
		Discrepancies: make([]string, 0),
	}

	for k, goldenVal := range d.goldenConfig {
		runtimeVal, ok := d.runtime[k]
		if !ok {
			report.Discrepancies = append(report.Discrepancies, fmt.Sprintf("Missing configuration key: %s", k))
			continue
		}

		if fmt.Sprintf("%v", goldenVal) != fmt.Sprintf("%v", runtimeVal) {
			report.Discrepancies = append(report.Discrepancies, 
				fmt.Sprintf("Drift in %s: expected %v, got %v", k, goldenVal, runtimeVal))
		}
	}

	report.DriftDetected = len(report.Discrepancies) > 0
	if report.DriftDetected {
		d.logger.Warn("Configuration drift detected", "discrepancies", len(report.Discrepancies))
	}

	return report
}
