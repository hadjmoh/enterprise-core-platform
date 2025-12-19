package index

import (
	"sync"
)

// IndexMetadata contains metadata about an inverted index file.
type IndexMetadata struct {
	MinTimestamp     string `json:"min_timestamp"`
	MaxTimestamp     string `json:"max_timestamp"`
	TokenCount       int    `json:"token_count"`
	EventCount       int    `json:"event_count"`
	SchemaVersion    int    `json:"schema_version"`
	TokenizerVersion int    `json:"tokenizer_version"`
}

// InvertedIndex maps tokens to event locations/IDs
type InvertedIndex struct {
	// Map of token -> list of event sequences/offsets
	Postings map[string][]int64 `json:"postings"`
	mu       sync.RWMutex
}

// IndexMerger defines the interface for combining inverted indexes.
type IndexMerger interface {
	Merge(indices []*InvertedIndex) (*InvertedIndex, error)
}

func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		Postings: make(map[string][]int64),
	}
}

func (idx *InvertedIndex) Add(token string, sequence int64) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.Postings[token] = append(idx.Postings[token], sequence)
}

func (idx *InvertedIndex) Get(token string) []int64 {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.Postings[token]
}

func (idx *InvertedIndex) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.Postings = make(map[string][]int64)
}

func (idx *InvertedIndex) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.Postings)
}
