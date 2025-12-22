package cluster

import (
	"context"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"sync"
	"time"
)

// ClusterMaster orchestrates the cluster topology
type ClusterMaster struct {
	nodeMgr   *NodeManager
	shardMgr  *ShardManager
	logger    *logger.Logger
	stopChan  chan struct{}
	mu        sync.Mutex
}

func NewClusterMaster(nodeMgr *NodeManager, shardMgr *ShardManager, l *logger.Logger) *ClusterMaster {
	return &ClusterMaster{
		nodeMgr:  nodeMgr,
		shardMgr: shardMgr,
		logger:   l,
		stopChan: make(chan struct{}),
	}
}

func (m *ClusterMaster) Start(ctx context.Context) {
	m.logger.Info("Cluster Master starting")
	ticker := time.NewTicker(10 * time.Second)
	
	go func() {
		for {
			select {
			case <-ticker.C:
				m.rebalanceShards()
				m.checkHealth()
			case <-m.stopChan:
				ticker.Stop()
				return
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (m *ClusterMaster) Stop() {
	close(m.stopChan)
}

func (m *ClusterMaster) rebalanceShards() {
	m.mu.Lock()
	defer m.mu.Unlock()

	nodes := m.nodeMgr.GetNodes()
	var indexers []Node
	for _, n := range nodes {
		if (n.Role == RoleIndexer || n.Role == RoleAll) && n.Status == NodeOnline {
			indexers = append(indexers, n)
		}
	}

	if len(indexers) == 0 {
		return
	}

	// Simple round-robin for newly created shards (In a real system, shards are created by data ingestion)
	// For PoC, let's ensure at least 3 shards exist and are assigned
	for i := 1; i <= 3; i++ {
		shardID := fmt.Sprintf("shard-%02d", i)
		if _, ok := m.shardMgr.GetShard(shardID); !ok {
			// Allocate to least loaded or round-robin
			target := indexers[(i-1)%len(indexers)]
			m.shardMgr.UpdateShard(&Shard{
				ID:         shardID,
				NodeIDs:    []string{target.ID},
				Status:     ShardActive,
				RangeStart: int64(i * 1000),
				RangeEnd:   int64((i + 1) * 1000),
			})
			m.logger.Info("Allocated new shard", "shard_id", shardID, "node_id", target.ID)
		}
	}
}

func (m *ClusterMaster) checkHealth() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Detect offline nodes and trigger recovery
	shards := m.shardMgr.GetShards()
	nodes := m.nodeMgr.GetNodes()
	
	nodeMap := make(map[string]Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	for _, s := range shards {
		healthyCopies := 0
		var offlineNode string
		for _, nid := range s.NodeIDs {
			if n, ok := nodeMap[nid]; ok && n.Status == NodeOnline {
				healthyCopies++
			} else {
				offlineNode = nid
			}
		}

		if healthyCopies == 0 {
			m.logger.Error("CRITICAL: Shard is data-missing (all copies offline)", fmt.Errorf("shard %s has no healthy nodes", s.ID))
			s.Status = ShardOffline
			m.shardMgr.UpdateShard(&s)
		} else if healthyCopies < len(s.NodeIDs) || len(s.NodeIDs) == 1 { // Simple case: expect 1 copy for now
			// Recovery: pick a new healthy node if possible
			m.recoverShard(&s, offlineNode)
		}
	}
}

func (m *ClusterMaster) recoverShard(s *Shard, failedNodeID string) {
	nodes := m.nodeMgr.GetNodes()
	var candidates []Node
	for _, n := range nodes {
		if (n.Role == RoleIndexer || n.Role == RoleAll) && n.Status == NodeOnline {
			// Don't pick a node that already has this shard
			alreadyHas := false
			for _, nid := range s.NodeIDs {
				if nid == n.ID {
					alreadyHas = true
					break
				}
			}
			if !alreadyHas {
				candidates = append(candidates, n)
			}
		}
	}

	if len(candidates) > 0 {
		// Pick first available candidate
		target := candidates[0]
		if failedNodeID != "" {
			// Replace failed node
			for i, nid := range s.NodeIDs {
				if nid == failedNodeID {
					s.NodeIDs[i] = target.ID
					break
				}
			}
		} else if len(s.NodeIDs) < 1 { // Ensure at least 1
			s.NodeIDs = append(s.NodeIDs, target.ID)
		}
		
		s.Status = ShardActive
		m.shardMgr.UpdateShard(s)
		m.logger.Warn("Triggered shard recovery", "shard_id", s.ID, "new_node", target.ID)
	}
}
