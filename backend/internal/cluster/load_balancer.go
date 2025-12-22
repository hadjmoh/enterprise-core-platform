package cluster

import (
	"enterprise-core/backend/pkg/logger"
	"math"
	"sync"
)

// LoadBalancingStrategy defines how nodes are selected
type LoadBalancingStrategy string

const (
	StrategyLeastLoaded  LoadBalancingStrategy = "least_loaded"
	StrategyRoundRobin   LoadBalancingStrategy = "round_robin"
	StrategyLocalityAware LoadBalancingStrategy = "locality_aware"
)

// NodeScore represents a node's suitability for handling a request
type NodeScore struct {
	NodeID string
	Score  float64
}

// LoadBalancer intelligently distributes search requests across nodes
type LoadBalancer struct {
	strategy LoadBalancingStrategy
	logger   *logger.Logger
	rrIndex  int
	mu       sync.Mutex
}

func NewLoadBalancer(strategy LoadBalancingStrategy, l *logger.Logger) *LoadBalancer {
	return &LoadBalancer{
		strategy: strategy,
		logger:   l,
		rrIndex:  0,
	}
}

// SelectNodes picks the best nodes for a given task
func (lb *LoadBalancer) SelectNodes(candidates []Node, count int) []Node {
	if len(candidates) == 0 {
		return nil
	}

	if count <= 0 || count > len(candidates) {
		count = len(candidates)
	}

	switch lb.strategy {
	case StrategyLeastLoaded:
		return lb.selectLeastLoaded(candidates, count)
	case StrategyRoundRobin:
		return lb.selectRoundRobin(candidates, count)
	case StrategyLocalityAware:
		return lb.selectLeastLoaded(candidates, count) // Locality is handled upstream
	default:
		return lb.selectRoundRobin(candidates, count)
	}
}

func (lb *LoadBalancer) selectLeastLoaded(candidates []Node, count int) []Node {
	scores := make([]NodeScore, len(candidates))
	for i, node := range candidates {
		scores[i] = NodeScore{
			NodeID: node.ID,
			Score:  lb.calculateScore(node),
		}
	}

	// Sort by score (higher is better)
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].Score > scores[i].Score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	result := make([]Node, 0, count)
	for i := 0; i < count && i < len(scores); i++ {
		for _, node := range candidates {
			if node.ID == scores[i].NodeID {
				result = append(result, node)
				break
			}
		}
	}

	return result
}

func (lb *LoadBalancer) selectRoundRobin(candidates []Node, count int) []Node {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	result := make([]Node, 0, count)
	for i := 0; i < count; i++ {
		idx := (lb.rrIndex + i) % len(candidates)
		result = append(result, candidates[idx])
	}
	lb.rrIndex = (lb.rrIndex + count) % len(candidates)

	return result
}

// calculateScore computes a node's suitability (0-100, higher is better)
func (lb *LoadBalancer) calculateScore(node Node) float64 {
	// Base score starts at 100
	score := 100.0

	// Penalize high CPU usage (0-100% CPU reduces score by 0-50 points)
	cpuPenalty := (float64(node.CPUUsage) / 100.0) * 50.0
	score -= cpuPenalty

	// Penalize high memory usage (0-100% Memory reduces score by 0-30 points)
	memPenalty := (float64(node.MemUsage) / 100.0) * 30.0
	score -= memPenalty

	// Penalize active searches (each active search reduces score by 5 points)
	searchPenalty := float64(node.ActiveSearches) * 5.0
	score -= searchPenalty

	// Ensure score doesn't go negative
	return math.Max(0, score)
}

// GetBestNode returns the single best node from candidates
func (lb *LoadBalancer) GetBestNode(candidates []Node) *Node {
	nodes := lb.SelectNodes(candidates, 1)
	if len(nodes) > 0 {
		return &nodes[0]
	}
	return nil
}
