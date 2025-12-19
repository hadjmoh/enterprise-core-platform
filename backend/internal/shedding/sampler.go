package shedding

import (
	"enterprise-core/backend/pkg/logger"
	"math/rand"
	"sync"
	"time"
)

// Sampler implements probabilistic event sampling
type Sampler struct {
	sampleRate float64 // 0.0 to 1.0
	logger     *logger.Logger
	mu         sync.RWMutex
	
	// Adaptive sampling
	adaptive      bool
	targetLoad    float64
	currentLoad   float64
	droppedCount  int64
	sampledCount  int64
}

func NewSampler(sampleRate float64, logger *logger.Logger) *Sampler {
	rand.Seed(time.Now().UnixNano())
	
	return &Sampler{
		sampleRate: sampleRate,
		logger:     logger,
		adaptive:   false,
	}
}

func NewAdaptiveSampler(targetLoad float64, logger *logger.Logger) *Sampler {
	rand.Seed(time.Now().UnixNano())
	
	return &Sampler{
		sampleRate:  1.0,
		logger:      logger,
		adaptive:    true,
		targetLoad:  targetLoad,
		currentLoad: 0.0,
	}
}

func (s *Sampler) ShouldSample() bool {
	s.mu.RLock()
	rate := s.sampleRate
	s.mu.RUnlock()

	if rand.Float64() < rate {
		s.mu.Lock()
		s.sampledCount++
		s.mu.Unlock()
		return true
	}

	s.mu.Lock()
	s.droppedCount++
	s.mu.Unlock()
	return false
}

func (s *Sampler) UpdateLoad(currentLoad float64) {
	if !s.adaptive {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentLoad = currentLoad

	// Adjust sample rate based on load
	if currentLoad > s.targetLoad {
		// Reduce sample rate to shed load
		reduction := (currentLoad - s.targetLoad) / s.targetLoad
		s.sampleRate = 1.0 - reduction
		if s.sampleRate < 0.1 {
			s.sampleRate = 0.1 // Minimum 10% sampling
		}
		s.logger.Info("Adaptive sampling reduced", "rate", s.sampleRate, "load", currentLoad)
	} else {
		// Increase sample rate back to 100%
		s.sampleRate = 1.0
	}
}

func (s *Sampler) SetSampleRate(rate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rate < 0.0 {
		rate = 0.0
	} else if rate > 1.0 {
		rate = 1.0
	}

	s.sampleRate = rate
	s.logger.Info("Sample rate updated", "rate", rate)
}

func (s *Sampler) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := s.sampledCount + s.droppedCount
	dropRate := 0.0
	if total > 0 {
		dropRate = float64(s.droppedCount) / float64(total)
	}

	return map[string]interface{}{
		"sample_rate":    s.sampleRate,
		"sampled_count":  s.sampledCount,
		"dropped_count":  s.droppedCount,
		"drop_rate":      dropRate,
		"adaptive":       s.adaptive,
		"current_load":   s.currentLoad,
	}
}

func (s *Sampler) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sampledCount = 0
	s.droppedCount = 0
}
