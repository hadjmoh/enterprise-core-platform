package cluster

import (
	"sync"
	"time"
)

type ShardStatus string

const (
	ShardActive      ShardStatus = "active"
	ShardReplicating ShardStatus = "replicating"
	ShardOffline     ShardStatus = "offline"
)

// Shard represents a data partition in the cluster
type Shard struct {
	ID          string      `json:"id"`
	NodeIDs     []string    `json:"node_ids"` // List of nodes holding this shard
	RangeStart  int64       `json:"range_start"`
	RangeEnd    int64       `json:"range_end"`
	Status      ShardStatus `json:"status"`
	LastUpdated time.Time   `json:"last_updated"`
}

// ShardManager tracks shard distribution across the cluster
type ShardManager struct {
	shards map[string]*Shard
	mu     sync.RWMutex
}

func NewShardManager() *ShardManager {
	return &ShardManager{
		shards: make(map[string]*Shard),
	}
}

func (m *ShardManager) UpdateShard(s *Shard) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s.LastUpdated = time.Now()
	m.shards[s.ID] = s
}

func (m *ShardManager) GetShards() []Shard {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	shards := make([]Shard, 0, len(m.shards))
	for _, s := range m.shards {
		shards = append(shards, *s)
	}
	return shards
}

func (m *ShardManager) GetShard(id string) (*Shard, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	s, ok := m.shards[id]
	return s, ok
}

func (m *ShardManager) GetShardsForNode(nodeID string) []Shard {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var nodeShards []Shard
	for _, s := range m.shards {
		for _, nid := range s.NodeIDs {
			if nid == nodeID {
				nodeShards = append(nodeShards, *s)
				break
			}
		}
	}
	return nodeShards
}

func (m *ShardManager) ReassignNodeShards(oldNodeID, newNodeID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, s := range m.shards {
		for i, nid := range s.NodeIDs {
			if nid == oldNodeID {
				s.NodeIDs[i] = newNodeID
				s.LastUpdated = time.Now()
			}
		}
	}
}
