package storage

import (
	"context"
	"enterprise-core/backend/internal/buffer"
	"time"
)

// StorageEngine defines the interface for all storage tiers (Hot, Warm, Cold)
type StorageEngine interface {
	// Write inserts an event into the storage engine
	Write(ctx context.Context, event buffer.Event) error
	
	// Query searches for events matching the criteria
	Query(ctx context.Context, query Query) ([]buffer.Event, error)
	
	// Stats returns storage usage and performance metrics
	Stats() Stats
	
	// NewReader creates a reader for the given query
	NewReader(ctx context.Context, query Query) (StorageReader, error)
	
	// Close gracefully shuts down the storage engine
	Close() error
}

// StorageReader defines the interface for streaming events from storage
type StorageReader interface {
	Next(ctx context.Context) (buffer.Event, error)
	Close() error
}

// PartitionState represents the current lifecycle status of a data shard
type PartitionState string

const (
	StateOpen       PartitionState = "OPEN"
	StateSealed     PartitionState = "SEALED"
	StateCompacting PartitionState = "COMPACTING"
	StateArchived   PartitionState = "ARCHIVED"
	StateWarm       PartitionState = "WARM"
	StateCold       PartitionState = "COLD"
	StateDeleted    PartitionState = "DELETED"
)

// Query defines the search criteria for the storage engine
type Query struct {
	StartTime    time.Time
	EndTime      time.Time
	Keywords     []string
	FieldFilters map[string]string
	Limit        int
	Offset       int
}

// Stats provides metrics about the storage engine
type Stats struct {
	EventCount     uint64
	DiskUsageBytes uint64
	Uptime         time.Duration
}
