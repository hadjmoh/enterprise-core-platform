package app

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PathSanitizer ensures file paths are safely scoped to a root directory
type PathSanitizer struct {
	RootPath string
}

// NewPathSanitizer creates a new sanitizer for the given root
func NewPathSanitizer(root string) (*PathSanitizer, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for root: %w", err)
	}
	return &PathSanitizer{RootPath: absRoot}, nil
}

// Sanitize ensures the joined path (root + subpath) is within the root directory
func (ps *PathSanitizer) Sanitize(subpath string) (string, error) {
	// Join the root with the user-provided subpath
	fullPath := filepath.Join(ps.RootPath, subpath)
	
	// Clean the path to resolve dots (e.g. /foo/../bar -> /bar)
	cleanPath := filepath.Clean(fullPath)
	
	// Verify that the clean path still has the root prefix
	if !strings.HasPrefix(cleanPath, ps.RootPath) {
		return "", fmt.Errorf("path traversal attempt detected: %s is outside root %s", subpath, ps.RootPath)
	}
	
	return cleanPath, nil
}
