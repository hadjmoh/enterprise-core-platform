package analytics

import (
	"encoding/json"
	"enterprise-core/backend/internal/buffer"
	"math"
	"time"
)

// FeatureExtractor updates the feature store based on incoming events
type FeatureExtractor struct {
	store *FeatureStore
}

// NewFeatureExtractor creates a new feature extractor
func NewFeatureExtractor(store *FeatureStore) *FeatureExtractor {
	return &FeatureExtractor{
		store: store,
	}
}

// ProcessEvent extracts features from an event
func (e *FeatureExtractor) ProcessEvent(event buffer.Event) {
	// Parse event data to extract relevant fields
	// This is highly dependent on the event schema.
	// We'll use a generic map for flexibility in this prototype.
	
	// 1. User Behavioral Features
	if user, ok := event.Data["user"].(string); ok && user != "" {
		// Count logins
		if action, ok := event.Data["action"].(string); ok && action == "login" {
			e.store.UpdateFeature("user:"+user, "login_count_1h", 1, "inc")
		}
		
		// Resource access
		if _, ok := event.Data["resource"].(string); ok {
			e.store.UpdateFeature("user:"+user, "resource_access_count_1h", 1, "inc")
		}
	}

	// 2. IP/Network Features
	if srcIP, ok := event.Data["src_ip"].(string); ok && srcIP != "" {
		// Bytes out
		if bytesOut, ok := event.Data["bytes_out"].(float64); ok {
			e.store.UpdateFeature("ip:"+srcIP, "bytes_out_1h", bytesOut, "sum")
		} else if bytesOutInt, ok := event.Data["bytes_out"].(int); ok {
			e.store.UpdateFeature("ip:"+srcIP, "bytes_out_1h", float64(bytesOutInt), "sum")
		}
	}

	// 3. Asset/Infrastructure Features
	if host, ok := event.Data["host"].(string); ok && host != "" {
		// CPU Usage (requires specialized event type or parsing)
		if cpu, ok := event.Data["cpu_usage"].(float64); ok {
			e.store.UpdateFeature("asset:"+host, "cpu_avg_1h", cpu, "avg")
		}
	}
}
