package buffer

import (
	"enterprise-core/backend/pkg/logger"
	"sync"
	"time"
)

// BackpressureController manages flow control
type BackpressureController struct {
	maxBufferSize     int
	currentLoad       int
	threshold         float64 // 0.0 to 1.0
	circuitOpen       bool
	circuitOpenTime   time.Time
	circuitTimeout    time.Duration
	logger            *logger.Logger
	mu                sync.RWMutex
	rejectedCount     int64
	acceptedCount     int64
}

func NewBackpressureController(maxBufferSize int, threshold float64, logger *logger.Logger) *BackpressureController {
	return &BackpressureController{
		maxBufferSize:  maxBufferSize,
		threshold:      threshold,
		circuitTimeout: 30 * time.Second,
		logger:         logger,
	}
}

func (b *BackpressureController) ShouldAccept() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check circuit breaker
	if b.circuitOpen {
		if time.Since(b.circuitOpenTime) > b.circuitTimeout {
			b.circuitOpen = false
			b.logger.Info("Circuit breaker closed - resuming normal operation")
		} else {
			b.rejectedCount++
			return false
		}
	}

	// Check load
	loadRatio := float64(b.currentLoad) / float64(b.maxBufferSize)
	
	if loadRatio > b.threshold {
		b.rejectedCount++
		
		// Open circuit if severely overloaded
		if loadRatio > 0.95 {
			b.circuitOpen = true
			b.circuitOpenTime = time.Now()
			b.logger.Warn("Circuit breaker opened - system overloaded", 
				"load", b.currentLoad, 
				"max", b.maxBufferSize)
		}
		
		return false
	}

	b.acceptedCount++
	return true
}

func (b *BackpressureController) UpdateLoad(currentLoad int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.currentLoad = currentLoad
}

func (b *BackpressureController) GetStats() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return map[string]interface{}{
		"current_load":   b.currentLoad,
		"max_buffer":     b.maxBufferSize,
		"load_ratio":     float64(b.currentLoad) / float64(b.maxBufferSize),
		"circuit_open":   b.circuitOpen,
		"rejected_count": b.rejectedCount,
		"accepted_count": b.acceptedCount,
	}
}

func (b *BackpressureController) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.rejectedCount = 0
	b.acceptedCount = 0
}
