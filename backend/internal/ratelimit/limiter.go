package ratelimit

import (
	"context"
	"enterprise-core/backend/pkg/logger"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limiter manages rate limiting for multiple sources
type Limiter struct {
	limiters map[string]*rate.Limiter
	logger   *logger.Logger
	mu       sync.RWMutex
	
	// Default limits
	defaultRate  rate.Limit
	defaultBurst int
}

func NewLimiter(ratePerSecond int, burst int, logger *logger.Logger) *Limiter {
	return &Limiter{
		limiters:     make(map[string]*rate.Limiter),
		logger:       logger,
		defaultRate:  rate.Limit(ratePerSecond),
		defaultBurst: burst,
	}
}

func (l *Limiter) Allow(key string) bool {
	limiter := l.getLimiter(key)
	return limiter.Allow()
}

func (l *Limiter) Wait(key string, timeout time.Duration) bool {
	limiter := l.getLimiter(key)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	err := limiter.Wait(ctx)
	return err == nil
}

func (l *Limiter) getLimiter(key string) *rate.Limiter {
	l.mu.RLock()
	limiter, ok := l.limiters[key]
	l.mu.RUnlock()

	if ok {
		return limiter
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, ok := l.limiters[key]; ok {
		return limiter
	}

	limiter = rate.NewLimiter(l.defaultRate, l.defaultBurst)
	l.limiters[key] = limiter
	l.logger.Info("Created rate limiter", "key", key, "rate", l.defaultRate, "burst", l.defaultBurst)
	
	return limiter
}

func (l *Limiter) SetLimit(key string, ratePerSecond int, burst int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter := rate.NewLimiter(rate.Limit(ratePerSecond), burst)
	l.limiters[key] = limiter
	l.logger.Info("Set custom rate limit", "key", key, "rate", ratePerSecond, "burst", burst)
}

func (l *Limiter) Remove(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.limiters, key)
}

func (l *Limiter) GetStats() map[string]interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return map[string]interface{}{
		"total_limiters": len(l.limiters),
		"default_rate":   l.defaultRate,
		"default_burst":  l.defaultBurst,
	}
}
