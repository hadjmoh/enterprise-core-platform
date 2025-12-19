package buffer

import (
	"context"
	"enterprise-core/backend/pkg/logger"
	"sync"
)

// Event represents a normalized event ready for persistence
type Event struct {
	Timestamp string
	Source    string
	Data      map[string]interface{}
}

// RingBuffer provides a thread-safe circular buffer for events
type RingBuffer struct {
	events   chan Event
	size     int
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	handlers []EventHandler
}

// EventHandler processes events from the buffer
type EventHandler interface {
	Handle(event Event) error
}

func NewRingBuffer(size int, logger *logger.Logger) *RingBuffer {
	ctx, cancel := context.WithCancel(context.Background())
	return &RingBuffer{
		events:   make(chan Event, size),
		size:     size,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
		handlers: make([]EventHandler, 0),
	}
}

func (r *RingBuffer) AddHandler(handler EventHandler) {
	r.handlers = append(r.handlers, handler)
}

func (r *RingBuffer) Start() {
	r.wg.Add(1)
	go r.processEvents()
	r.logger.Info("Ring Buffer started", "size", r.size)
}

func (r *RingBuffer) Push(event Event) error {
	select {
	case r.events <- event:
		return nil
	case <-r.ctx.Done():
		return r.ctx.Err()
	default:
		// Buffer full - drop oldest or block based on strategy
		r.logger.Warn("Ring Buffer full, dropping event")
		return nil
	}
}

func (r *RingBuffer) processEvents() {
	defer r.wg.Done()

	for {
		select {
		case <-r.ctx.Done():
			r.logger.Info("Ring Buffer stopping, draining remaining events")
			r.drain()
			return
		case event := <-r.events:
			for _, handler := range r.handlers {
				if err := handler.Handle(event); err != nil {
					r.logger.Error("Handler error", err)
				}
			}
		}
	}
}

func (r *RingBuffer) drain() {
	for {
		select {
		case event := <-r.events:
			for _, handler := range r.handlers {
				handler.Handle(event)
			}
		default:
			return
		}
	}
}

func (r *RingBuffer) Stop() {
	r.cancel()
	r.wg.Wait()
	close(r.events)
	r.logger.Info("Ring Buffer stopped")
}

func (r *RingBuffer) Size() int {
	return len(r.events)
}
