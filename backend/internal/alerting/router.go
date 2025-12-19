package alerting

import (
	"enterprise-core/backend/pkg/logger"
	"sync"
	"time"
)

// Router manages alert distribution to multiple channels
type Router struct {
	channels      []Channel
	dedup         *Deduplicator
	logger        *logger.Logger
	mu            sync.RWMutex
	enabled       bool
	severityMap   map[string]int
}

type Alert struct {
	ID          string
	Title       string
	Description string
	Severity    string
	Source      string
	Timestamp   time.Time
	Metadata    map[string]interface{}
}

type Channel interface {
	Send(alert *Alert) error
	Name() string
}

func NewRouter(logger *logger.Logger) *Router {
	return &Router{
		channels: make([]Channel, 0),
		dedup:    NewDeduplicator(5 * time.Minute),
		logger:   logger,
		enabled:  true,
		severityMap: map[string]int{
			"low":      1,
			"medium":   2,
			"high":     3,
			"critical": 4,
		},
	}
}

func (r *Router) AddChannel(channel Channel) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.channels = append(r.channels, channel)
	r.logger.Info("Added alert channel", "name", channel.Name())
}

func (r *Router) SendAlert(alert *Alert) error {
	if !r.enabled {
		return nil
	}

	// Check for duplicate
	if r.dedup.IsDuplicate(alert) {
		r.logger.Info("Suppressed duplicate alert", "id", alert.ID, "title", alert.Title)
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Send to all channels
	for _, channel := range r.channels {
		if err := channel.Send(alert); err != nil {
			r.logger.Error("Failed to send alert", err, "channel", channel.Name(), "alert", alert.Title)
		} else {
			r.logger.Info("Alert sent", "channel", channel.Name(), "alert", alert.Title, "severity", alert.Severity)
		}
	}

	return nil
}

func (r *Router) Enable() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled = true
}

func (r *Router) Disable() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled = false
}

// Deduplicator prevents duplicate alerts
type Deduplicator struct {
	seen   map[string]time.Time
	window time.Duration
	mu     sync.RWMutex
}

func NewDeduplicator(window time.Duration) *Deduplicator {
	return &Deduplicator{
		seen:   make(map[string]time.Time),
		window: window,
	}
}

func (d *Deduplicator) IsDuplicate(alert *Alert) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := alert.ID + ":" + alert.Title
	
	if lastSeen, ok := d.seen[key]; ok {
		if time.Since(lastSeen) < d.window {
			return true
		}
	}

	d.seen[key] = time.Now()
	
	// Clean old entries
	d.cleanup()
	
	return false
}

func (d *Deduplicator) cleanup() {
	now := time.Now()
	for key, timestamp := range d.seen {
		if now.Sub(timestamp) > d.window {
			delete(d.seen, key)
		}
	}
}
