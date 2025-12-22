package cluster

import (
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"time"
)

// RollbackManager coordinates safe rollbacks across the cluster
type RollbackManager struct {
	deploySrv *DeploymentServer
	nodeMgr   *NodeManager
	txCoord   *TransactionCoordinator
	logger    *logger.Logger
}

func NewRollbackManager(deploySrv *DeploymentServer, nodeMgr *NodeManager, txCoord *TransactionCoordinator, l *logger.Logger) *RollbackManager {
	return &RollbackManager{
		deploySrv: deploySrv,
		nodeMgr:   nodeMgr,
		txCoord:   txCoord,
		logger:    l,
	}
}

// RollbackToVersion performs a cluster-wide rollback to a specific config version
func (rm *RollbackManager) RollbackToVersion(targetVersion string) error {
	rm.logger.Warn("Initiating cluster-wide rollback", "target_version", targetVersion)

	// Verify target version exists
	_, ok := rm.deploySrv.GetStore().Get(targetVersion)
	if !ok {
		return fmt.Errorf("target version %s not found", targetVersion)
	}

	// Check quorum
	if !rm.checkQuorum() {
		return fmt.Errorf("insufficient nodes online for safe rollback")
	}

	// Get all online nodes
	nodes := rm.nodeMgr.GetNodes()
	var participants []string
	for _, n := range nodes {
		if n.Status == NodeOnline {
			participants = append(participants, n.ID)
		}
	}

	// Create rollback transaction
	tx, err := rm.txCoord.BeginTransaction("rollback", participants, map[string]interface{}{
		"target_version": targetVersion,
	})
	if err != nil {
		return fmt.Errorf("failed to begin rollback transaction: %w", err)
	}

	// Execute rollback using deployment server
	if err := rm.deploySrv.RollbackConfig(targetVersion); err != nil {
		rm.txCoord.AbortTransaction(tx.ID, err.Error())
		return fmt.Errorf("rollback failed: %w", err)
	}

	// Commit transaction
	if err := rm.txCoord.CommitPhase(tx.ID); err != nil {
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	rm.logger.Info("Cluster-wide rollback completed", "target_version", targetVersion, "tx_id", tx.ID)
	return nil
}

// ProgressiveRollback rolls back nodes one at a time for safety
func (rm *RollbackManager) ProgressiveRollback(targetVersion string) error {
	rm.logger.Info("Starting progressive rollback", "target_version", targetVersion)

	nodes := rm.nodeMgr.GetNodes()
	var onlineNodes []Node
	for _, n := range nodes {
		if n.Status == NodeOnline {
			onlineNodes = append(onlineNodes, n)
		}
	}

	if len(onlineNodes) == 0 {
		return fmt.Errorf("no online nodes available")
	}

	// Roll back one node at a time
	for i, node := range onlineNodes {
		rm.logger.Info("Rolling back node", 
			"node_id", node.ID, 
			"progress", fmt.Sprintf("%d/%d", i+1, len(onlineNodes)))

		// In a real implementation, this would send rollback command to specific node
		// For now, we just log the action
		time.Sleep(100 * time.Millisecond) // Simulate rollback delay
	}

	// Update deployment server state
	if err := rm.deploySrv.GetStore().Rollback(targetVersion); err != nil {
		return err
	}

	rm.logger.Info("Progressive rollback completed", "target_version", targetVersion)
	return nil
}

// checkQuorum verifies that enough nodes are online for safe operations
func (rm *RollbackManager) checkQuorum() bool {
	nodes := rm.nodeMgr.GetNodes()
	onlineCount := 0
	totalCount := len(nodes)

	for _, n := range nodes {
		if n.Status == NodeOnline {
			onlineCount++
		}
	}

	// Require at least 50% of nodes to be online
	quorum := (totalCount / 2) + 1
	hasQuorum := onlineCount >= quorum

	rm.logger.Debug("Quorum check", 
		"online", onlineCount, 
		"total", totalCount, 
		"required", quorum,
		"has_quorum", hasQuorum)

	return hasQuorum
}

// GetRollbackStatus returns the current rollback status
func (rm *RollbackManager) GetRollbackStatus() map[string]interface{} {
	currentPkg, _ := rm.deploySrv.GetStore().GetCurrent()
	
	status := map[string]interface{}{
		"quorum_available": rm.checkQuorum(),
		"available_versions": rm.deploySrv.GetStore().ListVersions(),
	}

	if currentPkg != nil {
		status["current_version"] = currentPkg.Version
	}

	return status
}
