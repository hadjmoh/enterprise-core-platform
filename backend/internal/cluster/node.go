package cluster

import (
	"sync"
	"time"
)

type NodeRole string

const (
	RoleIndexer    NodeRole = "indexer"
	RoleSearchHead NodeRole = "search_head"
	RoleAll        NodeRole = "all"
)

// NodeStatus used in NodeManager is now shared with monitor.go

// Node represents a cluster member
type Node struct {
	ID             string     `json:"id"`
	Address        string     `json:"address"`
	Role           NodeRole   `json:"role"`
	Status         NodeStatus `json:"status"`
	LastSeen       time.Time  `json:"last_seen"`
	CPUUsage       float32    `json:"cpu_usage"`
	MemUsage       float32    `json:"mem_usage"`
	ActiveSearches int        `json:"active_searches"`
}

// NodeManager tracks active peers in the cluster
type NodeManager struct {
	nodes map[string]*Node
	mu    sync.RWMutex
}

func NewNodeManager() *NodeManager {
	return &NodeManager{
		nodes: make(map[string]*Node),
	}
}

// UpdateNode registers or updates a node's status
func (m *NodeManager) UpdateNode(n *Node) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if existing, ok := m.nodes[n.ID]; ok {
		existing.Status = n.Status
		existing.LastSeen = time.Now()
		existing.CPUUsage = n.CPUUsage
		existing.MemUsage = n.MemUsage
		existing.Address = n.Address
	} else {
		n.LastSeen = time.Now()
		m.nodes[n.ID] = n
	}
}

// GetNodes returns all known nodes
func (m *NodeManager) GetNodes() []Node {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	nodes := make([]Node, 0, len(m.nodes))
	for _, n := range m.nodes {
		nodes = append(nodes, *n)
	}
	return nodes
}

// GetNode returns a specific node by ID
func (m *NodeManager) GetNode(id string) (*Node, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	n, ok := m.nodes[id]
	if !ok {
		return nil, false
	}
	return n, true
}

// MarkOffline flags nodes that haven't been seen recently
func (m *NodeManager) MarkOffline(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now()
	for _, n := range m.nodes {
		if now.Sub(n.LastSeen) > timeout {
			n.Status = NodeOffline
		}
	}
}
