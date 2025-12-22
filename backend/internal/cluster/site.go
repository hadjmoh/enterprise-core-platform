package cluster

import (
	"enterprise-core/backend/pkg/logger"
	"sync"
	"time"
)

// SiteStatus represents the health status of a site
type SiteStatus string

const (
	SiteOnline  SiteStatus = "online"
	SiteOffline SiteStatus = "offline"
	SiteDegraded SiteStatus = "degraded"
)

// Site represents a geographic location in the cluster
type Site struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Location    string     `json:"location"` // e.g., "us-east-1", "eu-west-1"
	Status      SiteStatus `json:"status"`
	IsPrimary   bool       `json:"is_primary"`
	Priority    int        `json:"priority"` // Lower number = higher priority for failover
	Nodes       []string   `json:"nodes"`    // Node IDs in this site
	LastChecked time.Time  `json:"last_checked"`
}

// SiteManager tracks multiple sites and their health
type SiteManager struct {
	sites  map[string]*Site
	logger *logger.Logger
	mu     sync.RWMutex
}

func NewSiteManager(l *logger.Logger) *SiteManager {
	return &SiteManager{
		sites:  make(map[string]*Site),
		logger: l,
	}
}

// RegisterSite adds a new site to the cluster
func (sm *SiteManager) RegisterSite(site *Site) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	site.LastChecked = time.Now()
	sm.sites[site.ID] = site
	sm.logger.Info("Site registered", "site_id", site.ID, "location", site.Location, "is_primary", site.IsPrimary)
}

// UpdateSiteStatus updates the status of a site
func (sm *SiteManager) UpdateSiteStatus(siteID string, status SiteStatus) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if site, ok := sm.sites[siteID]; ok {
		site.Status = status
		site.LastChecked = time.Now()
		sm.logger.Debug("Site status updated", "site_id", siteID, "status", status)
	}
}

// GetSite retrieves a site by ID
func (sm *SiteManager) GetSite(siteID string) (*Site, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	site, ok := sm.sites[siteID]
	return site, ok
}

// GetAllSites returns all registered sites
func (sm *SiteManager) GetAllSites() []*Site {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	sites := make([]*Site, 0, len(sm.sites))
	for _, site := range sm.sites {
		sites = append(sites, site)
	}
	return sites
}

// GetPrimarySite returns the current primary site
func (sm *SiteManager) GetPrimarySite() (*Site, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	for _, site := range sm.sites {
		if site.IsPrimary && site.Status == SiteOnline {
			return site, true
		}
	}
	return nil, false
}

// GetBackupSites returns all backup sites sorted by priority
func (sm *SiteManager) GetBackupSites() []*Site {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	backups := make([]*Site, 0)
	for _, site := range sm.sites {
		if !site.IsPrimary && site.Status == SiteOnline {
			backups = append(backups, site)
		}
	}
	
	// Sort by priority (lower number = higher priority)
	for i := 0; i < len(backups)-1; i++ {
		for j := i + 1; j < len(backups); j++ {
			if backups[j].Priority < backups[i].Priority {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}
	
	return backups
}

// PromoteSite promotes a backup site to primary
func (sm *SiteManager) PromoteSite(siteID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	// Demote current primary
	for _, site := range sm.sites {
		if site.IsPrimary {
			site.IsPrimary = false
			sm.logger.Info("Site demoted from primary", "site_id", site.ID)
		}
	}
	
	// Promote new primary
	if site, ok := sm.sites[siteID]; ok {
		site.IsPrimary = true
		sm.logger.Info("Site promoted to primary", "site_id", siteID)
		return nil
	}
	
	return nil
}
