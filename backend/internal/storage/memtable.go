package storage

import (
	"enterprise-core/backend/internal/buffer"
	"sort"
	"sync"
)

// Memtable buffers events in memory before they are flushed to SSTables
type Memtable struct {
	events []buffer.Event
	mu     sync.RWMutex
	size   int64
	limit  int64
}

func NewMemtable(limitBytes int64) *Memtable {
	return &Memtable{
		events: make([]buffer.Event, 0),
		limit:  limitBytes,
	}
}

func (m *Memtable) Push(event buffer.Event) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events = append(m.events, event)
	// Approximate size calculation (simplified)
	m.size += int64(len(event.Timestamp) + len(event.Source))
	for k, v := range event.Data {
		m.size += int64(len(k))
		if s, ok := v.(string); ok {
			m.size += int64(len(s))
		}
	}

	return m.size >= m.limit
}

func (m *Memtable) GetEvents() []buffer.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Sort events by timestamp before returning for flushing
	sorted := make([]buffer.Event, len(m.events))
	copy(sorted, m.events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp < sorted[j].Timestamp
	})
	return sorted
}

func (m *Memtable) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = m.events[:0]
	m.size = 0
}

func (m *Memtable) Size() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.size
}

func (m *Memtable) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.events)
}
