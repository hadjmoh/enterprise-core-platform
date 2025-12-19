package storage

import (
	"enterprise-core/backend/internal/buffer"
	"sync"
)

// Tombstone constants for event marking
const (
	TombstoneField = "_is_tombstone"
)

// IsTombstone checks if an event is a deletion marker
func IsTombstone(event buffer.Event) bool {
	v, ok := event.Data[TombstoneField]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

// Compactor defines the interface for merging SSTables and removing tombstones.
type Compactor interface {
	Trigger(partitionPath string) error
	Stats() CompactionStats
}

type CompactionStats struct {
	FilesMergedTotal uint64
	BytesSavedTotal  uint64
	LastRunTimestamp string
}

// DefaultCompactor is a stub implementation
type DefaultCompactor struct {
	mu sync.Mutex
}

func (c *DefaultCompactor) Trigger(partitionPath string) error {
	return nil // To be implemented in later sessions
}

func (c *DefaultCompactor) Stats() CompactionStats {
	return CompactionStats{}
}
