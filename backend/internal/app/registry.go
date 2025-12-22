package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

// RegistryAppInfo represents an app in the registry/store
type RegistryAppInfo struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Title        string   `json:"title"`
	Author       string   `json:"author"`
	Category     string   `json:"category"`
	Description  string   `json:"description"`
	LongDesc     string   `json:"longDescription,omitempty"`
	Icon         string   `json:"icon,omitempty"`
	Screenshots  []string `json:"screenshots,omitempty"`
	Permissions  []string `json:"permissions,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Downloads    int      `json:"downloads,omitempty"`
	Rating       float64  `json:"rating,omitempty"`
	DownloadURL  string   `json:"downloadUrl,omitempty"`
}

// AppRegistry manages the catalog of available apps
type AppRegistry struct {
	mu    sync.RWMutex
	apps  map[string]*RegistryAppInfo
	path  string
}

// RegistryData represents the JSON structure of the registry file
type RegistryData struct {
	Apps []*RegistryAppInfo `json:"apps"`
}

// NewAppRegistry creates a new app registry
func NewAppRegistry(registryPath string) (*AppRegistry, error) {
	registry := &AppRegistry{
		apps: make(map[string]*RegistryAppInfo),
		path: registryPath,
	}
	
	if err := registry.LoadRegistry(); err != nil {
		return nil, fmt.Errorf("failed to load registry: %w", err)
	}
	
	return registry, nil
}

// LoadRegistry loads the app catalog from the JSON file
func (r *AppRegistry) LoadRegistry() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// If file doesn't exist, start with empty registry
	if _, err := os.Stat(r.path); os.IsNotExist(err) {
		return nil
	}
	
	data, err := os.ReadFile(r.path)
	if err != nil {
		return fmt.Errorf("failed to read registry file: %w", err)
	}
	
	var registryData RegistryData
	if err := json.Unmarshal(data, &registryData); err != nil {
		return fmt.Errorf("failed to parse registry JSON: %w", err)
	}
	
	// Build the map
	for _, app := range registryData.Apps {
		r.apps[app.Name] = app
	}
	
	return nil
}

// ListAll returns all apps in the registry
func (r *AppRegistry) ListAll() []*RegistryAppInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	apps := make([]*RegistryAppInfo, 0, len(r.apps))
	for _, app := range r.apps {
		apps = append(apps, app)
	}
	
	return apps
}

// GetByName retrieves a specific app by name
func (r *AppRegistry) GetByName(name string) (*RegistryAppInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	app, exists := r.apps[name]
	if !exists {
		return nil, fmt.Errorf("app not found: %s", name)
	}
	
	return app, nil
}

// Search filters apps by query string (searches name, title, description, tags)
func (r *AppRegistry) Search(query string) []*RegistryAppInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	query = strings.ToLower(query)
	results := make([]*RegistryAppInfo, 0)
	
	for _, app := range r.apps {
		// Search in name, title, description, and tags
		if strings.Contains(strings.ToLower(app.Name), query) ||
			strings.Contains(strings.ToLower(app.Title), query) ||
			strings.Contains(strings.ToLower(app.Description), query) ||
			containsTag(app.Tags, query) {
			results = append(results, app)
		}
	}
	
	return results
}

// FilterByCategory returns apps in a specific category
func (r *AppRegistry) FilterByCategory(category string) []*RegistryAppInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	results := make([]*RegistryAppInfo, 0)
	for _, app := range r.apps {
		if strings.EqualFold(app.Category, category) {
			results = append(results, app)
		}
	}
	
	return results
}

// Helper function to check if tags contain query
func containsTag(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}
