package api

import (
	"net/http"
	"strconv"

	"enterprise-core/backend/internal/cloud"
	"github.com/gin-gonic/gin"
)

// CloudAPI handles cloud asset and event endpoints
type CloudAPI struct {
	graph      *cloud.AssetGraph
	normalizer *cloud.Normalizer
}

// NewCloudAPI creates a new cloud API handler
func NewCloudAPI(graph *cloud.AssetGraph) *CloudAPI {
	return &CloudAPI{
		graph:      graph,
		normalizer: cloud.NewNormalizer(),
	}
}

// RegisterRoutes registers cloud API routes
func (api *CloudAPI) RegisterRoutes(r *gin.RouterGroup) {
	cloud := r.Group("/cloud")
	{
		cloud.GET("/assets", api.ListAssets)
		cloud.GET("/assets/:id", api.GetAsset)
		cloud.GET("/assets/:id/graph", api.GetAssetGraph)
		cloud.GET("/assets/:id/blast-radius", api.GetBlastRadius)
		cloud.GET("/stats", api.GetStats)
		cloud.POST("/assets", api.IngestAsset)
	}
}

// ListAssets returns filtered list of cloud assets
// GET /api/cloud/assets?provider=aws&type=ec2&region=us-east-1
func (api *CloudAPI) ListAssets(c *gin.Context) {
	filter := cloud.AssetFilter{
		Provider: cloud.Provider(c.Query("provider")),
		Type:     c.Query("type"),
		Region:   c.Query("region"),
		Account:  c.Query("account"),
	}
	
	// Parse risk range
	if minRisk := c.Query("min_risk"); minRisk != "" {
		if val, err := strconv.Atoi(minRisk); err == nil {
			filter.MinRisk = val
		}
	}
	
	if maxRisk := c.Query("max_risk"); maxRisk != "" {
		if val, err := strconv.Atoi(maxRisk); err == nil {
			filter.MaxRisk = val
		}
	}
	
	assets := api.graph.FilterAssets(filter)
	
	c.JSON(http.StatusOK, gin.H{
		"assets": assets,
		"count":  len(assets),
	})
}

// GetAsset returns a single asset by ID
// GET /api/cloud/assets/:id
func (api *CloudAPI) GetAsset(c *gin.Context) {
	id := c.Param("id")
	
	asset := api.graph.GetAsset(id)
	if asset == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	
	c.JSON(http.StatusOK, asset)
}

// GetAssetGraph returns connected assets
// GET /api/cloud/assets/:id/graph?depth=2
func (api *CloudAPI) GetAssetGraph(c *gin.Context) {
	id := c.Param("id")
	depth := 2
	
	if depthStr := c.Query("depth"); depthStr != "" {
		if val, err := strconv.Atoi(depthStr); err == nil {
			depth = val
		}
	}
	
	center := api.graph.GetAsset(id)
	if center == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	
	connected := api.graph.GetConnected(id, depth)
	
	c.JSON(http.StatusOK, gin.H{
		"center":    center,
		"connected": connected,
		"depth":     depth,
	})
}

// GetBlastRadius calculates potential impact of asset compromise
// GET /api/cloud/assets/:id/blast-radius
func (api *CloudAPI) GetBlastRadius(c *gin.Context) {
	id := c.Param("id")
	
	asset := api.graph.GetAsset(id)
	if asset == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	
	radius := api.graph.CalculateBlastRadius(id)
	
	c.JSON(http.StatusOK, gin.H{
		"asset_id":     id,
		"blast_radius": radius,
		"severity":     api.calculateSeverity(radius),
	})
}

// GetStats returns cloud asset statistics
// GET /api/cloud/stats
func (api *CloudAPI) GetStats(c *gin.Context) {
	stats := api.graph.GetStats()
	c.JSON(http.StatusOK, stats)
}

// IngestAsset ingests a new cloud asset (for testing/manual ingestion)
// POST /api/cloud/assets
func (api *CloudAPI) IngestAsset(c *gin.Context) {
	var req struct {
		Provider cloud.Provider         `json:"provider"`
		RawData  map[string]interface{} `json:"raw_data"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Normalize asset
	asset, err := api.normalizer.NormalizeAsset(req.Provider, req.RawData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Add to graph
	api.graph.AddAsset(asset)
	
	c.JSON(http.StatusCreated, asset)
}

// calculateSeverity determines severity based on blast radius
func (api *CloudAPI) calculateSeverity(radius int) string {
	if radius > 100 {
		return "critical"
	} else if radius > 50 {
		return "high"
	} else if radius > 20 {
		return "medium"
	}
	return "low"
}
