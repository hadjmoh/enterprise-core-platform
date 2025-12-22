package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"enterprise-core/backend/internal/app"
	"enterprise-core/backend/pkg/logger"
)

type AppHandler struct {
	manager *app.AppManager
	logger  *logger.Logger
}

func NewAppHandler(manager *app.AppManager, log *logger.Logger) *AppHandler {
	return &AppHandler{
		manager: manager,
		logger:  log,
	}
}

func (h *AppHandler) ListApps(w http.ResponseWriter, r *http.Request) {
	apps := h.manager.ListApps()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps)
}

func (h *AppHandler) RegisterApp(w http.ResponseWriter, r *http.Request) {
	// For now, this is a stub that accepts a manifest JSON in the body
	var manifest app.AppManifest
	if err := json.NewDecoder(r.Body).Decode(&manifest); err != nil {
		http.Error(w, "Invalid manifest: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.manager.RegisterApp(&manifest); err != nil {
		http.Error(w, "Failed to register app: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}

func (h *AppHandler) InstallApp(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	// Max upload size: 50MB
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("bundle")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Verify extension
	if !strings.HasSuffix(header.Filename, ".tar.gz") {
		http.Error(w, "Only .tar.gz bundles are supported", http.StatusBadRequest)
		return
	}

	// Save to temp file
	tempFile, err := os.CreateTemp("", "app-upload-*.tar.gz")
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name()) // Cleanup

	if _, err := io.Copy(tempFile, file); err != nil {
		http.Error(w, "Upload failed", http.StatusInternalServerError)
		return
	}
	tempFile.Close()

	if err := h.manager.InstallApp(tempFile.Name()); err != nil {
		http.Error(w, "Installation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "installed"})
}

func (h *AppHandler) UninstallApp(w http.ResponseWriter, r *http.Request) {
	// Path param /api/v1/apps/{name}
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	appName := parts[4] // /api/v1/apps/my-app

	if err := h.manager.UninstallApp(appName); err != nil {
		http.Error(w, "Uninstall failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "uninstalled"})
}

func (h *AppHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/apps/{name}/dashboards/{id}
	path := r.URL.Path
	parts := strings.Split(path, "/")
	// Expected: /api/v1/apps/my-app/dashboards/main
	// 0: "", 1: api, 2: v1, 3: apps, 4: my-app, 5: dashboards, 6: main
	if len(parts) < 7 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	appName := parts[4]
	dashboardID := parts[6]

	layout, err := h.manager.GetDashboard(appName, dashboardID)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(layout)
}

// ListAvailableApps returns all apps from the registry
func (h *AppHandler) ListAvailableApps(w http.ResponseWriter, r *http.Request) {
apps := h.manager.GetAvailableApps()

// Enrich with installation status
type AppWithStatus struct {
*app.RegistryAppInfo
Installed bool `json:"installed"`
}

enriched := make([]AppWithStatus, len(apps))
for i, a := range apps {
enriched[i] = AppWithStatus{
RegistryAppInfo: a,
Installed:   h.manager.IsInstalled(a.Name),
}
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(enriched)
}

// GetAppDetailsHandler returns detailed metadata for a specific app
func (h *AppHandler) GetAppDetailsHandler(w http.ResponseWriter, r *http.Request) {
// Path: /api/v1/apps/available/{name}
path := r.URL.Path
parts := strings.Split(path, "/")
if len(parts) < 6 {
http.Error(w, "Invalid path", http.StatusBadRequest)
return
}
appName := parts[5]

appMeta, err := h.manager.GetAppDetails(appName)
if err != nil {
http.Error(w, err.Error(), http.StatusNotFound)
return
}

// Enrich with installation status
type AppWithStatus struct {
*app.RegistryAppInfo
Installed bool `json:"installed"`
}

enriched := AppWithStatus{
RegistryAppInfo: appMeta,
Installed:   h.manager.IsInstalled(appName),
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(enriched)
}

// SearchAppsHandler searches for apps matching a query
func (h *AppHandler) SearchAppsHandler(w http.ResponseWriter, r *http.Request) {
query := r.URL.Query().Get("q")
if query == "" {
// Return all apps if no query
h.ListAvailableApps(w, r)
return
}

apps := h.manager.SearchApps(query)

// Enrich with installation status
type AppWithStatus struct {
*app.RegistryAppInfo
Installed bool `json:"installed"`
}

enriched := make([]AppWithStatus, len(apps))
for i, a := range apps {
enriched[i] = AppWithStatus{
RegistryAppInfo: a,
Installed:   h.manager.IsInstalled(a.Name),
}
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(enriched)
}
