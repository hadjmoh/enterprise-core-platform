package app

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"enterprise-core/backend/pkg/logger"
)

// WebhookEndpoint represents a registered webhook
type WebhookEndpoint struct {
	ID        string    `json:"id"`
	AppName   string    `json:"app_name"`
	URL       string    `json:"url"`
	Secret    string    `json:"-"` // Never expose in JSON
	Events    []string  `json:"events"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// WebhookDelivery tracks webhook delivery attempts
type WebhookDelivery struct {
	ID           string    `json:"id"`
	WebhookID    string    `json:"webhook_id"`
	Event        Event     `json:"event"`
	StatusCode   int       `json:"status_code"`
	ResponseBody string    `json:"response_body"`
	Attempt      int       `json:"attempt"`
	Success      bool      `json:"success"`
	DeliveredAt  time.Time `json:"delivered_at"`
}

// WebhookManager manages webhook registration and delivery
type WebhookManager struct {
	mu        sync.RWMutex
	webhooks  map[string]*WebhookEndpoint // webhook ID -> endpoint
	byApp     map[string][]*WebhookEndpoint // app name -> webhooks
	deliveries map[string][]*WebhookDelivery // webhook ID -> deliveries
	logger    *logger.Logger
	client    *http.Client
}

// NewWebhookManager creates a new webhook manager
func NewWebhookManager(log *logger.Logger) *WebhookManager {
	return &WebhookManager{
		webhooks:   make(map[string]*WebhookEndpoint),
		byApp:      make(map[string][]*WebhookEndpoint),
		deliveries: make(map[string][]*WebhookDelivery),
		logger:     log,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RegisterWebhook registers a new webhook endpoint
func (wm *WebhookManager) RegisterWebhook(appName, url, secret string, events []string) (*WebhookEndpoint, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	webhook := &WebhookEndpoint{
		ID:        generateWebhookID(),
		AppName:   appName,
		URL:       url,
		Secret:    secret,
		Events:    events,
		Active:    true,
		CreatedAt: time.Now(),
	}

	wm.webhooks[webhook.ID] = webhook
	wm.byApp[appName] = append(wm.byApp[appName], webhook)

	wm.logger.Info(fmt.Sprintf("Registered webhook for app %s: %s", appName, webhook.ID))
	return webhook, nil
}

// DeleteWebhook removes a webhook
func (wm *WebhookManager) DeleteWebhook(webhookID string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	webhook, exists := wm.webhooks[webhookID]
	if !exists {
		return fmt.Errorf("webhook not found: %s", webhookID)
	}

	// Remove from byApp map
	appWebhooks := wm.byApp[webhook.AppName]
	for i, w := range appWebhooks {
		if w.ID == webhookID {
			wm.byApp[webhook.AppName] = append(appWebhooks[:i], appWebhooks[i+1:]...)
			break
		}
	}

	delete(wm.webhooks, webhookID)
	delete(wm.deliveries, webhookID)

	wm.logger.Info(fmt.Sprintf("Deleted webhook: %s", webhookID))
	return nil
}

// ListWebhooks returns all webhooks for an app
func (wm *WebhookManager) ListWebhooks(appName string) []*WebhookEndpoint {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	webhooks := wm.byApp[appName]
	result := make([]*WebhookEndpoint, len(webhooks))
	copy(result, webhooks)
	return result
}

// DeliverWebhooks sends an event to all subscribed webhooks
func (wm *WebhookManager) DeliverWebhooks(event Event) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	for _, webhook := range wm.webhooks {
		if !webhook.Active {
			continue
		}

		// Check if webhook is subscribed to this event type
		subscribed := false
		for _, eventType := range webhook.Events {
			if eventType == event.Type || eventType == "*" {
				subscribed = true
				break
			}
		}

		if !subscribed {
			continue
		}

		// Deliver asynchronously with retries
		go wm.deliverWithRetry(webhook, event, 3)
	}
}

// deliverWithRetry attempts delivery with exponential backoff
func (wm *WebhookManager) deliverWithRetry(webhook *WebhookEndpoint, event Event, maxAttempts int) {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		delivery := wm.deliver(webhook, event, attempt)
		
		wm.mu.Lock()
		wm.deliveries[webhook.ID] = append(wm.deliveries[webhook.ID], delivery)
		wm.mu.Unlock()

		if delivery.Success {
			return
		}

		// Exponential backoff: 1s, 2s, 4s
		if attempt < maxAttempts {
			time.Sleep(time.Duration(1<<uint(attempt-1)) * time.Second)
		}
	}

	wm.logger.Warn(fmt.Sprintf("Webhook delivery failed after %d attempts: %s", maxAttempts, webhook.ID))
}

// deliver sends a single webhook delivery
func (wm *WebhookManager) deliver(webhook *WebhookEndpoint, event Event, attempt int) *WebhookDelivery {
	delivery := &WebhookDelivery{
		ID:          generateDeliveryID(),
		WebhookID:   webhook.ID,
		Event:       event,
		Attempt:     attempt,
		DeliveredAt: time.Now(),
	}

	// Serialize event payload
	payload, err := json.Marshal(event)
	if err != nil {
		delivery.Success = false
		delivery.ResponseBody = fmt.Sprintf("Failed to serialize event: %v", err)
		return delivery
	}

	// Create request
	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(payload))
	if err != nil {
		delivery.Success = false
		delivery.ResponseBody = fmt.Sprintf("Failed to create request: %v", err)
		return delivery
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Event-Type", event.Type)
	req.Header.Set("X-Event-ID", event.ID)
	
	// Add signature
	signature := wm.generateSignature(payload, webhook.Secret)
	req.Header.Set("X-Webhook-Signature", signature)

	// Send request
	resp, err := wm.client.Do(req)
	if err != nil {
		delivery.Success = false
		delivery.ResponseBody = fmt.Sprintf("Request failed: %v", err)
		return delivery
	}
	defer resp.Body.Close()

	delivery.StatusCode = resp.StatusCode
	delivery.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

	// Read response body (limited)
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	delivery.ResponseBody = string(buf[:n])

	return delivery
}

// generateSignature creates HMAC-SHA256 signature
func (wm *WebhookManager) generateSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// ValidateSignature verifies webhook signature
func (wm *WebhookManager) ValidateSignature(payload []byte, signature, secret string) bool {
	expected := wm.generateSignature(payload, secret)
	return hmac.Equal([]byte(signature), []byte(expected))
}

// GetDeliveries returns delivery history for a webhook
func (wm *WebhookManager) GetDeliveries(webhookID string) []*WebhookDelivery {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	deliveries := wm.deliveries[webhookID]
	result := make([]*WebhookDelivery, len(deliveries))
	copy(result, deliveries)
	return result
}

// Helper functions
func generateWebhookID() string {
	return "wh_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

func generateDeliveryID() string {
	return "del_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}
