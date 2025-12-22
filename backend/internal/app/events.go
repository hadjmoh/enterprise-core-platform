package app

import (
	"time"
)

// Event represents a platform event that can trigger webhooks
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// Standard event types
const (
	EventAlertCreated      = "alert.created"
	EventAlertUpdated      = "alert.updated"
	EventIncidentCreated   = "incident.created"
	EventIncidentResolved  = "incident.resolved"
	EventDataIngested      = "data.ingested"
	EventAppInstalled      = "app.installed"
	EventAppUninstalled    = "app.uninstalled"
	EventAppUpdated        = "app.updated"
)

// EventBus manages event publishing and subscription
type EventBus struct {
	subscribers map[string][]chan Event
	webhookMgr  *WebhookManager
}

// NewEventBus creates a new event bus
func NewEventBus(webhookMgr *WebhookManager) *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan Event),
		webhookMgr:  webhookMgr,
	}
}

// Publish sends an event to all subscribers and triggers webhooks
func (eb *EventBus) Publish(event Event) {
	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Notify in-process subscribers
	if channels, exists := eb.subscribers[event.Type]; exists {
		for _, ch := range channels {
			select {
			case ch <- event:
			default:
				// Channel full, skip
			}
		}
	}

	// Trigger webhooks asynchronously
	if eb.webhookMgr != nil {
		go eb.webhookMgr.DeliverWebhooks(event)
	}
}

// Subscribe registers a channel to receive events of a specific type
func (eb *EventBus) Subscribe(eventType string) <-chan Event {
	ch := make(chan Event, 100)
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	return ch
}

// CreateEvent is a helper to create a new event
func CreateEvent(eventType, source string, payload map[string]interface{}) Event {
	return Event{
		ID:        generateEventID(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now(),
		Payload:   payload,
	}
}

// Simple event ID generator (in production, use UUID)
func generateEventID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
