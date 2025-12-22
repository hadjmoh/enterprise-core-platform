package app

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"regexp"
)

// AppManifest defines the structure of 'prospect.yaml'
type AppManifest struct {
	APIVersion  string       `yaml:"apiVersion"`
	Kind        string       `yaml:"kind"`
	Metadata    AppMetadata  `yaml:"metadata"`
	Permissions []string     `yaml:"permissions"`
	Inputs      []AppInput   `yaml:"inputs,omitempty"`
	Dashboards  []AppDashboard `yaml:"dashboards,omitempty"`
}

type AppMetadata struct {
	Name        string   `yaml:"name"`
	DisplayName string   `yaml:"displayName"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	License     string   `yaml:"license"`
	Icon        string   `yaml:"icon,omitempty"` // Base64 or path
	Tags        []string `yaml:"tags,omitempty"`
}

type AppInput struct {
	Type        string                 `yaml:"type"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Properties  map[string]interface{} `yaml:"properties"`
}

type AppDashboard struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
	Path  string `yaml:"path"` // Path to dashboard definition file
}

// ParseManifest reads and validates a manifest from a reader
func ParseManifest(r io.Reader) (*AppManifest, error) {
	var m AppManifest
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	if err := m.Validate(); err != nil {
		return nil, err
	}

	return &m, nil
}

// Validate ensures the manifest meets all requirements
func (m *AppManifest) Validate() error {
	if m.APIVersion != "prospect/v1alpha1" {
		return fmt.Errorf("unsupported apiVersion: %s", m.APIVersion)
	}
	if m.Kind != "App" {
		return fmt.Errorf("invalid kind: %s, expected 'App'", m.Kind)
	}
	
	nameRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !nameRegex.MatchString(m.Metadata.Name) {
		return fmt.Errorf("invalid app name '%s': must consist of lowercase alphanumeric characters and hyphens", m.Metadata.Name)
	}

	if m.Metadata.Version == "" {
		return fmt.Errorf("version is required")
	}

	return nil
}
