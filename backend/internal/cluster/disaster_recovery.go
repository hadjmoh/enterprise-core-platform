package cluster

import (
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"time"
)

// DRCoordinator orchestrates disaster recovery operations
type DRCoordinator struct {
	siteMgr     *SiteManager
	replMgr     *ReplicationManager
	txCoord     *TransactionCoordinator
	logger      *logger.Logger
	rpoMinutes  int // Recovery Point Objective in minutes
	rtoMinutes  int // Recovery Time Objective in minutes
}

func NewDRCoordinator(siteMgr *SiteManager, replMgr *ReplicationManager, txCoord *TransactionCoordinator, l *logger.Logger) *DRCoordinator {
	return &DRCoordinator{
		siteMgr:    siteMgr,
		replMgr:    replMgr,
		txCoord:    txCoord,
		logger:     l,
		rpoMinutes: 15, // Default: 15 minutes RPO
		rtoMinutes: 30, // Default: 30 minutes RTO
	}
}

// MonitorPrimarySite continuously monitors the primary site health
func (dr *DRCoordinator) MonitorPrimarySite() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		primary, ok := dr.siteMgr.GetPrimarySite()
		if !ok || primary.Status != SiteOnline {
			dr.logger.Warn("Primary site is offline or unavailable")
			// In a real implementation, this would trigger automatic failover
			// For now, we just log the condition
		}
	}
}

// TriggerFailover initiates failover to a backup site
func (dr *DRCoordinator) TriggerFailover(targetSiteID string) error {
	dr.logger.Warn("Initiating disaster recovery failover", "target_site", targetSiteID)
	
	startTime := time.Now()
	
	// Verify target site is available
	targetSite, ok := dr.siteMgr.GetSite(targetSiteID)
	if !ok {
		return fmt.Errorf("target site %s not found", targetSiteID)
	}
	
	if targetSite.Status != SiteOnline {
		return fmt.Errorf("target site %s is not online", targetSiteID)
	}
	
	// Check replication lag
	maxLag := dr.replMgr.GetMaxReplicationLag()
	if maxLag > dr.rpoMinutes*60 {
		dr.logger.Warn("Replication lag exceeds RPO", 
			"lag_seconds", maxLag, 
			"rpo_seconds", dr.rpoMinutes*60)
	}
	
	// Create failover transaction
	tx, err := dr.txCoord.BeginTransaction("failover", []string{targetSiteID}, map[string]interface{}{
		"target_site": targetSiteID,
		"start_time":  startTime,
	})
	if err != nil {
		return fmt.Errorf("failed to begin failover transaction: %w", err)
	}
	
	// Promote target site to primary
	if err := dr.siteMgr.PromoteSite(targetSiteID); err != nil {
		dr.txCoord.AbortTransaction(tx.ID, err.Error())
		return fmt.Errorf("failed to promote site: %w", err)
	}
	
	// Commit transaction
	if err := dr.txCoord.CommitPhase(tx.ID); err != nil {
		return fmt.Errorf("failed to commit failover: %w", err)
	}
	
	duration := time.Since(startTime)
	dr.logger.Info("Failover completed", 
		"target_site", targetSiteID, 
		"duration_seconds", duration.Seconds(),
		"tx_id", tx.ID)
	
	// Check if RTO was met
	if duration.Minutes() > float64(dr.rtoMinutes) {
		dr.logger.Warn("Failover exceeded RTO", 
			"duration_minutes", duration.Minutes(), 
			"rto_minutes", dr.rtoMinutes)
	}
	
	return nil
}

// GetDRStatus returns the current disaster recovery status
func (dr *DRCoordinator) GetDRStatus() map[string]interface{} {
	primary, hasPrimary := dr.siteMgr.GetPrimarySite()
	backups := dr.siteMgr.GetBackupSites()
	replStatus := dr.replMgr.GetReplicationStatus()
	
	status := map[string]interface{}{
		"rpo_minutes":        dr.rpoMinutes,
		"rto_minutes":        dr.rtoMinutes,
		"has_primary":        hasPrimary,
		"backup_sites_count": len(backups),
		"replication_streams": len(replStatus),
		"max_replication_lag": dr.replMgr.GetMaxReplicationLag(),
	}
	
	if hasPrimary {
		status["primary_site"] = primary.ID
		status["primary_location"] = primary.Location
	}
	
	return status
}

// TestDR performs a DR test without actual failover
func (dr *DRCoordinator) TestDR(targetSiteID string) error {
	dr.logger.Info("Starting DR test", "target_site", targetSiteID)
	
	// Verify target site
	targetSite, ok := dr.siteMgr.GetSite(targetSiteID)
	if !ok {
		return fmt.Errorf("target site %s not found", targetSiteID)
	}
	
	if targetSite.Status != SiteOnline {
		return fmt.Errorf("target site %s is not online", targetSiteID)
	}
	
	// Check replication health
	replStatus := dr.replMgr.GetReplicationStatus()
	healthyStreams := 0
	for _, status := range replStatus {
		if status.IsHealthy {
			healthyStreams++
		}
	}
	
	dr.logger.Info("DR test completed", 
		"target_site", targetSiteID,
		"healthy_replication_streams", healthyStreams,
		"total_streams", len(replStatus))
	
	return nil
}
