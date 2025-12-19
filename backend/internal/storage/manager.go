package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PartitionMetadata stores information about a specific time shard
type PartitionMetadata struct {
	State           PartitionState `json:"state"`
	TargetTier      PartitionState `json:"target_tier,omitempty"`
	EventCount      int64          `json:"event_count"`
	SizeInBytes     int64          `json:"size_in_bytes"`
	LastModifiedAt  time.Time      `json:"last_modified_at"`
	LastMigrationAt time.Time      `json:"last_migration_at,omitempty"`
	DeleteAt        time.Time      `json:"delete_at,omitempty"`
	SchemaVersion   int            `json:"schema_version"`
}

// ShardMetadata tracks health and content of a virtual shard
type ShardMetadata struct {
	ShardID          int       `json:"shard_id"`
	ParentPartition  string    `json:"parent_partition"`
	EventCount       int64     `json:"event_count"`
	MinTimestamp     string    `json:"min_timestamp"`
	MaxTimestamp     string    `json:"max_timestamp"`
	RegistryVersion  int       `json:"registry_version"`
	LastModifiedAt   time.Time `json:"last_modified_at"`
}

// PartitionManager handles time-based data sharding on disk
type PartitionManager struct {
	basePath string
	mu       sync.RWMutex
}

func NewPartitionManager(basePath string) *PartitionManager {
	return &PartitionManager{
		basePath: basePath,
	}
}

// GetRelativePath returns the standard YYYY/MM/DD/HH structure
func (m *PartitionManager) GetRelativePath(t time.Time) string {
	return filepath.Join(
		t.Format("2006"),
		t.Format("01"),
		t.Format("02"),
		t.Format("15"),
	)
}

// GetPartitionPath returns the directory path for a given timestamp
// Format: basePath/YYYY/MM/DD/HH
func (m *PartitionManager) GetPartitionPath(t time.Time) string {
	t = t.UTC()
	return filepath.Join(
		m.basePath,
		t.Format("2006"),
		t.Format("01"),
		t.Format("02"),
		t.Format("15"),
	)
}

// GetShardPath returns the path for a specific shard within a partition
func (m *PartitionManager) GetShardPath(t time.Time, shardID int) string {
	return filepath.Join(m.GetPartitionPath(t), fmt.Sprintf("shard_%02d", shardID))
}

func (m *PartitionManager) ListPartitionsInTimerange(start, end time.Time) []string {
	var paths []string
	
	// Force UTC for directory traversal
	start = start.UTC()
	end = end.UTC()
	
	// Basic implementation: iterate hourly from start to end
	current := time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), 0, 0, 0, time.UTC)
	for current.Before(end) || current.Equal(end) {
		paths = append(paths, m.GetPartitionPath(current))
		current = current.Add(time.Hour)
	}
	
	return paths
}

// GetPartitions returns the paths of existing partitions between start and end
func (m *PartitionManager) GetPartitions(start, end time.Time, tierPaths ...string) []string {
	allRels := m.ListPartitionsInTimerange(start, end)
	var existing []string
	
	searchPaths := []string{m.basePath}
	searchPaths = append(searchPaths, tierPaths...)

	for _, rel := range allRels {
		subPath, _ := filepath.Rel(m.basePath, rel)
		for _, base := range searchPaths {
			p := filepath.Join(base, subPath)
			if info, err := os.Stat(p); err == nil && info.IsDir() {
				existing = append(existing, p)
				break
			}
		}
	}
	return existing
}

// EnsurePartition creates the necessary directories and initializes metadata if missing
func (m *PartitionManager) EnsurePartition(t time.Time) error {
	path := m.GetPartitionPath(t)
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	metaPath := filepath.Join(path, "metadata.json")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		initialMeta := PartitionMetadata{
			State:          StateOpen,
			SchemaVersion:  1,
			LastModifiedAt: time.Now(),
		}
		data, _ := json.MarshalIndent(initialMeta, "", "  ")
		return os.WriteFile(metaPath, data, 0644)
	}
	return nil
}

// GetMetadata returns the metadata for a partition
func (m *PartitionManager) GetMetadata(t time.Time) (*PartitionMetadata, error) {
	path := filepath.Join(m.GetPartitionPath(t), "metadata.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var meta PartitionMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// SetState updates the partition state atomically
func (m *PartitionManager) SetState(t time.Time, state PartitionState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta, err := m.GetMetadata(t)
	if err != nil {
		return err
	}

	meta.State = state
	return m.writeMetadataAtomic(t, meta)
}

func (m *PartitionManager) writeMetadataAtomic(t time.Time, meta *PartitionMetadata) error {
	path := filepath.Join(m.GetPartitionPath(t), "metadata.json")
	tmpPath := path + ".tmp"

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	// Ensure data is synced to disk (fsync)
	f, err := os.Open(tmpPath)
	if err == nil {
		f.Sync()
		f.Close()
	}

	return os.Rename(tmpPath, path)
}

// UpdateMetadata updates the metadata file for a partition atomically
func (m *PartitionManager) UpdateMetadata(t time.Time, eventSize int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta, err := m.GetMetadata(t)
	if err != nil {
		// If fails to read, we might be initialization phase, EnsurePartition should have run
		return err
	}

	meta.EventCount++
	meta.SizeInBytes += int64(eventSize)
	meta.LastModifiedAt = time.Now()

	return m.writeMetadataAtomic(t, meta)
}

// RebuildRegistry reconstructs the registry.json by scanning SSTable footers.
func (m *PartitionManager) RebuildRegistry(t time.Time) error {
	path := m.GetPartitionPath(t)
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	reg := NewRegistry(path)
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".jsonl" {
			filePath := filepath.Join(path, f.Name())
			it, err := NewSSTableIterator(filePath)
			if err != nil {
				continue
			}
			reg.Add(SSTableInfo{
				Filename:     f.Name(),
				MinTimestamp: it.footer.MinTimestamp,
				MaxTimestamp: it.footer.MaxTimestamp,
				EventCount:   it.footer.EventCount,
			})
			it.Close()
		}
	}
	return reg.Save()
}

// WriteShardMetadataAtomic saves shard metadata atomically
func (m *PartitionManager) WriteShardMetadataAtomic(shardPath string, meta *ShardMetadata) error {
	path := filepath.Join(shardPath, "shard_metadata.json")
	tmpPath := path + ".tmp"

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	f, err := os.Open(tmpPath)
	if err == nil {
		f.Sync()
		f.Close()
	}

	return os.Rename(tmpPath, path)
}
