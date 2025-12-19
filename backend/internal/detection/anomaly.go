package detection

import (
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/pkg/logger"
	"math"
	"sync"
	"time"
)

// AnomalyDetector identifies statistical anomalies in event streams
type AnomalyDetector struct {
	baselines map[string]*Baseline
	logger    *logger.Logger
	mu        sync.RWMutex
	threshold float64 // Standard deviations from mean
}

type Baseline struct {
	Metric     string
	Mean       float64
	StdDev     float64
	Count      int64
	LastUpdate time.Time
	Values     []float64 // Rolling window
	WindowSize int
}

type Anomaly struct {
	Metric    string
	Value     float64
	Expected  float64
	Deviation float64
	Severity  string
	Timestamp time.Time
}

func NewAnomalyDetector(threshold float64, logger *logger.Logger) *AnomalyDetector {
	return &AnomalyDetector{
		baselines: make(map[string]*Baseline),
		logger:    logger,
		threshold: threshold, // e.g., 3.0 for 3 standard deviations
	}
}

func (a *AnomalyDetector) ProcessEvent(event buffer.Event) []*Anomaly {
	a.mu.Lock()
	defer a.mu.Unlock()

	anomalies := make([]*Anomaly, 0)

	// Extract numeric metrics from event
	metrics := a.extractMetrics(event.Data)

	for metric, value := range metrics {
		baseline := a.getOrCreateBaseline(metric)
		
		// Check if value is anomalous
		if baseline.Count > 10 { // Need minimum samples
			deviation := math.Abs(value-baseline.Mean) / baseline.StdDev
			
			if deviation > a.threshold {
				severity := a.calculateSeverity(deviation)
				anomaly := &Anomaly{
					Metric:    metric,
					Value:     value,
					Expected:  baseline.Mean,
					Deviation: deviation,
					Severity:  severity,
					Timestamp: time.Now(),
				}
				anomalies = append(anomalies, anomaly)
				a.logger.Info("Anomaly detected", 
					"metric", metric, 
					"value", value, 
					"expected", baseline.Mean,
					"deviation", deviation,
					"severity", severity)
			}
		}

		// Update baseline
		a.updateBaseline(baseline, value)
	}

	return anomalies
}

func (a *AnomalyDetector) extractMetrics(data map[string]interface{}) map[string]float64 {
	metrics := make(map[string]float64)

	// Extract common numeric fields
	numericFields := []string{"bytes", "duration", "response_time", "cpu", "memory", "connections"}
	
	for _, field := range numericFields {
		if val, ok := data[field]; ok {
			if fval, ok := toFloat64(val); ok {
				metrics[field] = fval
			}
		}
	}

	return metrics
}

func (a *AnomalyDetector) getOrCreateBaseline(metric string) *Baseline {
	if baseline, ok := a.baselines[metric]; ok {
		return baseline
	}

	baseline := &Baseline{
		Metric:     metric,
		Mean:       0,
		StdDev:     1,
		Count:      0,
		LastUpdate: time.Now(),
		Values:     make([]float64, 0, 1000),
		WindowSize: 1000,
	}
	a.baselines[metric] = baseline
	return baseline
}

func (a *AnomalyDetector) updateBaseline(baseline *Baseline, value float64) {
	// Add to rolling window
	baseline.Values = append(baseline.Values, value)
	if len(baseline.Values) > baseline.WindowSize {
		baseline.Values = baseline.Values[1:]
	}

	baseline.Count++
	baseline.LastUpdate = time.Now()

	// Recalculate mean and standard deviation
	if len(baseline.Values) > 0 {
		baseline.Mean = calculateMean(baseline.Values)
		baseline.StdDev = calculateStdDev(baseline.Values, baseline.Mean)
	}
}

func (a *AnomalyDetector) calculateSeverity(deviation float64) string {
	if deviation > 5.0 {
		return "critical"
	} else if deviation > 4.0 {
		return "high"
	} else if deviation > 3.0 {
		return "medium"
	}
	return "low"
}

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 1.0
	}
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	return math.Sqrt(variance)
}
