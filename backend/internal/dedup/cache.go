package dedup

import (
	"crypto/sha256"
	"encoding/hex"
	"enterprise-core/backend/pkg/logger"
	"sync"
	"time"
)

// Cache implements event deduplication
type Cache struct {
	seen   map[string]time.Time
	ttl    time.Duration
	logger *logger.Logger
	mu     sync.RWMutex
}

func NewCache(ttl time.Duration, logger *logger.Logger) *Cache {
	cache := &Cache{
		seen:   make(map[string]time.Time),
		ttl:    ttl,
		logger: logger,
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

func (c *Cache) IsDuplicate(eventID string, data map[string]interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Generate unique key
	key := c.generateKey(eventID, data)

	// Check if seen
	if lastSeen, ok := c.seen[key]; ok {
		if time.Since(lastSeen) < c.ttl {
			return true
		}
	}

	// Mark as seen
	c.seen[key] = time.Now()
	return false
}

func (c *Cache) generateKey(eventID string, data map[string]interface{}) string {
	// If event has explicit ID, use it
	if eventID != "" {
		return eventID
	}

	// Otherwise, hash the event data
	h := sha256.New()
	
	// Hash important fields
	if timestamp, ok := data["@timestamp"].(string); ok {
		h.Write([]byte(timestamp))
	}
	if source, ok := data["source"].(string); ok {
		h.Write([]byte(source))
	}
	if message, ok := data["message"].(string); ok {
		h.Write([]byte(message))
	}

	return hex.EncodeToString(h.Sum(nil))
}

func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, timestamp := range c.seen {
		if now.Sub(timestamp) > c.ttl {
			delete(c.seen, key)
		}
	}
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.seen)
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seen = make(map[string]time.Time)
	c.logger.Info("Deduplication cache cleared")
}
