package cluster

import (
	"crypto/sha256"
	"encoding/hex"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"sync"
	"time"
)

// ConfigPackage represents a versioned configuration bundle
type ConfigPackage struct {
	Version   string            `json:"version"`
	Checksum  string            `json:"checksum"`
	Content   map[string]string `json:"content"`
	CreatedAt time.Time         `json:"created_at"`
	Author    string            `json:"author"`
}

// ConfigStore manages versioned configurations
type ConfigStore struct {
	configs map[string]*ConfigPackage
	current string
	mu      sync.RWMutex
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{
		configs: make(map[string]*ConfigPackage),
	}
}

func (s *ConfigStore) Store(pkg *ConfigPackage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configs[pkg.Version] = pkg
	s.current = pkg.Version
}

func (s *ConfigStore) Get(version string) (*ConfigPackage, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pkg, ok := s.configs[version]
	return pkg, ok
}

func (s *ConfigStore) GetCurrent() (*ConfigPackage, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.current == "" {
		return nil, false
	}
	pkg, ok := s.configs[s.current]
	return pkg, ok
}

func (s *ConfigStore) ListVersions() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	versions := make([]string, 0, len(s.configs))
	for v := range s.configs {
		versions = append(versions, v)
	}
	return versions
}

func (s *ConfigStore) Rollback(version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.configs[version]; !ok {
		return fmt.Errorf("version %s not found", version)
	}
	s.current = version
	return nil
}

// DeploymentServer orchestrates configuration deployment across the cluster
type DeploymentServer struct {
	store    *ConfigStore
	nodeMgr  *NodeManager
	logger   *logger.Logger
}

func NewDeploymentServer(nodeMgr *NodeManager, l *logger.Logger) *DeploymentServer {
	return &DeploymentServer{
		store:   NewConfigStore(),
		nodeMgr: nodeMgr,
		logger:  l,
	}
}

func (d *DeploymentServer) GetStore() *ConfigStore {
	return d.store
}

// CreateConfig creates a new configuration package
func (d *DeploymentServer) CreateConfig(version string, content map[string]string, author string) (*ConfigPackage, error) {
	// Calculate checksum
	checksum := d.calculateChecksum(content)
	
	pkg := &ConfigPackage{
		Version:   version,
		Checksum:  checksum,
		Content:   content,
		CreatedAt: time.Now(),
		Author:    author,
	}
	
	d.store.Store(pkg)
	d.logger.Info("Created new config package", "version", version, "checksum", checksum)
	
	return pkg, nil
}

// DeployConfig pushes a configuration to all cluster nodes
func (d *DeploymentServer) DeployConfig(version string) error {
	pkg, ok := d.store.Get(version)
	if !ok {
		return fmt.Errorf("config version %s not found", version)
	}
	
	nodes := d.nodeMgr.GetNodes()
	var targets []Node
	for _, n := range nodes {
		if n.Status == NodeOnline {
			targets = append(targets, n)
		}
	}
	
	d.logger.Info("Deploying config to cluster", 
		"version", version, 
		"nodes", len(targets),
		"checksum", pkg.Checksum)
	
	// In a real implementation, this would use gRPC to push configs
	// For now, we just log the deployment
	for _, node := range targets {
		d.logger.Debug("Deploying config to node", "node_id", node.ID, "version", version)
	}
	
	return nil
}

// RollbackConfig reverts to a previous configuration version
func (d *DeploymentServer) RollbackConfig(version string) error {
	if err := d.store.Rollback(version); err != nil {
		return err
	}
	
	d.logger.Warn("Rolling back configuration", "version", version)
	return d.DeployConfig(version)
}

func (d *DeploymentServer) calculateChecksum(content map[string]string) string {
	h := sha256.New()
	for k, v := range content {
		h.Write([]byte(k + "=" + v + ";"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateChecksum verifies a config package's integrity
func (d *DeploymentServer) ValidateChecksum(pkg *ConfigPackage) bool {
	expected := d.calculateChecksum(pkg.Content)
	return expected == pkg.Checksum
}

// DeploymentHistory tracks deployment operations
type DeploymentHistory struct {
	Version   string    `json:"version"`
	Action    string    `json:"action"` // "deploy", "rollback"
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
}

// GetDeploymentHistory returns the deployment history
func (d *DeploymentServer) GetDeploymentHistory() []DeploymentHistory {
	// In a real implementation, this would be persisted
	// For now, return empty history
	return []DeploymentHistory{}
}

// AtomicDeploy performs an atomic deployment with prepare/commit phases
func (d *DeploymentServer) AtomicDeploy(version string, txCoord *TransactionCoordinator) error {
	pkg, ok := d.store.Get(version)
	if !ok {
		return fmt.Errorf("config version %s not found", version)
	}

	nodes := d.nodeMgr.GetNodes()
	var participants []string
	for _, n := range nodes {
		if n.Status == NodeOnline {
			participants = append(participants, n.ID)
		}
	}

	// Begin transaction
	tx, err := txCoord.BeginTransaction("deployment", participants, map[string]interface{}{
		"version":  version,
		"checksum": pkg.Checksum,
	})
	if err != nil {
		return fmt.Errorf("failed to begin deployment transaction: %w", err)
	}

	// Prepare phase
	if err := txCoord.PreparePhase(tx.ID); err != nil {
		txCoord.AbortTransaction(tx.ID, err.Error())
		return fmt.Errorf("prepare phase failed: %w", err)
	}

	// Deploy
	if err := d.DeployConfig(version); err != nil {
		txCoord.AbortTransaction(tx.ID, err.Error())
		return fmt.Errorf("deployment failed: %w", err)
	}

	// Commit phase
	if err := txCoord.CommitPhase(tx.ID); err != nil {
		return fmt.Errorf("commit phase failed: %w", err)
	}

	d.logger.Info("Atomic deployment completed", "version", version, "tx_id", tx.ID)
	return nil
}
