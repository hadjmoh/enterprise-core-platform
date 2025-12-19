package index

import (
	"sort"
)

// SparseIndex stores periodic (timestamp, sequence) mappings for fast time-range jumps.
type SparseIndex struct {
	Metadata SparseMetadata `json:"metadata"`
	Entries  []SparseEntry  `json:"entries"`
}

type SparseMetadata struct {
	EventCount    int64  `json:"event_count"`
	MinTimestamp  string `json:"min_timestamp"`
	MaxTimestamp  string `json:"max_timestamp"`
	IndexVersion  int    `json:"index_version"`
	SchemaVersion int    `json:"schema_version"`
}

type SparseEntry struct {
	Timestamp string `json:"timestamp"`
	Sequence  int64  `json:"sequence"`
}

func NewSparseIndex() *SparseIndex {
	return &SparseIndex{
		Metadata: SparseMetadata{
			IndexVersion:  1,
			SchemaVersion: 1,
		},
		Entries: make([]SparseEntry, 0),
	}
}

func (si *SparseIndex) Add(timestamp string, sequence int64) {
	si.Metadata.EventCount++
	if si.Metadata.MinTimestamp == "" || timestamp < si.Metadata.MinTimestamp {
		si.Metadata.MinTimestamp = timestamp
	}
	if si.Metadata.MaxTimestamp == "" || timestamp > si.Metadata.MaxTimestamp {
		si.Metadata.MaxTimestamp = timestamp
	}

	si.Entries = append(si.Entries, SparseEntry{
		Timestamp: timestamp,
		Sequence:  sequence,
	})
}

// FindNearest returns the sequence number of the entry closest to (but not after) the given timestamp.
func (si *SparseIndex) FindNearest(timestamp string) int64 {
	if len(si.Entries) == 0 {
		return 0
	}
	
	idx := sort.Search(len(si.Entries), func(i int) bool {
		return si.Entries[i].Timestamp >= timestamp
	})

	if idx == 0 {
		return si.Entries[0].Sequence
	}
	if idx == len(si.Entries) {
		return si.Entries[len(si.Entries)-1].Sequence
	}
	
	// If exact match found or first entry > timestamp, return that or previous.
	// For simplicity, return the entry just before or at.
	return si.Entries[idx-1].Sequence
}
