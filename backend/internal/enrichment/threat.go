package enrichment

import (
	"enterprise-core/backend/pkg/logger"
	"sync"
	"time"
)

// ThreatIntelEnricher adds threat intelligence context to events
type ThreatIntelEnricher struct {
	cache  map[string]*ThreatInfo
	logger *logger.Logger
	mu     sync.RWMutex
	ttl    time.Duration
}

type ThreatInfo struct {
	IsMalicious bool
	Reputation  int // 0-100, lower is worse
	Categories  []string
	LastSeen    time.Time
	Source      string
}

func NewThreatIntelEnricher(logger *logger.Logger) *ThreatIntelEnricher {
	return &ThreatIntelEnricher{
		cache:  make(map[string]*ThreatInfo),
		logger: logger,
		ttl:    1 * time.Hour,
	}
}

func (t *ThreatIntelEnricher) Enrich(data map[string]interface{}) map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Check IP addresses
	ipFields := []string{"src_ip", "dst_ip", "client_ip", "remote_addr", "ip"}
	
	for _, field := range ipFields {
		if ipStr, ok := data[field].(string); ok {
			if info := t.lookup(ipStr); info != nil {
				prefix := field + "_threat"
				data[prefix+"_malicious"] = info.IsMalicious
				data[prefix+"_reputation"] = info.Reputation
				data[prefix+"_categories"] = info.Categories
				data[prefix+"_source"] = info.Source
			}
		}
	}

	// Check file hashes
	hashFields := []string{"md5", "sha1", "sha256", "file_hash"}
	
	for _, field := range hashFields {
		if hash, ok := data[field].(string); ok {
			if info := t.lookup(hash); info != nil {
				data["file_threat_malicious"] = info.IsMalicious
				data["file_threat_categories"] = info.Categories
			}
		}
	}

	return data
}

func (t *ThreatIntelEnricher) lookup(indicator string) *ThreatInfo {
	// Check cache first
	if info, ok := t.cache[indicator]; ok {
		if time.Since(info.LastSeen) < t.ttl {
			return info
		}
	}

	// TODO: Integrate with real threat intel APIs
	// - AlienVault OTX
	// - AbuseIPDB
	// - VirusTotal
	// - Custom feeds
	
	// For now, return nil (no threat data)
	return nil
}

func (t *ThreatIntelEnricher) UpdateCache(indicator string, info *ThreatInfo) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	info.LastSeen = time.Now()
	t.cache[indicator] = info
}

func (t *ThreatIntelEnricher) ClearExpired() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	for key, info := range t.cache {
		if now.Sub(info.LastSeen) > t.ttl {
			delete(t.cache, key)
		}
	}
}
