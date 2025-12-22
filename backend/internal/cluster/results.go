package cluster

import (
	"encoding/json"
	"enterprise-core/backend/internal/buffer"
	"sort"
	"time"
)

// SearchResult represents a partial result set from a single node
type SearchResult struct {
	NodeID string         `json:"node_id"`
	Events []buffer.Event `json:"events"`
	Error  string         `json:"error,omitempty"`
}

// ResultMerger handles combining partial results into a final sorted set
type ResultMerger struct {
	limit int
}

func NewResultMerger(limit int) *ResultMerger {
	return &ResultMerger{
		limit: limit,
	}
}

// Merge combines multiple partial results, sorts them by time (descending), and applies the limit
func (m *ResultMerger) Merge(results []SearchResult) []buffer.Event {
	var allEvents []buffer.Event
	for _, res := range results {
		if res.Error == "" {
			allEvents = append(allEvents, res.Events...)
		}
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(allEvents, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, allEvents[i].Timestamp)
		t2, _ := time.Parse(time.RFC3339, allEvents[j].Timestamp)
		return t1.After(t2)
	})

	// Apply limit
	if m.limit > 0 && len(allEvents) > m.limit {
		return allEvents[:m.limit]
	}

	return allEvents
}

// UnmarshalEvents helper to convert gRPC JSON results back to event objects
func UnmarshalEvents(data []byte) ([]buffer.Event, error) {
	var events []buffer.Event
	err := json.Unmarshal(data, &events)
	return events, err
}
