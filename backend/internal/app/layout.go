package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DashboardLayout defines the structure of a dashboard JSON file
type DashboardLayout struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Rows        []Row  `json:"rows"`
}

type Row struct {
	Height  string   `json:"height"` // e.g. "400px" or "auto"
	Columns []Column `json:"columns"`
}

type Column struct {
	Width  int     `json:"width"` // 1-12 grid system
	Panels []Panel `json:"panels"`
}

type Panel struct {
	ID      string                 `json:"id"`
	Title   string                 `json:"title"`
	Type    string                 `json:"type"` // line, bar, area, stat, table
	Query   string                 `json:"query"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// ParseLayout reads a dashboard layout from a reader
func ParseLayout(r io.Reader) (*DashboardLayout, error) {
	var layout DashboardLayout
	if err := json.NewDecoder(r).Decode(&layout); err != nil {
		return nil, fmt.Errorf("failed to parse dashboard layout: %w", err)
	}
	return &layout, nil
}

// GetDashboard retrieves a specific dashboard for an installed app
// It assumes the app is installed at <root>/<appName>
// and dashboards are in <root>/<appName>/dashboards/<dashboardID>.json
func (am *AppManager) GetDashboard(appName, dashboardID string) (*DashboardLayout, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Verify app exists
	if _, exists := am.installedApps[appName]; !exists {
		return nil, fmt.Errorf("app '%s' not found", appName)
	}

	// Sanitize app path
	appPath, err := am.sanitizer.Sanitize(appName)
	if err != nil {
		return nil, err
	}

	// Construct dashboard path
	// We need to be careful here. sanitize(appName) gives us the app root.
	// We want to join with "dashboards" and "<dashboardID>.json"
	// But we should also verify that the resulting file is definitely inside the app root.
	// The Sanitize method on PathSanitizer checks against the Manager's root.
	// Here we want to ensure we don't slip out of the app's dashboard dir.
	
	dashboardDir := filepath.Join(appPath, "dashboards")
	filename := dashboardID + ".json"
	
	// Quick directory traversal check on dashboardID
	if dashboardID == "" || filepath.Base(filename) != filename {
		return nil, fmt.Errorf("invalid dashboard ID")
	}

	fullPath := filepath.Join(dashboardDir, filename)
	
	// Verify it still starts with appPath to be ultra safe (redundant if using filepath.Join correctly with clean ID, but good practice)
	// Actually better to use the sanitizer again if strictly needed, but manual check is fine here.
	cleanPath := filepath.Clean(fullPath)
	if filepath.Dir(filepath.Dir(cleanPath)) != filepath.Clean(appPath) {
		// Expecting cleanPath to be .../appName/dashboards/file.json
		// So Dir is .../dashboards
		// Dir(Dir) is .../appName
		// This check is a bit brittle, let's just trust filepath.Join with a validated filename (Base check above).
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("dashboard '%s' not found", dashboardID)
		}
		return nil, err
	}
	defer file.Close()

	return ParseLayout(file)
}
