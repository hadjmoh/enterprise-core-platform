package cluster

import (
	"sync"
	"time"
)

// NodeStatus represents the current state of a cluster node
type NodeStatus string

const (
	NodeOnline   NodeStatus = "operational"
	NodeOffline  NodeStatus = "offline"
	NodeDegraded NodeStatus = "degraded"
)

// ClusterNode represents a single compute/storage node in the cluster
type ClusterNode struct {
	ID        string     `json:"id"`
	Address   string     `json:"address"`
	Status    NodeStatus `json:"status"`
	Uptime    int64      `json:"uptime_seconds"`
	LastSeen  time.Time  `json:"last_seen"`
	Metrics   Metrics    `json:"metrics"`
}

// Metrics captures resource utilization for a node
type Metrics struct {
	CPUUsage    float64 `json:"cpu_usage"`    // 0.0 - 100.0
	MemUsage    float64 `json:"mem_usage"`    // 0.0 - 100.0
	DiskUsage   float64 `json:"disk_usage"`   // 0.0 - 100.0
	IngestRate  float64 `json:"ingest_rate"`  // events per second
	ErrorRate   float64 `json:"error_rate"`   // percentage
}

// ResourceMonitor tracks the status of all nodes in the cluster
type ResourceMonitor struct {
	nodes map[string]*ClusterNode
	mu    sync.RWMutex
}

func NewResourceMonitor() *ResourceMonitor {
	rm := &ResourceMonitor{
		nodes: make(map[string]*ClusterNode),
	}
	// Seed initial nodes for PoC
	rm.seedNodes()
	return rm
}

func (rm *ResourceMonitor) seedNodes() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.nodes["node-01"] = &ClusterNode{
		ID:       "node-01",
		Address:  "10.0.0.1",
		Status:   NodeOnline,
		Uptime:   86400,
		LastSeen: time.Now(),
		Metrics: Metrics{
			CPUUsage:   45.5,
			MemUsage:   62.1,
			DiskUsage:  30.0,
			IngestRate: 5000,
		},
	}
	rm.nodes["node-02"] = &ClusterNode{
		ID:       "node-02",
		Address:  "10.0.0.2",
		Status:   NodeOnline,
		Uptime:   86400,
		LastSeen: time.Now(),
		Metrics: Metrics{
			CPUUsage:   50.2,
			MemUsage:   58.4,
			DiskUsage:  28.5,
			IngestRate: 4800,
		},
	}
}

func (rm *ResourceMonitor) GetNodes() []*ClusterNode {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	nodes := make([]*ClusterNode, 0, len(rm.nodes))
	for _, n := range rm.nodes {
		nodes = append(nodes, n)
	}
	return nodes
}

func (rm *ResourceMonitor) GetClusterMetrics() Metrics {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if len(rm.nodes) == 0 {
		return Metrics{}
	}

	var totalCPU, totalMem, totalDisk, totalIngest float64
	for _, n := range rm.nodes {
		totalCPU += n.Metrics.CPUUsage
		totalMem += n.Metrics.MemUsage
		totalDisk += n.Metrics.DiskUsage
		totalIngest += n.Metrics.IngestRate
	}

	count := float64(len(rm.nodes))
	return Metrics{
		CPUUsage:   totalCPU / count,
		MemUsage:   totalMem / count,
		DiskUsage:  totalDisk / count,
		IngestRate: totalIngest, // Aggregated ingest
	}
}

func (rm *ResourceMonitor) AddNode(node *ClusterNode) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.nodes[node.ID] = node
}

func (rm *ResourceMonitor) RemoveNode(id string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.nodes, id)
}
