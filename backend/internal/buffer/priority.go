package buffer

import (
	"context"
	"enterprise-core/backend/pkg/logger"
	"sync"
)

// PriorityBuffer manages multiple priority queues
type PriorityBuffer struct {
	queues   map[string]chan Event // priority -> queue
	handlers []EventHandler
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	sizes    map[string]int
}

const (
	PriorityCritical = "critical"
	PriorityHigh     = "high"
	PriorityMedium   = "medium"
	PriorityLow      = "low"
)

func NewPriorityBuffer(sizes map[string]int, logger *logger.Logger) *PriorityBuffer {
	ctx, cancel := context.WithCancel(context.Background())
	
	queues := make(map[string]chan Event)
	for priority, size := range sizes {
		queues[priority] = make(chan Event, size)
	}

	return &PriorityBuffer{
		queues:   queues,
		handlers: make([]EventHandler, 0),
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
		sizes:    sizes,
	}
}

func (p *PriorityBuffer) Start() {
	// Start workers for each priority (critical gets more workers)
	workerCounts := map[string]int{
		PriorityCritical: 4,
		PriorityHigh:     3,
		PriorityMedium:   2,
		PriorityLow:      1,
	}

	for priority, count := range workerCounts {
		if queue, ok := p.queues[priority]; ok {
			for i := 0; i < count; i++ {
				p.wg.Add(1)
				go p.processQueue(priority, queue)
			}
		}
	}

	p.logger.Info("Priority Buffer started", "queues", len(p.queues))
}

func (p *PriorityBuffer) Push(event Event) error {
	priority := p.getPriority(event)
	
	queue, ok := p.queues[priority]
	if !ok {
		queue = p.queues[PriorityMedium] // Default to medium
	}

	select {
	case queue <- event:
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
		// Queue full - drop low priority events first
		if priority == PriorityLow {
			p.logger.Warn("Dropped low priority event - queue full")
			return nil
		}
		// Block for higher priority
		select {
		case queue <- event:
			return nil
		case <-p.ctx.Done():
			return p.ctx.Err()
		}
	}
}

func (p *PriorityBuffer) getPriority(event Event) string {
	// Check event metadata for priority
	if priority, ok := event.Data["priority"].(string); ok {
		return priority
	}

	// Check severity
	if severity, ok := event.Data["severity"].(string); ok {
		switch severity {
		case "critical":
			return PriorityCritical
		case "high":
			return PriorityHigh
		case "medium":
			return PriorityMedium
		}
	}

	// Default priority
	return PriorityMedium
}

func (p *PriorityBuffer) processQueue(priority string, queue chan Event) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			p.drainQueue(queue)
			return
		case event := <-queue:
			for _, handler := range p.handlers {
				if err := handler.Handle(event); err != nil {
					p.logger.Error("Handler error", err)
				}
			}
		}
	}
}

func (p *PriorityBuffer) drainQueue(queue chan Event) {
	for {
		select {
		case event := <-queue:
			for _, handler := range p.handlers {
				handler.Handle(event)
			}
		default:
			return
		}
	}
}

func (p *PriorityBuffer) AddHandler(handler EventHandler) {
	p.handlers = append(p.handlers, handler)
}

func (p *PriorityBuffer) Stop() {
	p.cancel()
	p.wg.Wait()
	for _, queue := range p.queues {
		close(queue)
	}
	p.logger.Info("Priority Buffer stopped")
}

func (p *PriorityBuffer) GetQueueSize(priority string) int {
	if queue, ok := p.queues[priority]; ok {
		return len(queue)
	}
	return 0
}
