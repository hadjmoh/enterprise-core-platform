package tenant

import (
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"sync"
	"time"
)

// QuotaEnforcer manages quota limits per tenant
type QuotaEnforcer struct {
	manager      *Manager
	rateLimiters map[string]*RateLimiter
	logger       *logger.Logger
	mu           sync.RWMutex
}

type RateLimiter struct {
	limit     int64
	current   int64
	window    time.Duration
	resetTime time.Time
	mu        sync.Mutex
}

func NewQuotaEnforcer(manager *Manager, logger *logger.Logger) *QuotaEnforcer {
	return &QuotaEnforcer{
		manager:      manager,
		rateLimiters: make(map[string]*RateLimiter),
		logger:       logger,
	}
}

func (q *QuotaEnforcer) CheckRateLimit(tenantID string) error {
	tenant, err := q.manager.GetTenant(tenantID)
	if err != nil {
		return err
	}

	if tenant.Quota.MaxEventsPerSecond == 0 {
		return nil // No limit
	}

	limiter := q.getRateLimiter(tenantID, tenant.Quota.MaxEventsPerSecond)
	
	if !limiter.Allow() {
		return fmt.Errorf("rate limit exceeded for tenant %s", tenantID)
	}

	return nil
}

func (q *QuotaEnforcer) CheckStorageQuota(tenantID string, additionalBytes int64) error {
	tenant, err := q.manager.GetTenant(tenantID)
	if err != nil {
		return err
	}

	if tenant.Quota.MaxStorageBytes == 0 {
		return nil // No limit
	}

	if tenant.Quota.CurrentStorage+additionalBytes > tenant.Quota.MaxStorageBytes {
		return fmt.Errorf("storage quota exceeded for tenant %s", tenantID)
	}

	return nil
}

func (q *QuotaEnforcer) UpdateStorageUsage(tenantID string, bytes int64) error {
	tenant, err := q.manager.GetTenant(tenantID)
	if err != nil {
		return err
	}

	tenant.Quota.CurrentStorage += bytes
	return nil
}

func (q *QuotaEnforcer) getRateLimiter(tenantID string, limit int64) *RateLimiter {
	q.mu.RLock()
	limiter, ok := q.rateLimiters[tenantID]
	q.mu.RUnlock()

	if ok {
		return limiter
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// Double-check
	if limiter, ok := q.rateLimiters[tenantID]; ok {
		return limiter
	}

	limiter = &RateLimiter{
		limit:     limit,
		current:   0,
		window:    1 * time.Second,
		resetTime: time.Now().Add(1 * time.Second),
	}

	q.rateLimiters[tenantID] = limiter
	return limiter
}

func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	
	// Reset if window expired
	if now.After(r.resetTime) {
		r.current = 0
		r.resetTime = now.Add(r.window)
	}

	if r.current >= r.limit {
		return false
	}

	r.current++
	return true
}

func (q *QuotaEnforcer) GetUsage(tenantID string) map[string]interface{} {
	tenant, err := q.manager.GetTenant(tenantID)
	if err != nil {
		return nil
	}

	q.mu.RLock()
	limiter, hasLimiter := q.rateLimiters[tenantID]
	q.mu.RUnlock()

	usage := map[string]interface{}{
		"storage_used":  tenant.Quota.CurrentStorage,
		"storage_limit": tenant.Quota.MaxStorageBytes,
	}

	if hasLimiter {
		limiter.mu.Lock()
		usage["rate_current"] = limiter.current
		usage["rate_limit"] = limiter.limit
		limiter.mu.Unlock()
	}

	return usage
}
