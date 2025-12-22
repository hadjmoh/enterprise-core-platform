package analytics

import (
	"math"
)

// AnomalyScore represents the result of anomaly detection
type AnomalyScore struct {
	Score       float64 `json:"score"`        // Z-Score or deviation magnitude
	IsAnomaly   bool    `json:"is_anomaly"`   // True if Score > Threshold
	Description string  `json:"description"`  // "High bytes out"
	FeatureName string  `json:"feature_name"` // "bytes_out_1h"
	Value       float64 `json:"value"`        // Current value
	Baseline    float64 `json:"baseline"`     // Expected mean/median
}

// AnomalyDetector uses statistical methods to identify outliers
type AnomalyDetector struct {
	ZScoreThreshold float64
	MinSamples      int
}

// NewAnomalyDetector creates a new detector
func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{
		ZScoreThreshold: 3.0, // Standard deviation cutoff
		MinSamples:      5,   // Need some history
	}
}

// Detect checks a feature store for anomalies
// This is a simplified "snapshot" detection. Real detection would track history per entity.
// For this PoC, we will compare an entity's current value against the population mean 
// (or assume the FeatureStore value *is* the rolling average/stat if we had historical time-series).
//
// Refined Approach for PoC:
// We will treat the FeatureStore 'Value' as the "Current Value".
// We need a baseline. Since we built a simple KV store, we don't have deep history per key.
// We will implement "Population Analysis": Calculate mean/stddev across ALL entities for a feature.
// Then flag entities that are outliers compared to the population.
func (d *AnomalyDetector) DetectPopulationAnomalies(store *FeatureStore, featureName string) []AnomalyScore {
	// 1. Gather all values for the feature
	values := []float64{}
	entityIDs := []string{}

	allFeatures := store.Features
	for entityID, features := range allFeatures {
		if val, ok := features[featureName]; ok {
			values = append(values, val.Value)
			entityIDs = append(entityIDs, entityID)
		}
	}

	if len(values) < d.MinSamples {
		return nil
	}

	// 2. Calculate Mean and StdDev
	mean, stdDev := calculateStats(values)
	if stdDev == 0 {
		return nil
	}

	// 3. Score each entity
	anomalies := []AnomalyScore{}
	for i, val := range values {
		zScore := (val - mean) / stdDev
		
		if math.Abs(zScore) > d.ZScoreThreshold {
			anomalies = append(anomalies, AnomalyScore{
				Score:       zScore,
				IsAnomaly:   true,
				Description: fmt.Sprintf("Value %.2f is %.2f sigmas from mean %.2f", val, zScore, mean),
				FeatureName: featureName,
				Value:       val,
				Baseline:    mean,
			})
			fmt.Printf("ANOMALY DETECTED: Entity=%s Feature=%s Z=%.2f\n", entityIDs[i], featureName, zScore)
		}
	}

	return anomalies
}

func calculateStats(values []float64) (float64, float64) {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	varianceSum := 0.0
	for _, v := range values {
		varianceSum += math.Pow(v-mean, 2)
	}
	variance := varianceSum / float64(len(values))
	stdDev := math.Sqrt(variance)

	return mean, stdDev
}
