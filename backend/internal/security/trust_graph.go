package security

import (
	"sync"
	"time"
)

// TrustLevel represents the degree of confidence in an entity or flow
type TrustLevel string

const (
	TrustVerified   TrustLevel = "verified"
	TrustReputable  TrustLevel = "reputable"
	TrustNeutral    TrustLevel = "neutral"
	TrustSuspicious TrustLevel = "suspicious"
	TrustUntrusted  TrustLevel = "untrusted"
)

// TrustNode represents an entity in the trust graph
type TrustNode struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // user, asset, service, process
	TrustLevel TrustLevel             `json:"trust_level"`
	Metadata   map[string]interface{} `json:"metadata"`
	LastSeen   time.Time              `json:"last_seen"`
}

// TrustEdge represents a flow or relationship between nodes
type TrustEdge struct {
	From       string                 `json:"from"`
	To         string                 `json:"to"`
	Relation   string                 `json:"relation"` // produced, consumed, authenticated
	TrustLevel TrustLevel             `json:"trust_level"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TrustGraph manages trust relationships and lineage verification
type TrustGraph struct {
	nodes map[string]*TrustNode
	edges []*TrustEdge
	mu    sync.RWMutex
}

func NewTrustGraph() *TrustGraph {
	return &TrustGraph{
		nodes: make(map[string]*TrustNode),
		edges: make([]*TrustEdge, 0),
	}
}

// UpdateNode adds or updates a node in the graph
func (g *TrustGraph) UpdateNode(id, nodeType string, level TrustLevel, metadata map[string]interface{}) {
	g.mu.Lock()
	defer g.mu.Unlock()

	node, ok := g.nodes[id]
	if !ok {
		node = &TrustNode{
			ID:   id,
			Type: nodeType,
		}
		g.nodes[id] = node
	}
	node.TrustLevel = level
	node.Metadata = metadata
	node.LastSeen = time.Now()
}

// AddEdge records a relationship or flow
func (g *TrustGraph) AddEdge(from, to, relation string, level TrustLevel, metadata map[string]interface{}) {
	g.mu.Lock()
	defer g.mu.Unlock()

	edge := &TrustEdge{
		From:       from,
		To:         to,
		Relation:   relation,
		TrustLevel: level,
		Timestamp:  time.Now(),
		Metadata:   metadata,
	}
	g.edges = append(g.edges, edge)
}

// GetTrustSummary returns high-level trust metrics
func (g *TrustGraph) GetTrustSummary() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	counts := make(map[TrustLevel]int)
	for _, node := range g.nodes {
		counts[node.TrustLevel]++
	}

	return map[string]interface{}{
		"total_nodes":    len(g.nodes),
		"total_edges":    len(g.edges),
		"trust_distribution": counts,
	}
}

// GetNodes returns all trust nodes
func (g *TrustGraph) GetNodes() []*TrustNode {
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodes := make([]*TrustNode, 0, len(g.nodes))
	for _, n := range g.nodes {
		nodes = append(nodes, n)
	}
	return nodes
}

// GetEdges returns all trust edges
func (g *TrustGraph) GetEdges() []*TrustEdge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.edges
}
