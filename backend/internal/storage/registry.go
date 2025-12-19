package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Registry tracks all SSTables within a storage partition.
type Registry struct {
	basePath string
	sstables []SSTableInfo
	mu       sync.RWMutex
}

// SSTableInfo contains registry metadata for a specific SSTable.
type SSTableInfo struct {
	Filename     string `json:"filename"`
	MinTimestamp string `json:"min_timestamp"`
	MaxTimestamp string `json:"max_timestamp"`
	EventCount   int64  `json:"event_count"`
}

func NewRegistry(partitionPath string) *Registry {
	return &Registry{
		basePath: partitionPath,
		sstables: make([]SSTableInfo, 0),
	}
}

// Load reads the registry from disk if it exists.
func (r *Registry) Load() error {
	path := filepath.Join(r.basePath, "registry.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return json.Unmarshal(data, &r.sstables)
}

// Add appends a new SSTable to the registry and persists it.
func (r *Registry) Add(info SSTableInfo) error {
	r.mu.Lock()
	r.sstables = append(r.sstables, info)
	r.mu.Unlock()
	return r.Save()
}

// Save persists the registry to disk atomically.
func (r *Registry) Save() error {
	r.mu.RLock()
	data, err := json.MarshalIndent(r.sstables, "", "  ")
	r.mu.RUnlock()
	if err != nil {
		return err
	}

	path := filepath.Join(r.basePath, "registry.json")
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// List returns all active SSTables in sequence.
func (r *Registry) List() []SSTableInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var list []SSTableInfo
	for _, info := range r.sstables {
		list = append(list, info)
	}
	return list
}
