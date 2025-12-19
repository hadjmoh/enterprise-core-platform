package cluster

import (
	"enterprise-core/backend/internal/ha"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"sync"
)

// Coordinator manages cluster-wide operations
type Coordinator struct {
	leader     *ha.LeaderElection
	health     *ha.HealthMonitor
	nodeID     string
	peers      map[string]string // nodeID -> address
	logger     *logger.Logger
	mu         sync.RWMutex
	haMode     string // active-passive, active-active
}

func NewCoordinator(nodeID string, haMode string, leader *ha.LeaderElection, health *ha.HealthMonitor, logger *logger.Logger) *Coordinator {
	return &Coordinator{
		nodeID: nodeID,
		haMode: haMode,
		leader: leader,
		health: health,
		peers:  make(map[string]string),
		logger: logger,
	}
}

func (c *Coordinator) Start() error {
	c.logger.Info("Starting cluster coordinator", "nodeID", c.nodeID, "mode", c.haMode)
	
	// Start leader election
	go func() {
		if err := c.leader.Run(); err != nil {
			c.logger.Error("Leader election failed", err)
		}
	}()

	return nil
}

func (c *Coordinator) IsActive() bool {
	if c.haMode == "active-active" {
		return true
	}
	return c.leader.IsLeader()
}

func (c *Coordinator) AddPeer(id, addr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.peers[id] = addr
}

func (c *Coordinator) GetPeers() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	peers := make(map[string]string)
	for k, v := range c.peers {
		if k != c.nodeID {
			peers[k] = v
		}
	}
	return peers
}

func (c *Coordinator) GetClusterStatus() map[string]interface{} {
	return map[string]interface{}{
		"node_id":    c.nodeID,
		"ha_mode":    c.haMode,
		"is_leader":  c.leader.IsLeader(),
		"healthy":    c.health.IsOverallHealthy(),
		"peer_count": len(c.peers),
	}
}

func (c *Coordinator) Stop() error {
	c.logger.Info("Stopping cluster coordinator", "nodeID", c.nodeID)
	return c.leader.Resign()
}
