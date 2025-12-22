package cloud

import (
	"sync"
)

// AssetGraph manages cloud assets and their relationships
type AssetGraph struct {
	mu            sync.RWMutex
	assets        map[string]*CloudAsset
	relationships []AssetRelationship
	indexByType   map[string][]string // type -> asset IDs
	indexByRegion map[string][]string // region -> asset IDs
}

// NewAssetGraph creates a new asset graph
func NewAssetGraph() *AssetGraph {
	return &AssetGraph{
		assets:        make(map[string]*CloudAsset),
		relationships: make([]AssetRelationship, 0),
		indexByType:   make(map[string][]string),
		indexByRegion: make(map[string][]string),
	}
}

// AddAsset adds or updates an asset in the graph
func (g *AssetGraph) AddAsset(asset *CloudAsset) {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	g.assets[asset.ID] = asset
	
	// Update indexes
	g.indexByType[asset.Type] = append(g.indexByType[asset.Type], asset.ID)
	g.indexByRegion[asset.Region] = append(g.indexByRegion[asset.Region], asset.ID)
}

// GetAsset retrieves an asset by ID
func (g *AssetGraph) GetAsset(id string) *CloudAsset {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	return g.assets[id]
}

// AddRelationship adds a relationship between assets
func (g *AssetGraph) AddRelationship(rel AssetRelationship) {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	g.relationships = append(g.relationships, rel)
}

// GetConnected finds all assets connected to a given asset within depth
func (g *AssetGraph) GetConnected(assetID string, depth int) []*CloudAsset {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	visited := make(map[string]bool)
	result := make([]*CloudAsset, 0)
	
	g.bfs(assetID, depth, visited, &result)
	
	return result
}

// bfs performs breadth-first search for connected assets
func (g *AssetGraph) bfs(startID string, maxDepth int, visited map[string]bool, result *[]*CloudAsset) {
	if maxDepth < 0 {
		return
	}
	
	queue := []struct {
		id    string
		depth int
	}{{startID, 0}}
	
	visited[startID] = true
	
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		
		if current.depth > maxDepth {
			continue
		}
		
		// Add current asset to result
		if asset := g.assets[current.id]; asset != nil {
			*result = append(*result, asset)
		}
		
		// Find connected assets
		for _, rel := range g.relationships {
			var nextID string
			if rel.FromID == current.id {
				nextID = rel.ToID
			} else if rel.ToID == current.id {
				nextID = rel.FromID
			}
			
			if nextID != "" && !visited[nextID] {
				visited[nextID] = true
				queue = append(queue, struct {
					id    string
					depth int
				}{nextID, current.depth + 1})
			}
		}
	}
}

// FilterAssets returns assets matching the filter criteria
func (g *AssetGraph) FilterAssets(filter AssetFilter) []*CloudAsset {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	result := make([]*CloudAsset, 0)
	
	for _, asset := range g.assets {
		if !g.matchesFilter(asset, filter) {
			continue
		}
		result = append(result, asset)
	}
	
	return result
}

// matchesFilter checks if an asset matches the filter criteria
func (g *AssetGraph) matchesFilter(asset *CloudAsset, filter AssetFilter) bool {
	if filter.Provider != "" && asset.Provider != filter.Provider {
		return false
	}
	
	if filter.Type != "" && asset.Type != filter.Type {
		return false
	}
	
	if filter.Region != "" && asset.Region != filter.Region {
		return false
	}
	
	if filter.Account != "" && asset.Account != filter.Account {
		return false
	}
	
	if filter.MinRisk > 0 && asset.RiskScore < filter.MinRisk {
		return false
	}
	
	if filter.MaxRisk > 0 && asset.RiskScore > filter.MaxRisk {
		return false
	}
	
	// Check tags
	for key, value := range filter.Tags {
		if asset.Tags[key] != value {
			return false
		}
	}
	
	return true
}

// CalculateBlastRadius calculates how many assets could be impacted if given asset is compromised
func (g *AssetGraph) CalculateBlastRadius(assetID string) int {
	connected := g.GetConnected(assetID, 10) // Max depth 10
	return len(connected)
}

// FindPaths finds all paths between two assets (simplified implementation)
func (g *AssetGraph) FindPaths(fromID, toID string, maxDepth int) [][]string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	paths := make([][]string, 0)
	currentPath := []string{fromID}
	visited := make(map[string]bool)
	
	g.dfs(fromID, toID, maxDepth, currentPath, visited, &paths)
	
	return paths
}

// dfs performs depth-first search for paths
func (g *AssetGraph) dfs(current, target string, maxDepth int, path []string, visited map[string]bool, paths *[][]string) {
	if len(path) > maxDepth {
		return
	}
	
	if current == target {
		// Found a path
		pathCopy := make([]string, len(path))
		copy(pathCopy, path)
		*paths = append(*paths, pathCopy)
		return
	}
	
	visited[current] = true
	
	// Explore neighbors
	for _, rel := range g.relationships {
		var next string
		if rel.FromID == current {
			next = rel.ToID
		} else if rel.ToID == current {
			next = rel.FromID
		}
		
		if next != "" && !visited[next] {
			g.dfs(next, target, maxDepth, append(path, next), visited, paths)
		}
	}
	
	visited[current] = false
}

// GetStats returns statistics about the asset graph
func (g *AssetGraph) GetStats() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	providerCounts := make(map[Provider]int)
	typeCounts := make(map[string]int)
	
	for _, asset := range g.assets {
		providerCounts[asset.Provider]++
		typeCounts[asset.Type]++
	}
	
	return map[string]interface{}{
		"total_assets":      len(g.assets),
		"total_relationships": len(g.relationships),
		"by_provider":       providerCounts,
		"by_type":           typeCounts,
	}
}
