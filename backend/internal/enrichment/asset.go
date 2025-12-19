package enrichment

import (
	"enterprise-core/backend/pkg/logger"
	"sync"
)

// AssetEnricher adds asset metadata to events
type AssetEnricher struct {
	assets map[string]*AssetInfo
	logger *logger.Logger
	mu     sync.RWMutex
}

type AssetInfo struct {
	Hostname    string
	Owner       string
	Department  string
	Criticality string // low, medium, high, critical
	Environment string // dev, staging, production
	Tags        []string
}

func NewAssetEnricher(logger *logger.Logger) *AssetEnricher {
	return &AssetEnricher{
		assets: make(map[string]*AssetInfo),
		logger: logger,
	}
}

func (a *AssetEnricher) Enrich(data map[string]interface{}) map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Try to find asset by IP or hostname
	var assetKey string
	
	if ip, ok := data["src_ip"].(string); ok {
		assetKey = ip
	} else if hostname, ok := data["hostname"].(string); ok {
		assetKey = hostname
	}

	if assetKey == "" {
		return data
	}

	if asset, ok := a.assets[assetKey]; ok {
		data["asset_hostname"] = asset.Hostname
		data["asset_owner"] = asset.Owner
		data["asset_department"] = asset.Department
		data["asset_criticality"] = asset.Criticality
		data["asset_environment"] = asset.Environment
		data["asset_tags"] = asset.Tags
	}

	return data
}

func (a *AssetEnricher) LoadAssets(assets map[string]*AssetInfo) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.assets = assets
	a.logger.Info("Loaded assets", "count", len(assets))
}

func (a *AssetEnricher) AddAsset(key string, info *AssetInfo) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.assets[key] = info
}
