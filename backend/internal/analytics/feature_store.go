package analytics

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// FeatureValue represents the value of a feature
type FeatureValue struct {
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"` // For averages
}

// FeatureStore manages feature vectors for entities
type FeatureStore struct {
	mu       sync.RWMutex
	Features map[string]map[string]*FeatureValue // entityID -> featureName -> Value
	filepath string
}

// NewFeatureStore creates a new in-memory feature store
func NewFeatureStore(filepath string) *FeatureStore {
	fs := &FeatureStore{
		Features: make(map[string]map[string]*FeatureValue),
		filepath: filepath,
	}
	// Try loading from disk
	if err := fs.load(); err != nil {
		fmt.Printf("Warning: Failed to load feature store: %v\n", err)
	}
	
	// Periodic snapshot
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			fs.Snapshot()
		}
	}()
	
	return fs
}

// GetFeature retrieves a feature value
func (fs *FeatureStore) GetFeature(entityID, featureName string) (float64, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if entityFeatures, ok := fs.Features[entityID]; ok {
		if val, ok := entityFeatures[featureName]; ok {
			return val.Value, true
		}
	}
	return 0, false
}

// UpdateFeature updates a feature value (e.g., set or aggregate)
func (fs *FeatureStore) UpdateFeature(entityID, featureName string, value float64, operation string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, ok := fs.Features[entityID]; !ok {
		fs.Features[entityID] = make(map[string]*FeatureValue)
	}

	fv, ok := fs.Features[entityID][featureName]
	if !ok {
		fv = &FeatureValue{Timestamp: time.Now()}
		fs.Features[entityID][featureName] = fv
	}

	switch operation {
	case "set":
		fv.Value = value
		fv.Count = 1
	case "sum", "inc":
		fv.Value += value
		fv.Count++
	case "avg":
		// Rolling average: new_avg = ((old_avg * count) + new_val) / (count + 1)
		fv.Value = ((fv.Value * float64(fv.Count)) + value) / float64(fv.Count+1)
		fv.Count++
	case "max":
		if value > fv.Value {
			fv.Value = value
		}
	}
	fv.Timestamp = time.Now()
}

// Snapshot persists the store to disk
func (fs *FeatureStore) Snapshot() error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, err := json.MarshalIndent(fs.Features, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.filepath, data, 0644)
}

func (fs *FeatureStore) load() error {
	data, err := os.ReadFile(fs.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &fs.Features)
}

// GetAllFeatures returns all features for an entity (for debugging/API)
func (fs *FeatureStore) GetAllFeatures(entityID string) map[string]float64 {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	result := make(map[string]float64)
	if features, ok := fs.Features[entityID]; ok {
		for k, v := range features {
			result[k] = v.Value
		}
	}
	return result
}
