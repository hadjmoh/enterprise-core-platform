package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"enterprise-core/backend/internal/app"
	"enterprise-core/backend/pkg/logger"
)

type SandboxHandler struct {
	manager *app.AppManager
	logger  *logger.Logger
}

func NewSandboxHandler(manager *app.AppManager, log *logger.Logger) *SandboxHandler {
	return &SandboxHandler{
		manager: manager,
		logger:  log,
	}
}

// GetSandboxStatus returns the sandbox status for an app
func (h *SandboxHandler) GetSandboxStatus(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/apps/{name}/sandbox
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	appName := parts[4]

	sandbox, err := h.manager.GetSandbox(appName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"app_name":     sandbox.GetAppName(),
		"active":       sandbox.IsActive(),
		"created_at":   sandbox.GetCreatedAt(),
		"limits":       sandbox.GetLimits(),
		"capabilities": sandbox.GetCapabilities(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetResourceUsage returns current resource usage for an app
func (h *SandboxHandler) GetResourceUsage(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/apps/{name}/resources
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	appName := parts[4]

	usage, err := h.manager.GetResourceUsage(appName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

// UpdateLimits modifies resource limits for an app
func (h *SandboxHandler) UpdateLimits(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/apps/{name}/limits
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	appName := parts[4]

	var limits app.ResourceLimits
	if err := json.NewDecoder(r.Body).Decode(&limits); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.manager.UpdateResourceLimits(appName, limits); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}
