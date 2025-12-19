package tenant

import (
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"sync"
)

// Manager handles multi-tenant operations
type Manager struct {
	tenants map[string]*Tenant
	logger  *logger.Logger
	mu      sync.RWMutex
}

type Tenant struct {
	ID          string
	Name        string
	Enabled     bool
	Quota       *Quota
	Metadata    map[string]interface{}
	CreatedAt   string
}

type Quota struct {
	MaxEventsPerSecond int64
	MaxStorageBytes    int64
	MaxBufferSize      int
	CurrentStorage     int64
	CurrentBuffer      int
}

func NewManager(logger *logger.Logger) *Manager {
	return &Manager{
		tenants: make(map[string]*Tenant),
		logger:  logger,
	}
}

func (m *Manager) CreateTenant(id, name string, quota *Quota) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tenants[id]; exists {
		return fmt.Errorf("tenant already exists: %s", id)
	}

	tenant := &Tenant{
		ID:       id,
		Name:     name,
		Enabled:  true,
		Quota:    quota,
		Metadata: make(map[string]interface{}),
	}

	m.tenants[id] = tenant
	m.logger.Info("Tenant created", "id", id, "name", name)
	return nil
}

func (m *Manager) GetTenant(id string) (*Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenant, ok := m.tenants[id]
	if !ok {
		return nil, fmt.Errorf("tenant not found: %s", id)
	}

	return tenant, nil
}

func (m *Manager) UpdateQuota(tenantID string, quota *Quota) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, ok := m.tenants[tenantID]
	if !ok {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	tenant.Quota = quota
	m.logger.Info("Tenant quota updated", "id", tenantID)
	return nil
}

func (m *Manager) DisableTenant(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, ok := m.tenants[id]
	if !ok {
		return fmt.Errorf("tenant not found: %s", id)
	}

	tenant.Enabled = false
	m.logger.Info("Tenant disabled", "id", id)
	return nil
}

func (m *Manager) ListTenants() []*Tenant {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenants := make([]*Tenant, 0, len(m.tenants))
	for _, tenant := range m.tenants {
		tenants = append(tenants, tenant)
	}

	return tenants
}

func (m *Manager) ExtractTenantID(data map[string]interface{}) string {
	// Try common tenant ID fields
	fields := []string{"tenant_id", "tenant", "organization_id", "org_id", "customer_id"}
	
	for _, field := range fields {
		if id, ok := data[field].(string); ok && id != "" {
			return id
		}
	}

	// Default tenant
	return "default"
}
