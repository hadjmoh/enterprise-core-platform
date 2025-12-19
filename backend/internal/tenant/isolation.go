package tenant

import (
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"path/filepath"
	"sync"
)

// Isolation provides data isolation per tenant
type Isolation struct {
	buffers     map[string]*buffer.RingBuffer
	storagePaths map[string]string
	manager     *Manager
	logger      *logger.Logger
	mu          sync.RWMutex
}

func NewIsolation(manager *Manager, baseStoragePath string, logger *logger.Logger) *Isolation {
	return &Isolation{
		buffers:      make(map[string]*buffer.RingBuffer),
		storagePaths: make(map[string]string),
		manager:      manager,
		logger:       logger,
	}
}

func (i *Isolation) GetBuffer(tenantID string) (*buffer.RingBuffer, error) {
	i.mu.RLock()
	buf, ok := i.buffers[tenantID]
	i.mu.RUnlock()

	if ok {
		return buf, nil
	}

	// Create buffer for tenant
	return i.createBuffer(tenantID)
}

func (i *Isolation) createBuffer(tenantID string) (*buffer.RingBuffer, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Double-check after acquiring write lock
	if buf, ok := i.buffers[tenantID]; ok {
		return buf, nil
	}

	// Get tenant quota
	tenant, err := i.manager.GetTenant(tenantID)
	if err != nil {
		return nil, err
	}

	if !tenant.Enabled {
		return nil, fmt.Errorf("tenant is disabled: %s", tenantID)
	}

	// Create isolated buffer
	bufferSize := tenant.Quota.MaxBufferSize
	if bufferSize == 0 {
		bufferSize = 10000 // Default
	}

	buf := buffer.NewRingBuffer(bufferSize, i.logger)
	buf.Start()

	i.buffers[tenantID] = buf
	i.logger.Info("Created tenant buffer", "tenant", tenantID, "size", bufferSize)

	return buf, nil
}

func (i *Isolation) GetStoragePath(tenantID string) string {
	i.mu.RLock()
	path, ok := i.storagePaths[tenantID]
	i.mu.RUnlock()

	if ok {
		return path
	}

	// Create storage path for tenant
	return i.createStoragePath(tenantID)
}

func (i *Isolation) createStoragePath(tenantID string) string {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Double-check
	if path, ok := i.storagePaths[tenantID]; ok {
		return path
	}

	path := filepath.Join("./data/tenants", tenantID, "events")
	i.storagePaths[tenantID] = path
	i.logger.Info("Created tenant storage path", "tenant", tenantID, "path", path)

	return path
}

func (i *Isolation) StopTenant(tenantID string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if buf, ok := i.buffers[tenantID]; ok {
		buf.Stop()
		delete(i.buffers, tenantID)
		i.logger.Info("Stopped tenant buffer", "tenant", tenantID)
	}

	return nil
}

func (i *Isolation) GetStats(tenantID string) map[string]interface{} {
	i.mu.RLock()
	defer i.mu.RUnlock()

	stats := make(map[string]interface{})

	if buf, ok := i.buffers[tenantID]; ok {
		stats["buffer_size"] = buf.Size()
	}

	if path, ok := i.storagePaths[tenantID]; ok {
		stats["storage_path"] = path
	}

	return stats
}
