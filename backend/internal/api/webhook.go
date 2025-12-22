package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"enterprise-core/backend/internal/app"
	"enterprise-core/backend/pkg/logger"
)

type WebhookHandler struct {
	manager *app.WebhookManager
	logger  *logger.Logger
}

func NewWebhookHandler(manager *app.WebhookManager, log *logger.Logger) *WebhookHandler {
	return &WebhookHandler{
		manager: manager,
		logger:  log,
	}
}

// ListWebhooks returns all webhooks for an app
func (h *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/apps/{name}/webhooks
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	appName := parts[4]

	webhooks := h.manager.ListWebhooks(appName)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhooks)
}

// RegisterWebhook creates a new webhook
func (h *WebhookHandler) RegisterWebhook(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/apps/{name}/webhooks
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	appName := parts[4]

	var req struct {
		URL    string   `json:"url"`
		Secret string   `json:"secret"`
		Events []string `json:"events"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}
	if req.Secret == "" {
		http.Error(w, "Secret is required", http.StatusBadRequest)
		return
	}
	if len(req.Events) == 0 {
		http.Error(w, "At least one event type is required", http.StatusBadRequest)
		return
	}

	webhook, err := h.manager.RegisterWebhook(appName, req.URL, req.Secret, req.Events)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhook)
}

// DeleteWebhook removes a webhook
func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/apps/{name}/webhooks/{id}
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 7 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	webhookID := parts[6]

	if err := h.manager.DeleteWebhook(webhookID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// GetWebhookLogs returns delivery history for a webhook
func (h *WebhookHandler) GetWebhookLogs(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/apps/{name}/webhooks/{id}/logs
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 8 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	webhookID := parts[6]

	deliveries := h.manager.GetDeliveries(webhookID)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deliveries)
}
