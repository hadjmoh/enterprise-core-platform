package cluster

import (
	"enterprise-core/backend/pkg/logger"
	"sync"
	"time"
)

// ReplicationStrategy defines how data is replicated across sites
type ReplicationStrategy string

const (
	ReplicationSync      ReplicationStrategy = "synchronous"
	ReplicationAsync     ReplicationStrategy = "asynchronous"
	ReplicationSemiSync  ReplicationStrategy = "semi_synchronous"
)

// ReplicationStatus tracks replication health
type ReplicationStatus struct {
	SourceSite      string    `json:"source_site"`
	TargetSite      string    `json:"target_site"`
	Strategy        ReplicationStrategy `json:"strategy"`
	LagSeconds      int       `json:"lag_seconds"`
	LastReplicated  time.Time `json:"last_replicated"`
	BytesReplicated int64     `json:"bytes_replicated"`
	IsHealthy       bool      `json:"is_healthy"`
}

// ReplicationManager handles cross-site data replication
type ReplicationManager struct {
	siteMgr  *SiteManager
	strategy ReplicationStrategy
	status   map[string]*ReplicationStatus // key: "source->target"
	logger   *logger.Logger
	mu       sync.RWMutex
}

func NewReplicationManager(siteMgr *SiteManager, strategy ReplicationStrategy, l *logger.Logger) *ReplicationManager {
	return &ReplicationManager{
		siteMgr:  siteMgr,
		strategy: strategy,
		status:   make(map[string]*ReplicationStatus),
		logger:   l,
	}
}

// StartReplication initiates replication between two sites
func (rm *ReplicationManager) StartReplication(sourceSiteID, targetSiteID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	key := sourceSiteID + "->" + targetSiteID
	rm.status[key] = &ReplicationStatus{
		SourceSite:      sourceSiteID,
		TargetSite:      targetSiteID,
		Strategy:        rm.strategy,
		LagSeconds:      0,
		LastReplicated:  time.Now(),
		BytesReplicated: 0,
		IsHealthy:       true,
	}
	
	rm.logger.Info("Replication started", 
		"source", sourceSiteID, 
		"target", targetSiteID, 
		"strategy", rm.strategy)
	
	return nil
}

// GetReplicationStatus returns the status of all replication streams
func (rm *ReplicationManager) GetReplicationStatus() []*ReplicationStatus {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	statuses := make([]*ReplicationStatus, 0, len(rm.status))
	for _, status := range rm.status {
		statuses = append(statuses, status)
	}
	return statuses
}

// UpdateReplicationLag updates the replication lag for a stream
func (rm *ReplicationManager) UpdateReplicationLag(sourceSiteID, targetSiteID string, lagSeconds int) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	key := sourceSiteID + "->" + targetSiteID
	if status, ok := rm.status[key]; ok {
		status.LagSeconds = lagSeconds
		status.LastReplicated = time.Now()
		status.IsHealthy = lagSeconds < 60 // Healthy if lag < 60 seconds
	}
}

// GetMaxReplicationLag returns the maximum replication lag across all streams
func (rm *ReplicationManager) GetMaxReplicationLag() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	maxLag := 0
	for _, status := range rm.status {
		if status.LagSeconds > maxLag {
			maxLag = status.LagSeconds
		}
	}
	return maxLag
}
